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
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"code.cloudfoundry.org/bytefmt"
	"github.com/openfaas/faas-cli/schema"

	"github.com/alexellis/derek/auth"
	"github.com/alexellis/hmac"
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

func fetchTemplates(filePath string) error {
	templateRepos := formatTemplateRepos()

	for _, repo := range templateRepos {
		pullCmd := exec.Command("faas-cli", "template", "pull", repo)
		pullCmd.Dir = filePath
		err := pullCmd.Start()
		if err != nil {
			return fmt.Errorf("Failed to start faas-cli template pull: %t", err)
		}
		err = pullCmd.Wait()
		if err != nil {
			return fmt.Errorf("Failed to wait faas-cli template pull: %t", err)
		}
	}

	return nil
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
			if header.Name != "/config" {
				header.Name = filepath.Join("context", header.Name)
			}

			header.Name = strings.TrimPrefix(header.Name, "/")

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

		tars = append(tars,
			tarEntry{fileName: tarPath,
				functionName: strings.TrimSpace(k),
				imageName:    imageName,
			})
	}

	return tars, nil
}

func formatImageShaTag(registry string, function *stack.Function, sha string, owner string, repo string) string {
	imageName := function.Image

	repoIndex := strings.LastIndex(imageName, "/")
	if repoIndex > -1 {
		imageName = imageName[repoIndex+1:]
	}

	sha = sdk.FormatShortSHA(sha)

	imageName = schema.BuildImageName(schema.BranchAndSHAFormat, imageName, sha, buildBranch())

	var imageRef string
	sharedRepo := strings.HasSuffix(registry, "/")
	if sharedRepo {
		imageRef = registry[:len(registry)-1] + "/" + owner + "-" + repo + "-" + imageName
	} else {
		imageRef = registry + "/" + owner + "/" + repo + "-" + imageName
	}

	return imageRef
}

type githubAuthToken struct {
	appID          string
	installationID int
	privateKeyPath string
	token          string
}

func (t *githubAuthToken) getToken() (string, error) {
	if t.token != "" {
		return t.token, nil
	}

	privateKey, ioErr := ioutil.ReadFile(t.privateKeyPath)
	if ioErr != nil {
		return "", ioErr
	}

	token, err := auth.MakeAccessTokenForInstallation(t.appID, t.installationID, string(privateKey))

	if err != nil {
		return "", err
	}

	t.token = token

	return token, nil
}

func (t *githubAuthToken) getInstallationID() int {
	return t.installationID
}

type tokener interface {
	getToken() (string, error)
	getInstallationID() int
}

