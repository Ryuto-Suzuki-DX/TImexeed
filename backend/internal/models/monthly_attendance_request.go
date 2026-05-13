package models

import "time"

/*
 * 〇 月次勤怠申請
 *
 * 従業員が対象月の勤怠を月次申請し、
 * 管理者が承認・否認するための親テーブル。
 *
 * このテーブルに入れるもの：
 * 	・対象ユーザー
 * 	・対象年月
 * 	・月次申請状態
 * 	・申請日時
 * 	・申請メモ
 * 	・承認者
 * 	・承認日時
 * 	・否認理由
 * 	・取り下げ日時
 * 	・取り下げ理由
 *
 * このテーブルに入れないもの：
 * 	・日別の予定区分
 * 	・日別の予定時間
 * 	・日別の実績区分
 * 	・日別の実績時間
 * 	・日別交通費
 * 	・休憩時間
 *
 * 理由：
 * 	AttendanceDay は日別勤怠データだけを管理する。
 * 	MonthlyAttendanceRequest は、その月全体の申請・承認状態だけを管理する。
 *
 * 状態：
 * 	・PENDING  = 申請中
 * 	・APPROVED = 承認済み
 * 	・REJECTED = 否認済み
 * 	・CANCELED = 取り下げ済み
 *
 * 補足：
 * 	未申請はレコードなしで表現する。
 * 	そのため NOT_SUBMITTED はDBには保存しない。
 */
type MonthlyAttendanceRequest struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// 対象ユーザーID
	UserID uint `gorm:"not null;index" json:"userId"`

	// 対象年
	TargetYear int `gorm:"not null;index" json:"targetYear"`

	// 対象月
	TargetMonth int `gorm:"not null;index" json:"targetMonth"`

	// 月次申請状態
	// 例：PENDING, APPROVED, REJECTED, CANCELED
	Status string `gorm:"type:varchar(30);not null" json:"status"`

	// 申請メモ
	RequestMemo *string `gorm:"type:text" json:"requestMemo"`

	// 申請日時
	RequestedAt *time.Time `json:"requestedAt"`

	// 承認者ID
	ApprovedBy *uint `json:"approvedBy"`

	// 承認日時
	ApprovedAt *time.Time `json:"approvedAt"`

	// 否認理由
	RejectedReason *string `gorm:"type:text" json:"rejectedReason"`

	// 否認日時
	RejectedAt *time.Time `json:"rejectedAt"`

	// 取り下げ理由
	CanceledReason *string `gorm:"type:text" json:"canceledReason"`

	// 取り下げ日時
	CanceledAt *time.Time `json:"canceledAt"`

	// 論理削除フラグ
	IsDeleted bool `gorm:"not null;default:false" json:"isDeleted"`

	// 作成日時
	CreatedAt time.Time `json:"createdAt"`

	// 更新日時
	UpdatedAt time.Time `json:"updatedAt"`

	// 論理削除日時
	DeletedAt *time.Time `json:"deletedAt"`
}
