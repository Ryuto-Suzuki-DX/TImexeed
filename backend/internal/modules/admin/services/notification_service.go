package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 管理者用お知らせService interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type NotificationService interface {
	SearchNotifications(userID uint, req types.SearchNotificationsRequest) results.Result
	ReadNotification(userID uint, req types.ReadNotificationRequest) results.Result
	CountUnreadNotifications(userID uint, req types.CountUnreadNotificationsRequest) results.Result
	CreateNotificationForAllUsers(req types.CreateNotificationForAllUsersRequest) results.Result
	CreateNotificationForUser(userID uint, title string, message string) results.Result
	DeleteNotification(req types.DeleteNotificationRequest) results.Result
}

/*
 * 管理者用お知らせService
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・ログイン中管理者本人のお知らせ検索、既読、未読件数取得を行う
 * ・管理者による全員宛お知らせ作成、削除を行う
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
		ID: notification.ID,

		UserID: notification.UserID,

		Title:   notification.Title,
		Message: notification.Message,

		IsRead: notification.IsRead,
		ReadAt: notification.ReadAt,

		IsDeleted: notification.IsDeleted,
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
		DeletedAt: notification.DeletedAt,
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
	normalizedCondition, normalizeResult := utils.NormalizePageSearchCondition(
		utils.PageSearchCondition{
			Keyword: req.Keyword,
			Offset:  req.Offset,
			Limit:   req.Limit,
		},
		"SEARCH_NOTIFICATIONS_INVALID_OFFSET",
		"検索開始位置が正しくありません",
	)
	if normalizeResult.Error {
		return normalizeResult
	}

	req.Keyword = normalizedCondition.Keyword
	req.Offset = normalizedCondition.Offset
	req.Limit = normalizedCondition.Limit

	searchQuery, buildSearchResult := service.notificationBuilder.BuildSearchNotificationsQuery(userID, req)
	if buildSearchResult.Error {
		return buildSearchResult
	}

	notifications, findResult := service.notificationRepository.FindNotifications(searchQuery)
	if findResult.Error {
		return findResult
	}

	countQuery, buildCountResult := service.notificationBuilder.BuildCountSearchNotificationsQuery(userID, req)
	if buildCountResult.Error {
		return buildCountResult
	}

	total, countResult := service.notificationRepository.CountNotifications(countQuery)
	if countResult.Error {
		return countResult
	}

	notificationResponses := make([]types.NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		notificationResponses = append(notificationResponses, toNotificationResponse(notification))
	}

	hasMore := utils.HasMore(total, req.Offset, len(notifications))

	return results.OK(
		types.SearchNotificationsResponse{
			Notifications: notificationResponses,
			Total:         total,
			Offset:        req.Offset,
			Limit:         req.Limit,
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
 * ログイン中管理者本人のお知らせだけを既読にする。
 */
func (service *notificationService) ReadNotification(
	userID uint,
	req types.ReadNotificationRequest,
) results.Result {
	findQuery, buildFindResult := service.notificationBuilder.BuildFindNotificationByUserIDAndIDQuery(userID, req.NotificationID)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentNotification, findResult := service.notificationRepository.FindNotification(findQuery)
	if findResult.Error {
		return findResult
	}

	readNotification, buildReadResult := service.notificationBuilder.BuildReadNotificationModel(currentNotification)
	if buildReadResult.Error {
		return buildReadResult
	}

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
 * ログイン中管理者本人の未読件数を取得する。
 */
func (service *notificationService) CountUnreadNotifications(
	userID uint,
	req types.CountUnreadNotificationsRequest,
) results.Result {
	query, buildResult := service.notificationBuilder.BuildCountUnreadNotificationsQuery(userID)
	if buildResult.Error {
		return buildResult
	}

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
 * 注意：
 * ・USERだけでなくADMINも対象に含める
 */
func (service *notificationService) CreateNotificationForAllUsers(
	req types.CreateNotificationForAllUsersRequest,
) results.Result {
	users, findUsersResult := service.notificationRepository.FindActiveUsers()
	if findUsersResult.Error {
		return findUsersResult
	}

	notifications, buildResult := service.notificationBuilder.BuildCreateNotificationsForAllUsersModels(users, req)
	if buildResult.Error {
		return buildResult
	}

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
 * 個別ユーザー宛お知らせ作成
 *
 * 月次勤怠承認・否認など、内部処理から特定ユーザーへ通知するときに使う。
 */
func (service *notificationService) CreateNotificationForUser(
	userID uint,
	title string,
	message string,
) results.Result {
	notification, buildResult := service.notificationBuilder.BuildCreateNotificationForUserModel(
		userID,
		title,
		message,
	)
	if buildResult.Error {
		return buildResult
	}

	createdNotifications, createResult := service.notificationRepository.CreateNotifications(
		[]models.Notification{notification},
	)
	if createResult.Error {
		return createResult
	}

	return results.Created(
		map[string]any{
			"createdCount": len(createdNotifications),
			"userId":       userID,
		},
		"CREATE_NOTIFICATION_FOR_USER_SUCCESS",
		"ユーザー宛のお知らせを作成しました",
		nil,
	)
}

/*
 * お知らせ論理削除
 *
 * 管理者機能なので user_id では絞らず、notificationId で対象を取得する。
 */
func (service *notificationService) DeleteNotification(
	req types.DeleteNotificationRequest,
) results.Result {
	findQuery, buildFindResult := service.notificationBuilder.BuildFindNotificationByIDQuery(req.NotificationID)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentNotification, findResult := service.notificationRepository.FindNotification(findQuery)
	if findResult.Error {
		return findResult
	}

	deletedNotification, buildDeleteResult := service.notificationBuilder.BuildDeleteNotificationModel(currentNotification)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	_, saveResult := service.notificationRepository.SaveNotification(deletedNotification)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteNotificationResponse{
			NotificationID: req.NotificationID,
		},
		"DELETE_NOTIFICATION_SUCCESS",
		"お知らせを削除しました",
		nil,
	)
}
