package types

import "time"

/*
 * 〇 管理者 月次通勤定期検索リクエスト
 */
type SearchMonthlyCommuterPassRequest struct {
	TargetUserID uint `form:"targetUserId" json:"targetUserId" binding:"required"`
	TargetYear   int  `form:"targetYear" json:"targetYear" binding:"required"`
	TargetMonth  int  `form:"targetMonth" json:"targetMonth" binding:"required"`
}

/*
 * 〇 管理者 月次通勤定期レスポンス
 */
type MonthlyCommuterPassResponse struct {
	ID     uint `json:"id"`
	UserID uint `json:"userId"`

	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	CommuterFrom   *string `json:"commuterFrom"`
	CommuterTo     *string `json:"commuterTo"`
	CommuterMethod *string `json:"commuterMethod"`
	CommuterAmount *int    `json:"commuterAmount"`

	IsDeleted bool       `json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

/*
 * 〇 管理者 月次通勤定期検索レスポンス
 *
 * 同じユーザー・対象年月の有効な通勤定期をすべて返す。
 */
type SearchMonthlyCommuterPassResponse struct {
	TargetUserID uint `json:"targetUserId"`
	TargetYear   int  `json:"targetYear"`
	TargetMonth  int  `json:"targetMonth"`

	MonthlyCommuterPasses []MonthlyCommuterPassResponse `json:"monthlyCommuterPasses"`
	TotalCommuterAmount   int                           `json:"totalCommuterAmount"`
}

/*
 * 〇 管理者 月次通勤定期差分保存リクエスト
 *
 * 画面に残っている通勤定期だけを送る。
 * ・monthlyCommuterPassIdあり：更新
 * ・monthlyCommuterPassIdなし：新規作成
 * ・DBに存在するがRequestにないID：論理削除
 */
type UpdateMonthlyCommuterPassesRequest struct {
	TargetUserID uint `json:"targetUserId" binding:"required"`
	TargetYear   int  `json:"targetYear" binding:"required"`
	TargetMonth  int  `json:"targetMonth" binding:"required"`

	CommuterPasses []UpdateMonthlyCommuterPassItemRequest `json:"commuterPasses"`
}

/*
 * 〇 管理者 月次通勤定期差分保存：明細
 */
type UpdateMonthlyCommuterPassItemRequest struct {
	MonthlyCommuterPassID *uint `json:"monthlyCommuterPassId"`

	CommuterFrom   *string `json:"commuterFrom"`
	CommuterTo     *string `json:"commuterTo"`
	CommuterMethod *string `json:"commuterMethod"`
	CommuterAmount *int    `json:"commuterAmount"`
}

/*
 * 〇 管理者 月次通勤定期差分保存レスポンス
 */
type UpdateMonthlyCommuterPassesResponse struct {
	TargetUserID uint `json:"targetUserId"`
	TargetYear   int  `json:"targetYear"`
	TargetMonth  int  `json:"targetMonth"`

	MonthlyCommuterPasses         []MonthlyCommuterPassResponse `json:"monthlyCommuterPasses"`
	SavedMonthlyCommuterPassCount int                           `json:"savedMonthlyCommuterPassCount"`
	TotalCommuterAmount           int                           `json:"totalCommuterAmount"`
}

/*
 * 〇 旧単体更新リクエスト（互換用）
 *
 * 新しい月次勤怠全体保存では UpdateMonthlyCommuterPassesRequest を使用する。
 */
type UpdateMonthlyCommuterPassRequest struct {
	TargetUserID uint `json:"targetUserId" binding:"required"`
	TargetYear   int  `json:"targetYear" binding:"required"`
	TargetMonth  int  `json:"targetMonth" binding:"required"`

	CommuterFrom   *string `json:"commuterFrom"`
	CommuterTo     *string `json:"commuterTo"`
	CommuterMethod *string `json:"commuterMethod"`
	CommuterAmount *int    `json:"commuterAmount"`
}

/*
 * 〇 旧単体更新レスポンス（互換用）
 */
type UpdateMonthlyCommuterPassResponse struct {
	MonthlyCommuterPass MonthlyCommuterPassResponse `json:"monthlyCommuterPass"`
}

/*
 * 〇 管理者 月次通勤定期削除リクエスト
 *
 * 対象年月の有効な通勤定期をすべて論理削除する。
 */
type DeleteMonthlyCommuterPassRequest struct {
	TargetUserID uint `json:"targetUserId" binding:"required"`
	TargetYear   int  `json:"targetYear" binding:"required"`
	TargetMonth  int  `json:"targetMonth" binding:"required"`
}

/*
 * 〇 管理者 月次通勤定期削除レスポンス
 */
type DeleteMonthlyCommuterPassResponse struct {
	TargetUserID uint `json:"targetUserId"`
	TargetYear   int  `json:"targetYear"`
	TargetMonth  int  `json:"targetMonth"`

	DeletedMonthlyCommuterPassCount int `json:"deletedMonthlyCommuterPassCount"`
}
