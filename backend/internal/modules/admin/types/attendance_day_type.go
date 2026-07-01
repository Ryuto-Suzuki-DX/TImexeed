package types

import "time"

/*
 * 管理者 勤怠日別検索Request
 *
 * POST /admin/attendance-days/search
 */
type SearchAttendanceDaysRequest struct {
	TargetUserID uint `json:"targetUserId"`
	TargetYear   int  `json:"targetYear"`
	TargetMonth  int  `json:"targetMonth"`
}

/*
 * 管理者 勤怠日別検索Response
 */
type SearchAttendanceDaysResponse struct {
	TargetUserID             uint                             `json:"targetUserId"`
	TargetYear               int                              `json:"targetYear"`
	TargetMonth              int                              `json:"targetMonth"`
	MonthlyAttendanceRequest MonthlyAttendanceRequestResponse `json:"monthlyAttendanceRequest"`
	AttendanceDays           []AttendanceDayResponse          `json:"attendanceDays"`
}

/*
 * 管理者 勤怠日別Response
 *
 * 注意：
 * ・予定区分は attendance_types のIDを返す
 * ・実績状態は constants/attendance_status_constants.go の固定値を返す
 * ・実績状態は attendance_types のIDではない
 * ・日別交通費は別テーブルで管理するため、このResponseには含めない
 */
type AttendanceDayResponse struct {
	ID     uint `json:"id"`
	UserID uint `json:"userId"`

	WorkDate time.Time `json:"workDate"`

	PlanAttendanceTypeID uint   `json:"planAttendanceTypeId"`
	ActualWorkStatus     string `json:"actualWorkStatus"`

	PlanStartAt   *time.Time `json:"planStartAt"`
	PlanEndAt     *time.Time `json:"planEndAt"`
	ActualStartAt *time.Time `json:"actualStartAt"`
	ActualEndAt   *time.Time `json:"actualEndAt"`

	ScheduledWorkMinutes *int `json:"scheduledWorkMinutes"`

	RemoteWorkAllowanceFlag bool `json:"remoteWorkAllowanceFlag"`

	IsDeleted bool       `json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

/*
 * 管理者 勤怠日別更新Request
 *
 * 注意：
 * ・このAPIは直接公開しない想定
 * ・monthly_attendances/update の月次全体保存から内部的に使う
 * ・管理者APIでは targetUserId を request body で受け取る
 * ・予定区分は planAttendanceTypeId
 * ・実績状態は actualWorkStatus
 * ・日別交通費は別テーブルで管理するため、このRequestには含めない
 */
type UpdateAttendanceDayRequest struct {
	TargetUserID uint   `json:"targetUserId"`
	WorkDate     string `json:"workDate"`

	PlanAttendanceTypeID uint    `json:"planAttendanceTypeId"`
	ActualWorkStatus     *string `json:"actualWorkStatus"`

	CommonStartAt *string `json:"commonStartAt"`
	CommonEndAt   *string `json:"commonEndAt"`

	PlanStartAt   *string `json:"planStartAt"`
	PlanEndAt     *string `json:"planEndAt"`
	ActualStartAt *string `json:"actualStartAt"`
	ActualEndAt   *string `json:"actualEndAt"`

	ScheduledWorkMinutes *int `json:"scheduledWorkMinutes"`

	RemoteWorkAllowanceFlag bool `json:"remoteWorkAllowanceFlag"`
}

/*
 * 管理者 勤怠日別更新Response
 */
type UpdateAttendanceDayResponse struct {
	AttendanceDay AttendanceDayResponse `json:"attendanceDay"`
}

/*
 * 管理者 勤怠日別削除Request
 *
 * 注意：
 * ・現時点では直接公開しない想定
 */
type DeleteAttendanceDayRequest struct {
	TargetUserID uint   `json:"targetUserId"`
	WorkDate     string `json:"workDate"`
}

/*
 * 管理者 勤怠日別削除Response
 */
type DeleteAttendanceDayResponse struct {
	TargetUserID uint   `json:"targetUserId"`
	WorkDate     string `json:"workDate"`
}
