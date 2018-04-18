package sdk

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
}
