package entity

import "gorm.io/gorm"

// User defines a user account
type User struct {
	gorm.Model
	DisplayName string
	Email       string
	Password    string
	Locked      bool
	Admin       bool
}
