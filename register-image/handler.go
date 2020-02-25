package function

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/openfaas/openfaas-cloud/sdk"
)

// ImageRequest requests to create a Docker registry entry
type ImageRequest struct {
	Image string `json:"image"` // Image is the full image path and registry combined

}

// GetRepository returns the name only without the server or tag
func (r ImageRequest) GetRepository() string {
	image := r.Image[strings.Index(r.Image, "/")+1:]
	tagIndex := strings.Index(image, ":")
	if tagIndex > -1 {
		image = image[:tagIndex]
	}
	return strings.ToLower(image)
}

// Handle accepts an ImageRequest and creates the registry entry if required
// in a container registry such as ECR using secrets made available through
// the function.
func Handle(req []byte) string {
	validate := sdk.HmacEnabled()

	if validate {
		hmacErr := sdk.ValidHMAC(&req, "payload-secret", os.Getenv("Http_X_Cloud_Signature"))
		if hmacErr != nil {
			log.Printf("hmac error %s\n", hmacErr.Error())
			os.Exit(1)
		}
	}

	r := ImageRequest{}
	err := json.Unmarshal(req, &r)

	if err != nil {
		return fmt.Errorf(`unable to unmarshal request: "%q" input: %s\n`, string(req), err.Error()).Error()
	}

	if len(r.Image) == 0 {
		return fmt.Errorf(`the field "image" is required in the request`).Error()
	}

	client := ecr.New(session.New(), aws.NewConfig().WithRegion(os.Getenv("AWS_DEFAULT_REGION")))

	mutability := "MUTABLE"
	image := r.GetRepository()
	repoReq := ecr.CreateRepositoryInput{
		ImageTagMutability: &mutability,
		RepositoryName:     &image,
	}

	log.Printf("Attempting to create repo: %s\n", image)

	output, err := client.CreateRepository(&repoReq)

	if err != nil {
		log.Printf("CreateRepository error: %s\n", err.Error())
		os.Exit(1)
	}

	return fmt.Sprintf("Created repo: %s\n", output.String())
}
