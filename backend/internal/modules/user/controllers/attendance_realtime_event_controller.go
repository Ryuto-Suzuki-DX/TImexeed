package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用 勤怠リアルタイムイベントController
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * 注意：
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・Requestを別の型へ詰め直さない
 * ・c.JSONは直接使わず responses.JSON を使う
 * ・従業員APIでは userId / targetUserId を request body で受け取らない
 */
type AttendanceRealtimeEventController struct {
	attendanceRealtimeEventService services.AttendanceRealtimeEventService
}

/*
 * AttendanceRealtimeEventController生成
 */
func NewAttendanceRealtimeEventController(
	attendanceRealtimeEventService services.AttendanceRealtimeEventService,
) *AttendanceRealtimeEventController {
	return &AttendanceRealtimeEventController{
		attendanceRealtimeEventService: attendanceRealtimeEventService,
	}
}

/*
 * 勤怠リアルタイムイベントController用ログインユーザーID取得
 */
func getAttendanceRealtimeEventLoginUserID(c *gin.Context, actionCode string) (uint, results.Result) {
	userIDValue, exists := c.Get("userId")
	if !exists {
		return 0, results.Unauthorized(
			actionCode+"_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		)
	}

	loginUserID, ok := userIDValue.(uint)
	if !ok || loginUserID == 0 {
		return 0, results.Unauthorized(
			actionCode+"_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	return loginUserID, results.OK(
		nil,
		actionCode+"_VALID_USER_ID",
		"",
		nil,
	)
}

/*
 * 勤怠リアルタイムイベント作成
 *
 * POST /user/attendance-realtime-events/create
 *
 * 用途：
 * ・従業員本人がmypageで出勤・退勤・その他ボタンを押した事実を記録する
 *
 * 仕様：
 * ・フロントから userId / targetUserId は送らない
 * ・AuthMiddlewareでJWTから取得した userId を使う
 * ・同じユーザーが同じ日に同じイベント種別を登録できるのは1回だけ
 * ・月次勤怠には反映しない
 */
func (controller *AttendanceRealtimeEventController) CreateAttendanceRealtimeEvent(c *gin.Context) {
	loginUserID, userIDResult := getAttendanceRealtimeEventLoginUserID(c, "CREATE_ATTENDANCE_REALTIME_EVENT")
	if userIDResult.Error {
		responses.JSON(c, userIDResult)
		return
	}

	var req types.CreateAttendanceRealtimeEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"CREATE_ATTENDANCE_REALTIME_EVENT_INVALID_REQUEST",
			"勤怠リアルタイムイベント作成のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.attendanceRealtimeEventService.CreateAttendanceRealtimeEvent(
		loginUserID,
		req,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	responses.JSON(c, result)
}

/*
 * 本日の勤怠リアルタイムイベント状態取得
 *
 * POST /user/attendance-realtime-events/today
 *
 * 用途：
 * ・mypage表示時に、出勤・退勤・その他ボタンを押せるか判定する
 *
 * 仕様：
 * ・フロントから userId / targetUserId は送らない
 * ・AuthMiddlewareでJWTから取得した userId を使う
 */
func (controller *AttendanceRealtimeEventController) GetTodayAttendanceRealtimeEvents(c *gin.Context) {
	loginUserID, userIDResult := getAttendanceRealtimeEventLoginUserID(c, "GET_TODAY_ATTENDANCE_REALTIME_EVENTS")
	if userIDResult.Error {
		responses.JSON(c, userIDResult)
		return
	}

	var req types.GetTodayAttendanceRealtimeEventsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"GET_TODAY_ATTENDANCE_REALTIME_EVENTS_INVALID_REQUEST",
			"本日の勤怠リアルタイムイベント状態取得のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.attendanceRealtimeEventService.GetTodayAttendanceRealtimeEvents(loginUserID, req)

	responses.JSON(c, result)
}
