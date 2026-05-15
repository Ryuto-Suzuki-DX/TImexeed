package models

import "time"

/*
 * 外部ストレージリンク
 *
 * Google Driveなど、Timexeed外で管理するフォルダURLやファイルURLを保存する。
 *
 * 主な用途：
 * ・経費レシート格納先
 * ・給与明細出力先
 * ・勤怠CSV出力先
 * ・その他、外部ストレージ上の管理場所
 *
 * 運用ルール：
 * ・Timexeedではファイル本体を保持しない
 * ・Google Driveなどの外部ストレージURLだけを保持する
 * ・管理者設定画面配下で管理する想定
 */
type ExternalStorageLink struct {
	ID uint `gorm:"primaryKey" json:"id"`

	/*
	 * リンク種別
	 *
	 * 例：
	 * ・EXPENSE_RECEIPT_BOX
	 * ・SALARY_STATEMENT_BOX
	 * ・ATTENDANCE_EXPORT_BOX
	 * ・OTHER
	 */
	LinkType string `gorm:"size:80;not null;index" json:"linkType"`

	/*
	 * 表示名
	 *
	 * 例：
	 * ・経費レシート格納先
	 * ・給与明細出力先
	 * ・勤怠CSV出力先
	 */
	LinkName string `gorm:"size:100;not null" json:"linkName"`

	/*
	 * 外部ストレージURL
	 *
	 * Google DriveなどのフォルダURL、またはファイルURL。
	 */
	URL string `gorm:"type:text;not null" json:"url"`

	/*
	 * 説明
	 */
	Description *string `gorm:"type:text" json:"description"`

	/*
	 * 管理メモ
	 */
	Memo *string `gorm:"type:text" json:"memo"`

	/*
	 * 論理削除
	 */
	IsDeleted bool       `gorm:"not null;default:false" json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}