func getRepositoryURL(e sdk.PushEvent, authToken tokener) (string, error) {
	cu := e.Repository.CloneURL

	if e.Repository.Private {
		u, err := url.Parse(cu)

		if err != nil {
			return "", fmt.Errorf("couldn't parse URL in getRepositoryURL: %t", err)
		}

		token, err := authToken.getToken()

		if err != nil {
			return "", fmt.Errorf("cannot get auth token: %t", err)
		}

		iid := authToken.getInstallationID()

		u.User = url.UserPassword(strconv.Itoa(iid), token)

		return u.String(), nil
	}

	return cu, nil
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

	var cloneURL string
	var cloneErr error

	switch pushEvent.SCM {
	case GitHub:

		cloneURL, cloneErr = GitHubCloneURL(pushEvent)
		if cloneErr != nil {
			return "", fmt.Errorf("error while creating GitLab CloneURL: %s", cloneErr.Error())
		}

	case GitLab:

		if pushEvent.Repository.Private {
			cloneURL, cloneErr = GitLabCloneURL(pushEvent)
			if cloneErr != nil {
				return "", fmt.Errorf("error while creating GitLab CloneURL: %s", cloneErr.Error())
			}
		} else {
			cloneURL = pushEvent.Repository.CloneURL
		}
	}

	git := exec.Command("git", "clone", cloneURL)
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

func deploy(tars []tarEntry, pushEvent sdk.PushEvent, stack *stack.Services, status *sdk.Status, payloadSecret string) error {

	failedFunctions := []string{}
	owner := pushEvent.Repository.Owner.Login
	repoName := pushEvent.Repository.Name
	url := pushEvent.Repository.CloneURL
	afterCommitID := pushEvent.AfterCommitID
	installationID := pushEvent.Installation.ID
	sourceManagement := pushEvent.SCM
	privateRepo := pushEvent.Repository.Private
	repositoryURL := pushEvent.Repository.RepositoryURL

	ownerID := pushEvent.Repository.Owner.ID

	c := http.Client{}
	gatewayURL := os.Getenv("gateway_url")

	for _, tarEntry := range tars {
		fmt.Println("Deploying service - " + tarEntry.functionName)
		status.AddStatus(sdk.StatusPending, fmt.Sprintf("%s function build started", tarEntry.functionName),
			sdk.BuildFunctionContext(tarEntry.functionName))
		statusErr := reportStatus(status, pushEvent.SCM)
		if statusErr != nil {
			log.Printf(statusErr.Error())
		}

		fileOpen, err := os.Open(tarEntry.fileName)

		if err != nil {
			return err
		}

		defer fileOpen.Close()

		fileInfo, statErr := fileOpen.Stat()
		if statErr == nil {
			msg := fmt.Sprintf("Building %s. Tar: %s",
				tarEntry.functionName,
				bytefmt.ByteSize(uint64(fileInfo.Size())))

			log.Printf("%s\n", msg)

			auditEvent := sdk.AuditEvent{
				Message: msg,
				Owner:   pushEvent.Repository.Owner.Login,
				Repo:    pushEvent.Repository.Name,
				Source:  Source,
			}
			sdk.PostAudit(auditEvent)
		}

		tarFileBytes, tarReadErr := ioutil.ReadAll(fileOpen)
		if tarReadErr != nil {
			return tarReadErr
		}

		digest := hmac.Sign(tarFileBytes, []byte(payloadSecret))

		postBodyReader := bytes.NewReader(tarFileBytes)

		httpReq, _ := http.NewRequest(http.MethodPost, gatewayURL+"function/buildshiprun", postBodyReader)

		httpReq.Header.Add(sdk.CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))

		httpReq.Header.Add("Repo", repoName)
		httpReq.Header.Add("Owner", owner)
		httpReq.Header.Add("Url", url)
		httpReq.Header.Add("Installation_id", fmt.Sprintf("%d", installationID))
		httpReq.Header.Add("Service", tarEntry.functionName)
		httpReq.Header.Add("Image", tarEntry.imageName)
		httpReq.Header.Add("Sha", afterCommitID)
		httpReq.Header.Add("Scm", sourceManagement)
		httpReq.Header.Add("Private", strconv.FormatBool(privateRepo))
		httpReq.Header.Add("Repo-URL", repositoryURL)
		httpReq.Header.Add("Owner-ID", fmt.Sprintf("%d,", ownerID))

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

		// Marshal user labels
		if stack.Functions[tarEntry.functionName].Labels != nil {
			jsonBytes, marshalErr := json.Marshal(stack.Functions[tarEntry.functionName].Labels)
			if marshalErr != nil {
				log.Printf("Error marshaling labels for function %s, %s", tarEntry.functionName, marshalErr)
			}

			httpReq.Header.Add("Labels", string(jsonBytes))
		}

		// Marshal annotations
		if stack.Functions[tarEntry.functionName].Annotations != nil {
			jsonBytes, marshalErr := json.Marshal(stack.Functions[tarEntry.functionName].Annotations)
			if marshalErr != nil {
				log.Printf("Error marshaling annotations for function %s, %s", tarEntry.functionName, marshalErr)
			}

			httpReq.Header.Add("Annotations", string(jsonBytes))
		}

		res, reqErr := c.Do(httpReq)
		if reqErr != nil {
			failedFunctions = append(failedFunctions, tarEntry.functionName)
			fmt.Fprintf(os.Stderr, fmt.Errorf("unable to deploy function via buildshiprun: %s", reqErr.Error()).Error())
			continue
		}

		if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
			failedFunctions = append(failedFunctions, tarEntry.functionName)
			fmt.Fprintf(os.Stderr, fmt.Errorf("unable to deploy function via buildshiprun: invalid status code %d for %s",
				res.StatusCode, tarEntry.functionName).Error())
		} else {
			fmt.Println("Service deployed ", tarEntry.functionName, res.Status, owner)
		}
	}

	if len(failedFunctions) > 0 {
		return fmt.Errorf("%s failed to be deployed via buildshiprun", strings.Join(failedFunctions, ","))
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

	bytesOut, err := ioutil.ReadFile(secretPath)

	if err != nil {
		return fmt.Errorf("unable to read secret: %s", secretPath)
	}

	payloadSecret, secretErr := sdk.ReadSecret("payload-secret")
	if secretErr != nil {
		return secretErr
	}

	c := http.Client{}

	reader := bytes.NewReader(bytesOut)
	httpReq, _ := http.NewRequest(http.MethodPost, gatewayURL+"function/import-secrets", reader)

	httpReq.Header.Add("Owner", owner)

	digest := hmac.Sign(bytesOut, []byte(payloadSecret))
	httpReq.Header.Add(sdk.CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))

	res, reqErr := c.Do(httpReq)

	if reqErr != nil {
		fmt.Fprintf(os.Stderr, fmt.Errorf("error reaching import-secrets function: %s", reqErr.Error()).Error())
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusAccepted && res.StatusCode != http.StatusOK {
		if res.Body != nil {
			resBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return fmt.Errorf("error reading response from import-secrets: %s", err.Error())
			}

			return fmt.Errorf("import-secrets returned unexpected status: %s", string(resBytes))

		}
		return fmt.Errorf("import-secrets returned unknown error, status: %d", res.StatusCode)
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

func formatTemplateRepos() []string {
	templateRepos := []string{"https://github.com/openfaas/templates"}

	if envTemplates := os.Getenv("custom_templates"); len(envTemplates) > 0 {
		customTemplates := strings.Split(envTemplates, ",")
		for _, repo := range customTemplates {
			if strings.Contains(envTemplates, "https://") || strings.Contains(envTemplates, "http://") {
				repo = strings.Trim(repo, " ")
				templateRepos = append(templateRepos, repo)
			} else {
				fmt.Println("Non-valid template URL is configured in custom_templates: ", repo)
			}
		}
	}
	return templateRepos
}

func reportStatus(status *sdk.Status, SCM string) error {
	if SCM == GitHub {
		reportGitHubStatus(status)
	} else if SCM == GitLab {
		reportGitLabStatus(status)
	} else {
		return fmt.Errorf("non-supported SCM: %s", SCM)
	}
	return nil
}

func reportGitLabStatus(status *sdk.Status) {

	payloadSecret, secretErr := sdk.ReadSecret("payload-secret")
	if secretErr != nil {
		log.Printf("unexpected error while reading secret: %s", secretErr)
	}

	suffix := os.Getenv("dns_suffix")
	gatewayURL := os.Getenv("gateway_url")
	gatewayURL = sdk.CreateServiceURL(gatewayURL, suffix)

	statusBytes, marshalErr := json.Marshal(status)
	if marshalErr != nil {
		log.Printf("error while marshalling request: %s", marshalErr.Error())
	}

	statusReader := bytes.NewReader(statusBytes)
	req, reqErr := http.NewRequest(http.MethodPost, gatewayURL+"function/gitlab-status", statusReader)
	if reqErr != nil {
		log.Printf("error while making request to gitlab-status: `%s`", reqErr.Error())
	}

	digest := hmac.Sign(statusBytes, []byte(payloadSecret))
	req.Header.Add(sdk.CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))

	client := http.Client{}

	res, resErr := client.Do(req)
	if resErr != nil {
		log.Printf("unexpected error while retrieving response: %s", resErr.Error())
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if res.StatusCode != http.StatusAccepted {
		log.Printf("unexpected status code: %d", res.StatusCode)
	}

	_, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		log.Printf("unexpected error while reading response body: %s", bodyErr.Error())
	}
	status.CommitStatuses = make(map[string]sdk.CommitStatus)
}

