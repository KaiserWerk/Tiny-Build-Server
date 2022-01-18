package entity

import (
	"gorm.io/gorm"
	"time"
)

// BuildExecution consists of metadata for a build definition
type BuildExecution struct {
	gorm.Model
	BuildDefinitionID uint
	ManuallyRunBy     uint
	ActionLog         string
	Result            string
	ArtifactPath      string
	ExecutionTime     float64
	ExecutedAt        time.Time
}
