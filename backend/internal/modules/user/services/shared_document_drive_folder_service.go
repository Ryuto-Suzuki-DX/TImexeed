package services

import (
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 従業員用 共有資料DriveフォルダService interface
 */
type SharedDocumentDriveFolderService interface {
	SearchSharedDocumentDriveFolders(userID uint, req types.SearchSharedDocumentDriveFoldersRequest) results.Result
	DetailSharedDocumentDriveFolder(userID uint, req types.SharedDocumentDriveFolderDetailRequest) results.Result
}

/*
 * 従業員用 共有資料DriveフォルダService
 *
 * 役割：
 * ・Controllerから受け取ったログインユーザーIDをもとに処理を進める
 * ・本人に共有されている資料だけ返す
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
) *sharedDocumentDriveFolderService {
	return &sharedDocumentDriveFolderService{
		sharedDocumentDriveFolderBuilder:    sharedDocumentDriveFolderBuilder,
		sharedDocumentDriveFolderRepository: sharedDocumentDriveFolderRepository,
	}
}

/*
 * RowをResponseへ変換する
 */
func toSharedDocumentDriveFolderResponse(
	row types.SharedDocumentDriveFolderRow,
) types.SharedDocumentDriveFolderResponse {
	return types.SharedDocumentDriveFolderResponse{
		ID: row.ID,

		FolderName:    row.FolderName,
		Description:   row.Description,
		DriveFolderID: row.DriveFolderID,
		FolderURL:     row.FolderURL,
		SyncedAt:      row.SyncedAt,

		SharedAt:  row.SharedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

/*
 * 共有資料Driveフォルダ検索
 *
 * ユーザー側。
 * JWTから取得した本人userIdに共有されている資料だけ返す。
 */
func (service *sharedDocumentDriveFolderService) SearchSharedDocumentDriveFolders(
	userID uint,
	req types.SearchSharedDocumentDriveFoldersRequest,
) results.Result {
	normalizedCondition, normalizeResult := utils.NormalizePageSearchCondition(
		utils.PageSearchCondition{
			Keyword: req.Keyword,
			Offset:  req.Offset,
			Limit:   req.Limit,
		},
		"SEARCH_MY_SHARED_DOCUMENT_DRIVE_FOLDERS_INVALID_OFFSET",
		"検索開始位置が正しくありません",
	)
	if normalizeResult.Error {
		return normalizeResult
	}

	req.Keyword = normalizedCondition.Keyword
	req.Offset = normalizedCondition.Offset
	req.Limit = normalizedCondition.Limit

	searchQuery, countQuery, buildResult := service.sharedDocumentDriveFolderBuilder.BuildSearchSharedDocumentDriveFoldersQuery(userID, req)
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
		"SEARCH_MY_SHARED_DOCUMENT_DRIVE_FOLDERS_SUCCESS",
		"共有資料Driveフォルダ一覧を取得しました",
		nil,
	)
}

/*
 * 共有資料Driveフォルダ詳細
 *
 * 本人に共有されている資料だけ取得可能。
 */
func (service *sharedDocumentDriveFolderService) DetailSharedDocumentDriveFolder(
	userID uint,
	req types.SharedDocumentDriveFolderDetailRequest,
) results.Result {
	query, buildResult := service.sharedDocumentDriveFolderBuilder.BuildFindSharedDocumentDriveFolderDetailQuery(
		userID,
		req.TargetSharedDocumentDriveFolderID,
	)
	if buildResult.Error {
		return buildResult
	}

	row, findResult := service.sharedDocumentDriveFolderRepository.FindSharedDocumentDriveFolderRow(query)
	if findResult.Error {
		return findResult
	}

	return results.OK(
		types.SharedDocumentDriveFolderDetailResponse{
			SharedDocumentDriveFolder: toSharedDocumentDriveFolderResponse(row),
		},
		"DETAIL_MY_SHARED_DOCUMENT_DRIVE_FOLDER_SUCCESS",
		"共有資料Driveフォルダを取得しました",
		nil,
	)
}
