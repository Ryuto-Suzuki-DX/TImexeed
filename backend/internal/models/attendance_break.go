package models

import "time"

/*
 * 〇 各日の休憩
 *
 * 1日の中で複数回発生する休憩を管理する。
 *
 * 例：
 * 	・12:00 - 13:00
 * 	・15:00 - 15:15
 * 	・22:00 - 22:30
 *
 * 夜勤も考えるため、時刻だけではなく日時で持つ。
 */
type AttendanceBreak struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// 紐づく勤怠日ID
	AttendanceDayID uint `gorm:"not null;index" json:"attendanceDayId"`

	// 休憩開始日時
	BreakStartAt time.Time `gorm:"not null" json:"breakStartAt"`

	// 休憩終了日時
	BreakEndAt time.Time `gorm:"not null" json:"breakEndAt"`

	// 休憩メモ
	BreakMemo *string `gorm:"type:text" json:"breakMemo"`

	// 論理削除フラグ
	IsDeleted bool `gorm:"not null;default:false" json:"isDeleted"`

	// 作成日時
	CreatedAt time.Time `json:"createdAt"`

	// 更新日時
	UpdatedAt time.Time `json:"updatedAt"`

	// 論理削除日時
	DeletedAt *time.Time `json:"deletedAt"`

	// 勤怠日
	AttendanceDay AttendanceDay `gorm:"foreignKey:AttendanceDayID" json:"attendanceDay,omitempty"`
}
