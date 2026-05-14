package builders

import (
	"encoding/csv"
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type HolidayDateBuilder interface {
	BuildCreateHolidayDateModels(req types.ImportHolidayDatesRequest) ([]models.HolidayDate, int, results.Result)
	BuildSearchHolidayDatesQuery(req types.SearchHolidayDatesRequest) (*gorm.DB, results.Result)
}

/*
 * 管理者用祝日Builder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Serviceから受け取ったCSV文字列を解析してDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Delete / Create はRepositoryに任せる
 * ・祝日CSV取り込みは管理者側だけで行う
 * ・インポート時は既存holiday_datesを全削除してから全投入する方針
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
 * 祝日CSVから作成用Model一覧を作る
 *
 * 想定CSV：
 * 国民の祝日・休日月日,国民の祝日・休日名称
 * 2026/1/1,元日
 * 2026/1/12,成人の日
 *
 * 対応日付形式：
 * ・yyyy/M/d
 * ・yyyy/MM/dd
 * ・yyyy-M-d
 * ・yyyy-MM-dd
 *
 * 注意：
 * ・ヘッダー行、空行、不正行はスキップする
 * ・CSV内で同じ日付が重複した場合は、最初の1件を採用し、以降はスキップする
 */
func (builder *holidayDateBuilder) BuildCreateHolidayDateModels(
	req types.ImportHolidayDatesRequest,
) ([]models.HolidayDate, int, results.Result) {
	csvText := strings.TrimSpace(removeUTF8BOM(req.CsvText))
	if csvText == "" {
		return nil, 0, results.BadRequest(
			"BUILD_CREATE_HOLIDAY_DATE_MODELS_EMPTY_CSV_TEXT",
			"祝日CSVデータの作成に失敗しました",
			nil,
		)
	}

	reader := csv.NewReader(strings.NewReader(csvText))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, 0, results.BadRequest(
			"BUILD_CREATE_HOLIDAY_DATE_MODELS_INVALID_CSV",
			"祝日CSVの解析に失敗しました",
			err.Error(),
		)
	}

	holidayDates := make([]models.HolidayDate, 0, len(records))
	seenHolidayDateMap := make(map[string]bool)
	skippedCount := 0

	for _, record := range records {
		if len(record) < 2 {
			skippedCount += 1
			continue
		}

		dateText := strings.TrimSpace(removeUTF8BOM(record[0]))
		nameText := strings.TrimSpace(record[1])

		if dateText == "" || nameText == "" {
			skippedCount += 1
			continue
		}

		holidayDate, parseResult := parseHolidayDateText(dateText)
		if parseResult.Error {
			skippedCount += 1
			continue
		}

		holidayDateKey := holidayDate.Format("2006-01-02")
		if seenHolidayDateMap[holidayDateKey] {
			skippedCount += 1
			continue
		}

		seenHolidayDateMap[holidayDateKey] = true

		holidayDates = append(holidayDates, models.HolidayDate{
			HolidayDate: holidayDate,
			HolidayName: nameText,
			IsDeleted:   false,
		})
	}

	return holidayDates, skippedCount, results.OK(
		nil,
		"BUILD_CREATE_HOLIDAY_DATE_MODELS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 祝日検索用クエリ作成
 *
 * 対象年月の祝日を取得する。
 *
 * 注意：
 * ・祝日は全ユーザー共通のため userId では絞り込まない
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

/*
 * UTF-8 BOMを除去する
 */
func removeUTF8BOM(value string) string {
	return strings.TrimPrefix(value, "\uFEFF")
}

/*
 * CSVの日付文字列をtime.Timeへ変換する
 */
func parseHolidayDateText(value string) (time.Time, results.Result) {
	dateText := strings.TrimSpace(value)

	layouts := []string{
		"2006/1/2",
		"2006/01/02",
		"2006-1-2",
		"2006-01-02",
	}

	for _, layout := range layouts {
		parsedDate, err := time.ParseInLocation(layout, dateText, time.Local)
		if err == nil {
			return time.Date(
					parsedDate.Year(),
					parsedDate.Month(),
					parsedDate.Day(),
					0,
					0,
					0,
					0,
					time.Local,
				),
				results.OK(
					nil,
					"PARSE_HOLIDAY_DATE_TEXT_SUCCESS",
					"",
					nil,
				)
		}
	}

	return time.Time{}, results.BadRequest(
		"PARSE_HOLIDAY_DATE_TEXT_FAILED",
		"祝日の日付形式が正しくありません",
		map[string]any{
			"holidayDate": dateText,
			"format":      "yyyy/M/d または yyyy-MM-dd",
		},
	)
}
