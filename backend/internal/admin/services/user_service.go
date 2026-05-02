package services

import (
	"strings"

	"timexeed/backend/internal/admin/builders"
	"timexeed/backend/internal/admin/repositories"
	"timexeed/backend/internal/admin/types"
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

/*
 * 管理者用ユーザーService
 *
 * Serviceの責務:
 * - Builderにクエリ/model作成を依頼する
 * - RepositoryにDB実行を依頼する
 * - 実行結果をレスポンス型に整える
 * - ControllerへResultで返す
 */

// インターフェース
type UserServiceInterface interface {
	InvalidRequest() results.Result
	InvalidUserID() results.Result
	SearchUsers(req types.SearchUsersRequest) results.Result
	GetUser(id uint) results.Result
	CreateUser(req types.CreateUserRequest) results.Result
	UpdateUser(id uint, req types.UpdateUserRequest) results.Result
	DeleteUser(id uint, loginUserID uint) results.Result
}

type UserService struct {
	db                   *gorm.DB
	userRepository       repositories.UserRepositoryInterface
	userBuilder          builders.UserBuilderInterface
	departmentRepository repositories.DepartmentRepositoryInterface
	departmentBuilder    builders.DepartmentBuilderInterface
}

/*
 * UserServiceを生成する
 */
func NewUserService(
	db *gorm.DB,
	userRepository *repositories.UserRepository,
	userBuilder *builders.UserBuilder,
	departmentRepository *repositories.DepartmentRepository,
	departmentBuilder *builders.DepartmentBuilder,
) *UserService {
	return &UserService{
		db:                   db,
		userRepository:       userRepository,
		userBuilder:          userBuilder,
		departmentRepository: departmentRepository,
		departmentBuilder:    departmentBuilder,
	}
}

/*
 * 不正なリクエスト形式
 */
func (s *UserService) InvalidRequest() results.Result {
	return results.ValidationError(map[string]string{
		"request": "リクエスト形式が正しくありません",
	})
}

/*
 * 不正なユーザーID
 */
func (s *UserService) InvalidUserID() results.Result {
	return results.BadRequest("INVALID_USER_ID", "ユーザーIDが正しくありません")
}

/*
 * ユーザー一覧取得
 */
func (s *UserService) SearchUsers(req types.SearchUsersRequest) results.Result {
	condition := types.SearchUsersCondition{
		Keyword:        strings.TrimSpace(req.Keyword),
		IncludeDeleted: req.IncludeDeleted,
	}

	query := s.userBuilder.BuildSearchUsersQuery(s.db, condition)

	users, err := s.userRepository.FindUsers(query)
	if err != nil {
		return results.InternalServerError("ユーザー一覧の取得に失敗しました")
	}

	departmentNameMap, err := s.buildDepartmentNameMap(users)
	if err != nil {
		return results.InternalServerError("所属情報の取得に失敗しました")
	}

	userResponses := make([]types.UserResponse, 0, len(users))

	for _, user := range users {
		userResponses = append(userResponses, toUserResponse(user, departmentNameMap))
	}

	return results.Success(types.SearchUsersResponse{
		Users: userResponses,
	}, "ユーザー一覧を取得しました")
}

/*
 * ユーザー詳細取得
 */
func (s *UserService) GetUser(id uint) results.Result {
	query := s.userBuilder.BuildFindActiveUserByIDQuery(s.db, id)

	user, err := s.userRepository.FindUser(query)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return results.NotFound("ユーザーが見つかりません")
		}

		return results.InternalServerError("ユーザーの取得に失敗しました")
	}

	departmentName := s.findDepartmentName(user.DepartmentID)

	return results.Success(toUserResponseByDepartmentName(user, departmentName), "ユーザーを取得しました")
}

/*
 * ユーザー新規作成
 */
