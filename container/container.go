package container

import (
	"github.com/xbmlz/webber/config"
	"github.com/xbmlz/webber/datasource/db"
	"github.com/xbmlz/webber/log"
)

type Container struct {
	log.Logger

	DB *db.DB
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
	if c.Logger == nil {
		c.Logger = log.NewWithConfg(cfg)
	}

	c.DB = db.New(cfg, c.Logger)
}
