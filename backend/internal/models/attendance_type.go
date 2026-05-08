package models

import "time"

/*
 * 〇 勤怠区分マスタ
 *
 * 通常勤務、有給、欠勤、病欠、特別休暇、介護休業、育児休業、休職など、
 * 勤怠入力で選択する区分を管理するマスタ。
 *
 * このマスタは、フロントとバックエンドの両方で使う。
 *
 * フロント：
 * 	・入力欄を出すか隠すか
 * 	・予定と実績を分けて入力するか
 * 	・共通時間だけ入力させるか
 *
 * バックエンド：
 * 	・予定と実績を同期するか
 * 	・遅刻、早退、欠勤、病欠などを無効化するか
 * 	・休憩や交通費を許可するか
 */
type AttendanceType struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// システム用コード
	// 例：WORK, PAID_LEAVE, CAREGIVER_LEAVE, SUSPENSION
	Code string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`

	// 画面表示名
	// 例：通常勤務、有給、介護休業、休職
	Name string `gorm:"type:varchar(100);not null" json:"name"`

	// 区分カテゴリ
	// 例：WORK, LEAVE, HOLIDAY, ABSENCE, SUSPENSION
	Category string `gorm:"type:varchar(50);not null" json:"category"`

	// 勤務扱いにするか
	// 通常勤務は true、有給や休職などは会社ルールにより false/true を決める
	IsWorked bool `gorm:"not null;default:false" json:"isWorked"`

	// 申請が必要か
	// 有給、特別休暇、介護休業などは true 想定
	RequiresRequest bool `gorm:"not null;default:false" json:"requiresRequest"`

	// 予定と実績を同じ内容で保存するか
	// true の場合、バックエンドで予定時間を実績時間にもコピーする
	// 例：有給、欠勤、病欠、休職、介護休業など
	SyncPlanActual bool `gorm:"not null;default:false" json:"syncPlanActual"`

	// 実績時間を個別入力できるか
	// 通常勤務は true、有給などの同期対象は false 想定
	AllowActualTimeInput bool `gorm:"not null;default:true" json:"allowActualTimeInput"`

	// 休憩入力できるか
	// 通常勤務は true、有給や欠勤などは false 想定
	AllowBreakInput bool `gorm:"not null;default:true" json:"allowBreakInput"`

	// 日別交通費入力できるか
	// 通常勤務は true、有給や休職などは基本 false 想定
	AllowTransportInput bool `gorm:"not null;default:true" json:"allowTransportInput"`

	// 遅刻入力できるか
	AllowLateFlag bool `gorm:"not null;default:true" json:"allowLateFlag"`

	// 早退入力できるか
	AllowEarlyLeaveFlag bool `gorm:"not null;default:true" json:"allowEarlyLeaveFlag"`

	// 欠勤入力できるか
	AllowAbsenceFlag bool `gorm:"not null;default:true" json:"allowAbsenceFlag"`

	// 病欠入力できるか
	AllowSickLeaveFlag bool `gorm:"not null;default:true" json:"allowSickLeaveFlag"`

	// 画面表示順
	DisplayOrder int `gorm:"not null;default:0" json:"displayOrder"`

	// 使用中か
	// 削除せず非表示にしたい場合に使う
	IsActive bool `gorm:"not null;default:true" json:"isActive"`

	// 論理削除フラグ
	IsDeleted bool `gorm:"not null;default:false" json:"isDeleted"`

	// 作成日時
	CreatedAt time.Time `json:"createdAt"`

	// 更新日時
	UpdatedAt time.Time `json:"updatedAt"`

	// 論理削除日時
	DeletedAt *time.Time `json:"deletedAt"`
}
