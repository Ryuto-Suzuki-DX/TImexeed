package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用日別交通費Controller
 *
 * 公開APIでは検索だけを扱う。
 *
 * 保存はmonthly_attendances/updateの月次全体保存から
 * AttendanceTransportExpenseServiceを内部的に呼び出す。
 */
type AttendanceTransportExpenseController struct {
	attendanceTransportExpenseService services.AttendanceTransportExpenseService
}

/*
 * AttendanceTransportExpenseController生成
 */
func NewAttendanceTransportExpenseController(
	attendanceTransportExpenseService services.AttendanceTransportExpenseService,
) *AttendanceTransportExpenseController {
	return &AttendanceTransportExpenseController{
		attendanceTransportExpenseService: attendanceTransportExpenseService,
	}
}

/*
 * 日別交通費検索
 *
 * POST /admin/attendance-transport-expenses/search
 */
func (controller *AttendanceTransportExpenseController) SearchAttendanceTransportExpenses(
	c *gin.Context,
) {
	var req types.SearchAttendanceTransportExpensesRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_REQUEST",
			"日別交通費検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result :=
		controller.attendanceTransportExpenseService.
			SearchAttendanceTransportExpenses(req)

	responses.JSON(c, result)
}
