package cron

import (
	"context"
	"github.com/sirupsen/logrus"
	"time"
)

type Cron struct {
	fDaily []Job
	f4Hour []Job
	logger *logrus.Entry
	ctx    context.Context
	cf     func()
}

func New(l *logrus.Entry) *Cron {
	c := Cron{
		fDaily: make([]Job, 0),
		f4Hour: make([]Job, 0),
		logger: l,
	}
	c.ctx, c.cf = context.WithCancel(context.Background())

	return &c
}

func (c *Cron) AddDaily(j Job) {
	c.fDaily = append(c.fDaily, j)
}

func (c *Cron) Add4Hourly(j Job) {
	c.f4Hour = append(c.f4Hour, j)
}

func (c *Cron) runDailyJobs() {
	if len(c.fDaily) == 0 {
		c.logger.Trace("no jobs queued")
		return
	}
	for _, j := range c.fDaily {
		go func(job Job, fLogger *logrus.Entry) {
			if !job.Enabled {
				fLogger.Tracef("job '%s' is not Enabled, skipping", job.Name)
				return
			}
			fLogger.Tracef("job '%s' started", job.Name)
			if err := job.Work(); err != nil {
				fLogger.Tracef("job '%s' failed: %s", job.Name, err.Error())
			} else {
				fLogger.Tracef("job '%s' ran successfully", job.Name)
			}
		}(j, c.logger)
	}
}

func (c *Cron) run4HourJobs() {
	if len(c.f4Hour) == 0 {
		c.logger.Trace("no jobs queued")
		return
	}
	for _, j := range c.f4Hour {
		go func(job Job, fLogger *logrus.Entry) {
			if !job.Enabled {
				fLogger.Tracef("job '%s' is not Enabled, skipping", job.Name)
				return
			}
			fLogger.Tracef("job '%s' started", job.Name)
			if err := job.Work(); err != nil {
				fLogger.Tracef("job '%s' failed: %s", job.Name, err.Error())
			} else {
				fLogger.Tracef("job '%s' ran successfully", job.Name)
			}
		}(j, c.logger)
	}
}

func (c *Cron) Run() {
	c.logger.Tracef("starting up %d cronjob(s)", len(c.fDaily)+len(c.f4Hour))
	fDailyTicker := time.NewTicker(24 * time.Hour)
	f4HourTicker := time.NewTicker(4 * time.Hour)

	go c.run(c.ctx, fDailyTicker, f4HourTicker)
}

func (c *Cron) Stop() {
	c.logger.Trace("stopping all cronjobs")
	c.cf()
}

func (c *Cron) run(ctx context.Context, dailyTicker *time.Ticker, fhTicker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-dailyTicker.C:
			go c.runDailyJobs()
		case <-fhTicker.C:
			go c.run4HourJobs()
		}
	}
}
