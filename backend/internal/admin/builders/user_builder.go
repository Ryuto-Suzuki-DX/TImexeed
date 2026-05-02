package builders

import (
	"time"

	"timexeed/backend/internal/admin/types"
	"timexeed/backend/internal/models"

	"gorm.io/gorm"
)

/*
 * 管理者用ユーザーBuilder
 *
 * Builderの責務:
 * - 検索クエリを作成する
 * - 取得クエリを作成する
 * - 重複確認クエリを作成する
 * - 作成用modelを組み立てる
 * - 更新用modelを組み立てる
 * - 削除用modelを組み立てる
 */

// インターフェース
type UserBuilderInterface interface {
	BuildSearchUsersQuery(db *gorm.DB, condition types.SearchUsersCondition) *gorm.DB
	BuildFindActiveUserByIDQuery(db *gorm.DB, id uint) *gorm.DB
	BuildCountActiveUserByEmailQuery(db *gorm.DB, email string) *gorm.DB
	BuildCountActiveUserByEmailExceptIDQuery(db *gorm.DB, email string, id uint) *gorm.DB
	BuildCreateUserModel(req types.CreateUserRequest, passwordHash string) models.User
	BuildUpdateUserModel(user models.User, req types.UpdateUserRequest, passwordHash string) models.User
	BuildDeleteUserModel(user models.User) models.User
}

type UserBuilder struct{}

/*
 * UserBuilderを生成する
 */
func NewUserBuilder() *UserBuilder {
	return &UserBuilder{}
}

/*
 * ユーザー一覧取得クエリを作成する
 */
func (b *UserBuilder) BuildSearchUsersQuery(db *gorm.DB, condition types.SearchUsersCondition) *gorm.DB {
	query := db.Model(&models.User{})

	if !condition.IncludeDeleted {
		query = query.Where("is_deleted = ?", false)
	}

	if condition.Keyword != "" {
		keyword := "%" + condition.Keyword + "%"

		query = query.Where(
			"name LIKE ? OR email LIKE ?",
			keyword,
			keyword,
		)
	}

	return query.Order("id ASC")
}

/*
 * 有効ユーザー詳細取得クエリを作成する
 */
func (b *UserBuilder) BuildFindActiveUserByIDQuery(db *gorm.DB, id uint) *gorm.DB {
	return db.
		Model(&models.User{}).
		Where("id = ? AND is_deleted = ?", id, false)
}

/*
 * メールアドレス重複確認クエリを作成する
 */
func (b *UserBuilder) BuildCountActiveUserByEmailQuery(db *gorm.DB, email string) *gorm.DB {
	return db.
		Model(&models.User{}).
		Where("email = ? AND is_deleted = ?", email, false)
}

/*
 * 自分以外のメールアドレス重複確認クエリを作成する
 */
func (b *UserBuilder) BuildCountActiveUserByEmailExceptIDQuery(db *gorm.DB, email string, id uint) *gorm.DB {
	return db.
		Model(&models.User{}).
		Where("email = ? AND id <> ? AND is_deleted = ?", email, id, false)
}

/*
 * 作成用ユーザーmodelを作成する
 */
func (b *UserBuilder) BuildCreateUserModel(req types.CreateUserRequest, passwordHash string) models.User {
	return models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         req.Role,
		DepartmentID: req.DepartmentID,
		IsDeleted:    false,
	}
}

/*
 * 更新用ユーザーmodelを作成する
 *
 * 既存userに更新内容を詰めて返す
 * passwordHashが空の場合はパスワードを変更しない
 */
func (b *UserBuilder) BuildUpdateUserModel(user models.User, req types.UpdateUserRequest, passwordHash string) models.User {
	user.Name = req.Name
	user.Email = req.Email
	user.Role = req.Role
	user.DepartmentID = req.DepartmentID

	if passwordHash != "" {
		user.PasswordHash = passwordHash
	}

	return user
}

/*
 * 論理削除用ユーザーmodelを作成する
 */
func (b *UserBuilder) BuildDeleteUserModel(user models.User) models.User {
	now := time.Now()

	user.IsDeleted = true
	user.DeletedAt = &now

	return user
}
