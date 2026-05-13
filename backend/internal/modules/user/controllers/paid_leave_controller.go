package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用有給Controller
 *
 * 役割：
 * ・JWTからログイン中ユーザーIDを取得する
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * 注意：
 * ・従業員APIでは targetUserId を受け取らない
 * ・対象ユーザーIDは必ずJWTから取得する
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・c.JSONは直接使わず responses.JSON を使う
 */
type PaidLeaveController struct {
	paidLeaveService services.PaidLeaveService
}

/*
 * PaidLeaveController生成
 */
func NewPaidLeaveController(paidLeaveService services.PaidLeaveService) *PaidLeaveController {
	return &PaidLeaveController{
		paidLeaveService: paidLeaveService,
	}
}

/*
 * 有給残数取得
 *
 * GET /user/paid-leave/balance
 */
func (controller *PaidLeaveController) GetPaidLeaveBalance(c *gin.Context) {
	// AuthMiddlewareでContextにセットされたログインユーザーIDを取得する
	loginUserID := c.GetUint("userId")

	if loginUserID == 0 {
		responses.JSON(c, results.Unauthorized(
			"GET_PAID_LEAVE_BALANCE_UNAUTHORIZED",
			"ログイン情報を取得できませんでした",
			nil,
		))
		return
	}

	result := controller.paidLeaveService.GetPaidLeaveBalance(loginUserID)

	responses.JSON(c, result)
}
