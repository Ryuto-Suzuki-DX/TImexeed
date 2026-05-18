package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type PersonalInformationDriveFolderRepository interface {
	FindPersonalInformationDriveFolderRows(query *gorm.DB) ([]types.PersonalInformationDriveFolderSearchRow, results.Result)
	CountPersonalInformationDriveFolderRows(query *gorm.DB) (int64, results.Result)
	FindPersonalInformationDriveFolder(query *gorm.DB) (models.PersonalInformationDriveFolder, results.Result)
	CreatePersonalInformationDriveFolder(folder models.PersonalInformationDriveFolder) (models.PersonalInformationDriveFolder, results.Result)
	SavePersonalInformationDriveFolder(folder models.PersonalInformationDriveFolder) (models.PersonalInformationDriveFolder, results.Result)
	FindUser(query *gorm.DB) (models.User, results.Result)
	FindUsers(query *gorm.DB) ([]models.User, results.Result)
	FindExternalStorageLink(query *gorm.DB) (models.ExternalStorageLink, results.Result)
}

/*
 * 管理者用 個人情報DriveフォルダRepository
 *
 * 役割：
 * ・DB処理を実行する
 * ・DB処理で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・ControllerのRequestは受け取らない
 * ・検索条件の組み立てはBuilderに任せる
 * ・業務判断はServiceに任せる
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
 * 個人情報Driveフォルダ検索
 */
