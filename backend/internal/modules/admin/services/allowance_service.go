package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"

	"gorm.io/gorm"
)

type AllowanceService interface {
	SearchAllowanceTypes(types.SearchAllowanceTypesRequest) results.Result
	GetAllowanceTypeDetail(types.AllowanceTypeDetailRequest) results.Result
	CreateAllowanceType(types.CreateAllowanceTypeRequest) results.Result
	UpdateAllowanceType(types.UpdateAllowanceTypeRequest) results.Result
	DeleteAllowanceType(types.DeleteAllowanceTypeRequest) results.Result
	SearchMonthlyAllowances(types.SearchMonthlyAllowancesRequest) results.Result
	GetMonthlyAllowanceDetail(types.MonthlyAllowanceDetailRequest) results.Result
	CreateMonthlyAllowance(types.CreateMonthlyAllowanceRequest) results.Result
	UpdateMonthlyAllowance(types.UpdateMonthlyAllowanceRequest) results.Result
	DeleteMonthlyAllowance(types.DeleteMonthlyAllowanceRequest) results.Result
}

type allowanceService struct {
	db         *gorm.DB
	builder    builders.AllowanceBuilder
	repository repositories.AllowanceRepository
}

func NewAllowanceService(db *gorm.DB, b builders.AllowanceBuilder, r repositories.AllowanceRepository) *allowanceService {
	return &allowanceService{db: db, builder: b, repository: r}
}
func typeResponse(v models.AllowanceType) types.AllowanceTypeResponse {
	return types.AllowanceTypeResponse{ID: v.ID, Name: v.Name, DisplayOrder: v.DisplayOrder, IsActive: v.IsActive, IsDeleted: v.IsDeleted, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, DeletedAt: v.DeletedAt}
}
func monthlyResponse(v models.MonthlyAllowance, name string) types.MonthlyAllowanceResponse {
	return types.MonthlyAllowanceResponse{ID: v.ID, UserID: v.UserID, TargetYear: v.TargetYear, TargetMonth: v.TargetMonth, AllowanceTypeID: v.AllowanceTypeID, AllowanceTypeName: name, Amount: v.Amount, Memo: v.Memo, IsDeleted: v.IsDeleted, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, DeletedAt: v.DeletedAt}
}

