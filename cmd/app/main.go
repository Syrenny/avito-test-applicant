package main

import (
	"avito-test-applicant/config"
	"avito-test-applicant/internal/app"

	log "github.com/sirupsen/logrus"
)

const configPath = "config/config.yaml"

func main() {
	// Configuration
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.RunMigrations(cfg)
	app.Run(cfg)
}
