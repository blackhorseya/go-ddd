package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/blackhorseya/go-ddd/internal/shared/adapter/http/response"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

// AccessTokenValidator validates an access token and returns the credential ID.
// This is a minimal interface — any BC providing JWT validation can satisfy it.
type AccessTokenValidator interface {
	ValidateAccessToken(c context.Context, token string) (credentialID string, err error)
}

// Auth returns a middleware that validates JWT access tokens.
// On success it injects the credential ID into context via contextx.WithUserID.
func Auth(validator AccessTokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			response.Unauthorized(c, "invalid authorization format")
			c.Abort()
			return
		}

		credentialID, err := validator.ValidateAccessToken(c.Request.Context(), token)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		ctx := contextx.WithUserID(c.Request.Context(), credentialID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
