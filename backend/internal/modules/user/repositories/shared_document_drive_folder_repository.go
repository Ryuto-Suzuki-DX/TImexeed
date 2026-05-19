package repositories

import (
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type SharedDocumentDriveFolderRepository interface {
	FindSharedDocumentDriveFolderRows(query *gorm.DB) ([]types.SharedDocumentDriveFolderRow, results.Result)
	CountSharedDocumentDriveFolderRows(query *gorm.DB) (int64, results.Result)
	FindSharedDocumentDriveFolderRow(query *gorm.DB) (types.SharedDocumentDriveFolderRow, results.Result)
}

/*
 * 従業員用 共有資料DriveフォルダRepository
 *
 * 役割：
 * ・DB処理を実行する
 * ・本人に共有されている資料だけ、Builderで作成されたクエリを実行する
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
 * 共有資料Driveフォルダ一覧取得
 */
func (repository *sharedDocumentDriveFolderRepository) FindSharedDocumentDriveFolderRows(
	query *gorm.DB,
) ([]types.SharedDocumentDriveFolderRow, results.Result) {
	var rows []types.SharedDocumentDriveFolderRow

	if query == nil {
		return rows, results.InternalServerError(
			"FIND_MY_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_EMPTY_QUERY",
			"共有資料Driveフォルダ一覧の取得に失敗しました",
			nil,
		)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return rows, results.InternalServerError(
			"FIND_MY_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_FAILED",
			"共有資料Driveフォルダ一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return rows, results.OK(nil, "FIND_MY_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_SUCCESS", "", nil)
}

/*
 * 共有資料Driveフォルダ件数取得
 */
func (repository *sharedDocumentDriveFolderRepository) CountSharedDocumentDriveFolderRows(
	query *gorm.DB,
) (int64, results.Result) {
	var total int64

	if query == nil {
		return 0, results.InternalServerError(
			"COUNT_MY_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_EMPTY_QUERY",
			"共有資料Driveフォルダ件数の取得に失敗しました",
			nil,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_MY_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_FAILED",
			"共有資料Driveフォルダ件数の取得に失敗しました",
			err.Error(),
		)
	}

	return total, results.OK(nil, "COUNT_MY_SHARED_DOCUMENT_DRIVE_FOLDER_ROWS_SUCCESS", "", nil)
}

/*
 * 共有資料Driveフォルダ1件取得
 */
func (repository *sharedDocumentDriveFolderRepository) FindSharedDocumentDriveFolderRow(
	query *gorm.DB,
) (types.SharedDocumentDriveFolderRow, results.Result) {
	var row types.SharedDocumentDriveFolderRow

	if query == nil {
		return row, results.InternalServerError(
			"FIND_MY_SHARED_DOCUMENT_DRIVE_FOLDER_ROW_EMPTY_QUERY",
			"共有資料Driveフォルダの取得に失敗しました",
			nil,
		)
	}

	if err := query.Scan(&row).Error; err != nil {
		return row, results.InternalServerError(
			"FIND_MY_SHARED_DOCUMENT_DRIVE_FOLDER_ROW_FAILED",
			"共有資料Driveフォルダの取得に失敗しました",
			err.Error(),
		)
	}

	if row.ID == 0 {
		return row, results.NotFound(
			"FIND_MY_SHARED_DOCUMENT_DRIVE_FOLDER_ROW_NOT_FOUND",
			"共有資料Driveフォルダが見つかりません",
			nil,
		)
	}

	return row, results.OK(nil, "FIND_MY_SHARED_DOCUMENT_DRIVE_FOLDER_ROW_SUCCESS", "", nil)
}
