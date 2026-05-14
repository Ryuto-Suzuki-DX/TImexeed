package types

import "time"

/*
 * 〇 管理者 月次通勤定期 Type
 *
 * 管理者が対象ユーザーの対象年月の通勤定期を扱う型。
 *
 * 重要：
 * ・管理者APIでは対象ユーザーIDを targetUserId としてRequestで受け取る
 * ・userId はJWTから取得しない
 * ・targetUserId + targetYear + targetMonth で対象データを特定する
 * ・管理者側では月次申請状態による編集ロックを行わない
 *
 * user側との差分：
 * ・user側はJWTからログイン中ユーザーIDを取得する
 * ・admin側はrequest bodyのtargetUserIdで対象ユーザーを指定する
 * ・user側は月次申請状態により編集不可になる場合がある
 * ・admin側は月次申請状態に関係なく編集できる
 */

/*
 * 〇 月次通勤定期検索リクエスト
 *
 * 管理者が対象ユーザーの対象年月の通勤定期を取得する。
 */
type SearchMonthlyCommuterPassRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`
}

/*
 * 〇 月次通勤定期更新リクエスト
 *
 * 対象ユーザーの対象年月の通勤定期を更新する。
 *
 * 仕様：
 * ・未登録なら新規作成する
 * ・登録済みなら更新する
 *
 * 注意：
 * ・targetUserId + targetYear + targetMonth で対象データを特定する
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
type UpdateMonthlyCommuterPassRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

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
 * 管理者が対象ユーザーの対象年月の通勤定期を論理削除する。
 *
 * 注意：
 * ・targetUserId + targetYear + targetMonth で対象データを特定する
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
type DeleteMonthlyCommuterPassRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

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

	// 対象ユーザーID
	UserID uint `json:"userId"`

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
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId"`

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
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId"`

	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`
}
