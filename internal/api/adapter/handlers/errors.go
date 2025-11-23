package handlers

import (
	apigen "avito-test-applicant/internal/api/gen"
)

func makeAPIError(code apigen.ErrorResponseErrorCode, message string) apigen.ErrorResponse {
	var err apigen.ErrorResponse
	err.Error.Code = code
	err.Error.Message = message
	return err
}
