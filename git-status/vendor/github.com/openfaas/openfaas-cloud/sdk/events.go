package sdk

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

// PushEvent as received from GitHub
type PushEvent struct {
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

	return &info, err
}

func GetEnv(service string) map[string]string {
	envVars := make(map[string]string)
	envErr := json.Unmarshal([]byte(os.Getenv("Http_Env")), &envVars)

	if envErr != nil {
		log.Printf("Error un-marshaling env-vars for function %s, %s", service, envErr)
		envVars = make(map[string]string)
	}
	return envVars
}
