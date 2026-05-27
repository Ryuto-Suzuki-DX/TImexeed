package controllers

import (
	"errors"
	"io"

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
func NewSharedDocumentDriveFolderController(sharedDocumentDriveFolderService services.SharedDocumentDriveFolderService) *SharedDocumentDriveFolderController {
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
			"SEARCH_SHARED_DOCUMENT_DRIVE_FOLDERS_BIND_FAILED",
			"共有資料Driveフォルダ検索条件の読み取りに失敗しました",
			err.Error(),
		))
		return
	}

	responses.JSON(c, controller.sharedDocumentDriveFolderService.SearchSharedDocumentDriveFolders(req))
}

/*
 * 共有資料Driveフォルダ詳細
 */
func (controller *SharedDocumentDriveFolderController) DetailSharedDocumentDriveFolder(c *gin.Context) {
	var req types.SharedDocumentDriveFolderDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"DETAIL_SHARED_DOCUMENT_DRIVE_FOLDER_BIND_FAILED",
			"共有資料Driveフォルダ詳細条件の読み取りに失敗しました",
			err.Error(),
		))
		return
	}

	responses.JSON(c, controller.sharedDocumentDriveFolderService.DetailSharedDocumentDriveFolder(req))
}

/*
 * 共有資料Driveフォルダ作成
 */
func (controller *SharedDocumentDriveFolderController) CreateSharedDocumentDriveFolder(c *gin.Context) {
	var req types.CreateSharedDocumentDriveFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_BIND_FAILED",
			"共有資料Driveフォルダ作成条件の読み取りに失敗しました",
			err.Error(),
		))
		return
	}

	responses.JSON(c, controller.sharedDocumentDriveFolderService.CreateSharedDocumentDriveFolder(req))
}

/*
 * 共有資料Driveフォルダ更新
 */
func (controller *SharedDocumentDriveFolderController) UpdateSharedDocumentDriveFolder(c *gin.Context) {
	var req types.UpdateSharedDocumentDriveFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_BIND_FAILED",
			"共有資料Driveフォルダ更新条件の読み取りに失敗しました",
			err.Error(),
		))
		return
	}

	responses.JSON(c, controller.sharedDocumentDriveFolderService.UpdateSharedDocumentDriveFolder(req))
}

/*
 * 共有資料Driveフォルダ削除
 */
func (controller *SharedDocumentDriveFolderController) DeleteSharedDocumentDriveFolder(c *gin.Context) {
	var req types.DeleteSharedDocumentDriveFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"DELETE_SHARED_DOCUMENT_DRIVE_FOLDER_BIND_FAILED",
			"共有資料Driveフォルダ削除条件の読み取りに失敗しました",
			err.Error(),
		))
		return
	}

	responses.JSON(c, controller.sharedDocumentDriveFolderService.DeleteSharedDocumentDriveFolder(req))
}

/*
 * 共有資料Driveフォルダ権限同期
 *
 * bodyなし、または {} の場合は全件同期する。
 * targetSharedDocumentDriveFolderId が指定された場合は1件だけ同期する。
 */
func (controller *SharedDocumentDriveFolderController) SyncSharedDocumentDriveFolder(c *gin.Context) {
	var req types.SyncSharedDocumentDriveFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		responses.JSON(c, results.BadRequest(
			"SYNC_SHARED_DOCUMENT_DRIVE_FOLDER_BIND_FAILED",
			"共有資料Driveフォルダ同期条件の読み取りに失敗しました",
			err.Error(),
		))
		return
	}

	responses.JSON(c, controller.sharedDocumentDriveFolderService.SyncSharedDocumentDriveFolder(req))
}
