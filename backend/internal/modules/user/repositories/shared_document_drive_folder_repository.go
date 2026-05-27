package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 従業員用 共有資料DriveフォルダRepository interface
 */
type SharedDocumentDriveFolderRepository interface {
	FindSharedDocumentDriveFolderRows(query *gorm.DB) ([]types.SharedDocumentDriveFolderSearchRow, results.Result)
	CountSharedDocumentDriveFolderRows(query *gorm.DB) (int64, results.Result)
	FindSharedDocumentDriveFolder(query *gorm.DB) (models.SharedDocumentDriveFolder, results.Result)
}

/*
 * 従業員用 共有資料DriveフォルダRepository
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
			"FIND_USER_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_EMPTY_QUERY",
			"共有資料Driveフォルダ一覧の取得に失敗しました",
			nil,
		)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return rows, results.InternalServerError(
			"FIND_USER_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_FAILED",
			"共有資料Driveフォルダ一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return rows, results.OK(
		nil,
		"FIND_USER_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 共有資料Driveフォルダ検索件数取得
 */
func (repository *sharedDocumentDriveFolderRepository) CountSharedDocumentDriveFolderRows(query *gorm.DB) (int64, results.Result) {
	var total int64

	if query == nil {
		return 0, results.InternalServerError(
			"COUNT_USER_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_EMPTY_QUERY",
			"共有資料Driveフォルダ件数の取得に失敗しました",
			nil,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_USER_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_FAILED",
			"共有資料Driveフォルダ件数の取得に失敗しました",
			err.Error(),
		)
	}

	return total, results.OK(
		nil,
		"COUNT_USER_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 共有資料Driveフォルダ1件取得
 */
func (repository *sharedDocumentDriveFolderRepository) FindSharedDocumentDriveFolder(query *gorm.DB) (models.SharedDocumentDriveFolder, results.Result) {
	var folder models.SharedDocumentDriveFolder

	if query == nil {
		return folder, results.InternalServerError(
			"FIND_USER_SHARED_DOCUMENT_DRIVE_FOLDER_EMPTY_QUERY",
			"共有資料Driveフォルダの取得に失敗しました",
			nil,
		)
	}

	if err := query.First(&folder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return folder, results.NotFound(
				"FIND_USER_SHARED_DOCUMENT_DRIVE_FOLDER_NOT_FOUND",
				"共有資料Driveフォルダが見つかりません",
				nil,
			)
		}

		return folder, results.InternalServerError(
			"FIND_USER_SHARED_DOCUMENT_DRIVE_FOLDER_FAILED",
			"共有資料Driveフォルダの取得に失敗しました",
			err.Error(),
		)
	}

	return folder, results.OK(
		nil,
		"FIND_USER_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS",
		"",
		nil,
	)
}
