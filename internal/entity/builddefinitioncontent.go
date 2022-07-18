package entity

// BuildDefinitionContent is the counterpart of the YAML string
// of a build definition
type BuildDefinitionContent struct {
	ProjectType string     `yaml:"project_type"`
	Repository  Repository `yaml:"repository"`
	Setup       []string   `yaml:"Setup,omitempty"`
	Test        []string   `yaml:"test,omitempty"`
	PreBuild    []string   `yaml:"pre_build,omitempty"`
	Build       []string   `yaml:"build"`
	PostBuild   []string   `yaml:"post_build,omitempty"`
	Deployments struct {
		LocalDeployments  []LocalDeployment  `yaml:"local_deployments,omitempty"`
		EmailDeployments  []EmailDeployment  `yaml:"email_deployments,omitempty"`
		RemoteDeployments []RemoteDeployment `yaml:"remote_deployments,omitempty"`
	} `yaml:"deployments"`
}

type Repository struct {
	Hoster       string `yaml:"hoster"`
	Url          string `yaml:"hoster_url"`
	Name         string `yaml:"name"`
	AccessUser   string `yaml:"access_user"`
	AccessSecret string `yaml:"access_secret"`
	Branch       string `yaml:"branch"`
}

type LocalDeployment struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type EmailDeployment struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
}

type RemoteDeployment struct {
	Enabled             bool     `yaml:"enabled"`
	Host                string   `yaml:"host"`
	Port                int      `yaml:"port"`
	ConnectionType      string   `yaml:"connection_type"`
	Username            string   `yaml:"username"`
	Password            string   `yaml:"password"`
	WorkingDirectory    string   `yaml:"working_directory"`
	PreDeploymentSteps  []string `yaml:"pre_deployment_steps"`
	PostDeploymentSteps []string `yaml:"post_deployment_steps"`
}

func (bdc *BuildDefinitionContent) GetSteps() []string {
	allSteps := make([]string, 0)
	allSteps = append(allSteps, bdc.Setup...)
	allSteps = append(allSteps, bdc.Test...)
	allSteps = append(allSteps, bdc.PreBuild...)
	allSteps = append(allSteps, bdc.Build...)
	allSteps = append(allSteps, bdc.PostBuild...)

	return allSteps
}
