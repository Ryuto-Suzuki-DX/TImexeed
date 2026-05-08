package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用休憩Controller
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
type AttendanceBreakController struct {
	attendanceBreakService services.AttendanceBreakService
}

/*
 * AttendanceBreakController生成
 */
func NewAttendanceBreakController(attendanceBreakService services.AttendanceBreakService) *AttendanceBreakController {
	return &AttendanceBreakController{
		attendanceBreakService: attendanceBreakService,
	}
}

/*
 * 休憩検索
 *
 * POST /user/attendance-breaks/search
 */
func (controller *AttendanceBreakController) SearchAttendanceBreaks(c *gin.Context) {
	var req types.SearchAttendanceBreaksRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"SEARCH_ATTENDANCE_BREAKS_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"SEARCH_ATTENDANCE_BREAKS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをSearchAttendanceBreaksRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_ATTENDANCE_BREAKS_INVALID_REQUEST",
			"休憩検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.attendanceBreakService.SearchAttendanceBreaks(userID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 休憩作成
 *
 * POST /user/attendance-breaks/create
 */
func (controller *AttendanceBreakController) CreateAttendanceBreak(c *gin.Context) {
	var req types.CreateAttendanceBreakRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"CREATE_ATTENDANCE_BREAK_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"CREATE_ATTENDANCE_BREAK_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをCreateAttendanceBreakRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"CREATE_ATTENDANCE_BREAK_INVALID_REQUEST",
			"休憩作成のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.attendanceBreakService.CreateAttendanceBreak(userID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 休憩更新
 *
 * monthly_attendanceからのみ呼び出される
 *
 */

/*
 * 休憩削除
 *
 * POST /user/attendance-breaks/delete
 */
func (controller *AttendanceBreakController) DeleteAttendanceBreak(c *gin.Context) {
	var req types.DeleteAttendanceBreakRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"DELETE_ATTENDANCE_BREAK_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"DELETE_ATTENDANCE_BREAK_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをDeleteAttendanceBreakRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"DELETE_ATTENDANCE_BREAK_INVALID_REQUEST",
			"休憩削除のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.attendanceBreakService.DeleteAttendanceBreak(userID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
