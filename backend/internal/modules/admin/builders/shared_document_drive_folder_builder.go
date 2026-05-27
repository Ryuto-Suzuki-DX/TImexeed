package builders

import (
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用 共有資料DriveフォルダBuilder interface
 */
type SharedDocumentDriveFolderBuilder interface {
	BuildSearchSharedDocumentDriveFoldersQuery(req types.SearchSharedDocumentDriveFoldersRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindActiveSharedDocumentDriveFolderByIDQuery(folderID uint) (*gorm.DB, results.Result)
	BuildFindAllActiveSharedDocumentDriveFoldersQuery() (*gorm.DB, results.Result)
	BuildFindActiveExternalStorageLinkByLinkTypeQuery(linkType string) (*gorm.DB, results.Result)
	BuildFindAllActiveUsersQuery() (*gorm.DB, results.Result)
	BuildFindActiveAdminUsersQuery() (*gorm.DB, results.Result)

	BuildCreateSharedDocumentDriveFolderModel(folderName string, description *string, driveFolderID string, folderURL string) (models.SharedDocumentDriveFolder, results.Result)
	BuildUpdateSharedDocumentDriveFolderModel(currentFolder models.SharedDocumentDriveFolder, folderName string, description *string) (models.SharedDocumentDriveFolder, results.Result)
	BuildDeleteSharedDocumentDriveFolderModel(currentFolder models.SharedDocumentDriveFolder) (models.SharedDocumentDriveFolder, results.Result)
	BuildSyncedSharedDocumentDriveFolderModel(currentFolder models.SharedDocumentDriveFolder, syncedAt time.Time) (models.SharedDocumentDriveFolder, results.Result)
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
			shared_document_drive_folders.created_at AS created_at,
			shared_document_drive_folders.updated_at AS updated_at
		`).
		Where("shared_document_drive_folders.is_deleted = ?", false).
		Where("shared_document_drive_folders.deleted_at IS NULL")

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
 * 有効な共有資料Driveフォルダ全件取得用クエリ作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildFindAllActiveSharedDocumentDriveFoldersQuery() (*gorm.DB, results.Result) {
	query := builder.db.
		Model(&models.SharedDocumentDriveFolder{}).
		Where("is_deleted = ?", false).
		Where("deleted_at IS NULL").
		Order("id ASC")

	return query, results.OK(nil, "BUILD_FIND_ALL_ACTIVE_SHARED_DOCUMENT_DRIVE_FOLDERS_QUERY_SUCCESS", "", nil)
}

/*
 * 有効な外部ストレージリンク1件取得用クエリ作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildFindActiveExternalStorageLinkByLinkTypeQuery(linkType string) (*gorm.DB, results.Result) {
	linkType = strings.TrimSpace(linkType)
	if linkType == "" {
		return nil, results.BadRequest(
			"BUILD_FIND_ACTIVE_EXTERNAL_STORAGE_LINK_BY_LINK_TYPE_QUERY_EMPTY_LINK_TYPE",
			"外部ストレージリンク取得条件の作成に失敗しました",
			nil,
		)
	}

	query := builder.db.
		Model(&models.ExternalStorageLink{}).
		Where("link_type = ?", linkType).
		Where("is_deleted = ?", false).
		Where("deleted_at IS NULL").
		Order("id ASC")

	return query, results.OK(nil, "BUILD_FIND_ACTIVE_EXTERNAL_STORAGE_LINK_BY_LINK_TYPE_QUERY_SUCCESS", "", nil)
}

/*
 * 有効なUSER全員取得用クエリ作成
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildFindAllActiveUsersQuery() (*gorm.DB, results.Result) {
	query := builder.db.
		Model(&models.User{}).
		Where("role = ?", "USER").
		Where("is_deleted = ?", false).
		Where("deleted_at IS NULL").
		Where("retirement_date IS NULL OR retirement_date > CURRENT_DATE").
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
		Where("deleted_at IS NULL").
		Where("retirement_date IS NULL OR retirement_date > CURRENT_DATE").
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
			"共有資料Driveフォルダ名が入力されていません",
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
) (models.SharedDocumentDriveFolder, results.Result) {
	if currentFolder.ID == 0 {
		return models.SharedDocumentDriveFolder{}, results.BadRequest(
			"BUILD_UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_EMPTY_CURRENT_FOLDER",
			"共有資料Driveフォルダ更新データの作成に失敗しました",
			nil,
		)
	}

	folderName = strings.TrimSpace(folderName)
	if folderName == "" {
		return models.SharedDocumentDriveFolder{}, results.BadRequest(
			"BUILD_UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_MODEL_EMPTY_FOLDER_NAME",
			"共有資料Driveフォルダ名が入力されていません",
			nil,
		)
	}

	currentFolder.FolderName = folderName
	currentFolder.Description = normalizeSharedDocumentDriveFolderOptionalString(description)
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
