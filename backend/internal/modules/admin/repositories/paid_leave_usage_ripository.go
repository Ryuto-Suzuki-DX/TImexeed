package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type PaidLeaveUsageRepository interface {
	FindUser(query *gorm.DB) (models.User, results.Result)
	FindPaidLeaveUsages(query *gorm.DB) ([]models.PaidLeaveUsage, results.Result)
	CountPaidLeaveUsages(query *gorm.DB) (int64, results.Result)
	SumPaidLeaveUsageDays(query *gorm.DB) (float64, results.Result)
	CreatePaidLeaveUsage(paidLeaveUsage models.PaidLeaveUsage) (models.PaidLeaveUsage, results.Result)
	FindPaidLeaveUsage(query *gorm.DB) (models.PaidLeaveUsage, results.Result)
	SavePaidLeaveUsage(paidLeaveUsage models.PaidLeaveUsage) (models.PaidLeaveUsage, results.Result)
}

/*
 * 管理者用有給使用日Repository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・DBへのCreate / Saveを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 */
type paidLeaveUsageRepository struct {
	db *gorm.DB
}

/*
 * PaidLeaveUsageRepository生成
 */
func NewPaidLeaveUsageRepository(db *gorm.DB) PaidLeaveUsageRepository {
	return &paidLeaveUsageRepository{db: db}
}

/*
 * ユーザー1件取得
 *
 * 有給使用日の対象ユーザー存在確認や、
 * 有給残数計算時の入社日取得で使う。
 */
func (repository *paidLeaveUsageRepository) FindUser(query *gorm.DB) (models.User, results.Result) {
	if query == nil {
		return models.User{}, results.InternalServerError(
			"FIND_PAID_LEAVE_USAGE_USER_QUERY_IS_NIL",
			"対象ユーザーの取得に失敗しました",
			nil,
		)
	}

	var user models.User

	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, results.NotFound(
				"PAID_LEAVE_USAGE_USER_NOT_FOUND",
				"対象ユーザーが見つかりません",
				nil,
			)
		}

		return models.User{}, results.InternalServerError(
			"FIND_PAID_LEAVE_USAGE_USER_FAILED",
			"対象ユーザーの取得に失敗しました",
			err.Error(),
		)
	}

	return user, results.OK(
		nil,
		"FIND_PAID_LEAVE_USAGE_USER_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給使用日一覧取得
 */
func (repository *paidLeaveUsageRepository) FindPaidLeaveUsages(query *gorm.DB) ([]models.PaidLeaveUsage, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_PAID_LEAVE_USAGES_QUERY_IS_NIL",
			"有給使用日一覧の取得に失敗しました",
			nil,
		)
	}

	var paidLeaveUsages []models.PaidLeaveUsage

	if err := query.Find(&paidLeaveUsages).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_PAID_LEAVE_USAGES_FAILED",
			"有給使用日一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return paidLeaveUsages, results.OK(
		nil,
		"FIND_PAID_LEAVE_USAGES_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給使用日件数取得
 */
func (repository *paidLeaveUsageRepository) CountPaidLeaveUsages(query *gorm.DB) (int64, results.Result) {
	if query == nil {
		return 0, results.InternalServerError(
			"COUNT_PAID_LEAVE_USAGES_QUERY_IS_NIL",
			"有給使用日件数の取得に失敗しました",
			nil,
		)
	}

	var count int64

	if err := query.Count(&count).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_PAID_LEAVE_USAGES_FAILED",
			"有給使用日件数の取得に失敗しました",
			err.Error(),
		)
	}

	return count, results.OK(
		nil,
		"COUNT_PAID_LEAVE_USAGES_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給使用日数合計取得
 *
 * 有給残数計算で使う。
 *
 * 注意：
 * ・検索条件はBuilderで作成する
 * ・ここでは usage_days の合計だけを取得する
 * ・該当データがない場合は 0 を返す
 */
func (repository *paidLeaveUsageRepository) SumPaidLeaveUsageDays(query *gorm.DB) (float64, results.Result) {
	if query == nil {
		return 0, results.InternalServerError(
			"SUM_PAID_LEAVE_USAGE_DAYS_QUERY_IS_NIL",
			"有給使用日数合計の取得に失敗しました",
			nil,
		)
	}

	var usedDays float64

	if err := query.Select("COALESCE(SUM(usage_days), 0)").Scan(&usedDays).Error; err != nil {
		return 0, results.InternalServerError(
			"SUM_PAID_LEAVE_USAGE_DAYS_FAILED",
			"有給使用日数合計の取得に失敗しました",
			err.Error(),
		)
	}

	return usedDays, results.OK(
		nil,
		"SUM_PAID_LEAVE_USAGE_DAYS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給使用日作成
 */
func (repository *paidLeaveUsageRepository) CreatePaidLeaveUsage(
	paidLeaveUsage models.PaidLeaveUsage,
) (models.PaidLeaveUsage, results.Result) {
	if err := repository.db.Create(&paidLeaveUsage).Error; err != nil {
		return models.PaidLeaveUsage{}, results.InternalServerError(
			"CREATE_PAID_LEAVE_USAGE_FAILED",
			"有給使用日の作成に失敗しました",
			err.Error(),
		)
	}

	return paidLeaveUsage, results.OK(
		nil,
		"CREATE_PAID_LEAVE_USAGE_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給使用日1件取得
 *
 * 更新・削除対象の取得で使う。
 */
func (repository *paidLeaveUsageRepository) FindPaidLeaveUsage(
	query *gorm.DB,
) (models.PaidLeaveUsage, results.Result) {
	if query == nil {
		return models.PaidLeaveUsage{}, results.InternalServerError(
			"FIND_PAID_LEAVE_USAGE_QUERY_IS_NIL",
			"有給使用日の取得に失敗しました",
			nil,
		)
	}

	var paidLeaveUsage models.PaidLeaveUsage

	if err := query.First(&paidLeaveUsage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PaidLeaveUsage{}, results.NotFound(
				"PAID_LEAVE_USAGE_NOT_FOUND",
				"対象の有給使用日が見つかりません",
				nil,
			)
		}

		return models.PaidLeaveUsage{}, results.InternalServerError(
			"FIND_PAID_LEAVE_USAGE_FAILED",
			"有給使用日の取得に失敗しました",
			err.Error(),
		)
	}

	return paidLeaveUsage, results.OK(
		nil,
		"FIND_PAID_LEAVE_USAGE_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給使用日保存
 *
 * 更新・論理削除で使う。
 */
func (repository *paidLeaveUsageRepository) SavePaidLeaveUsage(
	paidLeaveUsage models.PaidLeaveUsage,
) (models.PaidLeaveUsage, results.Result) {
	if paidLeaveUsage.ID == 0 {
		return models.PaidLeaveUsage{}, results.InternalServerError(
			"SAVE_PAID_LEAVE_USAGE_EMPTY_ID",
			"有給使用日の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&paidLeaveUsage).Error; err != nil {
		return models.PaidLeaveUsage{}, results.InternalServerError(
			"SAVE_PAID_LEAVE_USAGE_FAILED",
			"有給使用日の保存に失敗しました",
			err.Error(),
		)
	}

	return paidLeaveUsage, results.OK(
		nil,
		"SAVE_PAID_LEAVE_USAGE_SUCCESS",
		"",
		nil,
	)
}
