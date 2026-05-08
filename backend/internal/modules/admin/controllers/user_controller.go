package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用ユーザーController
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
type UserController struct {
	userService services.UserService
}

/*
 * UserController生成
 */
func NewUserController(userService services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

/*
 * 検索
 *
 * POST /admin/users/search
 */
func (controller *UserController) SearchUsers(c *gin.Context) {
	var req types.SearchUsersRequest

	// リクエストJSONをSearchUsersRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"SEARCH_USERS_INVALID_REQUEST",
			"ユーザー検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.userService.SearchUsers(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 詳細
 *
 * POST /admin/users/detail
 */
func (controller *UserController) GetUserDetail(c *gin.Context) {
	var req types.UserDetailRequest

	// リクエストJSONをUserDetailRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"GET_USER_DETAIL_INVALID_REQUEST",
			"ユーザー詳細取得のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.userService.GetUserDetail(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 新規作成
 *
 * POST /admin/users/create
 */
func (controller *UserController) CreateUser(c *gin.Context) {
	var req types.CreateUserRequest

	// リクエストJSONをCreateUserRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"CREATE_USER_INVALID_REQUEST",
			"ユーザー作成のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.userService.CreateUser(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * ユーザー更新
 *
 * POST /admin/users/update
 */
func (controller *UserController) UpdateUser(c *gin.Context) {
	var req types.UpdateUserRequest

	// リクエストJSONをUpdateUserRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"UPDATE_USER_INVALID_REQUEST",
			"ユーザー更新のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.userService.UpdateUser(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 論理削除
 *
 * POST /admin/users/delete
 */
func (controller *UserController) DeleteUser(c *gin.Context) {
	var req types.DeleteUserRequest

	// リクエストJSONをDeleteUserRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"DELETE_USER_INVALID_REQUEST",
			"ユーザー削除のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.userService.DeleteUser(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
