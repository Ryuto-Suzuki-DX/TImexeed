package types

import "time"

/*
 * 〇 従業員 勤怠日別 Type
 *
 * 従業員本人の勤怠日別データを扱う型。
 *
 * 重要：
 * ・AttendanceDay は日別勤怠データだけを持つ
 * ・日別交通費は AttendanceTransportExpense で管理する
 * ・申請状態、承認状態は AttendanceDay では持たない
 * ・月次申請状態は MonthlyAttendanceRequestResponse として返す
 * ・従業員APIでは userId / targetUserId をRequestで受け取らない
 */

/*
 * 勤怠検索 Request
 */
type SearchAttendanceDaysRequest struct {
	TargetYear  int `json:"targetYear" binding:"required"`
	TargetMonth int `json:"targetMonth" binding:"required"`
}

/*
 * 勤怠更新 Request
 *
 * APIとして直接公開しない。
 * monthly_attendances/update の月次全体保存から内部的に使う。
 *
 * 注意：
 * ・日別交通費はこのRequestでは扱わない
 * ・日別交通費は月次保存Serviceから専用Serviceへ渡す
 */
type UpdateAttendanceDayRequest struct {
	WorkDate string `json:"workDate" binding:"required"`

	PlanAttendanceTypeID uint `json:"planAttendanceTypeId" binding:"required"`

	ActualWorkStatus *string `json:"actualWorkStatus"`

	PlanStartAt *string `json:"planStartAt"`
	PlanEndAt   *string `json:"planEndAt"`

	ActualStartAt *string `json:"actualStartAt"`
	ActualEndAt   *string `json:"actualEndAt"`

	CommonStartAt *string `json:"commonStartAt"`
	CommonEndAt   *string `json:"commonEndAt"`

	ScheduledWorkMinutes *int `json:"scheduledWorkMinutes"`

	RemoteWorkAllowanceFlag bool `json:"remoteWorkAllowanceFlag"`
}

/*
 * 勤怠削除 Request
 */
type DeleteAttendanceDayRequest struct {
	WorkDate string `json:"workDate" binding:"required"`
}

/*
 * 勤怠日別 Response
 *
 * AttendanceDay 自体のデータだけを返す。
 * 日別交通費は専用の検索Responseで返す。
 */
type AttendanceDayResponse struct {
	ID uint `json:"id"`

	WorkDate time.Time `json:"workDate"`

	PlanAttendanceTypeID uint   `json:"planAttendanceTypeId"`
	ActualWorkStatus     string `json:"actualWorkStatus"`

	PlanStartAt *time.Time `json:"planStartAt"`
	PlanEndAt   *time.Time `json:"planEndAt"`

	ActualStartAt *time.Time `json:"actualStartAt"`
	ActualEndAt   *time.Time `json:"actualEndAt"`

	ScheduledWorkMinutes *int `json:"scheduledWorkMinutes"`

	RemoteWorkAllowanceFlag bool `json:"remoteWorkAllowanceFlag"`

	IsDeleted bool `json:"isDeleted"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

/*
 * 勤怠検索 Response
 */
type SearchAttendanceDaysResponse struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	MonthlyAttendanceRequest MonthlyAttendanceRequestResponse `json:"monthlyAttendanceRequest"`

	AttendanceDays []AttendanceDayResponse `json:"attendanceDays"`
}

/*
 * 勤怠更新 Response
 */
type UpdateAttendanceDayResponse struct {
	AttendanceDay AttendanceDayResponse `json:"attendanceDay"`
}

/*
 * 勤怠削除 Response
 */
type DeleteAttendanceDayResponse struct {
	WorkDate string `json:"workDate"`
}
