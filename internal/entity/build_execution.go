package entity

import "time"

type BuildExecution struct {
	Id                int
	BuildDefinitionId int
	ManuallyRunBy     bool
	ActionLog         string
	Result            string
	ArtifactPath      string
	ExecutionTime     float64
	ExecutedAt        time.Time
}
