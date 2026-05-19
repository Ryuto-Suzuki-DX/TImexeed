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
 * driveFolderUrlOrId:
 * ・Google DriveフォルダURL
 * ・またはフォルダID
 */
type CreateSharedDocumentDriveFolderRequest struct {
	FolderName         string  `json:"folderName"`
	Description        *string `json:"description"`
	DriveFolderURLOrID string  `json:"driveFolderUrlOrId"`
}

/*
 * 管理者用 共有資料Driveフォルダ更新Request
 */
type UpdateSharedDocumentDriveFolderRequest struct {
	TargetSharedDocumentDriveFolderID uint    `json:"targetSharedDocumentDriveFolderId"`
	FolderName                        string  `json:"folderName"`
	Description                       *string `json:"description"`
	DriveFolderURLOrID                string  `json:"driveFolderUrlOrId"`
}

/*
 * 管理者用 共有資料Driveフォルダ削除Request
 */
type DeleteSharedDocumentDriveFolderRequest struct {
	TargetSharedDocumentDriveFolderID uint `json:"targetSharedDocumentDriveFolderId"`
}

/*
 * 管理者用 共有資料Driveフォルダ同期Request
 */
type SyncSharedDocumentDriveFolderRequest struct {
	TargetSharedDocumentDriveFolderID uint `json:"targetSharedDocumentDriveFolderId"`
}

/*
 * 管理者用 共有資料Driveフォルダ共有ユーザー更新Request
 *
 * 通常時：
 * ・targetUserIds を共有対象の最終状態として扱う
 *
 * shareAllUsers = true の場合：
 * ・targetUserIds は無視する
 * ・有効なUSER全員を共有対象にする
 *
 * 注意：
 * ・共有対象の削除もこのAPIで行う
 * ・targetUserIds = [] かつ shareAllUsers = false の場合、共有対象を全削除する
 */
type UpdateSharedDocumentDriveFolderUsersRequest struct {
	TargetSharedDocumentDriveFolderID uint   `json:"targetSharedDocumentDriveFolderId"`
	TargetUserIDs                     []uint `json:"targetUserIds"`
	ShareAllUsers                     bool   `json:"shareAllUsers"`
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

	SharedUserCount int64 `json:"sharedUserCount"`

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
 * 管理者用 共有資料Driveフォルダ共有ユーザーResponse
 */
type SharedDocumentDriveFolderUserResponse struct {
	ID uint `json:"id"`

	SharedDocumentDriveFolderID uint `json:"sharedDocumentDriveFolderId"`

	UserID    uint   `json:"userId"`
	UserName  string `json:"userName"`
	UserEmail string `json:"userEmail"`
	UserRole  string `json:"userRole"`

	SyncedAt *time.Time `json:"syncedAt"`

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
	SharedDocumentDriveFolder SharedDocumentDriveFolderResponse       `json:"sharedDocumentDriveFolder"`
	SharedUsers               []SharedDocumentDriveFolderUserResponse `json:"sharedUsers"`
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
	SharedDocumentDriveFolder SharedDocumentDriveFolderResponse       `json:"sharedDocumentDriveFolder"`
	SharedUsers               []SharedDocumentDriveFolderUserResponse `json:"sharedUsers"`
}

/*
 * 管理者用 共有資料Driveフォルダ共有ユーザー更新Response
 */
type UpdateSharedDocumentDriveFolderUsersResponse struct {
	SharedDocumentDriveFolderID uint                                    `json:"sharedDocumentDriveFolderId"`
	SharedUsers                 []SharedDocumentDriveFolderUserResponse `json:"sharedUsers"`
}
