package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 管理者用外部ストレージリンクService interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type ExternalStorageLinkService interface {
	SearchExternalStorageLinks(req types.SearchExternalStorageLinksRequest) results.Result
	GetExternalStorageLinkDetail(req types.ExternalStorageLinkDetailRequest) results.Result
	CreateExternalStorageLink(req types.CreateExternalStorageLinkRequest) results.Result
	UpdateExternalStorageLink(req types.UpdateExternalStorageLinkRequest) results.Result
	DeleteExternalStorageLink(req types.DeleteExternalStorageLinkRequest) results.Result
}

/*
 * 管理者用外部ストレージリンクService
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや更新用Modelを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB処理を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type externalStorageLinkService struct {
	externalStorageLinkBuilder    builders.ExternalStorageLinkBuilder
	externalStorageLinkRepository repositories.ExternalStorageLinkRepository
}

/*
 * ExternalStorageLinkService生成
 */
func NewExternalStorageLinkService(
	externalStorageLinkBuilder builders.ExternalStorageLinkBuilder,
	externalStorageLinkRepository repositories.ExternalStorageLinkRepository,
) ExternalStorageLinkService {
	return &externalStorageLinkService{
		externalStorageLinkBuilder:    externalStorageLinkBuilder,
		externalStorageLinkRepository: externalStorageLinkRepository,
	}
}

/*
 * models.ExternalStorageLinkをフロント返却用ExternalStorageLinkResponseへ変換する
 *
 * 日付はtime.Time / *time.Timeのまま返す。
 * 表示形式の整形はフロント側で行う。
 */
func toExternalStorageLinkResponse(externalStorageLink models.ExternalStorageLink) types.ExternalStorageLinkResponse {
	return types.ExternalStorageLinkResponse{
		ID:          externalStorageLink.ID,
		LinkType:    externalStorageLink.LinkType,
		LinkName:    externalStorageLink.LinkName,
		URL:         externalStorageLink.URL,
		Description: externalStorageLink.Description,
		Memo:        externalStorageLink.Memo,
		IsDeleted:   externalStorageLink.IsDeleted,
		CreatedAt:   externalStorageLink.CreatedAt,
		UpdatedAt:   externalStorageLink.UpdatedAt,
		DeletedAt:   externalStorageLink.DeletedAt,
	}
}

/*
 * 検索
 *
 * ページング方針：
 * ・初回は offset=0, limit=50
 * ・さらに表示するときは、フロントで現在表示済みの件数を offset として送る
 * ・limit が未指定、0以下の場合は 50件にする
 * ・limit が 50件を超える場合も 50件に丸める
 */
func (service *externalStorageLinkService) SearchExternalStorageLinks(req types.SearchExternalStorageLinksRequest) results.Result {
	normalizedCondition, normalizeResult := utils.NormalizePageSearchCondition(
		utils.PageSearchCondition{
			Keyword: req.Keyword,
			Offset:  req.Offset,
			Limit:   req.Limit,
		},
		"SEARCH_EXTERNAL_STORAGE_LINKS_INVALID_OFFSET",
		"検索開始位置が正しくありません",
	)
	if normalizeResult.Error {
		return normalizeResult
	}

	req.Keyword = normalizedCondition.Keyword
	req.Offset = normalizedCondition.Offset
	req.Limit = normalizedCondition.Limit

	searchQuery, countQuery, buildResult := service.externalStorageLinkBuilder.BuildSearchExternalStorageLinksQuery(req)
	if buildResult.Error {
		return buildResult
	}

	externalStorageLinks, findResult := service.externalStorageLinkRepository.FindExternalStorageLinks(searchQuery)
	if findResult.Error {
		return findResult
	}

	total, countResult := service.externalStorageLinkRepository.CountExternalStorageLinks(countQuery)
	if countResult.Error {
		return countResult
	}

	externalStorageLinkResponses := make([]types.ExternalStorageLinkResponse, 0, len(externalStorageLinks))
	for _, externalStorageLink := range externalStorageLinks {
		externalStorageLinkResponses = append(externalStorageLinkResponses, toExternalStorageLinkResponse(externalStorageLink))
	}

	hasMore := utils.HasMore(total, req.Offset, len(externalStorageLinks))

	return results.OK(
		types.SearchExternalStorageLinksResponse{
			ExternalStorageLinks: externalStorageLinkResponses,
			Total:                total,
			Offset:               req.Offset,
			Limit:                req.Limit,
			HasMore:              hasMore,
		},
		"SEARCH_EXTERNAL_STORAGE_LINKS_SUCCESS",
		"外部ストレージリンク一覧を取得しました",
		nil,
	)
}

/*
 * 詳細
 */
func (service *externalStorageLinkService) GetExternalStorageLinkDetail(req types.ExternalStorageLinkDetailRequest) results.Result {
	query, buildResult := service.externalStorageLinkBuilder.BuildFindExternalStorageLinkByIDQuery(req.ExternalStorageLinkID)
	if buildResult.Error {
		return buildResult
	}

	externalStorageLink, findResult := service.externalStorageLinkRepository.FindExternalStorageLink(query)
	if findResult.Error {
		return findResult
	}

	return results.OK(
		types.ExternalStorageLinkDetailResponse{
			ExternalStorageLink: toExternalStorageLinkResponse(externalStorageLink),
		},
		"GET_EXTERNAL_STORAGE_LINK_DETAIL_SUCCESS",
		"外部ストレージリンク詳細を取得しました",
		nil,
	)
}

