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
 * 管理者用 個人情報DriveフォルダService interface
 */
type PersonalInformationDriveFolderService interface {
	SearchPersonalInformationDriveFolders(req types.SearchPersonalInformationDriveFoldersRequest) results.Result
	SyncPersonalInformationDriveFolder(req types.SyncPersonalInformationDriveFolderRequest) results.Result
	ViewPersonalInformationDriveFolder(req types.ViewPersonalInformationDriveFolderRequest) results.Result
}

/*
 * 管理者用 個人情報DriveフォルダService
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや保存用Modelを作成する
 * ・RepositoryでDB処理を実行する
 * ・Google Drive上のフォルダ作成/権限同期をStorage層に依頼する
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 */
type personalInformationDriveFolderService struct {
	personalInformationDriveFolderBuilder    builders.PersonalInformationDriveFolderBuilder
	personalInformationDriveFolderRepository repositories.PersonalInformationDriveFolderRepository
	googleDriveService                       storage.GoogleDriveService
}

/*
 * PersonalInformationDriveFolderService生成
 */
func NewPersonalInformationDriveFolderService(
	personalInformationDriveFolderBuilder builders.PersonalInformationDriveFolderBuilder,
	personalInformationDriveFolderRepository repositories.PersonalInformationDriveFolderRepository,
	googleDriveService storage.GoogleDriveService,
) *personalInformationDriveFolderService {
	return &personalInformationDriveFolderService{
		personalInformationDriveFolderBuilder:    personalInformationDriveFolderBuilder,
		personalInformationDriveFolderRepository: personalInformationDriveFolderRepository,
		googleDriveService:                       googleDriveService,
	}
}

/*
 * models.PersonalInformationDriveFolderをResponseへ変換する
 */
func toPersonalInformationDriveFolderResponse(
	folder models.PersonalInformationDriveFolder,
	user models.User,
) types.PersonalInformationDriveFolderResponse {
	return types.PersonalInformationDriveFolderResponse{
		ID: folder.ID,

		UserID:    user.ID,
		UserName:  user.Name,
		UserEmail: user.Email,

		ExternalStorageLinkID: folder.ExternalStorageLinkID,
		FolderName:            folder.FolderName,
		DriveFolderID:         folder.DriveFolderID,
		FolderURL:             folder.FolderURL,
		SyncedAt:              folder.SyncedAt,

		CreatedAt: folder.CreatedAt,
		UpdatedAt: folder.UpdatedAt,
	}
}

/*
 * 検索
 *
 * 管理者用。
 * フォルダ未作成ユーザーも出すため、users主軸で検索する。
 */
func (service *personalInformationDriveFolderService) SearchPersonalInformationDriveFolders(req types.SearchPersonalInformationDriveFoldersRequest) results.Result {
	normalizedCondition, normalizeResult := utils.NormalizePageSearchCondition(
		utils.PageSearchCondition{
			Keyword: req.Keyword,
			Offset:  req.Offset,
			Limit:   req.Limit,
		},
		"SEARCH_PERSONAL_INFORMATION_DRIVE_FOLDERS_INVALID_OFFSET",
		"検索開始位置が正しくありません",
	)
	if normalizeResult.Error {
		return normalizeResult
	}

	req.Keyword = normalizedCondition.Keyword
	req.Offset = normalizedCondition.Offset
	req.Limit = normalizedCondition.Limit

	searchQuery, countQuery, buildResult := service.personalInformationDriveFolderBuilder.BuildSearchPersonalInformationDriveFoldersQuery(req)
	if buildResult.Error {
		return buildResult
	}

	rows, findResult := service.personalInformationDriveFolderRepository.FindPersonalInformationDriveFolderRows(searchQuery)
	if findResult.Error {
		return findResult
	}

	total, countResult := service.personalInformationDriveFolderRepository.CountPersonalInformationDriveFolderRows(countQuery)
	if countResult.Error {
		return countResult
	}

	hasMore := utils.HasMore(total, req.Offset, len(rows))

	return results.OK(
		types.SearchPersonalInformationDriveFoldersResponse{
			PersonalInformationDriveFolders: rows,
			Total:                           total,
			Offset:                          req.Offset,
			Limit:                           req.Limit,
			HasMore:                         hasMore,
		},
		"SEARCH_PERSONAL_INFORMATION_DRIVE_FOLDERS_SUCCESS",
		"個人情報Driveフォルダ一覧を取得しました",
		nil,
	)
}

/*
 * 個人情報Driveフォルダ作成/権限同期
 *
 * 処理内容：
 * 1. 対象ユーザーを取得
 * 2. PERSONAL_INFORMATION_DRIVE_ROOT の外部ストレージリンクを取得
 * 3. 対象ユーザー用フォルダがDB未登録ならDrive上に作成
 * 4. 管理者全員と対象ユーザー本人のDrive権限を同期
 * 5. DBのDriveフォルダ情報と syncedAt を保存
 */
