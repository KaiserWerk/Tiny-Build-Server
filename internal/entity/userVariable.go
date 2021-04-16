package entity

type UserVariable struct {
	Id          int
	UserEntryId int
	Variable    string
	Value       string
	Public      bool
}
