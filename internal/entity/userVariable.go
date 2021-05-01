package entity

// UserVariable has a name and a value and is, if available,
// programmatically inserted into the content of a
// build definition
type UserVariable struct {
	Id          int
	UserEntryId int
	Variable    string
	Value       string
	Public      bool
}
