package repositories

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type HolidayDateRepository interface {
	FindHolidayDates(query *gorm.DB) ([]models.HolidayDate, results.Result)
	DeleteAllHolidayDates() results.Result
	CreateHolidayDates(holidayDates []models.HolidayDate) ([]models.HolidayDate, results.Result)
}

/*
 * 管理者用祝日Repository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・既存の祝日マスタを物理削除する
 * ・CSVから作成した祝日マスタを一括登録する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 * ・CSV解析はBuilderに任せる
 * ・祝日CSVインポートは、既存データを全削除してから全投入する
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

/*
 * 祝日全件物理削除
 *
 * CSV再取り込み時に使う。
 *
 * 注意：
 * ・論理削除ではなく物理削除する
 * ・祝日マスタは外部CSVから再生成する参照マスタなので、履歴は持たない
 */
func (repository *holidayDateRepository) DeleteAllHolidayDates() results.Result {
	if err := repository.db.Unscoped().Where("1 = 1").Delete(&models.HolidayDate{}).Error; err != nil {
		return results.InternalServerError(
			"DELETE_ALL_HOLIDAY_DATES_FAILED",
			"既存の祝日データの削除に失敗しました",
			err.Error(),
		)
	}

	return results.OK(
		nil,
		"DELETE_ALL_HOLIDAY_DATES_SUCCESS",
		"",
		nil,
	)
}

/*
 * 祝日一括作成
 *
 * CSVから作成した祝日データを全件登録する。
 */
func (repository *holidayDateRepository) CreateHolidayDates(
	holidayDates []models.HolidayDate,
) ([]models.HolidayDate, results.Result) {
	if len(holidayDates) == 0 {
		return nil, results.InternalServerError(
			"CREATE_HOLIDAY_DATES_EMPTY_DATA",
			"祝日データの登録に失敗しました",
			nil,
		)
	}

	if err := repository.db.Create(&holidayDates).Error; err != nil {
		return nil, results.InternalServerError(
			"CREATE_HOLIDAY_DATES_FAILED",
			"祝日データの登録に失敗しました",
			err.Error(),
		)
	}

	return holidayDates, results.OK(
		nil,
		"CREATE_HOLIDAY_DATES_SUCCESS",
		"",
		nil,
	)
}
