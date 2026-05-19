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
	BuildFindNotificationByUserIDAndIDQuery(userID uint, notificationID uint) (*gorm.DB, results.Result)
	BuildReadNotificationModel(currentNotification models.Notification) (models.Notification, results.Result)
	BuildCountUnreadNotificationsQuery(userID uint) (*gorm.DB, results.Result)

	BuildCreateNotificationsForAllUsersModels(users []models.User, req types.CreateNotificationForAllUsersRequest) ([]models.Notification, results.Result)
	BuildCreateNotificationForUserModel(userID uint, title string, message string) (models.Notification, results.Result)
	BuildFindNotificationByIDQuery(notificationID uint) (*gorm.DB, results.Result)
	BuildDeleteNotificationModel(currentNotification models.Notification) (models.Notification, results.Result)
}

/*
 * 管理者用お知らせBuilder
 *
 * 役割：
 * ・Serviceから受け取った値をもとにGORMクエリを作成する
 * ・Serviceから受け取ったModelをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Save / Create はRepositoryに任せる
 * ・管理者本人宛のお知らせ取得では、userIDはJWTから取得した管理者IDを使う
 * ・全員宛作成では、Service/Repositoryで取得した有効ユーザー一覧をもとにnotificationsを作成する
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
 * お知らせ検索用の基本クエリ作成
 *
 * 一覧取得用クエリと件数取得用クエリで同じ検索条件を使う。
 */
func (builder *notificationBuilder) buildSearchNotificationsBaseQuery(
	userID uint,
	req types.SearchNotificationsRequest,
) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_NOTIFICATIONS_QUERY_INVALID_USER_ID",
			"お知らせ検索条件の作成に失敗しました",
			map[string]any{
				"userId": userID,
			},
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
		"BUILD_SEARCH_NOTIFICATIONS_BASE_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * お知らせ検索用クエリ作成
 *
 * ログイン中管理者本人のお知らせ一覧を取得する。
 *
 * 注意：
 * ・userID はJWTから取得したログイン中管理者ID
 * ・フロントから userId / targetUserId は受け取らない
 * ・論理削除済みのお知らせは対象外
 * ・keyword がある場合は title / message を部分一致検索する
 * ・新しいお知らせから順に取得する
 */
