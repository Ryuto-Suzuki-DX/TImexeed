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
}

/*
 * 従業員用有給Repository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 */
type paidLeaveRepository struct {
	db *gorm.DB
}

/*
 * PaidLeaveRepository生成
 */
func NewPaidLeaveRepository(db *gorm.DB) PaidLeaveRepository {
	return &paidLeaveRepository{db: db}
}

/*
 * ユーザー1件取得
 *
 * 有給残数計算時の入社日取得で使う。
 */
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

	return user, results.OK(
		nil,
		"FIND_USER_PAID_LEAVE_USER_SUCCESS",
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

	return usedDays, results.OK(
		nil,
		"SUM_USER_PAID_LEAVE_USAGE_DAYS_SUCCESS",
		"",
		nil,
	)
}
