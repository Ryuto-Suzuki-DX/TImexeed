package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用お知らせController
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * 注意：
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・Requestを別の型へ詰め直さない
 * ・c.JSONは直接使わず responses.JSON を使う
 * ・従業員APIでは userId / targetUserId を request body で受け取らない
 *
 * エラー方針：
 * ・Controllerで発生したエラーはControllerでcode/messageを決める
 * ・Serviceで発生したエラーはServiceでcode/messageを決める
 * ・Builderで発生したエラーはBuilderでcode/messageを決める
 * ・Repositoryで発生したエラーはRepositoryでcode/messageを決める
 * ・Controllerは最終的に responses.JSON で返す
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
 * お知らせ検索
 *
 * POST /user/notifications/search
 *
 * 用途：
 * ・従業員本人のお知らせ一覧を取得する
 * ・お知らせ画面に表示する
 *
 * 仕様：
 * ・フロントから userId / targetUserId は送らない
 * ・AuthMiddlewareでJWTから取得した userId を使う
 * ・ログイン中ユーザー本人のお知らせだけを取得する
 */
func (controller *NotificationController) SearchNotifications(c *gin.Context) {
	var req types.SearchNotificationsRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"SEARCH_NOTIFICATIONS_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	loginUserID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"SEARCH_NOTIFICATIONS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをSearchNotificationsRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"SEARCH_NOTIFICATIONS_INVALID_REQUEST",
			"お知らせ検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.notificationService.SearchNotifications(loginUserID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * お知らせ既読更新
 *
 * POST /user/notifications/read
 *
 * 用途：
 * ・従業員本人のお知らせを既読にする
 *
 * 仕様：
 * ・フロントから userId / targetUserId は送らない
 * ・AuthMiddlewareでJWTから取得した userId を使う
 * ・Service側で loginUserID + notificationId から対象お知らせを特定する
 * ・対象お知らせが存在すれば既読にする
 */
func (controller *NotificationController) ReadNotification(c *gin.Context) {
	var req types.ReadNotificationRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"READ_NOTIFICATION_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	loginUserID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"READ_NOTIFICATION_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをReadNotificationRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"READ_NOTIFICATION_INVALID_REQUEST",
			"お知らせ既読更新のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.notificationService.ReadNotification(loginUserID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}

/*
 * 未読お知らせ件数取得
 *
 * POST /user/notifications/unread-count
 *
 * 用途：
 * ・従業員本人の未読お知らせ件数を取得する
 * ・サイドメニューに NEW! を表示するために使う
 *
 * 仕様：
 * ・フロントから userId / targetUserId は送らない
 * ・AuthMiddlewareでJWTから取得した userId を使う
 * ・ログイン中ユーザー本人の未読お知らせ件数だけを取得する
 */
func (controller *NotificationController) CountUnreadNotifications(c *gin.Context) {
	var req types.CountUnreadNotificationsRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"COUNT_UNREAD_NOTIFICATIONS_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	loginUserID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"COUNT_UNREAD_NOTIFICATIONS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをCountUnreadNotificationsRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"COUNT_UNREAD_NOTIFICATIONS_INVALID_REQUEST",
			"未読お知らせ件数取得のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.notificationService.CountUnreadNotifications(loginUserID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
