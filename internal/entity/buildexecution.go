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
	Status            BuildStatus
	ArtifactPath      string
	ExecutionTime     float64
	ExecutedAt        time.Time
}

func NewBuildExecution(bdID, userID uint) *BuildExecution {
	return &BuildExecution{
		Model:             gorm.Model{},
		BuildDefinitionID: bdID,
		ManuallyRunBy:     userID,
		ActionLog:         "",
		Status:            StatusRunning,
		ArtifactPath:      "",
		ExecutionTime:     0,
		ExecutedAt:        time.Now(),
	}
}
