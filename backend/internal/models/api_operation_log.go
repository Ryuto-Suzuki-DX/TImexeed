package models

import "time"

/*
 * API操作ログ
 *
 * API単位で、誰が・いつ・どのAPIを実行したかを追跡する。
 *
 * 保存方針：
 * ・API実行時にDBへ保存する
 * ・毎日設定時刻に対象日分をCSV化してGoogle Driveへアップロードする
 * ・Google Drive側は半年分だけ保持する
 *
 * 注意：
 * ・パスワード、JWT、認証情報、個人番号などの機密情報は保存しない
 * ・request_body の丸ごと保存はしない
 */
type ApiOperationLog struct {
	ID uint `gorm:"primaryKey" json:"id"`

	/*
	 * ログインユーザー情報
	 */
	UserID *uint   `gorm:"index" json:"userId"`
	Email  *string `gorm:"size:255" json:"email"`
	Role   *string `gorm:"size:50" json:"role"`

	/*
	 * API情報
	 */
	Method string `gorm:"size:10;not null;index" json:"method"`
	Path   string `gorm:"size:255;not null;index" json:"path"`

	/*
	 * HTTP結果
	 */
	StatusCode int `gorm:"not null;index" json:"statusCode"`

	/*
	 * アクセス元
	 */
	ClientIP  string `gorm:"size:100" json:"clientIp"`
	UserAgent string `gorm:"type:text" json:"userAgent"`

	/*
	 * 処理時間
	 */
	DurationMs int64 `gorm:"not null" json:"durationMs"`

	/*
	 * エラー情報
	 *
	 * GinのレスポンスBodyをそのまま読む設計にはせず、
	 * 最初はstatusCodeだけで追跡する。
	 * 必要になったらService側の業務ログで詳細を残す。
	 */
	ErrorMessage *string `gorm:"type:text" json:"errorMessage"`

	/*
	 * 実行日時
	 */
	StartedAt  time.Time `gorm:"not null;index" json:"startedAt"`
	FinishedAt time.Time `gorm:"not null" json:"finishedAt"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
