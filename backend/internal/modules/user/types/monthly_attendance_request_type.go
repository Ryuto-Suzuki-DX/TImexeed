package types

import "time"

/*
 * 〇 従業員 月次勤怠申請 Type
 *
 * 従業員本人が対象月の月次勤怠を申請・取り下げするための型。
 *
 * 重要：
 * ・従業員APIでは userId / targetUserId をRequestで受け取らない
 * ・userId は Controller で JWT から取得し、Serviceへ渡す
 * ・管理者による承認・否認は admin 側の型で別管理する
 *
 * 状態：
 * ・NOT_SUBMITTED
 *     DBには保存しない。
 *     MonthlyAttendanceRequest のレコードが存在しない場合に、
 *     フロント返却用として使う未申請状態。
 *
 * ・PENDING
 *     申請中。
 *     従業員は勤怠を編集できない。
 *     従業員は取り下げできる。
 *
 * ・APPROVED
 *     承認済み。
 *     従業員は勤怠を編集できない。
 *     従業員は取り下げできない。
 *
 * ・REJECTED
 *     否認済み。
 *     従業員は勤怠を編集できる。
 *     従業員は再申請できる。
 *
 * ・CANCELED
 *     取り下げ済み。
 *     従業員は勤怠を編集できる。
 *     従業員は再申請できる。
 */

/*
 * 月次勤怠申請状態取得 Request
 *
 * POST /user/monthly-attendance-requests/status
 */
type GetMonthlyAttendanceRequestStatusRequest struct {
	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`
}

/*
 * 月次勤怠申請 Request
 *
 * POST /user/monthly-attendance-requests/submit
 */
type SubmitMonthlyAttendanceRequestRequest struct {
	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`

	// 申請メモ
	RequestMemo *string `json:"requestMemo"`
}

/*
 * 月次勤怠申請取り下げ Request
 *
 * POST /user/monthly-attendance-requests/cancel
 */
type CancelMonthlyAttendanceRequestRequest struct {
	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`

	// 取り下げ理由
	CanceledReason *string `json:"canceledReason"`
}

/*
 * 月次勤怠申請 Response
 *
 * 月次申請状態取得・申請・取り下げで共通して返す。
 *
 * 用途：
 * ・月次勤怠画面の状態表示
 * ・編集可否の判定
 * ・申請ボタン/取り下げボタンの表示制御
 */
type MonthlyAttendanceRequestResponse struct {
	// 月次勤怠申請ID
	// 未申請の場合は nil
	ID *uint `json:"id"`

	// 対象年
	TargetYear int `json:"targetYear"`

	// 対象月
	TargetMonth int `json:"targetMonth"`

	// 月次申請状態
	// 例：NOT_SUBMITTED, PENDING, APPROVED, REJECTED, CANCELED
	Status string `json:"status"`

	// レコードが存在するか
	// 未申請の場合は false
	Exists bool `json:"exists"`

	// 勤怠編集可能か
	// NOT_SUBMITTED / REJECTED / CANCELED は true
	// PENDING / APPROVED は false
	Editable bool `json:"editable"`

	// 月次申請できるか
	// NOT_SUBMITTED / REJECTED / CANCELED は true
	// PENDING / APPROVED は false
	CanSubmit bool `json:"canSubmit"`

	// 取り下げできるか
	// PENDING のみ true
	CanCancel bool `json:"canCancel"`

	// 申請メモ
	RequestMemo *string `json:"requestMemo"`

	// 申請日時
	RequestedAt *time.Time `json:"requestedAt"`

	// 承認者ID
	ApprovedBy *uint `json:"approvedBy"`

	// 承認日時
	ApprovedAt *time.Time `json:"approvedAt"`

	// 否認理由
	RejectedReason *string `json:"rejectedReason"`

	// 否認日時
	RejectedAt *time.Time `json:"rejectedAt"`

	// 取り下げ理由
	CanceledReason *string `json:"canceledReason"`

	// 取り下げ日時
	CanceledAt *time.Time `json:"canceledAt"`

	// 作成日時
	CreatedAt *time.Time `json:"createdAt"`

	// 更新日時
	UpdatedAt *time.Time `json:"updatedAt"`
}

/*
 * 月次勤怠申請状態取得 Response
 */
type GetMonthlyAttendanceRequestStatusResponse struct {
	MonthlyAttendanceRequest MonthlyAttendanceRequestResponse `json:"monthlyAttendanceRequest"`
}

/*
 * 月次勤怠申請 Response
 */
type SubmitMonthlyAttendanceRequestResponse struct {
	MonthlyAttendanceRequest MonthlyAttendanceRequestResponse `json:"monthlyAttendanceRequest"`
}

/*
 * 月次勤怠申請取り下げ Response
 */
type CancelMonthlyAttendanceRequestResponse struct {
	MonthlyAttendanceRequest MonthlyAttendanceRequestResponse `json:"monthlyAttendanceRequest"`
}
