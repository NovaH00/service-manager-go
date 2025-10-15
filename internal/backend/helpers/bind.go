package helpers

import (
	"errors"
	"net/http"
	"service-manager/internal/backend/api"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// BindOrAbort binds and validates the request. If it fails, it writes an
// error JSON response, aborts the request, and returns false.
func BindOrAbort[T any](c *gin.Context) (T, bool) {
	var req T

	// We must pass a pointer to ShouldBindJSON so it can modify 'req'.
	if err := c.ShouldBindJSON(&req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			// Validation error (e.g., missing required field).
			apiError := api.NewError(
				"Invalid request schema",
				"One or many fields of the request is not correct",
			)
			c.JSON(
				http.StatusUnprocessableEntity,
				apiError,
			)
			return req, false
		}

		apiError := api.NewError(
			"Bad request",
			err.Error(),
		)
		c.JSON(
			http.StatusBadRequest,
			apiError,
		)

		// Syntax error (e.g., malformed JSON).
		return req, false
	}

	return req, true
}
