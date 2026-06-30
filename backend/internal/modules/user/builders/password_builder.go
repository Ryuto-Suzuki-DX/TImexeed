// 配置先:
// backend/internal/modules/user/builders/password_builder.go

package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * パスワード Builder Interface
 */
type PasswordBuilder interface {
	BuildFindUserByIDQuery(
		userID uint,
	) (*gorm.DB, results.Result)

	BuildUpdatePasswordQuery(
		userID uint,
		passwordHash string,
		passwordChangedAt time.Time,
	) (*gorm.DB, results.Result)
}

/*
 * パスワード Builder
 */
type passwordBuilder struct {
	db *gorm.DB
}

/*
 * パスワード Builder生成
 */
func NewPasswordBuilder(db *gorm.DB) PasswordBuilder {
	return &passwordBuilder{
		db: db,
	}
}

/*
 * ユーザー取得クエリ生成
 */
func (builder *passwordBuilder) BuildFindUserByIDQuery(
	userID uint,
) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"USER_ID_REQUIRED",
			"ユーザーIDが指定されていません。",
			nil,
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("id = ?", userID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_USER_BY_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * パスワード更新クエリ生成
 */
func (builder *passwordBuilder) BuildUpdatePasswordQuery(
	userID uint,
	passwordHash string,
	passwordChangedAt time.Time,
) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"USER_ID_REQUIRED",
			"ユーザーIDが指定されていません。",
			nil,
		)
	}

	if passwordHash == "" {
		return nil, results.BadRequest(
			"PASSWORD_HASH_REQUIRED",
			"パスワード情報が指定されていません。",
			nil,
		)
	}

	updates := map[string]interface{}{
		"password_hash":        passwordHash,
		"must_change_password": false,
		"password_changed_at":  passwordChangedAt,
		"updated_at":           time.Now(),
	}

	query := builder.db.
		Model(&models.User{}).
		Where("id = ?", userID).
		Where("is_deleted = ?", false).
		Updates(updates)

	return query, results.OK(
		nil,
		"BUILD_UPDATE_PASSWORD_QUERY_SUCCESS",
		"",
		nil,
	)
}
