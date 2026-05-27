package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用 共有資料DriveフォルダRepository interface
 */
type SharedDocumentDriveFolderRepository interface {
	FindSharedDocumentDriveFolderRows(query *gorm.DB) ([]types.SharedDocumentDriveFolderSearchRow, results.Result)
	CountSharedDocumentDriveFolderRows(query *gorm.DB) (int64, results.Result)

	FindSharedDocumentDriveFolder(query *gorm.DB) (models.SharedDocumentDriveFolder, results.Result)
	FindSharedDocumentDriveFolders(query *gorm.DB) ([]models.SharedDocumentDriveFolder, results.Result)
	CreateSharedDocumentDriveFolder(folder models.SharedDocumentDriveFolder) (models.SharedDocumentDriveFolder, results.Result)
	SaveSharedDocumentDriveFolder(folder models.SharedDocumentDriveFolder) (models.SharedDocumentDriveFolder, results.Result)

	FindExternalStorageLink(query *gorm.DB) (models.ExternalStorageLink, results.Result)
	FindUsers(query *gorm.DB) ([]models.User, results.Result)
}

/*
 * 管理者用 共有資料DriveフォルダRepository
 */
type sharedDocumentDriveFolderRepository struct {
	db *gorm.DB
}

/*
 * SharedDocumentDriveFolderRepository生成
 */
func NewSharedDocumentDriveFolderRepository(db *gorm.DB) SharedDocumentDriveFolderRepository {
	return &sharedDocumentDriveFolderRepository{db: db}
}

/*
 * 共有資料Driveフォルダ検索
 */
