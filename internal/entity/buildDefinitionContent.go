package entity

type BuildDefinitionContent struct {
	ProjectType string `yaml:"project_type"`
	Repository  struct {
		Hoster       string `yaml:"hoster"`
		Url          string `yaml:"hoster_url"`
		Name         string `yaml:"name"`
		AccessUser   string `yaml:"access_user"`
		AccessSecret string `yaml:"access_secret"`
		Branch       string `yaml:"branch"`
	} `yaml:"repository"`
	Setup       []string `yaml:"setup,omitempty"`
	Test        []string `yaml:"test,omitempty"`
	PreBuild    []string `yaml:"pre_build,omitempty"`
	Build       []string `yaml:"build"`
	PostBuild   []string `yaml:"post_build,omitempty"`
	Deployments struct {
		LocalDeployments []struct {
			Enabled bool   `yaml:"enabled"`
			Path string `yaml:"path"`
		} `yaml:"local_deployments,omitempty"`
		EmailDeployments []struct {
			Enabled bool   `yaml:"enabled"`
			Address string `yaml:"address"`
		} `yaml:"email_deployments,omitempty"`
		RemoteDeployments []struct {
			Enabled             bool     `yaml:"enabled"`
			Host                string   `yaml:"host"`
			Port                int      `yaml:"port"`
			ConnectionType      string   `yaml:"connection_type"`
			Username            string   `yaml:"username"`
			Password            string   `yaml:"password"`
			WorkingDirectory    string   `yaml:"working_directory"`
			PreDeploymentSteps  []string `yaml:"pre_deployment_steps"`
			PostDeploymentSteps []string `yaml:"post_deployment_steps"`
		} `yaml:"remote_deployments,omitempty"`
	} `yaml:"deployments"`
}
