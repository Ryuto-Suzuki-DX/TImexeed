package services

import (
	"context"
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/storage"
	"timexeed/backend/internal/utils"
)

const sharedDocumentDriveRootLinkType = "SHARED_DOCUMENT_DRIVE_ROOT"

/*
 * 管理者用 共有資料DriveフォルダService interface
 */
type SharedDocumentDriveFolderService interface {
	SearchSharedDocumentDriveFolders(req types.SearchSharedDocumentDriveFoldersRequest) results.Result
	DetailSharedDocumentDriveFolder(req types.SharedDocumentDriveFolderDetailRequest) results.Result
	CreateSharedDocumentDriveFolder(req types.CreateSharedDocumentDriveFolderRequest) results.Result
	UpdateSharedDocumentDriveFolder(req types.UpdateSharedDocumentDriveFolderRequest) results.Result
	DeleteSharedDocumentDriveFolder(req types.DeleteSharedDocumentDriveFolderRequest) results.Result
	SyncSharedDocumentDriveFolder(req types.SyncSharedDocumentDriveFolderRequest) results.Result
}

/*
 * 管理者用 共有資料DriveフォルダService
 */
type sharedDocumentDriveFolderService struct {
	sharedDocumentDriveFolderBuilder    builders.SharedDocumentDriveFolderBuilder
	sharedDocumentDriveFolderRepository repositories.SharedDocumentDriveFolderRepository
	googleDriveService                  storage.GoogleDriveService
}

/*
 * SharedDocumentDriveFolderService生成
 */
func NewSharedDocumentDriveFolderService(
	sharedDocumentDriveFolderBuilder builders.SharedDocumentDriveFolderBuilder,
	sharedDocumentDriveFolderRepository repositories.SharedDocumentDriveFolderRepository,
	googleDriveService storage.GoogleDriveService,
) *sharedDocumentDriveFolderService {
	return &sharedDocumentDriveFolderService{
		sharedDocumentDriveFolderBuilder:    sharedDocumentDriveFolderBuilder,
		sharedDocumentDriveFolderRepository: sharedDocumentDriveFolderRepository,
		googleDriveService:                  googleDriveService,
	}
}

/*
 * models.SharedDocumentDriveFolderをResponseへ変換する
 */
func toSharedDocumentDriveFolderResponse(folder models.SharedDocumentDriveFolder) types.SharedDocumentDriveFolderResponse {
	return types.SharedDocumentDriveFolderResponse{
		ID: folder.ID,

		FolderName:    folder.FolderName,
		Description:   folder.Description,
		DriveFolderID: folder.DriveFolderID,
		FolderURL:     folder.FolderURL,
		SyncedAt:      folder.SyncedAt,

		CreatedAt: folder.CreatedAt,
		UpdatedAt: folder.UpdatedAt,
	}
}

/*
 * models.SharedDocumentDriveFolder一覧をResponse一覧へ変換する
 */
func toSharedDocumentDriveFolderResponses(folders []models.SharedDocumentDriveFolder) []types.SharedDocumentDriveFolderResponse {
	responses := make([]types.SharedDocumentDriveFolderResponse, 0, len(folders))
	for _, folder := range folders {
		responses = append(responses, toSharedDocumentDriveFolderResponse(folder))
	}

	return responses
}

/*
 * 検索
 */
