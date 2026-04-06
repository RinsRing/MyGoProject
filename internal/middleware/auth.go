package middleware

import (
	"MyGoProject/internal/auth"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "需要登陆才能访问"})
			c.Abort()
			return
		}
		userID, err := auth.ParseToken(authHeader)
		if err != nil {
			c.JSON(401, gin.H{"error": "无效token"})
			c.Abort()
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}
