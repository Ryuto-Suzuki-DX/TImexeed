package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用 共有資料DriveフォルダController
 *
 * 従業員側では閲覧のみ。
 * 作成・更新・削除・Drive権限同期は管理者側APIで行う。
 */
type SharedDocumentDriveFolderController struct {
	sharedDocumentDriveFolderService services.SharedDocumentDriveFolderService
}

/*
 * SharedDocumentDriveFolderController生成
 */
func NewSharedDocumentDriveFolderController(
	sharedDocumentDriveFolderService services.SharedDocumentDriveFolderService,
) *SharedDocumentDriveFolderController {
	return &SharedDocumentDriveFolderController{
		sharedDocumentDriveFolderService: sharedDocumentDriveFolderService,
	}
}

/*
 * 共有資料Driveフォルダ検索
 */
func (controller *SharedDocumentDriveFolderController) SearchSharedDocumentDriveFolders(c *gin.Context) {
	var req types.SearchSharedDocumentDriveFoldersRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_USER_SHARED_DOCUMENT_DRIVE_FOLDERS_INVALID_REQUEST",
			"共有資料Driveフォルダ検索リクエストが正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.sharedDocumentDriveFolderService.SearchSharedDocumentDriveFolders(req)
	responses.JSON(c, result)
}

/*
 * 共有資料Driveフォルダ詳細
 */
func (controller *SharedDocumentDriveFolderController) DetailSharedDocumentDriveFolder(c *gin.Context) {
	var req types.SharedDocumentDriveFolderDetailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"DETAIL_USER_SHARED_DOCUMENT_DRIVE_FOLDER_INVALID_REQUEST",
			"共有資料Driveフォルダ詳細リクエストが正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.sharedDocumentDriveFolderService.DetailSharedDocumentDriveFolder(req)
	responses.JSON(c, result)
}
