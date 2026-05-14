package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用祝日Controller
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * このControllerで扱うもの：
 * ・祝日CSV取り込み
 * ・対象年月の祝日検索
 *
 * このControllerで扱わないもの：
 * ・祝日の単体作成
 * ・祝日の単体更新
 * ・祝日の単体削除
 * ・DB処理
 * ・業務ルール
 *
 * 注意：
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・Requestを別の型へ詰め直さない
 * ・c.JSONは直接使わず responses.JSON を使う
 *
 * 祝日管理方針：
 * ・祝日マスタは全ユーザー共通
 * ・管理者側だけが祝日CSVを取り込める
 * ・従業員側では祝日を参照するだけ
 *
 * エラー方針：
 * ・Controllerで発生したエラーはControllerでcode/messageを決める
 * ・Serviceで発生したエラーはServiceでcode/messageを決める
 * ・Builderで発生したエラーはBuilderでcode/messageを決める
 * ・Repositoryで発生したエラーはRepositoryでcode/messageを決める
 * ・Controllerは最終的に responses.JSON で返す
 */
type HolidayDateController struct {
	holidayDateService services.HolidayDateService
}

/*
 * HolidayDateController生成
 */
func NewHolidayDateController(holidayDateService services.HolidayDateService) *HolidayDateController {
	return &HolidayDateController{
		holidayDateService: holidayDateService,
	}
}

/*
 * 祝日CSV取り込み
 *
 * POST /admin/holiday-dates/import
 *
 * 用途：
 * ・管理者がCSVファイルの内容を取り込む
 * ・holiday_dates に祝日データを登録、または更新する
 *
 * 仕様：
 * ・multipart/form-data ではなく JSON で受け取る
 * ・フロント側でCSVファイルを文字列として読み取り、csvText に詰めて送る
 */
func (controller *HolidayDateController) ImportHolidayDates(c *gin.Context) {
	var req types.ImportHolidayDatesRequest

	// リクエストJSONをImportHolidayDatesRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"IMPORT_HOLIDAY_DATES_INVALID_REQUEST",
			"祝日CSV取り込みのリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.holidayDateService.ImportHolidayDates(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 祝日検索
 *
 * POST /admin/holiday-dates/search
 *
 * 用途：
 * ・管理者画面で対象年月の祝日一覧を取得する
 * ・CSV取り込み結果を確認する
 *
 * 仕様：
 * ・祝日は全ユーザー共通のため、ユーザーIDでは絞り込まない
 */
func (controller *HolidayDateController) SearchHolidayDates(c *gin.Context) {
	var req types.SearchHolidayDatesRequest

	// リクエストJSONをSearchHolidayDatesRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_HOLIDAY_DATES_INVALID_REQUEST",
			"祝日検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.holidayDateService.SearchHolidayDates(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
