package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
)

type MonthlyCommuterPassService interface {
	SearchMonthlyCommuterPass(userID uint, req types.SearchMonthlyCommuterPassRequest) results.Result
	UpdateMonthlyCommuterPasses(userID uint, req types.UpdateMonthlyCommuterPassesRequest) results.Result
	UpdateMonthlyCommuterPass(userID uint, req types.UpdateMonthlyCommuterPassRequest) results.Result
	DeleteMonthlyCommuterPass(userID uint, req types.DeleteMonthlyCommuterPassRequest) results.Result
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
	monthlyStatus string,
) types.MonthlyCommuterPassResponse {
	return types.MonthlyCommuterPassResponse{
		ID:             monthlyCommuterPass.ID,
		TargetYear:     monthlyCommuterPass.TargetYear,
		TargetMonth:    monthlyCommuterPass.TargetMonth,
		CommuterFrom:   monthlyCommuterPass.CommuterFrom,
		CommuterTo:     monthlyCommuterPass.CommuterTo,
		CommuterMethod: monthlyCommuterPass.CommuterMethod,
		CommuterAmount: monthlyCommuterPass.CommuterAmount,
		MonthlyStatus:  monthlyStatus,
		IsDeleted:      monthlyCommuterPass.IsDeleted,
		CreatedAt:      monthlyCommuterPass.CreatedAt,
		UpdatedAt:      monthlyCommuterPass.UpdatedAt,
		DeletedAt:      monthlyCommuterPass.DeletedAt,
	}
}

func toMonthlyCommuterPassResponses(
	monthlyCommuterPasses []models.MonthlyCommuterPass,
	monthlyStatus string,
) ([]types.MonthlyCommuterPassResponse, int) {
	responses := make([]types.MonthlyCommuterPassResponse, 0, len(monthlyCommuterPasses))
	totalCommuterAmount := 0

	for _, monthlyCommuterPass := range monthlyCommuterPasses {
		responses = append(
			responses,
			toMonthlyCommuterPassResponse(monthlyCommuterPass, monthlyStatus),
		)

		if monthlyCommuterPass.CommuterAmount != nil {
			totalCommuterAmount += *monthlyCommuterPass.CommuterAmount
		}
	}

	return responses, totalCommuterAmount
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
 * 対象月の月次申請状態を取得する。
 *
 * 月次申請レコードが存在しない場合は NOT_SUBMITTED を返す。
 */
func (service *monthlyCommuterPassService) getMonthlyAttendanceStatus(
	userID uint,
	targetYear int,
	targetMonth int,
	actionCode string,
) (string, results.Result) {
	query, buildResult :=
		service.monthlyAttendanceRequestBuilder.
			BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
				userID,
				targetYear,
				targetMonth,
			)
	if buildResult.Error {
		return "", buildResult
	}

	monthlyAttendanceRequest, findResult :=
		service.monthlyAttendanceRequestRepository.
			FindMonthlyAttendanceRequest(query)

	if findResult.Error && findResult.Code == "MONTHLY_ATTENDANCE_REQUEST_NOT_FOUND" {
		return "NOT_SUBMITTED", results.OK(
			nil,
			actionCode+"_MONTHLY_ATTENDANCE_STATUS_NOT_SUBMITTED",
			"",
			nil,
		)
	}

	if findResult.Error {
		return "", findResult
	}

	return monthlyAttendanceRequest.Status, results.OK(
		nil,
		actionCode+"_MONTHLY_ATTENDANCE_STATUS_SUCCESS",
		"",
		nil,
	)
}

/*
 * PENDING / APPROVED の場合は従業員側から更新・削除できない。
 */
