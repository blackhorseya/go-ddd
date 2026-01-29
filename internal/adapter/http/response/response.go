// Package response provides a unified response format for RESTful APIs.
// It ensures consistent response structure across all API endpoints.
package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

// Response represents a unified API response structure.
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    Meta   `json:"meta"`
}

// Meta contains metadata about the response.
type Meta struct {
	TraceID    string      `json:"trace_id,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination contains pagination information for list responses.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Error represents an error response.
type Error struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// FieldError represents a validation error for a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// newMeta creates a new Meta with trace ID from context.
func newMeta(c *gin.Context) Meta {
	traceID := contextx.GetTraceID(c.Request.Context())
	return Meta{
		TraceID:   traceID,
		Timestamp: time.Now().UTC(),
	}
}

// OK sends a successful response with data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    newMeta(c),
	})
}

// Created sends a 201 Created response with data.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
		Meta:    newMeta(c),
	})
}

// NoContent sends a 204 No Content response.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// List sends a successful response with paginated data.
func List(c *gin.Context, data any, page, pageSize, total int) {
	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	meta := newMeta(c)
	meta.Pagination = &Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Err sends an error response with the given HTTP status code.
func Err(c *gin.Context, status int, code, message string) {
	c.JSON(status, Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
		},
		Meta: newMeta(c),
	})
}

// ErrWithDetails sends an error response with field-level details.
func ErrWithDetails(c *gin.Context, status int, code, message string, details []FieldError) {
	c.JSON(status, Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: newMeta(c),
	})
}
