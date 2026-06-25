package models

import "time"

/*
 * 勤怠リアルタイムイベント
 *
 * ユーザーがマイページで押した
 * ・出勤
 * ・退勤
 * のリアルタイム操作を記録する。
 *
 * 目的：
 * ・管理者が当日の出退勤状況をリアルタイムに確認できるようにする
 * ・毎日12時に速報メールを送信するための元データにする
 *
 * 注意：
 * ・これは月次勤怠の正式データではない
 * ・attendance_days の実績開始/終了とは分離する
 * ・user_id はリクエストでは受け取らず、JWTから取得する
 * ・同じユーザーが同じ日に同じイベント種別を登録できるのは1回だけ
 * ・登録後の取消・編集はしない
 * ・旧仕様のOTHERデータがDBに残る可能性はある
 */
type AttendanceRealtimeEvent struct {
	ID uint `gorm:"primaryKey" json:"id"`

	/*
	 * 対象ユーザー
	 *
	 * ユーザー側APIではJWTから取得する。
	 */
	UserID uint `gorm:"not null;uniqueIndex:idx_attendance_realtime_event_unique;index" json:"userId"`

	/*
	 * ユーザー情報
	 *
	 * 管理者一覧表示やメール本文作成時にJOINできるようにする。
	 */
	User User `gorm:"foreignKey:UserID" json:"user"`

	/*
	 * イベント日
	 *
	 * JST基準の日付。
	 * 管理者画面で「今日の出退勤」を検索するために使う。
	 */
	EventDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_attendance_realtime_event_unique;index" json:"eventDate"`

	/*
	 * イベント種別
	 *
	 * CLOCK_IN  ：出勤
	 * CLOCK_OUT ：退勤
	 *
	 * OTHERは旧仕様の互換性用データとして
	 * DBに残る可能性がある。
	 */
	EventType string `gorm:"size:30;not null;uniqueIndex:idx_attendance_realtime_event_unique;index" json:"eventType"`

	/*
	 * イベント発生日時
	 *
	 * ボタンを押した実日時。
	 */
	EventAt time.Time `gorm:"not null;index" json:"eventAt"`

	/*
	 * コメント
	 *
	 * 出勤・退勤時の任意コメントを保存する。
	 * 未入力の場合はnilになる。
	 */
	Note *string `gorm:"type:text" json:"note"`

	/*
	 * アクセス情報
	 *
	 * 不正操作や確認用。
	 */
	ClientIP  *string `gorm:"size:100" json:"clientIp"`
	UserAgent *string `gorm:"type:text" json:"userAgent"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
