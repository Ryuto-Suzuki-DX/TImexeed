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
	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`
}

/*
 * 〇 月次通勤定期更新リクエスト
 *
 * 対象年月の通勤定期を更新する。
 *
 * 仕様：
 * ・未登録なら新規作成する
 * ・登録済みなら更新する
 *
 * 注意：
 * ・userId / targetUserId は受け取らない
 * ・userID + targetYear + targetMonth で対象データを特定する
 */
type UpdateMonthlyCommuterPassRequest struct {
	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`

	// 定期：出発地
	CommuterFrom *string `json:"commuterFrom"`

	// 定期：目的地
	CommuterTo *string `json:"commuterTo"`

	// 定期：手段
	// 例：電車、バス、車
	CommuterMethod *string `json:"commuterMethod"`

	// 定期：金額
	CommuterAmount *int `json:"commuterAmount"`
}

/*
 * 〇 月次通勤定期削除リクエスト
 *
 * 従業員本人の対象年月の通勤定期を論理削除する。
 *
 * 注意：
 * ・userId / targetUserId は受け取らない
 * ・userID + targetYear + targetMonth で対象データを特定する
 */
type DeleteMonthlyCommuterPassRequest struct {
	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`
}

/*
 * 〇 月次通勤定期レスポンス
 */
type MonthlyCommuterPassResponse struct {
	ID uint `json:"id"`

	// 対象年
	TargetYear int `json:"targetYear"`

	// 対象月
	TargetMonth int `json:"targetMonth"`

	// 定期：出発地
	CommuterFrom *string `json:"commuterFrom"`

	// 定期：目的地
	CommuterTo *string `json:"commuterTo"`

	// 定期：手段
	CommuterMethod *string `json:"commuterMethod"`

	// 定期：金額
	CommuterAmount *int `json:"commuterAmount"`

	// 月次申請状態
	MonthlyStatus string `json:"monthlyStatus"`

	// 論理削除フラグ
	IsDeleted bool `json:"isDeleted"`

	// 作成日時
	CreatedAt time.Time `json:"createdAt"`

	// 更新日時
	UpdatedAt time.Time `json:"updatedAt"`

	// 論理削除日時
	DeletedAt *time.Time `json:"deletedAt"`
}

/*
 * 〇 月次通勤定期検索レスポンス
 *
 * 未登録の場合、monthlyCommuterPass は null で返す。
 */
type SearchMonthlyCommuterPassResponse struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	MonthlyCommuterPass *MonthlyCommuterPassResponse `json:"monthlyCommuterPass"`
}

/*
 * 〇 月次通勤定期更新レスポンス
 */
type UpdateMonthlyCommuterPassResponse struct {
	MonthlyCommuterPass MonthlyCommuterPassResponse `json:"monthlyCommuterPass"`
}

/*
 * 〇 月次通勤定期削除レスポンス
 */
type DeleteMonthlyCommuterPassResponse struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`
}