func (service *sharedDocumentDriveFolderService) SearchSharedDocumentDriveFolders(req types.SearchSharedDocumentDriveFoldersRequest) results.Result {
	normalizedCondition, normalizeResult := utils.NormalizePageSearchCondition(
		utils.PageSearchCondition{
			Keyword: req.Keyword,
			Offset:  req.Offset,
			Limit:   req.Limit,
		},
		"SEARCH_SHARED_DOCUMENT_DRIVE_FOLDERS_INVALID_OFFSET",
		"検索開始位置が正しくありません",
	)
	if normalizeResult.Error {
		return normalizeResult
	}

	req.Keyword = normalizedCondition.Keyword
	req.Offset = normalizedCondition.Offset
	req.Limit = normalizedCondition.Limit

	searchQuery, countQuery, buildResult := service.sharedDocumentDriveFolderBuilder.BuildSearchSharedDocumentDriveFoldersQuery(req)
	if buildResult.Error {
		return buildResult
	}

	rows, findResult := service.sharedDocumentDriveFolderRepository.FindSharedDocumentDriveFolderRows(searchQuery)
	if findResult.Error {
		return findResult
	}

	total, countResult := service.sharedDocumentDriveFolderRepository.CountSharedDocumentDriveFolderRows(countQuery)
	if countResult.Error {
		return countResult
	}

	hasMore := utils.HasMore(total, req.Offset, len(rows))

	return results.OK(
		types.SearchSharedDocumentDriveFoldersResponse{
			SharedDocumentDriveFolders: rows,
			Total:                      total,
			Offset:                     req.Offset,
			Limit:                      req.Limit,
			HasMore:                    hasMore,
		},
		"SEARCH_SHARED_DOCUMENT_DRIVE_FOLDERS_SUCCESS",
		"共有資料Driveフォルダ一覧を取得しました",
		nil,
	)
}

/*
 * 詳細
 */
func (service *sharedDocumentDriveFolderService) DetailSharedDocumentDriveFolder(req types.SharedDocumentDriveFolderDetailRequest) results.Result {
	folder, folderResult := service.findCurrentSharedDocumentDriveFolder(req.TargetSharedDocumentDriveFolderID)
	if folderResult.Error {
		return folderResult
	}

	return results.OK(
		types.SharedDocumentDriveFolderDetailResponse{
			SharedDocumentDriveFolder: toSharedDocumentDriveFolderResponse(folder),
		},
		"DETAIL_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS",
		"共有資料Driveフォルダを取得しました",
		nil,
	)
}

/*
 * 作成
 *
 * external_storage_links の SHARED_DOCUMENT_DRIVE_ROOT から親フォルダURLを取得し、
 * その親フォルダ配下に指定名のGoogle Driveフォルダを作成する。
 */
func (service *sharedDocumentDriveFolderService) CreateSharedDocumentDriveFolder(req types.CreateSharedDocumentDriveFolderRequest) results.Result {
	if service.googleDriveService == nil {
		return results.InternalServerError(
			"CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_GOOGLE_DRIVE_SERVICE_NOT_CONFIGURED",
			"Google Drive連携が設定されていません",
			nil,
		)
	}

	folderName := strings.TrimSpace(req.FolderName)
	if folderName == "" {
		return results.BadRequest(
			"CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_EMPTY_FOLDER_NAME",
			"共有資料Driveフォルダ名が入力されていません",
			nil,
		)
	}

	rootExternalStorageLink, rootResult := service.findSharedDocumentDriveRootExternalStorageLink()
	if rootResult.Error {
		return rootResult
	}

	parentFolderID, parseErr := service.googleDriveService.ParseFolderID(rootExternalStorageLink.URL)
	if parseErr != nil {
		return results.BadRequest(
			"CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_ROOT_FOLDER_URL_INVALID",
			"共有資料Drive親フォルダURLの形式が正しくありません",
			parseErr.Error(),
		)
	}

	folderMetadata, createFolderErr := service.googleDriveService.CreateFolder(context.Background(), parentFolderID, folderName)
	if createFolderErr != nil {
		return results.InternalServerError(
			"CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_CREATE_GOOGLE_DRIVE_FOLDER_FAILED",
			"Google Drive上の共有資料フォルダ作成に失敗しました",
			createFolderErr.Error(),
		)
	}

	createdFolderModel, buildResult := service.sharedDocumentDriveFolderBuilder.BuildCreateSharedDocumentDriveFolderModel(
		folderName,
		req.Description,
		folderMetadata.DriveFolderID,
		folderMetadata.FolderURL,
	)
	if buildResult.Error {
		return buildResult
	}

	createdFolder, createResult := service.sharedDocumentDriveFolderRepository.CreateSharedDocumentDriveFolder(createdFolderModel)
	if createResult.Error {
		return createResult
	}

	return results.Created(
		types.CreateSharedDocumentDriveFolderResponse{
			SharedDocumentDriveFolder: toSharedDocumentDriveFolderResponse(createdFolder),
		},
		"CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS",
		"共有資料Driveフォルダを作成しました",
		nil,
	)
}

