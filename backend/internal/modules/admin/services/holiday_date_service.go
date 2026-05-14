package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
)

/*
 * 管理者用祝日Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type HolidayDateService interface {
	ImportHolidayDates(req types.ImportHolidayDatesRequest) results.Result
	SearchHolidayDates(req types.SearchHolidayDatesRequest) results.Result
}

/*
 * 管理者用祝日Service
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや作成用Modelを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB処理を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 *
 * 祝日CSVインポート方針：
 * ・既存のholiday_datesを物理削除する
 * ・CSVから作成した祝日データを全件新規登録する
 * ・差分更新は行わない
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type holidayDateService struct {
	holidayDateBuilder    builders.HolidayDateBuilder
	holidayDateRepository repositories.HolidayDateRepository
}

/*
 * HolidayDateService生成
 */
func NewHolidayDateService(
	holidayDateBuilder builders.HolidayDateBuilder,
	holidayDateRepository repositories.HolidayDateRepository,
) *holidayDateService {
	return &holidayDateService{
		holidayDateBuilder:    holidayDateBuilder,
		holidayDateRepository: holidayDateRepository,
	}
}

/*
 * models.HolidayDateをフロント返却用HolidayDateResponseへ変換する
 *
 * 日付はtime.Timeのまま返す。
 * 表示形式の整形はフロント側で行う。
 */
func toHolidayDateResponse(holidayDate models.HolidayDate) types.HolidayDateResponse {
	return types.HolidayDateResponse{
		ID:          holidayDate.ID,
		HolidayDate: holidayDate.HolidayDate,
		HolidayName: holidayDate.HolidayName,
		CreatedAt:   holidayDate.CreatedAt,
		UpdatedAt:   holidayDate.UpdatedAt,
	}
}

/*
 * 祝日CSV取り込み
 *
 * 仕様：
 * ・既存のholiday_datesを物理削除する
 * ・CSVの内容を解析して、新しい祝日データを全件登録する
 * ・CSV取り込みは管理者側だけが行う
 *
 * 注意：
 * ・CsvTextが空の場合は取り込み不可
 * ・CSV解析とModel作成はBuilderに任せる
 * ・物理削除と一括作成はRepositoryに任せる
 *
 * DeletedCountについて：
 * ・現状のRepositoryは削除件数を返さないため、0固定で返す
 * ・削除件数を画面表示したくなった場合は、Repository側を削除件数返却に変更する
 */
func (service *holidayDateService) ImportHolidayDates(
	req types.ImportHolidayDatesRequest,
) results.Result {
	if req.CsvText == "" {
		return results.BadRequest(
			"IMPORT_HOLIDAY_DATES_EMPTY_CSV_TEXT",
			"祝日CSVの内容を入力してください",
			nil,
		)
	}

	holidayDates, skippedCount, buildResult := service.holidayDateBuilder.BuildCreateHolidayDateModels(req)
	if buildResult.Error {
		return buildResult
	}

	if len(holidayDates) == 0 {
		return results.BadRequest(
			"IMPORT_HOLIDAY_DATES_EMPTY_VALID_ROWS",
			"取り込み可能な祝日データがありません",
			map[string]any{
				"skippedCount": skippedCount,
			},
		)
	}

	deleteResult := service.holidayDateRepository.DeleteAllHolidayDates()
	if deleteResult.Error {
		return deleteResult
	}

	createdHolidayDates, createResult := service.holidayDateRepository.CreateHolidayDates(holidayDates)
	if createResult.Error {
		return createResult
	}

	return results.Created(
		types.ImportHolidayDatesResponse{
			DeletedCount:  0,
			ImportedCount: len(createdHolidayDates),
			SkippedCount:  skippedCount,
		},
		"IMPORT_HOLIDAY_DATES_SUCCESS",
		"祝日CSVを取り込みました",
		nil,
	)
}

/*
 * 祝日検索
 *
 * 対象年月の祝日一覧を取得する。
 *
 * 用途：
 * ・管理者画面で登録済み祝日を確認する
 * ・CSV取り込み前の確認に使う
 * ・CSV取り込み後の確認に使う
 */
func (service *holidayDateService) SearchHolidayDates(
	req types.SearchHolidayDatesRequest,
) results.Result {
	if req.TargetYear <= 0 {
		return results.BadRequest(
			"SEARCH_HOLIDAY_DATES_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return results.BadRequest(
			"SEARCH_HOLIDAY_DATES_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	query, buildResult := service.holidayDateBuilder.BuildSearchHolidayDatesQuery(req)
	if buildResult.Error {
		return buildResult
	}

	holidayDates, findResult := service.holidayDateRepository.FindHolidayDates(query)
	if findResult.Error {
		return findResult
	}

	holidayDateResponses := make([]types.HolidayDateResponse, 0, len(holidayDates))
	for _, holidayDate := range holidayDates {
		holidayDateResponses = append(holidayDateResponses, toHolidayDateResponse(holidayDate))
	}

	return results.OK(
		types.SearchHolidayDatesResponse{
			Holidays: holidayDateResponses,
		},
		"SEARCH_HOLIDAY_DATES_SUCCESS",
		"祝日一覧を取得しました",
		nil,
	)
}
