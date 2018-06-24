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
		Id int `json:"id"`
	}
}
