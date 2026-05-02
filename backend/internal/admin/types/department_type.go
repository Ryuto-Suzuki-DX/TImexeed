package types

/*
 * 所属モデルの定義
 */

/*
 * フロントへのレスポンス型
 */

// 所属一覧・詳細表示
type DepartmentResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	IsDeleted bool   `json:"isDeleted"`
}

type SearchDepartmentsResponse struct {
	Departments []DepartmentResponse `json:"departments"`
}

/*
 * フロントからのリクエスト型
 */

// 所属検索
type SearchDepartmentsRequest struct {
	Keyword        string `form:"keyword"`
	IncludeDeleted bool   `form:"includeDeleted"`
}

// 所属新規作成
type CreateDepartmentRequest struct {
	Name string `json:"name"`
}

// 所属更新
type UpdateDepartmentRequest struct {
	Name string `json:"name"`
}

/*
 * Service / Builder 内部で使う検索条件
 */
type SearchDepartmentsCondition struct {
	Keyword        string
	IncludeDeleted bool
}
