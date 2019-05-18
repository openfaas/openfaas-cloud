package sdk

// BuildResult represents a successful Docker build and
// push operation to a remote registry
type BuildResult struct {
	Log       []string `json:"log"`
	ImageName string   `json:"imageName"`
	Status    string   `json:"status"`
}
