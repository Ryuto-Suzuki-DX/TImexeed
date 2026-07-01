package services

import (
	"strings"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 従業員用日別交通費Service interface
 */
type AttendanceTransportExpenseService interface {
	SearchAttendanceTransportExpenses(
		userID uint,
		req types.SearchAttendanceTransportExpensesRequest,
	) results.Result
	UpdateAttendanceTransportExpensesByWorkDate(
		userID uint,
		req types.UpdateAttendanceTransportExpensesByWorkDateRequest,
	) results.Result
}

type attendanceTransportExpenseService struct {
	attendanceTransportExpenseBuilder    builders.AttendanceTransportExpenseBuilder
	attendanceTransportExpenseRepository repositories.AttendanceTransportExpenseRepository
	attendanceDayBuilder                 builders.AttendanceDayBuilder
	attendanceDayRepository              repositories.AttendanceDayRepository
	monthlyAttendanceRequestBuilder      builders.MonthlyAttendanceRequestBuilder
	monthlyAttendanceRequestRepository   repositories.MonthlyAttendanceRequestRepository
}

func NewAttendanceTransportExpenseService(
	attendanceTransportExpenseBuilder builders.AttendanceTransportExpenseBuilder,
	attendanceTransportExpenseRepository repositories.AttendanceTransportExpenseRepository,
	attendanceDayBuilder builders.AttendanceDayBuilder,
	attendanceDayRepository repositories.AttendanceDayRepository,
	monthlyAttendanceRequestBuilder builders.MonthlyAttendanceRequestBuilder,
	monthlyAttendanceRequestRepository repositories.MonthlyAttendanceRequestRepository,
) *attendanceTransportExpenseService {
	return &attendanceTransportExpenseService{
		attendanceTransportExpenseBuilder:    attendanceTransportExpenseBuilder,
		attendanceTransportExpenseRepository: attendanceTransportExpenseRepository,
		attendanceDayBuilder:                 attendanceDayBuilder,
		attendanceDayRepository:              attendanceDayRepository,
		monthlyAttendanceRequestBuilder:      monthlyAttendanceRequestBuilder,
		monthlyAttendanceRequestRepository:   monthlyAttendanceRequestRepository,
	}
}

func toAttendanceTransportExpenseResponse(
	model models.AttendanceTransportExpense,
) types.AttendanceTransportExpenseResponse {
	return types.AttendanceTransportExpenseResponse{
		ID:              model.ID,
		AttendanceDayID: model.AttendanceDayID,
		WorkDate:        model.AttendanceDay.WorkDate,
		SortOrder:       model.SortOrder,
		TransportFrom:   model.TransportFrom,
		TransportTo:     model.TransportTo,
		TransportMethod: model.TransportMethod,
		TransportAmount: model.TransportAmount,
		TransportMemo:   model.TransportMemo,
		IsDeleted:       model.IsDeleted,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
		DeletedAt:       model.DeletedAt,
	}
}

func (service *attendanceTransportExpenseService) SearchAttendanceTransportExpenses(
	userID uint,
	req types.SearchAttendanceTransportExpensesRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	if req.TargetYear <= 0 {
		return results.BadRequest(
			"SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{"targetYear": req.TargetYear},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return results.BadRequest(
			"SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{"targetMonth": req.TargetMonth},
		)
	}

	query, buildResult := service.attendanceTransportExpenseBuilder.
		BuildSearchAttendanceTransportExpensesQuery(userID, req)
	if buildResult.Error {
		return buildResult
	}

	expenses, findResult := service.attendanceTransportExpenseRepository.
		FindAttendanceTransportExpenses(query)
	if findResult.Error {
		return findResult
	}

	responseItems := make([]types.AttendanceTransportExpenseResponse, 0, len(expenses))
	for _, expense := range expenses {
		responseItems = append(responseItems, toAttendanceTransportExpenseResponse(expense))
	}

	return results.OK(
		types.SearchAttendanceTransportExpensesResponse{
			TargetYear:                  req.TargetYear,
			TargetMonth:                 req.TargetMonth,
			AttendanceTransportExpenses: responseItems,
		},
		"SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_SUCCESS",
		"日別交通費一覧を取得しました",
		nil,
	)
}

func (service *attendanceTransportExpenseService) UpdateAttendanceTransportExpensesByWorkDate(
	userID uint,
	req types.UpdateAttendanceTransportExpensesByWorkDateRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	workDate, err := utils.ParseDate(req.WorkDate)
	if err != nil {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_WORK_DATE",
			"対象日の形式が正しくありません",
			map[string]any{
				"workDate": req.WorkDate,
				"format":   "yyyy-MM-dd",
			},
		)
	}

	editableResult := service.validateMonthlyAttendanceEditable(
		userID,
		workDate.Year(),
		int(workDate.Month()),
	)
	if editableResult.Error {
		return editableResult
	}

	findAttendanceDayQuery, buildAttendanceDayResult := service.attendanceDayBuilder.
		BuildFindAttendanceDayByUserIDAndWorkDateQuery(userID, workDate)
	if buildAttendanceDayResult.Error {
		return buildAttendanceDayResult
	}

	attendanceDay, findAttendanceDayResult := service.attendanceDayRepository.
		FindAttendanceDay(findAttendanceDayQuery)
	if findAttendanceDayResult.Error {
		return findAttendanceDayResult
	}

	findCurrentQuery, buildCurrentResult := service.attendanceTransportExpenseBuilder.
		BuildFindAttendanceTransportExpensesByAttendanceDayIDQuery(attendanceDay.ID)
	if buildCurrentResult.Error {
		return buildCurrentResult
	}

	currentExpenses, findCurrentResult := service.attendanceTransportExpenseRepository.
		FindAttendanceTransportExpenses(findCurrentQuery)
	if findCurrentResult.Error {
		return findCurrentResult
	}

	currentExpenseMap := make(map[uint]models.AttendanceTransportExpense)
	for _, currentExpense := range currentExpenses {
		currentExpenseMap[currentExpense.ID] = currentExpense
	}

	requestedExpenseIDMap := make(map[uint]bool)
	savedCount := 0

	for index, expenseReq := range req.TransportExpenses {
		if validateResult := validateAttendanceTransportExpenseRequest(expenseReq, index); validateResult.Error {
			return validateResult
		}

		sortOrder := expenseReq.SortOrder
		if sortOrder <= 0 {
			sortOrder = index + 1
		}

		if expenseReq.AttendanceTransportExpenseID == nil ||
			*expenseReq.AttendanceTransportExpenseID == 0 {
			newExpense, buildCreateResult := service.attendanceTransportExpenseBuilder.
				BuildCreateAttendanceTransportExpenseModel(attendanceDay.ID, expenseReq, sortOrder)
			if buildCreateResult.Error {
				return buildCreateResult
			}

			_, createResult := service.attendanceTransportExpenseRepository.
				CreateAttendanceTransportExpense(newExpense)
			if createResult.Error {
				return createResult
			}

			savedCount++
			continue
		}

		expenseID := *expenseReq.AttendanceTransportExpenseID
		currentExpense, exists := currentExpenseMap[expenseID]
		if !exists {
			return results.BadRequest(
				"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_ID_NOT_FOUND_IN_TARGET_DAY",
				"対象日の日別交通費IDが正しくありません",
				map[string]any{
					"attendanceTransportExpenseId": expenseID,
					"workDate":                     req.WorkDate,
				},
			)
		}

		requestedExpenseIDMap[expenseID] = true

		updatedExpense, buildUpdateResult := service.attendanceTransportExpenseBuilder.
			BuildUpdateAttendanceTransportExpenseModel(currentExpense, expenseReq, sortOrder)
		if buildUpdateResult.Error {
			return buildUpdateResult
		}

		_, saveResult := service.attendanceTransportExpenseRepository.
			SaveAttendanceTransportExpense(updatedExpense)
		if saveResult.Error {
			return saveResult
		}

		savedCount++
	}

	for _, currentExpense := range currentExpenses {
		if requestedExpenseIDMap[currentExpense.ID] {
			continue
		}

		deletedExpense, buildDeleteResult := service.attendanceTransportExpenseBuilder.
			BuildDeleteAttendanceTransportExpenseModel(currentExpense)
		if buildDeleteResult.Error {
			return buildDeleteResult
		}

		_, saveResult := service.attendanceTransportExpenseRepository.
			SaveAttendanceTransportExpense(deletedExpense)
		if saveResult.Error {
			return saveResult
		}

		savedCount++
	}

	return results.OK(
		types.UpdateAttendanceTransportExpensesByWorkDateResponse{
			WorkDate:                             req.WorkDate,
			SavedAttendanceTransportExpenseCount: savedCount,
		},
		"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_SUCCESS",
		"日別交通費を保存しました",
		nil,
	)
}

