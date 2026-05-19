package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用 共有資料DriveフォルダController
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
 * 検索
 *
 * POST /admin/shared-document-drive-folders/search
 */
func (controller *SharedDocumentDriveFolderController) SearchSharedDocumentDriveFolders(c *gin.Context) {
	var req types.SearchSharedDocumentDriveFoldersRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_SHARED_DOCUMENT_DRIVE_FOLDERS_INVALID_REQUEST",
			"共有資料Driveフォルダ検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.sharedDocumentDriveFolderService.SearchSharedDocumentDriveFolders(req)

	responses.JSON(c, result)
}

/*
 * 詳細
 *
 * POST /admin/shared-document-drive-folders/detail
 */
func (controller *SharedDocumentDriveFolderController) DetailSharedDocumentDriveFolder(c *gin.Context) {
	var req types.SharedDocumentDriveFolderDetailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"DETAIL_SHARED_DOCUMENT_DRIVE_FOLDER_INVALID_REQUEST",
			"共有資料Driveフォルダ詳細のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.sharedDocumentDriveFolderService.DetailSharedDocumentDriveFolder(req)

	responses.JSON(c, result)
}

/*
 * 作成
 *
 * POST /admin/shared-document-drive-folders/create
 */
func (controller *SharedDocumentDriveFolderController) CreateSharedDocumentDriveFolder(c *gin.Context) {
	var req types.CreateSharedDocumentDriveFolderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_INVALID_REQUEST",
			"共有資料Driveフォルダ作成のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.sharedDocumentDriveFolderService.CreateSharedDocumentDriveFolder(req)

	responses.JSON(c, result)
}

/*
 * 更新
 *
 * POST /admin/shared-document-drive-folders/update
 */
func (controller *SharedDocumentDriveFolderController) UpdateSharedDocumentDriveFolder(c *gin.Context) {
	var req types.UpdateSharedDocumentDriveFolderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_INVALID_REQUEST",
			"共有資料Driveフォルダ更新のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.sharedDocumentDriveFolderService.UpdateSharedDocumentDriveFolder(req)

	responses.JSON(c, result)
}

/*
 * 削除
 *
 * POST /admin/shared-document-drive-folders/delete
 */
func (controller *SharedDocumentDriveFolderController) DeleteSharedDocumentDriveFolder(c *gin.Context) {
	var req types.DeleteSharedDocumentDriveFolderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"DELETE_SHARED_DOCUMENT_DRIVE_FOLDER_INVALID_REQUEST",
			"共有資料Driveフォルダ削除のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.sharedDocumentDriveFolderService.DeleteSharedDocumentDriveFolder(req)

	responses.JSON(c, result)
}

/*
 * 共有ユーザー更新
 *
 * POST /admin/shared-document-drive-folders/users/update
 */
func (controller *SharedDocumentDriveFolderController) UpdateSharedDocumentDriveFolderUsers(c *gin.Context) {
	var req types.UpdateSharedDocumentDriveFolderUsersRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_USERS_INVALID_REQUEST",
			"共有資料Driveフォルダ共有ユーザー更新のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.sharedDocumentDriveFolderService.UpdateSharedDocumentDriveFolderUsers(req)

	responses.JSON(c, result)
}

/*
 * 同期
 *
 * POST /admin/shared-document-drive-folders/sync
 */
func (controller *SharedDocumentDriveFolderController) SyncSharedDocumentDriveFolder(c *gin.Context) {
	var req types.SyncSharedDocumentDriveFolderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SYNC_SHARED_DOCUMENT_DRIVE_FOLDER_INVALID_REQUEST",
			"共有資料Driveフォルダ同期のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.sharedDocumentDriveFolderService.SyncSharedDocumentDriveFolder(req)

	responses.JSON(c, result)
}
