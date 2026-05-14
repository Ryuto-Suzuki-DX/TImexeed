package types

import "time"

/*
 * 管理者 お知らせ自動リマインド Type
 *
 * 管理者が毎月自動でお知らせを作成するためのルールを管理する。
 *
 * 注意：
 * ・これは実際のお知らせではない
 * ・実際に従業員へ表示されるお知らせは notifications テーブルに作成される
 * ・この型は notification_reminders テーブルの管理用
 * ・メール通知は今後対応なのでここでは扱わない
 */

type SearchNotificationRemindersRequest struct {
	Keyword         string `json:"keyword"`
	IncludeDisabled bool   `json:"includeDisabled"`
	IncludeDeleted  bool   `json:"includeDeleted"`
	Limit           int    `json:"limit"`
	Offset          int    `json:"offset"`
}

type SearchNotificationRemindersResponse struct {
	Reminders []NotificationReminderResponse `json:"reminders"`
	HasMore   bool                           `json:"hasMore"`
}

type NotificationReminderResponse struct {
	ID uint `json:"id"`

	Title   string `json:"title"`
	Message string `json:"message"`

	DayOffsetFromMonthEnd int `json:"dayOffsetFromMonthEnd"`
	SendHour              int `json:"sendHour"`
	SendMinute            int `json:"sendMinute"`

	IsEnabled bool `json:"isEnabled"`
	IsDeleted bool `json:"isDeleted"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

type CreateNotificationReminderRequest struct {
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`

	DayOffsetFromMonthEnd int `json:"dayOffsetFromMonthEnd"`
	SendHour              int `json:"sendHour"`
	SendMinute            int `json:"sendMinute"`
}

type CreateNotificationReminderResponse struct {
	Reminder NotificationReminderResponse `json:"reminder"`
}

type UpdateNotificationReminderRequest struct {
	ReminderID uint `json:"reminderId" binding:"required"`

	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`

	DayOffsetFromMonthEnd int  `json:"dayOffsetFromMonthEnd"`
	SendHour              int  `json:"sendHour"`
	SendMinute            int  `json:"sendMinute"`
	IsEnabled             bool `json:"isEnabled"`
}

type UpdateNotificationReminderResponse struct {
	Reminder NotificationReminderResponse `json:"reminder"`
}

type DeleteNotificationReminderRequest struct {
	ReminderID uint `json:"reminderId" binding:"required"`
}

type DeleteNotificationReminderResponse struct {
	Reminder NotificationReminderResponse `json:"reminder"`
}

type ToggleNotificationReminderEnabledRequest struct {
	ReminderID uint `json:"reminderId" binding:"required"`
	IsEnabled  bool `json:"isEnabled"`
}

type ToggleNotificationReminderEnabledResponse struct {
	Reminder NotificationReminderResponse `json:"reminder"`
}
