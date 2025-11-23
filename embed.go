package avitotestapplicant

import "embed"

//go:embed docs/swagger/*
//go:embed docs/openapi.yml
var SwaggerFS embed.FS
