/*
 * 管理者 お知らせ自動リマインド Type
 *
 * バックエンドの admin/types/notification_reminder.go に対応する。
 *
 * 役割：
 * ・自動リマインド設定一覧取得
 * ・自動リマインド設定作成
 * ・自動リマインド設定更新
 * ・自動リマインド設定削除
 * ・自動リマインド設定の有効/無効切替
 *
 * 注意：
 * ・これは実際のお知らせではない
 * ・実際にユーザーへ表示されるお知らせは notifications に作成される
 * ・メール通知は今後対応なのでここでは扱わない
 */

export type NotificationReminder = {
  id: number;

  title: string;
  message: string;

  dayOffsetFromMonthEnd: number;
  sendHour: number;
  sendMinute: number;

  isEnabled: boolean;
  isDeleted: boolean;

  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchNotificationRemindersRequest = {
  keyword: string;
  includeDisabled: boolean;
  includeDeleted: boolean;
  limit: number;
  offset: number;
};

export type SearchNotificationRemindersResponse = {
  reminders: NotificationReminder[];
  hasMore: boolean;
};

export type CreateNotificationReminderRequest = {
  title: string;
  message: string;

  dayOffsetFromMonthEnd: number;
  sendHour: number;
  sendMinute: number;
};

export type CreateNotificationReminderResponse = {
  reminder: NotificationReminder;
};

export type UpdateNotificationReminderRequest = {
  reminderId: number;

  title: string;
  message: string;

  dayOffsetFromMonthEnd: number;
  sendHour: number;
  sendMinute: number;
  isEnabled: boolean;
};

export type UpdateNotificationReminderResponse = {
  reminder: NotificationReminder;
};

export type DeleteNotificationReminderRequest = {
  reminderId: number;
};

export type DeleteNotificationReminderResponse = {
  reminder: NotificationReminder;
};

export type ToggleNotificationReminderEnabledRequest = {
  reminderId: number;
  isEnabled: boolean;
};

export type ToggleNotificationReminderEnabledResponse = {
  reminder: NotificationReminder;
};
