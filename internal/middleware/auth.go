package middleware

import (
	"net/http"
	"strings"

	"github.com/delivery-app/delivery-api/internal/util"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token from Authorization header
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := util.ParseToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminMiddleware checks if user has admin role
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// DriverMiddleware checks if user has driver role
func DriverMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "driver" {
			c.JSON(http.StatusForbidden, gin.H{"error": "driver access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ShipperMiddleware checks if user has shipper role
func ShipperMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "shipper" {
			c.JSON(http.StatusForbidden, gin.H{"error": "shipper access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
