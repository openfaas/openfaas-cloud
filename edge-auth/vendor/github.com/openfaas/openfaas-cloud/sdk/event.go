package sdk

// Event info used to pass events between functions
type Event struct {
	Service        string            `json:"service"`
	Owner          string            `json:"owner"`
	OwnerID        int               `json:"owner-id"`
	Repository     string            `json:"repository"`
	Image          string            `json:"image"`
	SHA            string            `json:"sha"`
	URL            string            `json:"url"`
	InstallationID int               `json:"installationID"`
	Environment    map[string]string `json:"environment"`
	Secrets        []string          `json:"secrets"`
	Private        bool              `json:"private"`
	SCM            string            `json:"scm"`
	RepoURL        string            `json:"repourl"`
	Labels         map[string]string `json:"labels"`
	Annotations    map[string]string `json:"annotations"`
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