func (repository *sharedDocumentDriveFolderRepository) FindSharedDocumentDriveFolderRows(query *gorm.DB) ([]types.SharedDocumentDriveFolderSearchRow, results.Result) {
	var rows []types.SharedDocumentDriveFolderSearchRow

	if query == nil {
		return rows, results.InternalServerError(
			"FIND_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_EMPTY_QUERY",
			"共有資料Driveフォルダ一覧の取得に失敗しました",
			nil,
		)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return rows, results.InternalServerError(
			"FIND_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_FAILED",
			"共有資料Driveフォルダ一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return rows, results.OK(nil, "FIND_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_SUCCESS", "", nil)
}

/*
 * 共有資料Driveフォルダ検索件数取得
 */
func (repository *sharedDocumentDriveFolderRepository) CountSharedDocumentDriveFolderRows(query *gorm.DB) (int64, results.Result) {
	var total int64

	if query == nil {
		return 0, results.InternalServerError(
			"COUNT_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_EMPTY_QUERY",
			"共有資料Driveフォルダ件数の取得に失敗しました",
			nil,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_FAILED",
			"共有資料Driveフォルダ件数の取得に失敗しました",
			err.Error(),
		)
	}

	return total, results.OK(nil, "COUNT_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_SUCCESS", "", nil)
}

/*
 * 共有資料Driveフォルダ1件取得
 */
func (repository *sharedDocumentDriveFolderRepository) FindSharedDocumentDriveFolder(query *gorm.DB) (models.SharedDocumentDriveFolder, results.Result) {
	var folder models.SharedDocumentDriveFolder

	if query == nil {
		return folder, results.InternalServerError(
			"FIND_SHARED_DOCUMENT_DRIVE_FOLDER_EMPTY_QUERY",
			"共有資料Driveフォルダの取得に失敗しました",
			nil,
		)
	}

	if err := query.First(&folder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return folder, results.NotFound(
				"FIND_SHARED_DOCUMENT_DRIVE_FOLDER_NOT_FOUND",
				"共有資料Driveフォルダが見つかりません",
				nil,
			)
		}

		return folder, results.InternalServerError(
			"FIND_SHARED_DOCUMENT_DRIVE_FOLDER_FAILED",
			"共有資料Driveフォルダの取得に失敗しました",
			err.Error(),
		)
	}

	return folder, results.OK(nil, "FIND_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * 共有資料Driveフォルダ複数取得
 */
func (repository *sharedDocumentDriveFolderRepository) FindSharedDocumentDriveFolders(query *gorm.DB) ([]models.SharedDocumentDriveFolder, results.Result) {
	var folders []models.SharedDocumentDriveFolder

	if query == nil {
		return folders, results.InternalServerError(
			"FIND_SHARED_DOCUMENT_DRIVE_FOLDERS_EMPTY_QUERY",
			"共有資料Driveフォルダ一覧の取得に失敗しました",
			nil,
		)
	}

	if err := query.Find(&folders).Error; err != nil {
		return folders, results.InternalServerError(
			"FIND_SHARED_DOCUMENT_DRIVE_FOLDERS_FAILED",
			"共有資料Driveフォルダ一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return folders, results.OK(nil, "FIND_SHARED_DOCUMENT_DRIVE_FOLDERS_SUCCESS", "", nil)
}

/*
 * 共有資料Driveフォルダ作成
 */
func (repository *sharedDocumentDriveFolderRepository) CreateSharedDocumentDriveFolder(folder models.SharedDocumentDriveFolder) (models.SharedDocumentDriveFolder, results.Result) {
	if err := repository.db.Create(&folder).Error; err != nil {
		return folder, results.InternalServerError(
			"CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_FAILED",
			"共有資料Driveフォルダ情報の登録に失敗しました",
			err.Error(),
		)
	}

	return folder, results.Created(nil, "CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * 共有資料Driveフォルダ保存
 */
func (repository *sharedDocumentDriveFolderRepository) SaveSharedDocumentDriveFolder(folder models.SharedDocumentDriveFolder) (models.SharedDocumentDriveFolder, results.Result) {
	if err := repository.db.Save(&folder).Error; err != nil {
		return folder, results.InternalServerError(
			"SAVE_SHARED_DOCUMENT_DRIVE_FOLDER_FAILED",
			"共有資料Driveフォルダ情報の保存に失敗しました",
			err.Error(),
		)
	}

	return folder, results.OK(nil, "SAVE_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * 外部ストレージリンク1件取得
 */
func (repository *sharedDocumentDriveFolderRepository) FindExternalStorageLink(query *gorm.DB) (models.ExternalStorageLink, results.Result) {
	var externalStorageLink models.ExternalStorageLink

	if query == nil {
		return externalStorageLink, results.InternalServerError(
			"FIND_EXTERNAL_STORAGE_LINK_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_EMPTY_QUERY",
			"共有資料Drive親フォルダ設定の取得に失敗しました",
			nil,
		)
	}

	if err := query.First(&externalStorageLink).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return externalStorageLink, results.NotFound(
				"SHARED_DOCUMENT_DRIVE_ROOT_EXTERNAL_STORAGE_LINK_NOT_FOUND",
				"共有資料Drive親フォルダ設定が見つかりません",
				nil,
			)
		}

		return externalStorageLink, results.InternalServerError(
			"FIND_EXTERNAL_STORAGE_LINK_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_FAILED",
			"共有資料Drive親フォルダ設定の取得に失敗しました",
			err.Error(),
		)
	}

	return externalStorageLink, results.OK(nil, "FIND_EXTERNAL_STORAGE_LINK_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * ユーザー複数取得
 */
func (repository *sharedDocumentDriveFolderRepository) FindUsers(query *gorm.DB) ([]models.User, results.Result) {
	var users []models.User

	if query == nil {
		return users, results.InternalServerError(
			"FIND_USERS_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_EMPTY_QUERY",
			"ユーザー一覧の取得に失敗しました",
			nil,
		)
	}

	if err := query.Find(&users).Error; err != nil {
		return users, results.InternalServerError(
			"FIND_USERS_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_FAILED",
			"ユーザー一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return users, results.OK(nil, "FIND_USERS_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS", "", nil)
}