func (s *UserService) CreateUser(req types.CreateUserRequest) results.Result {
	req = normalizeCreateUserRequest(req)

	validationErrors := validateCreateUserRequest(req)
	if len(validationErrors) > 0 {
		return results.ValidationError(validationErrors)
	}

	if req.DepartmentID != nil {
		if !s.existsDepartment(*req.DepartmentID) {
			return results.ValidationError(map[string]string{
				"departmentId": "所属が見つかりません",
			})
		}
	}

	countQuery := s.userBuilder.BuildCountActiveUserByEmailQuery(s.db, req.Email)

	count, err := s.userRepository.Count(countQuery)
	if err != nil {
		return results.InternalServerError("メールアドレスの確認に失敗しました")
	}

	if count > 0 {
		return results.Conflict("このメールアドレスはすでに使用されています")
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return results.InternalServerError("パスワードの暗号化に失敗しました")
	}

	user := s.userBuilder.BuildCreateUserModel(req, passwordHash)

	if err := s.userRepository.CreateUser(s.db, &user); err != nil {
		return results.InternalServerError("ユーザーの作成に失敗しました")
	}

	departmentName := s.findDepartmentName(user.DepartmentID)

	return results.Created(toUserResponseByDepartmentName(user, departmentName), "ユーザーを作成しました")
}

/*
 * ユーザー更新
 */
func (s *UserService) UpdateUser(id uint, req types.UpdateUserRequest) results.Result {
	req = normalizeUpdateUserRequest(req)

	validationErrors := validateUpdateUserRequest(req)
	if len(validationErrors) > 0 {
		return results.ValidationError(validationErrors)
	}

	if req.DepartmentID != nil {
		if !s.existsDepartment(*req.DepartmentID) {
			return results.ValidationError(map[string]string{
				"departmentId": "所属が見つかりません",
			})
		}
	}

	findQuery := s.userBuilder.BuildFindActiveUserByIDQuery(s.db, id)

	user, err := s.userRepository.FindUser(findQuery)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return results.NotFound("ユーザーが見つかりません")
		}

		return results.InternalServerError("ユーザーの取得に失敗しました")
	}

	countQuery := s.userBuilder.BuildCountActiveUserByEmailExceptIDQuery(s.db, req.Email, id)

	count, err := s.userRepository.Count(countQuery)
	if err != nil {
		return results.InternalServerError("メールアドレスの確認に失敗しました")
	}

	if count > 0 {
		return results.Conflict("このメールアドレスはすでに使用されています")
	}

	passwordHash := ""

	if req.Password != "" {
		hashedPassword, err := hashPassword(req.Password)
		if err != nil {
			return results.InternalServerError("パスワードの暗号化に失敗しました")
		}

		passwordHash = hashedPassword
	}

	user = s.userBuilder.BuildUpdateUserModel(user, req, passwordHash)

	if err := s.userRepository.SaveUser(s.db, &user); err != nil {
		return results.InternalServerError("ユーザーの更新に失敗しました")
	}

	departmentName := s.findDepartmentName(user.DepartmentID)

	return results.Success(toUserResponseByDepartmentName(user, departmentName), "ユーザーを更新しました")
}

/*
 * ユーザー論理削除
 */
func (s *UserService) DeleteUser(id uint, loginUserID uint) results.Result {
	if id == loginUserID {
		return results.BadRequest("CANNOT_DELETE_SELF", "自分自身のユーザーは削除できません")
	}

	findQuery := s.userBuilder.BuildFindActiveUserByIDQuery(s.db, id)

	user, err := s.userRepository.FindUser(findQuery)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return results.NotFound("ユーザーが見つかりません")
		}

		return results.InternalServerError("ユーザーの取得に失敗しました")
	}

	user = s.userBuilder.BuildDeleteUserModel(user)

	if err := s.userRepository.SaveUser(s.db, &user); err != nil {
		return results.InternalServerError("ユーザーの削除に失敗しました")
	}

	return results.Success(nil, "ユーザーを削除しました")
}

/*
 * ユーザー一覧用に所属IDと所属名のMapを作成する
 */
