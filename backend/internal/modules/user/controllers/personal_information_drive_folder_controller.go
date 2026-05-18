package controllers

import (
	"fmt"

	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用 個人情報DriveフォルダController
 *
 * 役割：
 * ・JWT認証後にgin.Contextへ入っているuserIdを取得する
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * 注意：
 * ・ユーザー側は検索しない
 * ・targetUserIdをrequest bodyでは受け取らない
 * ・本人userIdはJWTから取得する
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
 * 自分の個人情報Driveフォルダ取得
 *
 * POST /user/personal-information-drive-folders/get
 */
func (controller *PersonalInformationDriveFolderController) GetMyPersonalInformationDriveFolder(c *gin.Context) {
	userID, ok := getLoginUserIDFromContext(c)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"GET_MY_PERSONAL_INFORMATION_DRIVE_FOLDER_UNAUTHORIZED",
			"ログインユーザー情報を取得できません",
			nil,
		))
		return
	}

	result := controller.personalInformationDriveFolderService.GetMyPersonalInformationDriveFolder(userID)

	responses.JSON(c, result)
}

/*
 * gin.ContextからログインユーザーIDを取得する。
 *
 * AuthMiddleware側で userId をセットしている前提。
 * 既存ミドルウェアの実装差異に耐えるため、主要な数値型を受ける。
 */
func getLoginUserIDFromContext(c *gin.Context) (uint, bool) {
	value, exists := c.Get("userId")
	if !exists {
		return 0, false
	}

	switch typedValue := value.(type) {
	case uint:
		return typedValue, typedValue != 0
	case uint64:
		return uint(typedValue), typedValue != 0
	case int:
		return uint(typedValue), typedValue > 0
	case int64:
		return uint(typedValue), typedValue > 0
	case float64:
		return uint(typedValue), typedValue > 0
	default:
		_ = fmt.Sprintf("%v", typedValue)
		return 0, false
	}
}
