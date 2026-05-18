package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type PersonalInformationDriveFolderRepository interface {
	FindPersonalInformationDriveFolder(query *gorm.DB) (models.PersonalInformationDriveFolder, results.Result)
}

/*
 * 従業員用 個人情報DriveフォルダRepository
 *
 * 役割：
 * ・DB処理を実行する
 * ・DB処理で発生したエラーはRepositoryでcode/message/detailsを作って返す
 */
type personalInformationDriveFolderRepository struct {
	db *gorm.DB
}

/*
 * PersonalInformationDriveFolderRepository生成
 */
func NewPersonalInformationDriveFolderRepository(db *gorm.DB) PersonalInformationDriveFolderRepository {
	return &personalInformationDriveFolderRepository{db: db}
}

/*
 * 個人情報Driveフォルダ1件取得
 */
func (repository *personalInformationDriveFolderRepository) FindPersonalInformationDriveFolder(query *gorm.DB) (models.PersonalInformationDriveFolder, results.Result) {
	var folder models.PersonalInformationDriveFolder

	if query == nil {
		return folder, results.InternalServerError(
			"FIND_MY_PERSONAL_INFORMATION_DRIVE_FOLDER_EMPTY_QUERY",
			"個人情報Driveフォルダの取得に失敗しました",
			nil,
		)
	}

	if err := query.First(&folder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return folder, results.NotFound(
				"FIND_MY_PERSONAL_INFORMATION_DRIVE_FOLDER_NOT_FOUND",
				"個人情報Driveフォルダが見つかりません",
				nil,
			)
		}

		return folder, results.InternalServerError(
			"FIND_MY_PERSONAL_INFORMATION_DRIVE_FOLDER_FAILED",
			"個人情報Driveフォルダの取得に失敗しました",
			err.Error(),
		)
	}

	return folder, results.OK(nil, "FIND_MY_PERSONAL_INFORMATION_DRIVE_FOLDER_SUCCESS", "", nil)
}
