package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// GitLabProvider provider
type GitLabProvider struct {
	BaseURL string
	ApiURL  string
	Client  *http.Client
}

// NewGitLabProvider create a new GitLabProvider API provider
func NewGitLabProvider(c *http.Client, baseURL string, apiURL string) *GitLabProvider {
	return &GitLabProvider{
		Client:  c,
		BaseURL: baseURL,
		ApiURL:  apiURL,
	}
}

// GetProfile returns a profile for a user from GitLabProvider
func (gl *GitLabProvider) GetProfile(accessToken string) (*Profile, error) {
	var err error
	var gitlabProfile GitLabProfile

	req, reqErr := http.NewRequest(http.MethodGet, gl.ApiURL+"user", nil)
	req.Header.Add("Authorization", "bearer "+accessToken)

	if reqErr != nil {
		return nil, reqErr
	}

	res, err := gl.Client.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		var errorBytes string
		if res.Body != nil {
			defer res.Body.Close()

			bytesOut, _ := ioutil.ReadAll(res.Body)
			errorBytes = string(bytesOut)
		}

		return nil, fmt.Errorf("bad status code: %d, %s", res.StatusCode, errorBytes)
	}

	if res.Body != nil {
		defer res.Body.Close()

		bytesOut, _ := ioutil.ReadAll(res.Body)

		unmarshalErr := json.Unmarshal(bytesOut, &gitlabProfile)

		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
	}

	return &Profile{
		ID:        gitlabProfile.ID,
		Login:     gitlabProfile.Username,
		CreatedAt: gitlabProfile.CreatedAt,
		Email:     gitlabProfile.Email,
		Name:      gitlabProfile.Name,
		TwoFactor: gitlabProfile.TwoFactor,
	}, err
}

func (gl *GitLabProvider) GetUserProjects(accessToken string) (string, error) {

	apiURL := gl.ApiURL + "groups?min_access_level=10"
	req, reqErr := http.NewRequest(http.MethodGet, apiURL, nil)
	if reqErr != nil {
		return "", fmt.Errorf("error while making request to `%s` projects: %s", apiURL, reqErr.Error())
	}

	req.Header.Add("Authorization", "bearer "+accessToken)
	res, respErr := gl.Client.Do(req)
	if respErr != nil {
		return "", fmt.Errorf("error while requesting projects: %s", respErr.Error())
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status code from request to GitLab projects: %d", res.StatusCode)
	}

	body, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		return "", fmt.Errorf("error while reading body from GitLab projects: %s", bodyErr.Error())
	}

	var orgs []GitLabProject
	unmarshalErr := json.Unmarshal(body, &orgs)
	if unmarshalErr != nil {
		return "", fmt.Errorf("error while un-marshaling projects: %s, value: %q", unmarshalErr.Error(), body)
	}

	return formatGitLabOrganizations(orgs), nil
}

func formatGitLabOrganizations(orgs []GitLabProject) string {
	logins := []string{}
	for _, organization := range orgs {
		logins = append(logins, organization.Path)
	}

	return strings.Join(logins, ",")
}

type GitLabProject struct {
	ID   int    `json:"id"`
	Path string `json:"path"`
}

// GitLabProfile represents a GitLabProvider profile
type GitLabProfile struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	TwoFactor bool      `json:"two_factor_enabled"`
	CreatedAt time.Time `json:"created_at"`
}
