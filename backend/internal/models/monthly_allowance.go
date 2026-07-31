package models

import "time"

/*
 * 〇 月次手当
 *
 * ユーザーごと・対象年月ごとの手当明細を管理する。
 * 同じユーザー・年月に複数件登録できる。
 */
type MonthlyAllowance struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UserID          uint       `gorm:"not null;index" json:"userId"`
	TargetYear      int        `gorm:"not null;index" json:"targetYear"`
	TargetMonth     int        `gorm:"not null;index" json:"targetMonth"`
	AllowanceTypeID uint       `gorm:"not null;index" json:"allowanceTypeId"`
	Amount          int        `gorm:"not null;default:0" json:"amount"`
	Memo            string     `gorm:"type:text" json:"memo"`
	IsDeleted       bool       `gorm:"not null;default:false;index" json:"isDeleted"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt"`
}
