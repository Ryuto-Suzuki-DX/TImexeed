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
	UpdateExternalStorageLink(req types.UpdateExternalStorageLinkRequest) results.Result
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
 * ・管理者が任意に外部ストレージリンクを作成/削除する運用にはしない
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
 * 更新
 *
 * 固定されたリンク種別/リンク名は更新しない。
 * 管理者が変更できるのは URL / 説明 / 管理メモ のみ。
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
