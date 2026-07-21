package types

/*
 * 〇 管理者 月次勤怠全体保存リクエスト
 *
 * 管理者用月次勤怠画面の「全体保存」用。
 *
 * 保存対象：
 * ・月次通勤定期（複数件）
 * ・日別勤怠
 * ・日別交通費
 * ・日別休憩
 *
 * 重要：
 * ・管理者APIでは対象ユーザーIDを targetUserId としてRequestで受け取る
 * ・ControllerではJWTのuserIdを対象ユーザーIDとして使わない
 * ・管理者側では月次申請状態による編集ロックを行わない
 *
 * 保存方針：
 * ・このRequestをControllerでbindする
 * ・Service側で既存のadmin用Serviceへ処理を振り分ける
 * ・日別勤怠は AttendanceDayService.UpdateAttendanceDay を使う
 * ・日別交通費は AttendanceTransportExpenseService.UpdateAttendanceTransportExpensesByWorkDate を使う
 * ・休憩は AttendanceBreakService.UpdateAttendanceBreaksByWorkDate を使う
 * ・月次通勤定期は MonthlyCommuterPassService.UpdateMonthlyCommuterPasses を使う
 */
type UpdateMonthlyAttendanceRequest struct {
	TargetUserID uint `json:"targetUserId" binding:"required"`
	TargetYear   int  `json:"targetYear" binding:"required"`
	TargetMonth  int  `json:"targetMonth" binding:"required"`

	// 月次通勤定期一覧
	// 空配列の場合、対象年月の既存定期をすべて論理削除する。
	CommuterPasses []UpdateMonthlyAttendanceCommuterPassRequest `json:"commuterPasses"`

	AttendanceDays []UpdateMonthlyAttendanceDayRequest `json:"attendanceDays"`
}

/*
 * 〇 管理者 月次勤怠全体保存：月次通勤定期
 */
type UpdateMonthlyAttendanceCommuterPassRequest struct {
	// 月次通勤定期ID。新規作成の場合はnil。
	MonthlyCommuterPassID *uint `json:"monthlyCommuterPassId"`

	CommuterFrom   *string `json:"commuterFrom"`
	CommuterTo     *string `json:"commuterTo"`
	CommuterMethod *string `json:"commuterMethod"`
	CommuterAmount *int    `json:"commuterAmount"`
}

/*
 * 〇 管理者 月次勤怠全体保存：日別勤怠
 */
type UpdateMonthlyAttendanceDayRequest struct {
	WorkDate string `json:"workDate" binding:"required"`

	PlanAttendanceTypeID uint    `json:"planAttendanceTypeId" binding:"required"`
	ActualWorkStatus     *string `json:"actualWorkStatus"`

	CommonStartAt *string `json:"commonStartAt"`
	CommonEndAt   *string `json:"commonEndAt"`
	PlanStartAt   *string `json:"planStartAt"`
	PlanEndAt     *string `json:"planEndAt"`
	ActualStartAt *string `json:"actualStartAt"`
	ActualEndAt   *string `json:"actualEndAt"`

	ScheduledWorkMinutes    *int    `json:"scheduledWorkMinutes"`
	RemoteWorkAllowanceFlag bool    `json:"remoteWorkAllowanceFlag"`
	RequestMemo             *string `json:"requestMemo"`

	TransportExpenses []UpdateMonthlyAttendanceTransportExpenseRequest `json:"transportExpenses"`
	Breaks            []UpdateMonthlyAttendanceBreakRequest            `json:"breaks"`
}

/*
 * 〇 管理者 月次勤怠全体保存：日別交通費
 */
type UpdateMonthlyAttendanceTransportExpenseRequest struct {
	AttendanceTransportExpenseID *uint   `json:"attendanceTransportExpenseId"`
	SortOrder                    int     `json:"sortOrder"`
	TransportFrom                string  `json:"transportFrom" binding:"required"`
	TransportTo                  string  `json:"transportTo" binding:"required"`
	TransportMethod              string  `json:"transportMethod" binding:"required"`
	TransportAmount              int     `json:"transportAmount"`
	TransportMemo                *string `json:"transportMemo"`
}

/*
 * 〇 管理者 月次勤怠全体保存：休憩
 */
type UpdateMonthlyAttendanceBreakRequest struct {
	AttendanceBreakID *uint   `json:"attendanceBreakId"`
	BreakStartAt      string  `json:"breakStartAt" binding:"required"`
	BreakEndAt        string  `json:"breakEndAt" binding:"required"`
	BreakMemo         *string `json:"breakMemo"`
}

/*
 * 〇 管理者 月次勤怠全体保存レスポンス
 */
type UpdateMonthlyAttendanceResponse struct {
	TargetUserID uint `json:"targetUserId"`
	TargetYear   int  `json:"targetYear"`
	TargetMonth  int  `json:"targetMonth"`

	SavedMonthlyCommuterPassCount        int `json:"savedMonthlyCommuterPassCount"`
	SavedAttendanceDayCount              int `json:"savedAttendanceDayCount"`
	SavedAttendanceTransportExpenseCount int `json:"savedAttendanceTransportExpenseCount"`
	SavedAttendanceBreakCount            int `json:"savedAttendanceBreakCount"`
}
