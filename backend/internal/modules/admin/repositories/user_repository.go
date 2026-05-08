package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindUsers(query *gorm.DB) ([]models.User, results.Result)
	CountUsers(query *gorm.DB) (int64, results.Result)
	FindUser(query *gorm.DB) (models.User, results.Result)
	CreateUser(user models.User) (models.User, results.Result)
	SaveUser(user models.User) (models.User, results.Result)
}

/*
 * 管理者用ユーザーRepository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・DBへのCreate / Saveを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 */
type userRepository struct {
	db *gorm.DB
}

/*
 * UserRepository生成
 */
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

/*
 * ユーザー一覧取得
 */
func (repository *userRepository) FindUsers(query *gorm.DB) ([]models.User, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_USERS_QUERY_IS_NIL",
			"ユーザー一覧の取得に失敗しました",
			nil,
		)
	}

	var users []models.User

	if err := query.Find(&users).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_USERS_FAILED",
			"ユーザー一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return users, results.OK(
		nil,
		"FIND_USERS_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー件数取得
 */
func (repository *userRepository) CountUsers(query *gorm.DB) (int64, results.Result) {
	if query == nil {
		return 0, results.InternalServerError(
			"COUNT_USERS_QUERY_IS_NIL",
			"ユーザー件数の取得に失敗しました",
			nil,
		)
	}

	var count int64

	if err := query.Count(&count).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_USERS_FAILED",
			"ユーザー件数の取得に失敗しました",
			err.Error(),
		)
	}

	return count, results.OK(
		nil,
		"COUNT_USERS_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー1件取得
 */
func (repository *userRepository) FindUser(query *gorm.DB) (models.User, results.Result) {
	if query == nil {
		return models.User{}, results.InternalServerError(
			"FIND_USER_QUERY_IS_NIL",
			"ユーザー情報の取得に失敗しました",
			nil,
		)
	}

	var user models.User

	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, results.NotFound(
				"USER_NOT_FOUND",
				"対象ユーザーが見つかりません",
				nil,
			)
		}

		return models.User{}, results.InternalServerError(
			"FIND_USER_FAILED",
			"ユーザー情報の取得に失敗しました",
			err.Error(),
		)
	}

	return user, results.OK(
		nil,
		"FIND_USER_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー作成
 */
func (repository *userRepository) CreateUser(user models.User) (models.User, results.Result) {
	if err := repository.db.Create(&user).Error; err != nil {
		return models.User{}, results.InternalServerError(
			"CREATE_USER_FAILED",
			"ユーザーの作成に失敗しました",
			err.Error(),
		)
	}

	return user, results.OK(
		nil,
		"CREATE_USER_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー保存
 *
 * 更新・論理削除で使う。
 */
func (repository *userRepository) SaveUser(user models.User) (models.User, results.Result) {
	if user.ID == 0 {
		return models.User{}, results.InternalServerError(
			"SAVE_USER_EMPTY_ID",
			"ユーザー情報の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&user).Error; err != nil {
		return models.User{}, results.InternalServerError(
			"SAVE_USER_FAILED",
			"ユーザー情報の保存に失敗しました",
			err.Error(),
		)
	}

	return user, results.OK(
		nil,
		"SAVE_USER_SUCCESS",
		"",
		nil,
	)
}
