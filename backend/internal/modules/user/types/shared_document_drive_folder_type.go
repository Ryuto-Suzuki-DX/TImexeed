package types

import "time"

/*
 * 従業員用 共有資料Driveフォルダ検索Request
 *
 * ユーザー側は本人に共有されている資料だけ検索する。
 * userId はJWTから取得するため、Requestでは受け取らない。
 */
type SearchSharedDocumentDriveFoldersRequest struct {
	Keyword string `json:"keyword"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

/*
 * 従業員用 共有資料Driveフォルダ詳細Request
 *
 * targetSharedDocumentDriveFolderId:
 * ・URLには載せない
 * ・request bodyで受け取る
 *
 * 注意：
 * ・本人に共有されていない資料は取得不可
 */
type SharedDocumentDriveFolderDetailRequest struct {
	TargetSharedDocumentDriveFolderID uint `json:"targetSharedDocumentDriveFolderId"`
}

/*
 * 従業員用 共有資料DriveフォルダRow
 */
type SharedDocumentDriveFolderRow struct {
	ID uint `json:"id"`

	FolderName    string     `json:"folderName"`
	Description   *string    `json:"description"`
	DriveFolderID string     `json:"driveFolderId"`
	FolderURL     string     `json:"folderUrl"`
	SyncedAt      *time.Time `json:"syncedAt"`

	SharedAt  time.Time `json:"sharedAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

/*
 * 従業員用 共有資料DriveフォルダResponse
 */
type SharedDocumentDriveFolderResponse struct {
	ID uint `json:"id"`

	FolderName    string     `json:"folderName"`
	Description   *string    `json:"description"`
	DriveFolderID string     `json:"driveFolderId"`
	FolderURL     string     `json:"folderUrl"`
	SyncedAt      *time.Time `json:"syncedAt"`

	SharedAt  time.Time `json:"sharedAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

/*
 * 従業員用 共有資料Driveフォルダ検索Response
 */
type SearchSharedDocumentDriveFoldersResponse struct {
	SharedDocumentDriveFolders []SharedDocumentDriveFolderRow `json:"sharedDocumentDriveFolders"`
	Total                      int64                          `json:"total"`
	Offset                     int                            `json:"offset"`
	Limit                      int                            `json:"limit"`
	HasMore                    bool                           `json:"hasMore"`
}

/*
 * 従業員用 共有資料Driveフォルダ詳細Response
 */
type SharedDocumentDriveFolderDetailResponse struct {
	SharedDocumentDriveFolder SharedDocumentDriveFolderResponse `json:"sharedDocumentDriveFolder"`
}
