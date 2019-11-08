package sdk

// PipelineLog stores a log output from a given stage of
// a pipeline such as the container builder
type PipelineLog struct {
	RepoPath  string
	CommitSHA string
	Function  string
	Source    string
	Data      string
}
