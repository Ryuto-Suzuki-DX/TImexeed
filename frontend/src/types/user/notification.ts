/*
 * ユーザー お知らせ Type
 *
 * バックエンドの user/types/notification.go に対応する。
 *
 * 役割：
 * ・ログイン中ユーザー本人宛のお知らせ一覧取得
 * ・ログイン中ユーザー本人宛のお知らせ既読更新
 * ・ログイン中ユーザー本人宛の未読お知らせ件数取得
 *
 * 注意：
 * ・検索、既読、未読件数取得では userId / targetUserId は送らない
 * ・バックエンド側でJWTからログイン中ユーザーIDを取得する
 * ・keyword は title / message の検索用
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
  keyword: string;
  offset: number;
  limit: number;
};

export type SearchNotificationsResponse = {
  notifications: Notification[];
  total: number;
  offset: number;
  limit: number;
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
