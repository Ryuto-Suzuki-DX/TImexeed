package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 管理者用所属Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type DepartmentService interface {
	SearchDepartments(req types.SearchDepartmentsRequest) results.Result
	GetDepartmentDetail(req types.DepartmentDetailRequest) results.Result
	CreateDepartment(req types.CreateDepartmentRequest) results.Result
	UpdateDepartment(req types.UpdateDepartmentRequest) results.Result
	DeleteDepartment(req types.DeleteDepartmentRequest) results.Result
}

/*
 * 管理者用所属Service
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや更新用Modelを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB処理を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type departmentService struct {
	departmentBuilder    builders.DepartmentBuilder
	departmentRepository repositories.DepartmentRepository
}

/*
 * DepartmentService生成
 */
func NewDepartmentService(
	departmentBuilder builders.DepartmentBuilder,
	departmentRepository repositories.DepartmentRepository,
) DepartmentService {
	return &departmentService{
		departmentBuilder:    departmentBuilder,
		departmentRepository: departmentRepository,
	}
}

/*
 * models.Departmentをフロント返却用DepartmentResponseへ変換する
 *
 * 日付はtime.Time / *time.Timeのまま返す。
 * 表示形式の整形はフロント側で行う。
 */
func toDepartmentResponse(department models.Department) types.DepartmentResponse {
	return types.DepartmentResponse{
		ID:        department.ID,
		Name:      department.Name,
		IsDeleted: department.IsDeleted,
		CreatedAt: department.CreatedAt,
		UpdatedAt: department.UpdatedAt,
		DeletedAt: department.DeletedAt,
	}
}

/*
 * 検索
 *
 * ページング方針：
 * ・初回は offset=0, limit=50
 * ・さらに表示するときは、フロントで現在表示済みの件数を offset として送る
 * ・limit が未指定、0以下の場合は 50件にする
 * ・limit が 50件を超える場合も 50件に丸める
 *
 * hasMore：
 * ・総件数 total が offset + 今回取得件数 より多ければ true
 * ・それ以下なら false
 */
func (service *departmentService) SearchDepartments(req types.SearchDepartmentsRequest) results.Result {
	// ページング検索条件を共通関数で正規化する
	normalizedCondition, normalizeResult := utils.NormalizePageSearchCondition(
		utils.PageSearchCondition{
			Keyword: req.Keyword,
			Offset:  req.Offset,
			Limit:   req.Limit,
		},
		"SEARCH_DEPARTMENTS_INVALID_OFFSET",
		"検索開始位置が正しくありません",
	)
	if normalizeResult.Error {
		return normalizeResult
	}

	req.Keyword = normalizedCondition.Keyword
	req.Offset = normalizedCondition.Offset
	req.Limit = normalizedCondition.Limit

	// Builderで一覧検索用クエリと件数取得用クエリを作成する
	searchQuery, countQuery, buildResult := service.departmentBuilder.BuildSearchDepartmentsQuery(req)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryで所属一覧を取得する
	departments, findResult := service.departmentRepository.FindDepartments(searchQuery)
	if findResult.Error {
		return findResult
	}

	// Repositoryで検索条件に一致する総件数を取得する
	total, countResult := service.departmentRepository.CountDepartments(countQuery)
	if countResult.Error {
		return countResult
	}

	// DBモデルをフロント返却用Responseへ変換する
	departmentResponses := make([]types.DepartmentResponse, 0, len(departments))
	for _, department := range departments {
		departmentResponses = append(departmentResponses, toDepartmentResponse(department))
	}

	hasMore := utils.HasMore(total, req.Offset, len(departments))

	return results.OK(
		types.SearchDepartmentsResponse{
			Departments: departmentResponses,
			Total:       total,
			Offset:      req.Offset,
			Limit:       req.Limit,
			HasMore:     hasMore,
		},
		"SEARCH_DEPARTMENTS_SUCCESS",
		"所属一覧を取得しました",
		nil,
	)
}

/*
 * 詳細
 */
func (service *departmentService) GetDepartmentDetail(req types.DepartmentDetailRequest) results.Result {
	// Builderで詳細取得用クエリを作成する
	query, buildResult := service.departmentBuilder.BuildFindDepartmentByIDQuery(req.DepartmentID)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryで所属を取得する
	department, findResult := service.departmentRepository.FindDepartment(query)
	if findResult.Error {
		return findResult
	}

	return results.OK(
		types.DepartmentDetailResponse{
			Department: toDepartmentResponse(department),
		},
		"GET_DEPARTMENT_DETAIL_SUCCESS",
		"所属詳細を取得しました",
		nil,
	)
}

/*
 * 新規作成
 */
