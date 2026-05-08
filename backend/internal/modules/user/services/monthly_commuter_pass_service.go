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
 * ControllerがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・従業員APIでは userId / targetUserId をRequestで受け取らない
 * ・ControllerでAuthMiddleware由来のuserIdを取得し、Serviceへ渡す
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
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type monthlyCommuterPassService struct {
	monthlyCommuterPassBuilder    builders.MonthlyCommuterPassBuilder
	monthlyCommuterPassRepository repositories.MonthlyCommuterPassRepository
}

/*
 * MonthlyCommuterPassService生成
 */
func NewMonthlyCommuterPassService(
	monthlyCommuterPassBuilder builders.MonthlyCommuterPassBuilder,
	monthlyCommuterPassRepository repositories.MonthlyCommuterPassRepository,
) *monthlyCommuterPassService {
	return &monthlyCommuterPassService{
		monthlyCommuterPassBuilder:    monthlyCommuterPassBuilder,
		monthlyCommuterPassRepository: monthlyCommuterPassRepository,
	}
}

/*
 * models.MonthlyCommuterPassをフロント返却用MonthlyCommuterPassResponseへ変換する
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
		MonthlyStatus:  monthlyCommuterPass.MonthlyStatus,
		IsDeleted:      monthlyCommuterPass.IsDeleted,
		CreatedAt:      monthlyCommuterPass.CreatedAt,
		UpdatedAt:      monthlyCommuterPass.UpdatedAt,
		DeletedAt:      monthlyCommuterPass.DeletedAt,
	}
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

	// Builderで対象月次通勤定期取得用クエリを作成する
	query, buildResult := service.monthlyCommuterPassBuilder.BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(
		userID,
		req.TargetYear,
		req.TargetMonth,
	)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryで対象月次通勤定期を取得する
	monthlyCommuterPass, findResult := service.monthlyCommuterPassRepository.FindMonthlyCommuterPass(query)

	// 未登録の場合はnullで返す
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
 * 画面上は対象年月の通勤定期を更新する操作。
 *
 * 仕様：
 * ・userID + targetYear + targetMonth で既存通勤定期を検索する
 * ・存在しなければ新規作成する
 * ・存在すれば更新する
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

	// Builderで対象月次通勤定期取得用クエリを作成する
	findQuery, buildFindResult := service.monthlyCommuterPassBuilder.BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(
		userID,
		req.TargetYear,
		req.TargetMonth,
	)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象月次通勤定期を取得する
	currentMonthlyCommuterPass, findResult := service.monthlyCommuterPassRepository.FindMonthlyCommuterPass(findQuery)

	// 対象年月の通勤定期が存在しない場合は新規作成する
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

	// 既存月次通勤定期がある場合は、更新可能か確認する
	editableResult := validateMonthlyCommuterPassEditable(
		currentMonthlyCommuterPass,
		"UPDATE_MONTHLY_COMMUTER_PASS",
	)
	if editableResult.Error {
		return editableResult
	}

	// Builderで更新用Modelを作る
	monthlyCommuterPass, buildUpdateResult := service.monthlyCommuterPassBuilder.BuildUpdateMonthlyCommuterPassModel(
		currentMonthlyCommuterPass,
		req,
	)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	// Repositoryで月次通勤定期を保存する
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
 * userID + targetYear + targetMonth で対象データを取得し、論理削除する。
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

	// Builderで対象月次通勤定期取得用クエリを作成する
	findQuery, buildFindResult := service.monthlyCommuterPassBuilder.BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(
		userID,
		req.TargetYear,
		req.TargetMonth,
	)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象月次通勤定期を取得する
	currentMonthlyCommuterPass, findResult := service.monthlyCommuterPassRepository.FindMonthlyCommuterPass(findQuery)
	if findResult.Error {
		return findResult
	}

	// 既存月次通勤定期がある場合は、削除可能か確認する
	editableResult := validateMonthlyCommuterPassEditable(
		currentMonthlyCommuterPass,
		"DELETE_MONTHLY_COMMUTER_PASS",
	)
	if editableResult.Error {
		return editableResult
	}

	// Builderで論理削除用Modelを作る
	deletedMonthlyCommuterPass, buildDeleteResult := service.monthlyCommuterPassBuilder.BuildDeleteMonthlyCommuterPassModel(currentMonthlyCommuterPass)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	// Repositoryで月次通勤定期を保存する
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

/*
 * 月次通勤定期の更新・削除が可能か確認する
 *
 * 月次申請中・月次承認済みの場合は、従業員側から更新・削除できない。
 */
func validateMonthlyCommuterPassEditable(
	monthlyCommuterPass models.MonthlyCommuterPass,
	actionCode string,
) results.Result {
	if monthlyCommuterPass.MonthlyStatus == "PENDING" {
		return results.Conflict(
			actionCode+"_MONTHLY_STATUS_PENDING",
			"月次申請中のため、通勤定期を変更できません",
			map[string]any{
				"monthlyCommuterPassId": monthlyCommuterPass.ID,
				"monthlyStatus":         monthlyCommuterPass.MonthlyStatus,
			},
		)
	}

	if monthlyCommuterPass.MonthlyStatus == "APPROVED" {
		return results.Conflict(
			actionCode+"_MONTHLY_STATUS_APPROVED",
			"月次承認済みのため、通勤定期を変更できません",
			map[string]any{
				"monthlyCommuterPassId": monthlyCommuterPass.ID,
				"monthlyStatus":         monthlyCommuterPass.MonthlyStatus,
			},
		)
	}

	return results.OK(
		nil,
		actionCode+"_EDITABLE",
		"",
		nil,
	)
}
