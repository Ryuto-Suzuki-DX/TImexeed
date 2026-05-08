package models

import "time"

type Department struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Name      string     `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	IsDeleted bool       `gorm:"not null;default:false" json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}
