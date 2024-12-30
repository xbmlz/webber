package container

import (
	"github.com/xbmlz/webber/config"
	"github.com/xbmlz/webber/datasource/db"
	"github.com/xbmlz/webber/datasource/redis"
	"github.com/xbmlz/webber/log"
)

type Container struct {
	log.Logger

	Config config.Config

	DB    *db.DB
	Redis *redis.Redis
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
	if c.Logger == nil {
		c.Logger = log.NewWithConfg(cfg)
	}

	c.Config = cfg

	c.Redis = redis.New(cfg, c.Logger)

	c.DB = db.New(cfg, c.Logger)
}
