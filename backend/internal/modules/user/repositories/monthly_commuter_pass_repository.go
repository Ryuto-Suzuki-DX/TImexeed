package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type MonthlyCommuterPassRepository interface {
	FindMonthlyCommuterPass(query *gorm.DB) (models.MonthlyCommuterPass, results.Result)
	CreateMonthlyCommuterPass(monthlyCommuterPass models.MonthlyCommuterPass) (models.MonthlyCommuterPass, results.Result)
	SaveMonthlyCommuterPass(monthlyCommuterPass models.MonthlyCommuterPass) (models.MonthlyCommuterPass, results.Result)
}

/*
 * 従業員用月次通勤定期Repository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・DBへのCreate / Saveを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 * ・通勤定期の更新可否、月次申請状態チェックなどはServiceに任せる
 * ・MonthlyCommuterPass は申請状態を持たない
 */
type monthlyCommuterPassRepository struct {
	db *gorm.DB
}

/*
 * MonthlyCommuterPassRepository生成
 */
func NewMonthlyCommuterPassRepository(db *gorm.DB) MonthlyCommuterPassRepository {
	return &monthlyCommuterPassRepository{db: db}
}

/*
 * 月次通勤定期1件取得
 */
func (repository *monthlyCommuterPassRepository) FindMonthlyCommuterPass(
	query *gorm.DB,
) (models.MonthlyCommuterPass, results.Result) {
	if query == nil {
		return models.MonthlyCommuterPass{}, results.InternalServerError(
			"FIND_MONTHLY_COMMUTER_PASS_QUERY_IS_NIL",
			"月次通勤定期の取得に失敗しました",
			nil,
		)
	}

	var monthlyCommuterPass models.MonthlyCommuterPass

	if err := query.First(&monthlyCommuterPass).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.MonthlyCommuterPass{}, results.NotFound(
				"MONTHLY_COMMUTER_PASS_NOT_FOUND",
				"対象年月の通勤定期が見つかりません",
				nil,
			)
		}

		return models.MonthlyCommuterPass{}, results.InternalServerError(
			"FIND_MONTHLY_COMMUTER_PASS_FAILED",
			"月次通勤定期の取得に失敗しました",
			err.Error(),
		)
	}

	return monthlyCommuterPass, results.OK(
		nil,
		"FIND_MONTHLY_COMMUTER_PASS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次通勤定期作成
 */
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

/*
 * 月次通勤定期保存
 *
 * 更新・論理削除で使う。
 */
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
