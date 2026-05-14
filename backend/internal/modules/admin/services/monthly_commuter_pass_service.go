package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
)

/*
 * 管理者用月次通勤定期Service interface
 *
 * Controllerや月次勤怠全体保存ServiceがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・管理者APIでは対象ユーザーIDを targetUserId としてRequestで受け取る
 * ・ControllerではJWTのuserIdを対象ユーザーIDとして使わない
 * ・MonthlyCommuterPass は月次通勤定期データだけを持つ
 * ・管理者側では MonthlyAttendanceRequest の状態による編集ロックを行わない
 */
type MonthlyCommuterPassService interface {
	SearchMonthlyCommuterPass(req types.SearchMonthlyCommuterPassRequest) results.Result
	UpdateMonthlyCommuterPass(req types.UpdateMonthlyCommuterPassRequest) results.Result
	DeleteMonthlyCommuterPass(req types.DeleteMonthlyCommuterPassRequest) results.Result
}

/*
 * 管理者用月次通勤定期Service
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや保存用Modelを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB処理を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 *
 * 状態管理：
 * ・MonthlyCommuterPass 自体は申請状態を持たない
 * ・対象月の申請状態は MonthlyAttendanceRequest 側で管理する
 * ・管理者側では月次申請状態による編集ロックを行わない
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type monthlyCommuterPassService struct {
	monthlyCommuterPassBuilder         builders.MonthlyCommuterPassBuilder
	monthlyCommuterPassRepository      repositories.MonthlyCommuterPassRepository
	monthlyAttendanceRequestBuilder    builders.MonthlyAttendanceRequestBuilder
	monthlyAttendanceRequestRepository repositories.MonthlyAttendanceRequestRepository
}

/*
 * MonthlyCommuterPassService生成
 */
func NewMonthlyCommuterPassService(
	monthlyCommuterPassBuilder builders.MonthlyCommuterPassBuilder,
	monthlyCommuterPassRepository repositories.MonthlyCommuterPassRepository,
	monthlyAttendanceRequestBuilder builders.MonthlyAttendanceRequestBuilder,
	monthlyAttendanceRequestRepository repositories.MonthlyAttendanceRequestRepository,
) *monthlyCommuterPassService {
	return &monthlyCommuterPassService{
		monthlyCommuterPassBuilder:         monthlyCommuterPassBuilder,
		monthlyCommuterPassRepository:      monthlyCommuterPassRepository,
		monthlyAttendanceRequestBuilder:    monthlyAttendanceRequestBuilder,
		monthlyAttendanceRequestRepository: monthlyAttendanceRequestRepository,
	}
}

/*
 * models.MonthlyCommuterPassをフロント返却用MonthlyCommuterPassResponseへ変換する
 *
 * MonthlyCommuterPass は申請状態を持たない。
 * 月次申請状態は MonthlyAttendanceRequest 側で管理する。
 */
func toMonthlyCommuterPassResponse(
	monthlyCommuterPass models.MonthlyCommuterPass,
) types.MonthlyCommuterPassResponse {
	return types.MonthlyCommuterPassResponse{
		ID:     monthlyCommuterPass.ID,
		UserID: monthlyCommuterPass.UserID,

		TargetYear:  monthlyCommuterPass.TargetYear,
		TargetMonth: monthlyCommuterPass.TargetMonth,

		CommuterFrom:   monthlyCommuterPass.CommuterFrom,
		CommuterTo:     monthlyCommuterPass.CommuterTo,
		CommuterMethod: monthlyCommuterPass.CommuterMethod,
		CommuterAmount: monthlyCommuterPass.CommuterAmount,

		IsDeleted: monthlyCommuterPass.IsDeleted,
		CreatedAt: monthlyCommuterPass.CreatedAt,
		UpdatedAt: monthlyCommuterPass.UpdatedAt,
		DeletedAt: monthlyCommuterPass.DeletedAt,
	}
}

/*
 * 対象ユーザーIDのバリデーション
 */
