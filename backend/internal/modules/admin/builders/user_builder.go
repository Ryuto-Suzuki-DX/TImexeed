package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type UserBuilder interface {
	BuildSearchUsersQuery(req types.SearchUsersRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindUserByIDQuery(targetUserID uint) (*gorm.DB, results.Result)
	BuildCountActiveUserByEmailQuery(email string) (*gorm.DB, results.Result)
	BuildCountActiveUserByEmailExceptIDQuery(email string, targetUserID uint) (*gorm.DB, results.Result)
	BuildCreateUserModel(req types.CreateUserRequest, passwordHash string, hireDate time.Time) (models.User, results.Result)
	BuildUpdateUserModel(currentUser models.User, req types.UpdateUserRequest, hireDate time.Time, retirementDate *time.Time) (models.User, results.Result)
	BuildDeleteUserModel(currentUser models.User) (models.User, results.Result)
}

/*
 * 管理者用ユーザーBuilder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Serviceから受け取ったRequestをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Count / Create / Save はRepositoryに任せる
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
 * ユーザー検索用クエリ作成
 *
 * searchQuery：
 * ・一覧取得用
 * ・offset / limit / order を含む
 *
 * countQuery：
 * ・総件数取得用
 * ・offset / limit は含めない
 */
func (builder *userBuilder) BuildSearchUsersQuery(req types.SearchUsersRequest) (*gorm.DB, *gorm.DB, results.Result) {
	if req.Offset < 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_USERS_QUERY_INVALID_OFFSET",
			"ユーザー検索条件の作成に失敗しました",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	if req.Limit <= 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_USERS_QUERY_INVALID_LIMIT",
			"ユーザー検索条件の作成に失敗しました",
			map[string]any{
				"limit": req.Limit,
			},
		)
	}

	searchQuery := builder.db.Model(&models.User{})
	countQuery := builder.db.Model(&models.User{})

	searchQuery = applySearchUsersCondition(searchQuery, req)
	countQuery = applySearchUsersCondition(countQuery, req)

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
 * ユーザーID検索用クエリ作成
 *
 * 論理削除済みユーザーは対象外
 */
func (builder *userBuilder) BuildFindUserByIDQuery(targetUserID uint) (*gorm.DB, results.Result) {
	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_USER_BY_ID_QUERY_INVALID_TARGET_USER_ID",
			"ユーザー取得条件の作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("id = ?", targetUserID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_USER_BY_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有効ユーザーのメールアドレス件数確認クエリ作成
 *
 * 新規作成時の重複確認に使う。
 */
func (builder *userBuilder) BuildCountActiveUserByEmailQuery(email string) (*gorm.DB, results.Result) {
	if email == "" {
		return nil, results.BadRequest(
			"BUILD_COUNT_ACTIVE_USER_BY_EMAIL_QUERY_EMPTY_EMAIL",
			"メールアドレス重複確認条件の作成に失敗しました",
			nil,
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("email = ?", email).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_COUNT_ACTIVE_USER_BY_EMAIL_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 指定ユーザー以外の有効ユーザーのメールアドレス件数確認クエリ作成
 *
 * 更新時の重複確認に使う。
 */
func (builder *userBuilder) BuildCountActiveUserByEmailExceptIDQuery(email string, targetUserID uint) (*gorm.DB, results.Result) {
	if email == "" {
		return nil, results.BadRequest(
			"BUILD_COUNT_ACTIVE_USER_BY_EMAIL_EXCEPT_ID_QUERY_EMPTY_EMAIL",
			"メールアドレス重複確認条件の作成に失敗しました",
			nil,
		)
	}

	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_COUNT_ACTIVE_USER_BY_EMAIL_EXCEPT_ID_QUERY_INVALID_TARGET_USER_ID",
			"メールアドレス重複確認条件の作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("email = ?", email).
		Where("id <> ?", targetUserID).
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
	if passwordHash == "" {
		return models.User{}, results.BadRequest(
			"BUILD_CREATE_USER_MODEL_EMPTY_PASSWORD_HASH",
			"ユーザー作成データの作成に失敗しました",
			nil,
		)
	}

	if req.Role != "ADMIN" && req.Role != "USER" {
		return models.User{}, results.BadRequest(
			"BUILD_CREATE_USER_MODEL_INVALID_ROLE",
			"権限の値が正しくありません",
			map[string]any{
				"role": req.Role,
			},
		)
	}

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         req.Role,
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
		return models.User{}, results.BadRequest(
			"BUILD_UPDATE_USER_MODEL_EMPTY_CURRENT_USER",
			"ユーザー更新データの作成に失敗しました",
			nil,
		)
	}

	if req.Role != "ADMIN" && req.Role != "USER" {
		return models.User{}, results.BadRequest(
			"BUILD_UPDATE_USER_MODEL_INVALID_ROLE",
			"権限の値が正しくありません",
			map[string]any{
				"role": req.Role,
			},
		)
	}

	currentUser.Name = req.Name
	currentUser.Email = req.Email
	currentUser.Role = req.Role
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
func (builder *userBuilder) BuildDeleteUserModel(currentUser models.User) (models.User, results.Result) {
	if currentUser.ID == 0 {
		return models.User{}, results.BadRequest(
			"BUILD_DELETE_USER_MODEL_EMPTY_CURRENT_USER",
			"ユーザー削除データの作成に失敗しました",
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

/*
 * ユーザー検索条件をGORMクエリへ適用する
 */
func applySearchUsersCondition(query *gorm.DB, req types.SearchUsersRequest) *gorm.DB {
	if !req.IncludeDeleted {
		query = query.Where("is_deleted = ?", false)
	}

	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where(
			"name ILIKE ? OR email ILIKE ? OR role ILIKE ?",
			keyword,
			keyword,
			keyword,
		)
	}

	return query
}
