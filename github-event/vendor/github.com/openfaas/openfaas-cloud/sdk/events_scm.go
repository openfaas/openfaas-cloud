package sdk

// PushEvent is received from GitHub's push event subscription
type PushEvent struct {
	Ref           string `json:"ref"`
	Repository    PushEventRepository
	AfterCommitID string `json:"after"`
	Installation  PushEventInstallation
	SCM           string // SCM field is for internal use and not provided by GitHub
}

// PushEventRepository represents the repository from a push event
type PushEventRepository struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	CloneURL      string `json:"clone_url"`
	Private       bool   `json:"private"`
	ID            int64  `json:"id"`
	RepositoryURL string `json:"url"`

	Owner Owner `json:"owner"`
}

// Owner is the owner of a GitHub repo
type Owner struct {
	Login string `json:"login"`
	Email string `json:"email"`
	ID    int64  `json:"id"`
}

type PushEventInstallation struct {
	ID int `json:"id"`
}

// GitLabPushEvent as received from GitLab's system hook event
type GitLabPushEvent struct {
	Ref              string           `json:"ref"`
	UserUsername     string           `json:"user_username"`
	UserEmail        string           `json:"user_email"`
	GitLabProject    GitLabProject    `json:"project"`
	GitLabRepository GitLabRepository `json:"repository"`
	AfterCommitID    string           `json:"after"`
}

type GitLabProject struct {
	ID                int    `json:"id"`
	Namespace         string `json:"namespace"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"` //would be repo full name
	WebURL            string `json:"web_url"`
	VisibilityLevel   int    `json:"visibility_level"`
}

type GitLabRepository struct {
	CloneURL string `json:"git_http_url"`
}

type Customer struct {
	Sender     Sender              `json:"sender"`
	Repository PushEventRepository `json:"repository"`
}

type Sender struct {
	Login string `json:"login"`
}

type InstallationRepositoriesEvent struct {
	Action       string `json:"action"`
	Installation struct {
		Account struct {
			Login string
		}
	} `json:"installation"`
	RepositoriesRemoved []Installation `json:"repositories_removed"`
	RepositoriesAdded   []Installation `json:"repositories_added"`
	Repositories        []Installation `json:"repositories"`
}

type Installation struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}
