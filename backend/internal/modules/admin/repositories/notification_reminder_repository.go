package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用お知らせ自動リマインドRepository interface
 *
 * ServiceがRepositoryに求めるDB処理だけを定義する。
 */
type NotificationReminderRepository interface {
	FindNotificationReminders(query *gorm.DB) ([]models.NotificationReminder, results.Result)
	FindNotificationReminder(query *gorm.DB) (models.NotificationReminder, results.Result)
	CreateNotificationReminder(notificationReminder models.NotificationReminder) (models.NotificationReminder, results.Result)
	SaveNotificationReminder(notificationReminder models.NotificationReminder) (models.NotificationReminder, results.Result)
}

/*
 * 管理者用お知らせ自動リマインドRepository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・DBへのCreate / Saveを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 * ・作成可否、更新可否、削除可否などはServiceに任せる
 */
type notificationReminderRepository struct {
	db *gorm.DB
}

/*
 * NotificationReminderRepository生成
 */
func NewNotificationReminderRepository(db *gorm.DB) NotificationReminderRepository {
	return &notificationReminderRepository{db: db}
}

/*
 * 自動リマインド一覧取得
 */
func (repository *notificationReminderRepository) FindNotificationReminders(
	query *gorm.DB,
) ([]models.NotificationReminder, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_NOTIFICATION_REMINDERS_QUERY_IS_NIL",
			"自動リマインド一覧の取得に失敗しました",
			nil,
		)
	}

	var notificationReminders []models.NotificationReminder

	if err := query.Find(&notificationReminders).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_NOTIFICATION_REMINDERS_FAILED",
			"自動リマインド一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return notificationReminders, results.OK(
		nil,
		"FIND_NOTIFICATION_REMINDERS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 自動リマインド1件取得
 */
func (repository *notificationReminderRepository) FindNotificationReminder(
	query *gorm.DB,
) (models.NotificationReminder, results.Result) {
	if query == nil {
		return models.NotificationReminder{}, results.InternalServerError(
			"FIND_NOTIFICATION_REMINDER_QUERY_IS_NIL",
			"自動リマインド情報の取得に失敗しました",
			nil,
		)
	}

	var notificationReminder models.NotificationReminder

	if err := query.First(&notificationReminder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.NotificationReminder{}, results.NotFound(
				"NOTIFICATION_REMINDER_NOT_FOUND",
				"対象の自動リマインドが見つかりません",
				nil,
			)
		}

		return models.NotificationReminder{}, results.InternalServerError(
			"FIND_NOTIFICATION_REMINDER_FAILED",
			"自動リマインド情報の取得に失敗しました",
			err.Error(),
		)
	}

	return notificationReminder, results.OK(
		nil,
		"FIND_NOTIFICATION_REMINDER_SUCCESS",
		"",
		nil,
	)
}

/*
 * 自動リマインド作成
 */
func (repository *notificationReminderRepository) CreateNotificationReminder(
	notificationReminder models.NotificationReminder,
) (models.NotificationReminder, results.Result) {
	if notificationReminder.Title == "" {
		return models.NotificationReminder{}, results.InternalServerError(
			"CREATE_NOTIFICATION_REMINDER_EMPTY_TITLE",
			"自動リマインドの作成に失敗しました",
			nil,
		)
	}

	if notificationReminder.Message == "" {
		return models.NotificationReminder{}, results.InternalServerError(
			"CREATE_NOTIFICATION_REMINDER_EMPTY_MESSAGE",
			"自動リマインドの作成に失敗しました",
			nil,
		)
	}

	if err := repository.db.Create(&notificationReminder).Error; err != nil {
		return models.NotificationReminder{}, results.InternalServerError(
			"CREATE_NOTIFICATION_REMINDER_FAILED",
			"自動リマインドの作成に失敗しました",
			err.Error(),
		)
	}

	return notificationReminder, results.OK(
		nil,
		"CREATE_NOTIFICATION_REMINDER_SUCCESS",
		"",
		nil,
	)
}

/*
 * 自動リマインド保存
 *
 * 更新、論理削除、有効/無効切替で使う。
 */
func (repository *notificationReminderRepository) SaveNotificationReminder(
	notificationReminder models.NotificationReminder,
) (models.NotificationReminder, results.Result) {
	if notificationReminder.ID == 0 {
		return models.NotificationReminder{}, results.InternalServerError(
			"SAVE_NOTIFICATION_REMINDER_EMPTY_ID",
			"自動リマインド情報の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&notificationReminder).Error; err != nil {
		return models.NotificationReminder{}, results.InternalServerError(
			"SAVE_NOTIFICATION_REMINDER_FAILED",
			"自動リマインド情報の保存に失敗しました",
			err.Error(),
		)
	}

	return notificationReminder, results.OK(
		nil,
		"SAVE_NOTIFICATION_REMINDER_SUCCESS",
		"",
		nil,
	)
}
