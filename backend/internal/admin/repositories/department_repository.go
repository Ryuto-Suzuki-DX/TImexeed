package repositories

import (
	"timexeed/backend/internal/models"

	"gorm.io/gorm"
)

/*
 * 管理者用所属Repository
 *
 * Repositoryの責務:
 * - DB実行のみを担当する
 */
type DepartmentRepositoryInterface interface {
	FindDepartments(query *gorm.DB) ([]models.Department, error)
	FindDepartment(query *gorm.DB) (models.Department, error)
	Count(query *gorm.DB) (int64, error)
	CreateDepartment(db *gorm.DB, department *models.Department) error
	SaveDepartment(db *gorm.DB, department *models.Department) error
}

type DepartmentRepository struct{}

/*
 * DepartmentRepositoryを生成する
 */
func NewDepartmentRepository() *DepartmentRepository {
	return &DepartmentRepository{}
}

/*
 * 所属一覧取得
 */
func (r *DepartmentRepository) FindDepartments(query *gorm.DB) ([]models.Department, error) {
	var departments []models.Department

	if err := query.Find(&departments).Error; err != nil {
		return nil, err
	}

	return departments, nil
}

/*
 * 所属1件取得
 */
func (r *DepartmentRepository) FindDepartment(query *gorm.DB) (models.Department, error) {
	var department models.Department

	if err := query.First(&department).Error; err != nil {
		return models.Department{}, err
	}

	return department, nil
}

/*
 * 件数取得
 */
func (r *DepartmentRepository) Count(query *gorm.DB) (int64, error) {
	var count int64

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

/*
 * 所属作成
 */
func (r *DepartmentRepository) CreateDepartment(db *gorm.DB, department *models.Department) error {
	return db.Create(department).Error
}

/*
 * 所属保存
 */
func (r *DepartmentRepository) SaveDepartment(db *gorm.DB, department *models.Department) error {
	return db.Save(department).Error
}
