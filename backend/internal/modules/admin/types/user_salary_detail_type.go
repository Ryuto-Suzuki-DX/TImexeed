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
 * ・手当は月次手当機能で管理する
 * ・控除はユーザー給与詳細では管理しない
 */

/*
 * 給与区分
 */
const (
	SalaryTypeMonthly = "MONTHLY"
	SalaryTypeHourly  = "HOURLY"
	SalaryTypeDaily   = "DAILY"
)

type SearchUserSalaryDetailsRequest struct {
	TargetUserID   uint `json:"targetUserId" binding:"required"`
	IncludeDeleted bool `json:"includeDeleted"`
	Offset         int  `json:"offset"`
	Limit          int  `json:"limit"`
}

type GetUserSalaryDetailRequest struct {
	UserSalaryDetailID uint `json:"userSalaryDetailId" binding:"required"`
}

type CreateUserSalaryDetailRequest struct {
	TargetUserID    uint    `json:"targetUserId" binding:"required"`
	SalaryType      string  `json:"salaryType" binding:"required"`
	BaseAmount      int     `json:"baseAmount"`
	IsPayrollTarget bool    `json:"isPayrollTarget"`
	EffectiveFrom   string  `json:"effectiveFrom" binding:"required"`
	EffectiveTo     *string `json:"effectiveTo"`
	Memo            string  `json:"memo"`
}

type UpdateUserSalaryDetailRequest struct {
	UserSalaryDetailID uint    `json:"userSalaryDetailId" binding:"required"`
	SalaryType         string  `json:"salaryType" binding:"required"`
	BaseAmount         int     `json:"baseAmount"`
	IsPayrollTarget    bool    `json:"isPayrollTarget"`
	EffectiveFrom      string  `json:"effectiveFrom" binding:"required"`
	EffectiveTo        *string `json:"effectiveTo"`
	Memo               string  `json:"memo"`
}

type DeleteUserSalaryDetailRequest struct {
	UserSalaryDetailID uint `json:"userSalaryDetailId" binding:"required"`
}

type UserSalaryDetailResponse struct {
	ID              uint    `json:"id"`
	UserID          uint    `json:"userId"`
	SalaryType      string  `json:"salaryType"`
	BaseAmount      int     `json:"baseAmount"`
	IsPayrollTarget bool    `json:"isPayrollTarget"`
	EffectiveFrom   string  `json:"effectiveFrom"`
	EffectiveTo     *string `json:"effectiveTo"`
	Memo            string  `json:"memo"`
	IsDeleted       bool    `json:"isDeleted"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
	DeletedAt       *string `json:"deletedAt"`
}

type SearchUserSalaryDetailsResponse struct {
	UserSalaryDetails []UserSalaryDetailResponse `json:"userSalaryDetails"`
	HasMore           bool                       `json:"hasMore"`
}

type GetUserSalaryDetailResponse struct {
	UserSalaryDetail UserSalaryDetailResponse `json:"userSalaryDetail"`
}
type CreateUserSalaryDetailResponse struct {
	UserSalaryDetail UserSalaryDetailResponse `json:"userSalaryDetail"`
}
type UpdateUserSalaryDetailResponse struct {
	UserSalaryDetail UserSalaryDetailResponse `json:"userSalaryDetail"`
}
type DeleteUserSalaryDetailResponse struct {
	UserSalaryDetailID uint `json:"userSalaryDetailId"`
}
