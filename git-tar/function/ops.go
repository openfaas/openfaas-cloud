package function

import (
	"archive/tar"
	"bytes"
	"encoding/hex"
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

	"github.com/openfaas/faas-cli/schema"

	hmac "github.com/alexellis/hmac"
	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/openfaas-cloud/sdk"
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

		pushRepositoryURL := os.Getenv("push_repository_url")

		if len(pushRepositoryURL) == 0 {
			fmt.Fprintf(os.Stderr, "push_repository_url env-var not set")
			return nil, fmt.Errorf("push_repository_url env-var not set")
		}

		imageName := formatImageShaTag(pushRepositoryURL, &v, pushEvent.AfterCommitID,
			pushEvent.Repository.Owner.Login, pushEvent.Repository.Name)

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

func formatImageShaTag(registry string, function *stack.Function, sha string, owner string, repo string) string {
	imageName := function.Image

	repoIndex := strings.LastIndex(imageName, "/")
	if repoIndex > -1 {
		imageName = imageName[repoIndex+1:]
	}
	imageName = schema.BuildImageName(schema.SHAFormat, imageName, sha, "master")

	var imageRef string
	sharedRepo := strings.HasSuffix(registry, "/")
	if sharedRepo {
		imageRef = registry[:len(registry)-1] + "/" + owner + "-" + repo + "-" + imageName
	} else {
		imageRef = registry + "/" + owner + "/" + repo + "-" + imageName
	}

	return imageRef
}

func clone(pushEvent sdk.PushEvent) (string, error) {
	workDir := os.TempDir()
	destPath := path.Join(workDir, path.Join(pushEvent.Repository.Owner.Login, pushEvent.Repository.Name))

	if _, err := os.Stat(destPath); err == nil {
		truncateErr := os.RemoveAll(destPath)
		if truncateErr != nil {
			return "", truncateErr
		}
	}

	userDir := path.Join(workDir, pushEvent.Repository.Owner.Login)
	err := os.MkdirAll(userDir, 0777)

	if err != nil {
		return "", fmt.Errorf("cannot create user-dir: %s", userDir)
	}

	git := exec.Command("git", "clone", pushEvent.Repository.CloneURL)
	git.Dir = path.Join(workDir, pushEvent.Repository.Owner.Login)
	log.Println(git.Dir)
	err = git.Start()
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

func deploy(tars []tarEntry, pushEvent sdk.PushEvent, stack *stack.Services, status *sdk.Status) error {

	owner := pushEvent.Repository.Owner.Login
	repoName := pushEvent.Repository.Name
	url := pushEvent.Repository.CloneURL
	afterCommitID := pushEvent.AfterCommitID
	installationID := pushEvent.Installation.ID

	c := http.Client{}
	gatewayURL := os.Getenv("gateway_url")

	for _, tarEntry := range tars {
		fmt.Println("Deploying service - " + tarEntry.functionName)

		status.AddStatus(sdk.StatusPending, fmt.Sprintf("%s function deploy is in progress", tarEntry.functionName),
			sdk.BuildFunctionContext(tarEntry.functionName))
		reportStatus(status)
		log.Printf(status.AuthToken)

		fileOpen, err := os.Open(tarEntry.fileName)
		if err != nil {
			return err
		}

		httpReq, _ := http.NewRequest(http.MethodPost, gatewayURL+"function/buildshiprun", fileOpen)

		httpReq.Header.Add("Repo", repoName)
		httpReq.Header.Add("Owner", owner)
		httpReq.Header.Add("Url", url)
		httpReq.Header.Add("Installation_id", fmt.Sprintf("%d", installationID))
		httpReq.Header.Add("Service", tarEntry.functionName)
		httpReq.Header.Add("Image", tarEntry.imageName)
		httpReq.Header.Add("Sha", afterCommitID)

		envJSON, marshalErr := json.Marshal(stack.Functions[tarEntry.functionName].Environment)
		if marshalErr != nil {
			log.Printf("Error marshaling %d env-vars for function %s, %s", len(stack.Functions[tarEntry.functionName].Environment), tarEntry.functionName, marshalErr)
		}

		httpReq.Header.Add("Env", string(envJSON))

		secretsJSON, marshalErr := json.Marshal(stack.Functions[tarEntry.functionName].Secrets)
		if marshalErr != nil {
			log.Printf("Error marshaling secrets for function %s, %s", tarEntry.functionName, marshalErr)
		}

		httpReq.Header.Add("Secrets", string(secretsJSON))

		res, reqErr := c.Do(httpReq)
		if reqErr != nil {
			fmt.Fprintf(os.Stderr, fmt.Errorf("unable to deploy function via buildshiprun: %s", reqErr.Error()).Error())
		}

		fmt.Println("Service deployed ", tarEntry.functionName, res.Status, owner)
	}
	return nil
}

func importSecrets(pushEvent sdk.PushEvent, stack *stack.Services, clonePath string) error {
	gatewayURL := os.Getenv("gateway_url")

	secretCount := 0
	for _, fn := range stack.Functions {
		secretCount += len(fn.Secrets)
	}

	owner := pushEvent.Repository.Owner.Login
	secretPath := path.Join(clonePath, "secrets.yml")

	// No secrets supplied.
	if fileInfo, err := os.Stat(secretPath); fileInfo == nil || err != nil {
		return nil
	}

	fileBytes, err := ioutil.ReadFile(secretPath)

	if err != nil {
		return fmt.Errorf("unable to read secret: %s", secretPath)
	}

	webhookSecretKey := os.Getenv("github_webhook_secret")
	hash := hmac.Sign(fileBytes, []byte(webhookSecretKey))

	c := http.Client{}
	reader := bytes.NewReader(fileBytes)
	httpReq, _ := http.NewRequest(http.MethodPost, gatewayURL+"function/import-secrets", reader)

	httpReq.Header.Add("Owner", owner)
	httpReq.Header.Add("X-Hub-Signature", "sha1="+hex.EncodeToString(hash))

	res, reqErr := c.Do(httpReq)

	if reqErr != nil {
		fmt.Fprintf(os.Stderr, fmt.Errorf("error reaching import-secrets function: %s", reqErr.Error()).Error())
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted && res.StatusCode != http.StatusOK {
		resBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("error reading response from import-secrets: %s", err.Error())
		}

		return fmt.Errorf("import-secrets returned error: %s, res: %s", err.Error(), string(resBytes))
	}

	auditEvent := sdk.AuditEvent{
		Message: fmt.Sprintf("Parsed sealed secrets for owner: %s. Parsed %d secrets, from %d functions", owner, secretCount, len(stack.Functions)),
		Owner:   pushEvent.Repository.Owner.Login,
		Repo:    pushEvent.Repository.Name,
		Source:  Source,
	}

	sdk.PostAudit(auditEvent)

	fmt.Println("Parsed sealed secrets", res.Status, owner)

	return nil
}
