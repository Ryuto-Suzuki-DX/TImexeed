import { apiPost } from "@/api/client";
import type {
  CreateNotificationReminderRequest,
  CreateNotificationReminderResponse,
  DeleteNotificationReminderRequest,
  DeleteNotificationReminderResponse,
  SearchNotificationRemindersRequest,
  SearchNotificationRemindersResponse,
  ToggleNotificationReminderEnabledRequest,
  ToggleNotificationReminderEnabledResponse,
  UpdateNotificationReminderRequest,
  UpdateNotificationReminderResponse,
} from "@/types/admin/notificationReminder";

/*
 * 管理者 自動リマインド一覧取得
 *
 * POST /admin/notification-reminders/search
 */
export function searchNotificationReminders(request: SearchNotificationRemindersRequest) {
  return apiPost<SearchNotificationRemindersResponse, SearchNotificationRemindersRequest>(
    "/admin/notification-reminders/search",
    request
  );
}

/*
 * 管理者 自動リマインド作成
 *
 * POST /admin/notification-reminders/create
 */
export function createNotificationReminder(request: CreateNotificationReminderRequest) {
  return apiPost<CreateNotificationReminderResponse, CreateNotificationReminderRequest>(
    "/admin/notification-reminders/create",
    request
  );
}

/*
 * 管理者 自動リマインド更新
 *
 * POST /admin/notification-reminders/update
 */
export function updateNotificationReminder(request: UpdateNotificationReminderRequest) {
  return apiPost<UpdateNotificationReminderResponse, UpdateNotificationReminderRequest>(
    "/admin/notification-reminders/update",
    request
  );
}

/*
 * 管理者 自動リマインド削除
 *
 * POST /admin/notification-reminders/delete
 */
export function deleteNotificationReminder(request: DeleteNotificationReminderRequest) {
  return apiPost<DeleteNotificationReminderResponse, DeleteNotificationReminderRequest>(
    "/admin/notification-reminders/delete",
    request
  );
}

/*
 * 管理者 自動リマインド有効/無効切替
 *
 * POST /admin/notification-reminders/toggle-enabled
 */
export function toggleNotificationReminderEnabled(request: ToggleNotificationReminderEnabledRequest) {
  return apiPost<ToggleNotificationReminderEnabledResponse, ToggleNotificationReminderEnabledRequest>(
    "/admin/notification-reminders/toggle-enabled",
    request
  );
}