func (service *attendanceTransportExpenseService) validateMonthlyAttendanceEditable(
	userID uint,
	targetYear int,
	targetMonth int,
) results.Result {
	query, buildResult := service.monthlyAttendanceRequestBuilder.
		BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
			userID,
			targetYear,
			targetMonth,
		)
	if buildResult.Error {
		return buildResult
	}

	monthlyRequest, findResult := service.monthlyAttendanceRequestRepository.
		FindMonthlyAttendanceRequest(query)

	if findResult.Error && findResult.Code == "MONTHLY_ATTENDANCE_REQUEST_NOT_FOUND" {
		return results.OK(
			nil,
			"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_MONTHLY_ATTENDANCE_NOT_SUBMITTED",
			"",
			nil,
		)
	}

	if findResult.Error {
		return findResult
	}

	if monthlyRequest.Status == "PENDING" || monthlyRequest.Status == "APPROVED" {
		return results.Conflict(
			"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_MONTHLY_ATTENDANCE_NOT_EDITABLE",
			"月次申請中または月次承認済みのため、日別交通費を変更できません",
			map[string]any{
				"targetYear":  targetYear,
				"targetMonth": targetMonth,
				"status":      monthlyRequest.Status,
			},
		)
	}

	return results.OK(
		nil,
		"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_MONTHLY_ATTENDANCE_EDITABLE",
		"",
		nil,
	)
}

func validateAttendanceTransportExpenseRequest(
	req types.UpdateAttendanceTransportExpensesByWorkDateExpenseRequest,
	index int,
) results.Result {
	if strings.TrimSpace(req.TransportFrom) == "" {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_EMPTY_TRANSPORT_FROM",
			"日別交通費の出発地を入力してください",
			map[string]any{"index": index},
		)
	}

	if strings.TrimSpace(req.TransportTo) == "" {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_EMPTY_TRANSPORT_TO",
			"日別交通費の目的地を入力してください",
			map[string]any{"index": index},
		)
	}

	if strings.TrimSpace(req.TransportMethod) == "" {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_EMPTY_TRANSPORT_METHOD",
			"日別交通費の交通手段を入力してください",
			map[string]any{"index": index},
		)
	}

	if req.TransportAmount < 0 {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_TRANSPORT_AMOUNT",
			"日別交通費の金額が正しくありません",
			map[string]any{
				"index":           index,
				"transportAmount": req.TransportAmount,
			},
		)
	}

	return results.OK(nil, "VALIDATE_ATTENDANCE_TRANSPORT_EXPENSE_REQUEST_SUCCESS", "", nil)
}
