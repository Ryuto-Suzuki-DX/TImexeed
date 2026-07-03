package types

import "time"

/*
 * 管理者用お知らせ型定義
 *
 * 注意：
 * ・管理者本人宛のお知らせ検索/既読/未読件数は、JWTのuserIdを使う
 * ・検索/既読/未読件数では、フロントから userId / targetUserId は受け取らない
 * ・全員宛作成は、Repositoryで取得した有効ユーザー全員分のnotificationsを作成する
 */

/*
 * =========================================================
 * Request
 * =========================================================
 */

type SearchNotificationsRequest struct {
	Keyword string `json:"keyword"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

type ReadNotificationRequest struct {
	NotificationID uint `json:"notificationId" binding:"required"`
}

type CountUnreadNotificationsRequest struct{}

type CreateNotificationForAllUsersRequest struct {
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type DeleteNotificationRequest struct {
	NotificationID uint `json:"notificationId" binding:"required"`
}

/*
 * お知らせ既読状況取得リクエスト
 *
 * 管理者のお知らせ一覧に表示されているnotificationIdを受け取る。
 * Serviceで対象のお知らせのnotificationGroupIdを特定する。
 */
type GetNotificationReadStatusRequest struct {
	NotificationID uint `json:"notificationId" binding:"required"`
}

/*
 * =========================================================
 * Response
 * =========================================================
 */

type NotificationResponse struct {
	ID uint `json:"id"`

	UserID uint `json:"userId"`

	Title   string `json:"title"`
	Message string `json:"message"`

	IsRead bool       `json:"isRead"`
	ReadAt *time.Time `json:"readAt"`

	IsDeleted bool       `json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

type SearchNotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int64                  `json:"total"`
	Offset        int                    `json:"offset"`
	Limit         int                    `json:"limit"`
	HasMore       bool                   `json:"hasMore"`
}

type ReadNotificationResponse struct {
	Notification NotificationResponse `json:"notification"`
}

type CountUnreadNotificationsResponse struct {
	UnreadCount int64 `json:"unreadCount"`
}

type CreateNotificationForAllUsersResponse struct {
	CreatedCount int `json:"createdCount"`
}

type DeleteNotificationResponse struct {
	NotificationID uint `json:"notificationId"`
}

/*
 * お知らせ既読状況のユーザー情報
 *
 * 注意：
 * ・管理者は含めず、role = USER のユーザーだけを返す
 * ・既読ユーザーではreadAtに初回既読日時が入る
 * ・未読ユーザーではreadAtはnullになる
 */
type NotificationReadStatusUserResponse struct {
	UserID uint `json:"userId"`

	Name  string `json:"name"`
	Email string `json:"email"`

	DepartmentID   *uint   `json:"departmentId"`
	DepartmentName *string `json:"departmentName"`

	ReadAt *time.Time `json:"readAt"`
}

type GetNotificationReadStatusResponse struct {
	NotificationID      uint   `json:"notificationId"`
	NotificationGroupID string `json:"notificationGroupId"`

	Title   string `json:"title"`
	Message string `json:"message"`

	TargetUserCount int `json:"targetUserCount"`
	ReadUserCount   int `json:"readUserCount"`
	UnreadUserCount int `json:"unreadUserCount"`

	ReadUsers   []NotificationReadStatusUserResponse `json:"readUsers"`
	UnreadUsers []NotificationReadStatusUserResponse `json:"unreadUsers"`
}
