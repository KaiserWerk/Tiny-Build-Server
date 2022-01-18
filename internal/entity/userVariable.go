package entity

import "gorm.io/gorm"

// UserVariable has a name and a value and is, if available,
// programmatically inserted into the content of a
// build definition
type UserVariable struct {
	gorm.Model
	UserEntryId uint
	Variable    string
	Value       string
	Public      bool
}
