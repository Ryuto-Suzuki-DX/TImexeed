package types

import "time"

/*
 * 従業員 日別交通費検索Request
 *
 * POST /user/attendance-transport-expenses/search
 *
 * 注意：
 * ・userId / targetUserId は受け取らない
 * ・ログイン中ユーザーIDはControllerでJWTから取得する
 */
type SearchAttendanceTransportExpensesRequest struct {
	TargetYear  int `json:"targetYear" binding:"required"`
	TargetMonth int `json:"targetMonth" binding:"required"`
}

/*
 * 従業員 日別交通費Response
 */
type AttendanceTransportExpenseResponse struct {
	ID uint `json:"id"`

	AttendanceDayID uint      `json:"attendanceDayId"`
	WorkDate        time.Time `json:"workDate"`

	SortOrder int `json:"sortOrder"`

	TransportFrom   string  `json:"transportFrom"`
	TransportTo     string  `json:"transportTo"`
	TransportMethod string  `json:"transportMethod"`
	TransportAmount int     `json:"transportAmount"`
	TransportMemo   *string `json:"transportMemo"`

	IsDeleted bool       `json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

/*
 * 従業員 日別交通費検索Response
 */
type SearchAttendanceTransportExpensesResponse struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	AttendanceTransportExpenses []AttendanceTransportExpenseResponse `json:"attendanceTransportExpenses"`
}

/*
 * 月次勤怠全体保存用 日別交通費明細Request
 *
 * AttendanceTransportExpenseID:
 * ・nilまたは0：新規作成
 * ・1以上：既存明細更新
 */
type UpdateAttendanceTransportExpensesByWorkDateExpenseRequest struct {
	AttendanceTransportExpenseID *uint `json:"attendanceTransportExpenseId"`

	SortOrder int `json:"sortOrder"`

	TransportFrom   string  `json:"transportFrom"`
	TransportTo     string  `json:"transportTo"`
	TransportMethod string  `json:"transportMethod"`
	TransportAmount int     `json:"transportAmount"`
	TransportMemo   *string `json:"transportMemo"`
}

/*
 * 月次勤怠全体保存用 対象日の日別交通費差分保存Request
 *
 * 注意：
 * ・APIとして直接公開しない
 * ・monthly_attendances/updateから内部的に使用する
 * ・userIdはService引数として受け取る
 */
type UpdateAttendanceTransportExpensesByWorkDateRequest struct {
	WorkDate string `json:"workDate"`

	TransportExpenses []UpdateAttendanceTransportExpensesByWorkDateExpenseRequest `json:"transportExpenses"`
}

/*
 * 月次勤怠全体保存用 対象日の日別交通費差分保存Response
 */
type UpdateAttendanceTransportExpensesByWorkDateResponse struct {
	WorkDate string `json:"workDate"`

	SavedAttendanceTransportExpenseCount int `json:"savedAttendanceTransportExpenseCount"`
}
