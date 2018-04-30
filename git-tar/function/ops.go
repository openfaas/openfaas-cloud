package function

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/openfaas-cloud/sdk"

	"github.com/alexellis/derek/auth"
	"github.com/google/go-github/github"
)

type tarEntry struct {
	fileName     string
	functionName string
	imageName    string
}

type cfg struct {
	Ref      string  `json:"ref"`
	Frontend *string `json:"frontend,omitempty"`
}

type eventInfo struct {
	service    string
	owner      string
	repository string
	//image          string
	sha            string
	url            string
	installationID int
	//	environment    map[string]string
}

const (
	defaultPrivateKeyName = "private_key.pem"
)

func parseYAML(pushEvent sdk.PushEvent, filePath string) (*stack.Services, error) {
	parsed, err := stack.ParseYAMLFile(path.Join(filePath, "stack.yml"), "", "")
	return parsed, err
}

func shrinkwrap(pushEvent sdk.PushEvent, filePath string) (string, error) {
	buildCmd := exec.Command("faas-cli", "build", "-f", "stack.yml", "--shrinkwrap")
	buildCmd.Dir = filePath
	err := buildCmd.Start()
	if err != nil {
		return "", fmt.Errorf("Cannot start faas-cli build: %t", err)
	}
	err = buildCmd.Wait()

	return filePath, err
}

func makeTar(pushEvent sdk.PushEvent, filePath string, services *stack.Services) ([]tarEntry, error) {
	tars := []tarEntry{}

	fmt.Printf("Tar up %s\n", filePath)

	for k, v := range services.Functions {
		fmt.Println("Creating tar for: ", v.Handler, k)

		tarPath := path.Join(filePath, fmt.Sprintf("%s.tar", k))
		contextTar, err := os.Create(tarPath)
		if err != nil {
			return []tarEntry{}, err
		}

		tarWriter := tar.NewWriter(contextTar)
		defer tarWriter.Close()

		base := filepath.Join(filePath, filepath.Join("build", k))

		imageName := formatImageShaTag("registry:5000", &v, pushEvent.AfterCommitID)
		config := cfg{
			Ref: imageName,
		}

		configBytes, _ := json.Marshal(config)
		configErr := ioutil.WriteFile(path.Join(base, "config"), configBytes, 0600)
		if configErr != nil {
			return nil, configErr
		}

		// fmt.Println("Base: ", base, filePath, k)
		err = filepath.Walk(base, func(path string, f os.FileInfo, pathErr error) error {
			if pathErr != nil {
				return pathErr
			}

			if f.Name() == "context.tar" {
				return nil
			}

			targetFile, err1 := os.Open(path)
			log.Println(path)

			if err1 != nil {
				return err1
			}

			header, headerErr := tar.FileInfoHeader(f, f.Name())
			if headerErr != nil {
				return headerErr
			}

			header.Name = strings.TrimPrefix(path, base)
			// log.Printf("header.Name '%s'\n", header.Name)
			if header.Name != "/config" {
				header.Name = filepath.Join("context", header.Name)
			}

			header.Name = strings.TrimPrefix(header.Name, "/")

			// log.Println("tar - header.Name ", header.Name)
			if err1 = tarWriter.WriteHeader(header); err != nil {
				return err1
			}

			if f.Mode().IsDir() {
				return nil
			}

			_, err1 = io.Copy(tarWriter, targetFile)
			return err1
		})

		if err != nil {
			return []tarEntry{}, err
		}
		tars = append(tars, tarEntry{fileName: tarPath, functionName: strings.TrimSpace(k), imageName: imageName})
	}

	return tars, nil
}

func formatImageShaTag(registry string, function *stack.Function, sha string) string {
	tag := ":latest"
	imageName := function.Image
	tagIndex := strings.LastIndex(function.Image, ":")
	if tagIndex > 0 {
		tag = function.Image[tagIndex:]
		imageName = function.Image[:tagIndex]
	}

	imageName = registry + "/" + imageName + tag + "-" + sha
	return imageName
}

func clone(pushEvent sdk.PushEvent) (string, error) {
	workDir := os.TempDir()
	destPath := path.Join(workDir, pushEvent.Repository.Name)
	if _, err := os.Stat(destPath); err == nil {
		truncateErr := os.RemoveAll(destPath)
		if truncateErr != nil {
			return "", truncateErr
		}
	}

	git := exec.Command("git", "clone", pushEvent.Repository.CloneURL)
	git.Dir = workDir
	err := git.Start()
	if err != nil {
		return "", fmt.Errorf("Cannot start git: %t", err)
	}
	err = git.Wait()

	git = exec.Command("git", "checkout", pushEvent.AfterCommitID)
	git.Dir = destPath
	err = git.Start()
	if err != nil {
		return "", fmt.Errorf("Cannot start git checkout: %t", err)
	}
	err = git.Wait()

	return destPath, err
}

