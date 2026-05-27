package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 従業員用 共有資料DriveフォルダService interface
 *
 * 従業員側では閲覧のみ。
 * Driveフォルダの作成・更新・削除・権限同期は管理者側で行う。
 */
type SharedDocumentDriveFolderService interface {
	SearchSharedDocumentDriveFolders(req types.SearchSharedDocumentDriveFoldersRequest) results.Result
	DetailSharedDocumentDriveFolder(req types.SharedDocumentDriveFolderDetailRequest) results.Result
}

/*
 * 従業員用 共有資料DriveフォルダService
 */
type sharedDocumentDriveFolderService struct {
	sharedDocumentDriveFolderBuilder    builders.SharedDocumentDriveFolderBuilder
	sharedDocumentDriveFolderRepository repositories.SharedDocumentDriveFolderRepository
}

/*
 * SharedDocumentDriveFolderService生成
 */
func NewSharedDocumentDriveFolderService(
	sharedDocumentDriveFolderBuilder builders.SharedDocumentDriveFolderBuilder,
	sharedDocumentDriveFolderRepository repositories.SharedDocumentDriveFolderRepository,
) SharedDocumentDriveFolderService {
	return &sharedDocumentDriveFolderService{
		sharedDocumentDriveFolderBuilder:    sharedDocumentDriveFolderBuilder,
		sharedDocumentDriveFolderRepository: sharedDocumentDriveFolderRepository,
	}
}

/*
 * models.SharedDocumentDriveFolderをResponseへ変換する
 *
 * 従業員側ではDrive内部IDは返さない。
 * ユーザー画面で必要なのは表示名・説明・開くためのURLのみ。
 */
func toUserSharedDocumentDriveFolderResponse(folder models.SharedDocumentDriveFolder) types.SharedDocumentDriveFolderResponse {
	return types.SharedDocumentDriveFolderResponse{
		ID: folder.ID,

		FolderName:  folder.FolderName,
		Description: folder.Description,
		FolderURL:   folder.FolderURL,
		SyncedAt:    folder.SyncedAt,

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
		"SEARCH_USER_SHARED_DOCUMENT_DRIVE_FOLDERS_INVALID_OFFSET",
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
		"SEARCH_USER_SHARED_DOCUMENT_DRIVE_FOLDERS_SUCCESS",
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
			SharedDocumentDriveFolder: toUserSharedDocumentDriveFolderResponse(folder),
		},
		"DETAIL_USER_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS",
		"共有資料Driveフォルダを取得しました",
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

	return folder, results.OK(
		nil,
		"FIND_CURRENT_USER_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS",
		"",
		nil,
	)
}
