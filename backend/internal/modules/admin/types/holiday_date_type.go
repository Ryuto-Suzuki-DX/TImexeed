package types

import "time"

/*
 * 〇 管理者 祝日 Type
 *
 * 管理者が祝日CSVを取り込み、
 * 登録済み祝日を対象年月ごとに確認するための型。
 *
 * 役割：
 * ・CSV取り込み
 * ・対象年月の祝日検索
 * ・祝日マスタのレスポンス整形
 *
 * 注意：
 * ・祝日マスタ自体は全ユーザー共通
 * ・従業員側では祝日の登録、更新、削除は行わない
 * ・管理者側だけがCSV取り込みを行う
 * ・CSV取り込み時は既存データを物理削除し、新しいCSV内容を全件登録する
 */

/*
 * 祝日CSVインポートRequest
 *
 * 管理者画面でCSVファイルを選択し、
 * フロント側でファイル内容を文字列として読み取ってから送信する。
 *
 * つまり、バックエンドではmultipart/form-dataではなく、
 * 通常のJSONとして受け取る。
 *
 * 想定：
 * {
 *   "csvText": "国民の祝日・休日月日,国民の祝日・休日名称\n2026/1/1,元日"
 * }
 */
type ImportHolidayDatesRequest struct {
	CsvText string `json:"csvText"`
}

/*
 * 祝日検索Request
 *
 * 管理者画面で現在登録されている祝日を確認するため、
 * 対象年月の祝日一覧を取得する。
 *
 * 用途：
 * ・CSV取り込み前の確認
 * ・CSV取り込み後の確認
 * ・設定画面で月ごとの登録済み祝日を表示する
 */
type SearchHolidayDatesRequest struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`
}

/*
 * 祝日Response
 */
type HolidayDateResponse struct {
	ID uint `json:"id"`

	HolidayDate time.Time `json:"holidayDate"`
	HolidayName string    `json:"holidayName"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

/*
 * 祝日CSVインポートResponse
 *
 * DeletedCount：
 * ・CSV取り込み前に削除した既存祝日件数
 *
 * ImportedCount：
 * ・CSVから新規登録した件数
 *
 * SkippedCount：
 * ・ヘッダー行、空行、不正行、重複行などでスキップした件数
 */
type ImportHolidayDatesResponse struct {
	DeletedCount  int `json:"deletedCount"`
	ImportedCount int `json:"importedCount"`
	SkippedCount  int `json:"skippedCount"`
}

/*
 * 祝日検索Response
 *
 * 対象年月に登録されている祝日一覧を返す。
 */
type SearchHolidayDatesResponse struct {
	Holidays []HolidayDateResponse `json:"holidays"`
}
