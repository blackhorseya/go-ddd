// Package handler contains HTTP handlers for the Identity bounded context.
//
// Handlers translate HTTP requests into use case calls and format responses
// using the shared response package (internal/shared/adapter/http/response).
//
// Each handler should implement a Register(*gin.Engine) method to set up routes.
// See internal/shared/adapter/http/handler/health.go for the reference pattern.
package handler
