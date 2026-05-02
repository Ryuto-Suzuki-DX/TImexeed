/*
 * 〇Controller
 * ・bind
 * ・JSON → Request → Controller → Service
 * ・Service → Builder
 * ・Service → Repository
 * ・Service → Controller
 * ・Service → Result　	（サービスからコントローラへの型)
 * ・Service → Response	 (フロントへのレスポンス)
 *
 *
 */

package controllers

import (
	"fmt"
	"strconv"

	"timexeed/backend/internal/admin/services"
	"timexeed/backend/internal/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用ユーザーController
 *
 * Controllerの責務:
 * - リクエストを受け取る
 * - パラメータを取得する
 * - JSONをリクエスト型へ変換する
 * - Serviceへ処理を渡す
 * - Serviceの結果を共通レスポンスで返す
 */
type UserController struct {
	userService services.UserServiceInterface
}

/*
 * UserControllerを生成する
 */
func NewUserController(userService services.UserServiceInterface) *UserController {
	return &UserController{
		userService: userService,
	}
}

/*
 * ユーザー一覧取得
 * GET /admin/users
 */
func (uc *UserController) SearchUsers(c *gin.Context) {
	req := types.SearchUsersRequest{
		Keyword:        c.Query("keyword"),
		IncludeDeleted: c.Query("includeDeleted") == "true",
	}

	result := uc.userService.SearchUsers(req)

	responses.JSON(c, result)
}

/*
 * ユーザー詳細取得
 * GET /admin/users/:id
 */
func (uc *UserController) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		result := uc.userService.InvalidUserID()
		responses.JSON(c, result)
		return
	}

	result := uc.userService.GetUser(uint(id))

	responses.JSON(c, result)
}

/*
 * ユーザー新規作成
 * POST /admin/users
 */
func (uc *UserController) CreateUser(c *gin.Context) {
	var req types.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		result := uc.userService.InvalidRequest()
		responses.JSON(c, result)
		return
	}

	result := uc.userService.CreateUser(req)

	responses.JSON(c, result)
}

/*
 * ユーザー更新
 * PUT /admin/users/:id
 */
func (uc *UserController) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		result := uc.userService.InvalidUserID()
		responses.JSON(c, result)
		return
	}

	var req types.UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		result := uc.userService.InvalidRequest()
		responses.JSON(c, result)
		return
	}

	result := uc.userService.UpdateUser(uint(id), req)

	responses.JSON(c, result)
}

/*
 * ユーザー論理削除
 * DELETE /admin/users/:id
 */
func (uc *UserController) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		result := uc.userService.InvalidUserID()
		responses.JSON(c, result)
		return
	}

	loginUserID, err := getLoginUserID(c)
	if err != nil {
		result := results.Unauthorized("ログイン情報を取得できませんでした")
		responses.JSON(c, result)
		return
	}

	result := uc.userService.DeleteUser(uint(id), loginUserID)

	responses.JSON(c, result)
}

/*
 * Contextからログイン中ユーザーIDを取得する
 *
 * AuthMiddlewareで c.Set("userId", ...) されている前提
 */
func getLoginUserID(c *gin.Context) (uint, error) {
	value, exists := c.Get("userId")
	if !exists {
		return 0, fmt.Errorf("userId not found")
	}

	switch v := value.(type) {
	case uint:
		return v, nil
	case int:
		return uint(v), nil
	case float64:
		return uint(v), nil
	default:
		return 0, fmt.Errorf("invalid userId type")
	}
}
