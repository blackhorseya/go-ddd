package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Common error codes for consistent API responses.
// These codes can be used as i18n keys by frontend applications.
const (
	// General errors
	CodeInternalError    = "INTERNAL_ERROR"
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeConflict         = "CONFLICT"
	CodeValidationFailed = "VALIDATION_FAILED"
	CodeTooManyRequests  = "TOO_MANY_REQUESTS"

	// Resource-specific patterns (examples)
	// Use format: {RESOURCE}_{ACTION}_{REASON}
	// e.g., ORDER_CREATE_FAILED, USER_NOT_FOUND
)

// BadRequest sends a 400 Bad Request response.
func BadRequest(c *gin.Context, message string) {
	Err(c, http.StatusBadRequest, CodeBadRequest, message)
}

// ValidationFailed sends a 400 response with validation error details.
func ValidationFailed(c *gin.Context, details []FieldError) {
	ErrWithDetails(c, http.StatusBadRequest, CodeValidationFailed, "validation failed", details)
}

// Unauthorized sends a 401 Unauthorized response.
func Unauthorized(c *gin.Context, message string) {
	Err(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden sends a 403 Forbidden response.
func Forbidden(c *gin.Context, message string) {
	Err(c, http.StatusForbidden, CodeForbidden, message)
}

// NotFound sends a 404 Not Found response.
func NotFound(c *gin.Context, message string) {
	Err(c, http.StatusNotFound, CodeNotFound, message)
}

// Conflict sends a 409 Conflict response.
func Conflict(c *gin.Context, message string) {
	Err(c, http.StatusConflict, CodeConflict, message)
}

// TooManyRequests sends a 429 Too Many Requests response.
func TooManyRequests(c *gin.Context, message string) {
	Err(c, http.StatusTooManyRequests, CodeTooManyRequests, message)
}

// InternalError sends a 500 Internal Server Error response.
func InternalError(c *gin.Context, message string) {
	Err(c, http.StatusInternalServerError, CodeInternalError, message)
}
