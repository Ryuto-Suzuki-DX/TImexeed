package types

import "time"

/*
 * 〇 管理者 お知らせ Type
 *
 * 管理者側のお知らせ一覧取得・既読更新・未読件数取得・全員宛作成・削除で使用する。
 *
 * 注意：
 * ・管理者自身にもお知らせは届く
 * ・管理者のお知らせ検索、既読更新、未読件数取得では userId をリクエストで受け取らない
 * ・ControllerでJWTからログイン中の管理者IDを取得する
 * ・全員宛作成では、有効なADMIN/USER全員にnotificationsを作成する
 */

type SearchNotificationsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type SearchNotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	HasMore       bool                   `json:"hasMore"`
}

type NotificationResponse struct {
	ID uint `json:"id"`

	Title   string `json:"title"`
	Message string `json:"message"`

	IsRead bool       `json:"isRead"`
	ReadAt *time.Time `json:"readAt"`

	CreatedAt time.Time `json:"createdAt"`
}

/*
 * お知らせ既読更新Request
 *
 * 管理者IDはリクエストで受け取らない。
 * ControllerでJWTから取得する。
 */
type ReadNotificationRequest struct {
	NotificationID uint `json:"notificationId" binding:"required"`
}

type ReadNotificationResponse struct {
	Notification NotificationResponse `json:"notification"`
}

type CountUnreadNotificationsRequest struct {
}

type CountUnreadNotificationsResponse struct {
	UnreadCount int64 `json:"unreadCount"`
}

/*
 * 全員宛お知らせ作成Request
 *
 * is_deleted = false の全アカウントへ作成する。
 * USERだけでなくADMINも対象に含める。
 */
type CreateNotificationForAllUsersRequest struct {
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type CreateNotificationForAllUsersResponse struct {
	CreatedCount int `json:"createdCount"`
}

/*
 * お知らせ削除Request
 *
 * 管理者による論理削除。
 */
type DeleteNotificationRequest struct {
	NotificationID uint `json:"notificationId" binding:"required"`
}

type DeleteNotificationResponse struct {
	Notification NotificationResponse `json:"notification"`
}
