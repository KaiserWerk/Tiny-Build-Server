package entity

import "time"

type BuildExecution struct {
	Id                int
	BuildDefinitionId int
	ManuallyRunBy     int
	ActionLog         string
	Result            string
	ArtifactPath      string
	ExecutionTime     float64
	ExecutedAt        time.Time
}
