package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用所属Builder interface
 */
type DepartmentBuilder interface {
	BuildSearchDepartmentsQuery(req types.SearchDepartmentsRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindDepartmentByIDQuery(departmentID uint) (*gorm.DB, results.Result)
	BuildCountActiveDepartmentByNameQuery(name string) (*gorm.DB, results.Result)
	BuildCountActiveDepartmentByNameExceptIDQuery(name string, departmentID uint) (*gorm.DB, results.Result)
	BuildCountActiveUsersByDepartmentIDQuery(departmentID uint) (*gorm.DB, results.Result)
	BuildCreateDepartmentModel(req types.CreateDepartmentRequest) (models.Department, results.Result)
	BuildUpdateDepartmentModel(currentDepartment models.Department, req types.UpdateDepartmentRequest) (models.Department, results.Result)
	BuildDeleteDepartmentModel(currentDepartment models.Department) (models.Department, results.Result)
}

/*
 * 管理者用所属Builder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Serviceから受け取ったRequestをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Count / Create / Save はRepositoryに任せる
 */
type departmentBuilder struct {
	db *gorm.DB
}

/*
 * DepartmentBuilder生成
 */
func NewDepartmentBuilder(db *gorm.DB) DepartmentBuilder {
	return &departmentBuilder{
		db: db,
	}
}

/*
 * 所属検索用クエリ作成
 *
 * searchQuery：
 * ・一覧取得用
 * ・offset / limit / order を含む
 *
 * countQuery：
 * ・総件数取得用
 * ・offset / limit は含めない
 */
func (builder *departmentBuilder) BuildSearchDepartmentsQuery(req types.SearchDepartmentsRequest) (*gorm.DB, *gorm.DB, results.Result) {
	if req.Offset < 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_DEPARTMENTS_QUERY_INVALID_OFFSET",
			"所属検索条件の作成に失敗しました",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	if req.Limit <= 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_DEPARTMENTS_QUERY_INVALID_LIMIT",
			"所属検索条件の作成に失敗しました",
			map[string]any{
				"limit": req.Limit,
			},
		)
	}

	searchQuery := builder.db.Model(&models.Department{})
	countQuery := builder.db.Model(&models.Department{})

	searchQuery = applySearchDepartmentsCondition(searchQuery, req)
	countQuery = applySearchDepartmentsCondition(countQuery, req)

	searchQuery = searchQuery.
		Order("id ASC").
		Offset(req.Offset).
		Limit(req.Limit)

	return searchQuery, countQuery, results.OK(
		nil,
		"BUILD_SEARCH_DEPARTMENTS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 所属ID検索用クエリ作成
 *
 * 論理削除済み所属は対象外にする。
 */
func (builder *departmentBuilder) BuildFindDepartmentByIDQuery(departmentID uint) (*gorm.DB, results.Result) {
	if departmentID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_DEPARTMENT_BY_ID_QUERY_INVALID_DEPARTMENT_ID",
			"所属取得条件の作成に失敗しました",
			map[string]any{
				"departmentId": departmentID,
			},
		)
	}

	query := builder.db.
		Model(&models.Department{}).
		Where("id = ?", departmentID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_DEPARTMENT_BY_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有効所属の所属名件数確認クエリ作成
 *
 * 新規作成時の重複確認に使う。
 */
func (builder *departmentBuilder) BuildCountActiveDepartmentByNameQuery(name string) (*gorm.DB, results.Result) {
	if name == "" {
		return nil, results.BadRequest(
			"BUILD_COUNT_ACTIVE_DEPARTMENT_BY_NAME_QUERY_EMPTY_NAME",
			"所属名重複確認条件の作成に失敗しました",
			nil,
		)
	}

	query := builder.db.
		Model(&models.Department{}).
		Where("name = ?", name).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_COUNT_ACTIVE_DEPARTMENT_BY_NAME_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 指定所属以外の有効所属の所属名件数確認クエリ作成
 *
 * 更新時の重複確認に使う。
 */
func (builder *departmentBuilder) BuildCountActiveDepartmentByNameExceptIDQuery(name string, departmentID uint) (*gorm.DB, results.Result) {
	if name == "" {
		return nil, results.BadRequest(
			"BUILD_COUNT_ACTIVE_DEPARTMENT_BY_NAME_EXCEPT_ID_QUERY_EMPTY_NAME",
			"所属名重複確認条件の作成に失敗しました",
			nil,
		)
	}

	if departmentID == 0 {
		return nil, results.BadRequest(
			"BUILD_COUNT_ACTIVE_DEPARTMENT_BY_NAME_EXCEPT_ID_QUERY_INVALID_DEPARTMENT_ID",
			"所属名重複確認条件の作成に失敗しました",
			map[string]any{
				"departmentId": departmentID,
			},
		)
	}

	query := builder.db.
		Model(&models.Department{}).
		Where("name = ?", name).
		Where("id <> ?", departmentID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_COUNT_ACTIVE_DEPARTMENT_BY_NAME_EXCEPT_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 所属に紐づく有効ユーザー件数取得用クエリ作成
 *
 * 所属削除前の紐づきチェックに使う。
 */
func (builder *departmentBuilder) BuildCountActiveUsersByDepartmentIDQuery(departmentID uint) (*gorm.DB, results.Result) {
	if departmentID == 0 {
		return nil, results.BadRequest(
			"BUILD_COUNT_ACTIVE_USERS_BY_DEPARTMENT_ID_QUERY_INVALID_DEPARTMENT_ID",
			"所属に紐づくユーザー件数取得条件の作成に失敗しました",
			map[string]any{
				"departmentId": departmentID,
			},
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("department_id = ?", departmentID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_COUNT_ACTIVE_USERS_BY_DEPARTMENT_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 所属作成用Model作成
 */
func (builder *departmentBuilder) BuildCreateDepartmentModel(req types.CreateDepartmentRequest) (models.Department, results.Result) {
	if req.Name == "" {
		return models.Department{}, results.BadRequest(
			"BUILD_CREATE_DEPARTMENT_MODEL_EMPTY_NAME",
			"所属作成データの作成に失敗しました",
			nil,
		)
	}

	department := models.Department{
		Name:      req.Name,
		IsDeleted: false,
	}

	return department, results.OK(
		nil,
		"BUILD_CREATE_DEPARTMENT_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 所属更新用Model作成
 */
func (builder *departmentBuilder) BuildUpdateDepartmentModel(
	currentDepartment models.Department,
	req types.UpdateDepartmentRequest,
) (models.Department, results.Result) {
	if currentDepartment.ID == 0 {
		return models.Department{}, results.BadRequest(
			"BUILD_UPDATE_DEPARTMENT_MODEL_EMPTY_CURRENT_DEPARTMENT",
			"所属更新データの作成に失敗しました",
			nil,
		)
	}

	if req.Name == "" {
		return models.Department{}, results.BadRequest(
			"BUILD_UPDATE_DEPARTMENT_MODEL_EMPTY_NAME",
			"所属更新データの作成に失敗しました",
			nil,
		)
	}

	currentDepartment.Name = req.Name

	return currentDepartment, results.OK(
		nil,
		"BUILD_UPDATE_DEPARTMENT_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 所属論理削除用Model作成
 */
func (builder *departmentBuilder) BuildDeleteDepartmentModel(currentDepartment models.Department) (models.Department, results.Result) {
	if currentDepartment.ID == 0 {
		return models.Department{}, results.BadRequest(
			"BUILD_DELETE_DEPARTMENT_MODEL_EMPTY_CURRENT_DEPARTMENT",
			"所属削除データの作成に失敗しました",
			nil,
		)
	}

	now := time.Now()

	currentDepartment.IsDeleted = true
	currentDepartment.DeletedAt = &now

	return currentDepartment, results.OK(
		nil,
		"BUILD_DELETE_DEPARTMENT_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 所属検索条件をGORMクエリへ適用する
 */
func applySearchDepartmentsCondition(query *gorm.DB, req types.SearchDepartmentsRequest) *gorm.DB {
	if !req.IncludeDeleted {
		query = query.Where("is_deleted = ?", false)
	}

	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where(
			"name ILIKE ?",
			keyword,
		)
	}

	return query
}
