package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
)

/*
 * 管理者用お知らせService interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・管理者本人宛のお知らせ検索、既読更新、未読件数取得では userId / targetUserId をRequestで受け取らない
 * ・ControllerでAuthMiddleware由来のuserIdを取得し、Serviceへ渡す
 * ・全員宛作成では、有効なADMIN/USER全員にnotificationsを作成する
 */
type NotificationService interface {
	SearchNotifications(userID uint, req types.SearchNotificationsRequest) results.Result
	ReadNotification(userID uint, req types.ReadNotificationRequest) results.Result
	CountUnreadNotifications(userID uint, req types.CountUnreadNotificationsRequest) results.Result
	CreateNotificationForAllUsers(req types.CreateNotificationForAllUsersRequest) results.Result
	DeleteNotification(req types.DeleteNotificationRequest) results.Result
}

/*
 * 管理者用お知らせService
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや保存用Modelを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB処理を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type notificationService struct {
	notificationBuilder    builders.NotificationBuilder
	notificationRepository repositories.NotificationRepository
}

/*
 * NotificationService生成
 */
func NewNotificationService(
	notificationBuilder builders.NotificationBuilder,
	notificationRepository repositories.NotificationRepository,
) *notificationService {
	return &notificationService{
		notificationBuilder:    notificationBuilder,
		notificationRepository: notificationRepository,
	}
}

/*
 * models.Notificationをフロント返却用NotificationResponseへ変換する
 */
func toNotificationResponse(notification models.Notification) types.NotificationResponse {
	return types.NotificationResponse{
		ID:        notification.ID,
		Title:     notification.Title,
		Message:   notification.Message,
		IsRead:    notification.IsRead,
		ReadAt:    notification.ReadAt,
		CreatedAt: notification.CreatedAt,
	}
}

/*
 * お知らせ検索
 *
 * ログイン中管理者本人のお知らせ一覧を取得する。
 */
func (service *notificationService) SearchNotifications(
	userID uint,
	req types.SearchNotificationsRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"SEARCH_NOTIFICATIONS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	if req.Offset < 0 {
		return results.BadRequest(
			"SEARCH_NOTIFICATIONS_INVALID_OFFSET",
			"お知らせ検索の開始位置が正しくありません",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	// hasMore判定用に1件多く取得する
	searchLimit := req.Limit + 1

	// Builderでお知らせ検索用クエリを作成する
	query, buildResult := service.notificationBuilder.BuildSearchNotificationsQuery(userID, searchLimit, req.Offset)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryでお知らせ一覧を取得する
	notifications, findResult := service.notificationRepository.FindNotifications(query)
	if findResult.Error {
		return findResult
	}

	hasMore := false

	if len(notifications) > req.Limit {
		hasMore = true
		notifications = notifications[:req.Limit]
	}

	// DBモデルをフロント返却用Responseへ変換する
	notificationResponses := make([]types.NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		notificationResponses = append(notificationResponses, toNotificationResponse(notification))
	}

	return results.OK(
		types.SearchNotificationsResponse{
			Notifications: notificationResponses,
			HasMore:       hasMore,
		},
		"SEARCH_NOTIFICATIONS_SUCCESS",
		"お知らせ一覧を取得しました",
		nil,
	)
}

/*
 * お知らせ既読更新
 *
 * userID + notificationID で対象お知らせを取得し、既読にする。
 */
func (service *notificationService) ReadNotification(
	userID uint,
	req types.ReadNotificationRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"READ_NOTIFICATION_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	if req.NotificationID == 0 {
		return results.BadRequest(
			"READ_NOTIFICATION_INVALID_NOTIFICATION_ID",
			"お知らせIDが正しくありません",
			map[string]any{
				"notificationId": req.NotificationID,
			},
		)
	}

	// Builderで対象お知らせ取得用クエリを作成する
	findQuery, buildFindResult := service.notificationBuilder.BuildFindNotificationByUserIDAndIDQuery(userID, req.NotificationID)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象お知らせを取得する
	currentNotification, findResult := service.notificationRepository.FindNotification(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderで既読更新用Modelを作る
	readNotification, buildReadResult := service.notificationBuilder.BuildReadNotificationModel(currentNotification)
	if buildReadResult.Error {
		return buildReadResult
	}

	// Repositoryでお知らせを保存する
	savedNotification, saveResult := service.notificationRepository.SaveNotification(readNotification)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.ReadNotificationResponse{
			Notification: toNotificationResponse(savedNotification),
		},
		"READ_NOTIFICATION_SUCCESS",
		"お知らせを既読にしました",
		nil,
	)
}

