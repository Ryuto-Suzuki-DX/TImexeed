package types

import "time"

type SearchAllowanceTypesRequest struct {
	Keyword         string `json:"keyword"`
	IncludeInactive bool   `json:"includeInactive"`
	IncludeDeleted  bool   `json:"includeDeleted"`
	Offset          int    `json:"offset"`
	Limit           int    `json:"limit"`
}
type AllowanceTypeDetailRequest struct {
	AllowanceTypeID uint `json:"allowanceTypeId" binding:"required"`
}
type CreateAllowanceTypeRequest struct {
	Name         string `json:"name" binding:"required"`
	DisplayOrder int    `json:"displayOrder"`
	IsActive     bool   `json:"isActive"`
}
type UpdateAllowanceTypeRequest struct {
	AllowanceTypeID uint   `json:"allowanceTypeId" binding:"required"`
	Name            string `json:"name" binding:"required"`
	DisplayOrder    int    `json:"displayOrder"`
	IsActive        bool   `json:"isActive"`
}
type DeleteAllowanceTypeRequest struct {
	AllowanceTypeID uint `json:"allowanceTypeId" binding:"required"`
}
type AllowanceTypeResponse struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	DisplayOrder int        `json:"displayOrder"`
	IsActive     bool       `json:"isActive"`
	IsDeleted    bool       `json:"isDeleted"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"deletedAt"`
}
type SearchAllowanceTypesResponse struct {
	AllowanceTypes []AllowanceTypeResponse `json:"allowanceTypes"`
	Total          int64                   `json:"total"`
	Offset         int                     `json:"offset"`
	Limit          int                     `json:"limit"`
	HasMore        bool                    `json:"hasMore"`
}
type AllowanceTypeDetailResponse struct {
	AllowanceType AllowanceTypeResponse `json:"allowanceType"`
}
type CreateAllowanceTypeResponse struct {
	AllowanceType AllowanceTypeResponse `json:"allowanceType"`
}
type UpdateAllowanceTypeResponse struct {
	AllowanceType AllowanceTypeResponse `json:"allowanceType"`
}
type DeleteAllowanceTypeResponse struct {
	AllowanceTypeID uint `json:"allowanceTypeId"`
}
