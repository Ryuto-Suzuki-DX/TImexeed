package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
)

/*
 * 従業員用月次通勤定期Service interface
 *
 * Controllerや月次勤怠全体保存ServiceがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・従業員APIでは userId / targetUserId をRequestで受け取らない
 * ・ControllerでAuthMiddleware由来のuserIdを取得し、Serviceへ渡す
 * ・MonthlyCommuterPass は月次通勤定期データだけを持つ
 * ・編集可否は MonthlyAttendanceRequest を見て判断する
 */
type MonthlyCommuterPassService interface {
	SearchMonthlyCommuterPass(userID uint, req types.SearchMonthlyCommuterPassRequest) results.Result
	UpdateMonthlyCommuterPass(userID uint, req types.UpdateMonthlyCommuterPassRequest) results.Result
	DeleteMonthlyCommuterPass(userID uint, req types.DeleteMonthlyCommuterPassRequest) results.Result
}

/*
 * 従業員用月次通勤定期Service
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
 * ・対象月の申請状態は MonthlyAttendanceRequest から取得する
 * ・MonthlyAttendanceRequest が存在しない場合は未申請扱いにする
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
		ID:             monthlyCommuterPass.ID,
		TargetYear:     monthlyCommuterPass.TargetYear,
		TargetMonth:    monthlyCommuterPass.TargetMonth,
		CommuterFrom:   monthlyCommuterPass.CommuterFrom,
		CommuterTo:     monthlyCommuterPass.CommuterTo,
		CommuterMethod: monthlyCommuterPass.CommuterMethod,
		CommuterAmount: monthlyCommuterPass.CommuterAmount,
		IsDeleted:      monthlyCommuterPass.IsDeleted,
		CreatedAt:      monthlyCommuterPass.CreatedAt,
		UpdatedAt:      monthlyCommuterPass.UpdatedAt,
		DeletedAt:      monthlyCommuterPass.DeletedAt,
	}
}

/*
 * 対象月の月次通勤定期が編集可能か確認する
 *
 * PENDING / APPROVED の場合は従業員側から更新・削除できない。
 * NOT_SUBMITTED / REJECTED / CANCELED の場合は編集可能。
 */
func (service *monthlyCommuterPassService) validateMonthlyAttendanceEditable(
	userID uint,
	targetYear int,
	targetMonth int,
	actionCode string,
) results.Result {
	query, buildResult := service.monthlyAttendanceRequestBuilder.BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
		userID,
		targetYear,
		targetMonth,
	)
	if buildResult.Error {
		return buildResult
	}

	monthlyAttendanceRequest, findResult := service.monthlyAttendanceRequestRepository.FindMonthlyAttendanceRequest(query)

	if findResult.Error && findResult.Code == "MONTHLY_ATTENDANCE_REQUEST_NOT_FOUND" {
		return results.OK(
			nil,
			actionCode+"_MONTHLY_ATTENDANCE_EDITABLE",
			"",
			nil,
		)
	}

	if findResult.Error {
		return findResult
	}

	if monthlyAttendanceRequest.Status == "PENDING" {
		return results.Conflict(
			actionCode+"_MONTHLY_ATTENDANCE_REQUEST_PENDING",
			"月次申請中のため、通勤定期を変更できません",
			map[string]any{
				"targetYear":  targetYear,
				"targetMonth": targetMonth,
				"status":      monthlyAttendanceRequest.Status,
			},
		)
	}

	if monthlyAttendanceRequest.Status == "APPROVED" {
		return results.Conflict(
			actionCode+"_MONTHLY_ATTENDANCE_REQUEST_APPROVED",
			"月次承認済みのため、通勤定期を変更できません",
			map[string]any{
				"targetYear":  targetYear,
				"targetMonth": targetMonth,
				"status":      monthlyAttendanceRequest.Status,
			},
		)
	}

	return results.OK(
		nil,
		actionCode+"_MONTHLY_ATTENDANCE_EDITABLE",
		"",
		nil,
	)
}

/*
 * 月次通勤定期検索
 *
 * 対象年月のログイン中ユーザー本人の通勤定期を取得する。
 * 未登録の場合は monthlyCommuterPass = nil で返す。
 */
func (service *monthlyCommuterPassService) SearchMonthlyCommuterPass(
	userID uint,
	req types.SearchMonthlyCommuterPassRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"SEARCH_MONTHLY_COMMUTER_PASS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	if req.TargetYear <= 0 {
		return results.BadRequest(
			"SEARCH_MONTHLY_COMMUTER_PASS_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return results.BadRequest(
			"SEARCH_MONTHLY_COMMUTER_PASS_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	query, buildResult := service.monthlyCommuterPassBuilder.BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(
		userID,
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
 * ・userID + targetYear + targetMonth で既存通勤定期を検索する
 * ・存在しなければ新規作成する
 * ・存在すれば更新する
 * ・更新可否は MonthlyAttendanceRequest を見て判断する
 */
func (service *monthlyCommuterPassService) UpdateMonthlyCommuterPass(
	userID uint,
	req types.UpdateMonthlyCommuterPassRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"UPDATE_MONTHLY_COMMUTER_PASS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	if req.TargetYear <= 0 {
		return results.BadRequest(
			"UPDATE_MONTHLY_COMMUTER_PASS_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return results.BadRequest(
			"UPDATE_MONTHLY_COMMUTER_PASS_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	editableResult := service.validateMonthlyAttendanceEditable(
		userID,
		req.TargetYear,
		req.TargetMonth,
		"UPDATE_MONTHLY_COMMUTER_PASS",
	)
	if editableResult.Error {
		return editableResult
	}

	findQuery, buildFindResult := service.monthlyCommuterPassBuilder.BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(
		userID,
		req.TargetYear,
		req.TargetMonth,
	)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentMonthlyCommuterPass, findResult := service.monthlyCommuterPassRepository.FindMonthlyCommuterPass(findQuery)

	if findResult.Error && findResult.Code == "MONTHLY_COMMUTER_PASS_NOT_FOUND" {
		monthlyCommuterPass, buildCreateResult := service.monthlyCommuterPassBuilder.BuildCreateMonthlyCommuterPassModel(
			userID,
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
 */
func (service *monthlyCommuterPassService) DeleteMonthlyCommuterPass(
	userID uint,
	req types.DeleteMonthlyCommuterPassRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"DELETE_MONTHLY_COMMUTER_PASS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	if req.TargetYear <= 0 {
		return results.BadRequest(
			"DELETE_MONTHLY_COMMUTER_PASS_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return results.BadRequest(
			"DELETE_MONTHLY_COMMUTER_PASS_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	editableResult := service.validateMonthlyAttendanceEditable(
		userID,
		req.TargetYear,
		req.TargetMonth,
		"DELETE_MONTHLY_COMMUTER_PASS",
	)
	if editableResult.Error {
		return editableResult
	}

	findQuery, buildFindResult := service.monthlyCommuterPassBuilder.BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(
		userID,
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
			TargetYear:  req.TargetYear,
			TargetMonth: req.TargetMonth,
		},
		"DELETE_MONTHLY_COMMUTER_PASS_SUCCESS",
		"月次通勤定期を削除しました",
		nil,
	)
}
