package types

import "time"

/*
 * 〇 従業員 祝日 Type
 *
 * 従業員画面で対象年月の祝日を取得するための型。
 *
 * 役割：
 * 	・対象年月の祝日検索
 * 	・勤怠画面で土日祝判定に使う祝日情報の返却
 *
 * 注意：
 * 	・従業員側では祝日の登録、更新、削除は行わない
 * 	・CSV取り込みは管理者側APIで行う
 * 	・祝日マスタ自体は全ユーザー共通
 */

/*
 * 祝日検索Request
 *
 * 対象年月の祝日一覧を取得する。
 *
 * 例：
 * 	targetYear: 2026
 * 	targetMonth: 5
 */
type SearchHolidayDatesRequest struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`
}

/*
 * 祝日Response
 *
 * holidayDate:
 * 	祝日の日付
 *
 * holidayName:
 * 	祝日名
 */
type HolidayDateResponse struct {
	ID uint `json:"id"`

	HolidayDate time.Time `json:"holidayDate"`
	HolidayName string    `json:"holidayName"`
}

/*
 * 祝日検索Response
 *
 * 対象年月に該当する祝日一覧を返す。
 */
type SearchHolidayDatesResponse struct {
	Holidays []HolidayDateResponse `json:"holidays"`
}
