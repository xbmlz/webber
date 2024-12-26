# Webber

Webber is a fast web framework for Golang.

## Features

- [Dotenv Configuration]() - Load environment variables from .env file
- [Gin Web Framework]() - Fast and minimal web framework
- [GORM ORM]() - The fantastic ORM library for Golang
- [Redis]() - Redis client for Golang
- [Cron Job]() - Run cron job in Golang
- [Zap Logger]() - A high performance logging library for Golang

## Usage

```go
package main

import (
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/xbmlz/webber"
)

type User struct {
	ID   int    `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Name string `json:"name"`
}

func main() {
	// create app
	app := webber.New()

	// get config
	env := app.Config.GetString("APP_ENV", "dev")

	// migrate table
	app.MigrateDB(&User{})

	// seed data
	app.SeedDB(&User{ID: 1, Name: "John"})

	// routes
	app.AddStaticFiles("/static", "./static")

	app.Get("/ping", func(c *webber.Context) {
		// log
		c.Logger.Infof("App env: %s", env)

		user := User{}
		// get user from redis
		val, err := c.Redis.Get(c.Context, "user").Result()
		if err != nil && errors.Is(err, redis.Nil) {
			// get user from db
			c.DB.First(&user, 1)
			c.Logger.Infof("Got user from db: %v", user)
			// cache to redis
			userJson, _ := json.Marshal(user)
			c.Redis.Set(c.Context, "user", string(userJson), 0)
		} else {
			// unmarshal json
			json.Unmarshal([]byte(val), &user)
			c.JSON(200, map[string]interface{}{
				"user": user,
				"from": "cache",
			})
			return
		}

		// response json
		c.JSON(200, map[string]interface{}{
			"user": user,
			"from": "db",
		})
	})

	// cron job
	app.AddCronJob("@every 1s", func(c *webber.Context) {
		c.Logger.Info("Cron job running")
	})

	// run app on port 8080
	app.Run()
}

```

