/*
 * ユーザー お知らせ Type
 *
 * バックエンドの NotificationResponse / SearchNotificationsResponse /
 * ReadNotificationResponse / CountUnreadNotificationsResponse に対応する。
 */

export type Notification = {
  id: number;
  title: string;
  message: string;
  isRead: boolean;
  readAt: string | null;
  createdAt: string;
};

export type SearchNotificationsRequest = {
  limit: number;
  offset: number;
};

export type SearchNotificationsResponse = {
  notifications: Notification[];
  hasMore: boolean;
};

export type ReadNotificationRequest = {
  notificationId: number;
};

export type ReadNotificationResponse = {
  notification: Notification;
};

export type CountUnreadNotificationsRequest = Record<string, never>;

export type CountUnreadNotificationsResponse = {
  unreadCount: number;
};