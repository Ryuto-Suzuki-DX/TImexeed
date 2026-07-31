package models

import (
	"time"

	"gorm.io/gorm"
)

/*
 * 〇 ユーザー給与詳細
 *
 * ユーザーごとの給与計算に使う基本情報を管理する。
 *
 * 注意：
 * ・給与計算結果ではない
 * ・CSV出力結果ではない
 * ・経費精算情報ではない
 * ・残業/深夜/休日などの割増率は持たない
 * ・所定労働時間は持たない
 * ・手当はMonthlyAllowanceで管理する
 * ・控除はこのテーブルでは管理しない
 * ・ここでは個人ごとに違う給与区分・基本金額だけを持つ
 */
type UserSalaryDetail struct {
	ID              uint       `gorm:"primaryKey"`
	UserID          uint       `gorm:"not null;index"`
	SalaryType      string     `gorm:"type:varchar(20);not null"`
	BaseAmount      int        `gorm:"not null;default:0"`
	IsPayrollTarget bool       `gorm:"not null;default:true"`
	EffectiveFrom   time.Time  `gorm:"type:date;not null"`
	EffectiveTo     *time.Time `gorm:"type:date"`
	Memo            string     `gorm:"type:text"`
	IsDeleted       bool       `gorm:"not null;default:false"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}
