package builders

import (
	"strings"

	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type SharedDocumentDriveFolderBuilder interface {
	BuildSearchSharedDocumentDriveFoldersQuery(userID uint, req types.SearchSharedDocumentDriveFoldersRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindSharedDocumentDriveFolderDetailQuery(userID uint, folderID uint) (*gorm.DB, results.Result)
}

/*
 * 従業員用 共有資料DriveフォルダBuilder
 *
 * 役割：
 * ・JWTから取得した本人userIdをもとにGORMクエリを作成する
 * ・本人に共有されている資料だけ取得できるようにする
 * ・DB実行はRepositoryに任せる
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
 * shared_document_drive_folder_users を通して、
 * 本人に共有されている資料だけ返す。
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildSearchSharedDocumentDriveFoldersQuery(
	userID uint,
	req types.SearchSharedDocumentDriveFoldersRequest,
) (*gorm.DB, *gorm.DB, results.Result) {
	if userID == 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_MY_SHARED_DOCUMENT_DRIVE_FOLDERS_QUERY_INVALID_USER_ID",
			"共有資料Driveフォルダ検索条件の作成に失敗しました",
			map[string]any{
				"userId": userID,
			},
		)
	}

	if req.Offset < 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_MY_SHARED_DOCUMENT_DRIVE_FOLDERS_QUERY_INVALID_OFFSET",
			"共有資料Driveフォルダ検索条件の作成に失敗しました",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	if req.Limit <= 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_MY_SHARED_DOCUMENT_DRIVE_FOLDERS_QUERY_INVALID_LIMIT",
			"共有資料Driveフォルダ検索条件の作成に失敗しました",
			map[string]any{
				"limit": req.Limit,
			},
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
			shared_document_drive_folder_users.created_at AS shared_at,
			shared_document_drive_folders.updated_at AS updated_at
		`).
		Joins(`
			INNER JOIN shared_document_drive_folder_users
				ON shared_document_drive_folder_users.shared_document_drive_folder_id = shared_document_drive_folders.id
				AND shared_document_drive_folder_users.user_id = ?
				AND shared_document_drive_folder_users.is_deleted = false
				AND shared_document_drive_folder_users.deleted_at IS NULL
		`, userID).
		Where("shared_document_drive_folders.is_deleted = ?", false).
		Where("shared_document_drive_folders.deleted_at IS NULL")

	countQuery := builder.db.
		Table("shared_document_drive_folders").
		Joins(`
			INNER JOIN shared_document_drive_folder_users
				ON shared_document_drive_folder_users.shared_document_drive_folder_id = shared_document_drive_folders.id
				AND shared_document_drive_folder_users.user_id = ?
				AND shared_document_drive_folder_users.is_deleted = false
				AND shared_document_drive_folder_users.deleted_at IS NULL
		`, userID).
		Where("shared_document_drive_folders.is_deleted = ?", false).
		Where("shared_document_drive_folders.deleted_at IS NULL")

	searchQuery = applySearchMySharedDocumentDriveFoldersCondition(searchQuery, req)
	countQuery = applySearchMySharedDocumentDriveFoldersCondition(countQuery, req)

	searchQuery = searchQuery.
		Order("shared_document_drive_folder_users.created_at DESC").
		Order("shared_document_drive_folders.id ASC").
		Offset(req.Offset).
		Limit(req.Limit)

	return searchQuery, countQuery, results.OK(
		nil,
		"BUILD_SEARCH_MY_SHARED_DOCUMENT_DRIVE_FOLDERS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 共有資料Driveフォルダ詳細取得用クエリ作成
 *
 * 本人に共有されている資料だけ取得できる。
 */
func (builder *sharedDocumentDriveFolderBuilder) BuildFindSharedDocumentDriveFolderDetailQuery(
	userID uint,
	folderID uint,
) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MY_SHARED_DOCUMENT_DRIVE_FOLDER_DETAIL_QUERY_INVALID_USER_ID",
			"共有資料Driveフォルダ詳細条件の作成に失敗しました",
			map[string]any{
				"userId": userID,
			},
		)
	}

	if folderID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MY_SHARED_DOCUMENT_DRIVE_FOLDER_DETAIL_QUERY_INVALID_FOLDER_ID",
			"共有資料Driveフォルダ詳細条件の作成に失敗しました",
			map[string]any{
				"targetSharedDocumentDriveFolderId": folderID,
			},
		)
	}

	query := builder.db.
		Table("shared_document_drive_folders").
		Select(`
			shared_document_drive_folders.id AS id,
			shared_document_drive_folders.folder_name AS folder_name,
			shared_document_drive_folders.description AS description,
			shared_document_drive_folders.drive_folder_id AS drive_folder_id,
			shared_document_drive_folders.folder_url AS folder_url,
			shared_document_drive_folders.synced_at AS synced_at,
			shared_document_drive_folder_users.created_at AS shared_at,
			shared_document_drive_folders.updated_at AS updated_at
		`).
		Joins(`
			INNER JOIN shared_document_drive_folder_users
				ON shared_document_drive_folder_users.shared_document_drive_folder_id = shared_document_drive_folders.id
				AND shared_document_drive_folder_users.user_id = ?
				AND shared_document_drive_folder_users.is_deleted = false
				AND shared_document_drive_folder_users.deleted_at IS NULL
		`, userID).
		Where("shared_document_drive_folders.id = ?", folderID).
		Where("shared_document_drive_folders.is_deleted = ?", false).
		Where("shared_document_drive_folders.deleted_at IS NULL")

	return query, results.OK(
		nil,
		"BUILD_FIND_MY_SHARED_DOCUMENT_DRIVE_FOLDER_DETAIL_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 共有資料Driveフォルダ検索条件適用
 */
func applySearchMySharedDocumentDriveFoldersCondition(
	query *gorm.DB,
	req types.SearchSharedDocumentDriveFoldersRequest,
) *gorm.DB {
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
