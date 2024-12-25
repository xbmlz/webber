package webber

import (
	"github.com/robfig/cron/v3"
	"github.com/xbmlz/webber/container"
)

type Crontab struct {
	*cron.Cron
	container *container.Container
}

type CronFunc func(ctx *Context)

func NewCron(c *container.Container) *Crontab {
	cron := cron.New()

	return &Crontab{
		container: c,
		Cron:      cron,
	}
}
