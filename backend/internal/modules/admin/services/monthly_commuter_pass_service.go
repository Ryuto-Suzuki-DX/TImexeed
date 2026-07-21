package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
)

type MonthlyCommuterPassService interface {
	SearchMonthlyCommuterPass(req types.SearchMonthlyCommuterPassRequest) results.Result
	UpdateMonthlyCommuterPasses(req types.UpdateMonthlyCommuterPassesRequest) results.Result
	UpdateMonthlyCommuterPass(req types.UpdateMonthlyCommuterPassRequest) results.Result
	DeleteMonthlyCommuterPass(req types.DeleteMonthlyCommuterPassRequest) results.Result
}

type monthlyCommuterPassService struct {
	monthlyCommuterPassBuilder         builders.MonthlyCommuterPassBuilder
	monthlyCommuterPassRepository      repositories.MonthlyCommuterPassRepository
	monthlyAttendanceRequestBuilder    builders.MonthlyAttendanceRequestBuilder
	monthlyAttendanceRequestRepository repositories.MonthlyAttendanceRequestRepository
}

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

func toMonthlyCommuterPassResponses(
	monthlyCommuterPasses []models.MonthlyCommuterPass,
) ([]types.MonthlyCommuterPassResponse, int) {
	responses := make([]types.MonthlyCommuterPassResponse, 0, len(monthlyCommuterPasses))
	totalCommuterAmount := 0

	for _, monthlyCommuterPass := range monthlyCommuterPasses {
		responses = append(responses, toMonthlyCommuterPassResponse(monthlyCommuterPass))

		if monthlyCommuterPass.CommuterAmount != nil {
			totalCommuterAmount += *monthlyCommuterPass.CommuterAmount
		}
	}

	return responses, totalCommuterAmount
}

func validateMonthlyCommuterPassTargetUserID(
	targetUserID uint,
	actionCode string,
) results.Result {
	if targetUserID == 0 {
		return results.BadRequest(
			actionCode+"_INVALID_TARGET_USER_ID",
			"対象ユーザーIDが正しくありません",
			map[string]any{"targetUserId": targetUserID},
		)
	}

	return results.OK(nil, actionCode+"_VALID_TARGET_USER_ID", "", nil)
}

func validateMonthlyCommuterPassTargetMonth(
	targetYear int,
	targetMonth int,
	actionCode string,
) results.Result {
	if targetYear <= 0 {
		return results.BadRequest(
			actionCode+"_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{"targetYear": targetYear},
		)
	}

	if targetMonth < 1 || targetMonth > 12 {
		return results.BadRequest(
			actionCode+"_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{"targetMonth": targetMonth},
		)
	}

	return results.OK(nil, actionCode+"_VALID_TARGET_MONTH", "", nil)
}

/*
 * 対象ユーザー・対象年月の月次通勤定期をすべて取得する。
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

	query, buildResult := service.monthlyCommuterPassBuilder.
		BuildFindMonthlyCommuterPassesByUserIDAndTargetYearMonthQuery(
			req.TargetUserID,
			req.TargetYear,
			req.TargetMonth,
		)
	if buildResult.Error {
		return buildResult
	}

	monthlyCommuterPasses, findResult := service.monthlyCommuterPassRepository.
		FindMonthlyCommuterPasses(query)
	if findResult.Error {
		return findResult
	}

	responses, totalCommuterAmount := toMonthlyCommuterPassResponses(monthlyCommuterPasses)

	return results.OK(
		types.SearchMonthlyCommuterPassResponse{
			TargetUserID:          req.TargetUserID,
			TargetYear:            req.TargetYear,
			TargetMonth:           req.TargetMonth,
			MonthlyCommuterPasses: responses,
			TotalCommuterAmount:   totalCommuterAmount,
		},
		"SEARCH_MONTHLY_COMMUTER_PASS_SUCCESS",
		"月次通勤定期を取得しました",
		nil,
	)
}

/*
 * 月次通勤定期差分保存
 *
 * ・IDあり：更新
 * ・IDなし：新規作成
 * ・DBに存在するがRequestから消えたID：論理削除
 */