/*
 * 未読お知らせ件数取得
 *
 * ログイン中管理者本人の未読お知らせ件数を取得する。
 */
func (service *notificationService) CountUnreadNotifications(
	userID uint,
	req types.CountUnreadNotificationsRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"COUNT_UNREAD_NOTIFICATIONS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	// Builderで未読お知らせ件数取得用クエリを作成する
	query, buildResult := service.notificationBuilder.BuildCountUnreadNotificationsQuery(userID)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryで未読お知らせ件数を取得する
	unreadCount, countResult := service.notificationRepository.CountNotifications(query)
	if countResult.Error {
		return countResult
	}

	return results.OK(
		types.CountUnreadNotificationsResponse{
			UnreadCount: unreadCount,
		},
		"COUNT_UNREAD_NOTIFICATIONS_SUCCESS",
		"未読お知らせ件数を取得しました",
		nil,
	)
}

/*
 * 全員宛お知らせ作成
 *
 * is_deleted = false の全アカウントへ同じタイトル・本文のお知らせを作成する。
 *
 * 注意：
 * ・USERだけでなくADMINも対象に含める
 * ・通知は未読状態で作成する
 */
func (service *notificationService) CreateNotificationForAllUsers(
	req types.CreateNotificationForAllUsersRequest,
) results.Result {
	if req.Title == "" {
		return results.BadRequest(
			"CREATE_NOTIFICATION_FOR_ALL_USERS_EMPTY_TITLE",
			"お知らせタイトルを入力してください",
			nil,
		)
	}

	if req.Message == "" {
		return results.BadRequest(
			"CREATE_NOTIFICATION_FOR_ALL_USERS_EMPTY_MESSAGE",
			"お知らせ本文を入力してください",
			nil,
		)
	}

	// Repositoryで有効ユーザー一覧を取得する
	users, findUsersResult := service.notificationRepository.FindActiveUsers()
	if findUsersResult.Error {
		return findUsersResult
	}

	// Builderで全員宛お知らせ作成用Model配列を作る
	notifications, buildResult := service.notificationBuilder.BuildCreateNotificationsForAllUsersModels(users, req)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryでお知らせを一括作成する
	createdNotifications, createResult := service.notificationRepository.CreateNotifications(notifications)
	if createResult.Error {
		return createResult
	}

	return results.Created(
		types.CreateNotificationForAllUsersResponse{
			CreatedCount: len(createdNotifications),
		},
		"CREATE_NOTIFICATION_FOR_ALL_USERS_SUCCESS",
		"全員宛のお知らせを作成しました",
		nil,
	)
}

/*
 * お知らせ削除
 *
 * 管理者がお知らせを論理削除する。
 */
func (service *notificationService) DeleteNotification(
	req types.DeleteNotificationRequest,
) results.Result {
	if req.NotificationID == 0 {
		return results.BadRequest(
			"DELETE_NOTIFICATION_INVALID_NOTIFICATION_ID",
			"お知らせIDが正しくありません",
			map[string]any{
				"notificationId": req.NotificationID,
			},
		)
	}

	// Builderで対象お知らせ取得用クエリを作成する
	findQuery, buildFindResult := service.notificationBuilder.BuildFindNotificationByIDQuery(req.NotificationID)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象お知らせを取得する
	currentNotification, findResult := service.notificationRepository.FindNotification(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderで論理削除用Modelを作る
	deleteNotification, buildDeleteResult := service.notificationBuilder.BuildDeleteNotificationModel(currentNotification)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	// Repositoryでお知らせを保存する
	savedNotification, saveResult := service.notificationRepository.SaveNotification(deleteNotification)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteNotificationResponse{
			Notification: toNotificationResponse(savedNotification),
		},
		"DELETE_NOTIFICATION_SUCCESS",
		"お知らせを削除しました",
		nil,
	)
}
