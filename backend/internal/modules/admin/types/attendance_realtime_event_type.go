package types

import "time"

/*
 * 〇 管理者 勤怠リアルタイムイベント Type
 *
 * 管理者側で、従業員がmypageで押した出勤・退勤・その他の時刻を確認する。
 *
 * 注意：
 * ・管理者はイベントを作成しない
 * ・検索と一覧表示のみ
 * ・月次勤怠には反映しない
 */

/*
 * =========================================================
 * Request
 * =========================================================
 */

/*
 * 勤怠リアルタイムイベント検索リクエスト
 */
type SearchAttendanceRealtimeEventsRequest struct {
	TargetDate string   `json:"targetDate"`
	Keyword    string   `json:"keyword"`
	EventTypes []string `json:"eventTypes"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
}

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
	UserID    uint      `json:"userId"`
	UserName  string    `json:"userName"`
	UserEmail string    `json:"userEmail"`
	EventDate time.Time `json:"eventDate"`
	EventType string    `json:"eventType"`
	EventAt   time.Time `json:"eventAt"`
	Note      *string   `json:"note"`
	ClientIP  *string   `json:"clientIp"`
	UserAgent *string   `json:"userAgent"`
	CreatedAt time.Time `json:"createdAt"`
}

/*
 * 勤怠リアルタイムイベント検索レスポンス
 */
type SearchAttendanceRealtimeEventsResponse struct {
	Events  []AttendanceRealtimeEventResponse `json:"events"`
	Total   int64                             `json:"total"`
	Offset  int                               `json:"offset"`
	Limit   int                               `json:"limit"`
	HasMore bool                              `json:"hasMore"`
}
