package main

import (
	"avito-test-applicant/internal/app"
)

const configPath = "config/config.yaml"

func main() {
	app.Run(configPath)
}
