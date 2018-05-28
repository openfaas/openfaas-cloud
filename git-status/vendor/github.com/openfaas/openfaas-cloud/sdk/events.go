package sdk

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

// PushEvent as received from GitHub
type PushEvent struct {
	Ref        string `json:"ref"`
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
		Owner    struct {
			Login string `json:"login"`
			Email string `json:"email"`
		} `json:"owner"`
	}
	AfterCommitID string `json:"after"`
	Installation  struct {
		ID int `json:"id"`
	}
}

// EventInfo to pass/store events across functions
type Event struct {
	Service        string            `json:"service"`
	Owner          string            `json:"owner"`
	Repository     string            `json:"repository"`
	Image          string            `json:"image"`
	Sha            string            `json:"sha"`
	URL            string            `json:"url"`
	InstallationID int               `json:"installationID"`
	Environment    map[string]string `json:"environment"`
	Secrets        []string          `json:"secrets"`
}

// function to build Event from PushEvent
func BuildEventFromPushEvent(pushEvent PushEvent) *Event {
	info := Event{}

	info.Service = pushEvent.Repository.Name
	info.Owner = pushEvent.Repository.Owner.Login
	info.Repository = pushEvent.Repository.Name
	info.Sha = pushEvent.AfterCommitID
	info.URL = pushEvent.Repository.CloneURL
	info.InstallationID = pushEvent.Installation.ID

	return &info
}

// function to build Event from Environment Variable
func BuildEventFromEnv() (*Event, error) {
	var err error
	info := Event{}

	info.Service = os.Getenv("Http_Service")
	info.Owner = os.Getenv("Http_Owner")
	info.Repository = os.Getenv("Http_Repo")
	info.Sha = os.Getenv("Http_Sha")
	info.URL = os.Getenv("Http_Url")
	info.InstallationID, err = strconv.Atoi(os.Getenv("Http_Installation_id"))
	info.Environment = GetEnv(info.Service)
	info.Secrets = GetSecret(info.Owner, info.Service)

	return &info, err
}

func GetEnv(service string) map[string]string {
	envVars := make(map[string]string)
	envStr := os.Getenv("Http_Env")
	if len(envStr) > 0 {
		envErr := json.Unmarshal([]byte(envStr), &envVars)
		if envErr != nil {
			log.Printf("error un-marshaling env-vars for function %s, %s", service, envErr)
		}
	}
	return envVars
}

func GetSecret(owner, service string) []string {
	secretVars := []string{}
	secretsStr := os.Getenv("Http_Secrets")
	if len(secretsStr) > 0 {
		secretErr := json.Unmarshal([]byte(secretsStr), &secretVars)
		if secretErr != nil {
			log.Println("error un-marshaling env-vars for function %s, %s", service, secretErr)
		}
	}
	for i := 0; i < len(secretVars); i++ {
		secretVars[i] = owner + "-" + secretVars[i]
	}
	return secretVars
}
