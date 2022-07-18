package entity

import (
	"database/sql"
	"gorm.io/gorm"
)

// UserAction specific actions a user can execute against his own account
// without being logged in
type UserAction struct {
	gorm.Model
	UserId   uint
	Purpose  string
	Token    string
	Validity sql.NullTime
}