/*
 * 新規作成
 */
func (service *externalStorageLinkService) CreateExternalStorageLink(req types.CreateExternalStorageLinkRequest) results.Result {
	linkTypeCountQuery, buildLinkTypeCountResult := service.externalStorageLinkBuilder.BuildCountActiveExternalStorageLinkByLinkTypeQuery(req.LinkType)
	if buildLinkTypeCountResult.Error {
		return buildLinkTypeCountResult
	}

	linkTypeCount, linkTypeCountResult := service.externalStorageLinkRepository.CountExternalStorageLinks(linkTypeCountQuery)
	if linkTypeCountResult.Error {
		return linkTypeCountResult
	}

	if linkTypeCount > 0 {
		return results.Conflict(
			"CREATE_EXTERNAL_STORAGE_LINK_TYPE_ALREADY_EXISTS",
			"このリンク種別は既に使用されています",
			map[string]any{
				"linkType": req.LinkType,
			},
		)
	}

	externalStorageLink, buildExternalStorageLinkResult := service.externalStorageLinkBuilder.BuildCreateExternalStorageLinkModel(req)
	if buildExternalStorageLinkResult.Error {
		return buildExternalStorageLinkResult
	}

	createdExternalStorageLink, createResult := service.externalStorageLinkRepository.CreateExternalStorageLink(externalStorageLink)
	if createResult.Error {
		return createResult
	}

	return results.Created(
		types.CreateExternalStorageLinkResponse{
			ExternalStorageLink: toExternalStorageLinkResponse(createdExternalStorageLink),
		},
		"CREATE_EXTERNAL_STORAGE_LINK_SUCCESS",
		"外部ストレージリンクを作成しました",
		nil,
	)
}

/*
 * 更新
 */
func (service *externalStorageLinkService) UpdateExternalStorageLink(req types.UpdateExternalStorageLinkRequest) results.Result {
	findQuery, buildFindResult := service.externalStorageLinkBuilder.BuildFindExternalStorageLinkByIDQuery(req.ExternalStorageLinkID)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentExternalStorageLink, findResult := service.externalStorageLinkRepository.FindExternalStorageLink(findQuery)
	if findResult.Error {
		return findResult
	}

	linkTypeCountQuery, buildLinkTypeCountResult := service.externalStorageLinkBuilder.BuildCountActiveExternalStorageLinkByLinkTypeExceptIDQuery(req.LinkType, req.ExternalStorageLinkID)
	if buildLinkTypeCountResult.Error {
		return buildLinkTypeCountResult
	}

	linkTypeCount, linkTypeCountResult := service.externalStorageLinkRepository.CountExternalStorageLinks(linkTypeCountQuery)
	if linkTypeCountResult.Error {
		return linkTypeCountResult
	}

	if linkTypeCount > 0 {
		return results.Conflict(
			"UPDATE_EXTERNAL_STORAGE_LINK_TYPE_ALREADY_EXISTS",
			"このリンク種別は既に使用されています",
			map[string]any{
				"linkType":              req.LinkType,
				"externalStorageLinkId": req.ExternalStorageLinkID,
			},
		)
	}

	updatedExternalStorageLink, buildUpdateResult := service.externalStorageLinkBuilder.BuildUpdateExternalStorageLinkModel(
		currentExternalStorageLink,
		req,
	)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	savedExternalStorageLink, saveResult := service.externalStorageLinkRepository.SaveExternalStorageLink(updatedExternalStorageLink)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.UpdateExternalStorageLinkResponse{
			ExternalStorageLink: toExternalStorageLinkResponse(savedExternalStorageLink),
		},
		"UPDATE_EXTERNAL_STORAGE_LINK_SUCCESS",
		"外部ストレージリンクを更新しました",
		nil,
	)
}

/*
 * 論理削除
 */
func (service *externalStorageLinkService) DeleteExternalStorageLink(req types.DeleteExternalStorageLinkRequest) results.Result {
	findQuery, buildFindResult := service.externalStorageLinkBuilder.BuildFindExternalStorageLinkByIDQuery(req.ExternalStorageLinkID)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentExternalStorageLink, findResult := service.externalStorageLinkRepository.FindExternalStorageLink(findQuery)
	if findResult.Error {
		return findResult
	}

	deletedExternalStorageLink, buildDeleteResult := service.externalStorageLinkBuilder.BuildDeleteExternalStorageLinkModel(currentExternalStorageLink)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	_, saveResult := service.externalStorageLinkRepository.SaveExternalStorageLink(deletedExternalStorageLink)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteExternalStorageLinkResponse{
			ExternalStorageLinkID: req.ExternalStorageLinkID,
		},
		"DELETE_EXTERNAL_STORAGE_LINK_SUCCESS",
		"外部ストレージリンクを削除しました",
		nil,
	)
}
