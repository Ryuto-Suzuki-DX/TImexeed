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

/*
 * 管理者用 共有資料DriveフォルダService interface
 */
type SharedDocumentDriveFolderService interface {
	SearchSharedDocumentDriveFolders(req types.SearchSharedDocumentDriveFoldersRequest) results.Result
	DetailSharedDocumentDriveFolder(req types.SharedDocumentDriveFolderDetailRequest) results.Result
	CreateSharedDocumentDriveFolder(req types.CreateSharedDocumentDriveFolderRequest) results.Result
	UpdateSharedDocumentDriveFolder(req types.UpdateSharedDocumentDriveFolderRequest) results.Result
	DeleteSharedDocumentDriveFolder(req types.DeleteSharedDocumentDriveFolderRequest) results.Result
	UpdateSharedDocumentDriveFolderUsers(req types.UpdateSharedDocumentDriveFolderUsersRequest) results.Result
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

	sharedUsers, sharedUsersResult := service.findActiveSharedDocumentDriveFolderUserRows(folder.ID)
	if sharedUsersResult.Error {
		return sharedUsersResult
	}

	return results.OK(
		types.SharedDocumentDriveFolderDetailResponse{
			SharedDocumentDriveFolder: toSharedDocumentDriveFolderResponse(folder),
			SharedUsers:               sharedUsers,
		},
		"DETAIL_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS",
		"共有資料Driveフォルダを取得しました",
		nil,
	)
}

/*
 * 作成
 *
 * 管理者が作成済みのGoogle DriveフォルダURL/IDを登録する。
 * アプリ用GoogleアカウントにDrive上の編集権限がある前提。
 */
func (service *sharedDocumentDriveFolderService) CreateSharedDocumentDriveFolder(req types.CreateSharedDocumentDriveFolderRequest) results.Result {
	if service.googleDriveService == nil {
		return results.InternalServerError(
			"CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_GOOGLE_DRIVE_SERVICE_NOT_CONFIGURED",
			"Google Drive連携が設定されていません",
			nil,
		)
	}

	folderID, parseResult := service.parseFolderID(req.DriveFolderURLOrID)
	if parseResult.Error {
		return parseResult
	}

	folderMetadata, metadataErr := service.googleDriveService.GetFolderMetadata(context.Background(), folderID)
	if metadataErr != nil {
		return results.InternalServerError(
			"CREATE_SHARED_DOCUMENT_DRIVE_FOLDER_GET_FOLDER_METADATA_FAILED",
			"共有資料Driveフォルダの確認に失敗しました",
			metadataErr.Error(),
		)
	}

	folderName := strings.TrimSpace(req.FolderName)
	if folderName == "" {
		folderName = folderMetadata.FolderName
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
		"共有資料Driveフォルダを登録しました",
		nil,
	)
}

/*
 * 更新
 */
