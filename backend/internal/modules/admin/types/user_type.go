package types

import "time"

/*
 * 管理者用ユーザー型定義
 *
 * このファイルには、管理者用ユーザー機能で使う型をまとめる。
 *
 * まとめるもの：
 * ・Request型
 * ・Response型
 *
 * 方針：
 * ・URLにIDは載せない
 * ・詳細、更新、削除対象のユーザーIDは targetUserId で受け取る
 * ・ControllerではRequest型にbindして、そのままServiceへ渡す
 * ・Responseの日付は string に変換せず time.Time / *time.Time のまま返す
 * ・表示形式 yyyy-MM-dd などはフロント側で整形する
 */

/*
 * =========================================================
 * Request
 * =========================================================
 */

/*
 * ユーザー検索Request
 *
 * POST /admin/users/search
 *
 * body例：初回表示
 * {
 *   "keyword": "",
 *   "includeDeleted": false,
 *   "offset": 0,
 *   "limit": 50
 * }
 *
 * body例：フリーワード検索
 * {
 *   "keyword": "山田",
 *   "includeDeleted": false,
 *   "offset": 0,
 *   "limit": 50
 * }
 *
 * body例：さらに表示
 * {
 *   "keyword": "山田",
 *   "includeDeleted": false,
 *   "offset": 50,
 *   "limit": 50
 * }
 */
type SearchUsersRequest struct {
	// フリーワード検索
	// 名前、メールアドレスなどを対象にする想定
	Keyword string `json:"keyword"`

	// 削除済みユーザーも含めるか
	IncludeDeleted bool `json:"includeDeleted"`

	// 取得開始位置
	// 初回は0
	// さらに表示の場合は、現在フロントに表示済みの件数を入れる
	Offset int `json:"offset"`

	// 取得件数
	// 基本は50
	// 未指定、0以下、50超えの場合はService側で補正する
	Limit int `json:"limit"`
}

/*
 * ユーザー詳細取得Request
 *
 * POST /admin/users/detail
 *
 * body例：
 * {
 *   "targetUserId": 1
 * }
 */
type UserDetailRequest struct {
	// 詳細取得対象のユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`
}

/*
 * ユーザー新規作成Request
 *
 * POST /admin/users/create
 *
 * body例：
 * {
 *   "name": "山田太郎",
 *   "email": "yamada@example.com",
 *   "password": "password123",
 *   "role": "USER",
 *   "departmentId": 1,
 *   "hireDate": "2026-05-04"
 * }
 *
 * HireDate は "2026-05-04" のような文字列で受け取る。
 * time.Timeで直接受けるとJSONではRFC3339形式が必要になりやすいため、
 * Service側で日付文字列をparseする。
 */
type CreateUserRequest struct {
	// ユーザー名
	Name string `json:"name" binding:"required"`

	// メールアドレス
	Email string `json:"email" binding:"required,email"`

	// 初期パスワード
	// Service側でハッシュ化して保存する
	Password string `json:"password" binding:"required,min=8"`

	// 権限
	// USER または ADMIN を想定
	Role string `json:"role" binding:"required"`

	// 所属ID
	// 未所属の場合は null を許可する
	DepartmentID *uint `json:"departmentId"`

	// 入社日
	// yyyy-MM-dd形式
	HireDate string `json:"hireDate" binding:"required"`
}

/*
 * ユーザー更新Request
 *
 * POST /admin/users/update
 *
 * body例：
 * {
 *   "targetUserId": 1,
 *   "name": "山田太郎",
 *   "email": "yamada@example.com",
 *   "role": "USER",
 *   "departmentId": 1,
 *   "hireDate": "2026-05-04",
 *   "retirementDate": null
 * }
 *
 * RetirementDate は未退職なら null または空文字を想定する。
 */
type UpdateUserRequest struct {
	// 更新対象のユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// ユーザー名
	Name string `json:"name" binding:"required"`

	// メールアドレス
	Email string `json:"email" binding:"required,email"`

	// 権限
	// USER または ADMIN を想定
	Role string `json:"role" binding:"required"`

	// 所属ID
	// 未所属の場合は null を許可する
	DepartmentID *uint `json:"departmentId"`

	// 入社日
	// yyyy-MM-dd形式
	HireDate string `json:"hireDate" binding:"required"`

	// 退職日
	// 未退職の場合は null または空文字
	RetirementDate *string `json:"retirementDate"`
}

/*
 * ユーザー論理削除Request
 *
 * POST /admin/users/delete
 *
 * body例：
 * {
 *   "targetUserId": 1
 * }
 */
type DeleteUserRequest struct {
	// 論理削除対象のユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`
}

/*
 * =========================================================
 * Response
 * =========================================================
 */

/*
 * ユーザー1件分のResponse
 *
 * 日付は time.Time / *time.Time のまま返す。
 * フロント側で表示時に yyyy-MM-dd などへ整形する。
 */
type UserResponse struct {
	ID             uint       `json:"id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	Role           string     `json:"role"`
	DepartmentID   *uint      `json:"departmentId"`
	HireDate       time.Time  `json:"hireDate"`
	RetirementDate *time.Time `json:"retirementDate"`
	IsDeleted      bool       `json:"isDeleted"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"deletedAt"`
}

/*
 * ユーザー検索Response
 *
 * hasMore：
 * ・さらに表示するデータがある場合は true
 * ・すべて取得済みの場合は false
 *
 * フロント側の使い方：
 * ・hasMore が true なら「さらに表示」ボタンを表示する
 * ・hasMore が false なら「さらに表示」ボタンを非表示にする
 */
type SearchUsersResponse struct {
	Users   []UserResponse `json:"users"`
	Total   int64          `json:"total"`
	Offset  int            `json:"offset"`
	Limit   int            `json:"limit"`
	HasMore bool           `json:"hasMore"`
}

/*
 * ユーザー詳細Response
 */
type UserDetailResponse struct {
	User UserResponse `json:"user"`
}

/*
 * ユーザー作成Response
 */
type CreateUserResponse struct {
	User UserResponse `json:"user"`
}

/*
 * ユーザー更新Response
 */
type UpdateUserResponse struct {
	User UserResponse `json:"user"`
}

/*
 * ユーザー削除Response
 */
type DeleteUserResponse struct {
	TargetUserID uint `json:"targetUserId"`
}
