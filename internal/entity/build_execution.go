package entity

import "time"

type BuildExecution struct {
	Id                     int
	BuildDefinitionEntryId int
	ManuallyRunBy          bool
	ActionLog              string
	Result                 string
	ArtifactPath           string
	ExecutionTime          float64
	ExecutedAt             time.Time
}