func (s *UserService) buildDepartmentNameMap(users []models.User) (map[uint]string, error) {
	departmentIDs := make([]uint, 0)
	seen := map[uint]bool{}

	for _, user := range users {
		if user.DepartmentID == nil {
			continue
		}

		departmentID := *user.DepartmentID

		if seen[departmentID] {
			continue
		}

		seen[departmentID] = true
		departmentIDs = append(departmentIDs, departmentID)
	}

	departmentNameMap := map[uint]string{}

	if len(departmentIDs) == 0 {
		return departmentNameMap, nil
	}

	query := s.departmentBuilder.BuildFindActiveDepartmentsByIDsQuery(s.db, departmentIDs)

	departments, err := s.departmentRepository.FindDepartments(query)
	if err != nil {
		return nil, err
	}

	for _, department := range departments {
		departmentNameMap[department.ID] = department.Name
	}

	return departmentNameMap, nil
}

/*
 * 所属IDから所属名を取得する
 *
 * 所属未設定または削除済みの場合は空文字を返す
 */
func (s *UserService) findDepartmentName(departmentID *uint) string {
	if departmentID == nil {
		return ""
	}

	query := s.departmentBuilder.BuildFindActiveDepartmentByIDQuery(s.db, *departmentID)

	department, err := s.departmentRepository.FindDepartment(query)
	if err != nil {
		return ""
	}

	return department.Name
}

/*
 * 所属が存在するか確認する
 */
func (s *UserService) existsDepartment(departmentID uint) bool {
	query := s.departmentBuilder.BuildFindActiveDepartmentByIDQuery(s.db, departmentID)

	_, err := s.departmentRepository.FindDepartment(query)

	return err == nil
}

/*
 * modelをレスポンス型へ変換する
 */
func toUserResponse(user models.User, departmentNameMap map[uint]string) types.UserResponse {
	departmentName := ""

	if user.DepartmentID != nil {
		departmentName = departmentNameMap[*user.DepartmentID]
	}

	return toUserResponseByDepartmentName(user, departmentName)
}

/*
 * modelをレスポンス型へ変換する
 */
func toUserResponseByDepartmentName(user models.User, departmentName string) types.UserResponse {
	return types.UserResponse{
		ID:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		Role:           user.Role,
		DepartmentID:   user.DepartmentID,
		DepartmentName: departmentName,
		IsDeleted:      user.IsDeleted,
	}
}

/*
 * ユーザー作成リクエストの前処理
 */
func normalizeCreateUserRequest(req types.CreateUserRequest) types.CreateUserRequest {
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	req.Role = strings.TrimSpace(req.Role)

	return req
}

/*
 * ユーザー更新リクエストの前処理
 */
func normalizeUpdateUserRequest(req types.UpdateUserRequest) types.UpdateUserRequest {
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	req.Role = strings.TrimSpace(req.Role)

	return req
}

/*
 * ユーザー作成時の入力チェック
 */
func validateCreateUserRequest(req types.CreateUserRequest) map[string]string {
	errors := map[string]string{}

	if req.Name == "" {
		errors["name"] = "名前を入力してください"
	}

	if req.Email == "" {
		errors["email"] = "メールアドレスを入力してください"
	}

	if req.Password == "" {
		errors["password"] = "パスワードを入力してください"
	} else if len(req.Password) < 8 {
		errors["password"] = "パスワードは8文字以上で入力してください"
	}

	if req.Role == "" {
		errors["role"] = "権限を選択してください"
	} else if req.Role != "ADMIN" && req.Role != "USER" {
		errors["role"] = "権限が正しくありません"
	}

	return errors
}

/*
 * ユーザー更新時の入力チェック
 *
 * 更新時はパスワード未入力なら変更しない
 */
func validateUpdateUserRequest(req types.UpdateUserRequest) map[string]string {
	errors := map[string]string{}

	if req.Name == "" {
		errors["name"] = "名前を入力してください"
	}

	if req.Email == "" {
		errors["email"] = "メールアドレスを入力してください"
	}

	if req.Password != "" && len(req.Password) < 8 {
		errors["password"] = "パスワードは8文字以上で入力してください"
	}

	if req.Role == "" {
		errors["role"] = "権限を選択してください"
	} else if req.Role != "ADMIN" && req.Role != "USER" {
		errors["role"] = "権限が正しくありません"
	}

	return errors
}

/*
 * パスワードをハッシュ化する
 */
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}