func validateMonthlyCommuterPassTargetUserID(
	targetUserID uint,
	actionCode string,
) results.Result {
	if targetUserID == 0 {
		return results.BadRequest(
			actionCode+"_INVALID_TARGET_USER_ID",
			"対象ユーザーIDが正しくありません",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	return results.OK(
		nil,
		actionCode+"_VALID_TARGET_USER_ID",
		"",
		nil,
	)
}

/*
 * 対象年月のバリデーション
 */
func validateMonthlyCommuterPassTargetMonth(
	targetYear int,
	targetMonth int,
	actionCode string,
) results.Result {
	if targetYear <= 0 {
		return results.BadRequest(
			actionCode+"_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{
				"targetYear": targetYear,
			},
		)
	}

	if targetMonth < 1 || targetMonth > 12 {
		return results.BadRequest(
			actionCode+"_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{
				"targetMonth": targetMonth,
			},
		)
	}

	return results.OK(
		nil,
		actionCode+"_VALID_TARGET_MONTH",
		"",
		nil,
	)
}

/*
 * 月次通勤定期検索
 *
 * 対象ユーザーの対象年月の通勤定期を取得する。
 * 未登録の場合は monthlyCommuterPass = nil で返す。
 */
func (service *monthlyCommuterPassService) SearchMonthlyCommuterPass(
	req types.SearchMonthlyCommuterPassRequest,
) results.Result {
	validateUserResult := validateMonthlyCommuterPassTargetUserID(
		req.TargetUserID,
		"SEARCH_MONTHLY_COMMUTER_PASS",
	)
	if validateUserResult.Error {
		return validateUserResult
	}

	validateMonthResult := validateMonthlyCommuterPassTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"SEARCH_MONTHLY_COMMUTER_PASS",
	)
	if validateMonthResult.Error {
		return validateMonthResult
	}

	query, buildResult := service.monthlyCommuterPassBuilder.BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(
		req.TargetUserID,
		req.TargetYear,
		req.TargetMonth,
	)
	if buildResult.Error {
		return buildResult
	}

	monthlyCommuterPass, findResult := service.monthlyCommuterPassRepository.FindMonthlyCommuterPass(query)

	if findResult.Error && findResult.Code == "MONTHLY_COMMUTER_PASS_NOT_FOUND" {
		return results.OK(
			types.SearchMonthlyCommuterPassResponse{
				TargetUserID:        req.TargetUserID,
				TargetYear:          req.TargetYear,
				TargetMonth:         req.TargetMonth,
				MonthlyCommuterPass: nil,
			},
			"SEARCH_MONTHLY_COMMUTER_PASS_SUCCESS",
			"月次通勤定期を取得しました",
			nil,
		)
	}

	if findResult.Error {
		return findResult
	}

	response := toMonthlyCommuterPassResponse(monthlyCommuterPass)

	return results.OK(
		types.SearchMonthlyCommuterPassResponse{
			TargetUserID:        req.TargetUserID,
			TargetYear:          req.TargetYear,
			TargetMonth:         req.TargetMonth,
			MonthlyCommuterPass: &response,
		},
		"SEARCH_MONTHLY_COMMUTER_PASS_SUCCESS",
		"月次通勤定期を取得しました",
		nil,
	)
}

