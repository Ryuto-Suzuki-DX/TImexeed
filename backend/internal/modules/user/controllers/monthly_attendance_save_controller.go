package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用月次勤怠全体保存Controller
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * このControllerで扱うもの：
 * ・月次勤怠画面の全体保存
 * ・月次通勤定期
 * ・日別勤怠
 * ・日別休憩
 *
 * このControllerで扱わないもの：
 * ・月次申請
 * ・月次承認
 * ・有給申請
 * ・DB処理
 * ・業務ルール
 *
 * 注意：
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・Requestを別の型へ詰め直さない
 * ・c.JSONは直接使わず responses.JSON を使う
 * ・従業員APIでは userId / targetUserId を request body で受け取らない
 *
 * 命名方針：
 * ・このControllerは月次申請ではなく、月次勤怠画面の保存を担当する
 * ・そのため、本来は MonthlyAttendanceSaveController や MonthlyAttendanceBulkSaveController
 *   のような名前の方が役割が分かりやすい
 */
type MonthlyAttendanceSaveController struct {
	monthlyAttendanceService services.MonthlyAttendanceSaveService
}

/*
 * MonthlyAttendanceController生成
 */
func NewMonthlyAttendanceSaveController(
	monthlyAttendanceService services.MonthlyAttendanceSaveService,
) *MonthlyAttendanceSaveController {
	return &MonthlyAttendanceSaveController{
		monthlyAttendanceService: monthlyAttendanceService,
	}
}

/*
 * 月次勤怠全体保存
 *
 * POST /user/monthly-attendances/update
 *
 * 用途：
 * ・月次勤怠画面の全体保存
 * ・月次通勤定期、日別勤怠、休憩をまとめて保存する
 *
 * 注意：
 * ・SystemMessage はDB保存しない
 * ・画面表示用メッセージは、保存値ではなく表示時に組み立てる
 */
func (controller *MonthlyAttendanceSaveController) UpdateMonthlyAttendance(c *gin.Context) {
	var req types.UpdateMonthlyAttendanceRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"UPDATE_MONTHLY_ATTENDANCE_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"UPDATE_MONTHLY_ATTENDANCE_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをUpdateMonthlyAttendanceRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"UPDATE_MONTHLY_ATTENDANCE_INVALID_REQUEST",
			"月次勤怠全体保存のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.monthlyAttendanceService.UpdateMonthlyAttendance(userID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
