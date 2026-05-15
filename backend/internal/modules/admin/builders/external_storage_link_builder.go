package builders

import (
	"strings"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用外部ストレージリンクBuilder interface
 */
type ExternalStorageLinkBuilder interface {
	BuildSearchExternalStorageLinksQuery(req types.SearchExternalStorageLinksRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindExternalStorageLinkByIDQuery(externalStorageLinkID uint) (*gorm.DB, results.Result)
	BuildUpdateExternalStorageLinkModel(currentExternalStorageLink models.ExternalStorageLink, req types.UpdateExternalStorageLinkRequest) (models.ExternalStorageLink, results.Result)
}

/*
 * 管理者用外部ストレージリンクBuilder
 */
type externalStorageLinkBuilder struct {
	db *gorm.DB
}

/*
 * ExternalStorageLinkBuilder生成
 */
func NewExternalStorageLinkBuilder(db *gorm.DB) ExternalStorageLinkBuilder {
	return &externalStorageLinkBuilder{
		db: db,
	}
}

/*
 * 検索用Query作成
 */
func (builder *externalStorageLinkBuilder) BuildSearchExternalStorageLinksQuery(req types.SearchExternalStorageLinksRequest) (*gorm.DB, *gorm.DB, results.Result) {
	searchQuery := builder.db.Model(&models.ExternalStorageLink{})
	countQuery := builder.db.Model(&models.ExternalStorageLink{})

	if !req.IncludeDeleted {
		searchQuery = searchQuery.Where("is_deleted = ?", false)
		countQuery = countQuery.Where("is_deleted = ?", false)
	}

	if strings.TrimSpace(req.LinkType) != "" {
		linkType := strings.TrimSpace(req.LinkType)
		searchQuery = searchQuery.Where("link_type = ?", linkType)
		countQuery = countQuery.Where("link_type = ?", linkType)
	}

	if strings.TrimSpace(req.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(req.Keyword) + "%"
		searchQuery = searchQuery.Where(
			"link_type ILIKE ? OR link_name ILIKE ? OR url ILIKE ? OR description ILIKE ? OR memo ILIKE ?",
			keyword,
			keyword,
			keyword,
			keyword,
			keyword,
		)
		countQuery = countQuery.Where(
			"link_type ILIKE ? OR link_name ILIKE ? OR url ILIKE ? OR description ILIKE ? OR memo ILIKE ?",
			keyword,
			keyword,
			keyword,
			keyword,
			keyword,
		)
	}

	searchQuery = searchQuery.Order("id asc").Offset(req.Offset).Limit(req.Limit)

	return searchQuery, countQuery, results.OK(
		nil,
		"BUILD_SEARCH_EXTERNAL_STORAGE_LINKS_QUERY_SUCCESS",
		"外部ストレージリンク検索クエリを作成しました",
		nil,
	)
}

/*
 * ID検索用Query作成
 */
func (builder *externalStorageLinkBuilder) BuildFindExternalStorageLinkByIDQuery(externalStorageLinkID uint) (*gorm.DB, results.Result) {
	if externalStorageLinkID == 0 {
		return nil, results.BadRequest(
			"EXTERNAL_STORAGE_LINK_ID_REQUIRED",
			"外部ストレージリンクIDが指定されていません",
			nil,
		)
	}

	query := builder.db.Model(&models.ExternalStorageLink{}).
		Where("id = ?", externalStorageLinkID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_EXTERNAL_STORAGE_LINK_BY_ID_QUERY_SUCCESS",
		"外部ストレージリンク取得クエリを作成しました",
		nil,
	)
}

/*
 * 更新用Model作成
 *
 * 固定されたリンク種別/リンク名は更新対象にしない。
 * 管理者が変更できるのは URL / 説明 / 管理メモ のみ。
 */
func (builder *externalStorageLinkBuilder) BuildUpdateExternalStorageLinkModel(
	currentExternalStorageLink models.ExternalStorageLink,
	req types.UpdateExternalStorageLinkRequest,
) (models.ExternalStorageLink, results.Result) {
	url := strings.TrimSpace(req.URL)

	currentExternalStorageLink.URL = url
	currentExternalStorageLink.Description = req.Description
	currentExternalStorageLink.Memo = req.Memo

	return currentExternalStorageLink, results.OK(
		nil,
		"BUILD_UPDATE_EXTERNAL_STORAGE_LINK_MODEL_SUCCESS",
		"外部ストレージリンク更新用モデルを作成しました",
		nil,
	)
}
