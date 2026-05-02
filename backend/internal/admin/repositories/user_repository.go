package repositories

import (
	"timexeed/backend/internal/models"

	"gorm.io/gorm"
)

/*
 * 管理者用ユーザーRepository
 *
 * Repositoryの責務:
 * - Builderから渡されたクエリを実行する
 * - Create / Save などDB操作を実行する
 * - メッセージやレスポンス形式は作らない
 */

// インターフェース
type UserRepositoryInterface interface {
	FindUsers(query *gorm.DB) ([]models.User, error)
	FindUser(query *gorm.DB) (models.User, error)
	Count(query *gorm.DB) (int64, error)
	CreateUser(db *gorm.DB, user *models.User) error
	SaveUser(db *gorm.DB, user *models.User) error
}

type UserRepository struct{}

/*
 * UserRepositoryを生成する
 */
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

/*
 * ユーザー一覧を取得する
 */
func (r *UserRepository) FindUsers(query *gorm.DB) ([]models.User, error) {
	var users []models.User

	err := query.Find(&users).Error

	return users, err
}

/*
 * ユーザーを1件取得する
 */
func (r *UserRepository) FindUser(query *gorm.DB) (models.User, error) {
	var user models.User

	err := query.First(&user).Error

	return user, err
}

/*
 * 件数を取得する
 */
func (r *UserRepository) Count(query *gorm.DB) (int64, error) {
	var count int64

	err := query.Count(&count).Error

	return count, err
}

/*
 * ユーザーを作成する
 */
func (r *UserRepository) CreateUser(db *gorm.DB, user *models.User) error {
	return db.Create(user).Error
}

/*
 * ユーザーを保存する
 */
func (r *UserRepository) SaveUser(db *gorm.DB, user *models.User) error {
	return db.Save(user).Error
}
