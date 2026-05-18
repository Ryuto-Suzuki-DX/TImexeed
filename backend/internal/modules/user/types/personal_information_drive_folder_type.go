package types

import "time"

/*
 * 従業員用 個人情報DriveフォルダResponse
 *
 * ユーザー側は検索不要。
 * JWTから本人userIdを取得し、自分のフォルダだけ返す。
 */
type PersonalInformationDriveFolderResponse struct {
	ID uint `json:"id"`

	UserID uint `json:"userId"`

	FolderName    string     `json:"folderName"`
	DriveFolderID string     `json:"driveFolderId"`
	FolderURL     string     `json:"folderUrl"`
	SyncedAt      *time.Time `json:"syncedAt"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

/*
 * 従業員用 自分の個人情報Driveフォルダ取得Response
 */
type GetMyPersonalInformationDriveFolderResponse struct {
	PersonalInformationDriveFolder PersonalInformationDriveFolderResponse `json:"personalInformationDriveFolder"`
}
