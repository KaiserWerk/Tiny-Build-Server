package entity

type BuildDefinitionContent struct {
	ProjectType string `yaml:"project_type"`
	Repository  struct {
		Hoster       string `yaml:"hoster"`
		HosterURL    string `yaml:"hoster_url"`
		Name         string `yaml:"name"`
		AccessUser   string `yaml:"access_user"`
		AccessSecret string `yaml:"access_secret"`
		Branch       string `yaml:"branch"`
	} `yaml:"repository"`
	Setup       []string `yaml:"setup"`
	Test        []string `yaml:"test"`
	PreBuild    []string `yaml:"pre_build"`
	Build       []string `yaml:"build"`
	PostBuild   []string `yaml:"post_build"`
	Deployments []struct {
		Host                string   `yaml:"host"`
		Port                int      `yaml:"port"`
		ConnectionType      string   `yaml:"connection_type"`
		Username            string   `yaml:"username"`
		Password            string   `yaml:"password"`
		WorkingDirectory    string   `yaml:"working_directory"`
		PreDeploymentSteps  []string `yaml:"pre_deployment_steps"`
		PostDeploymentSteps []string `yaml:"post_deployment_steps"`
	} `yaml:"deployments"`
}
