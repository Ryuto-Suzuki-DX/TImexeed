package responses

import (
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * API共通レスポンス形式
 */
type ApiResponse struct {
	Data    any    `json:"data"`
	Error   bool   `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

/*
 * Service結果をJSONで返す
 *
 * Controllerでは基本的にこの関数だけを使う
 */
func JSON(c *gin.Context, result results.Result) {
	c.JSON(result.StatusCode, ApiResponse{
		Data:    result.Data,
		Error:   result.Error,
		Code:    result.Code,
		Message: result.Message,
		Details: result.Details,
	})
}
