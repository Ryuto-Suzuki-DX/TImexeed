package services

import (
	"strings"

	"timexeed/backend/internal/admin/builders"
	"timexeed/backend/internal/admin/repositories"
	"timexeed/backend/internal/admin/types"
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用所属Service
 *
 * Serviceの責務:
 * - Builderにクエリ/model作成を依頼する
 * - RepositoryにDB実行を依頼する
 * - 実行結果をレスポンス型に整える
 * - ControllerへResultで返す
 */

// インターフェース
type DepartmentServiceInterface interface {
	InvalidRequest() results.Result
	InvalidDepartmentID() results.Result
	SearchDepartments(req types.SearchDepartmentsRequest) results.Result
	GetDepartment(id uint) results.Result
	CreateDepartment(req types.CreateDepartmentRequest) results.Result
	UpdateDepartment(id uint, req types.UpdateDepartmentRequest) results.Result
	DeleteDepartment(id uint) results.Result
}

type DepartmentService struct {
	db                   *gorm.DB
	departmentRepository repositories.DepartmentRepositoryInterface
	departmentBuilder    builders.DepartmentBuilderInterface
}

/*
 * DepartmentServiceを生成する
 */
func NewDepartmentService(
	db *gorm.DB,
	departmentRepository *repositories.DepartmentRepository,
	departmentBuilder *builders.DepartmentBuilder,
) *DepartmentService {
	return &DepartmentService{
		db:                   db,
		departmentRepository: departmentRepository,
		departmentBuilder:    departmentBuilder,
	}
}

/*
 * 不正なリクエスト形式
 */
func (s *DepartmentService) InvalidRequest() results.Result {
	return results.ValidationError(map[string]string{
		"request": "リクエスト形式が正しくありません",
	})
}

/*
 * 不正な所属ID
 */
func (s *DepartmentService) InvalidDepartmentID() results.Result {
	return results.BadRequest("INVALID_DEPARTMENT_ID", "所属IDが正しくありません")
}

/*
 * 所属一覧取得
 */
func (s *DepartmentService) SearchDepartments(req types.SearchDepartmentsRequest) results.Result {
	condition := types.SearchDepartmentsCondition{
		Keyword:        strings.TrimSpace(req.Keyword),
		IncludeDeleted: req.IncludeDeleted,
	}

	query := s.departmentBuilder.BuildSearchDepartmentsQuery(s.db, condition)

	departments, err := s.departmentRepository.FindDepartments(query)
	if err != nil {
		return results.InternalServerError("所属一覧の取得に失敗しました")
	}

	departmentResponses := make([]types.DepartmentResponse, 0, len(departments))

	for _, department := range departments {
		departmentResponses = append(departmentResponses, toDepartmentResponse(department))
	}

	return results.Success(types.SearchDepartmentsResponse{
		Departments: departmentResponses,
	}, "所属一覧を取得しました")
}

/*
 * 所属詳細取得
 */
func (s *DepartmentService) GetDepartment(id uint) results.Result {
	query := s.departmentBuilder.BuildFindActiveDepartmentByIDQuery(s.db, id)

	department, err := s.departmentRepository.FindDepartment(query)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return results.NotFound("所属が見つかりません")
		}

		return results.InternalServerError("所属の取得に失敗しました")
	}

	return results.Success(toDepartmentResponse(department), "所属を取得しました")
}

/*
 * 所属新規作成
 */
func (s *DepartmentService) CreateDepartment(req types.CreateDepartmentRequest) results.Result {
	req = normalizeCreateDepartmentRequest(req)

	validationErrors := validateCreateDepartmentRequest(req)
	if len(validationErrors) > 0 {
		return results.ValidationError(validationErrors)
	}

	countQuery := s.departmentBuilder.BuildCountActiveDepartmentByNameQuery(s.db, req.Name)

	count, err := s.departmentRepository.Count(countQuery)
	if err != nil {
		return results.InternalServerError("所属名の確認に失敗しました")
	}

	if count > 0 {
		return results.Conflict("この所属名はすでに使用されています")
	}

	department := s.departmentBuilder.BuildCreateDepartmentModel(req)

	if err := s.departmentRepository.CreateDepartment(s.db, &department); err != nil {
		return results.InternalServerError("所属の作成に失敗しました")
	}

	return results.Created(toDepartmentResponse(department), "所属を作成しました")
}

