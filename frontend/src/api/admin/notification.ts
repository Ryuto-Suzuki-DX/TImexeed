import { apiPost } from "@/api/client";
import type {
  CountUnreadNotificationsRequest,
  CountUnreadNotificationsResponse,
  CreateNotificationForAllUsersRequest,
  CreateNotificationForAllUsersResponse,
  DeleteNotificationRequest,
  DeleteNotificationResponse,
  ReadNotificationRequest,
  ReadNotificationResponse,
  SearchNotificationsRequest,
  SearchNotificationsResponse,
} from "@/types/admin/notification";

/*
 * 管理者 お知らせ一覧取得
 *
 * POST /admin/notifications/search
 */
export function searchNotifications(request: SearchNotificationsRequest) {
  return apiPost<SearchNotificationsResponse, SearchNotificationsRequest>(
    "/admin/notifications/search",
    request
  );
}

/*
 * 管理者 お知らせ既読更新
 *
 * POST /admin/notifications/read
 */
export function readNotification(request: ReadNotificationRequest) {
  return apiPost<ReadNotificationResponse, ReadNotificationRequest>(
    "/admin/notifications/read",
    request
  );
}

/*
 * 管理者 未読お知らせ件数取得
 *
 * POST /admin/notifications/unread-count
 */
export function countUnreadNotifications(request: CountUnreadNotificationsRequest) {
  return apiPost<CountUnreadNotificationsResponse, CountUnreadNotificationsRequest>(
    "/admin/notifications/unread-count",
    request
  );
}

/*
 * 管理者 全員宛お知らせ作成
 *
 * POST /admin/notifications/create-for-all-users
 */
export function createNotificationForAllUsers(request: CreateNotificationForAllUsersRequest) {
  return apiPost<CreateNotificationForAllUsersResponse, CreateNotificationForAllUsersRequest>(
    "/admin/notifications/create-for-all-users",
    request
  );
}

/*
 * 管理者 お知らせ削除
 *
 * POST /admin/notifications/delete
 */
export function deleteNotification(request: DeleteNotificationRequest) {
  return apiPost<DeleteNotificationResponse, DeleteNotificationRequest>(
    "/admin/notifications/delete",
    request
  );
}
