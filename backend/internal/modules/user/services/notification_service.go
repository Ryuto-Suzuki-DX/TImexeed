package services

import (
	"timexeed/backend/internal/mail"
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
)

/*
 * 従業員用お知らせService interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・従業員APIでは userId / targetUserId をRequestで受け取らない
 * ・ControllerでAuthMiddleware由来のuserIdを取得し、Serviceへ渡す
 * ・通知作成系はControllerから直接呼ばず、月次勤怠申請などの内部処理から呼ぶ
 */
type NotificationService interface {
	SearchNotifications(userID uint, req types.SearchNotificationsRequest) results.Result
	ReadNotification(userID uint, req types.ReadNotificationRequest) results.Result
	CountUnreadNotifications(userID uint, req types.CountUnreadNotificationsRequest) results.Result
	CreateNotificationForUser(userID uint, title string, message string) results.Result
	CreateNotificationForAdmins(title string, message string) results.Result
}

/*
 * 従業員用お知らせService
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや保存用Modelを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB処理を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 * ・お知らせ作成後、可能であれば対象者へメール送信する
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
	mailService            mail.MailService
}

/*
 * NotificationService生成
 */
func NewNotificationService(
	notificationBuilder builders.NotificationBuilder,
	notificationRepository repositories.NotificationRepository,
	mailService mail.MailService,
) *notificationService {
	return &notificationService{
		notificationBuilder:    notificationBuilder,
		notificationRepository: notificationRepository,
		mailService:            mailService,
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
 * お知らせメール送信
 *
 * 注意：
 * ・メール送信は副処理
 * ・メール送信に失敗しても、お知らせ作成自体は失敗にしない
 */
func (service *notificationService) sendNotificationMail(
	to string,
	title string,
	message string,
) {
	if service.mailService == nil {
		return
	}

	if to == "" {
		return
	}

	service.mailService.SendNotificationMail(to, title, message)
}

/*
 * お知らせ検索
 *
 * ログイン中ユーザー本人のお知らせ一覧を取得する。
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

	// Builderでお知らせ検索件数取得用クエリを作成する
	countQuery, buildCountResult := service.notificationBuilder.BuildCountSearchNotificationsQuery(userID, req)
	if buildCountResult.Error {
		return buildCountResult
	}

	// Repositoryで検索条件に一致する総件数を取得する
	total, countResult := service.notificationRepository.CountNotifications(countQuery)
	if countResult.Error {
		return countResult
	}

	// Builderでお知らせ検索用クエリを作成する
	query, buildSearchResult := service.notificationBuilder.BuildSearchNotificationsQuery(userID, req)
	if buildSearchResult.Error {
		return buildSearchResult
	}

	// Repositoryでお知らせ一覧を取得する
	notifications, findResult := service.notificationRepository.FindNotifications(query)
	if findResult.Error {
		return findResult
	}

	hasMore := int64(req.Offset+len(notifications)) < total

	// DBモデルをフロント返却用Responseへ変換する
	notificationResponses := make([]types.NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		notificationResponses = append(notificationResponses, toNotificationResponse(notification))
	}

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
 * ログイン中ユーザー本人の未読お知らせ件数を取得する。
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
 * 個別ユーザー宛お知らせ作成
 *
 * 月次勤怠申請/取り下げなど、内部処理から本人へ通知するときに使う。
 */
func (service *notificationService) CreateNotificationForUser(
	userID uint,
	title string,
	message string,
) results.Result {
	if userID == 0 {
		return results.BadRequest(
			"CREATE_NOTIFICATION_FOR_USER_INVALID_USER_ID",
			"通知対象ユーザーIDが正しくありません",
			map[string]any{
				"userId": userID,
			},
		)
	}

	user, findUserResult := service.notificationRepository.FindUserByID(userID)
	if findUserResult.Error {
		return findUserResult
	}

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

	service.sendNotificationMail(user.Email, title, message)

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
 * 管理者全員宛お知らせ作成
 *
 * 月次勤怠申請/取り下げなど、内部処理から管理者へ通知するときに使う。
 */
func (service *notificationService) CreateNotificationForAdmins(
	title string,
	message string,
) results.Result {
	admins, findAdminsResult := service.notificationRepository.FindActiveAdmins()
	if findAdminsResult.Error {
		return findAdminsResult
	}

	notifications, buildResult := service.notificationBuilder.BuildCreateNotificationsForUsersModels(
		admins,
		title,
		message,
	)
	if buildResult.Error {
		return buildResult
	}

	createdNotifications, createResult := service.notificationRepository.CreateNotifications(notifications)
	if createResult.Error {
		return createResult
	}

	for _, admin := range admins {
		service.sendNotificationMail(admin.Email, title, message)
	}

	return results.Created(
		map[string]any{
			"createdCount": len(createdNotifications),
		},
		"CREATE_NOTIFICATION_FOR_ADMINS_SUCCESS",
		"管理者宛のお知らせを作成しました",
		nil,
	)
}