func (service *monthlyCommuterPassService) UpdateMonthlyCommuterPasses(
	req types.UpdateMonthlyCommuterPassesRequest,
) results.Result {
	validateUserResult := validateMonthlyCommuterPassTargetUserID(
		req.TargetUserID,
		"UPDATE_MONTHLY_COMMUTER_PASSES",
	)
	if validateUserResult.Error {
		return validateUserResult
	}

	validateMonthResult := validateMonthlyCommuterPassTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"UPDATE_MONTHLY_COMMUTER_PASSES",
	)
	if validateMonthResult.Error {
		return validateMonthResult
	}

	findQuery, buildFindResult := service.monthlyCommuterPassBuilder.
		BuildFindMonthlyCommuterPassesByUserIDAndTargetYearMonthQuery(
			req.TargetUserID,
			req.TargetYear,
			req.TargetMonth,
		)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentMonthlyCommuterPasses, findResult := service.monthlyCommuterPassRepository.
		FindMonthlyCommuterPasses(findQuery)
	if findResult.Error {
		return findResult
	}

	currentByID := make(map[uint]models.MonthlyCommuterPass, len(currentMonthlyCommuterPasses))
	for _, current := range currentMonthlyCommuterPasses {
		currentByID[current.ID] = current
	}

	requestedIDs := make(map[uint]bool)
	savedMonthlyCommuterPasses := make([]models.MonthlyCommuterPass, 0, len(req.CommuterPasses))
	savedCount := 0

	for _, commuterPassReq := range req.CommuterPasses {
		if commuterPassReq.MonthlyCommuterPassID == nil || *commuterPassReq.MonthlyCommuterPassID == 0 {
			createModel, buildCreateResult := service.monthlyCommuterPassBuilder.
				BuildCreateMonthlyCommuterPassModel(
					req.TargetUserID,
					req.TargetYear,
					req.TargetMonth,
					commuterPassReq,
				)
			if buildCreateResult.Error {
				return buildCreateResult
			}

			created, createResult := service.monthlyCommuterPassRepository.
				CreateMonthlyCommuterPass(createModel)
			if createResult.Error {
				return createResult
			}

			savedMonthlyCommuterPasses = append(savedMonthlyCommuterPasses, created)
			savedCount++
			continue
		}

		monthlyCommuterPassID := *commuterPassReq.MonthlyCommuterPassID
		if requestedIDs[monthlyCommuterPassID] {
			return results.BadRequest(
				"UPDATE_MONTHLY_COMMUTER_PASSES_DUPLICATE_ID",
				"同じ月次通勤定期IDが複数回指定されています",
				map[string]any{"monthlyCommuterPassId": monthlyCommuterPassID},
			)
		}
		requestedIDs[monthlyCommuterPassID] = true

		current, exists := currentByID[monthlyCommuterPassID]
		if !exists {
			return results.NotFound(
				"MONTHLY_COMMUTER_PASS_NOT_FOUND",
				"更新対象の月次通勤定期が見つかりません",
				map[string]any{"monthlyCommuterPassId": monthlyCommuterPassID},
			)
		}

		updateModel, buildUpdateResult := service.monthlyCommuterPassBuilder.
			BuildUpdateMonthlyCommuterPassModel(
				current,
				req.TargetUserID,
				req.TargetYear,
				req.TargetMonth,
				commuterPassReq,
			)
		if buildUpdateResult.Error {
			return buildUpdateResult
		}

		saved, saveResult := service.monthlyCommuterPassRepository.
			SaveMonthlyCommuterPass(updateModel)
		if saveResult.Error {
			return saveResult
		}

		savedMonthlyCommuterPasses = append(savedMonthlyCommuterPasses, saved)
		savedCount++
	}

	for _, current := range currentMonthlyCommuterPasses {
		if requestedIDs[current.ID] {
			continue
		}

		deleteModel, buildDeleteResult := service.monthlyCommuterPassBuilder.
			BuildDeleteMonthlyCommuterPassModel(current)
		if buildDeleteResult.Error {
			return buildDeleteResult
		}

		_, saveResult := service.monthlyCommuterPassRepository.
			SaveMonthlyCommuterPass(deleteModel)
		if saveResult.Error {
			return saveResult
		}

		savedCount++
	}

	responses, totalCommuterAmount := toMonthlyCommuterPassResponses(savedMonthlyCommuterPasses)

	return results.OK(
		types.UpdateMonthlyCommuterPassesResponse{
			TargetUserID:                  req.TargetUserID,
			TargetYear:                    req.TargetYear,
			TargetMonth:                   req.TargetMonth,
			MonthlyCommuterPasses:         responses,
			SavedMonthlyCommuterPassCount: savedCount,
			TotalCommuterAmount:           totalCommuterAmount,
		},
		"UPDATE_MONTHLY_COMMUTER_PASSES_SUCCESS",
		"月次通勤定期を保存しました",
		nil,
	)
}

