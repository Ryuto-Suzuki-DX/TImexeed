package repositories

import (
	"time"

	"timexeed/backend/internal/models"

	"gorm.io/gorm"
)

/*
 * お知らせ自動リマインド実行Repository interface
 */
type NotificationReminderExecutionRepository interface {
	FindEnabledNotificationReminders() ([]models.NotificationReminder, error)
	FindActiveNotificationUsers(currentDate time.Time) ([]models.User, error)
	CreateNotifications(notifications []models.Notification) error
}

/*
 * お知らせ自動リマインド実行Repository
 *
 * 役割：
 * ・有効な自動リマインド設定を取得する
 * ・通知対象となるユーザーを取得する
 * ・notificationsテーブルへお知らせを作成する
 */
type notificationReminderExecutionRepository struct {
	db *gorm.DB
}

/*
 * NotificationReminderExecutionRepository生成
 */
func NewNotificationReminderExecutionRepository(
	db *gorm.DB,
) NotificationReminderExecutionRepository {
	return &notificationReminderExecutionRepository{
		db: db,
	}
}

/*
 * 有効な自動リマインド設定一覧取得
 *
 * 対象：
 * ・is_enabled = true
 * ・is_deleted = false
 */
func (repository *notificationReminderExecutionRepository) FindEnabledNotificationReminders() (
	[]models.NotificationReminder,
	error,
) {
	var reminders []models.NotificationReminder

	err := repository.db.
		Model(&models.NotificationReminder{}).
		Where("is_enabled = ?", true).
		Where("is_deleted = ?", false).
		Order("id ASC").
		Find(&reminders).
		Error
	if err != nil {
		return nil, err
	}

	return reminders, nil
}

/*
 * 通知対象ユーザー一覧取得
 *
 * 対象：
 * ・is_deleted = false
 * ・退職日が未設定、または現在日以降
 */
func (repository *notificationReminderExecutionRepository) FindActiveNotificationUsers(
	currentDate time.Time,
) ([]models.User, error) {
	var users []models.User

	err := repository.db.
		Model(&models.User{}).
		Where("is_deleted = ?", false).
		Where(
			"(retirement_date IS NULL OR retirement_date >= ?)",
			currentDate.Format("2006-01-02"),
		).
		Order("id ASC").
		Find(&users).
		Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

/*
 * お知らせ一括作成
 *
 * 同じ自動リマインドから作成する通知には、
 * Service側で同じnotification_group_idを設定する。
 */
func (repository *notificationReminderExecutionRepository) CreateNotifications(
	notifications []models.Notification,
) error {
	if len(notifications) == 0 {
		return nil
	}

	return repository.db.Create(&notifications).Error
}
