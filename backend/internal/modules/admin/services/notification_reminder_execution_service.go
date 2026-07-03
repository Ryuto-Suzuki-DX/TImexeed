package services

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/repositories"

	"github.com/google/uuid"
)

/*
 * お知らせ自動リマインド実行Service interface
 */
type NotificationReminderExecutionService interface {
	ExecuteNotificationReminders() error
}

/*
 * お知らせ自動リマインド実行Service
 *
 * 1分ごとに呼び出され、現在日時と設定日時が一致した場合に
 * notificationsテーブルへ一般ユーザー分のお知らせを作成する。
 *
 * 注意：
 * ・実行履歴は保存しない
 * ・日付、時、分が完全一致した場合だけ実行する
 * ・メール送信処理は行わない
 */
type notificationReminderExecutionService struct {
	notificationReminderExecutionRepository repositories.NotificationReminderExecutionRepository
}

/*
 * NotificationReminderExecutionService生成
 */
func NewNotificationReminderExecutionService(
	notificationReminderExecutionRepository repositories.NotificationReminderExecutionRepository,
) NotificationReminderExecutionService {
	return &notificationReminderExecutionService{
		notificationReminderExecutionRepository: notificationReminderExecutionRepository,
	}
}

/*
 * 現在日時に一致する自動リマインドを実行する
 */
func (service *notificationReminderExecutionService) ExecuteNotificationReminders() error {
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return err
	}

	now := time.Now().In(location)

	reminders, err :=
		service.notificationReminderExecutionRepository.FindEnabledNotificationReminders()
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return nil
	}

	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		location,
	)

	monthEnd := time.Date(
		now.Year(),
		now.Month()+1,
		0,
		0,
		0,
		0,
		0,
		location,
	)

	daysUntilMonthEnd := int(monthEnd.Sub(today).Hours() / 24)

	dueReminders := make([]models.NotificationReminder, 0)

	for _, reminder := range reminders {
		if reminder.DayOffsetFromMonthEnd != daysUntilMonthEnd {
			continue
		}

		if reminder.SendHour != now.Hour() {
			continue
		}

		if reminder.SendMinute != now.Minute() {
			continue
		}

		dueReminders = append(dueReminders, reminder)
	}

	if len(dueReminders) == 0 {
		return nil
	}

	users, err :=
		service.notificationReminderExecutionRepository.FindActiveNotificationUsers(now)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		return nil
	}

	for _, reminder := range dueReminders {
		notificationGroupID := uuid.NewString()
		notifications := make([]models.Notification, 0, len(users))

		for _, user := range users {
			notifications = append(notifications, models.Notification{
				NotificationGroupID: &notificationGroupID,
				UserID:              user.ID,
				Title:               reminder.Title,
				Message:             reminder.Message,
				IsRead:              false,
				IsDeleted:           false,
			})
		}

		if err := service.notificationReminderExecutionRepository.CreateNotifications(
			notifications,
		); err != nil {
			return err
		}
	}

	return nil
}
