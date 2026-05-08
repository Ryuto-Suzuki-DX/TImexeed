package types

import "time"

/*
 * 管理者用所属型定義
 *
 * 所属管理で使うRequest/Responseをまとめる。
 */

/*
 * =========================================================
 * Request
 * =========================================================
 */

type SearchDepartmentsRequest struct {
	Keyword        string `json:"keyword"`
	IncludeDeleted bool   `json:"includeDeleted"`
	Offset         int    `json:"offset"`
	Limit          int    `json:"limit"`
}

type DepartmentDetailRequest struct {
	DepartmentID uint `json:"departmentId" binding:"required"`
}

type CreateDepartmentRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateDepartmentRequest struct {
	DepartmentID uint   `json:"departmentId" binding:"required"`
	Name         string `json:"name" binding:"required"`
}

type DeleteDepartmentRequest struct {
	DepartmentID uint `json:"departmentId" binding:"required"`
}

/*
 * =========================================================
 * Response
 * =========================================================
 */

type DepartmentResponse struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	IsDeleted bool       `json:"isDeleted"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

type SearchDepartmentsResponse struct {
	Departments []DepartmentResponse `json:"departments"`
	Total       int64                `json:"total"`
	Offset      int                  `json:"offset"`
	Limit       int                  `json:"limit"`
	HasMore     bool                 `json:"hasMore"`
}

type DepartmentDetailResponse struct {
	Department DepartmentResponse `json:"department"`
}

type CreateDepartmentResponse struct {
	Department DepartmentResponse `json:"department"`
}

type UpdateDepartmentResponse struct {
	Department DepartmentResponse `json:"department"`
}

type DeleteDepartmentResponse struct {
	DepartmentID uint `json:"departmentId"`
}