func (service *sharedDocumentDriveFolderService) UpdateSharedDocumentDriveFolder(req types.UpdateSharedDocumentDriveFolderRequest) results.Result {
	if service.googleDriveService == nil {
		return results.InternalServerError(
			"UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_GOOGLE_DRIVE_SERVICE_NOT_CONFIGURED",
			"Google Drive連携が設定されていません",
			nil,
		)
	}

	currentFolder, folderResult := service.findCurrentSharedDocumentDriveFolder(req.TargetSharedDocumentDriveFolderID)
	if folderResult.Error {
		return folderResult
	}

	folderID, parseResult := service.parseFolderID(req.DriveFolderURLOrID)
	if parseResult.Error {
		return parseResult
	}

	folderMetadata, metadataErr := service.googleDriveService.GetFolderMetadata(context.Background(), folderID)
	if metadataErr != nil {
		return results.InternalServerError(
			"UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_GET_FOLDER_METADATA_FAILED",
			"共有資料Driveフォルダの確認に失敗しました",
			metadataErr.Error(),
		)
	}

	folderName := strings.TrimSpace(req.FolderName)
	if folderName == "" {
		folderName = folderMetadata.FolderName
	}

	updatedFolderModel, buildResult := service.sharedDocumentDriveFolderBuilder.BuildUpdateSharedDocumentDriveFolderModel(
		currentFolder,
		folderName,
		req.Description,
		folderMetadata.DriveFolderID,
		folderMetadata.FolderURL,
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
 * 共有ユーザー更新
 *
 * 通常時：
 * ・targetUserIds を正として、共有対象を差し替える
 *
 * shareAllUsers = true の場合：
 * ・targetUserIds は無視する
 * ・有効なUSER全員を共有対象にする
 *
 * 注意：
 * ・targetUserIds = [] かつ shareAllUsers = false の場合、共有対象を全削除する
 * ・このAPIはDB上の共有対象だけを更新する
 * ・Drive権限へ反映するには SyncSharedDocumentDriveFolder を呼び出す
 */
func (service *sharedDocumentDriveFolderService) UpdateSharedDocumentDriveFolderUsers(req types.UpdateSharedDocumentDriveFolderUsersRequest) results.Result {
	folder, folderResult := service.findCurrentSharedDocumentDriveFolder(req.TargetSharedDocumentDriveFolderID)
	if folderResult.Error {
		return folderResult
	}

	targetUsers, targetUsersResult := service.resolveSharedDocumentDriveFolderTargetUsers(req)
	if targetUsersResult.Error {
		return targetUsersResult
	}

	targetUserIDs := make([]uint, 0, len(targetUsers))
	for _, targetUser := range targetUsers {
		targetUserIDs = append(targetUserIDs, targetUser.ID)
	}

	currentFolderUsers, currentUsersResult := service.findAllSharedDocumentDriveFolderUsers(folder.ID)
	if currentUsersResult.Error {
		return currentUsersResult
	}

	targetUserIDMap := sharedDocumentDriveFolderUintSliceToMap(targetUserIDs)
	currentUserMap := map[uint]models.SharedDocumentDriveFolderUser{}

	for _, currentFolderUser := range currentFolderUsers {
		currentUserMap[currentFolderUser.UserID] = currentFolderUser
	}

	for _, targetUserID := range targetUserIDs {
		currentFolderUser, exists := currentUserMap[targetUserID]
		if exists {
			activeModel, buildActiveResult := service.sharedDocumentDriveFolderBuilder.BuildActiveSharedDocumentDriveFolderUserModel(currentFolderUser)
			if buildActiveResult.Error {
				return buildActiveResult
			}

			if _, saveResult := service.sharedDocumentDriveFolderRepository.SaveSharedDocumentDriveFolderUser(activeModel); saveResult.Error {
				return saveResult
			}

			continue
		}

		createModel, buildCreateResult := service.sharedDocumentDriveFolderBuilder.BuildCreateSharedDocumentDriveFolderUserModel(folder.ID, targetUserID)
		if buildCreateResult.Error {
			return buildCreateResult
		}

		if _, createResult := service.sharedDocumentDriveFolderRepository.CreateSharedDocumentDriveFolderUser(createModel); createResult.Error {
			return createResult
		}
	}

	for _, currentFolderUser := range currentFolderUsers {
		if targetUserIDMap[currentFolderUser.UserID] {
			continue
		}

		if currentFolderUser.IsDeleted {
			continue
		}

		deletedModel, buildDeletedResult := service.sharedDocumentDriveFolderBuilder.BuildDeletedSharedDocumentDriveFolderUserModel(currentFolderUser)
		if buildDeletedResult.Error {
			return buildDeletedResult
		}

		if _, saveResult := service.sharedDocumentDriveFolderRepository.SaveSharedDocumentDriveFolderUser(deletedModel); saveResult.Error {
			return saveResult
		}
	}

	sharedUsers, sharedUsersResult := service.findActiveSharedDocumentDriveFolderUserRows(folder.ID)
	if sharedUsersResult.Error {
		return sharedUsersResult
	}

	return results.OK(
		types.UpdateSharedDocumentDriveFolderUsersResponse{
			SharedDocumentDriveFolderID: folder.ID,
			SharedUsers:                 sharedUsers,
		},
		"UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_USERS_SUCCESS",
		"共有対象ユーザーを更新しました",
		nil,
	)
}

/*
 * 同期
 *
 * 管理者全員: writer
 * 共有対象ユーザー: writer
 *
 * 注意：
 * ・共有資料フォルダ同期では、Timexeedで指定したユーザーに権限を追加/更新する
 * ・Drive上に手動で付いている直接権限は削除しない
 * ・完全同期で権限削除まで行いたい場合は、SyncPermissions の第4引数を true に戻す
 */
func (service *sharedDocumentDriveFolderService) SyncSharedDocumentDriveFolder(req types.SyncSharedDocumentDriveFolderRequest) results.Result {
	if service.googleDriveService == nil {
		return results.InternalServerError(
			"SYNC_SHARED_DOCUMENT_DRIVE_FOLDER_GOOGLE_DRIVE_SERVICE_NOT_CONFIGURED",
			"Google Drive連携が設定されていません",
			nil,
		)
	}

	folder, folderResult := service.findCurrentSharedDocumentDriveFolder(req.TargetSharedDocumentDriveFolderID)
	if folderResult.Error {
		return folderResult
	}

	sharedUsers, sharedUsersResult := service.findActiveSharedDocumentDriveFolderUsers(folder.ID)
	if sharedUsersResult.Error {
		return sharedUsersResult
	}

	permissions, permissionsResult := service.buildSharedDocumentDriveFolderPermissions(sharedUsers)
	if permissionsResult.Error {
		return permissionsResult
	}

	if err := service.googleDriveService.SyncPermissions(context.Background(), folder.DriveFolderID, permissions, false); err != nil {
		return results.InternalServerError(
			"SYNC_SHARED_DOCUMENT_DRIVE_FOLDER_PERMISSIONS_FAILED",
			"共有資料Driveフォルダの権限同期に失敗しました",
			err.Error(),
		)
	}

	now := time.Now()

	syncedFolderModel, buildSyncedFolderResult := service.sharedDocumentDriveFolderBuilder.BuildSyncedSharedDocumentDriveFolderModel(folder, now)
	if buildSyncedFolderResult.Error {
		return buildSyncedFolderResult
	}

	syncedFolder, saveFolderResult := service.sharedDocumentDriveFolderRepository.SaveSharedDocumentDriveFolder(syncedFolderModel)
	if saveFolderResult.Error {
		return saveFolderResult
	}

	for _, sharedUser := range sharedUsers {
		syncedUserModel, buildSyncedUserResult := service.sharedDocumentDriveFolderBuilder.BuildSyncedSharedDocumentDriveFolderUserModel(sharedUser, now)
		if buildSyncedUserResult.Error {
			return buildSyncedUserResult
		}

		if _, saveUserResult := service.sharedDocumentDriveFolderRepository.SaveSharedDocumentDriveFolderUser(syncedUserModel); saveUserResult.Error {
			return saveUserResult
		}
	}

	sharedUserRows, sharedUserRowsResult := service.findActiveSharedDocumentDriveFolderUserRows(folder.ID)
	if sharedUserRowsResult.Error {
		return sharedUserRowsResult
	}

	return results.OK(
		types.SyncSharedDocumentDriveFolderResponse{
			SharedDocumentDriveFolder: toSharedDocumentDriveFolderResponse(syncedFolder),
			SharedUsers:               sharedUserRows,
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
 * 有効な共有ユーザー表示用取得
 */
func (service *sharedDocumentDriveFolderService) findActiveSharedDocumentDriveFolderUserRows(folderID uint) ([]types.SharedDocumentDriveFolderUserResponse, results.Result) {
	query, buildResult := service.sharedDocumentDriveFolderBuilder.BuildFindActiveSharedDocumentDriveFolderUsersByFolderIDQuery(folderID)
	if buildResult.Error {
		return nil, buildResult
	}

	rows, findResult := service.sharedDocumentDriveFolderRepository.FindSharedDocumentDriveFolderUserRows(query)
	if findResult.Error {
		return nil, findResult
	}

	return rows, results.OK(nil, "FIND_ACTIVE_SHARED_DOCUMENT_DRIVE_FOLDER_USER_ROWS_SUCCESS", "", nil)
}

/*
 * 有効な共有ユーザーModel取得
 */
func (service *sharedDocumentDriveFolderService) findActiveSharedDocumentDriveFolderUsers(folderID uint) ([]models.SharedDocumentDriveFolderUser, results.Result) {
	query, buildResult := service.sharedDocumentDriveFolderBuilder.BuildFindAllSharedDocumentDriveFolderUsersByFolderIDQuery(folderID)
	if buildResult.Error {
		return nil, buildResult
	}

	folderUsers, findResult := service.sharedDocumentDriveFolderRepository.FindSharedDocumentDriveFolderUsers(query)
	if findResult.Error {
		return nil, findResult
	}

	activeUsers := make([]models.SharedDocumentDriveFolderUser, 0, len(folderUsers))
	for _, folderUser := range folderUsers {
		if folderUser.IsDeleted {
			continue
		}

		activeUsers = append(activeUsers, folderUser)
	}

	return activeUsers, results.OK(nil, "FIND_ACTIVE_SHARED_DOCUMENT_DRIVE_FOLDER_USERS_SUCCESS", "", nil)
}

/*
 * 論理削除済み含む共有ユーザーModel取得
 */
func (service *sharedDocumentDriveFolderService) findAllSharedDocumentDriveFolderUsers(folderID uint) ([]models.SharedDocumentDriveFolderUser, results.Result) {
	query, buildResult := service.sharedDocumentDriveFolderBuilder.BuildFindAllSharedDocumentDriveFolderUsersByFolderIDQuery(folderID)
	if buildResult.Error {
		return nil, buildResult
	}

	folderUsers, findResult := service.sharedDocumentDriveFolderRepository.FindSharedDocumentDriveFolderUsers(query)
	if findResult.Error {
		return nil, findResult
	}

	return folderUsers, results.OK(nil, "FIND_ALL_SHARED_DOCUMENT_DRIVE_FOLDER_USERS_SUCCESS", "", nil)
}

/*
 * 共有対象ユーザー解決
 *
 * shareAllUsers = true:
 * ・有効なUSER全員を返す
 *
 * shareAllUsers = false:
 * ・targetUserIds の有効なUSERだけを返す
 */
func (service *sharedDocumentDriveFolderService) resolveSharedDocumentDriveFolderTargetUsers(req types.UpdateSharedDocumentDriveFolderUsersRequest) ([]models.User, results.Result) {
	if req.ShareAllUsers {
		users, usersResult := service.findAllActiveUsers()
		if usersResult.Error {
			return nil, usersResult
		}

		return users, results.OK(nil, "RESOLVE_SHARED_DOCUMENT_DRIVE_FOLDER_TARGET_USERS_ALL_SUCCESS", "", nil)
	}

	uniqueUserIDs := uniqueSharedDocumentDriveFolderServiceUserIDs(req.TargetUserIDs)
	validUsers, validUsersResult := service.findValidTargetUsers(uniqueUserIDs)
	if validUsersResult.Error {
		return nil, validUsersResult
	}

	if len(validUsers) != len(uniqueUserIDs) {
		return nil, results.BadRequest(
			"UPDATE_SHARED_DOCUMENT_DRIVE_FOLDER_USERS_CONTAINS_INVALID_USER",
			"共有対象ユーザーに無効なユーザーが含まれています",
			map[string]any{
				"requestedUserIds": uniqueUserIDs,
				"validUserCount":   len(validUsers),
			},
		)
	}

	return validUsers, results.OK(nil, "RESOLVE_SHARED_DOCUMENT_DRIVE_FOLDER_TARGET_USERS_SELECTED_SUCCESS", "", nil)
}

/*
 * 有効な対象USER取得
 */
func (service *sharedDocumentDriveFolderService) findValidTargetUsers(userIDs []uint) ([]models.User, results.Result) {
	if len(userIDs) == 0 {
		return []models.User{}, results.OK(nil, "FIND_VALID_TARGET_USERS_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_EMPTY", "", nil)
	}

	query, buildResult := service.sharedDocumentDriveFolderBuilder.BuildFindActiveUsersByIDsQuery(userIDs)
	if buildResult.Error {
		return nil, buildResult
	}

	users, findResult := service.sharedDocumentDriveFolderRepository.FindUsers(query)
	if findResult.Error {
		return nil, findResult
	}

	return users, results.OK(nil, "FIND_VALID_TARGET_USERS_FOR_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS", "", nil)
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
 * DriveフォルダID解析
 */
func (service *sharedDocumentDriveFolderService) parseFolderID(folderURLOrID string) (string, results.Result) {
	folderURLOrID = strings.TrimSpace(folderURLOrID)
	if folderURLOrID == "" {
		return "", results.BadRequest(
			"SHARED_DOCUMENT_DRIVE_FOLDER_URL_OR_ID_EMPTY",
			"共有資料DriveフォルダのURLまたはIDが入力されていません",
			nil,
		)
	}

	folderID, err := service.googleDriveService.ParseFolderID(folderURLOrID)
	if err != nil {
		return "", results.BadRequest(
			"SHARED_DOCUMENT_DRIVE_FOLDER_URL_OR_ID_INVALID",
			"共有資料DriveフォルダのURLまたはID形式が正しくありません",
			err.Error(),
		)
	}

	return folderID, results.OK(nil, "SHARED_DOCUMENT_DRIVE_FOLDER_ID_PARSE_SUCCESS", "", nil)
}

/*
 * Drive権限作成
 *
 * 管理者全員: writer
 * 共有対象ユーザー: writer
 */
func (service *sharedDocumentDriveFolderService) buildSharedDocumentDriveFolderPermissions(sharedFolderUsers []models.SharedDocumentDriveFolderUser) ([]storage.GoogleDrivePermissionSetting, results.Result) {
	adminQuery, buildAdminQueryResult := service.sharedDocumentDriveFolderBuilder.BuildFindActiveAdminUsersQuery()
	if buildAdminQueryResult.Error {
		return nil, buildAdminQueryResult
	}

	adminUsers, findAdminsResult := service.sharedDocumentDriveFolderRepository.FindUsers(adminQuery)
	if findAdminsResult.Error {
		return nil, findAdminsResult
	}

	targetUserIDs := make([]uint, 0, len(sharedFolderUsers))
	for _, sharedFolderUser := range sharedFolderUsers {
		targetUserIDs = append(targetUserIDs, sharedFolderUser.UserID)
	}

	targetUsers, targetUsersResult := service.findValidTargetUsers(targetUserIDs)
	if targetUsersResult.Error {
		return nil, targetUsersResult
	}

	permissions := make([]storage.GoogleDrivePermissionSetting, 0, len(adminUsers)+len(targetUsers))
	emailMap := map[string]bool{}

	appendWriterPermission := func(email string) {
		email = strings.TrimSpace(email)
		if email == "" {
			return
		}

		if emailMap[email] {
			return
		}

		emailMap[email] = true
		permissions = append(permissions, storage.GoogleDrivePermissionSetting{
			EmailAddress: email,
			Role:         "writer",
		})
	}

	for _, adminUser := range adminUsers {
		appendWriterPermission(adminUser.Email)
	}

	for _, targetUser := range targetUsers {
		appendWriterPermission(targetUser.Email)
	}

	return permissions, results.OK(nil, "BUILD_SHARED_DOCUMENT_DRIVE_FOLDER_PERMISSIONS_SUCCESS", "", nil)
}

/*
 * uint重複排除
 */
func uniqueSharedDocumentDriveFolderServiceUserIDs(values []uint) []uint {
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

/*
 * uint slice to map
 */
func sharedDocumentDriveFolderUintSliceToMap(values []uint) map[uint]bool {
	valueMap := map[uint]bool{}

	for _, value := range values {
		valueMap[value] = true
	}

	return valueMap
}
