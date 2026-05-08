package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用所属Controller
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * 注意：
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・Requestを別の型へ詰め直さない
 * ・c.JSONは直接使わず responses.JSON を使う
 *
 * エラー方針：
 * ・Controllerで発生したエラーはControllerでcode/messageを決める
 * ・Serviceで発生したエラーはServiceでcode/messageを決める
 * ・Builderで発生したエラーはBuilderでcode/messageを決める
 * ・Repositoryで発生したエラーはRepositoryでcode/messageを決める
 * ・Controllerは最終的に responses.JSON で返す
 */
type DepartmentController struct {
	departmentService services.DepartmentService
}

/*
 * DepartmentController生成
 */
func NewDepartmentController(departmentService services.DepartmentService) *DepartmentController {
	return &DepartmentController{
		departmentService: departmentService,
	}
}

/*
 * 検索
 *
 * POST /admin/departments/search
 */
func (controller *DepartmentController) SearchDepartments(c *gin.Context) {
	var req types.SearchDepartmentsRequest

	// リクエストJSONをSearchDepartmentsRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"SEARCH_DEPARTMENTS_INVALID_REQUEST",
			"所属検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.departmentService.SearchDepartments(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 取得
 *
 * POST /admin/departments/detail
 */
func (controller *DepartmentController) GetDepartmentDetail(c *gin.Context) {
	var req types.DepartmentDetailRequest

	// リクエストJSONをDepartmentDetailRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"GET_DEPARTMENT_DETAIL_INVALID_REQUEST",
			"所属詳細取得のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.departmentService.GetDepartmentDetail(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 新規作成
 *
 * POST /admin/departments/create
 */
func (controller *DepartmentController) CreateDepartment(c *gin.Context) {
	var req types.CreateDepartmentRequest

	// リクエストJSONをCreateDepartmentRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"CREATE_DEPARTMENT_INVALID_REQUEST",
			"所属作成のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.departmentService.CreateDepartment(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 所属更新
 *
 * POST /admin/departments/update
 */
func (controller *DepartmentController) UpdateDepartment(c *gin.Context) {
	var req types.UpdateDepartmentRequest

	// リクエストJSONをUpdateDepartmentRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"UPDATE_DEPARTMENT_INVALID_REQUEST",
			"所属更新のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.departmentService.UpdateDepartment(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 論理削除
 *
 * POST /admin/departments/delete
 */
func (controller *DepartmentController) DeleteDepartment(c *gin.Context) {
	var req types.DeleteDepartmentRequest

	// リクエストJSONをDeleteDepartmentRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"DELETE_DEPARTMENT_INVALID_REQUEST",
			"所属削除のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.departmentService.DeleteDepartment(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
