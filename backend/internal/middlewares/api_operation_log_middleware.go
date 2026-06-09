package middlewares

import (
	"time"

	"timexeed/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/*
 * API操作ログMiddleware
 *
 * 役割：
 * ・API単位で操作ログをDBへ保存する
 * ・AuthMiddleware後に実行されることで userId / email / role を取得する
 *
 * 注意：
 * ・ログ保存に失敗しても、本体APIは失敗させない
 * ・request body は保存しない
 * ・/health や /db-health などは対象外にする
 */
func ApiOperationLogMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()

		c.Next()

		finishedAt := time.Now()
		durationMs := finishedAt.Sub(startedAt).Milliseconds()

		userID := getUintPointerFromContext(c, "userId")
		email := getStringPointerFromContext(c, "email")
		role := getStringPointerFromContext(c, "role")

		statusCode := c.Writer.Status()

		apiLog := models.ApiOperationLog{
			UserID:     userID,
			Email:      email,
			Role:       role,
			Method:     c.Request.Method,
			Path:       c.FullPath(),
			StatusCode: statusCode,
			ClientIP:   c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			DurationMs: durationMs,
			StartedAt:  startedAt,
			FinishedAt: finishedAt,
		}

		/*
		 * c.FullPath() が空になるケースがあるため、その場合はURLパスを使う。
		 */
		if apiLog.Path == "" && c.Request != nil && c.Request.URL != nil {
			apiLog.Path = c.Request.URL.Path
		}

		if len(c.Errors) > 0 {
			errorMessage := c.Errors.String()
			apiLog.ErrorMessage = &errorMessage
		}

		/*
		 * ログ保存失敗で本体APIを落とさない。
		 */
		_ = db.Create(&apiLog).Error
	}
}

func getUintPointerFromContext(c *gin.Context, key string) *uint {
	value, exists := c.Get(key)
	if !exists {
		return nil
	}

	switch typedValue := value.(type) {
	case uint:
		return &typedValue
	case int:
		converted := uint(typedValue)
		return &converted
	case float64:
		converted := uint(typedValue)
		return &converted
	default:
		return nil
	}
}

func getStringPointerFromContext(c *gin.Context, key string) *string {
	value, exists := c.Get(key)
	if !exists {
		return nil
	}

	typedValue, ok := value.(string)
	if !ok || typedValue == "" {
		return nil
	}

	return &typedValue
}
