package types

/*
 * ユーザー検索リクエスト
 */
type SearchUsersRequest struct {
	Keyword        string
	IncludeDeleted bool
}

/*
 * ユーザー検索条件
 * Builderへ渡す検索条件
 */
type SearchUsersCondition struct {
	Keyword        string
	IncludeDeleted bool
}

/*
 * ユーザー一覧レスポンス
 */
type SearchUsersResponse struct {
	Users []UserResponse `json:"users"`
}

/*
 * ユーザー共通レスポンス
 */
type UserResponse struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	DepartmentID   *uint  `json:"departmentId"`
	DepartmentName string `json:"departmentName"`
	IsDeleted      bool   `json:"isDeleted"`
}

/*
 * ユーザー新規作成リクエスト
 */
type CreateUserRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	DepartmentID *uint  `json:"departmentId"`
}

/*
 * ユーザー更新リクエスト
 *
 * Passwordが空の場合は変更しない
 */
type UpdateUserRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	DepartmentID *uint  `json:"departmentId"`
}
