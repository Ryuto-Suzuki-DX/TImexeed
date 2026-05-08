package models

import "time"

/*
 * 〇 各日の勤怠
 *
 * 1日ごとの予定・実績・申請状態・日別交通費を管理するメインテーブル。
 *
 * このテーブルに入れるもの：
 * 	・予定区分
 * 	・予定時間
 * 	・実績区分
 * 	・実績時間
 * 	・有給、休業、休職などの申請状態
 * 	・遅刻、早退、欠勤、病欠などの補足
 * 	・日別交通費
 *
 * 休憩は1日に複数回あるため、AttendanceBreak に分ける。
 * 月次通勤定期は月単位なので、MonthlyCommuterPass に分ける。
 */
type AttendanceDay struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// 対象ユーザーID
	UserID uint `gorm:"not null;index" json:"userId"`

	// 対象日
	WorkDate time.Time `gorm:"type:date;not null;index" json:"workDate"`

	// 予定区分ID
	PlanAttendanceTypeID uint `gorm:"not null" json:"planAttendanceTypeId"`

	// 実績区分ID
	ActualAttendanceTypeID uint `gorm:"not null" json:"actualAttendanceTypeId"`

	// 予定開始日時
	PlanStartAt *time.Time `json:"planStartAt"`

	// 予定終了日時
	PlanEndAt *time.Time `json:"planEndAt"`

	// 実績開始日時
	ActualStartAt *time.Time `json:"actualStartAt"`

	// 実績終了日時
	ActualEndAt *time.Time `json:"actualEndAt"`

	// 申請状態
	// 例：NONE, PENDING, APPROVED, REJECTED, CANCELED
	RequestStatus string `gorm:"type:varchar(30);not null;default:'NONE'" json:"requestStatus"`

	// 申請メモ
	RequestMemo *string `gorm:"type:text" json:"requestMemo"`

	// 承認者ID
	ApprovedBy *uint `json:"approvedBy"`

	// 承認日時
	ApprovedAt *time.Time `json:"approvedAt"`

	// 否認理由
	RejectedReason *string `gorm:"type:text" json:"rejectedReason"`

	// 遅刻フラグ
	LateFlag bool `gorm:"not null;default:false" json:"lateFlag"`

	// 早退フラグ
	EarlyLeaveFlag bool `gorm:"not null;default:false" json:"earlyLeaveFlag"`

	// 欠勤フラグ
	AbsenceFlag bool `gorm:"not null;default:false" json:"absenceFlag"`

	// 病欠フラグ
	SickLeaveFlag bool `gorm:"not null;default:false" json:"sickLeaveFlag"`

	// 画面表示用メッセージ
	// 例：残業15分、有給申請中、承認済みなど
	SystemMessage *string `gorm:"type:text" json:"systemMessage"`

	// 日別交通費：出発地
	TransportFrom *string `gorm:"type:varchar(100)" json:"transportFrom"`

	// 日別交通費：目的地
	TransportTo *string `gorm:"type:varchar(100)" json:"transportTo"`

	// 日別交通費：手段
	// 例：電車、バス、徒歩、車
	TransportMethod *string `gorm:"type:varchar(50)" json:"transportMethod"`

	// 日別交通費：金額
	TransportAmount *int `json:"transportAmount"`

	// 月次申請状態
	// 例：DRAFT, PENDING, APPROVED, REJECTED
	MonthlyStatus string `gorm:"type:varchar(30);not null;default:'DRAFT'" json:"monthlyStatus"`

	// 論理削除フラグ
	IsDeleted bool `gorm:"not null;default:false" json:"isDeleted"`

	// 作成日時
	CreatedAt time.Time `json:"createdAt"`

	// 更新日時
	UpdatedAt time.Time `json:"updatedAt"`

	// 論理削除日時
	DeletedAt *time.Time `json:"deletedAt"`

	// 予定区分
	PlanAttendanceType AttendanceType `gorm:"foreignKey:PlanAttendanceTypeID" json:"planAttendanceType,omitempty"`

	// 実績区分
	ActualAttendanceType AttendanceType `gorm:"foreignKey:ActualAttendanceTypeID" json:"actualAttendanceType,omitempty"`
}
