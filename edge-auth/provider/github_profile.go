package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// GitHub provider
type GitHub struct {
	Client *http.Client
}

// NewGitHub create a new GitHub API provider
func NewGitHub(c *http.Client) *GitHub {
	return &GitHub{
		Client: c,
	}
}

// GetProfile returns a profile for a user from GitHub
func (gh *GitHub) GetProfile(accessToken string) (*Profile, error) {
	var err error
	var githubProfile GitHubProfile
	profile := &Profile{}

	req, reqErr := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Add("Authorization", "token "+accessToken)
	if reqErr != nil {
		return profile, reqErr
	}

	res, err := gh.Client.Do(req)
	if err != nil {
		return profile, reqErr
	}

	if res.StatusCode != http.StatusOK {
		return profile, fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	if res.Body != nil {
		defer res.Body.Close()

		bytesOut, _ := ioutil.ReadAll(res.Body)

		unmarshalErr := json.Unmarshal(bytesOut, &githubProfile)
		if unmarshalErr != nil {
			return profile, unmarshalErr
		}
	}

	profile.TwoFactor = githubProfile.TwoFactor
	profile.Name = githubProfile.Name
	profile.Email = githubProfile.Email
	profile.CreatedAt = githubProfile.CreatedAt
	profile.Login = githubProfile.Login
	profile.ID = githubProfile.ID

	return profile, err
}

// GitHubProfile represents a GitHub profile
type GitHubProfile struct {
	ID        int       `json:"id"`
	Login     string    `json:"login"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	TwoFactor bool      `json:"two_factor_authentication"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetUserOrganizations using the API "List your organizations"
// https://developer.github.com/v3/orgs/#list-your-organizations
func (gh *GitHub) GetUserOrganizations(accessToken string) (string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/user/orgs")

	req, reqErr := http.NewRequest(http.MethodGet, apiURL, nil)
	if reqErr != nil {
		return "", fmt.Errorf("error while making request to `%s` organizations: %s", apiURL, reqErr.Error())
	}

	req.Header.Add("Authorization", "token "+accessToken)

	resp, respErr := gh.Client.Do(req)
	if respErr != nil {
		return "", fmt.Errorf("error while requesting organizations: %s", respErr.Error())
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status code from request to GitHub organizations: %d", resp.StatusCode)
	}

	body, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		return "", fmt.Errorf("error while reading body from GitHub organizations: %s", bodyErr.Error())
	}

	var orgs []Organization
	unmarshalErr := json.Unmarshal(body, &orgs)
	if unmarshalErr != nil {
		return "", fmt.Errorf("error while un-marshaling organizations: %s, value: %q", unmarshalErr.Error(), body)
	}

	return formatOrganizations(orgs), nil
}

func formatOrganizations(orgs []Organization) string {
	logins := []string{}
	for _, organization := range orgs {
		logins = append(logins, organization.Login)
	}

	return strings.Join(logins, ",")
}

type Organization struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
}
