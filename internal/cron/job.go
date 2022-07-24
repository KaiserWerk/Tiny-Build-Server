package cron

import "sync/atomic"

var counter uint32 = 0

type Job struct {
	id      uint32
	Name    string
	Enabled bool
	Work    func() error
}

func NewJob(name string, enabled bool, work func() error) Job {
	return Job{
		id:      atomic.AddUint32(&counter, 1),
		Name:    name,
		Enabled: enabled,
		Work:    work,
	}
}

func (j Job) ID() uint32 {
	return j.id
}
