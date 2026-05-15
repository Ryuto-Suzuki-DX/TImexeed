package types

/*
 * 〇 管理者 ユーザー給与詳細 Type
 *
 * 管理者だけがユーザーごとの給与詳細を操作する。
 *
 * 注意：
 * ・従業員側APIでは使わない
 * ・URLにIDを載せない
 * ・対象ユーザーID、給与詳細IDは request body で受け取る
 * ・日付はフロントから yyyy-MM-dd 形式の string で受け取り、Service層で time.Time に変換する
 */

/*
 * 給与区分
 */
const (
	SalaryTypeMonthly = "MONTHLY"
	SalaryTypeHourly  = "HOURLY"
	SalaryTypeDaily   = "DAILY"
)

/*
 * 〇 ユーザー給与詳細検索 Request
 *
 * 対象ユーザーに紐づく給与詳細履歴を検索する。
 */
type SearchUserSalaryDetailsRequest struct {
	TargetUserID uint `json:"targetUserId" binding:"required"`

	IncludeDeleted bool `json:"includeDeleted"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

/*
 * 〇 ユーザー給与詳細単体取得 Request
 */
type GetUserSalaryDetailRequest struct {
	UserSalaryDetailID uint `json:"userSalaryDetailId" binding:"required"`
}

/*
 * 〇 ユーザー給与詳細作成 Request
 */
type CreateUserSalaryDetailRequest struct {
	TargetUserID uint `json:"targetUserId" binding:"required"`

	SalaryType string `json:"salaryType" binding:"required"`
	// MONTHLY: 月給
	// HOURLY: 時給
	// DAILY: 日給

	BaseAmount int `json:"baseAmount"`
	// 月給なら月額
	// 時給なら時給
	// 日給なら日給

	ExtraAllowanceAmount int    `json:"extraAllowanceAmount"`
	ExtraAllowanceMemo   string `json:"extraAllowanceMemo"`

	FixedDeductionAmount int    `json:"fixedDeductionAmount"`
	FixedDeductionMemo   string `json:"fixedDeductionMemo"`

	IsPayrollTarget bool `json:"isPayrollTarget"`

	EffectiveFrom string  `json:"effectiveFrom" binding:"required"`
	EffectiveTo   *string `json:"effectiveTo"`

	Memo string `json:"memo"`
}

/*
 * 〇 ユーザー給与詳細更新 Request
 */
type UpdateUserSalaryDetailRequest struct {
	UserSalaryDetailID uint `json:"userSalaryDetailId" binding:"required"`

	SalaryType string `json:"salaryType" binding:"required"`

	BaseAmount int `json:"baseAmount"`

	ExtraAllowanceAmount int    `json:"extraAllowanceAmount"`
	ExtraAllowanceMemo   string `json:"extraAllowanceMemo"`

	FixedDeductionAmount int    `json:"fixedDeductionAmount"`
	FixedDeductionMemo   string `json:"fixedDeductionMemo"`

	IsPayrollTarget bool `json:"isPayrollTarget"`

	EffectiveFrom string  `json:"effectiveFrom" binding:"required"`
	EffectiveTo   *string `json:"effectiveTo"`

	Memo string `json:"memo"`
}

/*
 * 〇 ユーザー給与詳細削除 Request
 */
type DeleteUserSalaryDetailRequest struct {
	UserSalaryDetailID uint `json:"userSalaryDetailId" binding:"required"`
}

/*
 * 〇 ユーザー給与詳細 Response
 */
type UserSalaryDetailResponse struct {
	ID uint `json:"id"`

	UserID uint `json:"userId"`

	SalaryType string `json:"salaryType"`

	BaseAmount int `json:"baseAmount"`

	ExtraAllowanceAmount int    `json:"extraAllowanceAmount"`
	ExtraAllowanceMemo   string `json:"extraAllowanceMemo"`

	FixedDeductionAmount int    `json:"fixedDeductionAmount"`
	FixedDeductionMemo   string `json:"fixedDeductionMemo"`

	IsPayrollTarget bool `json:"isPayrollTarget"`

	EffectiveFrom string  `json:"effectiveFrom"`
	EffectiveTo   *string `json:"effectiveTo"`

	Memo string `json:"memo"`

	IsDeleted bool    `json:"isDeleted"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	DeletedAt *string `json:"deletedAt"`
}

/*
 * 〇 ユーザー給与詳細検索 Response
 */
type SearchUserSalaryDetailsResponse struct {
	UserSalaryDetails []UserSalaryDetailResponse `json:"userSalaryDetails"`
	HasMore           bool                       `json:"hasMore"`
}

/*
 * 〇 ユーザー給与詳細単体取得 Response
 */
type GetUserSalaryDetailResponse struct {
	UserSalaryDetail UserSalaryDetailResponse `json:"userSalaryDetail"`
}

/*
 * 〇 ユーザー給与詳細作成 Response
 */
type CreateUserSalaryDetailResponse struct {
	UserSalaryDetail UserSalaryDetailResponse `json:"userSalaryDetail"`
}

/*
 * 〇 ユーザー給与詳細更新 Response
 */
type UpdateUserSalaryDetailResponse struct {
	UserSalaryDetail UserSalaryDetailResponse `json:"userSalaryDetail"`
}

/*
 * 〇 ユーザー給与詳細削除 Response
 */
type DeleteUserSalaryDetailResponse struct {
	UserSalaryDetailID uint `json:"userSalaryDetailId"`
}
