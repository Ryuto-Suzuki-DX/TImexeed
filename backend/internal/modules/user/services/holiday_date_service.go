package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
)

/*
 * 従業員用祝日Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・従業員側では祝日の登録、更新、削除は行わない
 * ・CSV取り込みは管理者側APIで行う
 * ・祝日マスタは全ユーザー共通
 */
type HolidayDateService interface {
	SearchHolidayDates(req types.SearchHolidayDatesRequest) results.Result
}

/*
 * 従業員用祝日Service
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB処理を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 * ・祝日は全ユーザー共通のため userId では絞り込まない
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
 */
func toHolidayDateResponse(holidayDate models.HolidayDate) types.HolidayDateResponse {
	return types.HolidayDateResponse{
		ID:          holidayDate.ID,
		HolidayDate: holidayDate.HolidayDate,
		HolidayName: holidayDate.HolidayName,
	}
}

/*
 * 祝日検索
 *
 * 対象年月の祝日一覧を取得する。
 *
 * 仕様：
 * ・祝日は全ユーザー共通のため userId では絞り込まない
 * ・従業員側では参照のみ行う
 * ・登録、更新、削除、CSV取り込みは行わない
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