/*
 * 月次通勤定期更新
 *
 * APIとして直接公開しない。
 * monthly_attendances/update の月次全体保存から内部的に使う。
 *
 * 仕様：
 * ・targetUserId + targetYear + targetMonth で既存通勤定期を検索する
 * ・存在しなければ新規作成する
 * ・存在すれば更新する
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
func (service *monthlyCommuterPassService) UpdateMonthlyCommuterPass(
	req types.UpdateMonthlyCommuterPassRequest,
) results.Result {
	validateUserResult := validateMonthlyCommuterPassTargetUserID(
		req.TargetUserID,
		"UPDATE_MONTHLY_COMMUTER_PASS",
	)
	if validateUserResult.Error {
		return validateUserResult
	}

	validateMonthResult := validateMonthlyCommuterPassTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"UPDATE_MONTHLY_COMMUTER_PASS",
	)
	if validateMonthResult.Error {
		return validateMonthResult
	}

	findQuery, buildFindResult := service.monthlyCommuterPassBuilder.BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(
		req.TargetUserID,
		req.TargetYear,
		req.TargetMonth,
	)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentMonthlyCommuterPass, findResult := service.monthlyCommuterPassRepository.FindMonthlyCommuterPass(findQuery)

	if findResult.Error && findResult.Code == "MONTHLY_COMMUTER_PASS_NOT_FOUND" {
		monthlyCommuterPass, buildCreateResult := service.monthlyCommuterPassBuilder.BuildCreateMonthlyCommuterPassModel(
			req,
		)
		if buildCreateResult.Error {
			return buildCreateResult
		}

		createdMonthlyCommuterPass, createResult := service.monthlyCommuterPassRepository.CreateMonthlyCommuterPass(monthlyCommuterPass)
		if createResult.Error {
			return createResult
		}

		return results.Created(
			types.UpdateMonthlyCommuterPassResponse{
				MonthlyCommuterPass: toMonthlyCommuterPassResponse(createdMonthlyCommuterPass),
			},
			"CREATE_MONTHLY_COMMUTER_PASS_SUCCESS",
			"月次通勤定期を作成しました",
			nil,
		)
	}

	if findResult.Error {
		return findResult
	}

	monthlyCommuterPass, buildUpdateResult := service.monthlyCommuterPassBuilder.BuildUpdateMonthlyCommuterPassModel(
		currentMonthlyCommuterPass,
		req,
	)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	savedMonthlyCommuterPass, saveResult := service.monthlyCommuterPassRepository.SaveMonthlyCommuterPass(monthlyCommuterPass)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.UpdateMonthlyCommuterPassResponse{
			MonthlyCommuterPass: toMonthlyCommuterPassResponse(savedMonthlyCommuterPass),
		},
		"UPDATE_MONTHLY_COMMUTER_PASS_SUCCESS",
		"月次通勤定期を更新しました",
		nil,
	)
}

/*
 * 月次通勤定期削除
 *
 * APIとして直接公開しない。
 * monthly_attendances/update の月次全体保存から内部的に使う。
 *
 * 注意：
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
func (service *monthlyCommuterPassService) DeleteMonthlyCommuterPass(
	req types.DeleteMonthlyCommuterPassRequest,
) results.Result {
	validateUserResult := validateMonthlyCommuterPassTargetUserID(
		req.TargetUserID,
		"DELETE_MONTHLY_COMMUTER_PASS",
	)
	if validateUserResult.Error {
		return validateUserResult
	}

	validateMonthResult := validateMonthlyCommuterPassTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"DELETE_MONTHLY_COMMUTER_PASS",
	)
	if validateMonthResult.Error {
		return validateMonthResult
	}

	findQuery, buildFindResult := service.monthlyCommuterPassBuilder.BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(
		req.TargetUserID,
		req.TargetYear,
		req.TargetMonth,
	)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentMonthlyCommuterPass, findResult := service.monthlyCommuterPassRepository.FindMonthlyCommuterPass(findQuery)
	if findResult.Error {
		return findResult
	}

	deletedMonthlyCommuterPass, buildDeleteResult := service.monthlyCommuterPassBuilder.BuildDeleteMonthlyCommuterPassModel(currentMonthlyCommuterPass)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	_, saveResult := service.monthlyCommuterPassRepository.SaveMonthlyCommuterPass(deletedMonthlyCommuterPass)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteMonthlyCommuterPassResponse{
			TargetUserID: req.TargetUserID,
			TargetYear:   req.TargetYear,
			TargetMonth:  req.TargetMonth,
		},
		"DELETE_MONTHLY_COMMUTER_PASS_SUCCESS",
		"月次通勤定期を削除しました",
		nil,
	)
}
