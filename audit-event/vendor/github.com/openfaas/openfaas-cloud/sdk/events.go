package sdk

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

// Event info to pass/store events across functions
type Event struct {
	Service        string            `json:"service"`
	Owner          string            `json:"owner"`
	Repository     string            `json:"repository"`
	Image          string            `json:"image"`
	SHA            string            `json:"sha"`
	URL            string            `json:"url"`
	InstallationID int               `json:"installationID"`
	Environment    map[string]string `json:"environment"`
	Secrets        []string          `json:"secrets"`
}

// BuildEventFromPushEvent function to build Event from PushEvent
func BuildEventFromPushEvent(pushEvent PushEvent) *Event {
	info := Event{}

	info.Service = pushEvent.Repository.Name
	info.Owner = pushEvent.Repository.Owner.Login
	info.Repository = pushEvent.Repository.Name
	info.SHA = pushEvent.AfterCommitID
	info.URL = pushEvent.Repository.CloneURL
	info.InstallationID = pushEvent.Installation.ID

	return &info
}
