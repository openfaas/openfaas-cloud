package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	minio "github.com/minio/minio-go"
	"github.com/openfaas/openfaas-cloud/sdk"
)

// Handle sends the function logs to the
// pipeline to be exposed in the dashboard
func Handle(req []byte) string {
	method := os.Getenv("Http_Method")

	if method == http.MethodPost {
		hmacErr := sdk.ValidHMAC(&req, "payload-secret", os.Getenv("Http_X_Cloud_Signature"))
		if hmacErr != nil {
			log.Printf("hmac error %s\n", hmacErr.Error())
			os.Exit(1)
		}
	}

	region := regionName()

	bucketName := bucketName()

	minioClient, connectErr := connectToMinio(region)
	if connectErr != nil {
		log.Printf("S3/Minio connection error %s\n", connectErr.Error())
		os.Exit(1)
	}

	switch method {
	case http.MethodPost:

		pipelineLog := sdk.PipelineLog{}
		json.Unmarshal(req, &pipelineLog)

		minioClient.MakeBucket(bucketName, region)

		reader := bytes.NewReader([]byte(pipelineLog.Data))
		fullPath := getPath(bucketName, &pipelineLog)
		n, err := minioClient.PutObject(bucketName,
			fullPath,
			reader,
			int64(reader.Len()),
			minio.PutObjectOptions{})

		if err != nil {
			log.Printf("error writing: %s, error: %s", fullPath, err.Error())
			os.Exit(1)
		}
		return fmt.Sprintf("Wrote %d bytes to %s\n", n, fullPath)

	case http.MethodGet:
		queryRaw := os.Getenv("Http_Query")
		query, parseErr := url.ParseQuery(queryRaw)
		if parseErr != nil {
			return parseErr.Error()
		}

		p := sdk.PipelineLog{
			CommitSHA: query.Get("commitSHA"),
			Function:  query.Get("function"),
			RepoPath:  query.Get("repoPath"),
		}

		fullPath := getPath(bucketName, &p)
		log.Printf("Reading %s\n", fullPath)
		obj, err := minioClient.GetObject(bucketName, fullPath, minio.GetObjectOptions{})

		if err != nil {
			log.Printf("error reading: %s, error: %s", fullPath, err.Error())
			os.Exit(1)
		}

		logBytes, _ := ioutil.ReadAll(obj)

		return string(logBytes)
	}

	return fmt.Sprintf("pipeline-log, unknown request")
}

func connectToMinio(region string) (*minio.Client, error) {

	endpoint := os.Getenv("s3_url")

	tlsEnabled := tlsEnabled()

	secretKey, _ := sdk.ReadSecret("s3-secret-key")
	accessKey, _ := sdk.ReadSecret("s3-access-key")

	return minio.New(endpoint, accessKey, secretKey, tlsEnabled)
}

// getPath produces a string such as pipeline/alexellis/super-pancake-fn/commit-id/fn1/
func getPath(bucket string, p *sdk.PipelineLog) string {
	fileName := "build.log"
	return fmt.Sprintf("%s/%s/%s/%s/%s", bucket, p.RepoPath, p.CommitSHA, p.Function, fileName)
}

func tlsEnabled() bool {
	if connection := os.Getenv("s3_tls"); connection == "true" || connection == "1" {
		return true
	}
	return false
}

func bucketName() string {
	bucketName, exist := os.LookupEnv("s3_bucket")
	if exist == false || len(bucketName) == 0 {
		bucketName = "pipeline"
		log.Printf("Bucket name not found, set to default: %v\n", bucketName)
	}
	return bucketName
}

func regionName() string {
	regionName, exist := os.LookupEnv("s3_region")
	if exist == false || len(regionName) == 0 {
		regionName = "us-east-1"
	}
	return regionName
}
