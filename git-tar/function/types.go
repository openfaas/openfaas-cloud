package function

type buildConfig struct {
	Ref       string            `json:"ref"`
	Frontend  string            `json:"frontend,omitempty"`
	BuildArgs map[string]string `json:"buildArgs,omitempty"`
}