func (s *allowanceService) SearchAllowanceTypes(req types.SearchAllowanceTypesRequest) results.Result {
	n, nr := utils.NormalizePageSearchCondition(utils.PageSearchCondition{Keyword: req.Keyword, Offset: req.Offset, Limit: req.Limit}, "SEARCH_ALLOWANCE_TYPES_INVALID_OFFSET", "検索開始位置が正しくありません")
	if nr.Error {
		return nr
	}
	req.Keyword = n.Keyword
	req.Offset = n.Offset
	req.Limit = n.Limit
	q, c, br := s.builder.BuildSearchAllowanceTypesQuery(req)
	if br.Error {
		return br
	}
	rows, rr := s.repository.FindAllowanceTypes(q)
	if rr.Error {
		return rr
	}
	total, cr := s.repository.Count(c)
	if cr.Error {
		return cr
	}
	out := make([]types.AllowanceTypeResponse, 0, len(rows))
	for _, v := range rows {
		out = append(out, typeResponse(v))
	}
	return results.OK(types.SearchAllowanceTypesResponse{AllowanceTypes: out, Total: total, Offset: req.Offset, Limit: req.Limit, HasMore: utils.HasMore(total, req.Offset, len(rows))}, "SEARCH_ALLOWANCE_TYPES_SUCCESS", "手当種別一覧を取得しました", nil)
}
func (s *allowanceService) GetAllowanceTypeDetail(req types.AllowanceTypeDetailRequest) results.Result {
	q, br := s.builder.BuildFindAllowanceTypeByIDQuery(req.AllowanceTypeID)
	if br.Error {
		return br
	}
	v, rr := s.repository.FindAllowanceType(q)
	if rr.Error {
		return rr
	}
	return results.OK(types.AllowanceTypeDetailResponse{AllowanceType: typeResponse(v)}, "GET_ALLOWANCE_TYPE_DETAIL_SUCCESS", "手当種別詳細を取得しました", nil)
}
func (s *allowanceService) CreateAllowanceType(req types.CreateAllowanceTypeRequest) results.Result {
	var count int64
	if err := s.db.Model(&models.AllowanceType{}).Where("name = ? AND is_deleted = ?", req.Name, false).Count(&count).Error; err != nil {
		return results.InternalServerError("CHECK_ALLOWANCE_TYPE_NAME_FAILED", "手当種別名の確認に失敗しました", err.Error())
	}
	if count > 0 {
		return results.Conflict("ALLOWANCE_TYPE_NAME_EXISTS", "同じ名前の手当種別が既に登録されています", nil)
	}
	v, br := s.builder.BuildCreateAllowanceTypeModel(req)
	if br.Error {
		return br
	}
	v, rr := s.repository.CreateAllowanceType(v)
	if rr.Error {
		return rr
	}
	return results.Created(types.CreateAllowanceTypeResponse{AllowanceType: typeResponse(v)}, "CREATE_ALLOWANCE_TYPE_SUCCESS", "手当種別を作成しました", nil)
}
func (s *allowanceService) UpdateAllowanceType(req types.UpdateAllowanceTypeRequest) results.Result {
	q, br := s.builder.BuildFindAllowanceTypeByIDQuery(req.AllowanceTypeID)
	if br.Error {
		return br
	}
	v, rr := s.repository.FindAllowanceType(q)
	if rr.Error {
		return rr
	}
	var count int64
	if err := s.db.Model(&models.AllowanceType{}).Where("name = ? AND id <> ? AND is_deleted = ?", req.Name, req.AllowanceTypeID, false).Count(&count).Error; err != nil {
		return results.InternalServerError("CHECK_ALLOWANCE_TYPE_NAME_FAILED", "手当種別名の確認に失敗しました", err.Error())
	}
	if count > 0 {
		return results.Conflict("ALLOWANCE_TYPE_NAME_EXISTS", "同じ名前の手当種別が既に登録されています", nil)
	}
	v, br = s.builder.BuildUpdateAllowanceTypeModel(v, req)
	if br.Error {
		return br
	}
	v, rr = s.repository.SaveAllowanceType(v)
	if rr.Error {
		return rr
	}
	return results.OK(types.UpdateAllowanceTypeResponse{AllowanceType: typeResponse(v)}, "UPDATE_ALLOWANCE_TYPE_SUCCESS", "手当種別を更新しました", nil)
}
func (s *allowanceService) DeleteAllowanceType(req types.DeleteAllowanceTypeRequest) results.Result {
	q, br := s.builder.BuildFindAllowanceTypeByIDQuery(req.AllowanceTypeID)
	if br.Error {
		return br
	}
	v, rr := s.repository.FindAllowanceType(q)
	if rr.Error {
		return rr
	}
	v, br = s.builder.BuildDeleteAllowanceTypeModel(v)
	if br.Error {
		return br
	}
	_, rr = s.repository.SaveAllowanceType(v)
	if rr.Error {
		return rr
	}
	return results.OK(types.DeleteAllowanceTypeResponse{AllowanceTypeID: req.AllowanceTypeID}, "DELETE_ALLOWANCE_TYPE_SUCCESS", "手当種別を削除しました", nil)
}

