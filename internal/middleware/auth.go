package middleware

import (
	"docker-visual/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth returns a Gin middleware that validates the API key from the
// Authorization header. If the configured key is empty, the middleware
// is a no-op (development mode).
func APIKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth if no key is configured (dev mode)
		if apiKey == "" {
			c.Next()
			return
		}

		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
				"code":  "AUTH_MISSING",
			})
			return
		}

		// Support "Bearer <key>" format
		token := strings.TrimPrefix(header, "Bearer ")
		if token != apiKey {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "invalid api key",
				"code":  "AUTH_INVALID",
			})
			return
		}

		c.Next()
	}
}

// JWTAuth returns a Gin middleware that validates JWT tokens.
func JWTAuth(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
				"code":  "AUTH_MISSING",
			})
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(header, "Bearer ")
		if tokenString == header {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization format",
				"code":  "AUTH_INVALID_FORMAT",
			})
			return
		}

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
				"code":  "AUTH_INVALID_TOKEN",
			})
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
