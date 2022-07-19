package entity

import (
	"database/sql"
	"gorm.io/gorm"
)

type (
	// BuildDefinition defines a build definition, wherein the Raw field
	// contains the actual YAML string
	BuildDefinition struct {
		gorm.Model
		Caption         string
		Token           string
		Raw             string
		Data            BuildDefinitionContent `gorm:"-"`
		EditedBy        uint
		EditedAt        sql.NullTime
		CreatedBy       uint
		BuildExecutions []BuildExecution
		Deleted         bool `gorm:"notNull"`
	}
)
