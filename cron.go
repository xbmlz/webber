package webber

import (
	"github.com/robfig/cron/v3"
	"github.com/xbmlz/webber/container"
)

type crontab struct {
	*cron.Cron
	container *container.Container
}

type CronFunc func(ctx *Context)

func NewCron(c *container.Container) *crontab {
	cron := cron.New()

	return &crontab{
		container: c,
		Cron:      cron,
	}
}
