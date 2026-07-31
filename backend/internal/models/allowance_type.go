package models

import "time"

/*
 * 〇 手当種別マスター
 *
 * 月次手当で選択できる手当種別を管理する。
 * CSV出力時の列順はDisplayOrderを使用する。
 */
type AllowanceType struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"type:varchar(100);not null" json:"name"`
	DisplayOrder int        `gorm:"not null;default:0;index" json:"displayOrder"`
	IsActive     bool       `gorm:"not null;default:true;index" json:"isActive"`
	IsDeleted    bool       `gorm:"not null;default:false;index" json:"isDeleted"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"deletedAt"`
}