func (builder *notificationBuilder) BuildSearchNotificationsQuery(
	userID uint,
	req types.SearchNotificationsRequest,
) (*gorm.DB, results.Result) {
	if req.Limit <= 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_NOTIFICATIONS_QUERY_INVALID_LIMIT",
			"お知らせ検索条件の作成に失敗しました",
			map[string]any{
				"limit": req.Limit,
			},
		)
	}

	if req.Offset < 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_NOTIFICATIONS_QUERY_INVALID_OFFSET",
			"お知らせ検索条件の作成に失敗しました",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	query, buildResult := builder.buildSearchNotificationsBaseQuery(userID, req)
	if buildResult.Error {
		return nil, buildResult
	}

	query = query.
		Order("created_at DESC").
		Order("id DESC").
		Limit(req.Limit).
		Offset(req.Offset)

	return query, results.OK(
		nil,
		"BUILD_SEARCH_NOTIFICATIONS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * お知らせ検索件数取得用クエリ作成
 *
 * 検索条件に一致する総件数を取得する。
 */
func (builder *notificationBuilder) BuildCountSearchNotificationsQuery(
	userID uint,
	req types.SearchNotificationsRequest,
) (*gorm.DB, results.Result) {
	query, buildResult := builder.buildSearchNotificationsBaseQuery(userID, req)
	if buildResult.Error {
		return nil, buildResult
	}

	return query, results.OK(
		nil,
		"BUILD_COUNT_SEARCH_NOTIFICATIONS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザーID + お知らせIDでお知らせ1件取得用クエリ作成
 *
 * 既読更新時に使う。
 *
 * 注意：
 * ・userID はJWTから取得したログイン中管理者ID
 * ・notificationID はフロントから受け取ったお知らせID
 * ・ログイン中管理者本人のお知らせだけを対象にする
 * ・論理削除済みのお知らせは対象外
 */
func (builder *notificationBuilder) BuildFindNotificationByUserIDAndIDQuery(
	userID uint,
	notificationID uint,
) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_NOTIFICATION_QUERY_INVALID_USER_ID",
			"お知らせ取得条件の作成に失敗しました",
			map[string]any{
				"userId": userID,
			},
		)
	}

	if notificationID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_NOTIFICATION_QUERY_INVALID_NOTIFICATION_ID",
			"お知らせ取得条件の作成に失敗しました",
			map[string]any{
				"notificationId": notificationID,
			},
		)
	}

	query := builder.db.
		Model(&models.Notification{}).
		Where("id = ?", notificationID).
		Where("user_id = ?", userID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_NOTIFICATION_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * お知らせ既読更新用Model作成
 */
func (builder *notificationBuilder) BuildReadNotificationModel(
	currentNotification models.Notification,
) (models.Notification, results.Result) {
	if currentNotification.ID == 0 {
		return models.Notification{}, results.BadRequest(
			"BUILD_READ_NOTIFICATION_MODEL_EMPTY_CURRENT_NOTIFICATION",
			"お知らせ既読更新データの作成に失敗しました",
			nil,
		)
	}

	if currentNotification.IsRead {
		return currentNotification, results.OK(
			nil,
			"BUILD_READ_NOTIFICATION_MODEL_ALREADY_READ",
			"",
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
 * 未読お知らせ件数取得用クエリ作成
 *
 * ログイン中管理者本人の未読お知らせ件数を取得する。
 *
 * 注意：
 * ・userID はJWTから取得したログイン中管理者ID
 * ・フロントから userId / targetUserId は受け取らない
 * ・論理削除済みのお知らせは対象外
 * ・is_read = false のお知らせだけを対象にする
 */
func (builder *notificationBuilder) BuildCountUnreadNotificationsQuery(
	userID uint,
) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_COUNT_UNREAD_NOTIFICATIONS_QUERY_INVALID_USER_ID",
			"未読お知らせ件数取得条件の作成に失敗しました",
			map[string]any{
				"userId": userID,
			},
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
 * 全員宛お知らせ作成用Model配列作成
 *
 * 管理者が作成したお知らせを、全有効アカウント分のnotificationsとして作成する。
 *
 * 注意：
 * ・users はRepositoryで取得した is_deleted = false のユーザー一覧
 * ・USERだけでなくADMINも対象に含める
 * ・1ユーザーにつき1件のNotificationを作成する
 * ・既読状態は未読で作成する
 */
func (builder *notificationBuilder) BuildCreateNotificationsForAllUsersModels(
	users []models.User,
	req types.CreateNotificationForAllUsersRequest,
) ([]models.Notification, results.Result) {
	if req.Title == "" {
		return nil, results.BadRequest(
			"BUILD_CREATE_NOTIFICATIONS_FOR_ALL_USERS_MODELS_EMPTY_TITLE",
			"お知らせ作成データの作成に失敗しました",
			nil,
		)
	}

	if req.Message == "" {
		return nil, results.BadRequest(
			"BUILD_CREATE_NOTIFICATIONS_FOR_ALL_USERS_MODELS_EMPTY_MESSAGE",
			"お知らせ作成データの作成に失敗しました",
			nil,
		)
	}

	if len(users) == 0 {
		return nil, results.BadRequest(
			"BUILD_CREATE_NOTIFICATIONS_FOR_ALL_USERS_MODELS_EMPTY_USERS",
			"お知らせ作成対象のユーザーが存在しません",
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
			Title:     req.Title,
			Message:   req.Message,
			IsRead:    false,
			IsDeleted: false,
		})
	}

	if len(notifications) == 0 {
		return nil, results.BadRequest(
			"BUILD_CREATE_NOTIFICATIONS_FOR_ALL_USERS_MODELS_EMPTY_NOTIFICATIONS",
			"作成可能なお知らせデータがありません",
			nil,
		)
	}

	return notifications, results.OK(
		nil,
		"BUILD_CREATE_NOTIFICATIONS_FOR_ALL_USERS_MODELS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 個別ユーザー宛お知らせ作成用Model作成
 *
 * 月次勤怠承認・否認など、内部処理から特定ユーザーへ通知するときに使う。
 *
 * 注意：
 * ・API Request型には依存しない
 * ・1ユーザーにつき1件のNotificationを作成する
 * ・既読状態は未読で作成する
 */
func (builder *notificationBuilder) BuildCreateNotificationForUserModel(
	userID uint,
	title string,
	message string,
) (models.Notification, results.Result) {
	if userID == 0 {
		return models.Notification{}, results.BadRequest(
			"BUILD_CREATE_NOTIFICATION_FOR_USER_MODEL_INVALID_USER_ID",
			"お知らせ作成データの作成に失敗しました",
			map[string]any{
				"userId": userID,
			},
		)
	}

	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return models.Notification{}, results.BadRequest(
			"BUILD_CREATE_NOTIFICATION_FOR_USER_MODEL_EMPTY_TITLE",
			"お知らせタイトルを入力してください",
			nil,
		)
	}

	trimmedMessage := strings.TrimSpace(message)
	if trimmedMessage == "" {
		return models.Notification{}, results.BadRequest(
			"BUILD_CREATE_NOTIFICATION_FOR_USER_MODEL_EMPTY_MESSAGE",
			"お知らせ本文を入力してください",
			nil,
		)
	}

	notification := models.Notification{
		UserID:    userID,
		Title:     trimmedTitle,
		Message:   trimmedMessage,
		IsRead:    false,
		ReadAt:    nil,
		IsDeleted: false,
		DeletedAt: nil,
	}

	return notification, results.OK(
		nil,
		"BUILD_CREATE_NOTIFICATION_FOR_USER_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * お知らせIDでお知らせ1件取得用クエリ作成
 *
 * 管理者による削除時に使う。
 *
 * 注意：
 * ・削除は管理者機能なので、user_idでは絞らない
 * ・論理削除済みのお知らせは対象外
 */
func (builder *notificationBuilder) BuildFindNotificationByIDQuery(
	notificationID uint,
) (*gorm.DB, results.Result) {
	if notificationID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_NOTIFICATION_BY_ID_QUERY_INVALID_NOTIFICATION_ID",
			"お知らせ取得条件の作成に失敗しました",
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
 * お知らせ削除用Model作成
 *
 * 管理者がお知らせを論理削除する。
 *
 * 注意：
 * ・物理削除はしない
 * ・is_deleted = true にする
 * ・deleted_at を設定する
 */
func (builder *notificationBuilder) BuildDeleteNotificationModel(
	currentNotification models.Notification,
) (models.Notification, results.Result) {
	if currentNotification.ID == 0 {
		return models.Notification{}, results.BadRequest(
			"BUILD_DELETE_NOTIFICATION_MODEL_EMPTY_CURRENT_NOTIFICATION",
			"お知らせ削除データの作成に失敗しました",
			nil,
		)
	}

	if currentNotification.IsDeleted {
		return currentNotification, results.OK(
			nil,
			"BUILD_DELETE_NOTIFICATION_MODEL_ALREADY_DELETED",
			"",
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
