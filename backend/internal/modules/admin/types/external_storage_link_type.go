package types

import "time"

/*
 * 外部ストレージリンク検索Request
 *
 * 固定された外部ストレージリンク一覧を取得する。
 *
 * 管理者が自由に追加/削除するものではないため、
 * create/delete/detail 用のRequestは持たない。
 */
type SearchExternalStorageLinksRequest struct {
	Keyword        string `json:"keyword"`
	LinkType       string `json:"linkType"`
	IncludeDeleted bool   `json:"includeDeleted"`
	Offset         int    `json:"offset"`
	Limit          int    `json:"limit"`
}

/*
 * 外部ストレージリンク更新Request
 *
 * 固定されたリンク種別/リンク名は更新対象にしない。
 * 管理者が変更できるのはURL、説明、管理メモのみ。
 */
type UpdateExternalStorageLinkRequest struct {
	ExternalStorageLinkID uint    `json:"externalStorageLinkId" binding:"required"`
	URL                   string  `json:"url"`
	Description           *string `json:"description"`
	Memo                  *string `json:"memo"`
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
 * 外部ストレージリンク更新Response
 */
type UpdateExternalStorageLinkResponse struct {
	ExternalStorageLink ExternalStorageLinkResponse `json:"externalStorageLink"`
}
