package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

const PersonalInformationDriveRootLinkType = "PERSONAL_INFORMATION_DRIVE_ROOT"

type PersonalInformationDriveFolderBuilder interface {
	BuildSearchPersonalInformationDriveFoldersQuery(req types.SearchPersonalInformationDriveFoldersRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindActivePersonalInformationDriveFolderByUserIDQuery(targetUserID uint) (*gorm.DB, results.Result)
	BuildFindActiveUserByIDQuery(targetUserID uint) (*gorm.DB, results.Result)
	BuildFindActiveAdminUsersQuery() (*gorm.DB, results.Result)
	BuildFindPersonalInformationDriveRootLinkQuery() (*gorm.DB, results.Result)
	BuildCreatePersonalInformationDriveFolderModel(userID uint, externalStorageLinkID uint, folderName string, driveFolderID string, folderURL string, syncedAt time.Time) (models.PersonalInformationDriveFolder, results.Result)
	BuildUpdatePersonalInformationDriveFolderModel(currentFolder models.PersonalInformationDriveFolder, externalStorageLinkID uint, folderName string, driveFolderID string, folderURL string, syncedAt time.Time) (models.PersonalInformationDriveFolder, results.Result)
}

/*
 * 管理者用 個人情報DriveフォルダBuilder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Serviceから受け取った値をもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Count / Create / Save はRepositoryに任せる
 */
type personalInformationDriveFolderBuilder struct {
	db *gorm.DB
}

/*
 * PersonalInformationDriveFolderBuilder生成
 */
func NewPersonalInformationDriveFolderBuilder(db *gorm.DB) PersonalInformationDriveFolderBuilder {
	return &personalInformationDriveFolderBuilder{db: db}
}

/*
 * 個人情報Driveフォルダ検索用クエリ作成
 *
 * users を主軸にする。
 * 理由：
 * ・フォルダ未作成ユーザーも検索結果に出すため
 * ・管理者はユーザーを検索して、そこからフォルダ作成/表示へ進むため
 */
func (builder *personalInformationDriveFolderBuilder) BuildSearchPersonalInformationDriveFoldersQuery(req types.SearchPersonalInformationDriveFoldersRequest) (*gorm.DB, *gorm.DB, results.Result) {
	if req.Offset < 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_PERSONAL_INFORMATION_DRIVE_FOLDERS_QUERY_INVALID_OFFSET",
			"個人情報Driveフォルダ検索条件の作成に失敗しました",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	if req.Limit <= 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_PERSONAL_INFORMATION_DRIVE_FOLDERS_QUERY_INVALID_LIMIT",
			"個人情報Driveフォルダ検索条件の作成に失敗しました",
			map[string]any{
				"limit": req.Limit,
			},
		)
	}

	/*
	 * 注意：
	 * searchQuery と countQuery で同じ *gorm.DB を使い回さない。
	 * GORMのチェーンはStatementを引き継ぐため、共通baseQueryにJoinsを追加すると、
	 * countQuery側にもJOINが混ざり、同じJOINが二重に乗ることがある。
	 *
	 * その結果、PostgreSQLで
	 * "table name personal_information_drive_folders specified more than once"
	 * が発生する。
	 *
	 * 対策として、検索用と件数用は最初から別々に作る。
	 */
	searchQuery := builder.db.
		Table("users").
		Select(`
			users.id AS user_id,
			users.name AS user_name,
			users.email AS user_email,
			users.role AS user_role,
			users.department_id AS department_id,
			personal_information_drive_folders.id AS personal_information_drive_folder_id,
			personal_information_drive_folders.external_storage_link_id AS external_storage_link_id,
			personal_information_drive_folders.folder_name AS folder_name,
			personal_information_drive_folders.drive_folder_id AS drive_folder_id,
			personal_information_drive_folders.folder_url AS folder_url,
			personal_information_drive_folders.synced_at AS synced_at,
			personal_information_drive_folders.created_at AS folder_created_at,
			personal_information_drive_folders.updated_at AS folder_updated_at,
			CASE
				WHEN personal_information_drive_folders.id IS NULL THEN false
				ELSE true
			END AS folder_registered
		`).
		Joins(`
			LEFT JOIN personal_information_drive_folders
				ON personal_information_drive_folders.user_id = users.id
				AND personal_information_drive_folders.is_deleted = false
				AND personal_information_drive_folders.deleted_at IS NULL
		`).
		Where("users.is_deleted = ?", false)

	countQuery := builder.db.
		Table("users").
		Where("users.is_deleted = ?", false)

	searchQuery = applySearchPersonalInformationDriveFoldersCondition(searchQuery, req)
	countQuery = applySearchPersonalInformationDriveFoldersCondition(countQuery, req)

	searchQuery = searchQuery.
		Order("users.id ASC").
		Offset(req.Offset).
		Limit(req.Limit)

	return searchQuery, countQuery, results.OK(
		nil,
		"BUILD_SEARCH_PERSONAL_INFORMATION_DRIVE_FOLDERS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 対象ユーザーの有効な個人情報Driveフォルダ取得用クエリ作成
 */
func (builder *personalInformationDriveFolderBuilder) BuildFindActivePersonalInformationDriveFolderByUserIDQuery(targetUserID uint) (*gorm.DB, results.Result) {
	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ACTIVE_PERSONAL_INFORMATION_DRIVE_FOLDER_BY_USER_ID_QUERY_INVALID_TARGET_USER_ID",
			"個人情報Driveフォルダ取得条件の作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	query := builder.db.
		Model(&models.PersonalInformationDriveFolder{}).
		Where("user_id = ?", targetUserID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_ACTIVE_PERSONAL_INFORMATION_DRIVE_FOLDER_BY_USER_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 対象ユーザー取得用クエリ作成
 */
func (builder *personalInformationDriveFolderBuilder) BuildFindActiveUserByIDQuery(targetUserID uint) (*gorm.DB, results.Result) {
	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ACTIVE_USER_BY_ID_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_QUERY_INVALID_TARGET_USER_ID",
			"ユーザー取得条件の作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("id = ?", targetUserID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_ACTIVE_USER_BY_ID_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有効な管理者ユーザー取得用クエリ作成
 */
func (builder *personalInformationDriveFolderBuilder) BuildFindActiveAdminUsersQuery() (*gorm.DB, results.Result) {
	query := builder.db.
		Model(&models.User{}).
		Where("role = ?", "ADMIN").
		Where("is_deleted = ?", false).
		Order("id ASC")

	return query, results.OK(
		nil,
		"BUILD_FIND_ACTIVE_ADMIN_USERS_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 個人情報Drive親フォルダ設定取得用クエリ作成
 */
func (builder *personalInformationDriveFolderBuilder) BuildFindPersonalInformationDriveRootLinkQuery() (*gorm.DB, results.Result) {
	query := builder.db.
		Model(&models.ExternalStorageLink{}).
		Where("link_type = ?", PersonalInformationDriveRootLinkType).
		Where("is_deleted = ?", false).
		Order("id ASC")

	return query, results.OK(
		nil,
		"BUILD_FIND_PERSONAL_INFORMATION_DRIVE_ROOT_LINK_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 個人情報Driveフォルダ作成用Model作成
 */
func (builder *personalInformationDriveFolderBuilder) BuildCreatePersonalInformationDriveFolderModel(
	userID uint,
	externalStorageLinkID uint,
	folderName string,
	driveFolderID string,
	folderURL string,
	syncedAt time.Time,
) (models.PersonalInformationDriveFolder, results.Result) {
	if userID == 0 {
		return models.PersonalInformationDriveFolder{}, results.BadRequest(
			"BUILD_CREATE_PERSONAL_INFORMATION_DRIVE_FOLDER_MODEL_INVALID_USER_ID",
			"個人情報Driveフォルダ作成データの作成に失敗しました",
			map[string]any{"userId": userID},
		)
	}

	if externalStorageLinkID == 0 {
		return models.PersonalInformationDriveFolder{}, results.BadRequest(
			"BUILD_CREATE_PERSONAL_INFORMATION_DRIVE_FOLDER_MODEL_INVALID_EXTERNAL_STORAGE_LINK_ID",
			"個人情報Driveフォルダ作成データの作成に失敗しました",
			map[string]any{"externalStorageLinkId": externalStorageLinkID},
		)
	}

	if folderName == "" || driveFolderID == "" || folderURL == "" {
		return models.PersonalInformationDriveFolder{}, results.BadRequest(
			"BUILD_CREATE_PERSONAL_INFORMATION_DRIVE_FOLDER_MODEL_EMPTY_DRIVE_VALUE",
			"個人情報Driveフォルダ作成データの作成に失敗しました",
			map[string]any{
				"folderName":    folderName,
				"driveFolderId": driveFolderID,
				"folderUrl":     folderURL,
			},
		)
	}

	folder := models.PersonalInformationDriveFolder{
		UserID:                userID,
		ExternalStorageLinkID: externalStorageLinkID,
		FolderName:            folderName,
		DriveFolderID:         driveFolderID,
		FolderURL:             folderURL,
		SyncedAt:              &syncedAt,
		IsDeleted:             false,
	}

	return folder, results.OK(nil, "BUILD_CREATE_PERSONAL_INFORMATION_DRIVE_FOLDER_MODEL_SUCCESS", "", nil)
}

/*
 * 個人情報Driveフォルダ更新用Model作成
 */
func (builder *personalInformationDriveFolderBuilder) BuildUpdatePersonalInformationDriveFolderModel(
	currentFolder models.PersonalInformationDriveFolder,
	externalStorageLinkID uint,
	folderName string,
	driveFolderID string,
	folderURL string,
	syncedAt time.Time,
) (models.PersonalInformationDriveFolder, results.Result) {
	if currentFolder.ID == 0 {
		return models.PersonalInformationDriveFolder{}, results.BadRequest(
			"BUILD_UPDATE_PERSONAL_INFORMATION_DRIVE_FOLDER_MODEL_EMPTY_CURRENT_FOLDER",
			"個人情報Driveフォルダ更新データの作成に失敗しました",
			nil,
		)
	}

	if externalStorageLinkID == 0 {
		return models.PersonalInformationDriveFolder{}, results.BadRequest(
			"BUILD_UPDATE_PERSONAL_INFORMATION_DRIVE_FOLDER_MODEL_INVALID_EXTERNAL_STORAGE_LINK_ID",
			"個人情報Driveフォルダ更新データの作成に失敗しました",
			map[string]any{"externalStorageLinkId": externalStorageLinkID},
		)
	}

	if folderName == "" || driveFolderID == "" || folderURL == "" {
		return models.PersonalInformationDriveFolder{}, results.BadRequest(
			"BUILD_UPDATE_PERSONAL_INFORMATION_DRIVE_FOLDER_MODEL_EMPTY_DRIVE_VALUE",
			"個人情報Driveフォルダ更新データの作成に失敗しました",
			map[string]any{
				"folderName":    folderName,
				"driveFolderId": driveFolderID,
				"folderUrl":     folderURL,
			},
		)
	}

	currentFolder.ExternalStorageLinkID = externalStorageLinkID
	currentFolder.FolderName = folderName
	currentFolder.DriveFolderID = driveFolderID
	currentFolder.FolderURL = folderURL
	currentFolder.SyncedAt = &syncedAt
	currentFolder.IsDeleted = false

	return currentFolder, results.OK(nil, "BUILD_UPDATE_PERSONAL_INFORMATION_DRIVE_FOLDER_MODEL_SUCCESS", "", nil)
}

/*
 * 個人情報Driveフォルダ検索条件をGORMクエリへ適用する
 */
func applySearchPersonalInformationDriveFoldersCondition(query *gorm.DB, req types.SearchPersonalInformationDriveFoldersRequest) *gorm.DB {
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where(
			"users.name ILIKE ? OR users.email ILIKE ? OR users.role ILIKE ?",
			keyword,
			keyword,
			keyword,
		)
	}

	return query
}
