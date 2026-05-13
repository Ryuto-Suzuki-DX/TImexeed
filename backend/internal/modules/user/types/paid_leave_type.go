package types

import "time"

/*
 * 従業員用 有給 型定義
 *
 * このファイルには、従業員自身が確認する有給情報の型をまとめる。
 *
 * 対象API：
 * ・現時点の有給残数取得
 *
 * 方針：
 * ・従業員APIでは targetUserId を受け取らない
 * ・対象ユーザーIDは JWT から取得する
 * ・Responseの日付は string に変換せず time.Time / *time.Time のまま返す
 * ・表示形式 yyyy-MM-dd などはフロント側で整形する
 */

/*
 * =========================================================
 * Request
 * =========================================================
 */

/*
 * 有給残数取得Request
 *
 * GET /user/paid-leave/balance
 *
 * request body は不要。
 */

/*
 * =========================================================
 * Response
 * =========================================================
 */

/*
 * 有給残数Response
 *
 * 現時点の有給残数を返す。
 */
type PaidLeaveBalanceResponse struct {
	UserID uint `json:"userId"`

	TotalGrantedDays float64 `json:"totalGrantedDays"`
	UsedDays         float64 `json:"usedDays"`
	RemainingDays    float64 `json:"remainingDays"`

	NextGrantDate *time.Time `json:"nextGrantDate"`
	NextGrantDays float64    `json:"nextGrantDays"`

	RequiredUseDays          float64    `json:"requiredUseDays"`
	RequiredUseDeadline      *time.Time `json:"requiredUseDeadline"`
	RequiredUseRemainingDays float64    `json:"requiredUseRemainingDays"`
}
