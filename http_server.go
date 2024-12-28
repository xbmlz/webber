package webber

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/pprof"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/xbmlz/webber/container"
)

type httpServer struct {
	host     string
	port     int
	certFile string
	keyFile  string

	router *gin.Engine
	srv    *http.Server
}

func newHTTPServer(c *container.Container, host string, port int, mode string) *httpServer {

	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	pprof.Register(r)

	// TODO: Add default middleware
	r.Use(
		ginzap.Ginzap(c.Logger.GetLogger(), time.DateTime, true),
		ginzap.RecoveryWithZap(c.Logger.GetLogger(), true),
	)

	return &httpServer{
		host:   host,
		port:   port,
		router: r,
	}
}

func (s *httpServer) Run(c *container.Container) {
	if s.srv != nil {
		c.Logger.Warnf("Server already running on %s:%d", s.host, s.port)
	}

	c.Logger.Infof("Starting server on %s:%d", s.host, s.port)

	s.srv = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", s.host, s.port),
		Handler:           s.router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if s.certFile != "" && s.keyFile != "" {
		// check file is exists
		if _, err := os.Stat(s.certFile); os.IsNotExist(err) {
			c.Logger.Errorf("Error loading %s: %v", s.certFile, err)
		}
		if _, err := os.Stat(s.keyFile); os.IsNotExist(err) {
			c.Logger.Errorf("Error loading %s: %v", s.keyFile, err)
		}

		if err := s.srv.ListenAndServeTLS(s.certFile, s.keyFile); err != nil {
			c.Logger.Errorf("Error starting server: %v", err)
		}
		return
	}

	// If no certFile/keyFile is provided, run the HTTP server
	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		c.Logger.Errorf("Error starting server: %v", err)
	}
}

func (s *httpServer) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}

	return ShutdownWithContext(ctx, func(ctx context.Context) error {
		return s.srv.Shutdown(ctx)
	}, func() error {
		if err := s.srv.Close(); err != nil {
			return err
		}

		return nil
	})
}
