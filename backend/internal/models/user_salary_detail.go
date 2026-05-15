package models

import (
	"time"

	"gorm.io/gorm"
)

/*
 * 〇 ユーザー給与詳細
 *
 * ユーザーごとの給与計算に使う基本情報を管理する。
 *
 * 注意：
 * ・給与計算結果ではない
 * ・CSV出力結果ではない
 * ・経費精算情報ではない
 * ・残業/深夜/休日などの割増率は持たない
 * ・所定労働時間は持たない
 * ・通勤手当上限や在宅勤務補助は会社全体設定側で管理する
 * ・ここでは個人ごとに違う給与区分・基本金額・固定手当/固定控除だけを持つ
 */
type UserSalaryDetail struct {
	ID uint `gorm:"primaryKey"`

	UserID uint `gorm:"not null;index"`
	// 対象ユーザーID

	SalaryType string `gorm:"type:varchar(20);not null"`
	// MONTHLY: 月給
	// HOURLY: 時給
	// DAILY: 日給

	BaseAmount int `gorm:"not null;default:0"`
	// 月給なら月額
	// 時給なら時給
	// 日給なら日給

	ExtraAllowanceAmount int `gorm:"not null;default:0"`
	// その他固定手当
	// 例：役職手当、資格手当、住宅手当など
	// 経費ではなく、給与として毎月加算する固定的な手当

	ExtraAllowanceMemo string `gorm:"type:text"`
	// その他固定手当メモ

	FixedDeductionAmount int `gorm:"not null;default:0"`
	// その他固定控除
	// 給与から毎月差し引く固定的な控除

	FixedDeductionMemo string `gorm:"type:text"`
	// その他固定控除メモ

	IsPayrollTarget bool `gorm:"not null;default:true"`
	// 給与計算対象にするか

	EffectiveFrom time.Time `gorm:"type:date;not null"`
	// 適用開始日

	EffectiveTo *time.Time `gorm:"type:date"`
	// 適用終了日
	// nullなら現在も有効

	Memo string `gorm:"type:text"`
	// 全体メモ

	IsDeleted bool `gorm:"not null;default:false"`
	// 論理削除

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