func (service *personalInformationDriveFolderService) SyncPersonalInformationDriveFolder(req types.SyncPersonalInformationDriveFolderRequest) results.Result {
	if service.googleDriveService == nil {
		return results.InternalServerError(
			"SYNC_PERSONAL_INFORMATION_DRIVE_FOLDER_GOOGLE_DRIVE_SERVICE_NOT_CONFIGURED",
			"Google Drive連携が設定されていません",
			nil,
		)
	}

	user, userResult := service.findTargetUser(req.TargetUserID)
	if userResult.Error {
		return userResult
	}

	externalStorageLink, externalStorageLinkResult := service.findPersonalInformationDriveRootLink()
	if externalStorageLinkResult.Error {
		return externalStorageLinkResult
	}

	parentFolderID, parseResult := service.parseParentFolderID(externalStorageLink.URL)
	if parseResult.Error {
		return parseResult
	}

	currentFolder, currentFolderResult := service.findCurrentPersonalInformationDriveFolder(req.TargetUserID)
	currentFolderExists := !currentFolderResult.Error

	folderName := storage.BuildGoogleDriveUserFolderName("user", user.ID, user.Name)
	now := time.Now()
	var savedFolder models.PersonalInformationDriveFolder

	if currentFolderExists {
		folderMetadata, metadataErr := service.googleDriveService.GetFolderMetadata(context.Background(), currentFolder.DriveFolderID)
		if metadataErr != nil {
			return results.InternalServerError(
				"SYNC_PERSONAL_INFORMATION_DRIVE_FOLDER_GET_FOLDER_METADATA_FAILED",
				"個人情報Driveフォルダの確認に失敗しました",
				metadataErr.Error(),
			)
		}

		updatedFolder, buildUpdateResult := service.personalInformationDriveFolderBuilder.BuildUpdatePersonalInformationDriveFolderModel(
			currentFolder,
			externalStorageLink.ID,
			folderMetadata.FolderName,
			folderMetadata.DriveFolderID,
			folderMetadata.FolderURL,
			now,
		)
		if buildUpdateResult.Error {
			return buildUpdateResult
		}

		folder, saveResult := service.personalInformationDriveFolderRepository.SavePersonalInformationDriveFolder(updatedFolder)
		if saveResult.Error {
			return saveResult
		}

		savedFolder = folder
	} else {
		folderMetadata, createErr := service.googleDriveService.CreateFolder(context.Background(), parentFolderID, folderName)
		if createErr != nil {
			return results.InternalServerError(
				"SYNC_PERSONAL_INFORMATION_DRIVE_FOLDER_CREATE_FOLDER_FAILED",
				"個人情報Driveフォルダの作成に失敗しました",
				createErr.Error(),
			)
		}

		createdFolder, buildCreateResult := service.personalInformationDriveFolderBuilder.BuildCreatePersonalInformationDriveFolderModel(
			user.ID,
			externalStorageLink.ID,
			folderMetadata.FolderName,
			folderMetadata.DriveFolderID,
			folderMetadata.FolderURL,
			now,
		)
		if buildCreateResult.Error {
			return buildCreateResult
		}

		folder, createResult := service.personalInformationDriveFolderRepository.CreatePersonalInformationDriveFolder(createdFolder)
		if createResult.Error {
			return createResult
		}

		savedFolder = folder
	}

	permissions, permissionsResult := service.buildPersonalInformationDriveFolderPermissions(user)
	if permissionsResult.Error {
		return permissionsResult
	}

	if err := service.googleDriveService.SyncPermissions(context.Background(), savedFolder.DriveFolderID, permissions, true); err != nil {
		return results.InternalServerError(
			"SYNC_PERSONAL_INFORMATION_DRIVE_FOLDER_PERMISSIONS_FAILED",
			"個人情報Driveフォルダの権限同期に失敗しました",
			err.Error(),
		)
	}

	savedFolder.SyncedAt = &now
	savedFolder, saveResult := service.personalInformationDriveFolderRepository.SavePersonalInformationDriveFolder(savedFolder)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.SyncPersonalInformationDriveFolderResponse{
			PersonalInformationDriveFolder: toPersonalInformationDriveFolderResponse(savedFolder, user),
		},
		"SYNC_PERSONAL_INFORMATION_DRIVE_FOLDER_SUCCESS",
		"個人情報Driveフォルダを最新状態に更新しました",
		nil,
	)
}

/*
 * 個人情報Driveフォルダ表示
 *
 * 管理者用。
 * 対象ユーザーの個人情報DriveフォルダURLを返す。
 */
func (service *personalInformationDriveFolderService) ViewPersonalInformationDriveFolder(req types.ViewPersonalInformationDriveFolderRequest) results.Result {
	user, userResult := service.findTargetUser(req.TargetUserID)
	if userResult.Error {
		return userResult
	}

	folder, folderResult := service.findCurrentPersonalInformationDriveFolder(req.TargetUserID)
	if folderResult.Error {
		return folderResult
	}

	return results.OK(
		types.ViewPersonalInformationDriveFolderResponse{
			PersonalInformationDriveFolder: toPersonalInformationDriveFolderResponse(folder, user),
		},
		"VIEW_PERSONAL_INFORMATION_DRIVE_FOLDER_SUCCESS",
		"個人情報Driveフォルダを取得しました",
		nil,
	)
}

