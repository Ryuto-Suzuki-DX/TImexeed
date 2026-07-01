package models

import "time"

/*
 * 〇 日別交通費
 *
 * AttendanceDayに紐づく日別交通費明細を管理する。
 *
 * 1日に複数件登録できるため、AttendanceDayとは別テーブルにする。
 *
 * 例：
 * ・自宅 → 最寄駅：バス 220円
 * ・最寄駅 → 派遣先：電車 480円
 *
 * 注意：
 * ・月次通勤定期はMonthlyCommuterPassで管理する
 * ・このテーブルは都度発生した日別交通費だけを管理する
 * ・表示順はSortOrderで管理する
 * ・削除は論理削除とする
 */
type AttendanceTransportExpense struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// 対象勤怠日ID
	AttendanceDayID uint `gorm:"not null;index" json:"attendanceDayId"`

	// 表示順
	SortOrder int `gorm:"not null;default:0" json:"sortOrder"`

	// 出発地
	TransportFrom string `gorm:"type:varchar(100);not null" json:"transportFrom"`

	// 目的地
	TransportTo string `gorm:"type:varchar(100);not null" json:"transportTo"`

	// 交通手段
	// 例：電車、バス、徒歩、車、タクシー
	TransportMethod string `gorm:"type:varchar(50);not null" json:"transportMethod"`

	// 金額
	TransportAmount int `gorm:"not null;default:0" json:"transportAmount"`

	// 備考
	TransportMemo *string `gorm:"type:varchar(255)" json:"transportMemo"`

	// 論理削除フラグ
	IsDeleted bool `gorm:"not null;default:false;index" json:"isDeleted"`

	// 作成日時
	CreatedAt time.Time `json:"createdAt"`

	// 更新日時
	UpdatedAt time.Time `json:"updatedAt"`

	// 論理削除日時
	DeletedAt *time.Time `json:"deletedAt"`

	// 対象勤怠日
	AttendanceDay AttendanceDay `gorm:"foreignKey:AttendanceDayID" json:"attendanceDay,omitempty"`
}
