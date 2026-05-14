package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type HolidayDateBuilder interface {
	BuildSearchHolidayDatesQuery(req types.SearchHolidayDatesRequest) (*gorm.DB, results.Result)
}

/*
 * 従業員用祝日Builder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find はRepositoryに任せる
 * ・祝日の登録、更新、削除は行わない
 * ・CSV取り込みは管理者側APIで行う
 * ・祝日マスタは全ユーザー共通
 */
type holidayDateBuilder struct {
	db *gorm.DB
}

/*
 * HolidayDateBuilder生成
 */
func NewHolidayDateBuilder(db *gorm.DB) HolidayDateBuilder {
	return &holidayDateBuilder{db: db}
}

/*
 * 祝日検索用クエリ作成
 *
 * 対象年月の祝日を取得する。
 *
 * 注意：
 * ・祝日は全ユーザー共通のため userId では絞り込まない
 * ・従業員側では参照のみ行う
 * ・論理削除済みの祝日は対象外
 */
func (builder *holidayDateBuilder) BuildSearchHolidayDatesQuery(
	req types.SearchHolidayDatesRequest,
) (*gorm.DB, results.Result) {
	if req.TargetYear <= 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_HOLIDAY_DATES_QUERY_INVALID_TARGET_YEAR",
			"祝日検索条件の作成に失敗しました",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_HOLIDAY_DATES_QUERY_INVALID_TARGET_MONTH",
			"祝日検索条件の作成に失敗しました",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	startDate := time.Date(req.TargetYear, time.Month(req.TargetMonth), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	query := builder.db.
		Model(&models.HolidayDate{}).
		Where("holiday_date >= ?", startDate).
		Where("holiday_date < ?", endDate).
		Where("is_deleted = ?", false).
		Order("holiday_date ASC").
		Order("id ASC")

	return query, results.OK(
		nil,
		"BUILD_SEARCH_HOLIDAY_DATES_QUERY_SUCCESS",
		"",
		nil,
	)
}
