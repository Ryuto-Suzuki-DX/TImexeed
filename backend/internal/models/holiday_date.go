package models

import "time"

/*
 * 〇 祝日マスタ
 *
 * 国民の祝日CSVなどを取り込み、
 * 勤怠画面で休日判定に使うためのマスタ。
 *
 * このテーブルに入れるもの：
 * 	・祝日の日付
 * 	・祝日名
 *
 * このテーブルに入れないもの：
 * 	・土曜日
 * 	・日曜日
 * 	・会社独自休日
 * 	・ユーザーごとの休み
 *
 * 理由：
 * 	土日は日付から判定できる。
 * 	会社独自休日やユーザーごとの休みは、
 * 	将来的に別マスタで管理する可能性があるため。
 *
 * 補足：
 * 	祝日は年ごとに変わるため、
 * 	CSVインポートでDBに登録する。
 *
 * 	CSV取り込み時は date をキーにして upsert する。
 */
type HolidayDate struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// 祝日の日付
	// 例：2026-01-01
	HolidayDate time.Time `gorm:"type:date;not null;uniqueIndex" json:"holidayDate"`

	// 祝日名
	// 例：元日、成人の日、建国記念の日
	HolidayName string `gorm:"type:varchar(100);not null" json:"holidayName"`

	// 論理削除フラグ
	IsDeleted bool `gorm:"not null;default:false" json:"isDeleted"`

	// 作成日時
	CreatedAt time.Time `json:"createdAt"`

	// 更新日時
	UpdatedAt time.Time `json:"updatedAt"`

	// 論理削除日時
	DeletedAt *time.Time `json:"deletedAt"`
}
