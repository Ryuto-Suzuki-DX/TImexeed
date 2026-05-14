package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用月次勤怠申請Controller
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・承認、否認ではAuthMiddlewareでJWTから取得した管理者IDを取得する
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * このControllerで扱うもの：
 * ・対象ユーザーの対象月の月次申請状態取得
 * ・対象ユーザーの対象月の月次申請
 * ・対象ユーザーの対象月の月次申請取り下げ
 * ・月次勤怠申請の承認
 * ・月次勤怠申請の否認
 *
 * このControllerで扱わないもの：
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
 * ・管理者APIでは対象ユーザーIDを request body の targetUserId で受け取る
 * ・承認者ID、否認者IDは request body で受け取らず、JWTの管理者IDを使う
 *
 * 状態管理方針：
 * ・月次申請状態は MonthlyAttendanceRequest で管理する
 * ・未申請は MonthlyAttendanceRequest のレコードなしで表現する
 * ・申請、取り下げ、承認、否認の可否判定は Service 側で行う
 *
 * 管理者編集方針：
 * ・管理者は月次申請状態に関係なく勤怠を編集できる
 * ・このControllerで返す月次申請状態は表示、申請操作、承認操作のために使う
 * ・勤怠編集ロックには使わない
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
 * JWTから管理者IDを取得する
 *
 * 承認、否認で approvedBy / rejectedBy として使う。
 */
