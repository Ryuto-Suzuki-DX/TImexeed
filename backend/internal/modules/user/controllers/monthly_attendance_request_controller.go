package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用月次勤怠申請Controller
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * このControllerで扱うもの：
 * ・従業員本人の対象月の月次申請状態取得
 * ・従業員本人の対象月の月次申請
 * ・従業員本人の対象月の月次申請取り下げ
 *
 * このControllerで扱わないもの：
 * ・月次承認
 * ・月次否認
 * ・他ユーザーの月次申請操作
 * ・勤怠日別データの更新
 * ・休憩データの更新
 * ・月次通勤定期の更新
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
 * 状態管理方針：
 * ・月次申請状態は MonthlyAttendanceRequest で管理する
 * ・未申請は MonthlyAttendanceRequest のレコードなしで表現する
 * ・申請、取り下げの可否判定は Service 側で行う
 * ・管理者による承認、否認は管理者API側で行う
 *
 * エラー方針：
 * ・Controllerで発生したエラーはControllerでcode/messageを決める
 * ・Serviceで発生したエラーはServiceでcode/messageを決める
 * ・Builderで発生したエラーはBuilderでcode/messageを決める
 * ・Repositoryで発生したエラーはRepositoryでcode/messageを決める
 * ・Controllerは最終的に responses.JSON で返す
 */
type MonthlyAttendanceRequestController struct {
	monthlyAttendanceRequestService services.MonthlyAttendanceRequestService
}

/*
 * MonthlyAttendanceRequestController生成
 */
func NewMonthlyAttendanceRequestController(
	monthlyAttendanceRequestService services.MonthlyAttendanceRequestService,
) *MonthlyAttendanceRequestController {
	return &MonthlyAttendanceRequestController{
		monthlyAttendanceRequestService: monthlyAttendanceRequestService,
	}
}

/*
 * 月次勤怠申請状態取得
 *
 * POST /user/monthly-attendance-requests/status
 *
 * 用途：
 * ・従業員本人の対象月の月次申請状態を取得する
 * ・月次勤怠画面の編集可否判定に使う
 * ・申請ボタン、取り下げボタンの表示制御に使う
 *
 * 仕様：
 * ・フロントから userId / targetUserId は送らない
 * ・AuthMiddlewareでJWTから取得した userId を使う
 * ・ログイン中ユーザー本人の月次申請状態だけを取得する
 * ・対象レコードが存在しない場合は、Service側で未申請扱いとして返す
 */
func (controller *MonthlyAttendanceRequestController) GetMonthlyAttendanceRequestStatus(c *gin.Context) {
	var req types.GetMonthlyAttendanceRequestStatusRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"GET_MONTHLY_ATTENDANCE_REQUEST_STATUS_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	loginUserID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"GET_MONTHLY_ATTENDANCE_REQUEST_STATUS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをGetMonthlyAttendanceRequestStatusRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"GET_MONTHLY_ATTENDANCE_REQUEST_STATUS_INVALID_REQUEST",
			"月次勤怠申請状態取得のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.monthlyAttendanceRequestService.GetMonthlyAttendanceRequestStatus(loginUserID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 月次勤怠申請
 *
 * POST /user/monthly-attendance-requests/submit
 *
 * 用途：
 * ・従業員本人の対象月の勤怠を月次申請する
 *
 * 仕様：
 * ・フロントから userId / targetUserId は送らない
 * ・AuthMiddlewareでJWTから取得した userId を使う
 * ・ログイン中ユーザー本人の月次勤怠だけを申請する
 * ・未申請の場合は新規申請する
 * ・否認済み、取り下げ済みの場合は再申請できる
 * ・申請中、承認済みの場合はService側で拒否する
 */
func (controller *MonthlyAttendanceRequestController) SubmitMonthlyAttendanceRequest(c *gin.Context) {
	var req types.SubmitMonthlyAttendanceRequestRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"SUBMIT_MONTHLY_ATTENDANCE_REQUEST_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	loginUserID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"SUBMIT_MONTHLY_ATTENDANCE_REQUEST_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをSubmitMonthlyAttendanceRequestRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SUBMIT_MONTHLY_ATTENDANCE_REQUEST_INVALID_REQUEST",
			"月次勤怠申請のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.monthlyAttendanceRequestService.SubmitMonthlyAttendanceRequest(loginUserID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 月次勤怠申請取り下げ
 *
 * POST /user/monthly-attendance-requests/cancel
 *
 * 用途：
 * ・従業員本人の対象月の月次勤怠申請を取り下げる
 *
 * 仕様：
 * ・フロントから userId / targetUserId は送らない
 * ・AuthMiddlewareでJWTから取得した userId を使う
 * ・ログイン中ユーザー本人の月次勤怠申請だけを取り下げる
 * ・取り下げできるのは申請中の月次勤怠のみ
 * ・承認済み、否認済み、取り下げ済み、未申請の場合はService側で拒否する
 */
func (controller *MonthlyAttendanceRequestController) CancelMonthlyAttendanceRequest(c *gin.Context) {
	var req types.CancelMonthlyAttendanceRequestRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"CANCEL_MONTHLY_ATTENDANCE_REQUEST_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	loginUserID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"CANCEL_MONTHLY_ATTENDANCE_REQUEST_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをCancelMonthlyAttendanceRequestRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"CANCEL_MONTHLY_ATTENDANCE_REQUEST_INVALID_REQUEST",
			"月次勤怠申請取り下げのリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.monthlyAttendanceRequestService.CancelMonthlyAttendanceRequest(loginUserID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
