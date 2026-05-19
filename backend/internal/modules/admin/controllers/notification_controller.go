package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用お知らせController
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・ログイン中管理者IDをgin.Contextから取得する
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 */
type NotificationController struct {
	notificationService services.NotificationService
}

/*
 * NotificationController生成
 */
func NewNotificationController(notificationService services.NotificationService) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
	}
}

/*
 * ログイン中管理者ID取得
 *
 * 注意：
 * ・controllers package内で getLoginAdminID という名前は他Controllerと衝突しやすい
 * ・そのため、お知らせController専用名にしている
 */
func getNotificationLoginAdminID(c *gin.Context, actionCode string) (uint, results.Result) {
	userID := c.GetUint("userId")
	if userID == 0 {
		return 0, results.Unauthorized(
			actionCode+"_INVALID_LOGIN_ADMIN_ID",
			"認証情報の管理者IDが正しくありません",
			nil,
		)
	}

	return userID, results.OK(
		nil,
		actionCode+"_VALID_LOGIN_ADMIN_ID",
		"",
		nil,
	)
}

/*
 * お知らせ検索
 *
 * POST /admin/notifications/search
 */
func (controller *NotificationController) SearchNotifications(c *gin.Context) {
	loginAdminID, adminIDResult := getNotificationLoginAdminID(c, "SEARCH_NOTIFICATIONS")
	if adminIDResult.Error {
		responses.JSON(c, adminIDResult)
		return
	}

	var req types.SearchNotificationsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_NOTIFICATIONS_INVALID_REQUEST",
			"お知らせ検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.notificationService.SearchNotifications(loginAdminID, req)

	responses.JSON(c, result)
}

/*
 * お知らせ既読更新
 *
 * POST /admin/notifications/read
 */
func (controller *NotificationController) ReadNotification(c *gin.Context) {
	loginAdminID, adminIDResult := getNotificationLoginAdminID(c, "READ_NOTIFICATION")
	if adminIDResult.Error {
		responses.JSON(c, adminIDResult)
		return
	}

	var req types.ReadNotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"READ_NOTIFICATION_INVALID_REQUEST",
			"お知らせ既読更新のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.notificationService.ReadNotification(loginAdminID, req)

	responses.JSON(c, result)
}

/*
 * 未読お知らせ件数取得
 *
 * POST /admin/notifications/unread-count
 */
func (controller *NotificationController) CountUnreadNotifications(c *gin.Context) {
	loginAdminID, adminIDResult := getNotificationLoginAdminID(c, "COUNT_UNREAD_NOTIFICATIONS")
	if adminIDResult.Error {
		responses.JSON(c, adminIDResult)
		return
	}

	var req types.CountUnreadNotificationsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"COUNT_UNREAD_NOTIFICATIONS_INVALID_REQUEST",
			"未読お知らせ件数取得のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.notificationService.CountUnreadNotifications(loginAdminID, req)

	responses.JSON(c, result)
}

/*
 * 全員宛お知らせ作成
 *
 * POST /admin/notifications/create-for-all-users
 */
func (controller *NotificationController) CreateNotificationForAllUsers(c *gin.Context) {
	var req types.CreateNotificationForAllUsersRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"CREATE_NOTIFICATION_FOR_ALL_USERS_INVALID_REQUEST",
			"全員宛お知らせ作成のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.notificationService.CreateNotificationForAllUsers(req)

	responses.JSON(c, result)
}

/*
 * お知らせ論理削除
 *
 * POST /admin/notifications/delete
 */
func (controller *NotificationController) DeleteNotification(c *gin.Context) {
	var req types.DeleteNotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"DELETE_NOTIFICATION_INVALID_REQUEST",
			"お知らせ削除のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.notificationService.DeleteNotification(req)

	responses.JSON(c, result)
}
