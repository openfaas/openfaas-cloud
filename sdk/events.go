package sdk

// PushEvent as received from GitHub
type PushEventInstallation struct {
	ID int `json:"id"`
}

type PushEventRepository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	CloneURL string `json:"clone_url"`
	Private  bool   `json:"private"`

	Owner Owner `json:"owner"`
}

type PushEvent struct {
	Ref           string `json:"ref"`
	Repository    PushEventRepository
	AfterCommitID string `json:"after"`
	Installation  PushEventInstallation
}

// Owner is the owner of a GitHub repo
type Owner struct {
	Login string `json:"login"`
	Email string `json:"email"`
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
	Private        bool              `json:"private"`
}

type Customer struct {
	Sender Sender `json:"sender"`
}

type Sender struct {
	Login string `json:"login"`
}

// BuildEventFromPushEvent function to build Event from PushEvent
func BuildEventFromPushEvent(pushEvent PushEvent) *Event {
	info := Event{}

	info.Service = pushEvent.Repository.Name
	info.Owner = pushEvent.Repository.Owner.Login
	info.Repository = pushEvent.Repository.Name
	info.URL = pushEvent.Repository.CloneURL
	info.Private = pushEvent.Repository.Private

	info.SHA = pushEvent.AfterCommitID
	info.InstallationID = pushEvent.Installation.ID

	return &info
}
