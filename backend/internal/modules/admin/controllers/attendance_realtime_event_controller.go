package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用 勤怠リアルタイムイベントController
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * 注意：
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・Requestを別の型へ詰め直さない
 * ・c.JSONは直接使わず responses.JSON を使う
 * ・管理者はイベントを作成しない
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
 * 勤怠リアルタイムイベント検索
 *
 * POST /admin/attendance-realtime-events/search
 *
 * 用途：
 * ・管理者が従業員の出勤・退勤・その他ボタン押下状況を確認する
 *
 * 仕様：
 * ・管理者はイベントを作成しない
 * ・対象日未指定の場合はJSTの本日を検索する
 */
func (controller *AttendanceRealtimeEventController) SearchAttendanceRealtimeEvents(c *gin.Context) {
	var req types.SearchAttendanceRealtimeEventsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_ATTENDANCE_REALTIME_EVENTS_INVALID_REQUEST",
			"勤怠リアルタイムイベント検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.attendanceRealtimeEventService.SearchAttendanceRealtimeEvents(req)

	responses.JSON(c, result)
}
