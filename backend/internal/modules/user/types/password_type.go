// 配置先:
// backend/internal/modules/user/types/password_type.go

package types

/*
 * パスワード変更リクエスト
 */
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

/*
 * パスワード変更レスポンス
 */
type ChangePasswordResponse struct {
	MustChangePassword bool `json:"mustChangePassword"`
}
