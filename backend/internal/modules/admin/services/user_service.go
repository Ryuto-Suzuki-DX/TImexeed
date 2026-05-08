package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 管理者用ユーザーService interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type UserService interface {
	SearchUsers(req types.SearchUsersRequest) results.Result
	GetUserDetail(req types.UserDetailRequest) results.Result
	CreateUser(req types.CreateUserRequest) results.Result
	UpdateUser(req types.UpdateUserRequest) results.Result
	DeleteUser(req types.DeleteUserRequest) results.Result
}

/*
 * 管理者用ユーザーService
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや更新用Modelを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB処理を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type userService struct {
	userBuilder    builders.UserBuilder
	userRepository repositories.UserRepository
}

/*
 * UserService生成
 */
func NewUserService(
	userBuilder builders.UserBuilder,
	userRepository repositories.UserRepository,
) *userService {
	return &userService{
		userBuilder:    userBuilder,
		userRepository: userRepository,
	}
}

/*
 * models.Userをフロント返却用UserResponseへ変換する
 *
 * 日付はtime.Time / *time.Timeのまま返す。
 * 表示形式の整形はフロント側で行う。
 */
func toUserResponse(user models.User) types.UserResponse {
	return types.UserResponse{
		ID:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		Role:           user.Role,
		DepartmentID:   user.DepartmentID,
		HireDate:       user.HireDate,
		RetirementDate: user.RetirementDate,
		IsDeleted:      user.IsDeleted,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
		DeletedAt:      user.DeletedAt,
	}
}

/*
 * 検索
 *
 * ページング方針：
 * ・初回は offset=0, limit=50
 * ・さらに表示するときは、フロントで現在表示済みの件数を offset として送る
 * ・limit が未指定、0以下の場合は 50件にする
 * ・limit が 50件を超える場合も 50件に丸める
 *
 * hasMore：
 * ・総件数 total が offset + 今回取得件数 より多ければ true
 * ・それ以下なら false
 */
func (service *userService) SearchUsers(req types.SearchUsersRequest) results.Result {
	// ページング検索条件を共通関数で正規化する
	normalizedCondition, normalizeResult := utils.NormalizePageSearchCondition(
		utils.PageSearchCondition{
			Keyword: req.Keyword,
			Offset:  req.Offset,
			Limit:   req.Limit,
		},
		"SEARCH_USERS_INVALID_OFFSET",
		"検索開始位置が正しくありません",
	)
	if normalizeResult.Error {
		return normalizeResult
	}

	req.Keyword = normalizedCondition.Keyword
	req.Offset = normalizedCondition.Offset
	req.Limit = normalizedCondition.Limit

	// Builderで一覧検索用クエリと件数取得用クエリを作成する
	searchQuery, countQuery, buildResult := service.userBuilder.BuildSearchUsersQuery(req)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryでユーザー一覧を取得する
	users, findResult := service.userRepository.FindUsers(searchQuery)
	if findResult.Error {
		return findResult
	}

	// Repositoryで検索条件に一致する総件数を取得する
	total, countResult := service.userRepository.CountUsers(countQuery)
	if countResult.Error {
		return countResult
	}

	// DBモデルをフロント返却用Responseへ変換する
	userResponses := make([]types.UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, toUserResponse(user))
	}

	hasMore := utils.HasMore(total, req.Offset, len(users))

	return results.OK(
		types.SearchUsersResponse{
			Users:   userResponses,
			Total:   total,
			Offset:  req.Offset,
			Limit:   req.Limit,
			HasMore: hasMore,
		},
		"SEARCH_USERS_SUCCESS",
		"ユーザー一覧を取得しました",
		nil,
	)
}

/*
 * 詳細
 */
func (service *userService) GetUserDetail(req types.UserDetailRequest) results.Result {
	// Builderで詳細取得用クエリを作成する
	query, buildResult := service.userBuilder.BuildFindUserByIDQuery(req.TargetUserID)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryでユーザーを取得する
	user, findResult := service.userRepository.FindUser(query)
	if findResult.Error {
		return findResult
	}

	return results.OK(
		types.UserDetailResponse{
			User: toUserResponse(user),
		},
		"GET_USER_DETAIL_SUCCESS",
		"ユーザー詳細を取得しました",
		nil,
	)
}

/*
 * 新規作成
 */
