/*
 * 〇Controller
 * ・bind
 * ・JSON → Request → Controller → Service
 * ・Service → Builder
 * ・Service → Repository
 * ・Service → Controller
 * ・Service → Result   （サービスからコントローラへの型)
 * ・Service → Response (フロントへのレスポンス)
 */

package controllers

import (
	"strconv"

	"timexeed/backend/internal/admin/services"
	"timexeed/backend/internal/admin/types"
	"timexeed/backend/internal/responses"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用所属Controller
 *
 * Controllerの責務:
 * - リクエストを受け取る
 * - パラメータを取得する
 * - JSONをリクエスト型へ変換する
 * - Serviceへ処理を渡す
 * - Serviceの結果を共通レスポンスで返す
 */
type DepartmentController struct {
	departmentService services.DepartmentServiceInterface
}

/*
 * DepartmentControllerを生成する
 */
func NewDepartmentController(departmentService services.DepartmentServiceInterface) *DepartmentController {
	return &DepartmentController{
		departmentService: departmentService,
	}
}

/*
 * 所属一覧取得
 * GET /admin/departments
 */
func (dc *DepartmentController) SearchDepartments(c *gin.Context) {
	req := types.SearchDepartmentsRequest{
		Keyword:        c.Query("keyword"),
		IncludeDeleted: c.Query("includeDeleted") == "true",
	}

	result := dc.departmentService.SearchDepartments(req)

	responses.JSON(c, result)
}

/*
 * 所属詳細取得
 * GET /admin/departments/:id
 */
func (dc *DepartmentController) GetDepartment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		result := dc.departmentService.InvalidDepartmentID()
		responses.JSON(c, result)
		return
	}

	result := dc.departmentService.GetDepartment(uint(id))

	responses.JSON(c, result)
}

/*
 * 所属新規作成
 * POST /admin/departments
 */
func (dc *DepartmentController) CreateDepartment(c *gin.Context) {
	var req types.CreateDepartmentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		result := dc.departmentService.InvalidRequest()
		responses.JSON(c, result)
		return
	}

	result := dc.departmentService.CreateDepartment(req)

	responses.JSON(c, result)
}

/*
 * 所属更新
 * PUT /admin/departments/:id
 */
func (dc *DepartmentController) UpdateDepartment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		result := dc.departmentService.InvalidDepartmentID()
		responses.JSON(c, result)
		return
	}

	var req types.UpdateDepartmentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		result := dc.departmentService.InvalidRequest()
		responses.JSON(c, result)
		return
	}

	result := dc.departmentService.UpdateDepartment(uint(id), req)

	responses.JSON(c, result)
}

/*
 * 所属論理削除
 * DELETE /admin/departments/:id
 */
func (dc *DepartmentController) DeleteDepartment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		result := dc.departmentService.InvalidDepartmentID()
		responses.JSON(c, result)
		return
	}

	result := dc.departmentService.DeleteDepartment(uint(id))

	responses.JSON(c, result)
}
