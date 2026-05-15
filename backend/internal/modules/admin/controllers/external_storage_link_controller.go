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
 * ・固定された外部ストレージリンク一覧を取得する
 * ・固定された外部ストレージリンクのURL/説明/メモを更新する
 *
 * 注意：
 * ・管理者が任意にリンクを新規作成/削除する運用にはしない
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
