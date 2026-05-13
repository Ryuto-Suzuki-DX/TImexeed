package models

import "time"

/*
 * 〇 お知らせ
 *
 * ユーザーに表示するお知らせを管理する。
 *
 * 操作履歴ではなく、ユーザーが確認するためのお知らせ本文を保存する。
 *
 * 例：
 * 	・2026年5月の月次勤怠を申請しました。
 * 	・2026年5月の月次勤怠が承認されました。
 * 	・2026年5月の月次勤怠が否認されました。
 *
 * 通知種別・関連対象IDは持たない。
 * 未読 / 既読の表示を行うため、既読フラグと既読日時は持つ。
 */
type Notification struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// 通知先ユーザーID
	UserID uint `gorm:"not null;index" json:"userId"`

	// タイトル
	Title string `gorm:"type:varchar(150);not null" json:"title"`

	// 本文
	Message string `gorm:"type:text;not null" json:"message"`

	// 既読フラグ
	IsRead bool `gorm:"not null;default:false" json:"isRead"`

	// 既読日時
	ReadAt *time.Time `json:"readAt"`

	// 論理削除フラグ
	IsDeleted bool `gorm:"not null;default:false" json:"isDeleted"`

	// 作成日時
	CreatedAt time.Time `json:"createdAt"`

	// 更新日時
	UpdatedAt time.Time `json:"updatedAt"`

	// 論理削除日時
	DeletedAt *time.Time `json:"deletedAt"`
}
