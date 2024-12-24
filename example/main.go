package main

import (
	"github.com/xbmlz/webber"
)

func main() {
	app := webber.New()

	app.Get("/ping", func(c *webber.Context) {
		// get config
		env := app.Config.GetString("APP_ENV", "dev")
		// log
		c.Logger.Infof("App env: %s", env)

		// response json
		c.JSON(200, map[string]string{
			"app_env": env,
		})
	})

	app.Run()
}
