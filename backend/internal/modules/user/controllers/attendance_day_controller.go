package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用勤怠Controller
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * 注意：
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・Requestを別の型へ詰め直さない
 * ・c.JSONは直接使わず responses.JSON を使う
 * ・従業員APIでは userId / targetUserId を request body で受け取らない
 *
 * エラー方針：
 * ・Controllerで発生したエラーはControllerでcode/messageを決める
 * ・Serviceで発生したエラーはServiceでcode/messageを決める
 * ・Builderで発生したエラーはBuilderでcode/messageを決める
 * ・Repositoryで発生したエラーはRepositoryでcode/messageを決める
 * ・Controllerは最終的に responses.JSON で返す
 */
type AttendanceDayController struct {
	attendanceDayService services.AttendanceDayService
}

/*
 * AttendanceDayController生成
 */
func NewAttendanceDayController(attendanceDayService services.AttendanceDayService) *AttendanceDayController {
	return &AttendanceDayController{
		attendanceDayService: attendanceDayService,
	}
}

/*
 * 勤怠検索
 *
 * POST /user/attendance-days/search
 *
 * 用途：
 * ・従業員本人の対象月の勤怠一覧を取得する
 * ・月次一覧画面に表示する
 * ・月次申請前チェックにも使う
 *
 * 仕様：
 * ・フロントから userId / targetUserId は送らない
 * ・AuthMiddlewareでJWTから取得した userId を使う
 * ・ログイン中ユーザー本人の勤怠だけを取得する
 */
func (controller *AttendanceDayController) SearchAttendanceDays(c *gin.Context) {
	var req types.SearchAttendanceDaysRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"SEARCH_ATTENDANCE_DAYS_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	loginUserID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"SEARCH_ATTENDANCE_DAYS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをSearchAttendanceDaysRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"SEARCH_ATTENDANCE_DAYS_INVALID_REQUEST",
			"勤怠検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.attendanceDayService.SearchAttendanceDays(loginUserID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 勤怠更新
 *
 * monthly_attendanceからのみ呼び出されるのでAPI不要
 *
 */

/*
 * 勤怠削除
 *
 * POST /user/attendance-days/delete
 *
 * 用途：
 * ・従業員本人の1日分の勤怠を論理削除する
 *
 * 仕様：
 * ・フロントから userId / targetUserId は送らない
 * ・AuthMiddlewareでJWTから取得した userId を使う
 * ・Service側で loginUserID + workDate から対象勤怠を特定する
 * ・対象勤怠が存在すれば論理削除する
 */
func (controller *AttendanceDayController) DeleteAttendanceDay(c *gin.Context) {
	var req types.DeleteAttendanceDayRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"DELETE_ATTENDANCE_DAY_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	loginUserID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"DELETE_ATTENDANCE_DAY_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをDeleteAttendanceDayRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"DELETE_ATTENDANCE_DAY_INVALID_REQUEST",
			"勤怠削除のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.attendanceDayService.DeleteAttendanceDay(loginUserID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
