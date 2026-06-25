package types

import "time"

/*
 * 〇 従業員 勤怠リアルタイムイベント Type
 *
 * ユーザー側mypageの出勤・退勤ボタンで使用する。
 *
 * 注意：
 * ・ユーザーIDはリクエストで受け取らない
 * ・ControllerでJWTからログイン中ユーザーIDを取得する
 * ・月次勤怠には反映しない
 * ・同じユーザーが同じ日に同じイベント種別を登録できるのは1回だけ
 * ・登録後の取消・編集はしない
 * ・ユーザーが登録できるイベント種別は出勤と退勤のみ
 */

/*
 * =========================================================
 * Request
 * =========================================================
 */

/*
 * 勤怠リアルタイムイベント作成リクエスト
 */
type CreateAttendanceRealtimeEventRequest struct {
	EventType string `json:"eventType" binding:"required"`
	Note      string `json:"note"`
}

/*
 * 本日の勤怠リアルタイムイベント状態取得リクエスト
 *
 * ユーザーIDはリクエストで受け取らない。
 * ControllerでJWTから取得する。
 */
type GetTodayAttendanceRealtimeEventsRequest struct{}

/*
 * =========================================================
 * Response
 * =========================================================
 */

/*
 * 勤怠リアルタイムイベントレスポンス
 */
type AttendanceRealtimeEventResponse struct {
	ID        uint      `json:"id"`
	EventDate time.Time `json:"eventDate"`
	EventType string    `json:"eventType"`
	EventAt   time.Time `json:"eventAt"`
	Note      *string   `json:"note"`
	CreatedAt time.Time `json:"createdAt"`
}

/*
 * 勤怠リアルタイムイベント作成レスポンス
 */
type CreateAttendanceRealtimeEventResponse struct {
	Event AttendanceRealtimeEventResponse `json:"event"`
}

/*
 * 本日の勤怠リアルタイムイベント状態取得レスポンス
 *
 * 出勤・退勤それぞれについて、
 * ・登録済みか
 * ・押下時刻
 * ・コメント
 * を返す。
 */
type GetTodayAttendanceRealtimeEventsResponse struct {
	ClockInRecorded  bool       `json:"clockInRecorded"`
	ClockOutRecorded bool       `json:"clockOutRecorded"`
	ClockInAt        *time.Time `json:"clockInAt"`
	ClockOutAt       *time.Time `json:"clockOutAt"`
	ClockInNote      *string    `json:"clockInNote"`
	ClockOutNote     *string    `json:"clockOutNote"`

	Events []AttendanceRealtimeEventResponse `json:"events"`
}
