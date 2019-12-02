package test

type YamlSpec struct {
	ApiVersion string        `yaml:"apiVersion"`
	Kind       string        `yaml:"kind"`
	Metadata   MetadataItems `yaml:"metadata"`
	Spec       Spec          `yaml:"spec"`
	Subjects   []Subjects `yaml:"subjects"`
	RoleRef map[string]string `yaml:"roleRef"`
	Rules []Rules `yaml:"rules"`
}

type Rules struct {
	ApiGroups []string `yaml:"apiGroups"`
	Resources []string `yaml:"resources"`
	Verbs 	[]string `yaml:"verbs"`
}

type Subjects struct {
	Kind string	`yaml:"kind"`
	Name string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type Spec struct {
	TLS         []TLSSpec         `yaml:"tls,omitempty"`
	Rules       []SpecRules       `yaml:"rules,flow,omitempty"`
	PolicyTypes []string          `yaml:"policyTypes,flow,omitempty"`
	PodSelector map[string]string `yaml:"podSelector,omitempty"`
	Ingress     []NetworkIngress  `yaml:"ingress,omitempty"`
	Replicas	int `yaml:"replicas"`
	Selector	MatchLabelSelector `yaml:"selector,omitempty"`
	Template 	SpecTemplate `yaml:"template"`
	ACME	ACME `yaml:"acme"`
}

type ACME struct {
	Email string `yaml:"email"`
	Server string `yaml:"server"`
	PrivateKeySecretRef map[string]string `yaml:"privateKeySecretRef"`
	Solvers []Solver `yaml:"solvers"`
}

type Solver struct {
	DNSSolver DNSSolver `yaml:"dns01"`
}

type DNSSolver struct {
	Route53 Route53Solver 	`yaml:"route53"`
}

type Route53Solver struct {
	Region string 	`yaml:"region"`
	AccessKeyID	string 	`yaml:"accessKeyID"`
	SecretAccessKeySecretRef map[string]string `yaml:"secretAccessKeySecretRef"`
}

type MetadataItems struct {
	Name        string            `yaml:"name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
}

type TLSSpec struct {
	Hosts      []string `yaml:"hosts"`
	SecretName string   `yaml:"secretName"`
}
type SpecRules struct {
	Host string    `yaml:"host"`
	Http HttpPaths `yaml:"http"`
}

type HttpPaths struct {
	Paths []PathType `yaml:"paths,flow"`
}

type PathType struct {
	Path    string            `yaml:"path"`
	Backend map[string]string `yaml:"backend"`
}

type NetworkIngress struct {
	From []NetworkSelectors `yaml:"from,flow"`
}

type NetworkSelectors struct {
	Namespace NamespaceSelector  `yaml:"namespaceSelector,omitempty"`
	Pod       MatchLabelSelector `yaml:"podSelector,omitempty"`
}

type MatchLabelSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}

type NamespaceSelector struct {
	MatchLabels map[string]string  `yaml:"matchLabels"`
	PodSelector MatchLabelSelector `yaml:"podSelector,omitempty"`
}

type SpecTemplate struct {
	Metadata MetadataItems `yaml:"metadata"`
	Spec 	 TemplateSpec `yaml:"spec"`
}

type TemplateSpec struct {
	Volumes []DeploymentVolumes       `yaml:"volumes"`
	Containers []DeploymentContainers `yaml:"containers"`
}

type DeploymentVolumes struct {
	SecretName string `yaml:"name"`
	SecretInfo map[string]string `yaml:"secret"`
}

type DeploymentContainers struct {
	Name string `yaml:"name"`
	Args []string `yaml:"args"`
	Image string `yaml:"image"`
	ImagePullPolicy string `yaml:"imagePullPolicy"`
	ContainerReadinessProbe ReadinessProbe `yaml:"readinessProbe"`
	ContainerEnvironment []Environment `yaml:"env"`
	Ports []ContainerPort `yaml:"ports"`
	Volumes []ContainerVolume `yaml:"volumeMounts"`
	SecurityContext map[string]string `yaml:"securityContext"`
}

type ContainerVolume struct {
	Name string `yaml:"name"`
	ReadOnly bool `yaml:"readOnly"`
	MountPath string `yaml:"mountPath"`
}

type ContainerPort struct {
	Port int `yaml:"containerPort"`
	Protocol string `yaml:"protocol"`
}

type ReadinessProbe struct {
	HttpGet HttpProbe `yaml:"httpGet"`
	ExecProbe ExecProbe `yaml:"exec"`
	TimeoutSeconds int `yaml:"timeoutSeconds"`
}

type ExecProbe struct {
	Command []string `yaml:"command"`
}

type HttpProbe struct {
	Path string `yaml:"path"`
	Port int `yaml:"port"`
}

type Environment struct {
	Name string `yaml:"name"`
	Value string `yaml:"value"`
}

