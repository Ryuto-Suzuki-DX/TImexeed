package models

import "time"

/*
 * 〇 お知らせ自動リマインド設定
 *
 * 管理者が作成する「毎月自動でお知らせを作成するためのルール」を管理する。
 *
 * 例：
 * ・月末3日前の9:00に「月次勤怠申請をお願いします」を全従業員へ送る
 * ・月末当日の18:00に「勤怠申請の締切日です」を全従業員へ送る
 *
 * 注意：
 * ・このテーブル自体は、実際のお知らせ本文ではない
 * ・実際にユーザーへ表示されるお知らせは notifications テーブルへ作成する
 * ・メール通知はここでは扱わない
 * ・自動実行処理は別途バッチ、cron、schedulerなどで行う想定
 */
type NotificationReminder struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// タイトル
	Title string `gorm:"type:varchar(150);not null" json:"title"`

	// 本文
	Message string `gorm:"type:text;not null" json:"message"`

	/*
	 * 月末から何日前に送るか
	 *
	 * 0：月末当日
	 * 1：月末1日前
	 * 3：月末3日前
	 *
	 * 例：
	 * 2026年5月の場合、月末は5/31
	 * DayOffsetFromMonthEnd = 3 なら 5/28 に送る
	 */
	DayOffsetFromMonthEnd int `gorm:"not null;default:0" json:"dayOffsetFromMonthEnd"`

	/*
	 * 送信予定時刻：時
	 *
	 * 0〜23 を想定する。
	 */
	SendHour int `gorm:"not null;default:9" json:"sendHour"`

	/*
	 * 送信予定時刻：分
	 *
	 * 0〜59 を想定する。
	 */
	SendMinute int `gorm:"not null;default:0" json:"sendMinute"`

	// 有効フラグ
	IsEnabled bool `gorm:"not null;default:true" json:"isEnabled"`

	// 論理削除フラグ
	IsDeleted bool `gorm:"not null;default:false" json:"isDeleted"`

	// 作成日時
	CreatedAt time.Time `json:"createdAt"`

	// 更新日時
	UpdatedAt time.Time `json:"updatedAt"`

	// 論理削除日時
	DeletedAt *time.Time `json:"deletedAt"`
}
