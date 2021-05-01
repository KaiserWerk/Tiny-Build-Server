package entity

import (
	"database/sql"
	"time"
)

// BuildDefinition defines a build definition, wherein the Content fields
// contains the actual YAML string
type BuildDefinition struct {
	Id              int
	Caption         string
	Token           string
	Content         string
	EditedBy        int
	EditedAt        sql.NullTime
	CreatedBy       int
	CreatedAt       time.Time
	BuildExecutions []BuildExecution `gorm:"foreignKey:build_definition_id"`
}
