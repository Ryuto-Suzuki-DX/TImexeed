package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
 * 一般ユーザー権限チェック
 * AuthMiddlewareでセットされたroleがUSERか確認する
 */
func UserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   true,
				"message": "認証情報がありません",
			})
			c.Abort()
			return
		}

		if role != "USER" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   true,
				"message": "一般ユーザー権限がありません",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
