package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

type AllowanceController struct{ service services.AllowanceService }

func NewAllowanceController(service services.AllowanceService) *AllowanceController {
	return &AllowanceController{service: service}
}
func bindError(c *gin.Context, code, message string, err error) {
	responses.JSON(c, results.BadRequest(code, message, err.Error()))
}
func (x *AllowanceController) SearchAllowanceTypes(c *gin.Context) {
	var r types.SearchAllowanceTypesRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		bindError(c, "SEARCH_ALLOWANCE_TYPES_INVALID_REQUEST", "手当種別検索のリクエスト形式が正しくありません", e)
		return
	}
	responses.JSON(c, x.service.SearchAllowanceTypes(r))
}
func (x *AllowanceController) GetAllowanceTypeDetail(c *gin.Context) {
	var r types.AllowanceTypeDetailRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		bindError(c, "GET_ALLOWANCE_TYPE_DETAIL_INVALID_REQUEST", "手当種別詳細取得のリクエスト形式が正しくありません", e)
		return
	}
	responses.JSON(c, x.service.GetAllowanceTypeDetail(r))
}
func (x *AllowanceController) CreateAllowanceType(c *gin.Context) {
	var r types.CreateAllowanceTypeRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		bindError(c, "CREATE_ALLOWANCE_TYPE_INVALID_REQUEST", "手当種別作成のリクエスト形式が正しくありません", e)
		return
	}
	responses.JSON(c, x.service.CreateAllowanceType(r))
}
func (x *AllowanceController) UpdateAllowanceType(c *gin.Context) {
	var r types.UpdateAllowanceTypeRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		bindError(c, "UPDATE_ALLOWANCE_TYPE_INVALID_REQUEST", "手当種別更新のリクエスト形式が正しくありません", e)
		return
	}
	responses.JSON(c, x.service.UpdateAllowanceType(r))
}
func (x *AllowanceController) DeleteAllowanceType(c *gin.Context) {
	var r types.DeleteAllowanceTypeRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		bindError(c, "DELETE_ALLOWANCE_TYPE_INVALID_REQUEST", "手当種別削除のリクエスト形式が正しくありません", e)
		return
	}
	responses.JSON(c, x.service.DeleteAllowanceType(r))
}
func (x *AllowanceController) SearchMonthlyAllowances(c *gin.Context) {
	var r types.SearchMonthlyAllowancesRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		bindError(c, "SEARCH_MONTHLY_ALLOWANCES_INVALID_REQUEST", "月次手当検索のリクエスト形式が正しくありません", e)
		return
	}
	responses.JSON(c, x.service.SearchMonthlyAllowances(r))
}
func (x *AllowanceController) GetMonthlyAllowanceDetail(c *gin.Context) {
	var r types.MonthlyAllowanceDetailRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		bindError(c, "GET_MONTHLY_ALLOWANCE_DETAIL_INVALID_REQUEST", "月次手当詳細取得のリクエスト形式が正しくありません", e)
		return
	}
	responses.JSON(c, x.service.GetMonthlyAllowanceDetail(r))
}
func (x *AllowanceController) CreateMonthlyAllowance(c *gin.Context) {
	var r types.CreateMonthlyAllowanceRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		bindError(c, "CREATE_MONTHLY_ALLOWANCE_INVALID_REQUEST", "月次手当作成のリクエスト形式が正しくありません", e)
		return
	}
	responses.JSON(c, x.service.CreateMonthlyAllowance(r))
}
func (x *AllowanceController) UpdateMonthlyAllowance(c *gin.Context) {
	var r types.UpdateMonthlyAllowanceRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		bindError(c, "UPDATE_MONTHLY_ALLOWANCE_INVALID_REQUEST", "月次手当更新のリクエスト形式が正しくありません", e)
		return
	}
	responses.JSON(c, x.service.UpdateMonthlyAllowance(r))
}
func (x *AllowanceController) DeleteMonthlyAllowance(c *gin.Context) {
	var r types.DeleteMonthlyAllowanceRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		bindError(c, "DELETE_MONTHLY_ALLOWANCE_INVALID_REQUEST", "月次手当削除のリクエスト形式が正しくありません", e)
		return
	}
	responses.JSON(c, x.service.DeleteMonthlyAllowance(r))
}
