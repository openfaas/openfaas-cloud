package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

// GitLabProfile represents a GitLabProvider profile
type GitLabProfile struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	TwoFactor bool      `json:"two_factor_enabled"`
	CreatedAt time.Time `json:"created_at"`
}
