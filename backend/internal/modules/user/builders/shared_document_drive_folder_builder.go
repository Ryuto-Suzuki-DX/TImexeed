package builders

import (
	"strings"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 従業員用 共有資料DriveフォルダBuilder interface
 *
 * 従業員側では、全ユーザー向け共有資料Driveフォルダの閲覧だけを扱う。
 */
type SharedDocumentDriveFolderBuilder interface {
	BuildSearchSharedDocumentDriveFoldersQuery(req types.SearchSharedDocumentDriveFoldersRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindActiveSharedDocumentDriveFolderByIDQuery(folderID uint) (*gorm.DB, results.Result)
}

/*
 * 従業員用 共有資料DriveフォルダBuilder
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
 *
 * 従業員側では、shared_document_drive_folders に登録されている有効なフォルダを全件閲覧対象にする。
 * 共有ユーザー中間テーブルは使わない。
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildSearchSharedDocumentDriveFoldersQuery(req types.SearchSharedDocumentDriveFoldersRequest) (*gorm.DB, *gorm.DB, results.Result) {
	if req.Offset < 0 {
		return nil, nil, results.BadRequest(
			"BUILD_USER_SEARCH_SHARED_DOCUMENT_DRIVE_FOLDERS_QUERY_INVALID_OFFSET",
			"共有資料Driveフォルダ検索条件の作成に失敗しました",
			map[string]any{"offset": req.Offset},
		)
	}

	if req.Limit <= 0 {
		return nil, nil, results.BadRequest(
			"BUILD_USER_SEARCH_SHARED_DOCUMENT_DRIVE_FOLDERS_QUERY_INVALID_LIMIT",
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

	searchQuery = applyUserSearchSharedDocumentDriveFoldersCondition(searchQuery, req)
	countQuery = applyUserSearchSharedDocumentDriveFoldersCondition(countQuery, req)

	searchQuery = searchQuery.
		Order("shared_document_drive_folders.id ASC").
		Offset(req.Offset).
		Limit(req.Limit)

	return searchQuery, countQuery, results.OK(
		nil,
		"BUILD_USER_SEARCH_SHARED_DOCUMENT_DRIVE_FOLDERS_QUERY_SUCCESS",
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
			"BUILD_USER_FIND_ACTIVE_SHARED_DOCUMENT_DRIVE_FOLDER_BY_ID_QUERY_INVALID_ID",
			"共有資料Driveフォルダ取得条件の作成に失敗しました",
			map[string]any{"targetSharedDocumentDriveFolderId": folderID},
		)
	}

	query := builder.db.
		Model(&models.SharedDocumentDriveFolder{}).
		Where("id = ?", folderID).
		Where("is_deleted = ?", false).
		Where("deleted_at IS NULL")

	return query, results.OK(
		nil,
		"BUILD_USER_FIND_ACTIVE_SHARED_DOCUMENT_DRIVE_FOLDER_BY_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 検索条件適用
 */
func applyUserSearchSharedDocumentDriveFoldersCondition(query *gorm.DB, req types.SearchSharedDocumentDriveFoldersRequest) *gorm.DB {
	keyword := strings.TrimSpace(req.Keyword)
	if keyword == "" {
		return query
	}

	likeKeyword := "%" + keyword + "%"
	return query.Where(
		"shared_document_drive_folders.folder_name ILIKE ? OR shared_document_drive_folders.description ILIKE ?",
		likeKeyword,
		likeKeyword,
	)
}
