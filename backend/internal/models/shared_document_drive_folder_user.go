package models

import "time"

/*
 * 共有資料Driveフォルダ 共有対象ユーザー
 *
 * 共有資料フォルダとユーザーの紐づき。
 * USER側はこのテーブルをもとに、自分に共有された資料だけ表示する。
 */
type SharedDocumentDriveFolderUser struct {
	ID uint `gorm:"primaryKey"`

	SharedDocumentDriveFolderID uint `gorm:"not null;index:idx_shared_document_folder_user,unique"`
	UserID                      uint `gorm:"not null;index:idx_shared_document_folder_user,unique"`

	SyncedAt *time.Time

	IsDeleted bool `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
