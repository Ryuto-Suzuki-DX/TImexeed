package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用所属Repository interface
 */
type DepartmentRepository interface {
	FindDepartments(query *gorm.DB) ([]models.Department, results.Result)
	CountDepartments(query *gorm.DB) (int64, results.Result)
	FindDepartment(query *gorm.DB) (models.Department, results.Result)
	CreateDepartment(department models.Department) (models.Department, results.Result)
	SaveDepartment(department models.Department) (models.Department, results.Result)
	CountUsers(query *gorm.DB) (int64, results.Result)
}

/*
 * 管理者用所属Repository
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
type departmentRepository struct {
	db *gorm.DB
}

/*
 * DepartmentRepository生成
 */
func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{
		db: db,
	}
}

/*
 * 所属一覧取得
 */
func (repository *departmentRepository) FindDepartments(query *gorm.DB) ([]models.Department, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_DEPARTMENTS_QUERY_IS_NIL",
			"所属一覧の取得に失敗しました",
			nil,
		)
	}

	var departments []models.Department

	if err := query.Find(&departments).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_DEPARTMENTS_FAILED",
			"所属一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return departments, results.OK(
		nil,
		"FIND_DEPARTMENTS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 所属件数取得
 */
func (repository *departmentRepository) CountDepartments(query *gorm.DB) (int64, results.Result) {
	if query == nil {
		return 0, results.InternalServerError(
			"COUNT_DEPARTMENTS_QUERY_IS_NIL",
			"所属件数の取得に失敗しました",
			nil,
		)
	}

	var count int64

	if err := query.Count(&count).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_DEPARTMENTS_FAILED",
			"所属件数の取得に失敗しました",
			err.Error(),
		)
	}

	return count, results.OK(
		nil,
		"COUNT_DEPARTMENTS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 所属1件取得
 */
func (repository *departmentRepository) FindDepartment(query *gorm.DB) (models.Department, results.Result) {
	if query == nil {
		return models.Department{}, results.InternalServerError(
			"FIND_DEPARTMENT_QUERY_IS_NIL",
			"所属情報の取得に失敗しました",
			nil,
		)
	}

	var department models.Department

	if err := query.First(&department).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Department{}, results.NotFound(
				"DEPARTMENT_NOT_FOUND",
				"対象所属が見つかりません",
				nil,
			)
		}

		return models.Department{}, results.InternalServerError(
			"FIND_DEPARTMENT_FAILED",
			"所属情報の取得に失敗しました",
			err.Error(),
		)
	}

	return department, results.OK(
		nil,
		"FIND_DEPARTMENT_SUCCESS",
		"",
		nil,
	)
}

/*
 * 所属作成
 */
func (repository *departmentRepository) CreateDepartment(department models.Department) (models.Department, results.Result) {
	if err := repository.db.Create(&department).Error; err != nil {
		return models.Department{}, results.InternalServerError(
			"CREATE_DEPARTMENT_FAILED",
			"所属の作成に失敗しました",
			err.Error(),
		)
	}

	return department, results.OK(
		nil,
		"CREATE_DEPARTMENT_SUCCESS",
		"",
		nil,
	)
}

/*
 * 所属保存
 *
 * 更新・論理削除で使う。
 */
func (repository *departmentRepository) SaveDepartment(department models.Department) (models.Department, results.Result) {
	if department.ID == 0 {
		return models.Department{}, results.InternalServerError(
			"SAVE_DEPARTMENT_EMPTY_ID",
			"所属情報の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&department).Error; err != nil {
		return models.Department{}, results.InternalServerError(
			"SAVE_DEPARTMENT_FAILED",
			"所属情報の保存に失敗しました",
			err.Error(),
		)
	}

	return department, results.OK(
		nil,
		"SAVE_DEPARTMENT_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー件数取得
 *
 * 所属削除前に、所属へ紐づく有効ユーザー数を確認するために使う。
 */
func (repository *departmentRepository) CountUsers(query *gorm.DB) (int64, results.Result) {
	if query == nil {
		return 0, results.InternalServerError(
			"COUNT_USERS_BY_DEPARTMENT_QUERY_IS_NIL",
			"所属に紐づくユーザー件数の取得に失敗しました",
			nil,
		)
	}

	var count int64

	if err := query.Count(&count).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_USERS_BY_DEPARTMENT_FAILED",
			"所属に紐づくユーザー件数の取得に失敗しました",
			err.Error(),
		)
	}

	return count, results.OK(
		nil,
		"COUNT_USERS_BY_DEPARTMENT_SUCCESS",
		"",
		nil,
	)
}
