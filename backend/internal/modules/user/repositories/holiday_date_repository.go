package repositories

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type HolidayDateRepository interface {
	FindHolidayDates(query *gorm.DB) ([]models.HolidayDate, results.Result)
}

/*
 * 従業員用祝日Repository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 * ・祝日の登録、更新、削除は行わない
 * ・CSV取り込みは管理者側APIで行う
 * ・祝日マスタは全ユーザー共通
 */
type holidayDateRepository struct {
	db *gorm.DB
}

/*
 * HolidayDateRepository生成
 */
func NewHolidayDateRepository(db *gorm.DB) HolidayDateRepository {
	return &holidayDateRepository{db: db}
}

/*
 * 祝日一覧取得
 */
func (repository *holidayDateRepository) FindHolidayDates(query *gorm.DB) ([]models.HolidayDate, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_HOLIDAY_DATES_QUERY_IS_NIL",
			"祝日一覧の取得に失敗しました",
			nil,
		)
	}

	var holidayDates []models.HolidayDate

	if err := query.Find(&holidayDates).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_HOLIDAY_DATES_FAILED",
			"祝日一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return holidayDates, results.OK(
		nil,
		"FIND_HOLIDAY_DATES_SUCCESS",
		"",
		nil,
	)
}
