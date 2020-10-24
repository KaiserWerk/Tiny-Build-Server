package model

import "time"

type buildExecution struct {
	Id                int
	BuildDefinitionId int
	InitiatedBy       int
	ManualRun         bool
	ActionLog         string
	Result            string
	ArtifactPath      string
	ExecutionTime     float64
	ExecutedAt        time.Time
}
