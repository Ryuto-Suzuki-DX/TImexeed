package models

import "time"

/*
 * 〇 各日の勤怠
 *
 * 1日ごとの予定・実績・日別交通費を管理するメインテーブル。
 *
 * このテーブルに入れるもの：
 * 	・予定区分
 * 	・予定時間
 * 	・実績区分
 * 	・実績時間
 * 	・遅刻、早退、欠勤、病欠などの補足
 * 	・在宅勤務補助対象フラグ
 * 	・日別交通費
 *
 * このテーブルに入れないもの：
 * 	・月次申請状態
 * 	・月次承認状態
 * 	・申請メモ
 * 	・承認者
 * 	・承認日時
 * 	・否認理由
 * 	・画面表示用システムメッセージ
 *
 * 理由：
 * 	勤怠日別レコードは、あくまで「その日の勤怠データ」を持つ。
 * 	月次申請・承認の状態は MonthlyAttendanceRequest を見て判断する。
 *
 * 	画面表示用システムメッセージは、
 * 	保存データではなく、予定・実績・休憩・有給申請状態などから
 * 	画面表示時に計算して作る。
 *
 * 休憩は1日に複数回あるため、AttendanceBreak に分ける。
 * 月次通勤定期は月単位なので、MonthlyCommuterPass に分ける。
 */
type AttendanceDay struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// 対象ユーザーID
	UserID uint `gorm:"not null;index" json:"userId"`

	// 対象日
	WorkDate time.Time `gorm:"type:date;not null;index" json:"workDate"`

	// 予定区分ID
	PlanAttendanceTypeID uint `gorm:"not null" json:"planAttendanceTypeId"`

	// 実績区分ID
	ActualAttendanceTypeID uint `gorm:"not null" json:"actualAttendanceTypeId"`

	// 予定開始日時
	PlanStartAt *time.Time `json:"planStartAt"`

	// 予定終了日時
	PlanEndAt *time.Time `json:"planEndAt"`

	// 実績開始日時
	ActualStartAt *time.Time `json:"actualStartAt"`

	// 実績終了日時
	ActualEndAt *time.Time `json:"actualEndAt"`

	// 遅刻フラグ
	LateFlag bool `gorm:"not null;default:false" json:"lateFlag"`

	// 早退フラグ
	EarlyLeaveFlag bool `gorm:"not null;default:false" json:"earlyLeaveFlag"`

	// 欠勤フラグ
	AbsenceFlag bool `gorm:"not null;default:false" json:"absenceFlag"`

	// 病欠フラグ
	SickLeaveFlag bool `gorm:"not null;default:false" json:"sickLeaveFlag"`

	// 在宅勤務補助対象フラグ
	// 他の派遣会社に勤めていて、且つ在宅勤務の場合に従業員が選択する。
	RemoteWorkAllowanceFlag bool `gorm:"not null;default:false" json:"remoteWorkAllowanceFlag"`

	// 日別交通費：出発地
	TransportFrom *string `gorm:"type:varchar(100)" json:"transportFrom"`

	// 日別交通費：目的地
	TransportTo *string `gorm:"type:varchar(100)" json:"transportTo"`

	// 日別交通費：手段
	// 例：電車、バス、徒歩、車
	TransportMethod *string `gorm:"type:varchar(50)" json:"transportMethod"`

	// 日別交通費：金額
	TransportAmount *int `json:"transportAmount"`

	// 論理削除フラグ
	IsDeleted bool `gorm:"not null;default:false" json:"isDeleted"`

	// 作成日時
	CreatedAt time.Time `json:"createdAt"`

	// 更新日時
	UpdatedAt time.Time `json:"updatedAt"`

	// 論理削除日時
	DeletedAt *time.Time `json:"deletedAt"`

	// 予定区分
	PlanAttendanceType AttendanceType `gorm:"foreignKey:PlanAttendanceTypeID" json:"planAttendanceType,omitempty"`

	// 実績区分
	ActualAttendanceType AttendanceType `gorm:"foreignKey:ActualAttendanceTypeID" json:"actualAttendanceType,omitempty"`
}
