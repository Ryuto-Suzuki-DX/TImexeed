package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"timexeed/backend/internal/auth"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"data":    nil,
				"error":   true,
				"code":    "AUTH_HEADER_REQUIRED",
				"message": "認証情報がありません",
				"detail":  nil,
			})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"data":    nil,
				"error":   true,
				"code":    "INVALID_AUTH_HEADER",
				"message": "認証形式が正しくありません",
				"detail":  nil,
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := auth.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"data":    nil,
				"error":   true,
				"code":    "INVALID_TOKEN",
				"message": "トークンが正しくありません",
				"detail":  err.Error(),
			})
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}