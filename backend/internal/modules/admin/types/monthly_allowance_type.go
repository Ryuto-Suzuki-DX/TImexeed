package types

import "time"

type SearchMonthlyAllowancesRequest struct {
	TargetUserID    *uint `json:"targetUserId"`
	TargetYear      *int  `json:"targetYear"`
	TargetMonth     *int  `json:"targetMonth"`
	AllowanceTypeID *uint `json:"allowanceTypeId"`
	IncludeDeleted  bool  `json:"includeDeleted"`
	Offset          int   `json:"offset"`
	Limit           int   `json:"limit"`
}
type MonthlyAllowanceDetailRequest struct {
	MonthlyAllowanceID uint `json:"monthlyAllowanceId" binding:"required"`
}
type CreateMonthlyAllowanceRequest struct {
	TargetUserID    uint   `json:"targetUserId" binding:"required"`
	TargetYear      int    `json:"targetYear" binding:"required"`
	TargetMonth     int    `json:"targetMonth" binding:"required"`
	AllowanceTypeID uint   `json:"allowanceTypeId" binding:"required"`
	Amount          int    `json:"amount"`
	Memo            string `json:"memo"`
}
type UpdateMonthlyAllowanceRequest struct {
	MonthlyAllowanceID uint   `json:"monthlyAllowanceId" binding:"required"`
	TargetUserID       uint   `json:"targetUserId" binding:"required"`
	TargetYear         int    `json:"targetYear" binding:"required"`
	TargetMonth        int    `json:"targetMonth" binding:"required"`
	AllowanceTypeID    uint   `json:"allowanceTypeId" binding:"required"`
	Amount             int    `json:"amount"`
	Memo               string `json:"memo"`
}
type DeleteMonthlyAllowanceRequest struct {
	MonthlyAllowanceID uint `json:"monthlyAllowanceId" binding:"required"`
}
type MonthlyAllowanceResponse struct {
	ID                uint       `json:"id"`
	UserID            uint       `json:"userId"`
	TargetYear        int        `json:"targetYear"`
	TargetMonth       int        `json:"targetMonth"`
	AllowanceTypeID   uint       `json:"allowanceTypeId"`
	AllowanceTypeName string     `json:"allowanceTypeName"`
	Amount            int        `json:"amount"`
	Memo              string     `json:"memo"`
	IsDeleted         bool       `json:"isDeleted"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	DeletedAt         *time.Time `json:"deletedAt"`
}
type SearchMonthlyAllowancesResponse struct {
	MonthlyAllowances []MonthlyAllowanceResponse `json:"monthlyAllowances"`
	Total             int64                      `json:"total"`
	Offset            int                        `json:"offset"`
	Limit             int                        `json:"limit"`
	HasMore           bool                       `json:"hasMore"`
}
type MonthlyAllowanceDetailResponse struct {
	MonthlyAllowance MonthlyAllowanceResponse `json:"monthlyAllowance"`
}
type CreateMonthlyAllowanceResponse struct {
	MonthlyAllowance MonthlyAllowanceResponse `json:"monthlyAllowance"`
}
type UpdateMonthlyAllowanceResponse struct {
	MonthlyAllowance MonthlyAllowanceResponse `json:"monthlyAllowance"`
}
type DeleteMonthlyAllowanceResponse struct {
	MonthlyAllowanceID uint `json:"monthlyAllowanceId"`
}
