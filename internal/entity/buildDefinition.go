package entity

import (
	"database/sql"
	"gorm.io/gorm"
)

// BuildDefinition defines a build definition, wherein the Content fields
// contains the actual YAML string
type BuildDefinition struct {
	gorm.Model
	Caption         string
	Token           string
	Content         string
	EditedBy        uint
	EditedAt        sql.NullTime
	CreatedBy       uint
	BuildExecutions []BuildExecution //`gorm:"foreignKey:build_definition_id"`
	Deleted         bool             `gorm:"notNull"`
}
