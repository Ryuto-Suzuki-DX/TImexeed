package types

import "time"

/*
 * 外部ストレージリンク検索Request
 */
type SearchExternalStorageLinksRequest struct {
	Keyword        string `json:"keyword"`
	LinkType       string `json:"linkType"`
	IncludeDeleted bool   `json:"includeDeleted"`
	Offset         int    `json:"offset"`
	Limit          int    `json:"limit"`
}

/*
 * 外部ストレージリンク詳細Request
 */
type ExternalStorageLinkDetailRequest struct {
	ExternalStorageLinkID uint `json:"externalStorageLinkId" binding:"required"`
}

/*
 * 外部ストレージリンク作成Request
 */
type CreateExternalStorageLinkRequest struct {
	LinkType    string  `json:"linkType" binding:"required"`
	LinkName    string  `json:"linkName" binding:"required"`
	URL         string  `json:"url" binding:"required"`
	Description *string `json:"description"`
	Memo        *string `json:"memo"`
}

/*
 * 外部ストレージリンク更新Request
 */
type UpdateExternalStorageLinkRequest struct {
	ExternalStorageLinkID uint    `json:"externalStorageLinkId" binding:"required"`
	LinkType              string  `json:"linkType" binding:"required"`
	LinkName              string  `json:"linkName" binding:"required"`
	URL                   string  `json:"url" binding:"required"`
	Description           *string `json:"description"`
	Memo                  *string `json:"memo"`
}

/*
 * 外部ストレージリンク削除Request
 */
type DeleteExternalStorageLinkRequest struct {
	ExternalStorageLinkID uint `json:"externalStorageLinkId" binding:"required"`
}

/*
 * 外部ストレージリンクResponse
 */
type ExternalStorageLinkResponse struct {
	ID          uint       `json:"id"`
	LinkType    string     `json:"linkType"`
	LinkName    string     `json:"linkName"`
	URL         string     `json:"url"`
	Description *string    `json:"description"`
	Memo        *string    `json:"memo"`
	IsDeleted   bool       `json:"isDeleted"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
}

/*
 * 外部ストレージリンク検索Response
 */
type SearchExternalStorageLinksResponse struct {
	ExternalStorageLinks []ExternalStorageLinkResponse `json:"externalStorageLinks"`
	Total                int64                         `json:"total"`
	Offset               int                           `json:"offset"`
	Limit                int                           `json:"limit"`
	HasMore              bool                          `json:"hasMore"`
}

/*
 * 外部ストレージリンク詳細Response
 */
type ExternalStorageLinkDetailResponse struct {
	ExternalStorageLink ExternalStorageLinkResponse `json:"externalStorageLink"`
}

/*
 * 外部ストレージリンク作成Response
 */
type CreateExternalStorageLinkResponse struct {
	ExternalStorageLink ExternalStorageLinkResponse `json:"externalStorageLink"`
}

/*
 * 外部ストレージリンク更新Response
 */
type UpdateExternalStorageLinkResponse struct {
	ExternalStorageLink ExternalStorageLinkResponse `json:"externalStorageLink"`
}

/*
 * 外部ストレージリンク削除Response
 */
type DeleteExternalStorageLinkResponse struct {
	ExternalStorageLinkID uint `json:"externalStorageLinkId"`
}
