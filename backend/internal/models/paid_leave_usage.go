package models

import "time"

/*
 * 〇 有給使用履歴
 *
 * ユーザーごとの有給使用日数を管理する。
 *
 * 主な用途：
 * ・システム導入前の有給使用分を管理者が手動登録する
 * ・勤怠・有給申請から確定した有給使用分を保存する
 * ・有給残数計算時に、使用済み日数として集計する
 *
 * 使用日数：
 * ・1日休暇   → 1.0
 * ・半日休暇 → 0.5
 *
 * 注意：
 * ・有給残数そのものはこのテーブルに保存しない
 * ・残数は、付与日数 - 使用日数 の形でService側で計算する
 * ・削除は物理削除ではなく論理削除で行う
 */
type PaidLeaveUsage struct {
	ID uint `gorm:"primaryKey" json:"id"`

	/*
	 * 対象ユーザーID
	 *
	 * 管理者APIでは targetUserId を受け取り、
	 * この UserID に紐づく有給使用履歴を追加・更新・削除する。
	 */
	UserID uint `gorm:"not null;index" json:"userId"`

	/*
	 * ユーザー情報
	 *
	 * 一覧表示や管理者画面で名前を参照したい場合に使う。
	 */
	User User `gorm:"foreignKey:UserID" json:"user"`

	/*
	 * 有給を使用した日
	 */
	UsageDate time.Time `gorm:"type:date;not null;index" json:"usageDate"`

	/*
	 * 使用日数
	 *
	 * 例：
	 * 	1.0 = 1日
	 * 	0.5 = 半日
	 */
	UsageDays float64 `gorm:"type:numeric(4,1);not null" json:"usageDays"`

	/*
	 * 手動追加フラグ
	 *
	 * true:
	 * 	管理者が手動で追加した有給使用履歴
	 *
	 * false:
	 * 	勤怠・有給申請など、システム処理から作成された有給使用履歴
	 */
	IsManual bool `gorm:"not null;default:false" json:"isManual"`

	/*
	 * メモ
	 *
	 * 例：
	 * ・システム導入前使用分
	 * ・管理者調整
	 * ・半休
	 */
	Memo string `gorm:"type:varchar(255)" json:"memo"`

	/*
	 * 論理削除
	 */
	IsDeleted bool       `gorm:"not null;default:false" json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}
