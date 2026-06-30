// 配置先:
// backend/internal/modules/user/services/password_service.go

package services

import (
	"strings"
	"time"

	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * パスワード Service Interface
 */
type PasswordService interface {
	ChangePassword(
		userID uint,
		request types.ChangePasswordRequest,
	) results.Result
}

/*
 * パスワード Service
 */
type passwordService struct {
	passwordBuilder    builders.PasswordBuilder
	passwordRepository repositories.PasswordRepository
}

/*
 * パスワード Service生成
 */
func NewPasswordService(
	passwordBuilder builders.PasswordBuilder,
	passwordRepository repositories.PasswordRepository,
) PasswordService {
	return &passwordService{
		passwordBuilder:    passwordBuilder,
		passwordRepository: passwordRepository,
	}
}

/*
 * パスワード変更
 */
func (service *passwordService) ChangePassword(
	userID uint,
	request types.ChangePasswordRequest,
) results.Result {
	currentPassword := strings.TrimSpace(request.CurrentPassword)
	newPassword := strings.TrimSpace(request.NewPassword)

	if userID == 0 {
		return results.BadRequest(
			"USER_ID_REQUIRED",
			"ユーザーIDが指定されていません。",
			nil,
		)
	}

	if currentPassword == "" {
		return results.BadRequest(
			"CURRENT_PASSWORD_REQUIRED",
			"現在のパスワードを入力してください。",
			nil,
		)
	}

	if newPassword == "" {
		return results.BadRequest(
			"NEW_PASSWORD_REQUIRED",
			"新しいパスワードを入力してください。",
			nil,
		)
	}

	if len(newPassword) < 8 {
		return results.BadRequest(
			"NEW_PASSWORD_TOO_SHORT",
			"新しいパスワードは8文字以上で入力してください。",
			nil,
		)
	}

	if currentPassword == newPassword {
		return results.BadRequest(
			"NEW_PASSWORD_SAME_AS_CURRENT",
			"現在のパスワードと異なるパスワードを設定してください。",
			nil,
		)
	}

	findQuery, buildFindResult := service.passwordBuilder.BuildFindUserByIDQuery(userID)
	if buildFindResult.Error {
		return buildFindResult
	}

	user, findResult := service.passwordRepository.FindUser(findQuery)
	if findResult.Error {
		return findResult
	}

	if user == nil || user.ID == 0 || user.IsDeleted {
		return results.NotFound(
			"USER_NOT_FOUND",
			"ユーザーが見つかりません。",
			nil,
		)
	}

	if !utils.CheckPassword(currentPassword, user.PasswordHash) {
		return results.BadRequest(
			"CURRENT_PASSWORD_INCORRECT",
			"現在のパスワードが正しくありません。",
			nil,
		)
	}

	newPasswordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return results.InternalServerError(
			"PASSWORD_HASH_FAILED",
			"パスワードの暗号化に失敗しました。",
			err,
		)
	}

	passwordChangedAt := time.Now()

	updateQuery, buildUpdateResult := service.passwordBuilder.BuildUpdatePasswordQuery(
		userID,
		newPasswordHash,
		passwordChangedAt,
	)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	updateResult := service.passwordRepository.UpdatePassword(updateQuery)
	if updateResult.Error {
		return updateResult
	}

	response := types.ChangePasswordResponse{
		MustChangePassword: false,
	}

	return results.OK(
		response,
		"CHANGE_PASSWORD_SUCCESS",
		"パスワードを変更しました。",
		nil,
	)
}