func getLoginAdminID(c *gin.Context, actionCode string) (uint, results.Result) {
	userIDValue, exists := c.Get("userId")
	if !exists {
		return 0, results.Unauthorized(
			actionCode+"_ADMIN_ID_NOT_FOUND",
			"認証情報から管理者IDを取得できません",
			nil,
		)
	}

	loginAdminID, ok := userIDValue.(uint)
	if !ok {
		return 0, results.Unauthorized(
			actionCode+"_INVALID_ADMIN_ID",
			"認証情報の管理者IDが正しくありません",
			nil,
		)
	}

	if loginAdminID == 0 {
		return 0, results.Unauthorized(
			actionCode+"_EMPTY_ADMIN_ID",
			"認証情報の管理者IDが正しくありません",
			nil,
		)
	}

	return loginAdminID, results.OK(
		nil,
		actionCode+"_ADMIN_ID_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次勤怠申請状態取得
 *
 * POST /admin/monthly-attendance-requests/status
 *
 * 用途：
 * ・対象ユーザーの対象月の月次申請状態を取得する
 * ・管理者勤怠画面の状態表示に使う
 * ・管理者側の申請ボタン、取り下げボタンの表示制御に使う
 *
 * 仕様：
 * ・対象ユーザーIDは request body の targetUserId で受け取る
 * ・対象レコードが存在しない場合は、Service側で未申請扱いとして返す
 * ・管理者側では月次申請状態で勤怠編集ロックはしない
 */
func (controller *MonthlyAttendanceRequestController) GetMonthlyAttendanceRequestStatus(c *gin.Context) {
	var req types.GetMonthlyAttendanceRequestStatusRequest

	// リクエストJSONをGetMonthlyAttendanceRequestStatusRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"GET_MONTHLY_ATTENDANCE_REQUEST_STATUS_INVALID_REQUEST",
			"月次勤怠申請状態取得のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.monthlyAttendanceRequestService.GetMonthlyAttendanceRequestStatus(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 月次勤怠申請
 *
 * POST /admin/monthly-attendance-requests/submit
 *
 * 用途：
 * ・管理者が対象ユーザーの対象月の勤怠を代理で月次申請する
 *
 * 仕様：
 * ・対象ユーザーIDは request body の targetUserId で受け取る
 * ・未申請の場合は新規申請する
 * ・否認済み、取り下げ済みの場合は再申請できる
 * ・申請中、承認済みの場合はService側で拒否する
 */
func (controller *MonthlyAttendanceRequestController) SubmitMonthlyAttendanceRequest(c *gin.Context) {
	var req types.SubmitMonthlyAttendanceRequestRequest

	// リクエストJSONをSubmitMonthlyAttendanceRequestRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SUBMIT_MONTHLY_ATTENDANCE_REQUEST_INVALID_REQUEST",
			"月次勤怠申請のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.monthlyAttendanceRequestService.SubmitMonthlyAttendanceRequest(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 月次勤怠申請取り下げ
 *
 * POST /admin/monthly-attendance-requests/cancel
 *
 * 用途：
 * ・管理者が対象ユーザーの対象月の月次勤怠申請を代理で取り下げる
 *
 * 仕様：
 * ・対象ユーザーIDは request body の targetUserId で受け取る
 * ・取り下げできるのは申請中の月次勤怠のみ
 * ・承認済み、否認済み、取り下げ済み、未申請の場合はService側で拒否する
 */
func (controller *MonthlyAttendanceRequestController) CancelMonthlyAttendanceRequest(c *gin.Context) {
	var req types.CancelMonthlyAttendanceRequestRequest

	// リクエストJSONをCancelMonthlyAttendanceRequestRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"CANCEL_MONTHLY_ATTENDANCE_REQUEST_INVALID_REQUEST",
			"月次勤怠申請取り下げのリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.monthlyAttendanceRequestService.CancelMonthlyAttendanceRequest(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 月次勤怠申請承認
 *
 * POST /admin/monthly-attendance-requests/approve
 *
 * 用途：
 * ・管理者が月次勤怠申請を承認する
 *
 * 仕様：
 * ・targetRequestId で対象の月次勤怠申請を指定する
 * ・承認者IDは request body では受け取らない
 * ・AuthMiddlewareでJWTから取得した管理者IDを使う
 * ・承認できる状態かどうかはService側で判定する
 */
func (controller *MonthlyAttendanceRequestController) ApproveMonthlyAttendanceRequest(c *gin.Context) {
	var req types.ApproveMonthlyAttendanceRequestRequest

	loginAdminID, adminIDResult := getLoginAdminID(c, "APPROVE_MONTHLY_ATTENDANCE_REQUEST")
	if adminIDResult.Error {
		responses.JSON(c, adminIDResult)
		return
	}

	// リクエストJSONをApproveMonthlyAttendanceRequestRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"APPROVE_MONTHLY_ATTENDANCE_REQUEST_INVALID_REQUEST",
			"月次勤怠申請承認のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中管理者IDをServiceへ渡す
	result := controller.monthlyAttendanceRequestService.ApproveMonthlyAttendanceRequest(loginAdminID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 月次勤怠申請否認
 *
 * POST /admin/monthly-attendance-requests/reject
 *
 * 用途：
 * ・管理者が月次勤怠申請を否認する
 *
 * 仕様：
 * ・targetRequestId で対象の月次勤怠申請を指定する
 * ・否認者IDは request body では受け取らない
 * ・AuthMiddlewareでJWTから取得した管理者IDを使う
 * ・否認できる状態かどうかはService側で判定する
 */
func (controller *MonthlyAttendanceRequestController) RejectMonthlyAttendanceRequest(c *gin.Context) {
	var req types.RejectMonthlyAttendanceRequestRequest

	loginAdminID, adminIDResult := getLoginAdminID(c, "REJECT_MONTHLY_ATTENDANCE_REQUEST")
	if adminIDResult.Error {
		responses.JSON(c, adminIDResult)
		return
	}

	// リクエストJSONをRejectMonthlyAttendanceRequestRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"REJECT_MONTHLY_ATTENDANCE_REQUEST_INVALID_REQUEST",
			"月次勤怠申請否認のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中管理者IDをServiceへ渡す
	result := controller.monthlyAttendanceRequestService.RejectMonthlyAttendanceRequest(loginAdminID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
