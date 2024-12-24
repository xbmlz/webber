package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xbmlz/webber/config"
	"github.com/xbmlz/webber/datasource"
)

const (
	redisPingTimeout = 5 * time.Second
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	DB       int
}

type Redis struct {
	*redis.Client
	logger datasource.Logger
	config *Config
}

func New(cfg config.Config, logger datasource.Logger) *Redis {
	redisConfig := getConfig(cfg)

	if redisConfig.Host == "" {
		logger.Debugf("skipping redis connection initialization as 'REDIS_HOST' is not provided")
		return nil
	}

	rc := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Username: redisConfig.Username,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	rc.AddHook(&redisHook{config: redisConfig, logger: logger})

	ctx, cancel := context.WithTimeout(context.TODO(), redisPingTimeout)
	defer cancel()

	if err := rc.Ping(ctx).Err(); err == nil {
		logger.Infof("connected to redis at %s:%d", redisConfig.Host, redisConfig.Port)
	} else {
		logger.Errorf("failed to connect to redis at %s:%d: %v", redisConfig.Host, redisConfig.Port, err)
	}

	return &Redis{Client: rc, config: redisConfig, logger: logger}
}

func getConfig(c config.Config) *Config {
	port, _ := c.GetInt("REDIS_PORT", 6379)
	db, _ := c.GetInt("REDIS_DB", 0)

	return &Config{
		Host:     c.GetString("REDIS_HOST", "localhost"),
		Port:     port,
		Username: c.GetString("REDIS_USERNAME", ""),
		Password: c.GetString("REDIS_PASSWORD", ""),
		DB:       db,
	}
}

func (r *Redis) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}
