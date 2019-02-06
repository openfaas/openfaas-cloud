package provider

import (
	"strings"
	"time"
)

const (
	githubName = "github"
	gitlabName = "gitlab"
)

type Profile struct {
	ID        int       `json:"id"`
	Login     string    `json:"login"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	TwoFactor bool      `json:"two_factor"`
	CreatedAt time.Time `json:"created_at"`
}

type Provider interface {
	GetProfile(accessToken string) (*Profile, error)
}

var supportedProviders = []string{
	githubName,
	gitlabName,
}

func GetSupportedString() string {
	return strings.Join(supportedProviders, ", ")
}

func IsSupported(name string) bool {
	for _, sp := range supportedProviders {
		if strings.EqualFold(name, sp) {
			return true
		}
	}

	return false
}
