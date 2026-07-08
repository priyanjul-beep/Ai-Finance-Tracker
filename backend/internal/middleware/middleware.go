// Package middleware contains Gin middleware used across all routes.
package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/pkg/auth"
)

// ─── Auth ──────────────────────────────────────────────────────────────────────

// Auth validates the Bearer JWT and injects "userID" into the Gin context.
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code: "UNAUTHORIZED", Message: "authorization header required",
			})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code: "UNAUTHORIZED", Message: "invalid authorization format",
			})
			return
		}

		claims, err := auth.VerifyToken(parts[1], jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code: "UNAUTHORIZED", Message: "invalid or expired token",
			})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Next()
	}
}

// ─── CORS ──────────────────────────────────────────────────────────────────────

// CORS sets permissive cross-origin headers for the configured origins.
func CORS(allowedOrigins string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowedOrigins)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers",
			"Origin,Content-Type,Authorization,Accept,X-Requested-With,Cache-Control")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// ─── Security Headers ──────────────────────────────────────────────────────────

// SecurityHeaders adds common security headers (Helmet equivalent).
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=()")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

// ─── Rate Limit ────────────────────────────────────────────────────────────────

// RateLimitFunc is the function signature for a rate-limiter.
type RateLimitFunc func(ctx *gin.Context, key string, limit int, window time.Duration) (bool, error)

// RateLimit applies per-IP or per-user rate limiting.
// limiterFn is injected so the Redis implementation can be swapped out.
func RateLimit(limit int, window time.Duration, limiterFn RateLimitFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("rate_limit:%s", c.ClientIP())
		if uid, ok := c.Get("userID"); ok {
			key = fmt.Sprintf("rate_limit:%s", uid)
		}

		exceeded, err := limiterFn(c, key, limit, window)
		if err != nil || exceeded {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Code:    "RATE_LIMIT_EXCEEDED",
				Message: fmt.Sprintf("too many requests, limit is %d per %s", limit, window),
			})
			return
		}
		c.Next()
	}
}

// ─── Panic Recovery ────────────────────────────────────────────────────────────

// Recovery catches panics and returns 500 instead of crashing the server.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
					Code:    "INTERNAL_ERROR",
					Message: "an unexpected error occurred",
				})
			}
		}()
		c.Next()
	}
}

// ─── Request Logger ────────────────────────────────────────────────────────────

// RequestLogger logs every request with latency and status.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		// In production, replace with structured logger
		fmt.Printf("[%s] %d %s %s %s\n",
			time.Now().Format(time.RFC3339),
			c.Writer.Status(),
			c.Request.Method,
			c.Request.URL.Path,
			latency,
		)
	}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// GetUserID extracts the authenticated user's ID from the Gin context.
// Returns ("", false) if the middleware hasn't run.
func GetUserID(c *gin.Context) (string, bool) {
	v, ok := c.Get("userID")
	if !ok {
		return "", false
	}
	uid, ok := v.(string)
	return uid, ok
}
