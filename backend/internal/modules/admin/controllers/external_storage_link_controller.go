package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用外部ストレージリンクController
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
 */
type ExternalStorageLinkController struct {
	externalStorageLinkService services.ExternalStorageLinkService
}

/*
 * ExternalStorageLinkController生成
 */
func NewExternalStorageLinkController(externalStorageLinkService services.ExternalStorageLinkService) *ExternalStorageLinkController {
	return &ExternalStorageLinkController{
		externalStorageLinkService: externalStorageLinkService,
	}
}

/*
 * 検索
 *
 * POST /admin/external-storage-links/search
 */
func (controller *ExternalStorageLinkController) SearchExternalStorageLinks(c *gin.Context) {
	var req types.SearchExternalStorageLinksRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_EXTERNAL_STORAGE_LINKS_INVALID_REQUEST",
			"外部ストレージリンク検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.externalStorageLinkService.SearchExternalStorageLinks(req)
	responses.JSON(c, result)
}

/*
 * 取得
 *
 * POST /admin/external-storage-links/detail
 */
func (controller *ExternalStorageLinkController) GetExternalStorageLinkDetail(c *gin.Context) {
	var req types.ExternalStorageLinkDetailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"GET_EXTERNAL_STORAGE_LINK_DETAIL_INVALID_REQUEST",
			"外部ストレージリンク詳細取得のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.externalStorageLinkService.GetExternalStorageLinkDetail(req)
	responses.JSON(c, result)
}

/*
 * 新規作成
 *
 * POST /admin/external-storage-links/create
 */
func (controller *ExternalStorageLinkController) CreateExternalStorageLink(c *gin.Context) {
	var req types.CreateExternalStorageLinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"CREATE_EXTERNAL_STORAGE_LINK_INVALID_REQUEST",
			"外部ストレージリンク作成のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.externalStorageLinkService.CreateExternalStorageLink(req)
	responses.JSON(c, result)
}

/*
 * 更新
 *
 * POST /admin/external-storage-links/update
 */
func (controller *ExternalStorageLinkController) UpdateExternalStorageLink(c *gin.Context) {
	var req types.UpdateExternalStorageLinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"UPDATE_EXTERNAL_STORAGE_LINK_INVALID_REQUEST",
			"外部ストレージリンク更新のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.externalStorageLinkService.UpdateExternalStorageLink(req)
	responses.JSON(c, result)
}

/*
 * 論理削除
 *
 * POST /admin/external-storage-links/delete
 */
func (controller *ExternalStorageLinkController) DeleteExternalStorageLink(c *gin.Context) {
	var req types.DeleteExternalStorageLinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"DELETE_EXTERNAL_STORAGE_LINK_INVALID_REQUEST",
			"外部ストレージリンク削除のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.externalStorageLinkService.DeleteExternalStorageLink(req)
	responses.JSON(c, result)
}
