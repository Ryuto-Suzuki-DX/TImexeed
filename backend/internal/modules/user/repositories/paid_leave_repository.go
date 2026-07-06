package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type PaidLeaveRepository interface {
	FindUser(query *gorm.DB) (models.User, results.Result)
	SumPaidLeaveUsageDays(query *gorm.DB) (float64, results.Result)
	FindPaidLeaveUsage(query *gorm.DB) (models.PaidLeaveUsage, results.Result)
	CreatePaidLeaveUsage(paidLeaveUsage models.PaidLeaveUsage) (models.PaidLeaveUsage, results.Result)
	SavePaidLeaveUsage(paidLeaveUsage models.PaidLeaveUsage) (models.PaidLeaveUsage, results.Result)
}

/*
 * 従業員用有給Repository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・有給使用履歴のCreate / Saveを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 */
type paidLeaveRepository struct {
	db *gorm.DB
}

func NewPaidLeaveRepository(db *gorm.DB) PaidLeaveRepository {
	return &paidLeaveRepository{db: db}
}

func (repository *paidLeaveRepository) FindUser(query *gorm.DB) (models.User, results.Result) {
	if query == nil {
		return models.User{}, results.InternalServerError(
			"FIND_USER_PAID_LEAVE_USER_QUERY_IS_NIL",
			"ユーザー情報の取得に失敗しました",
			nil,
		)
	}

	var user models.User
	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, results.NotFound(
				"USER_PAID_LEAVE_USER_NOT_FOUND",
				"対象ユーザーが見つかりません",
				nil,
			)
		}

		return models.User{}, results.InternalServerError(
			"FIND_USER_PAID_LEAVE_USER_FAILED",
			"ユーザー情報の取得に失敗しました",
			err.Error(),
		)
	}

	return user, results.OK(nil, "FIND_USER_PAID_LEAVE_USER_SUCCESS", "", nil)
}

func (repository *paidLeaveRepository) SumPaidLeaveUsageDays(query *gorm.DB) (float64, results.Result) {
	if query == nil {
		return 0, results.InternalServerError(
			"SUM_USER_PAID_LEAVE_USAGE_DAYS_QUERY_IS_NIL",
			"有給使用日数合計の取得に失敗しました",
			nil,
		)
	}

	var usedDays float64
	if err := query.Select("COALESCE(SUM(usage_days), 0)").Scan(&usedDays).Error; err != nil {
		return 0, results.InternalServerError(
			"SUM_USER_PAID_LEAVE_USAGE_DAYS_FAILED",
			"有給使用日数合計の取得に失敗しました",
			err.Error(),
		)
	}

	return usedDays, results.OK(nil, "SUM_USER_PAID_LEAVE_USAGE_DAYS_SUCCESS", "", nil)
}

func (repository *paidLeaveRepository) FindPaidLeaveUsage(
	query *gorm.DB,
) (models.PaidLeaveUsage, results.Result) {
	if query == nil {
		return models.PaidLeaveUsage{}, results.InternalServerError(
			"FIND_USER_PAID_LEAVE_USAGE_QUERY_IS_NIL",
			"有給使用日の取得に失敗しました",
			nil,
		)
	}

	var paidLeaveUsage models.PaidLeaveUsage
	if err := query.First(&paidLeaveUsage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PaidLeaveUsage{}, results.NotFound(
				"USER_PAID_LEAVE_USAGE_NOT_FOUND",
				"対象の有給使用日が見つかりません",
				nil,
			)
		}

		return models.PaidLeaveUsage{}, results.InternalServerError(
			"FIND_USER_PAID_LEAVE_USAGE_FAILED",
			"有給使用日の取得に失敗しました",
			err.Error(),
		)
	}

	return paidLeaveUsage, results.OK(nil, "FIND_USER_PAID_LEAVE_USAGE_SUCCESS", "", nil)
}

func (repository *paidLeaveRepository) CreatePaidLeaveUsage(
	paidLeaveUsage models.PaidLeaveUsage,
) (models.PaidLeaveUsage, results.Result) {
	if err := repository.db.Create(&paidLeaveUsage).Error; err != nil {
		return models.PaidLeaveUsage{}, results.InternalServerError(
			"CREATE_USER_PAID_LEAVE_USAGE_FAILED",
			"有給使用日の作成に失敗しました",
			err.Error(),
		)
	}

	return paidLeaveUsage, results.OK(nil, "CREATE_USER_PAID_LEAVE_USAGE_SUCCESS", "", nil)
}

func (repository *paidLeaveRepository) SavePaidLeaveUsage(
	paidLeaveUsage models.PaidLeaveUsage,
) (models.PaidLeaveUsage, results.Result) {
	if paidLeaveUsage.ID == 0 {
		return models.PaidLeaveUsage{}, results.InternalServerError(
			"SAVE_USER_PAID_LEAVE_USAGE_EMPTY_ID",
			"有給使用日の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&paidLeaveUsage).Error; err != nil {
		return models.PaidLeaveUsage{}, results.InternalServerError(
			"SAVE_USER_PAID_LEAVE_USAGE_FAILED",
			"有給使用日の保存に失敗しました",
			err.Error(),
		)
	}

	return paidLeaveUsage, results.OK(nil, "SAVE_USER_PAID_LEAVE_USAGE_SUCCESS", "", nil)
}
