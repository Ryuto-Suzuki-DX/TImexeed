package builders

import (
	"strings"
	"time"
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type AllowanceBuilder interface {
	BuildSearchAllowanceTypesQuery(types.SearchAllowanceTypesRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindAllowanceTypeByIDQuery(uint) (*gorm.DB, results.Result)
	BuildCreateAllowanceTypeModel(types.CreateAllowanceTypeRequest) (models.AllowanceType, results.Result)
	BuildUpdateAllowanceTypeModel(models.AllowanceType, types.UpdateAllowanceTypeRequest) (models.AllowanceType, results.Result)
	BuildDeleteAllowanceTypeModel(models.AllowanceType) (models.AllowanceType, results.Result)
	BuildSearchMonthlyAllowancesQuery(types.SearchMonthlyAllowancesRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindMonthlyAllowanceByIDQuery(uint) (*gorm.DB, results.Result)
	BuildCreateMonthlyAllowanceModel(types.CreateMonthlyAllowanceRequest) (models.MonthlyAllowance, results.Result)
	BuildUpdateMonthlyAllowanceModel(models.MonthlyAllowance, types.UpdateMonthlyAllowanceRequest) (models.MonthlyAllowance, results.Result)
	BuildDeleteMonthlyAllowanceModel(models.MonthlyAllowance) (models.MonthlyAllowance, results.Result)
}

type allowanceBuilder struct{ db *gorm.DB }

func NewAllowanceBuilder(db *gorm.DB) AllowanceBuilder { return &allowanceBuilder{db: db} }

func (b *allowanceBuilder) BuildSearchAllowanceTypesQuery(req types.SearchAllowanceTypesRequest) (*gorm.DB, *gorm.DB, results.Result) {
	q := b.db.Model(&models.AllowanceType{})
	c := b.db.Model(&models.AllowanceType{})
	if !req.IncludeDeleted {
		q = q.Where("is_deleted = ?", false)
		c = c.Where("is_deleted = ?", false)
	}
	if !req.IncludeInactive {
		q = q.Where("is_active = ?", true)
		c = c.Where("is_active = ?", true)
	}
	if s := strings.TrimSpace(req.Keyword); s != "" {
		like := "%" + s + "%"
		q = q.Where("name ILIKE ?", like)
		c = c.Where("name ILIKE ?", like)
	}
	q = q.Order("display_order ASC").Order("id ASC").Offset(req.Offset).Limit(req.Limit)
	return q, c, results.OK(nil, "BUILD_SEARCH_ALLOWANCE_TYPES_QUERY_SUCCESS", "", nil)
}
func (b *allowanceBuilder) BuildFindAllowanceTypeByIDQuery(id uint) (*gorm.DB, results.Result) {
	if id == 0 {
		return nil, results.BadRequest("EMPTY_ALLOWANCE_TYPE_ID", "対象の手当種別が指定されていません", nil)
	}
	return b.db.Model(&models.AllowanceType{}).Where("id = ?", id).Where("is_deleted = ?", false), results.OK(nil, "BUILD_FIND_ALLOWANCE_TYPE_SUCCESS", "", nil)
}
func validateType(name string, order int) results.Result {
	if strings.TrimSpace(name) == "" {
		return results.BadRequest("EMPTY_ALLOWANCE_TYPE_NAME", "手当種別名を入力してください", nil)
	}
	if order < 0 {
		return results.BadRequest("INVALID_ALLOWANCE_TYPE_ORDER", "表示順は0以上で入力してください", nil)
	}
	return results.OK(nil, "VALID", "", nil)
}
func (b *allowanceBuilder) BuildCreateAllowanceTypeModel(req types.CreateAllowanceTypeRequest) (models.AllowanceType, results.Result) {
	if r := validateType(req.Name, req.DisplayOrder); r.Error {
		return models.AllowanceType{}, r
	}
	return models.AllowanceType{Name: strings.TrimSpace(req.Name), DisplayOrder: req.DisplayOrder, IsActive: req.IsActive}, results.OK(nil, "BUILD_CREATE_ALLOWANCE_TYPE_SUCCESS", "", nil)
}
func (b *allowanceBuilder) BuildUpdateAllowanceTypeModel(m models.AllowanceType, req types.UpdateAllowanceTypeRequest) (models.AllowanceType, results.Result) {
	if r := validateType(req.Name, req.DisplayOrder); r.Error {
		return models.AllowanceType{}, r
	}
	m.Name = strings.TrimSpace(req.Name)
	m.DisplayOrder = req.DisplayOrder
	m.IsActive = req.IsActive
	return m, results.OK(nil, "BUILD_UPDATE_ALLOWANCE_TYPE_SUCCESS", "", nil)
}
func (b *allowanceBuilder) BuildDeleteAllowanceTypeModel(m models.AllowanceType) (models.AllowanceType, results.Result) {
	now := time.Now()
	m.IsActive = false
	m.IsDeleted = true
	m.DeletedAt = &now
	return m, results.OK(nil, "BUILD_DELETE_ALLOWANCE_TYPE_SUCCESS", "", nil)
}

func (b *allowanceBuilder) BuildSearchMonthlyAllowancesQuery(req types.SearchMonthlyAllowancesRequest) (*gorm.DB, *gorm.DB, results.Result) {
	q := b.db.Table("monthly_allowances").Select("monthly_allowances.*, allowance_types.name AS allowance_type_name").Joins("LEFT JOIN allowance_types ON allowance_types.id = monthly_allowances.allowance_type_id")
	c := b.db.Table("monthly_allowances")
	if !req.IncludeDeleted {
		q = q.Where("monthly_allowances.is_deleted = ?", false)
		c = c.Where("is_deleted = ?", false)
	}
	if req.TargetUserID != nil {
		q = q.Where("monthly_allowances.user_id = ?", *req.TargetUserID)
		c = c.Where("user_id = ?", *req.TargetUserID)
	}
	if req.TargetYear != nil {
		q = q.Where("monthly_allowances.target_year = ?", *req.TargetYear)
		c = c.Where("target_year = ?", *req.TargetYear)
	}
	if req.TargetMonth != nil {
		q = q.Where("monthly_allowances.target_month = ?", *req.TargetMonth)
		c = c.Where("target_month = ?", *req.TargetMonth)
	}
	if req.AllowanceTypeID != nil {
		q = q.Where("monthly_allowances.allowance_type_id = ?", *req.AllowanceTypeID)
		c = c.Where("allowance_type_id = ?", *req.AllowanceTypeID)
	}
	q = q.Order("target_year DESC").Order("target_month DESC").Order("allowance_types.display_order ASC").Order("monthly_allowances.id ASC").Offset(req.Offset).Limit(req.Limit)
	return q, c, results.OK(nil, "BUILD_SEARCH_MONTHLY_ALLOWANCES_SUCCESS", "", nil)
}
func (b *allowanceBuilder) BuildFindMonthlyAllowanceByIDQuery(id uint) (*gorm.DB, results.Result) {
	if id == 0 {
		return nil, results.BadRequest("EMPTY_MONTHLY_ALLOWANCE_ID", "対象の月次手当が指定されていません", nil)
	}
	return b.db.Model(&models.MonthlyAllowance{}).Where("id = ?", id).Where("is_deleted = ?", false), results.OK(nil, "BUILD_FIND_MONTHLY_ALLOWANCE_SUCCESS", "", nil)
}
func validateMonthly(userID uint, year, month int, typeID uint, amount int) results.Result {
	if userID == 0 {
		return results.BadRequest("EMPTY_TARGET_USER_ID", "対象ユーザーが指定されていません", nil)
	}
	if year < 2000 || year > 2100 {
		return results.BadRequest("INVALID_TARGET_YEAR", "対象年が正しくありません", nil)
	}
	if month < 1 || month > 12 {
		return results.BadRequest("INVALID_TARGET_MONTH", "対象月が正しくありません", nil)
	}
	if typeID == 0 {
		return results.BadRequest("EMPTY_ALLOWANCE_TYPE_ID", "手当種別が指定されていません", nil)
	}
	if amount < 0 {
		return results.BadRequest("INVALID_ALLOWANCE_AMOUNT", "手当金額は0円以上で入力してください", nil)
	}
	return results.OK(nil, "VALID", "", nil)
}
func (b *allowanceBuilder) BuildCreateMonthlyAllowanceModel(req types.CreateMonthlyAllowanceRequest) (models.MonthlyAllowance, results.Result) {
	if r := validateMonthly(req.TargetUserID, req.TargetYear, req.TargetMonth, req.AllowanceTypeID, req.Amount); r.Error {
		return models.MonthlyAllowance{}, r
	}
	return models.MonthlyAllowance{UserID: req.TargetUserID, TargetYear: req.TargetYear, TargetMonth: req.TargetMonth, AllowanceTypeID: req.AllowanceTypeID, Amount: req.Amount, Memo: strings.TrimSpace(req.Memo)}, results.OK(nil, "BUILD_CREATE_MONTHLY_ALLOWANCE_SUCCESS", "", nil)
}
func (b *allowanceBuilder) BuildUpdateMonthlyAllowanceModel(m models.MonthlyAllowance, req types.UpdateMonthlyAllowanceRequest) (models.MonthlyAllowance, results.Result) {
	if r := validateMonthly(req.TargetUserID, req.TargetYear, req.TargetMonth, req.AllowanceTypeID, req.Amount); r.Error {
		return models.MonthlyAllowance{}, r
	}
	m.UserID = req.TargetUserID
	m.TargetYear = req.TargetYear
	m.TargetMonth = req.TargetMonth
	m.AllowanceTypeID = req.AllowanceTypeID
	m.Amount = req.Amount
	m.Memo = strings.TrimSpace(req.Memo)
	return m, results.OK(nil, "BUILD_UPDATE_MONTHLY_ALLOWANCE_SUCCESS", "", nil)
}
func (b *allowanceBuilder) BuildDeleteMonthlyAllowanceModel(m models.MonthlyAllowance) (models.MonthlyAllowance, results.Result) {
	now := time.Now()
	m.IsDeleted = true
	m.DeletedAt = &now
	return m, results.OK(nil, "BUILD_DELETE_MONTHLY_ALLOWANCE_SUCCESS", "", nil)
}
