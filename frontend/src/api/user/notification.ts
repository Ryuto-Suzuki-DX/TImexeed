import { apiPost } from "@/api/client";
import type {
  CountUnreadNotificationsRequest,
  CountUnreadNotificationsResponse,
  ReadNotificationRequest,
  ReadNotificationResponse,
  SearchNotificationsRequest,
  SearchNotificationsResponse,
} from "@/types/user/notification";

/*
 * お知らせ一覧取得
 *
 * POST /user/notifications/search
 */
export function searchNotifications(request: SearchNotificationsRequest) {
  return apiPost<SearchNotificationsResponse, SearchNotificationsRequest>("/user/notifications/search", request);
}

/*
 * お知らせ既読更新
 *
 * POST /user/notifications/read
 */
export function readNotification(request: ReadNotificationRequest) {
  return apiPost<ReadNotificationResponse, ReadNotificationRequest>("/user/notifications/read", request);
}

/*
 * 未読お知らせ件数取得
 *
 * POST /user/notifications/unread-count
 */
export function countUnreadNotifications(request: CountUnreadNotificationsRequest) {
  return apiPost<CountUnreadNotificationsResponse, CountUnreadNotificationsRequest>("/user/notifications/unread-count", request);
}