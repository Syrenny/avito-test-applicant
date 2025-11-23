package app

import (
	avitotestapplicant "avito-test-applicant"
	"avito-test-applicant/config"
	"avito-test-applicant/internal/api/adapter/handlers"
	"avito-test-applicant/internal/api/adapter/middleware"
	apigen "avito-test-applicant/internal/api/gen"
	"avito-test-applicant/internal/repo"
	"avito-test-applicant/internal/service"
	"avito-test-applicant/pkg/httpserver"
	"avito-test-applicant/pkg/postgres"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"

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
	pg, err := postgres.New(cfg.PG.URL, cfg.PG.MaxPoolSize)
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - pgdb.NewServices: %w", err))
	}
	defer pg.Close()

	log.Info("Initializing transaction manager...")
	trManager := postgres.NewTransactionManager(pg.Pool)

	// Repositories
	log.Info("Initializing repositories...")
	repositories := repo.NewRepositories(pg, trmpgx.DefaultCtxGetter)

	// Services dependencies
	log.Info("Initializing services...")
	deps := service.ServicesDependencies{
		Repos:     repositories,
		TrManager: trManager,
	}
	services := service.NewServices(deps)

	// Echo
	log.Info("Initializing handlers and routes...")
	e := echo.New()
	// setup handler validator as go-playground/validator
	e.Validator = &requestValidator{v: gv.New()}

	// Swagger UI
	staticFS := http.FS(avitotestapplicant.SwaggerFS)
	fileServer := http.FileServer(staticFS)
	e.GET("/*", echo.WrapHandler(http.StripPrefix("/", fileServer)))

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// HTTP error handler
	e.HTTPErrorHandler = middleware.NewHTTPErrorHandler(log.StandardLogger())

	// HTTP server
	serverImpl := handlers.NewServer(services)
	strictServer := apigen.NewStrictHandler(serverImpl, nil)
	apigen.RegisterHandlers(e, strictServer)

	// HTTP server wrapper
	log.Info("Starting http server...")
	log.Debugf("Server port: %s", cfg.HTTP.Port)
	httpServer := httpserver.New(e, httpserver.Port(cfg.HTTP.Port))

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
