package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type SharedDocumentDriveFolderRepository interface {
	FindSharedDocumentDriveFolderRows(query *gorm.DB) ([]types.SharedDocumentDriveFolderSearchRow, results.Result)
	CountSharedDocumentDriveFolderRows(query *gorm.DB) (int64, results.Result)

	FindSharedDocumentDriveFolder(query *gorm.DB) (models.SharedDocumentDriveFolder, results.Result)
	CreateSharedDocumentDriveFolder(folder models.SharedDocumentDriveFolder) (models.SharedDocumentDriveFolder, results.Result)
	SaveSharedDocumentDriveFolder(folder models.SharedDocumentDriveFolder) (models.SharedDocumentDriveFolder, results.Result)

	FindSharedDocumentDriveFolderUserRows(query *gorm.DB) ([]types.SharedDocumentDriveFolderUserResponse, results.Result)
	FindSharedDocumentDriveFolderUsers(query *gorm.DB) ([]models.SharedDocumentDriveFolderUser, results.Result)
	CreateSharedDocumentDriveFolderUser(folderUser models.SharedDocumentDriveFolderUser) (models.SharedDocumentDriveFolderUser, results.Result)
	SaveSharedDocumentDriveFolderUser(folderUser models.SharedDocumentDriveFolderUser) (models.SharedDocumentDriveFolderUser, results.Result)

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
 * 共有ユーザー表示用取得
 */
func (repository *sharedDocumentDriveFolderRepository) FindSharedDocumentDriveFolderUserRows(query *gorm.DB) ([]types.SharedDocumentDriveFolderUserResponse, results.Result) {
	var rows []types.SharedDocumentDriveFolderUserResponse

	if query == nil {
		return rows, results.InternalServerError(
			"FIND_SHARED_DOCUMENT_DRIVE_FOLDER_USER_ROWS_EMPTY_QUERY",
			"共有資料Driveフォルダ共有ユーザー一覧の取得に失敗しました",
			nil,
		)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return rows, results.InternalServerError(
			"FIND_SHARED_DOCUMENT_DRIVE_FOLDER_USER_ROWS_FAILED",
			"共有資料Driveフォルダ共有ユーザー一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return rows, results.OK(nil, "FIND_SHARED_DOCUMENT_DRIVE_FOLDER_USER_ROWS_SUCCESS", "", nil)
}

/*
 * 共有ユーザーModel取得
 */
func (repository *sharedDocumentDriveFolderRepository) FindSharedDocumentDriveFolderUsers(query *gorm.DB) ([]models.SharedDocumentDriveFolderUser, results.Result) {
	var folderUsers []models.SharedDocumentDriveFolderUser

	if query == nil {
		return folderUsers, results.InternalServerError(
			"FIND_SHARED_DOCUMENT_DRIVE_FOLDER_USERS_EMPTY_QUERY",
			"共有資料Driveフォルダ共有ユーザー一覧の取得に失敗しました",
			nil,
		)
	}

	if err := query.Find(&folderUsers).Error; err != nil {
		return folderUsers, results.InternalServerError(
			"FIND_SHARED_DOCUMENT_DRIVE_FOLDER_USERS_FAILED",
			"共有資料Driveフォルダ共有ユーザー一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return folderUsers, results.OK(nil, "FIND_SHARED_DOCUMENT_DRIVE_FOLDER_USERS_SUCCESS", "", nil)
}

/*
 * 共有ユーザー作成
 */
func (repository *sharedDocumentDriveFolderRepository) CreateSharedDocumentDriveFolderUser(folderUser models.SharedDocumentDriveFolderUser) (models.SharedDocumentDriveFolderUser, results.Result) {
	if err := repository.db.Create(&folderUser).Error; err != nil {
		return folderUser, results.InternalServerError(
			"CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_USER_FAILED",
			"共有資料Driveフォルダ共有ユーザー情報の登録に失敗しました",
			err.Error(),
		)
	}

	return folderUser, results.Created(nil, "CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_USER_SUCCESS", "", nil)
}

/*
 * 共有ユーザー保存
 */
func (repository *sharedDocumentDriveFolderRepository) SaveSharedDocumentDriveFolderUser(folderUser models.SharedDocumentDriveFolderUser) (models.SharedDocumentDriveFolderUser, results.Result) {
	if err := repository.db.Save(&folderUser).Error; err != nil {
		return folderUser, results.InternalServerError(
			"SAVE_SHARED_DOCUMENT_DRIVE_FOLDER_USER_FAILED",
			"共有資料Driveフォルダ共有ユーザー情報の保存に失敗しました",
			err.Error(),
		)
	}

	return folderUser, results.OK(nil, "SAVE_SHARED_DOCUMENT_DRIVE_FOLDER_USER_SUCCESS", "", nil)
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