/*
 * 更新
 *
 * Drive上のフォルダID/URLは変更しない。
 * 表示名・説明のみDB上で更新する。
 */
func (service *sharedDocumentDriveFolderService) UpdateSharedDocumentDriveFolder(req types.UpdateSharedDocumentDriveFolderRequest) results.Result {
	currentFolder, folderResult := service.findCurrentSharedDocumentDriveFolder(req.TargetSharedDocumentDriveFolderID)
	if folderResult.Error {
		return folderResult
	}

	updatedFolderModel, buildResult := service.sharedDocumentDriveFolderBuilder.BuildUpdateSharedDocumentDriveFolderModel(
		currentFolder,
		req.FolderName,
		req.Description,
	)
	if buildResult.Error {
		return buildResult
	}

	updatedFolder, saveResult := service.sharedDocumentDriveFolderRepository.SaveSharedDocumentDriveFolder(updatedFolderModel)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.UpdateSharedDocumentDriveFolderResponse{
			SharedDocumentDriveFolder: toSharedDocumentDriveFolderResponse(updatedFolder),
		},
		"UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS",
		"共有資料Driveフォルダを更新しました",
		nil,
	)
}

/*
 * 削除
 *
 * DBは論理削除。
 * Drive上のフォルダ自体は削除しない。
 */
func (service *sharedDocumentDriveFolderService) DeleteSharedDocumentDriveFolder(req types.DeleteSharedDocumentDriveFolderRequest) results.Result {
	currentFolder, folderResult := service.findCurrentSharedDocumentDriveFolder(req.TargetSharedDocumentDriveFolderID)
	if folderResult.Error {
		return folderResult
	}

	deletedFolderModel, buildResult := service.sharedDocumentDriveFolderBuilder.BuildDeleteSharedDocumentDriveFolderModel(currentFolder)
	if buildResult.Error {
		return buildResult
	}

	deletedFolder, saveResult := service.sharedDocumentDriveFolderRepository.SaveSharedDocumentDriveFolder(deletedFolderModel)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteSharedDocumentDriveFolderResponse{
			SharedDocumentDriveFolderID: deletedFolder.ID,
		},
		"DELETE_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS",
		"共有資料Driveフォルダを削除しました",
		nil,
	)
}

/*
 * 同期
 *
 * targetSharedDocumentDriveFolderId = 0:
 * ・有効な共有資料Driveフォルダ全件を同期する
 *
 * targetSharedDocumentDriveFolderId > 0:
 * ・指定された共有資料Driveフォルダ1件だけ同期する
 *
 * 権限：
 * ・有効なADMIN: writer
 * ・有効なUSER: reader
 *
 * 注意：
 * ・Drive上に手動で付いている直接権限は削除しない
 * ・完全同期で権限削除まで行いたい場合は、SyncPermissions の第4引数を true にする
 */
