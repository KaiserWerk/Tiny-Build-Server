package entity

// User defines a user account
type User struct {
	Id          int
	Displayname string
	Email       string
	Password    string
	Locked      bool
	Admin       bool
}
