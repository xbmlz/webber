package container

import (
	"github.com/xbmlz/webber/config"
	"github.com/xbmlz/webber/log"
)

type Container struct {
	log.Logger
}

func New(cfg config.Config) *Container {
	if cfg == nil {
		return &Container{}
	}

	c := &Container{}

	c.init(cfg)
	return c
}

func (c *Container) init(cfg config.Config) {
	// TODO: Add initialization code here
	c.Logger = log.NewWithConfg(cfg)
}
