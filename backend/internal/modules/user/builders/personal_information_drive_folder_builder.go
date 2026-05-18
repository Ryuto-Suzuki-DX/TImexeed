package builders

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type PersonalInformationDriveFolderBuilder interface {
	BuildFindActivePersonalInformationDriveFolderByUserIDQuery(userID uint) (*gorm.DB, results.Result)
}

/*
 * 従業員用 個人情報DriveフォルダBuilder
 *
 * 役割：
 * ・JWTから取得した本人userIdをもとにGORMクエリを作成する
 * ・DB実行はRepositoryに任せる
 */
type personalInformationDriveFolderBuilder struct {
	db *gorm.DB
}

/*
 * PersonalInformationDriveFolderBuilder生成
 */
func NewPersonalInformationDriveFolderBuilder(db *gorm.DB) PersonalInformationDriveFolderBuilder {
	return &personalInformationDriveFolderBuilder{db: db}
}

/*
 * 本人の有効な個人情報Driveフォルダ取得用クエリ作成
 */
func (builder *personalInformationDriveFolderBuilder) BuildFindActivePersonalInformationDriveFolderByUserIDQuery(userID uint) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MY_PERSONAL_INFORMATION_DRIVE_FOLDER_QUERY_INVALID_USER_ID",
			"個人情報Driveフォルダ取得条件の作成に失敗しました",
			map[string]any{
				"userId": userID,
			},
		)
	}

	query := builder.db.
		Model(&models.PersonalInformationDriveFolder{}).
		Where("user_id = ?", userID).
		Where("is_deleted = ?", false)

	return query, results.OK(nil, "BUILD_FIND_MY_PERSONAL_INFORMATION_DRIVE_FOLDER_QUERY_SUCCESS", "", nil)
}