func (s *allowanceService) SearchMonthlyAllowances(req types.SearchMonthlyAllowancesRequest) results.Result {
	n, nr := utils.NormalizePageSearchCondition(utils.PageSearchCondition{Offset: req.Offset, Limit: req.Limit}, "SEARCH_MONTHLY_ALLOWANCES_INVALID_OFFSET", "検索開始位置が正しくありません")
	if nr.Error {
		return nr
	}
	req.Offset = n.Offset
	req.Limit = n.Limit
	q, c, br := s.builder.BuildSearchMonthlyAllowancesQuery(req)
	if br.Error {
		return br
	}
	rows, rr := s.repository.FindMonthlyAllowances(q)
	if rr.Error {
		return rr
	}
	total, cr := s.repository.Count(c)
	if cr.Error {
		return cr
	}
	return results.OK(types.SearchMonthlyAllowancesResponse{MonthlyAllowances: rows, Total: total, Offset: req.Offset, Limit: req.Limit, HasMore: utils.HasMore(total, req.Offset, len(rows))}, "SEARCH_MONTHLY_ALLOWANCES_SUCCESS", "月次手当一覧を取得しました", nil)
}
func (s *allowanceService) GetMonthlyAllowanceDetail(req types.MonthlyAllowanceDetailRequest) results.Result {
	q, br := s.builder.BuildFindMonthlyAllowanceByIDQuery(req.MonthlyAllowanceID)
	if br.Error {
		return br
	}
	v, rr := s.repository.FindMonthlyAllowanceModel(q)
	if rr.Error {
		return rr
	}
	var t models.AllowanceType
	s.db.First(&t, v.AllowanceTypeID)
	return results.OK(types.MonthlyAllowanceDetailResponse{MonthlyAllowance: monthlyResponse(v, t.Name)}, "GET_MONTHLY_ALLOWANCE_DETAIL_SUCCESS", "月次手当詳細を取得しました", nil)
}
func (s *allowanceService) validateRefs(userID, typeID uint) results.Result {
	var users int64
	if err := s.db.Model(&models.User{}).Where("id = ? AND role = ? AND is_deleted = ?", userID, "USER", false).Count(&users).Error; err != nil {
		return results.InternalServerError("CHECK_TARGET_USER_FAILED", "対象ユーザーの確認に失敗しました", err.Error())
	}
	if users == 0 {
		return results.NotFound("TARGET_USER_NOT_FOUND", "対象ユーザーが見つかりません", nil)
	}
	var t models.AllowanceType
	if err := s.db.Where("id = ? AND is_deleted = ?", typeID, false).First(&t).Error; err != nil {
		return results.NotFound("ALLOWANCE_TYPE_NOT_FOUND", "対象の手当種別が見つかりません", nil)
	}
	if !t.IsActive {
		return results.BadRequest("ALLOWANCE_TYPE_INACTIVE", "利用停止中の手当種別は選択できません", nil)
	}
	return results.OK(nil, "VALID", "", nil)
}
func (s *allowanceService) CreateMonthlyAllowance(req types.CreateMonthlyAllowanceRequest) results.Result {
	if r := s.validateRefs(req.TargetUserID, req.AllowanceTypeID); r.Error {
		return r
	}
	v, br := s.builder.BuildCreateMonthlyAllowanceModel(req)
	if br.Error {
		return br
	}
	v, rr := s.repository.CreateMonthlyAllowance(v)
	if rr.Error {
		return rr
	}
	var t models.AllowanceType
	s.db.First(&t, v.AllowanceTypeID)
	return results.Created(types.CreateMonthlyAllowanceResponse{MonthlyAllowance: monthlyResponse(v, t.Name)}, "CREATE_MONTHLY_ALLOWANCE_SUCCESS", "月次手当を作成しました", nil)
}
func (s *allowanceService) UpdateMonthlyAllowance(req types.UpdateMonthlyAllowanceRequest) results.Result {
	q, br := s.builder.BuildFindMonthlyAllowanceByIDQuery(req.MonthlyAllowanceID)
	if br.Error {
		return br
	}
	v, rr := s.repository.FindMonthlyAllowanceModel(q)
	if rr.Error {
		return rr
	}
	if r := s.validateRefs(req.TargetUserID, req.AllowanceTypeID); r.Error {
		return r
	}
	v, br = s.builder.BuildUpdateMonthlyAllowanceModel(v, req)
	if br.Error {
		return br
	}
	v, rr = s.repository.SaveMonthlyAllowance(v)
	if rr.Error {
		return rr
	}
	var t models.AllowanceType
	s.db.First(&t, v.AllowanceTypeID)
	return results.OK(types.UpdateMonthlyAllowanceResponse{MonthlyAllowance: monthlyResponse(v, t.Name)}, "UPDATE_MONTHLY_ALLOWANCE_SUCCESS", "月次手当を更新しました", nil)
}
func (s *allowanceService) DeleteMonthlyAllowance(req types.DeleteMonthlyAllowanceRequest) results.Result {
	q, br := s.builder.BuildFindMonthlyAllowanceByIDQuery(req.MonthlyAllowanceID)
	if br.Error {
		return br
	}
	v, rr := s.repository.FindMonthlyAllowanceModel(q)
	if rr.Error {
		return rr
	}
	v, br = s.builder.BuildDeleteMonthlyAllowanceModel(v)
	if br.Error {
		return br
	}
	_, rr = s.repository.SaveMonthlyAllowance(v)
	if rr.Error {
		return rr
	}
	return results.OK(types.DeleteMonthlyAllowanceResponse{MonthlyAllowanceID: req.MonthlyAllowanceID}, "DELETE_MONTHLY_ALLOWANCE_SUCCESS", "月次手当を削除しました", nil)
}
