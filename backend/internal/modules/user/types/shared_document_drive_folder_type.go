package types

import "time"

/*
 * 従業員用 共有資料Driveフォルダ検索Request
 *
 * 従業員側では、全ユーザー向けに公開されている共有資料Driveフォルダを閲覧するだけ。
 * 個別ユーザーの共有設定やDrive権限同期は扱わない。
 */
type SearchSharedDocumentDriveFoldersRequest struct {
	Keyword string `json:"keyword"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

/*
 * 従業員用 共有資料Driveフォルダ詳細Request
 */
type SharedDocumentDriveFolderDetailRequest struct {
	TargetSharedDocumentDriveFolderID uint `json:"targetSharedDocumentDriveFolderId"`
}

/*
 * 従業員用 共有資料Driveフォルダ検索Row
 */
type SharedDocumentDriveFolderSearchRow struct {
	ID uint `json:"id"`

	FolderName  string     `json:"folderName"`
	Description *string    `json:"description"`
	FolderURL   string     `json:"folderUrl"`
	SyncedAt    *time.Time `json:"syncedAt"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

/*
 * 従業員用 共有資料DriveフォルダResponse
 */
type SharedDocumentDriveFolderResponse struct {
	ID uint `json:"id"`

	FolderName  string     `json:"folderName"`
	Description *string    `json:"description"`
	FolderURL   string     `json:"folderUrl"`
	SyncedAt    *time.Time `json:"syncedAt"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

/*
 * 従業員用 共有資料Driveフォルダ検索Response
 */
type SearchSharedDocumentDriveFoldersResponse struct {
	SharedDocumentDriveFolders []SharedDocumentDriveFolderSearchRow `json:"sharedDocumentDriveFolders"`
	Total                      int64                                `json:"total"`
	Offset                     int                                  `json:"offset"`
	Limit                      int                                  `json:"limit"`
	HasMore                    bool                                 `json:"hasMore"`
}

/*
 * 従業員用 共有資料Driveフォルダ詳細Response
 */
type SharedDocumentDriveFolderDetailResponse struct {
	SharedDocumentDriveFolder SharedDocumentDriveFolderResponse `json:"sharedDocumentDriveFolder"`
}