func (service *userService) CreateUser(req types.CreateUserRequest) results.Result {
	// 入社日を日付型へ変換する
	hireDate, err := utils.ParseDate(req.HireDate)
	if err != nil {
		return results.BadRequest(
			"CREATE_USER_INVALID_HIRE_DATE",
			"入社日の形式が正しくありません",
			map[string]any{
				"hireDate": req.HireDate,
				"format":   "yyyy-MM-dd",
			},
		)
	}

	// Builderでメールアドレス重複確認用クエリを作成する
	emailCountQuery, buildEmailCountResult := service.userBuilder.BuildCountActiveUserByEmailQuery(req.Email)
	if buildEmailCountResult.Error {
		return buildEmailCountResult
	}

	// Repositoryでメールアドレス重複確認を実行する
	emailCount, emailCountResult := service.userRepository.CountUsers(emailCountQuery)
	if emailCountResult.Error {
		return emailCountResult
	}

	if emailCount > 0 {
		return results.Conflict(
			"CREATE_USER_EMAIL_ALREADY_EXISTS",
			"このメールアドレスは既に使用されています",
			map[string]any{
				"email": req.Email,
			},
		)
	}

	// パスワードをハッシュ化する
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return results.InternalServerError(
			"CREATE_USER_HASH_PASSWORD_FAILED",
			"パスワードの暗号化に失敗しました",
			err.Error(),
		)
	}

	// Builderで作成用Modelを作る
	user, buildUserResult := service.userBuilder.BuildCreateUserModel(req, passwordHash, hireDate)
	if buildUserResult.Error {
		return buildUserResult
	}

	// Repositoryでユーザーを作成する
	createdUser, createResult := service.userRepository.CreateUser(user)
	if createResult.Error {
		return createResult
	}

	return results.Created(
		types.CreateUserResponse{
			User: toUserResponse(createdUser),
		},
		"CREATE_USER_SUCCESS",
		"ユーザーを作成しました",
		nil,
	)
}

/*
 * 更新
 */
func (service *userService) UpdateUser(req types.UpdateUserRequest) results.Result {
	// 入社日を日付型へ変換する
	hireDate, err := utils.ParseDate(req.HireDate)
	if err != nil {
		return results.BadRequest(
			"UPDATE_USER_INVALID_HIRE_DATE",
			"入社日の形式が正しくありません",
			map[string]any{
				"hireDate": req.HireDate,
				"format":   "yyyy-MM-dd",
			},
		)
	}

	// 退職日を日付型へ変換する
	retirementDate, err := utils.ParseOptionalDate(req.RetirementDate)
	if err != nil {
		return results.BadRequest(
			"UPDATE_USER_INVALID_RETIREMENT_DATE",
			"退職日の形式が正しくありません",
			map[string]any{
				"retirementDate": req.RetirementDate,
				"format":         "yyyy-MM-dd",
			},
		)
	}

	// Builderで対象ユーザー取得用クエリを作成する
	findQuery, buildFindResult := service.userBuilder.BuildFindUserByIDQuery(req.TargetUserID)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象ユーザーを取得する
	currentUser, findResult := service.userRepository.FindUser(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderでメールアドレス重複確認用クエリを作成する
	emailCountQuery, buildEmailCountResult := service.userBuilder.BuildCountActiveUserByEmailExceptIDQuery(req.Email, req.TargetUserID)
	if buildEmailCountResult.Error {
		return buildEmailCountResult
	}

	// Repositoryでメールアドレス重複確認を実行する
	emailCount, emailCountResult := service.userRepository.CountUsers(emailCountQuery)
	if emailCountResult.Error {
		return emailCountResult
	}

	if emailCount > 0 {
		return results.Conflict(
			"UPDATE_USER_EMAIL_ALREADY_EXISTS",
			"このメールアドレスは既に使用されています",
			map[string]any{
				"email":        req.Email,
				"targetUserId": req.TargetUserID,
			},
		)
	}

	// Builderで更新用Modelを作る
	updatedUser, buildUpdateResult := service.userBuilder.BuildUpdateUserModel(
		currentUser,
		req,
		hireDate,
		retirementDate,
	)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	// Repositoryでユーザーを更新する
	savedUser, saveResult := service.userRepository.SaveUser(updatedUser)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.UpdateUserResponse{
			User: toUserResponse(savedUser),
		},
		"UPDATE_USER_SUCCESS",
		"ユーザーを更新しました",
		nil,
	)
}

/*
 * ユーザー論理削除
 */
func (service *userService) DeleteUser(req types.DeleteUserRequest) results.Result {
	// Builderで対象ユーザー取得用クエリを作成する
	findQuery, buildFindResult := service.userBuilder.BuildFindUserByIDQuery(req.TargetUserID)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象ユーザーを取得する
	currentUser, findResult := service.userRepository.FindUser(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderで論理削除用Modelを作る
	deletedUser, buildDeleteResult := service.userBuilder.BuildDeleteUserModel(currentUser)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	// Repositoryでユーザーを保存する
	_, saveResult := service.userRepository.SaveUser(deletedUser)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteUserResponse{
			TargetUserID: req.TargetUserID,
		},
		"DELETE_USER_SUCCESS",
		"ユーザーを削除しました",
		nil,
	)
}
