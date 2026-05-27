package builders

import (
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用お知らせBuilder interface
 *
 * ServiceがBuilderに求める処理だけを定義する。
 */
type NotificationBuilder interface {
	BuildSearchNotificationsQuery(userID uint, req types.SearchNotificationsRequest) (*gorm.DB, results.Result)
	BuildCountSearchNotificationsQuery(userID uint, req types.SearchNotificationsRequest) (*gorm.DB, results.Result)
	BuildCountUnreadNotificationsQuery(userID uint) (*gorm.DB, results.Result)
	BuildFindNotificationByUserIDAndIDQuery(userID uint, notificationID uint) (*gorm.DB, results.Result)
	BuildFindNotificationByIDQuery(notificationID uint) (*gorm.DB, results.Result)
	BuildReadNotificationModel(currentNotification models.Notification) (models.Notification, results.Result)
	BuildDeleteNotificationModel(currentNotification models.Notification) (models.Notification, results.Result)
	BuildCreateNotificationsForAllUsersModels(users []models.User, req types.CreateNotificationForAllUsersRequest) ([]models.Notification, results.Result)
	BuildCreateNotificationForUserModel(userID uint, title string, message string) (models.Notification, results.Result)
	BuildCreateNotificationsForUsersModels(users []models.User, title string, message string) ([]models.Notification, results.Result)
}

/*
 * 管理者用お知らせBuilder
 *
 * 役割：
 * ・Serviceから受け取った値をもとにGORMクエリを作成する
 * ・Serviceから受け取った値をもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DBアクセスはしない
 * ・query.Find / query.First / db.Create / db.Save はRepositoryで行う
 * ・業務処理の流れはServiceに任せる
 */
type notificationBuilder struct {
	db *gorm.DB
}

/*
 * NotificationBuilder生成
 */
func NewNotificationBuilder(db *gorm.DB) NotificationBuilder {
	return &notificationBuilder{db: db}
}

/*
 * お知らせ検索用Query作成
 *
 * ログイン中管理者本人のお知らせだけを取得する。
 */
func (builder *notificationBuilder) BuildSearchNotificationsQuery(
	userID uint,
	req types.SearchNotificationsRequest,
) (*gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, results.InternalServerError(
			"BUILD_SEARCH_NOTIFICATIONS_QUERY_DB_IS_NIL",
			"お知らせ検索条件の作成に失敗しました",
			nil,
		)
	}

	if userID == 0 {
		return nil, results.Unauthorized(
			"BUILD_SEARCH_NOTIFICATIONS_QUERY_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	query := builder.db.
		Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Where("is_deleted = ?", false)

	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.Where(
			"(title ILIKE ? OR message ILIKE ?)",
			likeKeyword,
			likeKeyword,
		)
	}

	query = query.
		Order("created_at DESC").
		Order("id DESC").
		Offset(req.Offset).
		Limit(req.Limit)

	return query, results.OK(
		nil,
		"BUILD_SEARCH_NOTIFICATIONS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * お知らせ検索件数取得用Query作成
 */
func (builder *notificationBuilder) BuildCountSearchNotificationsQuery(
	userID uint,
	req types.SearchNotificationsRequest,
) (*gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, results.InternalServerError(
			"BUILD_COUNT_SEARCH_NOTIFICATIONS_QUERY_DB_IS_NIL",
			"お知らせ検索件数条件の作成に失敗しました",
			nil,
		)
	}

	if userID == 0 {
		return nil, results.Unauthorized(
			"BUILD_COUNT_SEARCH_NOTIFICATIONS_QUERY_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	query := builder.db.
		Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Where("is_deleted = ?", false)

	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.Where(
			"(title ILIKE ? OR message ILIKE ?)",
			likeKeyword,
			likeKeyword,
		)
	}

	return query, results.OK(
		nil,
		"BUILD_COUNT_SEARCH_NOTIFICATIONS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 未読お知らせ件数取得用Query作成
 *
 * ログイン中管理者本人の未読お知らせ件数を取得する。
 */
func (builder *notificationBuilder) BuildCountUnreadNotificationsQuery(
	userID uint,
) (*gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, results.InternalServerError(
			"BUILD_COUNT_UNREAD_NOTIFICATIONS_QUERY_DB_IS_NIL",
			"未読お知らせ件数条件の作成に失敗しました",
			nil,
		)
	}

	if userID == 0 {
		return nil, results.Unauthorized(
			"BUILD_COUNT_UNREAD_NOTIFICATIONS_QUERY_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	query := builder.db.
		Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Where("is_read = ?", false).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_COUNT_UNREAD_NOTIFICATIONS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * userID + notificationID でお知らせ1件取得用Query作成
 *
 * 管理者本人のお知らせ既読更新で使う。
 */
func (builder *notificationBuilder) BuildFindNotificationByUserIDAndIDQuery(
	userID uint,
	notificationID uint,
) (*gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, results.InternalServerError(
			"BUILD_FIND_NOTIFICATION_BY_USER_ID_AND_ID_QUERY_DB_IS_NIL",
			"お知らせ取得条件の作成に失敗しました",
			nil,
		)
	}

	if userID == 0 {
		return nil, results.Unauthorized(
			"BUILD_FIND_NOTIFICATION_BY_USER_ID_AND_ID_QUERY_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	if notificationID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_NOTIFICATION_BY_USER_ID_AND_ID_QUERY_INVALID_NOTIFICATION_ID",
			"お知らせIDが正しくありません",
			map[string]any{
				"notificationId": notificationID,
			},
		)
	}

	query := builder.db.
		Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Where("id = ?", notificationID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_NOTIFICATION_BY_USER_ID_AND_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * notificationID でお知らせ1件取得用Query作成
 *
 * 管理者による論理削除で使う。
 */
func (builder *notificationBuilder) BuildFindNotificationByIDQuery(
	notificationID uint,
) (*gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, results.InternalServerError(
			"BUILD_FIND_NOTIFICATION_BY_ID_QUERY_DB_IS_NIL",
			"お知らせ取得条件の作成に失敗しました",
			nil,
		)
	}

	if notificationID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_NOTIFICATION_BY_ID_QUERY_INVALID_NOTIFICATION_ID",
			"お知らせIDが正しくありません",
			map[string]any{
				"notificationId": notificationID,
			},
		)
	}

	query := builder.db.
		Model(&models.Notification{}).
		Where("id = ?", notificationID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_NOTIFICATION_BY_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 既読更新用Model作成
 */
func (builder *notificationBuilder) BuildReadNotificationModel(
	currentNotification models.Notification,
) (models.Notification, results.Result) {
	if currentNotification.ID == 0 {
		return models.Notification{}, results.InternalServerError(
			"BUILD_READ_NOTIFICATION_MODEL_EMPTY_ID",
			"お知らせ既読更新情報の作成に失敗しました",
			nil,
		)
	}

	now := time.Now()
	currentNotification.IsRead = true
	currentNotification.ReadAt = &now

	return currentNotification, results.OK(
		nil,
		"BUILD_READ_NOTIFICATION_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 論理削除用Model作成
 */
func (builder *notificationBuilder) BuildDeleteNotificationModel(
	currentNotification models.Notification,
) (models.Notification, results.Result) {
	if currentNotification.ID == 0 {
		return models.Notification{}, results.InternalServerError(
			"BUILD_DELETE_NOTIFICATION_MODEL_EMPTY_ID",
			"お知らせ削除情報の作成に失敗しました",
			nil,
		)
	}

	now := time.Now()
	currentNotification.IsDeleted = true
	currentNotification.DeletedAt = &now

	return currentNotification, results.OK(
		nil,
		"BUILD_DELETE_NOTIFICATION_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 全員宛お知らせ作成用Models作成
 */
func (builder *notificationBuilder) BuildCreateNotificationsForAllUsersModels(
	users []models.User,
	req types.CreateNotificationForAllUsersRequest,
) ([]models.Notification, results.Result) {
	return builder.BuildCreateNotificationsForUsersModels(
		users,
		req.Title,
		req.Message,
	)
}

/*
 * 個別ユーザー宛お知らせ作成用Model作成
 */
func (builder *notificationBuilder) BuildCreateNotificationForUserModel(
	userID uint,
	title string,
	message string,
) (models.Notification, results.Result) {
	title = strings.TrimSpace(title)
	message = strings.TrimSpace(message)

	if userID == 0 {
		return models.Notification{}, results.BadRequest(
			"BUILD_CREATE_NOTIFICATION_FOR_USER_MODEL_INVALID_USER_ID",
			"通知対象ユーザーIDが正しくありません",
			map[string]any{
				"userId": userID,
			},
		)
	}

	if title == "" {
		return models.Notification{}, results.BadRequest(
			"BUILD_CREATE_NOTIFICATION_FOR_USER_MODEL_EMPTY_TITLE",
			"お知らせタイトルを入力してください",
			nil,
		)
	}

	if message == "" {
		return models.Notification{}, results.BadRequest(
			"BUILD_CREATE_NOTIFICATION_FOR_USER_MODEL_EMPTY_MESSAGE",
			"お知らせ本文を入力してください",
			nil,
		)
	}

	return models.Notification{
			UserID:    userID,
			Title:     title,
			Message:   message,
			IsRead:    false,
			IsDeleted: false,
		}, results.OK(
			nil,
			"BUILD_CREATE_NOTIFICATION_FOR_USER_MODEL_SUCCESS",
			"",
			nil,
		)
}

/*
 * 複数ユーザー宛お知らせ作成用Models作成
 */
func (builder *notificationBuilder) BuildCreateNotificationsForUsersModels(
	users []models.User,
	title string,
	message string,
) ([]models.Notification, results.Result) {
	title = strings.TrimSpace(title)
	message = strings.TrimSpace(message)

	if len(users) == 0 {
		return nil, results.BadRequest(
			"BUILD_CREATE_NOTIFICATIONS_FOR_USERS_MODELS_EMPTY_USERS",
			"通知対象ユーザーが存在しません",
			nil,
		)
	}

	if title == "" {
		return nil, results.BadRequest(
			"BUILD_CREATE_NOTIFICATIONS_FOR_USERS_MODELS_EMPTY_TITLE",
			"お知らせタイトルを入力してください",
			nil,
		)
	}

	if message == "" {
		return nil, results.BadRequest(
			"BUILD_CREATE_NOTIFICATIONS_FOR_USERS_MODELS_EMPTY_MESSAGE",
			"お知らせ本文を入力してください",
			nil,
		)
	}

	notifications := make([]models.Notification, 0, len(users))

	for _, user := range users {
		if user.ID == 0 {
			continue
		}

		notifications = append(notifications, models.Notification{
			UserID:    user.ID,
			Title:     title,
			Message:   message,
			IsRead:    false,
			IsDeleted: false,
		})
	}

	if len(notifications) == 0 {
		return nil, results.BadRequest(
			"BUILD_CREATE_NOTIFICATIONS_FOR_USERS_MODELS_EMPTY_NOTIFICATIONS",
			"作成対象のお知らせが存在しません",
			nil,
		)
	}

	return notifications, results.OK(
		nil,
		"BUILD_CREATE_NOTIFICATIONS_FOR_USERS_MODELS_SUCCESS",
		"",
		nil,
	)
}
