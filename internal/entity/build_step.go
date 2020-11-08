package entity

type BuildStep struct {
	Id            int
	BuildTargetId int
	Caption       string
	Command       string
	Enabled       bool
}