func (service *sharedDocumentDriveFolderService) SyncSharedDocumentDriveFolder(req types.SyncSharedDocumentDriveFolderRequest) results.Result {
	if service.googleDriveService == nil {
		return results.InternalServerError(
			"SYNC_SHARED_DOCUMENT_DRIVE_FOLDER_GOOGLE_DRIVE_SERVICE_NOT_CONFIGURED",
			"Google Drive連携が設定されていません",
			nil,
		)
	}

	folders, foldersResult := service.findSyncTargetSharedDocumentDriveFolders(req.TargetSharedDocumentDriveFolderID)
	if foldersResult.Error {
		return foldersResult
	}

	if len(folders) == 0 {
		return results.BadRequest(
			"SYNC_SHARED_DOCUMENT_DRIVE_FOLDER_TARGET_EMPTY",
			"同期対象の共有資料Driveフォルダがありません",
			nil,
		)
	}

	adminUsers, adminUsersResult := service.findAllActiveAdminUsers()
	if adminUsersResult.Error {
		return adminUsersResult
	}

	generalUsers, generalUsersResult := service.findAllActiveUsers()
	if generalUsersResult.Error {
		return generalUsersResult
	}

	permissions, permissionsResult := service.buildSharedDocumentDriveFolderPermissions(adminUsers, generalUsers)
	if permissionsResult.Error {
		return permissionsResult
	}

	syncedAt := time.Now()
	syncedFolders := make([]models.SharedDocumentDriveFolder, 0, len(folders))

	for _, folder := range folders {
		if err := service.googleDriveService.SyncPermissions(context.Background(), folder.DriveFolderID, permissions, false); err != nil {
			return results.InternalServerError(
				"SYNC_SHARED_DOCUMENT_DRIVE_FOLDER_PERMISSIONS_FAILED",
				"共有資料Driveフォルダの権限同期に失敗しました",
				map[string]any{
					"sharedDocumentDriveFolderId": folder.ID,
					"driveFolderId":               folder.DriveFolderID,
					"error":                       err.Error(),
				},
			)
		}

		syncedFolderModel, buildSyncedFolderResult := service.sharedDocumentDriveFolderBuilder.BuildSyncedSharedDocumentDriveFolderModel(folder, syncedAt)
		if buildSyncedFolderResult.Error {
			return buildSyncedFolderResult
		}

		syncedFolder, saveFolderResult := service.sharedDocumentDriveFolderRepository.SaveSharedDocumentDriveFolder(syncedFolderModel)
		if saveFolderResult.Error {
			return saveFolderResult
		}

		syncedFolders = append(syncedFolders, syncedFolder)
	}

	return results.OK(
		types.SyncSharedDocumentDriveFolderResponse{
			SharedDocumentDriveFolders: toSharedDocumentDriveFolderResponses(syncedFolders),
			SyncedFolderCount:          len(syncedFolders),
			TargetAdminCount:           len(adminUsers),
			TargetUserCount:            len(generalUsers),
			SyncedAt:                   syncedAt,
		},
		"SYNC_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS",
		"共有資料Driveフォルダの権限を同期しました",
		nil,
	)
}

/*
 * 共有資料Driveフォルダ取得
 */
