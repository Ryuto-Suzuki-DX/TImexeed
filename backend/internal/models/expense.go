package models

import "time"

/*
 * 〇 経費
 *
 * 上長確認済みの経費を登録するためのテーブル。
 *
 * 注意：
 * ・承認フローは持たない
 * ・ステータス管理もしない
 * ・登録されている経費 = 上長確認済みとして扱う
 * ・対象月は year / month に分けず、date 型で月初日として保持する
 * ・領収書ファイル情報もこのテーブルに保持する
 */
type Expense struct {
	ID uint `gorm:"primaryKey" json:"id"`

	/*
	 * 登録対象ユーザー
	 *
	 * USER側ではJWTから取得する。
	 * ADMIN側では request body の targetUserId で受け取る。
	 */
	UserID uint `gorm:"not null;index" json:"userId"`
	User   User `gorm:"foreignKey:UserID" json:"user"`

	/*
	 * 対象月
	 *
	 * DBでは date 型で保持する。
	 * 例：2026年5月分なら 2026-05-01
	 *
	 * フロントからは "2026-05" で受け取り、
	 * バックエンド内部で 2026-05-01 に変換する。
	 */
	TargetMonth time.Time `gorm:"type:date;not null;index" json:"targetMonth"`

	/*
	 * 経費発生日
	 */
	ExpenseDate time.Time `gorm:"type:date;not null;index" json:"expenseDate"`

	/*
	 * 金額
	 */
	Amount int `gorm:"not null" json:"amount"`

	/*
	 * 内容
	 */
	Description string `gorm:"type:text;not null" json:"description"`

	/*
	 * メモ
	 */
	Memo *string `gorm:"type:text" json:"memo"`

	/*
	 * 登録時の元ファイル名
	 *
	 * ユーザー画面に表示するのは基本これ。
	 */
	OriginalFileName *string `gorm:"size:255" json:"originalFileName"`

	/*
	 * Google Driveへ実際に保存したファイル名
	 */
	StoredFileName *string `gorm:"size:255" json:"storedFileName"`

	/*
	 * Google Drive上のファイルURL
	 *
	 * 注意：
	 * ユーザーには直接返さない。
	 * アプリ側の receipt/view API 経由で表示する。
	 */
	FileURL *string `gorm:"type:text" json:"fileUrl"`

	/*
	 * Google Drive ファイルID
	 */
	DriveFileID *string `gorm:"size:255;index" json:"driveFileId"`

	/*
	 * 保存先として参照した external_storage_links.id
	 *
	 * 経費レシート保存先は ExternalStorageLink で管理する。
	 */
	ExternalStorageLinkID *uint `gorm:"index" json:"externalStorageLinkId"`

	/*
	 * MIMEタイプ
	 */
	MimeType *string `gorm:"size:100" json:"mimeType"`

	/*
	 * ファイルサイズ byte
	 */
	SizeBytes *int64 `json:"sizeBytes"`

	/*
	 * 論理削除
	 */
	IsDeleted bool       `gorm:"not null;default:false" json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}
