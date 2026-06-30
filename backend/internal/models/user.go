package models

import "time"

type User struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `gorm:"type:varchar(100);not null" json:"name"`
	Email          string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash   string     `gorm:"type:varchar(255);not null" json:"-"`
	Role           string     `gorm:"type:varchar(20);not null;default:'USER'" json:"role"`
	DepartmentID   *uint      `json:"departmentId"`
	HireDate       time.Time  `gorm:"type:date;not null" json:"hireDate"`
	RetirementDate *time.Time `gorm:"type:date" json:"retirementDate"`

	// 初期パスワード変更が必要かどうか
	MustChangePassword bool `gorm:"not null;default:false" json:"mustChangePassword"`

	// 最後にパスワードを変更した日時
	PasswordChangedAt *time.Time `json:"passwordChangedAt"`

	IsDeleted bool       `gorm:"not null;default:false" json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}
