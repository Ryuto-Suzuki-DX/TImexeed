package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
)

/*
 * 従業員用 個人情報DriveフォルダService interface
 */
type PersonalInformationDriveFolderService interface {
	GetMyPersonalInformationDriveFolder(userID uint) results.Result
}

/*
 * 従業員用 個人情報DriveフォルダService
 *
 * 役割：
 * ・Controllerから受け取ったログインユーザーIDをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリを作成する
 * ・RepositoryでDB処理を実行する
 */
type personalInformationDriveFolderService struct {
	personalInformationDriveFolderBuilder    builders.PersonalInformationDriveFolderBuilder
	personalInformationDriveFolderRepository repositories.PersonalInformationDriveFolderRepository
}

/*
 * PersonalInformationDriveFolderService生成
 */
func NewPersonalInformationDriveFolderService(
	personalInformationDriveFolderBuilder builders.PersonalInformationDriveFolderBuilder,
	personalInformationDriveFolderRepository repositories.PersonalInformationDriveFolderRepository,
) *personalInformationDriveFolderService {
	return &personalInformationDriveFolderService{
		personalInformationDriveFolderBuilder:    personalInformationDriveFolderBuilder,
		personalInformationDriveFolderRepository: personalInformationDriveFolderRepository,
	}
}

/*
 * models.PersonalInformationDriveFolderをResponseへ変換する
 */
func toPersonalInformationDriveFolderResponse(folder models.PersonalInformationDriveFolder) types.PersonalInformationDriveFolderResponse {
	return types.PersonalInformationDriveFolderResponse{
		ID: folder.ID,

		UserID: folder.UserID,

		FolderName:    folder.FolderName,
		DriveFolderID: folder.DriveFolderID,
		FolderURL:     folder.FolderURL,
		SyncedAt:      folder.SyncedAt,

		CreatedAt: folder.CreatedAt,
		UpdatedAt: folder.UpdatedAt,
	}
}

/*
 * 自分の個人情報Driveフォルダ取得
 *
 * ユーザー側は検索不要。
 * JWTから取得した本人userIdのフォルダだけ返す。
 */
func (service *personalInformationDriveFolderService) GetMyPersonalInformationDriveFolder(userID uint) results.Result {
	query, buildResult := service.personalInformationDriveFolderBuilder.BuildFindActivePersonalInformationDriveFolderByUserIDQuery(userID)
	if buildResult.Error {
		return buildResult
	}

	folder, findResult := service.personalInformationDriveFolderRepository.FindPersonalInformationDriveFolder(query)
	if findResult.Error {
		return findResult
	}

	return results.OK(
		types.GetMyPersonalInformationDriveFolderResponse{
			PersonalInformationDriveFolder: toPersonalInformationDriveFolderResponse(folder),
		},
		"GET_MY_PERSONAL_INFORMATION_DRIVE_FOLDER_SUCCESS",
		"個人情報Driveフォルダを取得しました",
		nil,
	)
}
