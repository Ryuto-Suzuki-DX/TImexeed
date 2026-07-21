package types

import "time"

/*
 * 〇 月次通勤定期検索リクエスト
 *
 * 従業員本人の対象年月の通勤定期を取得する。
 *
 * 注意：
 * ・userId / targetUserId は受け取らない
 * ・ログイン中ユーザーIDはControllerでJWTから取得してServiceへ渡す
 */
type SearchMonthlyCommuterPassRequest struct {
	TargetYear  int `form:"targetYear" json:"targetYear" binding:"required"`
	TargetMonth int `form:"targetMonth" json:"targetMonth" binding:"required"`
}

/*
 * 〇 月次通勤定期レスポンス
 */
type MonthlyCommuterPassResponse struct {
	ID uint `json:"id"`

	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	CommuterFrom   *string `json:"commuterFrom"`
	CommuterTo     *string `json:"commuterTo"`
	CommuterMethod *string `json:"commuterMethod"`
	CommuterAmount *int    `json:"commuterAmount"`

	// 対象月の月次申請状態
	MonthlyStatus string `json:"monthlyStatus"`

	IsDeleted bool       `json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

/*
 * 〇 月次通勤定期検索レスポンス
 *
 * 未登録の場合も monthlyCommuterPasses は空配列で返す。
 */
type SearchMonthlyCommuterPassResponse struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	MonthlyStatus         string                        `json:"monthlyStatus"`
	MonthlyCommuterPasses []MonthlyCommuterPassResponse `json:"monthlyCommuterPasses"`
	TotalCommuterAmount   int                           `json:"totalCommuterAmount"`
}

/*
 * 〇 月次通勤定期差分保存リクエスト
 *
 * 画面に残っている通勤定期だけを送る。
 * ・monthlyCommuterPassIdあり：更新
 * ・monthlyCommuterPassIdなし：新規作成
 * ・DBに存在するがRequestにないID：論理削除
 */
type UpdateMonthlyCommuterPassesRequest struct {
	TargetYear  int `json:"targetYear" binding:"required"`
	TargetMonth int `json:"targetMonth" binding:"required"`

	CommuterPasses []UpdateMonthlyCommuterPassItemRequest `json:"commuterPasses"`
}

/*
 * 〇 月次通勤定期差分保存：明細
 */
type UpdateMonthlyCommuterPassItemRequest struct {
	MonthlyCommuterPassID *uint `json:"monthlyCommuterPassId"`

	CommuterFrom   *string `json:"commuterFrom"`
	CommuterTo     *string `json:"commuterTo"`
	CommuterMethod *string `json:"commuterMethod"`
	CommuterAmount *int    `json:"commuterAmount"`
}

/*
 * 〇 月次通勤定期差分保存レスポンス
 */
type UpdateMonthlyCommuterPassesResponse struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	MonthlyStatus                 string                        `json:"monthlyStatus"`
	MonthlyCommuterPasses         []MonthlyCommuterPassResponse `json:"monthlyCommuterPasses"`
	SavedMonthlyCommuterPassCount int                           `json:"savedMonthlyCommuterPassCount"`
	TotalCommuterAmount           int                           `json:"totalCommuterAmount"`
}

/*
 * 〇 旧単体更新リクエスト（互換用）
 *
 * 新しい月次勤怠全体保存では UpdateMonthlyCommuterPassesRequest を使う。
 */
type UpdateMonthlyCommuterPassRequest struct {
	TargetYear  int `json:"targetYear" binding:"required"`
	TargetMonth int `json:"targetMonth" binding:"required"`

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
 * 〇 月次通勤定期削除リクエスト
 *
 * 対象年月の有効な通勤定期をすべて論理削除する。
 */
type DeleteMonthlyCommuterPassRequest struct {
	TargetYear  int `json:"targetYear" binding:"required"`
	TargetMonth int `json:"targetMonth" binding:"required"`
}

/*
 * 〇 月次通勤定期削除レスポンス
 */
type DeleteMonthlyCommuterPassResponse struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	DeletedMonthlyCommuterPassCount int `json:"deletedMonthlyCommuterPassCount"`
}
