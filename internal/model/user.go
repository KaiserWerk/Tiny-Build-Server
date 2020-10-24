package model

type user struct {
	Id          int
	Displayname string
	Email       string
	Password    string
	Locked      bool
	Admin       bool
}
