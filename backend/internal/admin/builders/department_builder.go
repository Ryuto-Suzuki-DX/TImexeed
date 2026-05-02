package builders

import (
	"time"

	"timexeed/backend/internal/admin/types"
	"timexeed/backend/internal/models"

	"gorm.io/gorm"
)

/*
 * 管理者用所属Builder
 *
 * Builderの責務:
 * - DBクエリを組み立てる
 * - modelを組み立てる
 */
type DepartmentBuilderInterface interface {
	BuildSearchDepartmentsQuery(db *gorm.DB, condition types.SearchDepartmentsCondition) *gorm.DB
	BuildFindActiveDepartmentByIDQuery(db *gorm.DB, id uint) *gorm.DB
	BuildCountActiveDepartmentByNameQuery(db *gorm.DB, name string) *gorm.DB
	BuildCountActiveDepartmentByNameExceptIDQuery(db *gorm.DB, name string, id uint) *gorm.DB
	BuildCountActiveUsersByDepartmentIDQuery(db *gorm.DB, departmentID uint) *gorm.DB
	BuildCreateDepartmentModel(req types.CreateDepartmentRequest) models.Department
	BuildUpdateDepartmentModel(department models.Department, req types.UpdateDepartmentRequest) models.Department
	BuildDeleteDepartmentModel(department models.Department) models.Department
	BuildFindActiveDepartmentsByIDsQuery(db *gorm.DB, ids []uint) *gorm.DB
}

type DepartmentBuilder struct{}

/*
 * DepartmentBuilderを生成する
 */
func NewDepartmentBuilder() *DepartmentBuilder {
	return &DepartmentBuilder{}
}

/*
 * 所属一覧検索クエリ作成
 */
func (b *DepartmentBuilder) BuildSearchDepartmentsQuery(
	db *gorm.DB,
	condition types.SearchDepartmentsCondition,
) *gorm.DB {
	query := db.Model(&models.Department{})

	if !condition.IncludeDeleted {
		query = query.Where("is_deleted = ?", false)
	}

	if condition.Keyword != "" {
		likeKeyword := "%" + condition.Keyword + "%"
		query = query.Where("name LIKE ?", likeKeyword)
	}

	return query.Order("id ASC")
}

/*
 * 有効な所属をIDで取得するクエリ作成
 */
func (b *DepartmentBuilder) BuildFindActiveDepartmentByIDQuery(db *gorm.DB, id uint) *gorm.DB {
	return db.
		Model(&models.Department{}).
		Where("id = ? AND is_deleted = ?", id, false)
}

/*
 * 有効な所属名の件数取得クエリ作成
 */
func (b *DepartmentBuilder) BuildCountActiveDepartmentByNameQuery(db *gorm.DB, name string) *gorm.DB {
	return db.
		Model(&models.Department{}).
		Where("name = ? AND is_deleted = ?", name, false)
}

/*
 * 自分以外の有効な所属名の件数取得クエリ作成
 */
func (b *DepartmentBuilder) BuildCountActiveDepartmentByNameExceptIDQuery(
	db *gorm.DB,
	name string,
	id uint,
) *gorm.DB {
	return db.
		Model(&models.Department{}).
		Where("name = ? AND id <> ? AND is_deleted = ?", name, id, false)
}

/*
 * 所属に紐づく有効ユーザー件数取得クエリ作成
 */
func (b *DepartmentBuilder) BuildCountActiveUsersByDepartmentIDQuery(db *gorm.DB, departmentID uint) *gorm.DB {
	return db.
		Model(&models.User{}).
		Where("department_id = ? AND is_deleted = ?", departmentID, false)
}

/*
 * 所属作成model作成
 */
func (b *DepartmentBuilder) BuildCreateDepartmentModel(req types.CreateDepartmentRequest) models.Department {
	return models.Department{
		Name:      req.Name,
		IsDeleted: false,
	}
}

/*
 * 所属更新model作成
 */
func (b *DepartmentBuilder) BuildUpdateDepartmentModel(
	department models.Department,
	req types.UpdateDepartmentRequest,
) models.Department {
	department.Name = req.Name

	return department
}

/*
 * 所属論理削除model作成
 */
func (b *DepartmentBuilder) BuildDeleteDepartmentModel(department models.Department) models.Department {
	now := time.Now()

	department.IsDeleted = true
	department.DeletedAt = &now

	return department
}

/*
 * 有効な所属をID一覧で取得するクエリ作成
 */
func (b *DepartmentBuilder) BuildFindActiveDepartmentsByIDsQuery(db *gorm.DB, ids []uint) *gorm.DB {
	return db.
		Model(&models.Department{}).
		Where("id IN ? AND is_deleted = ?", ids, false)
}
