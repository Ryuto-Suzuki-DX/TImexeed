package builders

import (
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用ユーザーBuilder interface
 *
 * ServiceがBuilderに求める処理だけを定義する。
 */
type UserBuilder interface {
	BuildSearchUsersQuery(req types.SearchUsersRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildSearchBusinessTargetUsersQuery(req types.SearchBusinessTargetUsersRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindUserByIDQuery(userID uint) (*gorm.DB, results.Result)
	BuildCountActiveUserByEmailQuery(email string) (*gorm.DB, results.Result)
	BuildCountActiveUserByEmailExceptIDQuery(email string, exceptUserID uint) (*gorm.DB, results.Result)
	BuildCreateUserModel(req types.CreateUserRequest, passwordHash string, hireDate time.Time) (models.User, results.Result)
	BuildUpdateUserModel(currentUser models.User, req types.UpdateUserRequest, hireDate time.Time, retirementDate *time.Time) (models.User, results.Result)
	BuildDeleteUserModel(currentUser models.User) (models.User, results.Result)
}

/*
 * 管理者用ユーザーBuilder
 *
 * 役割：
 * ・Serviceから受け取った値をもとにGORMクエリを作成する
 * ・Serviceから受け取った値をもとにDB保存用Modelを作成する
 * ・Builderで発生したバリデーションエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Repositoryの処理は呼ばない
 * ・Serviceの業務フローは持たない
 */
type userBuilder struct {
	db *gorm.DB
}

/*
 * UserBuilder生成
 */
func NewUserBuilder(db *gorm.DB) UserBuilder {
	return &userBuilder{db: db}
}

/*
 * ユーザー一覧検索用クエリ作成
 *
 * 注意：
 * ・これはユーザー管理画面用
 * ・ADMIN / USER の両方を検索対象にする
 */
func (builder *userBuilder) BuildSearchUsersQuery(
	req types.SearchUsersRequest,
) (*gorm.DB, *gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, nil, results.InternalServerError(
			"BUILD_SEARCH_USERS_QUERY_DB_IS_NIL",
			"ユーザー検索の準備に失敗しました",
			nil,
		)
	}

	searchQuery := builder.db.Model(&models.User{})
	countQuery := builder.db.Model(&models.User{})

	if !req.IncludeDeleted {
		searchQuery = searchQuery.Where("is_deleted = ?", false)
		countQuery = countQuery.Where("is_deleted = ?", false)
	}

	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"

		searchQuery = searchQuery.Where(
			"name ILIKE ? OR email ILIKE ?",
			likeKeyword,
			likeKeyword,
		)

		countQuery = countQuery.Where(
			"name ILIKE ? OR email ILIKE ?",
			likeKeyword,
			likeKeyword,
		)
	}

	searchQuery = searchQuery.
		Order("id ASC").
		Offset(req.Offset).
		Limit(req.Limit)

	return searchQuery, countQuery, results.OK(
		nil,
		"BUILD_SEARCH_USERS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 業務対象ユーザー一覧検索用クエリ作成
 *
 * 注意：
 * ・勤怠、給与、経費、有給、個人情報Driveなどの対象ユーザー検索用
 * ・ADMIN は検索結果に含めない
 * ・削除済みユーザーも検索結果に含めない
 */
func (builder *userBuilder) BuildSearchBusinessTargetUsersQuery(
	req types.SearchBusinessTargetUsersRequest,
) (*gorm.DB, *gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, nil, results.InternalServerError(
			"BUILD_SEARCH_BUSINESS_TARGET_USERS_QUERY_DB_IS_NIL",
			"業務対象ユーザー検索の準備に失敗しました",
			nil,
		)
	}

	searchQuery := builder.db.Model(&models.User{}).
		Where("is_deleted = ?", false).
		Where("role = ?", "USER")

	countQuery := builder.db.Model(&models.User{}).
		Where("is_deleted = ?", false).
		Where("role = ?", "USER")

	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"

		searchQuery = searchQuery.Where(
			"name ILIKE ? OR email ILIKE ?",
			likeKeyword,
			likeKeyword,
		)

		countQuery = countQuery.Where(
			"name ILIKE ? OR email ILIKE ?",
			likeKeyword,
			likeKeyword,
		)
	}

	searchQuery = searchQuery.
		Order("id ASC").
		Offset(req.Offset).
		Limit(req.Limit)

	return searchQuery, countQuery, results.OK(
		nil,
		"BUILD_SEARCH_BUSINESS_TARGET_USERS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザーID検索用クエリ作成
 */
func (builder *userBuilder) BuildFindUserByIDQuery(userID uint) (*gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, results.InternalServerError(
			"BUILD_FIND_USER_BY_ID_QUERY_DB_IS_NIL",
			"ユーザー詳細取得の準備に失敗しました",
			nil,
		)
	}

	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_USER_BY_ID_QUERY_EMPTY_USER_ID",
			"対象ユーザーが指定されていません",
			nil,
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("id = ?", userID)

	return query, results.OK(
		nil,
		"BUILD_FIND_USER_BY_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有効ユーザーのメールアドレス件数確認用クエリ作成
 */
func (builder *userBuilder) BuildCountActiveUserByEmailQuery(email string) (*gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, results.InternalServerError(
			"BUILD_COUNT_ACTIVE_USER_BY_EMAIL_QUERY_DB_IS_NIL",
			"メールアドレス確認の準備に失敗しました",
			nil,
		)
	}

	trimmedEmail := strings.TrimSpace(email)
	if trimmedEmail == "" {
		return nil, results.BadRequest(
			"BUILD_COUNT_ACTIVE_USER_BY_EMAIL_QUERY_EMPTY_EMAIL",
			"メールアドレスを入力してください",
			nil,
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("email = ?", trimmedEmail).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_COUNT_ACTIVE_USER_BY_EMAIL_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 自分以外の有効ユーザーのメールアドレス件数確認用クエリ作成
 */
func (builder *userBuilder) BuildCountActiveUserByEmailExceptIDQuery(
	email string,
	exceptUserID uint,
) (*gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, results.InternalServerError(
			"BUILD_COUNT_ACTIVE_USER_BY_EMAIL_EXCEPT_ID_QUERY_DB_IS_NIL",
			"メールアドレス確認の準備に失敗しました",
			nil,
		)
	}

	trimmedEmail := strings.TrimSpace(email)
	if trimmedEmail == "" {
		return nil, results.BadRequest(
			"BUILD_COUNT_ACTIVE_USER_BY_EMAIL_EXCEPT_ID_QUERY_EMPTY_EMAIL",
			"メールアドレスを入力してください",
			nil,
		)
	}

	if exceptUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_COUNT_ACTIVE_USER_BY_EMAIL_EXCEPT_ID_QUERY_EMPTY_USER_ID",
			"対象ユーザーが指定されていません",
			nil,
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("email = ?", trimmedEmail).
		Where("id <> ?", exceptUserID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_COUNT_ACTIVE_USER_BY_EMAIL_EXCEPT_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー作成用Model作成
 */
func (builder *userBuilder) BuildCreateUserModel(
	req types.CreateUserRequest,
	passwordHash string,
	hireDate time.Time,
) (models.User, results.Result) {
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	role := strings.TrimSpace(req.Role)

	if name == "" {
		return models.User{}, results.BadRequest(
			"BUILD_CREATE_USER_MODEL_EMPTY_NAME",
			"ユーザー名を入力してください",
			nil,
		)
	}

	if email == "" {
		return models.User{}, results.BadRequest(
			"BUILD_CREATE_USER_MODEL_EMPTY_EMAIL",
			"メールアドレスを入力してください",
			nil,
		)
	}

	if role != "USER" && role != "ADMIN" {
		return models.User{}, results.BadRequest(
			"BUILD_CREATE_USER_MODEL_INVALID_ROLE",
			"権限が正しくありません",
			map[string]any{
				"role":    role,
				"allowed": []string{"USER", "ADMIN"},
			},
		)
	}

	if passwordHash == "" {
		return models.User{}, results.InternalServerError(
			"BUILD_CREATE_USER_MODEL_EMPTY_PASSWORD_HASH",
			"ユーザー作成の準備に失敗しました",
			nil,
		)
	}

	if hireDate.IsZero() {
		return models.User{}, results.BadRequest(
			"BUILD_CREATE_USER_MODEL_EMPTY_HIRE_DATE",
			"入社日が正しくありません",
			nil,
		)
	}

	user := models.User{
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		DepartmentID: req.DepartmentID,
		HireDate:     hireDate,
		IsDeleted:    false,
	}

	return user, results.OK(
		nil,
		"BUILD_CREATE_USER_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー更新用Model作成
 */
func (builder *userBuilder) BuildUpdateUserModel(
	currentUser models.User,
	req types.UpdateUserRequest,
	hireDate time.Time,
	retirementDate *time.Time,
) (models.User, results.Result) {
	if currentUser.ID == 0 {
		return models.User{}, results.InternalServerError(
			"BUILD_UPDATE_USER_MODEL_EMPTY_CURRENT_USER_ID",
			"ユーザー更新の準備に失敗しました",
			nil,
		)
	}

	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	role := strings.TrimSpace(req.Role)

	if name == "" {
		return models.User{}, results.BadRequest(
			"BUILD_UPDATE_USER_MODEL_EMPTY_NAME",
			"ユーザー名を入力してください",
			nil,
		)
	}

	if email == "" {
		return models.User{}, results.BadRequest(
			"BUILD_UPDATE_USER_MODEL_EMPTY_EMAIL",
			"メールアドレスを入力してください",
			nil,
		)
	}

	if role != "USER" && role != "ADMIN" {
		return models.User{}, results.BadRequest(
			"BUILD_UPDATE_USER_MODEL_INVALID_ROLE",
			"権限が正しくありません",
			map[string]any{
				"role":    role,
				"allowed": []string{"USER", "ADMIN"},
			},
		)
	}

	if hireDate.IsZero() {
		return models.User{}, results.BadRequest(
			"BUILD_UPDATE_USER_MODEL_EMPTY_HIRE_DATE",
			"入社日が正しくありません",
			nil,
		)
	}

	currentUser.Name = name
	currentUser.Email = email
	currentUser.Role = role
	currentUser.DepartmentID = req.DepartmentID
	currentUser.HireDate = hireDate
	currentUser.RetirementDate = retirementDate

	return currentUser, results.OK(
		nil,
		"BUILD_UPDATE_USER_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー論理削除用Model作成
 */
func (builder *userBuilder) BuildDeleteUserModel(
	currentUser models.User,
) (models.User, results.Result) {
	if currentUser.ID == 0 {
		return models.User{}, results.InternalServerError(
			"BUILD_DELETE_USER_MODEL_EMPTY_CURRENT_USER_ID",
			"ユーザー削除の準備に失敗しました",
			nil,
		)
	}

	now := time.Now()

	currentUser.IsDeleted = true
	currentUser.DeletedAt = &now

	return currentUser, results.OK(
		nil,
		"BUILD_DELETE_USER_MODEL_SUCCESS",
		"",
		nil,
	)
}
