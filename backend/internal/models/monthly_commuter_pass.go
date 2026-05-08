package models

import "time"

/*
 * 〇 月次通勤定期
 *
 * 月ごとの通勤定期を管理する。
 *
 * 日ごとの交通費は AttendanceDay に持たせる。
 * 月単位の定期代は、この MonthlyCommuterPass に持たせる。
 */
type MonthlyCommuterPass struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// 対象ユーザーID
	UserID uint `gorm:"not null;index" json:"userId"`

	// 対象年
	TargetYear int `gorm:"not null;index" json:"targetYear"`

	// 対象月
	TargetMonth int `gorm:"not null;index" json:"targetMonth"`

	// 定期：出発地
	CommuterFrom *string `gorm:"type:varchar(100)" json:"commuterFrom"`

	// 定期：目的地
	CommuterTo *string `gorm:"type:varchar(100)" json:"commuterTo"`

	// 定期：手段
	// 例：電車、バス、車
	CommuterMethod *string `gorm:"type:varchar(50)" json:"commuterMethod"`

	// 定期：金額
	CommuterAmount *int `json:"commuterAmount"`

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
}
