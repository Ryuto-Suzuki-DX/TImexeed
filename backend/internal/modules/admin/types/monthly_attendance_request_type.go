package types

import "time"

/*
 * 〇 管理者 月次勤怠申請 Type
 *
 * 管理者が対象ユーザーの月次勤怠申請状態を扱うための型。
 *
 * 重要：
 * ・管理者APIでは対象ユーザーIDを targetUserId としてRequestで受け取る
 * ・管理者は対象ユーザーの勤怠画面で、ユーザーと同様に申請・取り下げ操作ができる
 * ・管理者は別画面で、月次勤怠申請の承認・否認もできる
 * ・管理者側では月次申請状態に関係なく勤怠編集できる
 *
 * 状態：
 * ・NOT_SUBMITTED
 *     DBには保存しない。
 *     MonthlyAttendanceRequest のレコードが存在しない場合に、
 *     フロント返却用として使う未申請状態。
 *
 * ・PENDING
 *     申請中。
 *     管理者は勤怠編集できる。
 *     管理者は取り下げ、承認、否認できる。
 *
 * ・APPROVED
 *     承認済み。
 *     管理者は勤怠編集できる。
 *
 * ・REJECTED
 *     否認済み。
 *     管理者は勤怠編集できる。
 *     管理者は再申請できる。
 *
 * ・CANCELED
 *     取り下げ済み。
 *     管理者は勤怠編集できる。
 *     管理者は再申請できる。
 */

/*
 * 月次勤怠申請状態取得 Request
 *
 * POST /admin/monthly-attendance-requests/status
 */
type GetMonthlyAttendanceRequestStatusRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`
}

/*
 * 月次勤怠申請 Request
 *
 * POST /admin/monthly-attendance-requests/submit
 *
 * 注意：
 * ・管理者が対象ユーザーの月次勤怠申請を代理で申請する用途
 */
type SubmitMonthlyAttendanceRequestRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

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
 * POST /admin/monthly-attendance-requests/cancel
 *
 * 注意：
 * ・管理者が対象ユーザーの月次勤怠申請を代理で取り下げる用途
 */
type CancelMonthlyAttendanceRequestRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`

	// 取り下げ理由
	CanceledReason *string `json:"canceledReason"`
}

/*
 * 月次勤怠申請承認 Request
 *
 * POST /admin/monthly-attendance-requests/approve
 *
 * 注意：
 * ・承認者IDはRequestで受け取らない
 * ・ControllerでJWTから取得した管理者IDをServiceへ渡す
 */
type ApproveMonthlyAttendanceRequestRequest struct {
	// 月次勤怠申請ID
	TargetRequestID uint `json:"targetRequestId" binding:"required"`
}

/*
 * 月次勤怠申請否認 Request
 *
 * POST /admin/monthly-attendance-requests/reject
 *
 * 注意：
 * ・否認者IDはRequestで受け取らない
 * ・ControllerでJWTから取得した管理者IDをServiceへ渡す
 */
type RejectMonthlyAttendanceRequestRequest struct {
	// 月次勤怠申請ID
	TargetRequestID uint `json:"targetRequestId" binding:"required"`

	// 否認理由
	RejectedReason string `json:"rejectedReason" binding:"required"`
}

/*
 * 月次勤怠申請 Response
 *
 * 月次申請状態取得・申請・取り下げ・承認・否認で共通して返す。
 *
 * 用途：
 * ・管理者勤怠画面の状態表示
 * ・管理者月次承認画面の状態表示
 *
 * 注意：
 * ・Editable / CanSubmit / CanCancel は画面表示制御用
 * ・管理者側では Editable=false でも勤怠編集自体は許可する
 */
type MonthlyAttendanceRequestResponse struct {
	// 月次勤怠申請ID
	// 未申請の場合は nil
	ID *uint `json:"id"`

	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId"`

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
	// user側の表示制御と型を合わせるために返す。
	// 管理者側ではこの値で勤怠編集ロックはしない。
	Editable bool `json:"editable"`

	// 月次申請できるか
	// NOT_SUBMITTED / REJECTED / CANCELED は true
	// PENDING / APPROVED は false
	CanSubmit bool `json:"canSubmit"`

	// 取り下げできるか
	// PENDING のみ true
	CanCancel bool `json:"canCancel"`

	// 承認できるか
	// PENDING のみ true
	CanApprove bool `json:"canApprove"`

	// 否認できるか
	// PENDING のみ true
	CanReject bool `json:"canReject"`

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

/*
 * 月次勤怠申請承認 Response
 */
type ApproveMonthlyAttendanceRequestResponse struct {
	MonthlyAttendanceRequest MonthlyAttendanceRequestResponse `json:"monthlyAttendanceRequest"`
}

/*
 * 月次勤怠申請否認 Response
 */
type RejectMonthlyAttendanceRequestResponse struct {
	MonthlyAttendanceRequest MonthlyAttendanceRequestResponse `json:"monthlyAttendanceRequest"`
}
