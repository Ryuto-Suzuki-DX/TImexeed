package models

import "time"

/*
 * 共有資料Driveフォルダ
 *
 * 管理者がGoogle Drive上で作成した共有資料フォルダをTimexeedに登録する。
 * このフォルダに対して、管理者全員と共有対象ユーザーの権限を同期する。
 */
type SharedDocumentDriveFolder struct {
	ID uint `gorm:"primaryKey"`

	FolderName    string  `gorm:"not null"`
	Description   *string `gorm:"type:text"`
	DriveFolderID string  `gorm:"not null;uniqueIndex"`
	FolderURL     string  `gorm:"not null"`
	SyncedAt      *time.Time

	IsDeleted bool `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