/*
 * 旧単体更新処理（互換用）
 *
 * 新しい月次勤怠全体保存では UpdateMonthlyCommuterPasses を使用する。
 * 既存呼び出しを壊さないため、対象年月の先頭1件を更新し、未登録なら作成する。
 */
func (service *monthlyCommuterPassService) UpdateMonthlyCommuterPass(
	req types.UpdateMonthlyCommuterPassRequest,
) results.Result {
	findQuery, buildFindResult := service.monthlyCommuterPassBuilder.
		BuildFindMonthlyCommuterPassesByUserIDAndTargetYearMonthQuery(
			req.TargetUserID,
			req.TargetYear,
			req.TargetMonth,
		)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentMonthlyCommuterPasses, findResult := service.monthlyCommuterPassRepository.
		FindMonthlyCommuterPasses(findQuery)
	if findResult.Error {
		return findResult
	}

	itemReq := types.UpdateMonthlyCommuterPassItemRequest{
		CommuterFrom:   req.CommuterFrom,
		CommuterTo:     req.CommuterTo,
		CommuterMethod: req.CommuterMethod,
		CommuterAmount: req.CommuterAmount,
	}

	if len(currentMonthlyCommuterPasses) == 0 {
		createModel, buildCreateResult := service.monthlyCommuterPassBuilder.
			BuildCreateMonthlyCommuterPassModel(
				req.TargetUserID,
				req.TargetYear,
				req.TargetMonth,
				itemReq,
			)
		if buildCreateResult.Error {
			return buildCreateResult
		}

		created, createResult := service.monthlyCommuterPassRepository.
			CreateMonthlyCommuterPass(createModel)
		if createResult.Error {
			return createResult
		}

		return results.Created(
			types.UpdateMonthlyCommuterPassResponse{
				MonthlyCommuterPass: toMonthlyCommuterPassResponse(created),
			},
			"CREATE_MONTHLY_COMMUTER_PASS_SUCCESS",
			"月次通勤定期を作成しました",
			nil,
		)
	}

	current := currentMonthlyCommuterPasses[0]
	itemReq.MonthlyCommuterPassID = &current.ID

	updateModel, buildUpdateResult := service.monthlyCommuterPassBuilder.
		BuildUpdateMonthlyCommuterPassModel(
			current,
			req.TargetUserID,
			req.TargetYear,
			req.TargetMonth,
			itemReq,
		)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	saved, saveResult := service.monthlyCommuterPassRepository.
		SaveMonthlyCommuterPass(updateModel)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.UpdateMonthlyCommuterPassResponse{
			MonthlyCommuterPass: toMonthlyCommuterPassResponse(saved),
		},
		"UPDATE_MONTHLY_COMMUTER_PASS_SUCCESS",
		"月次通勤定期を更新しました",
		nil,
	)
}

/*
 * 対象年月の月次通勤定期をすべて論理削除する。
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

	findQuery, buildFindResult := service.monthlyCommuterPassBuilder.
		BuildFindMonthlyCommuterPassesByUserIDAndTargetYearMonthQuery(
			req.TargetUserID,
			req.TargetYear,
			req.TargetMonth,
		)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentMonthlyCommuterPasses, findResult := service.monthlyCommuterPassRepository.
		FindMonthlyCommuterPasses(findQuery)
	if findResult.Error {
		return findResult
	}

	deletedCount := 0
	for _, current := range currentMonthlyCommuterPasses {
		deletedModel, buildDeleteResult := service.monthlyCommuterPassBuilder.
			BuildDeleteMonthlyCommuterPassModel(current)
		if buildDeleteResult.Error {
			return buildDeleteResult
		}

		_, saveResult := service.monthlyCommuterPassRepository.
			SaveMonthlyCommuterPass(deletedModel)
		if saveResult.Error {
			return saveResult
		}

		deletedCount++
	}

	return results.OK(
		types.DeleteMonthlyCommuterPassResponse{
			TargetUserID:                    req.TargetUserID,
			TargetYear:                      req.TargetYear,
			TargetMonth:                     req.TargetMonth,
			DeletedMonthlyCommuterPassCount: deletedCount,
		},
		"DELETE_MONTHLY_COMMUTER_PASS_SUCCESS",
		"月次通勤定期を削除しました",
		nil,
	)
}