/*
 * 所属更新
 */
func (s *DepartmentService) UpdateDepartment(id uint, req types.UpdateDepartmentRequest) results.Result {
	req = normalizeUpdateDepartmentRequest(req)

	validationErrors := validateUpdateDepartmentRequest(req)
	if len(validationErrors) > 0 {
		return results.ValidationError(validationErrors)
	}

	findQuery := s.departmentBuilder.BuildFindActiveDepartmentByIDQuery(s.db, id)

	department, err := s.departmentRepository.FindDepartment(findQuery)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return results.NotFound("所属が見つかりません")
		}

		return results.InternalServerError("所属の取得に失敗しました")
	}

	countQuery := s.departmentBuilder.BuildCountActiveDepartmentByNameExceptIDQuery(s.db, req.Name, id)

	count, err := s.departmentRepository.Count(countQuery)
	if err != nil {
		return results.InternalServerError("所属名の確認に失敗しました")
	}

	if count > 0 {
		return results.Conflict("この所属名はすでに使用されています")
	}

	department = s.departmentBuilder.BuildUpdateDepartmentModel(department, req)

	if err := s.departmentRepository.SaveDepartment(s.db, &department); err != nil {
		return results.InternalServerError("所属の更新に失敗しました")
	}

	return results.Success(toDepartmentResponse(department), "所属を更新しました")
}

/*
 * 所属論理削除
 */
func (s *DepartmentService) DeleteDepartment(id uint) results.Result {
	findQuery := s.departmentBuilder.BuildFindActiveDepartmentByIDQuery(s.db, id)

	department, err := s.departmentRepository.FindDepartment(findQuery)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return results.NotFound("所属が見つかりません")
		}

		return results.InternalServerError("所属の取得に失敗しました")
	}

	countUserQuery := s.departmentBuilder.BuildCountActiveUsersByDepartmentIDQuery(s.db, id)

	count, err := s.departmentRepository.Count(countUserQuery)
	if err != nil {
		return results.InternalServerError("所属に紐づくユーザーの確認に失敗しました")
	}

	if count > 0 {
		return results.Conflict("この所属にユーザーが紐づいているため削除できません")
	}

	department = s.departmentBuilder.BuildDeleteDepartmentModel(department)

	if err := s.departmentRepository.SaveDepartment(s.db, &department); err != nil {
		return results.InternalServerError("所属の削除に失敗しました")
	}

	return results.Success(nil, "所属を削除しました")
}

/*
 * modelをレスポンス型へ変換する
 */
func toDepartmentResponse(department models.Department) types.DepartmentResponse {
	return types.DepartmentResponse{
		ID:        department.ID,
		Name:      department.Name,
		IsDeleted: department.IsDeleted,
	}
}

/*
 * 所属作成リクエストの前処理
 */
func normalizeCreateDepartmentRequest(req types.CreateDepartmentRequest) types.CreateDepartmentRequest {
	req.Name = strings.TrimSpace(req.Name)

	return req
}

/*
 * 所属更新リクエストの前処理
 */
func normalizeUpdateDepartmentRequest(req types.UpdateDepartmentRequest) types.UpdateDepartmentRequest {
	req.Name = strings.TrimSpace(req.Name)

	return req
}

/*
 * 所属作成時の入力チェック
 */
func validateCreateDepartmentRequest(req types.CreateDepartmentRequest) map[string]string {
	errors := map[string]string{}

	if req.Name == "" {
		errors["name"] = "所属名を入力してください"
	}

	return errors
}

/*
 * 所属更新時の入力チェック
 */
func validateUpdateDepartmentRequest(req types.UpdateDepartmentRequest) map[string]string {
	errors := map[string]string{}

	if req.Name == "" {
		errors["name"] = "所属名を入力してください"
	}

	return errors
}