func reportGitHubStatus(status *sdk.Status) {

	if !enableStatusReporting() {
		return
	}

	hmacKey, keyErr := getPayloadSecret()

	if keyErr != nil {
		log.Printf("failed to load hmac key for status, error " + keyErr.Error())

		return
	}

	gatewayURL := os.Getenv("gateway_url")

	_, reportErr := status.Report(gatewayURL, hmacKey)
	if reportErr != nil {
		log.Printf("failed to report status, error: %s", reportErr.Error())
	}
}

func formatGitLabCloneURL(pushEvent sdk.PushEvent, tokenAPI string) (string, error) {
	url, urlErr := url.Parse(pushEvent.Repository.CloneURL)
	if urlErr != nil {
		return "", fmt.Errorf("error while parsing URL: %s", urlErr.Error())
	}
	return fmt.Sprintf("https://%s:%s@%s%s", pushEvent.Repository.Owner.Login, tokenAPI, url.Host, url.Path), nil
}
func GitLabCloneURL(pushEvent sdk.PushEvent) (string, error) {
	tokenAPI, tokenErr := sdk.ReadSecret("gitlab-api-token")
	if tokenErr != nil {
		return "", fmt.Errorf("cannot read api token from GitLab in secret `gitlab-api-token`: %s", tokenErr.Error())
	}

	cloneURL, formatErr := formatGitLabCloneURL(pushEvent, tokenAPI)
	if formatErr != nil {
		return "", fmt.Errorf("error while formatting clone URL for GitLab: %s", formatErr.Error())
	}

	return cloneURL, nil
}

func GitHubCloneURL(pushEvent sdk.PushEvent) (string, error) {
	at := &githubAuthToken{
		appID:          os.Getenv("github_app_id"),
		installationID: pushEvent.Installation.ID,
		privateKeyPath: sdk.GetPrivateKeyPath(),
	}

	cloneURL, err := getRepositoryURL(pushEvent, at)
	if err != nil {
		return "", fmt.Errorf("cannot get repository url to clone: %t", err)
	}

	return cloneURL, nil
}

func checkCompatibleTemplates(stack *stack.Services, filePath string) (templatesErr error) {
	templates, ioErr := existingTemplates(filePath)
	if templatesErr != nil {
		templatesErr = ioErr
	}
	for functionName, function := range stack.Functions {
		for templateIndex, template := range templates {
			if template == function.Language {
				break
			} else if templateIndex == len(templates)-1 && template != function.Language {
				templatesErr = fmt.Errorf("Not supported language: `%s` for function: `%s`",
					function.Language,
					functionName,
				)
			}
		}
	}
	return templatesErr
}

func existingTemplates(filePath string) ([]string, error) {
	var existingTemplates []string
	templatePath := fmt.Sprintf("%s/template", filePath)
	files, ioErr := ioutil.ReadDir(templatePath)
	if ioErr != nil {
		return nil, fmt.Errorf("error while reading tempates directory: %s", ioErr)
	}
	for _, templateFolder := range files {
		if templateFolder.IsDir() {
			existingTemplates = append(existingTemplates, templateFolder.Name())
		}
	}
	return existingTemplates, nil
}
