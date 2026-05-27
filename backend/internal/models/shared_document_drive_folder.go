package models

import "time"

/*
 * 共有資料Driveフォルダ
 *
 * 管理者がTimexeed上から作成した全ユーザー向け共有資料フォルダを管理する。
 *
 * Google Drive上では、external_storage_links に登録された親フォルダ配下へ
 * このフォルダを作成し、作成されたDriveフォルダID・URLを保存する。
 *
 * 共有対象は個別ユーザー選択ではなく、全有効一般ユーザーを対象とする。
 * Drive権限同期は、管理者が同期ボタンを押したタイミングで実行する。
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
