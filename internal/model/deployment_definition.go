package model

type deploymentDefinition struct {
	Id                    int
	BuildDefinitionId     int
	Caption               string
	Host                  string
	Username              string
	Password              string
	ConnectionType        string
	WorkingDirectory      string
	PreDeploymentActions  string
	PostDeploymentActions string
}
