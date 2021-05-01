package entity

import (
	"database/sql"
)

// UserAction specific actions a user can execute against his own account
// without being logged in
type UserAction struct {
	Id       int
	UserId   int
	Purpose  string
	Token    string
	Validity sql.NullTime
}
