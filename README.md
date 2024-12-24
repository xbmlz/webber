# Webber

Webber is a fast web framework for Golang.

## Features

- [Dotenv Configuration]() - Load environment variables from .env file
- [Gin Web Framework]() - Fast and minimal web framework
- [GORM ORM]() - The fantastic ORM library for Golang
- [Redis]() - Redis client for Golang

## Usage

```go
package main

import (
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

    app.

	app.Get("/ping", func(c *webber.Context) {

		// log
		c.Logger.Infof("App env: %s", env)

		// create user
		c.DB.Create(&User{Name: "John"})
		// get user
		user := User{}
		c.DB.First(&user, 1)

		// response json
		c.JSON(200, map[string]interface{}{
			"app_env": env,
			"user":    user,
		})
	})

	// run app on port 8080
	app.Run()
}

```

