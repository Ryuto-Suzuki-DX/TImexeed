package repositories

import (
	"errors"
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type AllowanceRepository interface {
	FindAllowanceTypes(*gorm.DB) ([]models.AllowanceType, results.Result)
	Count(*gorm.DB) (int64, results.Result)
	FindAllowanceType(*gorm.DB) (models.AllowanceType, results.Result)
	CreateAllowanceType(models.AllowanceType) (models.AllowanceType, results.Result)
	SaveAllowanceType(models.AllowanceType) (models.AllowanceType, results.Result)
	FindMonthlyAllowances(*gorm.DB) ([]types.MonthlyAllowanceResponse, results.Result)
	FindMonthlyAllowanceModel(*gorm.DB) (models.MonthlyAllowance, results.Result)
	CreateMonthlyAllowance(models.MonthlyAllowance) (models.MonthlyAllowance, results.Result)
	SaveMonthlyAllowance(models.MonthlyAllowance) (models.MonthlyAllowance, results.Result)
}
type allowanceRepository struct{ db *gorm.DB }

func NewAllowanceRepository(db *gorm.DB) AllowanceRepository { return &allowanceRepository{db: db} }

func (r *allowanceRepository) Count(q *gorm.DB) (int64, results.Result) {
	if q == nil {
		return 0, results.InternalServerError("COUNT_QUERY_NIL", "件数の取得に失敗しました", nil)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return 0, results.InternalServerError("COUNT_FAILED", "件数の取得に失敗しました", err.Error())
	}
	return n, results.OK(nil, "COUNT_SUCCESS", "", nil)
}

func (r *allowanceRepository) FindAllowanceTypes(q *gorm.DB) ([]models.AllowanceType, results.Result) {
	var v []models.AllowanceType
	if err := q.Find(&v).Error; err != nil {
		return nil, results.InternalServerError("FIND_ALLOWANCE_TYPES_FAILED", "手当種別一覧の取得に失敗しました", err.Error())
	}
	return v, results.OK(nil, "FIND_ALLOWANCE_TYPES_SUCCESS", "", nil)
}

func (r *allowanceRepository) FindAllowanceType(q *gorm.DB) (models.AllowanceType, results.Result) {
	var v models.AllowanceType
	if err := q.First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return v, results.NotFound("ALLOWANCE_TYPE_NOT_FOUND", "対象の手当種別が見つかりません", nil)
		}
		return v, results.InternalServerError("FIND_ALLOWANCE_TYPE_FAILED", "手当種別の取得に失敗しました", err.Error())
	}
	return v, results.OK(nil, "FIND_ALLOWANCE_TYPE_SUCCESS", "", nil)
}

func (r *allowanceRepository) CreateAllowanceType(v models.AllowanceType) (models.AllowanceType, results.Result) {
	if err := r.db.Create(&v).Error; err != nil {
		return v, results.InternalServerError("CREATE_ALLOWANCE_TYPE_FAILED", "手当種別の作成に失敗しました", err.Error())
	}
	return v, results.OK(nil, "CREATE_ALLOWANCE_TYPE_SUCCESS", "", nil)
}

func (r *allowanceRepository) SaveAllowanceType(v models.AllowanceType) (models.AllowanceType, results.Result) {
	if err := r.db.Save(&v).Error; err != nil {
		return v, results.InternalServerError("SAVE_ALLOWANCE_TYPE_FAILED", "手当種別の保存に失敗しました", err.Error())
	}
	return v, results.OK(nil, "SAVE_ALLOWANCE_TYPE_SUCCESS", "", nil)
}

func (r *allowanceRepository) FindMonthlyAllowances(q *gorm.DB) ([]types.MonthlyAllowanceResponse, results.Result) {
	var v []types.MonthlyAllowanceResponse
	if err := q.Scan(&v).Error; err != nil {
		return nil, results.InternalServerError("FIND_MONTHLY_ALLOWANCES_FAILED", "月次手当一覧の取得に失敗しました", err.Error())
	}
	return v, results.OK(nil, "FIND_MONTHLY_ALLOWANCES_SUCCESS", "", nil)
}

func (r *allowanceRepository) FindMonthlyAllowanceModel(q *gorm.DB) (models.MonthlyAllowance, results.Result) {
	var v models.MonthlyAllowance
	if err := q.First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return v, results.NotFound("MONTHLY_ALLOWANCE_NOT_FOUND", "対象の月次手当が見つかりません", nil)
		}
		return v, results.InternalServerError("FIND_MONTHLY_ALLOWANCE_FAILED", "月次手当の取得に失敗しました", err.Error())
	}
	return v, results.OK(nil, "FIND_MONTHLY_ALLOWANCE_SUCCESS", "", nil)
}

func (r *allowanceRepository) CreateMonthlyAllowance(v models.MonthlyAllowance) (models.MonthlyAllowance, results.Result) {
	if err := r.db.Create(&v).Error; err != nil {
		return v, results.InternalServerError("CREATE_MONTHLY_ALLOWANCE_FAILED", "月次手当の作成に失敗しました", err.Error())
	}
	return v, results.OK(nil, "CREATE_MONTHLY_ALLOWANCE_SUCCESS", "", nil)
}

func (r *allowanceRepository) SaveMonthlyAllowance(v models.MonthlyAllowance) (models.MonthlyAllowance, results.Result) {
	if err := r.db.Save(&v).Error; err != nil {
		return v, results.InternalServerError("SAVE_MONTHLY_ALLOWANCE_FAILED", "月次手当の保存に失敗しました", err.Error())
	}
	return v, results.OK(nil, "SAVE_MONTHLY_ALLOWANCE_SUCCESS", "", nil)
}
