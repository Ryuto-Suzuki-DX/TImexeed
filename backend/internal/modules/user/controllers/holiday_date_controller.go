package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用祝日Controller
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * このControllerで扱うもの：
 * ・対象年月の祝日検索
 *
 * このControllerで扱わないもの：
 * ・祝日の登録
 * ・祝日の更新
 * ・祝日の削除
 * ・CSV取り込み
 * ・DB処理
 * ・業務ルール
 *
 * 注意：
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・Requestを別の型へ詰め直さない
 * ・c.JSONは直接使わず responses.JSON を使う
 * ・従業員APIでは userId / targetUserId を request body で受け取らない
 *
 * 祝日管理方針：
 * ・祝日マスタは全ユーザー共通
 * ・従業員側では祝日を参照するだけ
 * ・CSV取り込みや祝日マスタ更新は管理者側APIで行う
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
 * 祝日検索
 *
 * POST /user/holiday-dates/search
 *
 * 用途：
 * ・対象年月の祝日一覧を取得する
 * ・従業員の月次勤怠画面で土日祝判定に使う
 *
 * 仕様：
 * ・フロントから userId / targetUserId は送らない
 * ・祝日は全ユーザー共通のため、ログイン中ユーザーIDによる絞り込みは行わない
 * ・AuthMiddlewareによりログイン済みであることはroute側で保証する
 *
 * 注意：
 * ・従業員側では祝日の登録、更新、削除は行わない
 * ・CSV取り込みは管理者側APIで行う
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

	// bindしたRequest型をServiceへ渡す
	result := controller.holidayDateService.SearchHolidayDates(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
