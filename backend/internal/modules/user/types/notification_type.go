package types

import "time"

/*
 * 〇 ユーザー お知らせ Type
 *
 * ユーザー側のお知らせ一覧取得・既読更新で使用する。
 *
 * お知らせはユーザー向け表示用。
 * 通知種別・関連対象IDは持たない。
 */

/*
 * お知らせ一覧取得リクエスト
 *
 * ユーザーIDはリクエストで受け取らない。
 * ControllerでJWTから取得する。
 */
type SearchNotificationsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

/*
 * お知らせ一覧取得レスポンス
 */
type SearchNotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	HasMore       bool                   `json:"hasMore"`
}

/*
 * お知らせレスポンス
 */
type NotificationResponse struct {
	ID        uint       `json:"id"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	IsRead    bool       `json:"isRead"`
	ReadAt    *time.Time `json:"readAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

/*
 * お知らせ既読更新リクエスト
 *
 * ユーザーIDはリクエストで受け取らない。
 * ControllerでJWTから取得する。
 */
type ReadNotificationRequest struct {
	NotificationID uint `json:"notificationId" binding:"required"`
}

/*
 * お知らせ既読更新レスポンス
 */
type ReadNotificationResponse struct {
	Notification NotificationResponse `json:"notification"`
}

/*
 * 未読お知らせ件数取得リクエスト
 *
 * ユーザーIDはリクエストで受け取らない。
 * ControllerでJWTから取得する。
 */
type CountUnreadNotificationsRequest struct {
}

/*
 * 未読お知らせ件数取得レスポンス
 */
type CountUnreadNotificationsResponse struct {
	UnreadCount int64 `json:"unreadCount"`
}