func (service *monthlyCommuterPassService) validateMonthlyAttendanceEditable(
	userID uint,
	targetYear int,
	targetMonth int,
	actionCode string,
) results.Result {
	monthlyStatus, statusResult := service.getMonthlyAttendanceStatus(
		userID,
		targetYear,
		targetMonth,
		actionCode,
	)
	if statusResult.Error {
		return statusResult
	}

	if monthlyStatus == "PENDING" {
		return results.Conflict(
			actionCode+"_MONTHLY_ATTENDANCE_REQUEST_PENDING",
			"月次申請中のため、通勤定期を変更できません",
			map[string]any{
				"targetYear":  targetYear,
				"targetMonth": targetMonth,
				"status":      monthlyStatus,
			},
		)
	}

	if monthlyStatus == "APPROVED" {
		return results.Conflict(
			actionCode+"_MONTHLY_ATTENDANCE_REQUEST_APPROVED",
			"月次承認済みのため、通勤定期を変更できません",
			map[string]any{
				"targetYear":  targetYear,
				"targetMonth": targetMonth,
				"status":      monthlyStatus,
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
 * 対象年月のログイン中ユーザー本人の通勤定期をすべて取得する。
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

	validateMonthResult := validateMonthlyCommuterPassTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"SEARCH_MONTHLY_COMMUTER_PASS",
	)
	if validateMonthResult.Error {
		return validateMonthResult
	}

	monthlyStatus, statusResult := service.getMonthlyAttendanceStatus(
		userID,
		req.TargetYear,
		req.TargetMonth,
		"SEARCH_MONTHLY_COMMUTER_PASS",
	)
	if statusResult.Error {
		return statusResult
	}

	query, buildResult :=
		service.monthlyCommuterPassBuilder.
			BuildFindMonthlyCommuterPassesByUserIDAndTargetYearMonthQuery(
				userID,
				req.TargetYear,
				req.TargetMonth,
			)
	if buildResult.Error {
		return buildResult
	}

	monthlyCommuterPasses, findResult :=
		service.monthlyCommuterPassRepository.
			FindMonthlyCommuterPasses(query)
	if findResult.Error {
		return findResult
	}

	responses, totalCommuterAmount :=
		toMonthlyCommuterPassResponses(monthlyCommuterPasses, monthlyStatus)

	return results.OK(
		types.SearchMonthlyCommuterPassResponse{
			TargetYear:            req.TargetYear,
			TargetMonth:           req.TargetMonth,
			MonthlyStatus:         monthlyStatus,
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
 * ・空配列：対象年月の既存定期をすべて論理削除
 */
func (service *monthlyCommuterPassService) UpdateMonthlyCommuterPasses(
	userID uint,
	req types.UpdateMonthlyCommuterPassesRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"UPDATE_MONTHLY_COMMUTER_PASSES_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	validateMonthResult := validateMonthlyCommuterPassTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"UPDATE_MONTHLY_COMMUTER_PASSES",
	)
	if validateMonthResult.Error {
		return validateMonthResult
	}

	editableResult := service.validateMonthlyAttendanceEditable(
		userID,
		req.TargetYear,
		req.TargetMonth,
		"UPDATE_MONTHLY_COMMUTER_PASSES",
	)
	if editableResult.Error {
		return editableResult
	}

	monthlyStatus, statusResult := service.getMonthlyAttendanceStatus(
		userID,
		req.TargetYear,
		req.TargetMonth,
		"UPDATE_MONTHLY_COMMUTER_PASSES",
	)
	if statusResult.Error {
		return statusResult
	}

	findQuery, buildFindResult :=
		service.monthlyCommuterPassBuilder.
			BuildFindMonthlyCommuterPassesByUserIDAndTargetYearMonthQuery(
				userID,
				req.TargetYear,
				req.TargetMonth,
			)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentMonthlyCommuterPasses, findResult :=
		service.monthlyCommuterPassRepository.
			FindMonthlyCommuterPasses(findQuery)
	if findResult.Error {
		return findResult
	}

	currentByID := make(
		map[uint]models.MonthlyCommuterPass,
		len(currentMonthlyCommuterPasses),
	)
	for _, current := range currentMonthlyCommuterPasses {
		currentByID[current.ID] = current
	}

	requestedIDs := make(map[uint]bool)
	savedMonthlyCommuterPasses := make(
		[]models.MonthlyCommuterPass,
		0,
		len(req.CommuterPasses),
	)
	savedCount := 0

	for _, commuterPassReq := range req.CommuterPasses {
		if commuterPassReq.MonthlyCommuterPassID == nil ||
			*commuterPassReq.MonthlyCommuterPassID == 0 {
			createModel, buildCreateResult :=
				service.monthlyCommuterPassBuilder.
					BuildCreateMonthlyCommuterPassModel(
						userID,
						req.TargetYear,
						req.TargetMonth,
						commuterPassReq,
					)
			if buildCreateResult.Error {
				return buildCreateResult
			}

			created, createResult :=
				service.monthlyCommuterPassRepository.
					CreateMonthlyCommuterPass(createModel)
			if createResult.Error {
				return createResult
			}

			savedMonthlyCommuterPasses = append(
				savedMonthlyCommuterPasses,
				created,
			)
			savedCount++
			continue
		}

		monthlyCommuterPassID := *commuterPassReq.MonthlyCommuterPassID

		if requestedIDs[monthlyCommuterPassID] {
			return results.BadRequest(
				"UPDATE_MONTHLY_COMMUTER_PASSES_DUPLICATE_ID",
				"同じ月次通勤定期IDが複数回指定されています",
				map[string]any{
					"monthlyCommuterPassId": monthlyCommuterPassID,
				},
			)
		}
		requestedIDs[monthlyCommuterPassID] = true

		current, exists := currentByID[monthlyCommuterPassID]
		if !exists {
			return results.NotFound(
				"MONTHLY_COMMUTER_PASS_NOT_FOUND",
				"更新対象の月次通勤定期が見つかりません",
				map[string]any{
					"monthlyCommuterPassId": monthlyCommuterPassID,
				},
			)
		}

		updateModel, buildUpdateResult :=
			service.monthlyCommuterPassBuilder.
				BuildUpdateMonthlyCommuterPassModel(
					current,
					userID,
					req.TargetYear,
					req.TargetMonth,
					commuterPassReq,
				)
		if buildUpdateResult.Error {
			return buildUpdateResult
		}

		saved, saveResult :=
			service.monthlyCommuterPassRepository.
				SaveMonthlyCommuterPass(updateModel)
		if saveResult.Error {
			return saveResult
		}

		savedMonthlyCommuterPasses = append(
			savedMonthlyCommuterPasses,
			saved,
		)
		savedCount++
	}

	for _, current := range currentMonthlyCommuterPasses {
		if requestedIDs[current.ID] {
			continue
		}

		deleteModel, buildDeleteResult :=
			service.monthlyCommuterPassBuilder.
				BuildDeleteMonthlyCommuterPassModel(current)
		if buildDeleteResult.Error {
			return buildDeleteResult
		}

		_, saveResult :=
			service.monthlyCommuterPassRepository.
				SaveMonthlyCommuterPass(deleteModel)
		if saveResult.Error {
			return saveResult
		}

		savedCount++
	}

	responses, totalCommuterAmount :=
		toMonthlyCommuterPassResponses(
			savedMonthlyCommuterPasses,
			monthlyStatus,
		)

	return results.OK(
		types.UpdateMonthlyCommuterPassesResponse{
			TargetYear:                    req.TargetYear,
			TargetMonth:                   req.TargetMonth,
			MonthlyStatus:                 monthlyStatus,
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
 * 新しい月次勤怠全体保存では UpdateMonthlyCommuterPasses を使う。
 */
func (service *monthlyCommuterPassService) UpdateMonthlyCommuterPass(
	userID uint,
	req types.UpdateMonthlyCommuterPassRequest,
) results.Result {
	itemRequest := types.UpdateMonthlyCommuterPassItemRequest{
		CommuterFrom:   req.CommuterFrom,
		CommuterTo:     req.CommuterTo,
		CommuterMethod: req.CommuterMethod,
		CommuterAmount: req.CommuterAmount,
	}

	findQuery, buildFindResult :=
		service.monthlyCommuterPassBuilder.
			BuildFindMonthlyCommuterPassesByUserIDAndTargetYearMonthQuery(
				userID,
				req.TargetYear,
				req.TargetMonth,
			)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentMonthlyCommuterPasses, findResult :=
		service.monthlyCommuterPassRepository.
			FindMonthlyCommuterPasses(findQuery)
	if findResult.Error {
		return findResult
	}

	if len(currentMonthlyCommuterPasses) > 0 {
		itemRequest.MonthlyCommuterPassID =
			&currentMonthlyCommuterPasses[0].ID
	}

	updateResult := service.UpdateMonthlyCommuterPasses(
		userID,
		types.UpdateMonthlyCommuterPassesRequest{
			TargetYear:     req.TargetYear,
			TargetMonth:    req.TargetMonth,
			CommuterPasses: []types.UpdateMonthlyCommuterPassItemRequest{itemRequest},
		},
	)
	if updateResult.Error {
		return updateResult
	}

	updateResponse, ok :=
		updateResult.Data.(types.UpdateMonthlyCommuterPassesResponse)
	if !ok || len(updateResponse.MonthlyCommuterPasses) == 0 {
		return results.InternalServerError(
			"UPDATE_MONTHLY_COMMUTER_PASS_INVALID_RESPONSE",
			"月次通勤定期保存結果の形式が正しくありません",
			nil,
		)
	}

	return results.OK(
		types.UpdateMonthlyCommuterPassResponse{
			MonthlyCommuterPass: updateResponse.MonthlyCommuterPasses[0],
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

	updateResult := service.UpdateMonthlyCommuterPasses(
		userID,
		types.UpdateMonthlyCommuterPassesRequest{
			TargetYear:     req.TargetYear,
			TargetMonth:    req.TargetMonth,
			CommuterPasses: []types.UpdateMonthlyCommuterPassItemRequest{},
		},
	)
	if updateResult.Error {
		return updateResult
	}

	updateResponse, ok :=
		updateResult.Data.(types.UpdateMonthlyCommuterPassesResponse)
	if !ok {
		return results.InternalServerError(
			"DELETE_MONTHLY_COMMUTER_PASS_INVALID_RESPONSE",
			"月次通勤定期削除結果の形式が正しくありません",
			nil,
		)
	}

	return results.OK(
		types.DeleteMonthlyCommuterPassResponse{
			TargetYear:                      req.TargetYear,
			TargetMonth:                     req.TargetMonth,
			DeletedMonthlyCommuterPassCount: updateResponse.SavedMonthlyCommuterPassCount,
		},
		"DELETE_MONTHLY_COMMUTER_PASS_SUCCESS",
		"月次通勤定期を削除しました",
		nil,
	)
}
