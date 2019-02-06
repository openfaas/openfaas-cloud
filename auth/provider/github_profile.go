package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