/*
 * 対象ユーザー取得
 */
func (service *personalInformationDriveFolderService) findTargetUser(targetUserID uint) (models.User, results.Result) {
	query, buildResult := service.personalInformationDriveFolderBuilder.BuildFindActiveUserByIDQuery(targetUserID)
	if buildResult.Error {
		return models.User{}, buildResult
	}

	user, findResult := service.personalInformationDriveFolderRepository.FindUser(query)
	if findResult.Error {
		return models.User{}, findResult
	}

	return user, results.OK(nil, "FIND_TARGET_USER_FOR_PERSONAL_INFORMATION_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * 現在の個人情報Driveフォルダ取得
 */
func (service *personalInformationDriveFolderService) findCurrentPersonalInformationDriveFolder(targetUserID uint) (models.PersonalInformationDriveFolder, results.Result) {
	query, buildResult := service.personalInformationDriveFolderBuilder.BuildFindActivePersonalInformationDriveFolderByUserIDQuery(targetUserID)
	if buildResult.Error {
		return models.PersonalInformationDriveFolder{}, buildResult
	}

	folder, findResult := service.personalInformationDriveFolderRepository.FindPersonalInformationDriveFolder(query)
	if findResult.Error {
		return models.PersonalInformationDriveFolder{}, findResult
	}

	return folder, results.OK(nil, "FIND_CURRENT_PERSONAL_INFORMATION_DRIVE_FOLDER_SUCCESS", "", nil)
}

/*
 * 個人情報Drive親フォルダ設定取得
 */
func (service *personalInformationDriveFolderService) findPersonalInformationDriveRootLink() (models.ExternalStorageLink, results.Result) {
	query, buildResult := service.personalInformationDriveFolderBuilder.BuildFindPersonalInformationDriveRootLinkQuery()
	if buildResult.Error {
		return models.ExternalStorageLink{}, buildResult
	}

	externalStorageLink, findResult := service.personalInformationDriveFolderRepository.FindExternalStorageLink(query)
	if findResult.Error {
		return models.ExternalStorageLink{}, findResult
	}

	return externalStorageLink, results.OK(nil, "FIND_PERSONAL_INFORMATION_DRIVE_ROOT_LINK_SUCCESS", "", nil)
}

/*
 * 親フォルダID解析
 */
func (service *personalInformationDriveFolderService) parseParentFolderID(folderURLOrID string) (string, results.Result) {
	folderURLOrID = strings.TrimSpace(folderURLOrID)
	if folderURLOrID == "" {
		return "", results.BadRequest(
			"PERSONAL_INFORMATION_DRIVE_ROOT_LINK_URL_EMPTY",
			"個人情報Drive親フォルダのURLが設定されていません",
			nil,
		)
	}

	folderID, err := service.googleDriveService.ParseFolderID(folderURLOrID)
	if err != nil {
		return "", results.BadRequest(
			"PERSONAL_INFORMATION_DRIVE_ROOT_LINK_URL_INVALID",
			"個人情報Drive親フォルダのURL形式が正しくありません",
			err.Error(),
		)
	}

	return folderID, results.OK(nil, "PERSONAL_INFORMATION_DRIVE_ROOT_FOLDER_ID_PARSE_SUCCESS", "", nil)
}

/*
 * 個人情報Driveフォルダ権限作成
 *
 * ・対象ユーザー本人：reader
 * ・管理者全員：writer
 */
func (service *personalInformationDriveFolderService) buildPersonalInformationDriveFolderPermissions(targetUser models.User) ([]storage.GoogleDrivePermissionSetting, results.Result) {
	adminQuery, buildAdminQueryResult := service.personalInformationDriveFolderBuilder.BuildFindActiveAdminUsersQuery()
	if buildAdminQueryResult.Error {
		return nil, buildAdminQueryResult
	}

	adminUsers, findAdminsResult := service.personalInformationDriveFolderRepository.FindUsers(adminQuery)
	if findAdminsResult.Error {
		return nil, findAdminsResult
	}

	permissions := make([]storage.GoogleDrivePermissionSetting, 0, len(adminUsers)+1)

	if strings.TrimSpace(targetUser.Email) == "" {
		return nil, results.BadRequest(
			"BUILD_PERSONAL_INFORMATION_DRIVE_FOLDER_PERMISSIONS_EMPTY_TARGET_USER_EMAIL",
			"対象ユーザーのメールアドレスが空です",
			map[string]any{
				"targetUserId": targetUser.ID,
			},
		)
	}

	permissions = append(permissions, storage.GoogleDrivePermissionSetting{
		EmailAddress: targetUser.Email,
		Role:         "reader",
	})

	for _, adminUser := range adminUsers {
		if strings.TrimSpace(adminUser.Email) == "" {
			continue
		}

		permissions = append(permissions, storage.GoogleDrivePermissionSetting{
			EmailAddress: adminUser.Email,
			Role:         "writer",
		})
	}

	return permissions, results.OK(nil, "BUILD_PERSONAL_INFORMATION_DRIVE_FOLDER_PERMISSIONS_SUCCESS", "", nil)
}
