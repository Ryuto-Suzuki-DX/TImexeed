package models

import (
	"time"

	"gorm.io/gorm"
)

/*
 * 個人情報Driveフォルダ
 *
 * ユーザーごとの個人情報関連ファイルを格納するGoogle Driveフォルダを管理する。
 *
 * このフォルダは、管理者全員と対象ユーザー本人のみが閲覧できる。
 *
 * 用途：
 * ・ユーザーごとの個人情報ファイル置き場
 * ・Google Drive上のフォルダ自動生成
 * ・Google Drive権限の最新状態への同期
 * ・管理者側からのユーザー別フォルダ一覧/検索
 * ・従業員側からの自分のフォルダ表示
 *
 * 注意：
 * ・Timexeed側ではDrive内のファイル追加・削除・一覧管理は行わない
 * ・Timexeed側はフォルダ生成、権限付与、URL表示のみ担当する
 * ・管理者は全ユーザー分のフォルダを開ける
 * ・ユーザーは自分のフォルダだけ開ける
 */
type PersonalInformationDriveFolder struct {
	ID uint `gorm:"primaryKey" json:"id"`

	/*
	 * 対象ユーザーID
	 *
	 * 1ユーザーにつき1つの個人情報Driveフォルダを持つ。
	 */
	UserID uint `gorm:"not null;uniqueIndex" json:"userId"`

	/*
	 * 親の外部ストレージ設定ID
	 *
	 * external_storage_links の PERSONAL_INFORMATION_DRIVE_ROOT を参照する。
	 * この親フォルダ配下に、ユーザーごとのフォルダを作成する。
	 */
	ExternalStorageLinkID uint `gorm:"not null;index" json:"externalStorageLinkId"`

	/*
	 * Google Drive上のフォルダ名
	 *
	 * 例：
	 * 0001_山田太郎
	 * user_1_山田太郎
	 */
	FolderName string `gorm:"type:varchar(255);not null" json:"folderName"`

	/*
	 * Google DriveのフォルダID
	 *
	 * URLではなく、Drive API操作に必要なIDを保存する。
	 */
	DriveFolderID string `gorm:"type:varchar(255);not null;uniqueIndex" json:"driveFolderId"`

	/*
	 * Google Driveで開くためのURL
	 *
	 * 管理者画面、従業員画面ではこのURLを開く。
	 * ただし、API側で必ず権限チェックを行った上で返す。
	 */
	FolderURL string `gorm:"type:text;not null" json:"folderUrl"`

	/*
	 * 最後にDrive権限同期を行った日時
	 *
	 * 管理者の「最新状態に更新」ボタンで、
	 * フォルダ作成や管理者/本人の権限同期を行った日時。
	 */
	SyncedAt *time.Time `json:"syncedAt"`

	/*
	 * 論理削除
	 *
	 * ユーザーが退職・削除された場合でも、
	 * Google Drive上のフォルダを即削除する想定はしない。
	 * Timexeed側で非表示/無効扱いにしたい場合に使う。
	 */
	IsDeleted bool `gorm:"not null;default:false" json:"isDeleted"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
