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
 * 役割：
 * ・JWT認証後にgin.Contextへ入っているuserIdを取得する
 * ・Requestをbindする
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * 注意：
 * ・targetUserIdをrequest bodyでは受け取らない
 * ・本人userIdはJWTから取得する
 * ・共有されていない資料は取得できない
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
 * 自分に共有された共有資料Driveフォルダ検索
 *
 * POST /user/shared-document-drive-folders/search
 */
func (controller *SharedDocumentDriveFolderController) SearchSharedDocumentDriveFolders(c *gin.Context) {
	userID, ok := getLoginUserIDFromContext(c)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"SEARCH_MY_SHARED_DOCUMENT_DRIVE_FOLDERS_UNAUTHORIZED",
			"ログインユーザー情報を取得できません",
			nil,
		))
		return
	}

	var req types.SearchSharedDocumentDriveFoldersRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_MY_SHARED_DOCUMENT_DRIVE_FOLDERS_INVALID_REQUEST",
			"共有資料Driveフォルダ検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.sharedDocumentDriveFolderService.SearchSharedDocumentDriveFolders(userID, req)

	responses.JSON(c, result)
}

/*
 * 自分に共有された共有資料Driveフォルダ詳細
 *
 * POST /user/shared-document-drive-folders/detail
 */
func (controller *SharedDocumentDriveFolderController) DetailSharedDocumentDriveFolder(c *gin.Context) {
	userID, ok := getLoginUserIDFromContext(c)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"DETAIL_MY_SHARED_DOCUMENT_DRIVE_FOLDER_UNAUTHORIZED",
			"ログインユーザー情報を取得できません",
			nil,
		))
		return
	}

	var req types.SharedDocumentDriveFolderDetailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"DETAIL_MY_SHARED_DOCUMENT_DRIVE_FOLDER_INVALID_REQUEST",
			"共有資料Driveフォルダ詳細のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.sharedDocumentDriveFolderService.DetailSharedDocumentDriveFolder(userID, req)

	responses.JSON(c, result)
}
