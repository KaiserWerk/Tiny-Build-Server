package entity

import (
	"database/sql"
	"time"
)

type BuildDefinition struct {
	Id              int
	Caption         string
	Content         string
	EditedBy        int
	EditedAt        sql.NullTime
	CreatedBy       int
	CreatedAt       time.Time
	BuildExecutions []BuildExecution `gorm:"foreignKey:build_definition_id"`
}
