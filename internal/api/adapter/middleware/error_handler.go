package middleware

import (
	"avito-test-applicant/internal/api/adapter/apperrors"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func NewHTTPErrorHandler(log *logrus.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		// defensive: avoid panic if context already done
		if err == nil {
			return
		}

		// Log error with useful request context
		entry := log.WithFields(logrus.Fields{
			"path":   c.Path(),
			"method": c.Request().Method,
		}).WithError(err)

		entry.Error("request failed")

		// map known application errors -> HTTP responses
		if errors.Is(err, apperrors.ErrInvalidUUID) {
			if !c.Response().Committed {
				_ = c.JSON(http.StatusBadRequest, map[string]any{
					"error": "invalid uuid",
				})
			}
			return
		}

		// if it's an echo HTTPError, preserve code/message
		if httpErr, ok := err.(*echo.HTTPError); ok {
			code := httpErr.Code
			message := httpErr.Message
			if !c.Response().Committed {
				_ = c.JSON(code, map[string]any{
					"error": message,
				})
			}
			return
		}

		// fallback: unexpected internal error -> 500
		if !c.Response().Committed {
			_ = c.JSON(http.StatusInternalServerError, map[string]any{
				"error": "internal server error",
			})
		}
	}
}
