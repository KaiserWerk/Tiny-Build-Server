package entity

import (
	"database/sql"
)

type UserAction struct {
	Id       int
	UserId   int
	Purpose  string
	Token    string
	Validity sql.NullTime
}
