package webber

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xbmlz/webber/config"
	"github.com/xbmlz/webber/container"
	"github.com/xbmlz/webber/log"
)

// App is the main struct for the webber package
type App struct {
	// Config can be used to get configuration values from the environment
	Config config.Config

	container      *container.Container
	httpServer     *httpServer
	httpRegistered bool
}

func New() *App {
	app := &App{}
	app.loadConfig()
	app.container = container.New(app.Config)

	env := app.Config.GetString("APP_ENV", "dev")
	// HTTP Server
	host := app.Config.GetString("HTTP_HOST", "localhost")
	port, _ := app.Config.GetInt("HTTP_PORT", 8080)
	app.httpServer = newHTTPServer(app.container, host, port, env)
	app.httpServer.certFile = app.Config.GetString("CERT_FILE", "")
	app.httpServer.keyFile = app.Config.GetString("KEY_FILE", "")

	return app
}

func (a *App) loadConfig() {
	var configPath string
	if _, err := os.Stat("./configs"); err == nil {
		configPath = "./configs"
	}

	a.Config = config.New(configPath, log.New(log.LevelInfo))
}

func (a *App) Run() {
	// Create a context that is canceled on receiving termination signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Goroutine to handle shutdown when context is canceled
	go func() {
		<-ctx.Done()

		// Create a shutdown context with a timeout
		shutdownCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer done()

		_ = a.Shutdown(shutdownCtx)
	}()

	wg := sync.WaitGroup{}

	if a.httpRegistered {
		wg.Add(1)

		go func(s *httpServer) {
			defer wg.Done()
			s.Run(a.container)
		}(a.httpServer)
	}

	wg.Wait()
}

func (a *App) Shutdown(ctx context.Context) error {
	var err error
	if a.httpServer != nil {
		err = errors.Join(err, a.httpServer.Shutdown(ctx))
	}
	return err
}

func (a *App) addRoute(method, path string, handler HandlerFunc) {
	a.httpRegistered = true

	a.httpServer.router.Handle(method, path, func(ctx *gin.Context) {
		handler(&Context{
			Container: a.container,
			Context:   ctx,
		})
	})
}

func (a *App) Get(path string, handler HandlerFunc) {
	a.addRoute(http.MethodGet, path, handler)
}

func (a *App) Post(path string, handler HandlerFunc) {
	a.addRoute(http.MethodPost, path, handler)
}

func (a *App) Put(path string, handler HandlerFunc) {
	a.addRoute(http.MethodPut, path, handler)
}

func (a *App) Patch(path string, handler HandlerFunc) {
	a.addRoute(http.MethodPatch, path, handler)
}

func (a *App) Delete(path string, handler HandlerFunc) {
	a.addRoute(http.MethodDelete, path, handler)
}

func (a *App) Logger() log.Logger {
	return a.container.Logger
}

func (a *App) Use(middleware ...gin.HandlerFunc) {
	a.httpServer.router.Use(middleware...)
}

func (a *App) MigrateDB(values ...interface{}) error {
	return a.container.DB.AutoMigrate(values...)
}

func (a *App) SeedDB(values ...interface{}) error {
	for _, value := range values {
		if err := a.container.DB.FirstOrCreate(value).Error; err != nil {
			return err
		}
	}
	return nil
}
