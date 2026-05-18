package types

import "time"

/*
 * 管理者用 個人情報Driveフォルダ検索Request
 *
 * 管理者はユーザーのフリーワードで検索し、
 * 対象ユーザーの個人情報Driveフォルダへ遷移する。
 */
type SearchPersonalInformationDriveFoldersRequest struct {
	Keyword string `json:"keyword"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

/*
 * 管理者用 個人情報Driveフォルダ同期Request
 *
 * targetUserId：
 * ・URLには載せない
 * ・request bodyで受け取る
 */
type SyncPersonalInformationDriveFolderRequest struct {
	TargetUserID uint `json:"targetUserId"`
}

/*
 * 管理者用 個人情報Driveフォルダ表示Request
 */
type ViewPersonalInformationDriveFolderRequest struct {
	TargetUserID uint `json:"targetUserId"`
}

/*
 * 管理者用 個人情報Driveフォルダ検索Row
 *
 * users を主軸に LEFT JOIN して取得する。
 * フォルダ未作成ユーザーも一覧に出すため、フォルダ系はポインタにする。
 */
type PersonalInformationDriveFolderSearchRow struct {
	UserID       uint   `json:"userId"`
	UserName     string `json:"userName"`
	UserEmail    string `json:"userEmail"`
	UserRole     string `json:"userRole"`
	DepartmentID *uint  `json:"departmentId"`

	PersonalInformationDriveFolderID *uint      `json:"personalInformationDriveFolderId"`
	ExternalStorageLinkID            *uint      `json:"externalStorageLinkId"`
	FolderName                       *string    `json:"folderName"`
	DriveFolderID                    *string    `json:"driveFolderId"`
	FolderURL                        *string    `json:"folderUrl"`
	SyncedAt                         *time.Time `json:"syncedAt"`
	FolderCreatedAt                  *time.Time `json:"folderCreatedAt"`
	FolderUpdatedAt                  *time.Time `json:"folderUpdatedAt"`
	FolderRegistered                 bool       `json:"folderRegistered"`
}

/*
 * 管理者用 個人情報DriveフォルダResponse
 */
type PersonalInformationDriveFolderResponse struct {
	ID uint `json:"id"`

	UserID    uint   `json:"userId"`
	UserName  string `json:"userName"`
	UserEmail string `json:"userEmail"`

	ExternalStorageLinkID uint       `json:"externalStorageLinkId"`
	FolderName            string     `json:"folderName"`
	DriveFolderID         string     `json:"driveFolderId"`
	FolderURL             string     `json:"folderUrl"`
	SyncedAt              *time.Time `json:"syncedAt"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

/*
 * 管理者用 個人情報Driveフォルダ検索Response
 */
type SearchPersonalInformationDriveFoldersResponse struct {
	PersonalInformationDriveFolders []PersonalInformationDriveFolderSearchRow `json:"personalInformationDriveFolders"`
	Total                           int64                                     `json:"total"`
	Offset                          int                                       `json:"offset"`
	Limit                           int                                       `json:"limit"`
	HasMore                         bool                                      `json:"hasMore"`
}

/*
 * 管理者用 個人情報Driveフォルダ同期Response
 */
type SyncPersonalInformationDriveFolderResponse struct {
	PersonalInformationDriveFolder PersonalInformationDriveFolderResponse `json:"personalInformationDriveFolder"`
}

/*
 * 管理者用 個人情報Driveフォルダ表示Response
 */
type ViewPersonalInformationDriveFolderResponse struct {
	PersonalInformationDriveFolder PersonalInformationDriveFolderResponse `json:"personalInformationDriveFolder"`
}
