package entity

import "time"

type UserAction struct {
	Id       int
	UserId   int
	Purpose  string
	Token    string
	Validity time.Time
}