func (repository *personalInformationDriveFolderRepository) FindPersonalInformationDriveFolderRows(query *gorm.DB) ([]types.PersonalInformationDriveFolderSearchRow, results.Result) {
	var rows []types.PersonalInformationDriveFolderSearchRow

	if query == nil {
		return rows, results.InternalServerError(
			"FIND_PERSONAL_INFORMATION_DRIVE_FOLDER_ROWS_EMPTY_QUERY",
			"個人情報Driveフォルダ一覧の取得に失敗しました",
			nil,
		)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return rows, results.InternalServerError(
			"FIND_PERSONAL_INFORMATION_DRIVE_FOLDER_ROWS_FAILED",
			"個人情報Driveフォルダ一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return rows, results.OK(nil, "FIND_PERSONAL_INFORMATION_DRIVE_FOLDER_ROWS_SUCCESS", "", nil)
}

/*
 * 個人情報Driveフォルダ検索件数取得
 */
func (repository *personalInformationDriveFolderRepository) CountPersonalInformationDriveFolderRows(query *gorm.DB) (int64, results.Result) {
	var total int64

	if query == nil {
		return 0, results.InternalServerError(
			"COUNT_PERSONAL_INFORMATION_DRIVE_FOLDER_ROWS_EMPTY_QUERY",
			"個人情報Driveフォルダ件数の取得に失敗しました",
			nil,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_PERSONAL_INFORMATION_DRIVE_FOLDER_ROWS_FAILED",
			"個人情報Driveフォルダ件数の取得に失敗しました",
			err.Error(),
		)
	}

	return total, results.OK(nil, "COUNT_PERSONAL_INFORMATION_DRIVE_FOLDER_ROWS_SUCCESS", "", nil)
}

/*
 * 個人情報Driveフォルダ1件取得
 */
func (repository *personalInformationDriveFolderRepository) FindPersonalInformationDriveFolder(query *gorm.DB) (models.PersonalInformationDriveFolder, results.Result) {
	var folder models.PersonalInformationDriveFolder

	if query == nil {
		return folder, results.InternalServerError(
			"FIND_PERSONAL_INFORMATION_DRIVE_FOLDER_EMPTY_QUERY",
			"個人情報Driveフォルダの取得に失敗しました",
			nil,
		)
	}

	if err := query.First(&folder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return folder, results.NotFound(
				"FIND_PERSONAL_INFORMATION_DRIVE_FOLDER_NOT_FOUND",
				"個人情報Driveフォルダが見つかりません",
				nil,
			)
		}

		return folder, results.InternalServerError(
			"FIND_PERSONAL_INFORMATION_DRIVE_FOLDER_FAILED",
			"個人情報Driveフォルダの取得に失敗しました",
			err.Error(),
		)
	}

	return folder, results.OK(nil, "FIND_PERSONAL_INFORMATION_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * 個人情報Driveフォルダ作成
 */
func (repository *personalInformationDriveFolderRepository) CreatePersonalInformationDriveFolder(folder models.PersonalInformationDriveFolder) (models.PersonalInformationDriveFolder, results.Result) {
	if err := repository.db.Create(&folder).Error; err != nil {
		return folder, results.InternalServerError(
			"CREATE_PERSONAL_INFORMATION_DRIVE_FOLDER_FAILED",
			"個人情報Driveフォルダ情報の登録に失敗しました",
			err.Error(),
		)
	}

	return folder, results.Created(nil, "CREATE_PERSONAL_INFORMATION_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * 個人情報Driveフォルダ保存
 */
func (repository *personalInformationDriveFolderRepository) SavePersonalInformationDriveFolder(folder models.PersonalInformationDriveFolder) (models.PersonalInformationDriveFolder, results.Result) {
	if err := repository.db.Save(&folder).Error; err != nil {
		return folder, results.InternalServerError(
			"SAVE_PERSONAL_INFORMATION_DRIVE_FOLDER_FAILED",
			"個人情報Driveフォルダ情報の保存に失敗しました",
			err.Error(),
		)
	}

	return folder, results.OK(nil, "SAVE_PERSONAL_INFORMATION_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * ユーザー1件取得
 */
func (repository *personalInformationDriveFolderRepository) FindUser(query *gorm.DB) (models.User, results.Result) {
	var user models.User

	if query == nil {
		return user, results.InternalServerError(
			"FIND_USER_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_EMPTY_QUERY",
			"ユーザーの取得に失敗しました",
			nil,
		)
	}

	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, results.NotFound(
				"FIND_USER_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_NOT_FOUND",
				"ユーザーが見つかりません",
				nil,
			)
		}

		return user, results.InternalServerError(
			"FIND_USER_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_FAILED",
			"ユーザーの取得に失敗しました",
			err.Error(),
		)
	}

	return user, results.OK(nil, "FIND_USER_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * ユーザー複数取得
 */
func (repository *personalInformationDriveFolderRepository) FindUsers(query *gorm.DB) ([]models.User, results.Result) {
	var users []models.User

	if query == nil {
		return users, results.InternalServerError(
			"FIND_USERS_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_EMPTY_QUERY",
			"ユーザー一覧の取得に失敗しました",
			nil,
		)
	}

	if err := query.Find(&users).Error; err != nil {
		return users, results.InternalServerError(
			"FIND_USERS_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_FAILED",
			"ユーザー一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return users, results.OK(nil, "FIND_USERS_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * 外部ストレージリンク取得
 */
func (repository *personalInformationDriveFolderRepository) FindExternalStorageLink(query *gorm.DB) (models.ExternalStorageLink, results.Result) {
	var externalStorageLink models.ExternalStorageLink

	if query == nil {
		return externalStorageLink, results.InternalServerError(
			"FIND_EXTERNAL_STORAGE_LINK_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_EMPTY_QUERY",
			"外部ストレージリンクの取得に失敗しました",
			nil,
		)
	}

	if err := query.First(&externalStorageLink).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return externalStorageLink, results.NotFound(
				"FIND_EXTERNAL_STORAGE_LINK_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_NOT_FOUND",
				"個人情報Drive親フォルダ設定が見つかりません",
				map[string]any{
					"linkType": "PERSONAL_INFORMATION_DRIVE_ROOT",
				},
			)
		}

		return externalStorageLink, results.InternalServerError(
			"FIND_EXTERNAL_STORAGE_LINK_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_FAILED",
			"外部ストレージリンクの取得に失敗しました",
			err.Error(),
		)
	}

	return externalStorageLink, results.OK(nil, "FIND_EXTERNAL_STORAGE_LINK_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_SUCCESS", "", nil)
}
