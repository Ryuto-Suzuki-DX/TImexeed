package types

/*
 * 〇 月次勤怠全体保存リクエスト
 *
 * 月次勤怠画面の「全体保存」用。
 *
 * 保存対象：
 * ・月次通勤定期
 * ・日別勤怠
 * ・日別休憩
 *
 * 注意：
 * ・userId / targetUserId は受け取らない
 * ・ログイン中ユーザーIDはControllerでJWTから取得してServiceへ渡す
 */
type UpdateMonthlyAttendanceRequest struct {
	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`

	// 月次通勤定期
	CommuterPass *UpdateMonthlyAttendanceCommuterPassRequest `json:"commuterPass"`

	// 日別勤怠一覧
	AttendanceDays []UpdateMonthlyAttendanceDayRequest `json:"attendanceDays"`
}

/*
 * 〇 月次勤怠全体保存：月次通勤定期
 */
type UpdateMonthlyAttendanceCommuterPassRequest struct {
	// 定期：出発地
	CommuterFrom *string `json:"commuterFrom"`

	// 定期：目的地
	CommuterTo *string `json:"commuterTo"`

	// 定期：手段
	CommuterMethod *string `json:"commuterMethod"`

	// 定期：金額
	CommuterAmount *int `json:"commuterAmount"`
}

/*
 * 〇 月次勤怠全体保存：日別勤怠
 */
type UpdateMonthlyAttendanceDayRequest struct {
	// 対象日
	// 例：2026-05-05
	WorkDate string `json:"workDate" binding:"required"`

	// 予定区分ID
	PlanAttendanceTypeID uint `json:"planAttendanceTypeId" binding:"required"`

	// 実績区分ID
	ActualAttendanceTypeID *uint `json:"actualAttendanceTypeId"`

	// 共通開始日時
	CommonStartAt *string `json:"commonStartAt"`

	// 共通終了日時
	CommonEndAt *string `json:"commonEndAt"`

	// 予定開始日時
	PlanStartAt *string `json:"planStartAt"`

	// 予定終了日時
	PlanEndAt *string `json:"planEndAt"`

	// 実績開始日時
	ActualStartAt *string `json:"actualStartAt"`

	// 実績終了日時
	ActualEndAt *string `json:"actualEndAt"`

	// 遅刻フラグ
	LateFlag bool `json:"lateFlag"`

	// 早退フラグ
	EarlyLeaveFlag bool `json:"earlyLeaveFlag"`

	// 欠勤フラグ
	AbsenceFlag bool `json:"absenceFlag"`

	// 病欠フラグ
	SickLeaveFlag bool `json:"sickLeaveFlag"`

	// 申請メモ
	RequestMemo *string `json:"requestMemo"`

	// 日別交通費：出発地
	TransportFrom *string `json:"transportFrom"`

	// 日別交通費：目的地
	TransportTo *string `json:"transportTo"`

	// 日別交通費：手段
	TransportMethod *string `json:"transportMethod"`

	// 日別交通費：金額
	TransportAmount *int `json:"transportAmount"`

	// 休憩一覧
	Breaks []UpdateMonthlyAttendanceBreakRequest `json:"breaks"`
}

/*
 * 〇 月次勤怠全体保存：休憩
 *
 * 方針：
 * ・画面に残っている休憩だけ送る
 * ・保存時は既存休憩を削除して作り直す
 * ・そのため attendanceBreakId は使わない
 */
type UpdateMonthlyAttendanceBreakRequest struct {
	// 休憩開始日時
	BreakStartAt string `json:"breakStartAt" binding:"required"`

	// 休憩終了日時
	BreakEndAt string `json:"breakEndAt" binding:"required"`

	// 休憩メモ
	BreakMemo *string `json:"breakMemo"`
}

/*
 * 〇 月次勤怠全体保存レスポンス
 */
type UpdateMonthlyAttendanceResponse struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	SavedMonthlyCommuterPass  bool `json:"savedMonthlyCommuterPass"`
	SavedAttendanceDayCount   int  `json:"savedAttendanceDayCount"`
	SavedAttendanceBreakCount int  `json:"savedAttendanceBreakCount"`
}
