package builders

import (
	"strings"
	"time"

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
	BuildCountActiveExternalStorageLinkByLinkTypeQuery(linkType string) (*gorm.DB, results.Result)
	BuildCountActiveExternalStorageLinkByLinkTypeExceptIDQuery(linkType string, externalStorageLinkID uint) (*gorm.DB, results.Result)
	BuildCreateExternalStorageLinkModel(req types.CreateExternalStorageLinkRequest) (models.ExternalStorageLink, results.Result)
	BuildUpdateExternalStorageLinkModel(currentExternalStorageLink models.ExternalStorageLink, req types.UpdateExternalStorageLinkRequest) (models.ExternalStorageLink, results.Result)
	BuildDeleteExternalStorageLinkModel(currentExternalStorageLink models.ExternalStorageLink) (models.ExternalStorageLink, results.Result)
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
 * 有効なリンク種別件数取得用Query作成
 */
func (builder *externalStorageLinkBuilder) BuildCountActiveExternalStorageLinkByLinkTypeQuery(linkType string) (*gorm.DB, results.Result) {
	trimmedLinkType := strings.TrimSpace(linkType)
	if trimmedLinkType == "" {
		return nil, results.BadRequest(
			"EXTERNAL_STORAGE_LINK_TYPE_REQUIRED",
			"リンク種別が指定されていません",
			nil,
		)
	}

	query := builder.db.Model(&models.ExternalStorageLink{}).
		Where("link_type = ?", trimmedLinkType).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_COUNT_ACTIVE_EXTERNAL_STORAGE_LINK_BY_LINK_TYPE_QUERY_SUCCESS",
		"外部ストレージリンク種別件数取得クエリを作成しました",
		nil,
	)
}

/*
 * 自身を除く有効なリンク種別件数取得用Query作成
 */
func (builder *externalStorageLinkBuilder) BuildCountActiveExternalStorageLinkByLinkTypeExceptIDQuery(linkType string, externalStorageLinkID uint) (*gorm.DB, results.Result) {
	trimmedLinkType := strings.TrimSpace(linkType)
	if trimmedLinkType == "" {
		return nil, results.BadRequest(
			"EXTERNAL_STORAGE_LINK_TYPE_REQUIRED",
			"リンク種別が指定されていません",
			nil,
		)
	}

	if externalStorageLinkID == 0 {
		return nil, results.BadRequest(
			"EXTERNAL_STORAGE_LINK_ID_REQUIRED",
			"外部ストレージリンクIDが指定されていません",
			nil,
		)
	}

	query := builder.db.Model(&models.ExternalStorageLink{}).
		Where("link_type = ?", trimmedLinkType).
		Where("id <> ?", externalStorageLinkID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_COUNT_ACTIVE_EXTERNAL_STORAGE_LINK_BY_LINK_TYPE_EXCEPT_ID_QUERY_SUCCESS",
		"外部ストレージリンク種別重複確認クエリを作成しました",
		nil,
	)
}

/*
 * 作成用Model作成
 */
func (builder *externalStorageLinkBuilder) BuildCreateExternalStorageLinkModel(req types.CreateExternalStorageLinkRequest) (models.ExternalStorageLink, results.Result) {
	linkType := strings.TrimSpace(req.LinkType)
	linkName := strings.TrimSpace(req.LinkName)
	url := strings.TrimSpace(req.URL)

	if linkType == "" {
		return models.ExternalStorageLink{}, results.BadRequest(
			"CREATE_EXTERNAL_STORAGE_LINK_TYPE_REQUIRED",
			"リンク種別を入力してください",
			nil,
		)
	}

	if linkName == "" {
		return models.ExternalStorageLink{}, results.BadRequest(
			"CREATE_EXTERNAL_STORAGE_LINK_NAME_REQUIRED",
			"表示名を入力してください",
			nil,
		)
	}

	if url == "" {
		return models.ExternalStorageLink{}, results.BadRequest(
			"CREATE_EXTERNAL_STORAGE_LINK_URL_REQUIRED",
			"URLを入力してください",
			nil,
		)
	}

	externalStorageLink := models.ExternalStorageLink{
		LinkType:    linkType,
		LinkName:    linkName,
		URL:         url,
		Description: req.Description,
		Memo:        req.Memo,
		IsDeleted:   false,
	}

	return externalStorageLink, results.OK(
		nil,
		"BUILD_CREATE_EXTERNAL_STORAGE_LINK_MODEL_SUCCESS",
		"外部ストレージリンク作成用モデルを作成しました",
		nil,
	)
}

/*
 * 更新用Model作成
 */
func (builder *externalStorageLinkBuilder) BuildUpdateExternalStorageLinkModel(currentExternalStorageLink models.ExternalStorageLink, req types.UpdateExternalStorageLinkRequest) (models.ExternalStorageLink, results.Result) {
	linkType := strings.TrimSpace(req.LinkType)
	linkName := strings.TrimSpace(req.LinkName)
	url := strings.TrimSpace(req.URL)

	if linkType == "" {
		return models.ExternalStorageLink{}, results.BadRequest(
			"UPDATE_EXTERNAL_STORAGE_LINK_TYPE_REQUIRED",
			"リンク種別を入力してください",
			nil,
		)
	}

	if linkName == "" {
		return models.ExternalStorageLink{}, results.BadRequest(
			"UPDATE_EXTERNAL_STORAGE_LINK_NAME_REQUIRED",
			"表示名を入力してください",
			nil,
		)
	}

	if url == "" {
		return models.ExternalStorageLink{}, results.BadRequest(
			"UPDATE_EXTERNAL_STORAGE_LINK_URL_REQUIRED",
			"URLを入力してください",
			nil,
		)
	}

	currentExternalStorageLink.LinkType = linkType
	currentExternalStorageLink.LinkName = linkName
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

/*
 * 論理削除用Model作成
 */
func (builder *externalStorageLinkBuilder) BuildDeleteExternalStorageLinkModel(currentExternalStorageLink models.ExternalStorageLink) (models.ExternalStorageLink, results.Result) {
	now := time.Now()
	currentExternalStorageLink.IsDeleted = true
	currentExternalStorageLink.DeletedAt = &now

	return currentExternalStorageLink, results.OK(
		nil,
		"BUILD_DELETE_EXTERNAL_STORAGE_LINK_MODEL_SUCCESS",
		"外部ストレージリンク削除用モデルを作成しました",
		nil,
	)
}
