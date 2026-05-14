/*
 * 管理者 お知らせ Type
 *
 * バックエンドの admin/types/notification.go に対応する。
 *
 * 役割：
 * ・管理者本人宛のお知らせ一覧取得
 * ・管理者本人宛のお知らせ既読更新
 * ・管理者本人宛の未読お知らせ件数取得
 * ・全員宛お知らせ作成
 * ・お知らせ削除
 *
 * 注意：
 * ・管理者にも notifications は作成される
 * ・検索、既読、未読件数取得では userId / targetUserId は送らない
 * ・全員宛作成では ADMIN / USER 両方が対象になる
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

export type CreateNotificationForAllUsersRequest = {
  title: string;
  message: string;
};

export type CreateNotificationForAllUsersResponse = {
  createdCount: number;
};

export type DeleteNotificationRequest = {
  notificationId: number;
};

export type DeleteNotificationResponse = {
  notification: Notification;
};
