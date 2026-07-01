package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用日別交通費Controller
 *
 * 公開APIでは本人分の検索だけを扱う。
 * 保存はmonthly_attendances/updateからServiceを内部的に呼び出す。
 */
type AttendanceTransportExpenseController struct {
	attendanceTransportExpenseService services.AttendanceTransportExpenseService
}

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
 * POST /user/attendance-transport-expenses/search
 */
func (controller *AttendanceTransportExpenseController) SearchAttendanceTransportExpenses(
	c *gin.Context,
) {
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok || userID == 0 {
		responses.JSON(c, results.Unauthorized(
			"SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	var req types.SearchAttendanceTransportExpensesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_REQUEST",
			"日別交通費検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.attendanceTransportExpenseService.
		SearchAttendanceTransportExpenses(userID, req)

	responses.JSON(c, result)
}
