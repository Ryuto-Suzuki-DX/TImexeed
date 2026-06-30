// 配置先:
// backend/internal/modules/user/repositories/password_repository.go

package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * パスワード Repository Interface
 */
type PasswordRepository interface {
	FindUser(query *gorm.DB) (*models.User, results.Result)
	UpdatePassword(query *gorm.DB) results.Result
}

/*
 * パスワード Repository
 */
type passwordRepository struct{}

/*
 * パスワード Repository生成
 */
func NewPasswordRepository() PasswordRepository {
	return &passwordRepository{}
}

/*
 * ユーザー取得
 */
func (repository *passwordRepository) FindUser(
	query *gorm.DB,
) (*models.User, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_USER_QUERY_REQUIRED",
			"ユーザー取得クエリが指定されていません。",
			nil,
		)
	}

	var user models.User

	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, results.NotFound(
				"USER_NOT_FOUND",
				"ユーザーが見つかりません。",
				err,
			)
		}

		return nil, results.InternalServerError(
			"FIND_USER_FAILED",
			"ユーザー情報の取得に失敗しました。",
			err,
		)
	}

	return &user, results.OK(
		user,
		"FIND_USER_SUCCESS",
		"",
		nil,
	)
}

/*
 * パスワード更新
 */
func (repository *passwordRepository) UpdatePassword(
	query *gorm.DB,
) results.Result {
	if query == nil {
		return results.InternalServerError(
			"UPDATE_PASSWORD_QUERY_REQUIRED",
			"パスワード更新クエリが指定されていません。",
			nil,
		)
	}

	if query.Error != nil {
		return results.InternalServerError(
			"UPDATE_PASSWORD_FAILED",
			"パスワードの更新に失敗しました。",
			query.Error,
		)
	}

	if query.RowsAffected == 0 {
		return results.NotFound(
			"USER_NOT_FOUND",
			"更新対象のユーザーが見つかりません。",
			nil,
		)
	}

	return results.OK(
		nil,
		"UPDATE_PASSWORD_SUCCESS",
		"",
		nil,
	)
}
