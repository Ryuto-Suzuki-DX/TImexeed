package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用ユーザー給与詳細Controller
 *
 * 役割：
 * ・JSONのbindを行う
 * ・bindに成功したRequestをServiceへ渡す
 * ・Serviceから返ってきたResultを共通レスポンスで返す
 *
 * 注意：
 * ・Controllerでは業務ロジックを持たない
 * ・ControllerではDBに触らない
 * ・ControllerではResponse型への変換をしない
 */
type UserSalaryDetailController struct {
	userSalaryDetailService services.UserSalaryDetailService
}

/*
 * UserSalaryDetailController生成
 */
func NewUserSalaryDetailController(
	userSalaryDetailService services.UserSalaryDetailService,
) *UserSalaryDetailController {
	return &UserSalaryDetailController{
		userSalaryDetailService: userSalaryDetailService,
	}
}

/*
 * ユーザー給与詳細検索
 *
 * POST /admin/user-salary-details/search
 */
func (controller *UserSalaryDetailController) SearchUserSalaryDetails(c *gin.Context) {
	var req types.SearchUserSalaryDetailsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(
			c,
			results.BadRequest(
				"SEARCH_USER_SALARY_DETAILS_INVALID_REQUEST",
				"ユーザー給与詳細検索のリクエスト形式が正しくありません",
				err.Error(),
			),
		)
		return
	}

	result := controller.userSalaryDetailService.SearchUserSalaryDetails(req)
	responses.JSON(c, result)
}

/*
 * ユーザー給与詳細単体情報取得
 *
 * POST /admin/user-salary-details/get
 */
func (controller *UserSalaryDetailController) GetUserSalaryDetail(c *gin.Context) {
	var req types.GetUserSalaryDetailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(
			c,
			results.BadRequest(
				"GET_USER_SALARY_DETAIL_INVALID_REQUEST",
				"ユーザー給与詳細取得のリクエスト形式が正しくありません",
				err.Error(),
			),
		)
		return
	}

	result := controller.userSalaryDetailService.GetUserSalaryDetail(req)
	responses.JSON(c, result)
}

/*
 * ユーザー給与詳細新規作成
 *
 * POST /admin/user-salary-details/create
 */
func (controller *UserSalaryDetailController) CreateUserSalaryDetail(c *gin.Context) {
	var req types.CreateUserSalaryDetailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(
			c,
			results.BadRequest(
				"CREATE_USER_SALARY_DETAIL_INVALID_REQUEST",
				"ユーザー給与詳細作成のリクエスト形式が正しくありません",
				err.Error(),
			),
		)
		return
	}

	result := controller.userSalaryDetailService.CreateUserSalaryDetail(req)
	responses.JSON(c, result)
}

/*
 * ユーザー給与詳細更新
 *
 * POST /admin/user-salary-details/update
 */
func (controller *UserSalaryDetailController) UpdateUserSalaryDetail(c *gin.Context) {
	var req types.UpdateUserSalaryDetailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(
			c,
			results.BadRequest(
				"UPDATE_USER_SALARY_DETAIL_INVALID_REQUEST",
				"ユーザー給与詳細更新のリクエスト形式が正しくありません",
				err.Error(),
			),
		)
		return
	}

	result := controller.userSalaryDetailService.UpdateUserSalaryDetail(req)
	responses.JSON(c, result)
}

/*
 * ユーザー給与詳細論理削除
 *
 * POST /admin/user-salary-details/delete
 */
func (controller *UserSalaryDetailController) DeleteUserSalaryDetail(c *gin.Context) {
	var req types.DeleteUserSalaryDetailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(
			c,
			results.BadRequest(
				"DELETE_USER_SALARY_DETAIL_INVALID_REQUEST",
				"ユーザー給与詳細削除のリクエスト形式が正しくありません",
				err.Error(),
			),
		)
		return
	}

	result := controller.userSalaryDetailService.DeleteUserSalaryDetail(req)
	responses.JSON(c, result)
}
