package model

type buildStep struct {
	Id            int
	BuildTargetId int
	Caption       string
	Command       string
	Enabled       bool
}
