package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用お知らせ自動リマインドController
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
type NotificationReminderController struct {
	notificationReminderService services.NotificationReminderService
}

/*
 * NotificationReminderController生成
 */
func NewNotificationReminderController(
	notificationReminderService services.NotificationReminderService,
) *NotificationReminderController {
	return &NotificationReminderController{
		notificationReminderService: notificationReminderService,
	}
}

/*
 * 自動リマインド検索
 *
 * POST /admin/notification-reminders/search
 *
 * 用途：
 * ・管理者画面で自動リマインド設定一覧を取得する
 */
func (controller *NotificationReminderController) SearchNotificationReminders(c *gin.Context) {
	var req types.SearchNotificationRemindersRequest

	// リクエストJSONをSearchNotificationRemindersRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_NOTIFICATION_REMINDERS_INVALID_REQUEST",
			"自動リマインド検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をServiceへ渡す
	result := controller.notificationReminderService.SearchNotificationReminders(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 自動リマインド作成
 *
 * POST /admin/notification-reminders/create
 *
 * 用途：
 * ・管理者が自動リマインド設定を作成する
 */
func (controller *NotificationReminderController) CreateNotificationReminder(c *gin.Context) {
	var req types.CreateNotificationReminderRequest

	// リクエストJSONをCreateNotificationReminderRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"CREATE_NOTIFICATION_REMINDER_INVALID_REQUEST",
			"自動リマインド作成のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をServiceへ渡す
	result := controller.notificationReminderService.CreateNotificationReminder(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 自動リマインド更新
 *
 * POST /admin/notification-reminders/update
 *
 * 用途：
 * ・管理者が自動リマインド設定を更新する
 */
func (controller *NotificationReminderController) UpdateNotificationReminder(c *gin.Context) {
	var req types.UpdateNotificationReminderRequest

	// リクエストJSONをUpdateNotificationReminderRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"UPDATE_NOTIFICATION_REMINDER_INVALID_REQUEST",
			"自動リマインド更新のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をServiceへ渡す
	result := controller.notificationReminderService.UpdateNotificationReminder(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 自動リマインド削除
 *
 * POST /admin/notification-reminders/delete
 *
 * 用途：
 * ・管理者が自動リマインド設定を論理削除する
 */
func (controller *NotificationReminderController) DeleteNotificationReminder(c *gin.Context) {
	var req types.DeleteNotificationReminderRequest

	// リクエストJSONをDeleteNotificationReminderRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"DELETE_NOTIFICATION_REMINDER_INVALID_REQUEST",
			"自動リマインド削除のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をServiceへ渡す
	result := controller.notificationReminderService.DeleteNotificationReminder(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 自動リマインド有効/無効切替
 *
 * POST /admin/notification-reminders/toggle-enabled
 *
 * 用途：
 * ・管理者が自動リマインド設定の有効/無効を切り替える
 */
func (controller *NotificationReminderController) ToggleNotificationReminderEnabled(c *gin.Context) {
	var req types.ToggleNotificationReminderEnabledRequest

	// リクエストJSONをToggleNotificationReminderEnabledRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"TOGGLE_NOTIFICATION_REMINDER_ENABLED_INVALID_REQUEST",
			"自動リマインド有効状態更新のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をServiceへ渡す
	result := controller.notificationReminderService.ToggleNotificationReminderEnabled(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
