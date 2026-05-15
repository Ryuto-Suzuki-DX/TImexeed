package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用ユーザー給与詳細Repository interface
 */
type UserSalaryDetailRepository interface {
	FindUserSalaryDetails(query *gorm.DB) ([]models.UserSalaryDetail, results.Result)
	CountUserSalaryDetails(query *gorm.DB) (int64, results.Result)
	FindUserSalaryDetail(query *gorm.DB) (models.UserSalaryDetail, results.Result)
	CreateUserSalaryDetail(userSalaryDetail models.UserSalaryDetail) (models.UserSalaryDetail, results.Result)
	SaveUserSalaryDetail(userSalaryDetail models.UserSalaryDetail) (models.UserSalaryDetail, results.Result)
}

/*
 * 管理者用ユーザー給与詳細Repository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・DBへのCreate / Saveを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 * ・適用期間重複などの業務チェックはServiceに任せる
 */
type userSalaryDetailRepository struct {
	db *gorm.DB
}

/*
 * UserSalaryDetailRepository生成
 */
func NewUserSalaryDetailRepository(db *gorm.DB) UserSalaryDetailRepository {
	return &userSalaryDetailRepository{
		db: db,
	}
}

/*
 * ユーザー給与詳細一覧取得
 */
func (repository *userSalaryDetailRepository) FindUserSalaryDetails(query *gorm.DB) ([]models.UserSalaryDetail, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_USER_SALARY_DETAILS_QUERY_IS_NIL",
			"ユーザー給与詳細一覧の取得に失敗しました",
			nil,
		)
	}

	var userSalaryDetails []models.UserSalaryDetail

	if err := query.Find(&userSalaryDetails).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_USER_SALARY_DETAILS_FAILED",
			"ユーザー給与詳細一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return userSalaryDetails, results.OK(
		nil,
		"FIND_USER_SALARY_DETAILS_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー給与詳細件数取得
 */
func (repository *userSalaryDetailRepository) CountUserSalaryDetails(query *gorm.DB) (int64, results.Result) {
	if query == nil {
		return 0, results.InternalServerError(
			"COUNT_USER_SALARY_DETAILS_QUERY_IS_NIL",
			"ユーザー給与詳細件数の取得に失敗しました",
			nil,
		)
	}

	var count int64

	if err := query.Count(&count).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_USER_SALARY_DETAILS_FAILED",
			"ユーザー給与詳細件数の取得に失敗しました",
			err.Error(),
		)
	}

	return count, results.OK(
		nil,
		"COUNT_USER_SALARY_DETAILS_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー給与詳細1件取得
 */
func (repository *userSalaryDetailRepository) FindUserSalaryDetail(query *gorm.DB) (models.UserSalaryDetail, results.Result) {
	if query == nil {
		return models.UserSalaryDetail{}, results.InternalServerError(
			"FIND_USER_SALARY_DETAIL_QUERY_IS_NIL",
			"ユーザー給与詳細の取得に失敗しました",
			nil,
		)
	}

	var userSalaryDetail models.UserSalaryDetail

	if err := query.First(&userSalaryDetail).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.UserSalaryDetail{}, results.NotFound(
				"USER_SALARY_DETAIL_NOT_FOUND",
				"対象ユーザー給与詳細が見つかりません",
				nil,
			)
		}

		return models.UserSalaryDetail{}, results.InternalServerError(
			"FIND_USER_SALARY_DETAIL_FAILED",
			"ユーザー給与詳細の取得に失敗しました",
			err.Error(),
		)
	}

	return userSalaryDetail, results.OK(
		nil,
		"FIND_USER_SALARY_DETAIL_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー給与詳細作成
 */
func (repository *userSalaryDetailRepository) CreateUserSalaryDetail(userSalaryDetail models.UserSalaryDetail) (models.UserSalaryDetail, results.Result) {
	if err := repository.db.Create(&userSalaryDetail).Error; err != nil {
		return models.UserSalaryDetail{}, results.InternalServerError(
			"CREATE_USER_SALARY_DETAIL_FAILED",
			"ユーザー給与詳細の作成に失敗しました",
			err.Error(),
		)
	}

	return userSalaryDetail, results.OK(
		nil,
		"CREATE_USER_SALARY_DETAIL_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー給与詳細保存
 *
 * 更新・論理削除で使う。
 */
func (repository *userSalaryDetailRepository) SaveUserSalaryDetail(userSalaryDetail models.UserSalaryDetail) (models.UserSalaryDetail, results.Result) {
	if userSalaryDetail.ID == 0 {
		return models.UserSalaryDetail{}, results.InternalServerError(
			"SAVE_USER_SALARY_DETAIL_EMPTY_ID",
			"ユーザー給与詳細の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&userSalaryDetail).Error; err != nil {
		return models.UserSalaryDetail{}, results.InternalServerError(
			"SAVE_USER_SALARY_DETAIL_FAILED",
			"ユーザー給与詳細の保存に失敗しました",
			err.Error(),
		)
	}

	return userSalaryDetail, results.OK(
		nil,
		"SAVE_USER_SALARY_DETAIL_SUCCESS",
		"",
		nil,
	)
}
