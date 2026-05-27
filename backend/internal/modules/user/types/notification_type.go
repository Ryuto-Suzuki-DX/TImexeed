package types

import "time"

/*
 * 〇 ユーザー お知らせ Type
 *
 * ユーザー側のお知らせ一覧取得・既読更新・未読件数取得で使用する。
 *
 * お知らせはユーザー向け表示用。
 * 通知種別・関連対象IDは持たない。
 *
 * 注意：
 * ・ユーザーIDはリクエストで受け取らない
 * ・ControllerでJWTからログイン中ユーザーIDを取得する
 * ・検索では keyword を title / message に対して部分一致検索する
 * ・検索結果は既存の検索系と同じく total / offset / limit / hasMore を返す
 * ・通知作成は公開APIとしては提供しない
 * ・月次申請などの内部処理からServiceを呼び出してnotificationsを作成する
 */

/*
 * =========================================================
 * Request
 * =========================================================
 */

/*
 * お知らせ一覧取得リクエスト
 */
type SearchNotificationsRequest struct {
	Keyword string `json:"keyword"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
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
 * 未読お知らせ件数取得リクエスト
 *
 * ユーザーIDはリクエストで受け取らない。
 * ControllerでJWTから取得する。
 */
type CountUnreadNotificationsRequest struct{}

/*
 * =========================================================
 * Response
 * =========================================================
 */

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
 * お知らせ一覧取得レスポンス
 */
type SearchNotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int64                  `json:"total"`
	Offset        int                    `json:"offset"`
	Limit         int                    `json:"limit"`
	HasMore       bool                   `json:"hasMore"`
}

/*
 * お知らせ既読更新レスポンス
 */
type ReadNotificationResponse struct {
	Notification NotificationResponse `json:"notification"`
}

/*
 * 未読お知らせ件数取得レスポンス
 */
type CountUnreadNotificationsResponse struct {
	UnreadCount int64 `json:"unreadCount"`
}

/*
 * 内部処理用 個別お知らせ作成レスポンス
 *
 * 注意：
 * ・Controllerから直接返すためのAPIレスポンスではなく、
 *   月次申請などの内部処理でresults.Resultに詰めるために使う。
 */
type CreateNotificationForUserResponse struct {
	UserID       uint `json:"userId"`
	CreatedCount int  `json:"createdCount"`
}

/*
 * 内部処理用 管理者宛お知らせ作成レスポンス
 *
 * 注意：
 * ・管理者が0件の場合も主処理を止めない。
 */
type CreateNotificationForAdminsResponse struct {
	AdminCount   int `json:"adminCount"`
	CreatedCount int `json:"createdCount"`
}
