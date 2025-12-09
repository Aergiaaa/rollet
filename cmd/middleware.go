package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func (app *app) AuthMiddleware() gin.HandlerFunc {
	// Middleware to authenticate requests using JWT tokens
	return func(c *gin.Context) {

		// Extract the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		// Parse and validate the JWT token
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token is required"})
			c.Abort()
			return
		}

		// Parse and validate the JWT token
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			return []byte(app.jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract user information from token claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Extract user ID from token claims
		userId := int(claims["userId"].(float64))
		user, err := app.models.Users.Get(userId)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
