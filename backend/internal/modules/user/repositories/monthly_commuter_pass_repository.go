package repositories

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type MonthlyCommuterPassRepository interface {
	FindMonthlyCommuterPasses(query *gorm.DB) ([]models.MonthlyCommuterPass, results.Result)
	CreateMonthlyCommuterPass(monthlyCommuterPass models.MonthlyCommuterPass) (models.MonthlyCommuterPass, results.Result)
	SaveMonthlyCommuterPass(monthlyCommuterPass models.MonthlyCommuterPass) (models.MonthlyCommuterPass, results.Result)
}

/*
 * 従業員用月次通勤定期Repository
 *
 * 対象ユーザー・対象年月の通勤定期を複数件扱う。
 */
type monthlyCommuterPassRepository struct {
	db *gorm.DB
}

func NewMonthlyCommuterPassRepository(db *gorm.DB) MonthlyCommuterPassRepository {
	return &monthlyCommuterPassRepository{db: db}
}

/*
 * 月次通勤定期一覧取得
 *
 * 0件の場合も空配列で成功を返す。
 */
func (repository *monthlyCommuterPassRepository) FindMonthlyCommuterPasses(
	query *gorm.DB,
) ([]models.MonthlyCommuterPass, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_MONTHLY_COMMUTER_PASSES_QUERY_IS_NIL",
			"月次通勤定期の取得に失敗しました",
			nil,
		)
	}

	monthlyCommuterPasses := make([]models.MonthlyCommuterPass, 0)

	if err := query.Order("id ASC").Find(&monthlyCommuterPasses).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_MONTHLY_COMMUTER_PASSES_FAILED",
			"月次通勤定期の取得に失敗しました",
			err.Error(),
		)
	}

	return monthlyCommuterPasses, results.OK(
		nil,
		"FIND_MONTHLY_COMMUTER_PASSES_SUCCESS",
		"",
		nil,
	)
}

func (repository *monthlyCommuterPassRepository) CreateMonthlyCommuterPass(
	monthlyCommuterPass models.MonthlyCommuterPass,
) (models.MonthlyCommuterPass, results.Result) {
	if err := repository.db.Create(&monthlyCommuterPass).Error; err != nil {
		return models.MonthlyCommuterPass{}, results.InternalServerError(
			"CREATE_MONTHLY_COMMUTER_PASS_FAILED",
			"月次通勤定期の作成に失敗しました",
			err.Error(),
		)
	}

	return monthlyCommuterPass, results.OK(
		nil,
		"CREATE_MONTHLY_COMMUTER_PASS_SUCCESS",
		"",
		nil,
	)
}

func (repository *monthlyCommuterPassRepository) SaveMonthlyCommuterPass(
	monthlyCommuterPass models.MonthlyCommuterPass,
) (models.MonthlyCommuterPass, results.Result) {
	if monthlyCommuterPass.ID == 0 {
		return models.MonthlyCommuterPass{}, results.InternalServerError(
			"SAVE_MONTHLY_COMMUTER_PASS_EMPTY_ID",
			"月次通勤定期情報の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&monthlyCommuterPass).Error; err != nil {
		return models.MonthlyCommuterPass{}, results.InternalServerError(
			"SAVE_MONTHLY_COMMUTER_PASS_FAILED",
			"月次通勤定期情報の保存に失敗しました",
			err.Error(),
		)
	}

	return monthlyCommuterPass, results.OK(
		nil,
		"SAVE_MONTHLY_COMMUTER_PASS_SUCCESS",
		"",
		nil,
	)
}