func deploy(tars []tarEntry, pushEvent sdk.PushEvent, stack *stack.Services) error {

	owner := pushEvent.Repository.Owner.Login
	repoName := pushEvent.Repository.Name
	url := pushEvent.Repository.CloneURL
	afterCommitID := pushEvent.AfterCommitID
	installationId := pushEvent.Installation.Id

	c := http.Client{}
	gatewayURL := os.Getenv("gateway_url")

	event := getEvent(pushEvent)
	serviceValue := owner + "-" + repoName

	reportStatus("pending", fmt.Sprintf("%s build and deploy is in progress", serviceValue), "DEPLOY", event)

	hasFailure := false
	failedFunctions := make(map[string]string)

	for _, tarEntry := range tars {
		fmt.Println("Deploying service - " + tarEntry.functionName)

		fileOpen, err := os.Open(tarEntry.fileName)
		if err != nil {
			return err
		}

		httpReq, _ := http.NewRequest(http.MethodPost, gatewayURL+"function/buildshiprun", fileOpen)

		httpReq.Header.Add("Repo", repoName)
		httpReq.Header.Add("Owner", owner)
		httpReq.Header.Add("Url", url)
		httpReq.Header.Add("Installation_id", fmt.Sprintf("%d", installationId))
		httpReq.Header.Add("Service", tarEntry.functionName)
		httpReq.Header.Add("Image", tarEntry.imageName)
		httpReq.Header.Add("Sha", afterCommitID)

		envJSON, marshalErr := json.Marshal(stack.Functions[tarEntry.functionName].Environment)
		if marshalErr != nil {
			log.Printf("Error marshaling %d env-vars for function %s, %s", len(stack.Functions[tarEntry.functionName].Environment), tarEntry.functionName, marshalErr)
		}

		httpReq.Header.Add("Env", string(envJSON))

		res, reqErr := c.Do(httpReq)
		if reqErr != nil {
			fmt.Fprintf(os.Stderr, fmt.Errorf("unable to deploy function via buildshiprun: %s", reqErr.Error()).Error())
			hasFailure = true
			failedFunctions[tarEntry.functionName] = reqErr.Error()
		}

		fmt.Println("Service deployed ", tarEntry.functionName, res.Status, owner)
	}

	if !hasFailure {
		reportStatus("success", fmt.Sprintf("service successfully deployed as: %s", serviceValue), "DEPLOY", event)
	} else {
		reportStatus("failure", fmt.Sprintf("failed to deploy service %s: %v", serviceValue, failedFunctions), "DEPLOY", event)
	}

	return nil
}

func reportStatus(status string, desc string, statusContext string, event *eventInfo) {

	if os.Getenv("report_status") != "true" {
		return
	}

	url := event.url
	if status == "success" {
		publicURL := os.Getenv("gateway_public_url")
		// for success status if gateway's public url id set the deployed
		// function url is used in the commit status
		if publicURL == "" {
			publicURL = os.Getenv("gateway_url")
		}
		url = publicURL + "overview?user=" + event.owner + "&sha=" + event.sha + "&repo=" + event.repository
	}

	ctx := context.Background()

	// NOTE: currently vendored derek auth package doesn't take the private key as input;
	// but expect it to be present at : "/run/secrets/derek-private-key"
	// as docker /secrets dir has limited permission we are bound to use secret named
	// as "derek-private-key"
	// the below lines should  be uncommented once the package is updated in derek project
	// privateKeyPath := getPrivateKey()
	// token, tokenErr := auth.MakeAccessTokenForInstallation(os.Getenv("github_app_id"),
	//      event.installationID, privateKeyPath)

	repoStatus := buildStatus(status, desc, statusContext, url)

	log.Printf("Status: %s, GitHub AppID: %d, Repo: %s, Owner: %s", status, event.installationID, event.repository, event.owner)

	token, tokenErr := auth.MakeAccessTokenForInstallation(os.Getenv("github_app_id"), event.installationID)
	if tokenErr != nil {
		fmt.Printf("failed to report status %v, error: %s\n", repoStatus, tokenErr.Error())
		return
	}

	if token == "" {
		fmt.Printf("failed to report status %v, error: authentication failed Invalid token\n", repoStatus)
		return
	}

	client := auth.MakeClient(ctx, token)

	_, _, apiErr := client.Repositories.CreateStatus(ctx, event.owner, event.repository, event.sha, repoStatus)
	if apiErr != nil {
		fmt.Printf("failed to report status %v, error: %s\n", repoStatus, apiErr.Error())
		return
	}
}

func getPrivateKey() string {
	// we are taking the secrets name from the env, by default it is fixed
	// to private_key.pem.
	// Although user can make the secret with a specific name and provide
	// it in the stack.yaml and also specify the secret name in github.yml
	privateKeyName := os.Getenv("private_key")
	if privateKeyName == "" {
		privateKeyName = defaultPrivateKeyName
	}
	privateKeyPath := "/run/secrets/" + privateKeyName
	return privateKeyPath
}

func buildStatus(status string, desc string, context string, url string) *github.RepoStatus {
	return &github.RepoStatus{State: &status, TargetURL: &url, Description: &desc, Context: &context}
}

func getEvent(pushEvent sdk.PushEvent) *eventInfo {
	info := eventInfo{}

	info.service = pushEvent.Repository.Name
	info.owner = pushEvent.Repository.Owner.Login
	info.repository = pushEvent.Repository.Name
	info.sha = pushEvent.AfterCommitID
	info.url = pushEvent.Repository.CloneURL
	info.installationID = pushEvent.Installation.Id

	return &info
}
