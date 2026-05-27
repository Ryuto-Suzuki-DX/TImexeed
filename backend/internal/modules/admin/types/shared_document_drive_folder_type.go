package types

import "time"

/*
 * 管理者用 共有資料Driveフォルダ検索Request
 */
type SearchSharedDocumentDriveFoldersRequest struct {
	Keyword string `json:"keyword"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

/*
 * 管理者用 共有資料Driveフォルダ詳細Request
 */
type SharedDocumentDriveFolderDetailRequest struct {
	TargetSharedDocumentDriveFolderID uint `json:"targetSharedDocumentDriveFolderId"`
}

/*
 * 管理者用 共有資料Driveフォルダ作成Request
 *
 * Google Driveの親フォルダURL/IDは画面から受け取らない。
 * external_storage_links の SHARED_DOCUMENT_DRIVE_ROOT から取得する。
 */
type CreateSharedDocumentDriveFolderRequest struct {
	FolderName  string  `json:"folderName"`
	Description *string `json:"description"`
}

/*
 * 管理者用 共有資料Driveフォルダ更新Request
 *
 * Drive上のフォルダID/URLは更新しない。
 * 表示名・説明のみ更新する。
 */
type UpdateSharedDocumentDriveFolderRequest struct {
	TargetSharedDocumentDriveFolderID uint    `json:"targetSharedDocumentDriveFolderId"`
	FolderName                        string  `json:"folderName"`
	Description                       *string `json:"description"`
}

/*
 * 管理者用 共有資料Driveフォルダ削除Request
 */
type DeleteSharedDocumentDriveFolderRequest struct {
	TargetSharedDocumentDriveFolderID uint `json:"targetSharedDocumentDriveFolderId"`
}

/*
 * 管理者用 共有資料Driveフォルダ同期Request
 *
 * targetSharedDocumentDriveFolderId = 0 の場合：
 * ・有効な共有資料Driveフォルダ全件を同期する
 *
 * targetSharedDocumentDriveFolderId > 0 の場合：
 * ・指定された共有資料Driveフォルダ1件だけ同期する
 */
type SyncSharedDocumentDriveFolderRequest struct {
	TargetSharedDocumentDriveFolderID uint `json:"targetSharedDocumentDriveFolderId"`
}

/*
 * 管理者用 共有資料Driveフォルダ検索Row
 */
type SharedDocumentDriveFolderSearchRow struct {
	ID uint `json:"id"`

	FolderName    string     `json:"folderName"`
	Description   *string    `json:"description"`
	DriveFolderID string     `json:"driveFolderId"`
	FolderURL     string     `json:"folderUrl"`
	SyncedAt      *time.Time `json:"syncedAt"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

/*
 * 管理者用 共有資料DriveフォルダResponse
 */
type SharedDocumentDriveFolderResponse struct {
	ID uint `json:"id"`

	FolderName    string     `json:"folderName"`
	Description   *string    `json:"description"`
	DriveFolderID string     `json:"driveFolderId"`
	FolderURL     string     `json:"folderUrl"`
	SyncedAt      *time.Time `json:"syncedAt"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

/*
 * 管理者用 共有資料Driveフォルダ検索Response
 */
type SearchSharedDocumentDriveFoldersResponse struct {
	SharedDocumentDriveFolders []SharedDocumentDriveFolderSearchRow `json:"sharedDocumentDriveFolders"`
	Total                      int64                                `json:"total"`
	Offset                     int                                  `json:"offset"`
	Limit                      int                                  `json:"limit"`
	HasMore                    bool                                 `json:"hasMore"`
}

/*
 * 管理者用 共有資料Driveフォルダ詳細Response
 */
type SharedDocumentDriveFolderDetailResponse struct {
	SharedDocumentDriveFolder SharedDocumentDriveFolderResponse `json:"sharedDocumentDriveFolder"`
}

/*
 * 管理者用 共有資料Driveフォルダ作成Response
 */
type CreateSharedDocumentDriveFolderResponse struct {
	SharedDocumentDriveFolder SharedDocumentDriveFolderResponse `json:"sharedDocumentDriveFolder"`
}

/*
 * 管理者用 共有資料Driveフォルダ更新Response
 */
type UpdateSharedDocumentDriveFolderResponse struct {
	SharedDocumentDriveFolder SharedDocumentDriveFolderResponse `json:"sharedDocumentDriveFolder"`
}

/*
 * 管理者用 共有資料Driveフォルダ削除Response
 */
type DeleteSharedDocumentDriveFolderResponse struct {
	SharedDocumentDriveFolderID uint `json:"sharedDocumentDriveFolderId"`
}

/*
 * 管理者用 共有資料Driveフォルダ同期Response
 */
type SyncSharedDocumentDriveFolderResponse struct {
	SharedDocumentDriveFolders []SharedDocumentDriveFolderResponse `json:"sharedDocumentDriveFolders"`
	SyncedFolderCount          int                                 `json:"syncedFolderCount"`
	TargetAdminCount           int                                 `json:"targetAdminCount"`
	TargetUserCount            int                                 `json:"targetUserCount"`
	SyncedAt                   time.Time                           `json:"syncedAt"`
}
