package builders

import (
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 従業員用お知らせBuilder interface
 *
 * ServiceがBuilderに求める処理だけを定義する。
 */
type NotificationBuilder interface {
	BuildSearchNotificationsQuery(userID uint, req types.SearchNotificationsRequest) (*gorm.DB, results.Result)
	BuildCountSearchNotificationsQuery(userID uint, req types.SearchNotificationsRequest) (*gorm.DB, results.Result)
	BuildFindNotificationByUserIDAndIDQuery(userID uint, notificationID uint) (*gorm.DB, results.Result)
	BuildReadNotificationModel(currentNotification models.Notification) (models.Notification, results.Result)
	BuildCountUnreadNotificationsQuery(userID uint) (*gorm.DB, results.Result)
}

/*
 * 従業員用お知らせBuilder
 *
 * 役割：
 * ・Serviceから受け取った値をもとにGORMクエリを作成する
 * ・Serviceから受け取ったModelをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Save はRepositoryに任せる
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
 * ログイン中ユーザー本人のお知らせ一覧を取得する。
 *
 * 注意：
 * ・userID はJWTから取得したログイン中ユーザーID
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
 * ・userID はJWTから取得したログイン中ユーザーID
 * ・notificationID はフロントから受け取ったお知らせID
 * ・ログイン中ユーザー本人のお知らせだけを対象にする
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
 * ログイン中ユーザー本人の未読お知らせ件数を取得する。
 *
 * 注意：
 * ・userID はJWTから取得したログイン中ユーザーID
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
