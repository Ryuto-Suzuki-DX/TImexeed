package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用お知らせRepository interface
 *
 * ServiceがRepositoryに求めるDB処理だけを定義する。
 */
type NotificationRepository interface {
	FindNotifications(query *gorm.DB) ([]models.Notification, results.Result)
	FindNotification(query *gorm.DB) (models.Notification, results.Result)
	SaveNotification(notification models.Notification) (models.Notification, results.Result)
	CountNotifications(query *gorm.DB) (int64, results.Result)

	FindActiveUsers() ([]models.User, results.Result)
	CreateNotifications(notifications []models.Notification) ([]models.Notification, results.Result)
}

/*
 * 管理者用お知らせRepository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・DBへのSave / Createを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 * ・既読更新可否、全員宛作成可否、削除可否などはServiceに任せる
 */
type notificationRepository struct {
	db *gorm.DB
}

/*
 * NotificationRepository生成
 */
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

/*
 * お知らせ一覧取得
 */
func (repository *notificationRepository) FindNotifications(query *gorm.DB) ([]models.Notification, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_NOTIFICATIONS_QUERY_IS_NIL",
			"お知らせ一覧の取得に失敗しました",
			nil,
		)
	}

	var notifications []models.Notification

	if err := query.Find(&notifications).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_NOTIFICATIONS_FAILED",
			"お知らせ一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return notifications, results.OK(
		nil,
		"FIND_NOTIFICATIONS_SUCCESS",
		"",
		nil,
	)
}

/*
 * お知らせ1件取得
 */
func (repository *notificationRepository) FindNotification(query *gorm.DB) (models.Notification, results.Result) {
	if query == nil {
		return models.Notification{}, results.InternalServerError(
			"FIND_NOTIFICATION_QUERY_IS_NIL",
			"お知らせ情報の取得に失敗しました",
			nil,
		)
	}

	var notification models.Notification

	if err := query.First(&notification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Notification{}, results.NotFound(
				"NOTIFICATION_NOT_FOUND",
				"対象のお知らせが見つかりません",
				nil,
			)
		}

		return models.Notification{}, results.InternalServerError(
			"FIND_NOTIFICATION_FAILED",
			"お知らせ情報の取得に失敗しました",
			err.Error(),
		)
	}

	return notification, results.OK(
		nil,
		"FIND_NOTIFICATION_SUCCESS",
		"",
		nil,
	)
}

/*
 * お知らせ保存
 *
 * 既読更新、論理削除で使う。
 */
func (repository *notificationRepository) SaveNotification(notification models.Notification) (models.Notification, results.Result) {
	if notification.ID == 0 {
		return models.Notification{}, results.InternalServerError(
			"SAVE_NOTIFICATION_EMPTY_ID",
			"お知らせ情報の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&notification).Error; err != nil {
		return models.Notification{}, results.InternalServerError(
			"SAVE_NOTIFICATION_FAILED",
			"お知らせ情報の保存に失敗しました",
			err.Error(),
		)
	}

	return notification, results.OK(
		nil,
		"SAVE_NOTIFICATION_SUCCESS",
		"",
		nil,
	)
}

/*
 * お知らせ件数取得
 *
 * 未読件数取得などで使う。
 */
func (repository *notificationRepository) CountNotifications(query *gorm.DB) (int64, results.Result) {
	if query == nil {
		return 0, results.InternalServerError(
			"COUNT_NOTIFICATIONS_QUERY_IS_NIL",
			"お知らせ件数の取得に失敗しました",
			nil,
		)
	}

	var count int64

	if err := query.Count(&count).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_NOTIFICATIONS_FAILED",
			"お知らせ件数の取得に失敗しました",
			err.Error(),
		)
	}

	return count, results.OK(
		nil,
		"COUNT_NOTIFICATIONS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有効ユーザー一覧取得
 *
 * 全員宛お知らせ作成で使う。
 *
 * 注意：
 * ・USERだけでなくADMINも対象に含める
 * ・論理削除済みユーザーは対象外
 */
func (repository *notificationRepository) FindActiveUsers() ([]models.User, results.Result) {
	var users []models.User

	if err := repository.db.
		Model(&models.User{}).
		Where("is_deleted = ?", false).
		Order("id ASC").
		Find(&users).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_ACTIVE_USERS_FAILED",
			"お知らせ作成対象ユーザーの取得に失敗しました",
			err.Error(),
		)
	}

	return users, results.OK(
		nil,
		"FIND_ACTIVE_USERS_SUCCESS",
		"",
		nil,
	)
}

/*
 * お知らせ一括作成
 *
 * 全員宛お知らせ作成で使う。
 */
func (repository *notificationRepository) CreateNotifications(
	notifications []models.Notification,
) ([]models.Notification, results.Result) {
	if len(notifications) == 0 {
		return nil, results.InternalServerError(
			"CREATE_NOTIFICATIONS_EMPTY_NOTIFICATIONS",
			"お知らせの作成に失敗しました",
			nil,
		)
	}

	if err := repository.db.Create(&notifications).Error; err != nil {
		return nil, results.InternalServerError(
			"CREATE_NOTIFICATIONS_FAILED",
			"お知らせの作成に失敗しました",
			err.Error(),
		)
	}

	return notifications, results.OK(
		nil,
		"CREATE_NOTIFICATIONS_SUCCESS",
		"",
		nil,
	)
}