func (service *sharedDocumentDriveFolderService) findCurrentSharedDocumentDriveFolder(folderID uint) (models.SharedDocumentDriveFolder, results.Result) {
	query, buildResult := service.sharedDocumentDriveFolderBuilder.BuildFindActiveSharedDocumentDriveFolderByIDQuery(folderID)
	if buildResult.Error {
		return models.SharedDocumentDriveFolder{}, buildResult
	}

	folder, findResult := service.sharedDocumentDriveFolderRepository.FindSharedDocumentDriveFolder(query)
	if findResult.Error {
		return models.SharedDocumentDriveFolder{}, findResult
	}

	return folder, results.OK(nil, "FIND_CURRENT_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * 同期対象の共有資料Driveフォルダ取得
 */
func (service *sharedDocumentDriveFolderService) findSyncTargetSharedDocumentDriveFolders(targetFolderID uint) ([]models.SharedDocumentDriveFolder, results.Result) {
	if targetFolderID > 0 {
		folder, folderResult := service.findCurrentSharedDocumentDriveFolder(targetFolderID)
		if folderResult.Error {
			return nil, folderResult
		}

		return []models.SharedDocumentDriveFolder{folder}, results.OK(nil, "FIND_SYNC_TARGET_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS", "", nil)
	}

	query, buildResult := service.sharedDocumentDriveFolderBuilder.BuildFindAllActiveSharedDocumentDriveFoldersQuery()
	if buildResult.Error {
		return nil, buildResult
	}

	folders, findResult := service.sharedDocumentDriveFolderRepository.FindSharedDocumentDriveFolders(query)
	if findResult.Error {
		return nil, findResult
	}

	return folders, results.OK(nil, "FIND_SYNC_TARGET_SHARED_DOCUMENT_DRIVE_FOLDERS_SUCCESS", "", nil)
}

/*
 * 共有資料Drive親フォルダの外部ストレージリンク取得
 */
func (service *sharedDocumentDriveFolderService) findSharedDocumentDriveRootExternalStorageLink() (models.ExternalStorageLink, results.Result) {
	query, buildResult := service.sharedDocumentDriveFolderBuilder.BuildFindActiveExternalStorageLinkByLinkTypeQuery(sharedDocumentDriveRootLinkType)
	if buildResult.Error {
		return models.ExternalStorageLink{}, buildResult
	}

	externalStorageLink, findResult := service.sharedDocumentDriveFolderRepository.FindExternalStorageLink(query)
	if findResult.Error {
		return models.ExternalStorageLink{}, findResult
	}

	if strings.TrimSpace(externalStorageLink.URL) == "" {
		return models.ExternalStorageLink{}, results.BadRequest(
			"SHARED_DOCUMENT_DRIVE_ROOT_EXTERNAL_STORAGE_LINK_URL_EMPTY",
			"共有資料Drive親フォルダURLが設定されていません",
			nil,
		)
	}

	return externalStorageLink, results.OK(nil, "FIND_SHARED_DOCUMENT_DRIVE_ROOT_EXTERNAL_STORAGE_LINK_SUCCESS", "", nil)
}

/*
 * 有効なUSER全員取得
 */
func (service *sharedDocumentDriveFolderService) findAllActiveUsers() ([]models.User, results.Result) {
	query, buildResult := service.sharedDocumentDriveFolderBuilder.BuildFindAllActiveUsersQuery()
	if buildResult.Error {
		return nil, buildResult
	}

	users, findResult := service.sharedDocumentDriveFolderRepository.FindUsers(query)
	if findResult.Error {
		return nil, findResult
	}

	return users, results.OK(nil, "FIND_ALL_ACTIVE_USERS_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * 有効なADMIN全員取得
 */
func (service *sharedDocumentDriveFolderService) findAllActiveAdminUsers() ([]models.User, results.Result) {
	query, buildResult := service.sharedDocumentDriveFolderBuilder.BuildFindActiveAdminUsersQuery()
	if buildResult.Error {
		return nil, buildResult
	}

	users, findResult := service.sharedDocumentDriveFolderRepository.FindUsers(query)
	if findResult.Error {
		return nil, findResult
	}

	return users, results.OK(nil, "FIND_ALL_ACTIVE_ADMINS_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * Drive権限作成
 *
 * 管理者: writer
 * 一般ユーザー: reader
 */
func (service *sharedDocumentDriveFolderService) buildSharedDocumentDriveFolderPermissions(adminUsers []models.User, generalUsers []models.User) ([]storage.GoogleDrivePermissionSetting, results.Result) {
	permissions := make([]storage.GoogleDrivePermissionSetting, 0, len(adminUsers)+len(generalUsers))
	emailMap := map[string]bool{}

	appendPermission := func(email string, role string) {
		email = strings.ToLower(strings.TrimSpace(email))
		role = strings.TrimSpace(role)
		if email == "" || role == "" {
			return
		}

		if emailMap[email] {
			return
		}

		emailMap[email] = true
		permissions = append(permissions, storage.GoogleDrivePermissionSetting{
			EmailAddress: email,
			Role:         role,
		})
	}

	for _, adminUser := range adminUsers {
		appendPermission(adminUser.Email, "writer")
	}

	for _, generalUser := range generalUsers {
		appendPermission(generalUser.Email, "reader")
	}

	if len(permissions) == 0 {
		return nil, results.BadRequest(
			"BUILD_SHARED_DOCUMENT_DRIVE_FOLDER_PERMISSIONS_EMPTY",
			"共有資料Driveフォルダの同期対象ユーザーがいません",
			nil,
		)
	}

	return permissions, results.OK(nil, "BUILD_SHARED_DOCUMENT_DRIVE_FOLDER_PERMISSIONS_SUCCESS", "", nil)
}
