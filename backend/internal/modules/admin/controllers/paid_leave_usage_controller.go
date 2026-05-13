package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用有給使用日Controller
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * 注意：
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・Requestを別の型へ詰め直さない
 * ・c.JSONは直接使わず responses.JSON を使う
 *
 * エラー方針：
 * ・Controllerで発生したエラーはControllerでcode/messageを決める
 * ・Serviceで発生したエラーはServiceでcode/messageを決める
 * ・Builderで発生したエラーはBuilderでcode/messageを決める
 * ・Repositoryで発生したエラーはRepositoryでcode/messageを決める
 * ・Controllerは最終的に responses.JSON で返す
 */
type PaidLeaveUsageController struct {
	paidLeaveUsageService services.PaidLeaveUsageService
}

/*
 * PaidLeaveUsageController生成
 */
func NewPaidLeaveUsageController(paidLeaveUsageService services.PaidLeaveUsageService) *PaidLeaveUsageController {
	return &PaidLeaveUsageController{
		paidLeaveUsageService: paidLeaveUsageService,
	}
}

/*
 * 有給使用日検索
 *
 * POST /admin/paid-leave-usages/search
 */
func (controller *PaidLeaveUsageController) SearchPaidLeaveUsages(c *gin.Context) {
	var req types.SearchPaidLeaveUsagesRequest

	// リクエストJSONをSearchPaidLeaveUsagesRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"SEARCH_PAID_LEAVE_USAGES_INVALID_REQUEST",
			"有給使用日検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.paidLeaveUsageService.SearchPaidLeaveUsages(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 有給残数取得
 *
 * POST /admin/paid-leave-usages/balance
 */
func (controller *PaidLeaveUsageController) GetPaidLeaveBalance(c *gin.Context) {
	var req types.GetPaidLeaveBalanceRequest

	// リクエストJSONをGetPaidLeaveBalanceRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"GET_PAID_LEAVE_BALANCE_INVALID_REQUEST",
			"有給残数取得のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.paidLeaveUsageService.GetPaidLeaveBalance(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 過去有給使用日追加
 *
 * POST /admin/paid-leave-usages/create
 */
func (controller *PaidLeaveUsageController) CreatePaidLeaveUsage(c *gin.Context) {
	var req types.CreatePaidLeaveUsageRequest

	// リクエストJSONをCreatePaidLeaveUsageRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"CREATE_PAID_LEAVE_USAGE_INVALID_REQUEST",
			"有給使用日作成のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.paidLeaveUsageService.CreatePaidLeaveUsage(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 過去有給使用日編集
 *
 * POST /admin/paid-leave-usages/update
 */
func (controller *PaidLeaveUsageController) UpdatePaidLeaveUsage(c *gin.Context) {
	var req types.UpdatePaidLeaveUsageRequest

	// リクエストJSONをUpdatePaidLeaveUsageRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"UPDATE_PAID_LEAVE_USAGE_INVALID_REQUEST",
			"有給使用日更新のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.paidLeaveUsageService.UpdatePaidLeaveUsage(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 過去有給使用日削除
 *
 * POST /admin/paid-leave-usages/delete
 */
func (controller *PaidLeaveUsageController) DeletePaidLeaveUsage(c *gin.Context) {
	var req types.DeletePaidLeaveUsageRequest

	// リクエストJSONをDeletePaidLeaveUsageRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"DELETE_PAID_LEAVE_USAGE_INVALID_REQUEST",
			"有給使用日削除のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.paidLeaveUsageService.DeletePaidLeaveUsage(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
