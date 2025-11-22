package handlers

import (
	apigen "avito-test-applicant/internal/api/gen"
)

func makeAPIError(code string, message string) apigen.ErrorResponse {
	return apigen.ErrorResponse{
		Error: &apigen.ErrorResponse{
			Code:    code,
			Message: message,
		},
	}
}
