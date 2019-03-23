package function

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/github"
	"github.com/matipan/derek/auth"
	handler "github.com/openfaas-incubator/go-function-sdk"
)

const goProgram = `package function

import (
	"net/http"

	handler "github.com/openfaas-incubator/go-function-sdk"
)

// Handle is a function that always returns:  Hello world!
func Handle(req handler.Request) (handler.Response, error) {
	return handler.Response{
		Body:       []byte("%s"),
		StatusCode: http.StatusOK,
	}, nil
}
`

// Handle handles a function invocation.
func Handle(req handler.Request) (res handler.Response, err error) {
	cnf, err := NewConfig()
	if err != nil {
		return res, fmt.Errorf("unable to build config: %s", err)
	}

	// Create an access token for github app installation number
	token, err := auth.MakeAccessTokenForInstallation(cnf.ApplicationID, 694419, cnf.PrivateKey)
	if err != nil {
		return res, fmt.Errorf("unable to obtain access token: %s", err)
	}

	ctx := context.Background()
	client := MakeClient(ctx, token, cnf)
	// client := MakeUserClient()
	owner, repo, goFilePath := os.Getenv("functionOwner"), os.Getenv("functionRepo"), os.Getenv("goFilePath")

	// Download the file and obtain it's sha. If the file does not exist
	// then we should not proceed.
	sha, err := fetchFileSHA(ctx, client, owner, repo, goFilePath)
	if err != nil {
		return res, fmt.Errorf("unable to fetch function's .go file: %s", err)
	}

	now := time.Now()
	content := fmt.Sprintf(goProgram, now.String())
	if err = updateFile(ctx, client, owner, repo, goFilePath, sha, []byte(content)); err != nil {
		return res, fmt.Errorf("unable to update file: %s", err)
	}

	c := &http.Client{
		Timeout: 120 * time.Second,
	}
	var ok bool
	for j := 0; j < 10; j++ {
		fmt.Printf("Attempt: %d\n", j)
		time.Sleep(15 * time.Second)
		b, err := callFunction(c, os.Getenv("functionURL"))
		if err != nil {
			return res, fmt.Errorf("calling function failed: %s", err)
		}
		fmt.Printf("Function responded with: %s - Expected: %s\n", string(b), now.String())
		if string(b) == now.String() {
			ok = true
			break
		}
	}

	return handler.Response{
		Body: []byte(fmt.Sprintf("%v", ok)),
	}, nil
}

func strptr(s string) *string {
	return &s
}

func fetchFileSHA(ctx context.Context, client *github.Client, owner, repo, filepath string) (string, error) {
	file, _, contentRes, err := client.Repositories.GetContents(ctx, owner, repo, filepath, &github.RepositoryContentGetOptions{
		Ref: "master",
	})
	if err != nil {
		return "", err
	}
	fmt.Println("Rate limiting", contentRes.Rate)
	if contentRes.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github responded with a %d", contentRes.StatusCode)
	}
	return *file.SHA, nil
}

func updateFile(ctx context.Context, client *github.Client, owner, repo, filePath, fileSha string, content []byte) error {
	now := time.Now()
	_, contentRes, err := client.Repositories.UpdateFile(ctx, owner, repo, filePath, &github.RepositoryContentFileOptions{
		Message: strptr("Update function file for ofc e2e test"),
		Content: content,
		SHA:     &fileSha,
		Branch:  strptr("master"),
		Committer: &github.CommitAuthor{
			Date:  &now,
			Name:  strptr("OpenFaaS e2e test"),
			Email: strptr("notreal@openfaas.com"),
		},
	})
	if err != nil {
		return err
	}
	fmt.Println("Rate limiting", contentRes.Rate)
	if contentRes.StatusCode != http.StatusOK {
		return fmt.Errorf("github responded with a %d", contentRes.StatusCode)
	}
	return nil
}

func callFunction(c *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to build request for function: %s", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to call function: %s", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %s", err)
	}
	return b, nil
}
