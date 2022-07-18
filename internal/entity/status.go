package entity

type BuildStatus string

const (
	StatusSucceeded          BuildStatus = "succeeded"
	StatusFailed             BuildStatus = "failed"
	StatusRunning            BuildStatus = "running"
	StatusPartiallySucceeded BuildStatus = "partially_succeeded"
	StatusUnknown            BuildStatus = "unknown"
)

func (bs BuildStatus) String() string {
	return string(bs)
}
