package app

import (
	"avito-test-applicant/config"
	"avito-test-applicant/internal/repo"
	"avito-test-applicant/migrations/httpserver"
	"avito-test-applicant/migrations/postgres"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"

	gv "github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)


func Run(configPath string) {
	// Configuration
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Logger
	SetLogrus(cfg.Log.Level)

	// Repositories
	log.Info("Initializing postgres...")
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.MaxPoolSize))
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - pgdb.NewServices: %w", err))
	}
	defer pg.Close()

	// Repositories
	log.Info("Initializing repositories...")
	repositories := repo.NewRepositories(pg)

	// Services dependencies
	log.Info("Initializing services...")
	deps := service.ServicesDependencies{
		Repos:    repositories,
	}
	services := service.NewServices(deps)

	// Echo handler
	log.Info("Initializing handlers and routes...")
	handler := echo.New()
	// setup handler validator as go-playground/validator
	handler.Validator = &requestValidator{v: gv.New()}
	v1.NewRouter(handler, services)

	// HTTP server
	log.Info("Starting http server...")
	log.Debugf("Server port: %s", cfg.HTTP.Port)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	log.Info("Configuring graceful shutdown...")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Graceful shutdown
	log.Info("Shutting down...")
	err = httpServer.Shutdown()
	if err != nil {
		log.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}

// requestValidator adapts go-playground/validator to echo.Validator
type requestValidator struct{ v *gv.Validate }

func (cv *requestValidator) Validate(i interface{}) error {
	return cv.v.Struct(i)
}
