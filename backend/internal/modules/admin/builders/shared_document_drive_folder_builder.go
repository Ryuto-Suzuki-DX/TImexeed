package builders

import (
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type SharedDocumentDriveFolderBuilder interface {
	BuildSearchSharedDocumentDriveFoldersQuery(req types.SearchSharedDocumentDriveFoldersRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindActiveSharedDocumentDriveFolderByIDQuery(folderID uint) (*gorm.DB, results.Result)
	BuildFindActiveSharedDocumentDriveFolderUsersByFolderIDQuery(folderID uint) (*gorm.DB, results.Result)
	BuildFindAllSharedDocumentDriveFolderUsersByFolderIDQuery(folderID uint) (*gorm.DB, results.Result)
	BuildFindActiveUsersByIDsQuery(userIDs []uint) (*gorm.DB, results.Result)
	BuildFindAllActiveUsersQuery() (*gorm.DB, results.Result)
	BuildFindActiveAdminUsersQuery() (*gorm.DB, results.Result)

	BuildCreateSharedDocumentDriveFolderModel(folderName string, description *string, driveFolderID string, folderURL string) (models.SharedDocumentDriveFolder, results.Result)
	BuildUpdateSharedDocumentDriveFolderModel(currentFolder models.SharedDocumentDriveFolder, folderName string, description *string, driveFolderID string, folderURL string) (models.SharedDocumentDriveFolder, results.Result)
	BuildDeleteSharedDocumentDriveFolderModel(currentFolder models.SharedDocumentDriveFolder) (models.SharedDocumentDriveFolder, results.Result)
	BuildCreateSharedDocumentDriveFolderUserModel(folderID uint, userID uint) (models.SharedDocumentDriveFolderUser, results.Result)
	BuildActiveSharedDocumentDriveFolderUserModel(currentUser models.SharedDocumentDriveFolderUser) (models.SharedDocumentDriveFolderUser, results.Result)
	BuildDeletedSharedDocumentDriveFolderUserModel(currentUser models.SharedDocumentDriveFolderUser) (models.SharedDocumentDriveFolderUser, results.Result)
	BuildSyncedSharedDocumentDriveFolderModel(currentFolder models.SharedDocumentDriveFolder, syncedAt time.Time) (models.SharedDocumentDriveFolder, results.Result)
	BuildSyncedSharedDocumentDriveFolderUserModel(currentUser models.SharedDocumentDriveFolderUser, syncedAt time.Time) (models.SharedDocumentDriveFolderUser, results.Result)
}

/*
 * 管理者用 共有資料DriveフォルダBuilder
 */
type sharedDocumentDriveFolderBuilder struct {
	db *gorm.DB
}

/*
 * SharedDocumentDriveFolderBuilder生成
 */
func NewSharedDocumentDriveFolderBuilder(db *gorm.DB) SharedDocumentDriveFolderBuilder {
	return &sharedDocumentDriveFolderBuilder{db: db}
}

/*
 * 共有資料Driveフォルダ検索用クエリ作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildSearchSharedDocumentDriveFoldersQuery(req types.SearchSharedDocumentDriveFoldersRequest) (*gorm.DB, *gorm.DB, results.Result) {
	if req.Offset < 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_SHARED_DOCUMENT_DRIVE_FOLDERS_QUERY_INVALID_OFFSET",
			"共有資料Driveフォルダ検索条件の作成に失敗しました",
			map[string]any{"offset": req.Offset},
		)
	}

	if req.Limit <= 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_SHARED_DOCUMENT_DRIVE_FOLDERS_QUERY_INVALID_LIMIT",
			"共有資料Driveフォルダ検索条件の作成に失敗しました",
			map[string]any{"limit": req.Limit},
		)
	}

	searchQuery := builder.db.
		Table("shared_document_drive_folders").
		Select(`
			shared_document_drive_folders.id AS id,
			shared_document_drive_folders.folder_name AS folder_name,
			shared_document_drive_folders.description AS description,
			shared_document_drive_folders.drive_folder_id AS drive_folder_id,
			shared_document_drive_folders.folder_url AS folder_url,
			shared_document_drive_folders.synced_at AS synced_at,
			COUNT(shared_document_drive_folder_users.id) AS shared_user_count,
			shared_document_drive_folders.created_at AS created_at,
			shared_document_drive_folders.updated_at AS updated_at
		`).
		Joins(`
			LEFT JOIN shared_document_drive_folder_users
				ON shared_document_drive_folder_users.shared_document_drive_folder_id = shared_document_drive_folders.id
				AND shared_document_drive_folder_users.is_deleted = false
				AND shared_document_drive_folder_users.deleted_at IS NULL
		`).
		Where("shared_document_drive_folders.is_deleted = ?", false).
		Where("shared_document_drive_folders.deleted_at IS NULL").
		Group("shared_document_drive_folders.id")

	countQuery := builder.db.
		Table("shared_document_drive_folders").
		Where("shared_document_drive_folders.is_deleted = ?", false).
		Where("shared_document_drive_folders.deleted_at IS NULL")

	searchQuery = applySearchSharedDocumentDriveFoldersCondition(searchQuery, req)
	countQuery = applySearchSharedDocumentDriveFoldersCondition(countQuery, req)

	searchQuery = searchQuery.
		Order("shared_document_drive_folders.id ASC").
		Offset(req.Offset).
		Limit(req.Limit)

	return searchQuery, countQuery, results.OK(
		nil,
		"BUILD_SEARCH_SHARED_DOCUMENT_DRIVE_FOLDERS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有効な共有資料Driveフォルダ1件取得用クエリ作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildFindActiveSharedDocumentDriveFolderByIDQuery(folderID uint) (*gorm.DB, results.Result) {
	if folderID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ACTIVE_SHARED_DOCUMENT_DRIVE_FOLDER_BY_ID_QUERY_INVALID_ID",
			"共有資料Driveフォルダ取得条件の作成に失敗しました",
			map[string]any{"targetSharedDocumentDriveFolderId": folderID},
		)
	}

	query := builder.db.
		Model(&models.SharedDocumentDriveFolder{}).
		Where("id = ?", folderID).
		Where("is_deleted = ?", false).
		Where("deleted_at IS NULL")

	return query, results.OK(nil, "BUILD_FIND_ACTIVE_SHARED_DOCUMENT_DRIVE_FOLDER_BY_ID_QUERY_SUCCESS", "", nil)
}

/*
 * 有効な共有ユーザー表示用クエリ作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildFindActiveSharedDocumentDriveFolderUsersByFolderIDQuery(folderID uint) (*gorm.DB, results.Result) {
	if folderID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ACTIVE_SHARED_DOCUMENT_DRIVE_FOLDER_USERS_BY_FOLDER_ID_QUERY_INVALID_ID",
			"共有資料Driveフォルダ共有ユーザー取得条件の作成に失敗しました",
			map[string]any{"targetSharedDocumentDriveFolderId": folderID},
		)
	}

	query := builder.db.
		Table("shared_document_drive_folder_users").
		Select(`
			shared_document_drive_folder_users.id AS id,
			shared_document_drive_folder_users.shared_document_drive_folder_id AS shared_document_drive_folder_id,
			shared_document_drive_folder_users.user_id AS user_id,
			users.name AS user_name,
			users.email AS user_email,
			users.role AS user_role,
			shared_document_drive_folder_users.synced_at AS synced_at,
			shared_document_drive_folder_users.created_at AS created_at,
			shared_document_drive_folder_users.updated_at AS updated_at
		`).
		Joins(`
			INNER JOIN users
				ON users.id = shared_document_drive_folder_users.user_id
				AND users.is_deleted = false
		`).
		Where("shared_document_drive_folder_users.shared_document_drive_folder_id = ?", folderID).
		Where("shared_document_drive_folder_users.is_deleted = ?", false).
		Where("shared_document_drive_folder_users.deleted_at IS NULL").
		Order("users.id ASC")

	return query, results.OK(nil, "BUILD_FIND_ACTIVE_SHARED_DOCUMENT_DRIVE_FOLDER_USERS_BY_FOLDER_ID_QUERY_SUCCESS", "", nil)
}

/*
 * 論理削除済み含む共有ユーザー取得用クエリ作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildFindAllSharedDocumentDriveFolderUsersByFolderIDQuery(folderID uint) (*gorm.DB, results.Result) {
	if folderID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ALL_SHARED_DOCUMENT_DRIVE_FOLDER_USERS_BY_FOLDER_ID_QUERY_INVALID_ID",
			"共有資料Driveフォルダ共有ユーザー取得条件の作成に失敗しました",
			map[string]any{"targetSharedDocumentDriveFolderId": folderID},
		)
	}

	query := builder.db.
		Model(&models.SharedDocumentDriveFolderUser{}).
		Where("shared_document_drive_folder_id = ?", folderID)

	return query, results.OK(nil, "BUILD_FIND_ALL_SHARED_DOCUMENT_DRIVE_FOLDER_USERS_BY_FOLDER_ID_QUERY_SUCCESS", "", nil)
}

/*
 * 有効なUSER取得用クエリ作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildFindActiveUsersByIDsQuery(userIDs []uint) (*gorm.DB, results.Result) {
	uniqueUserIDs := uniqueSharedDocumentDriveFolderUserIDs(userIDs)
	if len(uniqueUserIDs) == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ACTIVE_USERS_BY_IDS_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_QUERY_EMPTY_USER_IDS",
			"共有対象ユーザー取得条件の作成に失敗しました",
			nil,
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("id IN ?", uniqueUserIDs).
		Where("role = ?", "USER").
		Where("is_deleted = ?", false).
		Order("id ASC")

	return query, results.OK(nil, "BUILD_FIND_ACTIVE_USERS_BY_IDS_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_QUERY_SUCCESS", "", nil)
}

/*
 * 有効なUSER全員取得用クエリ作成
 *
 * 全員追加ボタン用。
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildFindAllActiveUsersQuery() (*gorm.DB, results.Result) {
	query := builder.db.
		Model(&models.User{}).
		Where("role = ?", "USER").
		Where("is_deleted = ?", false).
		Order("id ASC")

	return query, results.OK(nil, "BUILD_FIND_ALL_ACTIVE_USERS_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_QUERY_SUCCESS", "", nil)
}

/*
 * 有効な管理者ユーザー取得用クエリ作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildFindActiveAdminUsersQuery() (*gorm.DB, results.Result) {
	query := builder.db.
		Model(&models.User{}).
		Where("role = ?", "ADMIN").
		Where("is_deleted = ?", false).
		Order("id ASC")

	return query, results.OK(nil, "BUILD_FIND_ACTIVE_ADMIN_USERS_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_QUERY_SUCCESS", "", nil)
}

/*
 * 共有資料Driveフォルダ作成用Model作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildCreateSharedDocumentDriveFolderModel(
	folderName string,
	description *string,
	driveFolderID string,
	folderURL string,
) (models.SharedDocumentDriveFolder, results.Result) {
	folderName = strings.TrimSpace(folderName)
	driveFolderID = strings.TrimSpace(driveFolderID)
	folderURL = strings.TrimSpace(folderURL)

	if folderName == "" {
		return models.SharedDocumentDriveFolder{}, results.BadRequest(
			"BUILD_CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_EMPTY_FOLDER_NAME",
			"共有資料Driveフォルダ作成データの作成に失敗しました",
			nil,
		)
	}

	if driveFolderID == "" || folderURL == "" {
		return models.SharedDocumentDriveFolder{}, results.BadRequest(
			"BUILD_CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_EMPTY_DRIVE_VALUE",
			"共有資料Driveフォルダ作成データの作成に失敗しました",
			map[string]any{
				"driveFolderId": driveFolderID,
				"folderUrl":     folderURL,
			},
		)
	}

	folder := models.SharedDocumentDriveFolder{
		FolderName:    folderName,
		Description:   normalizeSharedDocumentDriveFolderOptionalString(description),
		DriveFolderID: driveFolderID,
		FolderURL:     folderURL,
		IsDeleted:     false,
	}

	return folder, results.OK(nil, "BUILD_CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_SUCCESS", "", nil)
}

/*
 * 共有資料Driveフォルダ更新用Model作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildUpdateSharedDocumentDriveFolderModel(
	currentFolder models.SharedDocumentDriveFolder,
	folderName string,
	description *string,
	driveFolderID string,
	folderURL string,
) (models.SharedDocumentDriveFolder, results.Result) {
	if currentFolder.ID == 0 {
		return models.SharedDocumentDriveFolder{}, results.BadRequest(
			"BUILD_UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_EMPTY_CURRENT_FOLDER",
			"共有資料Driveフォルダ更新データの作成に失敗しました",
			nil,
		)
	}

	folderName = strings.TrimSpace(folderName)
	driveFolderID = strings.TrimSpace(driveFolderID)
	folderURL = strings.TrimSpace(folderURL)

	if folderName == "" {
		return models.SharedDocumentDriveFolder{}, results.BadRequest(
			"BUILD_UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_EMPTY_FOLDER_NAME",
			"共有資料Driveフォルダ更新データの作成に失敗しました",
			nil,
		)
	}

	if driveFolderID == "" || folderURL == "" {
		return models.SharedDocumentDriveFolder{}, results.BadRequest(
			"BUILD_UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_EMPTY_DRIVE_VALUE",
			"共有資料Driveフォルダ更新データの作成に失敗しました",
			map[string]any{
				"driveFolderId": driveFolderID,
				"folderUrl":     folderURL,
			},
		)
	}

	currentFolder.FolderName = folderName
	currentFolder.Description = normalizeSharedDocumentDriveFolderOptionalString(description)
	currentFolder.DriveFolderID = driveFolderID
	currentFolder.FolderURL = folderURL
	currentFolder.IsDeleted = false

	return currentFolder, results.OK(nil, "BUILD_UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_SUCCESS", "", nil)
}

/*
 * 共有資料Driveフォルダ削除用Model作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildDeleteSharedDocumentDriveFolderModel(currentFolder models.SharedDocumentDriveFolder) (models.SharedDocumentDriveFolder, results.Result) {
	if currentFolder.ID == 0 {
		return models.SharedDocumentDriveFolder{}, results.BadRequest(
			"BUILD_DELETE_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_EMPTY_CURRENT_FOLDER",
			"共有資料Driveフォルダ削除データの作成に失敗しました",
			nil,
		)
	}

	now := time.Now()
	currentFolder.IsDeleted = true
	currentFolder.DeletedAt = &now

	return currentFolder, results.OK(nil, "BUILD_DELETE_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_SUCCESS", "", nil)
}

/*
 * 共有ユーザー作成用Model作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildCreateSharedDocumentDriveFolderUserModel(folderID uint, userID uint) (models.SharedDocumentDriveFolderUser, results.Result) {
	if folderID == 0 || userID == 0 {
		return models.SharedDocumentDriveFolderUser{}, results.BadRequest(
			"BUILD_CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_USER_MODEL_INVALID_VALUE",
			"共有資料Driveフォルダ共有ユーザー作成データの作成に失敗しました",
			map[string]any{
				"sharedDocumentDriveFolderId": folderID,
				"userId":                      userID,
			},
		)
	}

	folderUser := models.SharedDocumentDriveFolderUser{
		SharedDocumentDriveFolderID: folderID,
		UserID:                      userID,
		IsDeleted:                   false,
	}

	return folderUser, results.OK(nil, "BUILD_CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_USER_MODEL_SUCCESS", "", nil)
}

/*
 * 共有ユーザー有効化Model作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildActiveSharedDocumentDriveFolderUserModel(currentUser models.SharedDocumentDriveFolderUser) (models.SharedDocumentDriveFolderUser, results.Result) {
	if currentUser.ID == 0 {
		return models.SharedDocumentDriveFolderUser{}, results.BadRequest(
			"BUILD_ACTIVE_SHARED_DOCUMENT_DRIVE_FOLDER_USER_MODEL_EMPTY_CURRENT_USER",
			"共有資料Driveフォルダ共有ユーザー更新データの作成に失敗しました",
			nil,
		)
	}

	currentUser.IsDeleted = false
	currentUser.DeletedAt = nil

	return currentUser, results.OK(nil, "BUILD_ACTIVE_SHARED_DOCUMENT_DRIVE_FOLDER_USER_MODEL_SUCCESS", "", nil)
}

/*
 * 共有ユーザー論理削除Model作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildDeletedSharedDocumentDriveFolderUserModel(currentUser models.SharedDocumentDriveFolderUser) (models.SharedDocumentDriveFolderUser, results.Result) {
	if currentUser.ID == 0 {
		return models.SharedDocumentDriveFolderUser{}, results.BadRequest(
			"BUILD_DELETED_SHARED_DOCUMENT_DRIVE_FOLDER_USER_MODEL_EMPTY_CURRENT_USER",
			"共有資料Driveフォルダ共有ユーザー削除データの作成に失敗しました",
			nil,
		)
	}

	now := time.Now()
	currentUser.IsDeleted = true
	currentUser.DeletedAt = &now

	return currentUser, results.OK(nil, "BUILD_DELETED_SHARED_DOCUMENT_DRIVE_FOLDER_USER_MODEL_SUCCESS", "", nil)
}

/*
 * 共有資料Driveフォルダ同期済みModel作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildSyncedSharedDocumentDriveFolderModel(currentFolder models.SharedDocumentDriveFolder, syncedAt time.Time) (models.SharedDocumentDriveFolder, results.Result) {
	if currentFolder.ID == 0 {
		return models.SharedDocumentDriveFolder{}, results.BadRequest(
			"BUILD_SYNCED_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_EMPTY_CURRENT_FOLDER",
			"共有資料Driveフォルダ同期データの作成に失敗しました",
			nil,
		)
	}

	currentFolder.SyncedAt = &syncedAt

	return currentFolder, results.OK(nil, "BUILD_SYNCED_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_SUCCESS", "", nil)
}

/*
 * 共有ユーザー同期済みModel作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildSyncedSharedDocumentDriveFolderUserModel(currentUser models.SharedDocumentDriveFolderUser, syncedAt time.Time) (models.SharedDocumentDriveFolderUser, results.Result) {
	if currentUser.ID == 0 {
		return models.SharedDocumentDriveFolderUser{}, results.BadRequest(
			"BUILD_SYNCED_SHARED_DOCUMENT_DRIVE_FOLDER_USER_MODEL_EMPTY_CURRENT_USER",
			"共有資料Driveフォルダ共有ユーザー同期データの作成に失敗しました",
			nil,
		)
	}

	currentUser.SyncedAt = &syncedAt

	return currentUser, results.OK(nil, "BUILD_SYNCED_SHARED_DOCUMENT_DRIVE_FOLDER_USER_MODEL_SUCCESS", "", nil)
}

/*
 * 検索条件適用
 */
func applySearchSharedDocumentDriveFoldersCondition(query *gorm.DB, req types.SearchSharedDocumentDriveFoldersRequest) *gorm.DB {
	if strings.TrimSpace(req.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(req.Keyword) + "%"
		query = query.Where(
			"shared_document_drive_folders.folder_name ILIKE ? OR shared_document_drive_folders.description ILIKE ? OR shared_document_drive_folders.drive_folder_id ILIKE ?",
			keyword,
			keyword,
			keyword,
		)
	}

	return query
}

/*
 * optional string整形
 */
func normalizeSharedDocumentDriveFolderOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmedValue := strings.TrimSpace(*value)
	if trimmedValue == "" {
		return nil
	}

	return &trimmedValue
}

/*
 * uint重複排除
 */
func uniqueSharedDocumentDriveFolderUserIDs(values []uint) []uint {
	seen := map[uint]bool{}
	uniqueValues := make([]uint, 0, len(values))

	for _, value := range values {
		if value == 0 {
			continue
		}

		if seen[value] {
			continue
		}

		seen[value] = true
		uniqueValues = append(uniqueValues, value)
	}

	return uniqueValues
}
