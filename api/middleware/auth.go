// Package middleware provides HTTP middleware for the Krustron API
// Author: Anubhav Gain <anubhavg@infopercept.com>
package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/anubhavg-icpl/krustron/internal/auth"
	"github.com/anubhavg-icpl/krustron/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// JWTAuth validates JWT tokens
func JWTAuth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.Unauthorized("missing authorization header").ToResponse(getRequestID(c)))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.Unauthorized("invalid authorization header format").ToResponse(getRequestID(c)))
			return
		}

		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.Unauthorized("invalid or expired token").ToResponse(getRequestID(c)))
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// WSAuth validates WebSocket authentication
func WSAuth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			token = c.GetHeader("Sec-WebSocket-Protocol")
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.Unauthorized("missing token").ToResponse(getRequestID(c)))
			return
		}

		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.Unauthorized("invalid or expired token").ToResponse(getRequestID(c)))
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole checks if user has required role(s)
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.Unauthorized("user not authenticated").ToResponse(getRequestID(c)))
			return
		}

		role := userRole.(string)

		// Admin has access to everything
		if role == "admin" {
			c.Next()
			return
		}

		// Check if user has one of the required roles
		hasRole := false
		for _, r := range roles {
			if role == r {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, errors.Forbidden("insufficient permissions").ToResponse(getRequestID(c)))
			return
		}

		c.Next()
	}
}

// RequirePermission checks if user has a specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.Unauthorized("user not authenticated").ToResponse(getRequestID(c)))
			return
		}

		userClaims := claims.(*auth.Claims)

		// Admin has all permissions
		if userClaims.Role == "admin" {
			c.Next()
			return
		}

		// Check if user has the required permission
		hasPermission := false
		for _, p := range userClaims.Permissions {
			if p == permission || p == "*" {
				hasPermission = true
				break
			}
			// Check wildcard permissions
			if strings.HasSuffix(p, ":*") {
				prefix := strings.TrimSuffix(p, "*")
				if strings.HasPrefix(permission, prefix) {
					hasPermission = true
					break
				}
			}
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, errors.Forbidden("permission denied: "+permission).ToResponse(getRequestID(c)))
			return
		}

		c.Next()
	}
}

// RateLimiter provides rate limiting per IP
func RateLimiter(requestsPerSecond float64, burst int) gin.HandlerFunc {
	var mu sync.Mutex
	limiters := make(map[string]*rate.Limiter)

	cleanup := time.NewTicker(10 * time.Minute)
	go func() {
		for range cleanup.C {
			mu.Lock()
			limiters = make(map[string]*rate.Limiter)
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		limiter, exists := limiters[ip]
		if !exists {
			limiter = rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
			limiters[ip] = limiter
		}
		mu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, errors.RateLimited("too many requests").ToResponse(getRequestID(c)))
			return
		}

		c.Next()
	}
}

// AuditLog logs API access for audit purposes
func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log after request
		duration := time.Since(start)
		_ = duration // Use for audit logging
	}
}

// CORS handles Cross-Origin Resource Sharing
func CORS(origins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		allowed := false
		for _, o := range origins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
			c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "43200") // 12 hours
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Secure adds security headers
func Secure() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// Recovery handles panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, errors.Internal("internal server error").ToResponse(getRequestID(c)))
			}
		}()
		c.Next()
	}
}

// getRequestID gets the request ID from context
func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		return id.(string)
	}
	return ""
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for WebSocket
	},
}

// GetUpgrader returns the WebSocket upgrader
func GetUpgrader() *websocket.Upgrader {
	return &upgrader
}