func (service *departmentService) CreateDepartment(req types.CreateDepartmentRequest) results.Result {
	// Builderで所属名重複確認用クエリを作成する
	nameCountQuery, buildNameCountResult := service.departmentBuilder.BuildCountActiveDepartmentByNameQuery(req.Name)
	if buildNameCountResult.Error {
		return buildNameCountResult
	}

	// Repositoryで所属名重複確認を実行する
	nameCount, nameCountResult := service.departmentRepository.CountDepartments(nameCountQuery)
	if nameCountResult.Error {
		return nameCountResult
	}

	if nameCount > 0 {
		return results.Conflict(
			"CREATE_DEPARTMENT_NAME_ALREADY_EXISTS",
			"この所属名は既に使用されています",
			map[string]any{
				"name": req.Name,
			},
		)
	}

	// Builderで作成用Modelを作る
	department, buildDepartmentResult := service.departmentBuilder.BuildCreateDepartmentModel(req)
	if buildDepartmentResult.Error {
		return buildDepartmentResult
	}

	// Repositoryで所属を作成する
	createdDepartment, createResult := service.departmentRepository.CreateDepartment(department)
	if createResult.Error {
		return createResult
	}

	return results.Created(
		types.CreateDepartmentResponse{
			Department: toDepartmentResponse(createdDepartment),
		},
		"CREATE_DEPARTMENT_SUCCESS",
		"所属を作成しました",
		nil,
	)
}

/*
 * 更新
 */
func (service *departmentService) UpdateDepartment(req types.UpdateDepartmentRequest) results.Result {
	// Builderで対象所属取得用クエリを作成する
	findQuery, buildFindResult := service.departmentBuilder.BuildFindDepartmentByIDQuery(req.DepartmentID)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象所属を取得する
	currentDepartment, findResult := service.departmentRepository.FindDepartment(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderで所属名重複確認用クエリを作成する
	nameCountQuery, buildNameCountResult := service.departmentBuilder.BuildCountActiveDepartmentByNameExceptIDQuery(req.Name, req.DepartmentID)
	if buildNameCountResult.Error {
		return buildNameCountResult
	}

	// Repositoryで所属名重複確認を実行する
	nameCount, nameCountResult := service.departmentRepository.CountDepartments(nameCountQuery)
	if nameCountResult.Error {
		return nameCountResult
	}

	if nameCount > 0 {
		return results.Conflict(
			"UPDATE_DEPARTMENT_NAME_ALREADY_EXISTS",
			"この所属名は既に使用されています",
			map[string]any{
				"name":         req.Name,
				"departmentId": req.DepartmentID,
			},
		)
	}

	// Builderで更新用Modelを作る
	updatedDepartment, buildUpdateResult := service.departmentBuilder.BuildUpdateDepartmentModel(
		currentDepartment,
		req,
	)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	// Repositoryで所属を更新する
	savedDepartment, saveResult := service.departmentRepository.SaveDepartment(updatedDepartment)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.UpdateDepartmentResponse{
			Department: toDepartmentResponse(savedDepartment),
		},
		"UPDATE_DEPARTMENT_SUCCESS",
		"所属を更新しました",
		nil,
	)
}

/*
 * 論理削除
 *
 * 注意：
 * ・所属に有効ユーザーが紐づいている場合は削除不可にする
 */
func (service *departmentService) DeleteDepartment(req types.DeleteDepartmentRequest) results.Result {
	// Builderで対象所属取得用クエリを作成する
	findQuery, buildFindResult := service.departmentBuilder.BuildFindDepartmentByIDQuery(req.DepartmentID)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象所属を取得する
	currentDepartment, findResult := service.departmentRepository.FindDepartment(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderで所属に紐づく有効ユーザー件数取得用クエリを作成する
	activeUserCountQuery, buildActiveUserCountResult := service.departmentBuilder.BuildCountActiveUsersByDepartmentIDQuery(req.DepartmentID)
	if buildActiveUserCountResult.Error {
		return buildActiveUserCountResult
	}

	// Repositoryで所属に紐づく有効ユーザー件数を取得する
	activeUserCount, activeUserCountResult := service.departmentRepository.CountUsers(activeUserCountQuery)
	if activeUserCountResult.Error {
		return activeUserCountResult
	}

	if activeUserCount > 0 {
		return results.Conflict(
			"DELETE_DEPARTMENT_HAS_ACTIVE_USERS",
			"この所属にユーザーが紐づいているため削除できません",
			map[string]any{
				"departmentId": req.DepartmentID,
				"userCount":    activeUserCount,
			},
		)
	}

	// Builderで論理削除用Modelを作る
	deletedDepartment, buildDeleteResult := service.departmentBuilder.BuildDeleteDepartmentModel(currentDepartment)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	// Repositoryで所属を保存する
	_, saveResult := service.departmentRepository.SaveDepartment(deletedDepartment)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteDepartmentResponse{
			DepartmentID: req.DepartmentID,
		},
		"DELETE_DEPARTMENT_SUCCESS",
		"所属を削除しました",
		nil,
	)
}
