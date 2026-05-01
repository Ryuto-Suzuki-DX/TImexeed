package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者権限チェック
 * AuthMiddlewareでセットされたroleがADMINか確認する
 */
func AdminMiddleware() gin.HandlerFunc {
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

		if role != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   true,
				"message": "管理者権限がありません",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
