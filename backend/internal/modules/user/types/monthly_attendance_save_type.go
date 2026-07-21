package types

/*
 * 〇 月次勤怠全体保存リクエスト
 *
 * 月次勤怠画面の「全体保存」用。
 *
 * 保存対象：
 * ・月次通勤定期（複数件）
 * ・日別勤怠
 * ・日別交通費
 * ・日別休憩
 *
 * 注意：
 * ・userId / targetUserId は受け取らない
 * ・ログイン中ユーザーIDはControllerでJWTから取得してServiceへ渡す
 * ・予定区分は PlanAttendanceTypeID
 * ・実績状態は ActualWorkStatus
 * ・ActualAttendanceTypeID は使わない
 */
type UpdateMonthlyAttendanceRequest struct {
	TargetYear  int `json:"targetYear" binding:"required"`
	TargetMonth int `json:"targetMonth" binding:"required"`

	// 月次通勤定期一覧
	//
	// 方針：
	// ・画面に残っている通勤定期だけ送る
	// ・monthlyCommuterPassId がある明細は更新する
	// ・monthlyCommuterPassId がない明細は新規作成する
	// ・DBに存在するがRequestから消えた明細は論理削除する
	// ・空配列の場合は対象年月の既存定期をすべて論理削除する
	CommuterPasses []UpdateMonthlyAttendanceCommuterPassRequest `json:"commuterPasses"`

	AttendanceDays []UpdateMonthlyAttendanceDayRequest `json:"attendanceDays"`
}

/*
 * 〇 月次勤怠全体保存：月次通勤定期
 */
type UpdateMonthlyAttendanceCommuterPassRequest struct {
	// 月次通勤定期ID
	// 新規作成の場合は nil
	MonthlyCommuterPassID *uint `json:"monthlyCommuterPassId"`

	CommuterFrom   *string `json:"commuterFrom"`
	CommuterTo     *string `json:"commuterTo"`
	CommuterMethod *string `json:"commuterMethod"`
	CommuterAmount *int    `json:"commuterAmount"`
}

/*
 * 〇 月次勤怠全体保存：日別勤怠
 */
type UpdateMonthlyAttendanceDayRequest struct {
	WorkDate string `json:"workDate" binding:"required"`

	PlanAttendanceTypeID uint    `json:"planAttendanceTypeId" binding:"required"`
	ActualWorkStatus     *string `json:"actualWorkStatus"`

	CommonStartAt *string `json:"commonStartAt"`
	CommonEndAt   *string `json:"commonEndAt"`

	PlanStartAt *string `json:"planStartAt"`
	PlanEndAt   *string `json:"planEndAt"`

	ActualStartAt *string `json:"actualStartAt"`
	ActualEndAt   *string `json:"actualEndAt"`

	ScheduledWorkMinutes    *int    `json:"scheduledWorkMinutes"`
	RemoteWorkAllowanceFlag bool    `json:"remoteWorkAllowanceFlag"`
	RequestMemo             *string `json:"requestMemo"`

	TransportExpenses []UpdateMonthlyAttendanceTransportExpenseRequest `json:"transportExpenses"`
	Breaks            []UpdateMonthlyAttendanceBreakRequest            `json:"breaks"`
}

/*
 * 〇 月次勤怠全体保存：日別交通費
 */
type UpdateMonthlyAttendanceTransportExpenseRequest struct {
	AttendanceTransportExpenseID *uint `json:"attendanceTransportExpenseId"`

	SortOrder int `json:"sortOrder"`

	TransportFrom   string `json:"transportFrom" binding:"required"`
	TransportTo     string `json:"transportTo" binding:"required"`
	TransportMethod string `json:"transportMethod" binding:"required"`
	TransportAmount int    `json:"transportAmount"`

	TransportMemo *string `json:"transportMemo"`
}

/*
 * 〇 月次勤怠全体保存：休憩
 */
type UpdateMonthlyAttendanceBreakRequest struct {
	BreakStartAt string  `json:"breakStartAt" binding:"required"`
	BreakEndAt   string  `json:"breakEndAt" binding:"required"`
	BreakMemo    *string `json:"breakMemo"`
}

/*
 * 〇 月次勤怠全体保存レスポンス
 */
type UpdateMonthlyAttendanceResponse struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	SavedMonthlyCommuterPassCount        int `json:"savedMonthlyCommuterPassCount"`
	SavedAttendanceDayCount              int `json:"savedAttendanceDayCount"`
	SavedAttendanceTransportExpenseCount int `json:"savedAttendanceTransportExpenseCount"`
	SavedAttendanceBreakCount            int `json:"savedAttendanceBreakCount"`
}
