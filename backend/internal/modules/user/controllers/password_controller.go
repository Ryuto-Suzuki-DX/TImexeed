// 配置先:
// backend/internal/modules/user/controllers/password_controller.go

package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * パスワード Controller Interface
 */
type PasswordController interface {
	ChangePassword(c *gin.Context)
}

/*
 * パスワード Controller
 */
type passwordController struct {
	passwordService services.PasswordService
}

/*
 * パスワード Controller生成
 */
func NewPasswordController(
	passwordService services.PasswordService,
) PasswordController {
	return &passwordController{
		passwordService: passwordService,
	}
}

/*
 * パスワード変更
 *
 * POST /user/password/change
 */
func (controller *passwordController) ChangePassword(c *gin.Context) {
	var request types.ChangePasswordRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		responses.JSON(
			c,
			results.BadRequest(
				"CHANGE_PASSWORD_REQUEST_INVALID",
				"リクエスト内容が正しくありません。",
				err,
			),
		)
		return
	}

	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(
			c,
			results.Unauthorized(
				"AUTHENTICATED_USER_NOT_FOUND",
				"ログイン情報を取得できませんでした。",
				nil,
			),
		)
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok || userID == 0 {
		responses.JSON(
			c,
			results.Unauthorized(
				"AUTHENTICATED_USER_INVALID",
				"ログイン情報が正しくありません。",
				nil,
			),
		)
		return
	}

	result := controller.passwordService.ChangePassword(userID, request)
	responses.JSON(c, result)
}
