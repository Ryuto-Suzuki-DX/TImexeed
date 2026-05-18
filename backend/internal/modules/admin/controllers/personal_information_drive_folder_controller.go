package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用 個人情報DriveフォルダController
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
type PersonalInformationDriveFolderController struct {
	personalInformationDriveFolderService services.PersonalInformationDriveFolderService
}

/*
 * PersonalInformationDriveFolderController生成
 */
func NewPersonalInformationDriveFolderController(
	personalInformationDriveFolderService services.PersonalInformationDriveFolderService,
) *PersonalInformationDriveFolderController {
	return &PersonalInformationDriveFolderController{
		personalInformationDriveFolderService: personalInformationDriveFolderService,
	}
}

/*
 * 検索
 *
 * POST /admin/personal-information-drive-folders/search
 */
func (controller *PersonalInformationDriveFolderController) SearchPersonalInformationDriveFolders(c *gin.Context) {
	var req types.SearchPersonalInformationDriveFoldersRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_PERSONAL_INFORMATION_DRIVE_FOLDERS_INVALID_REQUEST",
			"個人情報Driveフォルダ検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.personalInformationDriveFolderService.SearchPersonalInformationDriveFolders(req)

	responses.JSON(c, result)
}

/*
 * 作成/権限同期
 *
 * POST /admin/personal-information-drive-folders/sync
 */
func (controller *PersonalInformationDriveFolderController) SyncPersonalInformationDriveFolder(c *gin.Context) {
	var req types.SyncPersonalInformationDriveFolderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SYNC_PERSONAL_INFORMATION_DRIVE_FOLDER_INVALID_REQUEST",
			"個人情報Driveフォルダ更新のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.personalInformationDriveFolderService.SyncPersonalInformationDriveFolder(req)

	responses.JSON(c, result)
}

/*
 * 表示
 *
 * POST /admin/personal-information-drive-folders/view
 */
func (controller *PersonalInformationDriveFolderController) ViewPersonalInformationDriveFolder(c *gin.Context) {
	var req types.ViewPersonalInformationDriveFolderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"VIEW_PERSONAL_INFORMATION_DRIVE_FOLDER_INVALID_REQUEST",
			"個人情報Driveフォルダ表示のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.personalInformationDriveFolderService.ViewPersonalInformationDriveFolder(req)

	responses.JSON(c, result)
}
