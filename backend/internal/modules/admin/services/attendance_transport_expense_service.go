package services

import (
	"strings"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 管理者用日別交通費Service interface
 *
 * 注意：
 * ・管理者APIでは対象ユーザーIDをtargetUserIdとしてRequestで受け取る
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
type AttendanceTransportExpenseService interface {
	SearchAttendanceTransportExpenses(
		req types.SearchAttendanceTransportExpensesRequest,
	) results.Result
	UpdateAttendanceTransportExpensesByWorkDate(
		req types.UpdateAttendanceTransportExpensesByWorkDateRequest,
	) results.Result
}

/*
 * 管理者用日別交通費Service
 *
 * 月次全体保存での差分保存方針：
 * ・IDあり：既存明細を更新
 * ・IDなし：新規作成
 * ・DBに存在するがRequestから消えた明細：論理削除
 */
type attendanceTransportExpenseService struct {
	attendanceTransportExpenseBuilder    builders.AttendanceTransportExpenseBuilder
	attendanceTransportExpenseRepository repositories.AttendanceTransportExpenseRepository
	attendanceDayBuilder                 builders.AttendanceDayBuilder
	attendanceDayRepository              repositories.AttendanceDayRepository
}

/*
 * AttendanceTransportExpenseService生成
 */
func NewAttendanceTransportExpenseService(
	attendanceTransportExpenseBuilder builders.AttendanceTransportExpenseBuilder,
	attendanceTransportExpenseRepository repositories.AttendanceTransportExpenseRepository,
	attendanceDayBuilder builders.AttendanceDayBuilder,
	attendanceDayRepository repositories.AttendanceDayRepository,
) *attendanceTransportExpenseService {
	return &attendanceTransportExpenseService{
		attendanceTransportExpenseBuilder:    attendanceTransportExpenseBuilder,
		attendanceTransportExpenseRepository: attendanceTransportExpenseRepository,
		attendanceDayBuilder:                 attendanceDayBuilder,
		attendanceDayRepository:              attendanceDayRepository,
	}
}

/*
 * ModelをResponseへ変換
 */
func toAttendanceTransportExpenseResponse(
	attendanceTransportExpense models.AttendanceTransportExpense,
) types.AttendanceTransportExpenseResponse {
	return types.AttendanceTransportExpenseResponse{
		ID: attendanceTransportExpense.ID,

		AttendanceDayID: attendanceTransportExpense.AttendanceDayID,
		WorkDate:        attendanceTransportExpense.AttendanceDay.WorkDate,

		SortOrder: attendanceTransportExpense.SortOrder,

		TransportFrom:   attendanceTransportExpense.TransportFrom,
		TransportTo:     attendanceTransportExpense.TransportTo,
		TransportMethod: attendanceTransportExpense.TransportMethod,
		TransportAmount: attendanceTransportExpense.TransportAmount,
		TransportMemo:   attendanceTransportExpense.TransportMemo,

		IsDeleted: attendanceTransportExpense.IsDeleted,
		CreatedAt: attendanceTransportExpense.CreatedAt,
		UpdatedAt: attendanceTransportExpense.UpdatedAt,
		DeletedAt: attendanceTransportExpense.DeletedAt,
	}
}

/*
 * 日別交通費検索
 */
func (service *attendanceTransportExpenseService) SearchAttendanceTransportExpenses(
	req types.SearchAttendanceTransportExpensesRequest,
) results.Result {
	if req.TargetUserID == 0 {
		return results.BadRequest(
			"SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_TARGET_USER_ID",
			"対象ユーザーIDが正しくありません",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

	if req.TargetYear <= 0 {
		return results.BadRequest(
			"SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return results.BadRequest(
			"SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	query, buildResult :=
		service.attendanceTransportExpenseBuilder.
			BuildSearchAttendanceTransportExpensesQuery(req)
	if buildResult.Error {
		return buildResult
	}

	attendanceTransportExpenses, findResult :=
		service.attendanceTransportExpenseRepository.
			FindAttendanceTransportExpenses(query)
	if findResult.Error {
		return findResult
	}

	responses := make(
		[]types.AttendanceTransportExpenseResponse,
		0,
		len(attendanceTransportExpenses),
	)

	for _, attendanceTransportExpense := range attendanceTransportExpenses {
		responses = append(
			responses,
			toAttendanceTransportExpenseResponse(attendanceTransportExpense),
		)
	}

	return results.OK(
		types.SearchAttendanceTransportExpensesResponse{
			TargetUserID: req.TargetUserID,
			TargetYear:   req.TargetYear,
			TargetMonth:  req.TargetMonth,

			AttendanceTransportExpenses: responses,
		},
		"SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_SUCCESS",
		"日別交通費一覧を取得しました",
		nil,
	)
}

/*
 * 対象日の日別交通費を差分保存
 *
 * APIとして直接公開しない。
 * monthly_attendances/updateの月次全体保存から内部的に使う。
 */
func (service *attendanceTransportExpenseService) UpdateAttendanceTransportExpensesByWorkDate(
	req types.UpdateAttendanceTransportExpensesByWorkDateRequest,
) results.Result {
	if req.TargetUserID == 0 {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_INVALID_TARGET_USER_ID",
			"対象ユーザーIDが正しくありません",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
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

	findAttendanceDayQuery, buildAttendanceDayResult :=
		service.attendanceDayBuilder.
			BuildFindAttendanceDayByUserIDAndWorkDateQuery(
				req.TargetUserID,
				workDate,
			)
	if buildAttendanceDayResult.Error {
		return buildAttendanceDayResult
	}

	attendanceDay, findAttendanceDayResult :=
		service.attendanceDayRepository.
			FindAttendanceDay(findAttendanceDayQuery)
	if findAttendanceDayResult.Error {
		return findAttendanceDayResult
	}

	findCurrentQuery, buildCurrentResult :=
		service.attendanceTransportExpenseBuilder.
			BuildFindAttendanceTransportExpensesByAttendanceDayIDQuery(
				attendanceDay.ID,
			)
	if buildCurrentResult.Error {
		return buildCurrentResult
	}

	currentExpenses, findCurrentResult :=
		service.attendanceTransportExpenseRepository.
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
			newExpense, buildCreateResult :=
				service.attendanceTransportExpenseBuilder.
					BuildCreateAttendanceTransportExpenseModel(
						attendanceDay.ID,
						expenseReq,
						sortOrder,
					)
			if buildCreateResult.Error {
				return buildCreateResult
			}

			_, createResult :=
				service.attendanceTransportExpenseRepository.
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

		updatedExpense, buildUpdateResult :=
			service.attendanceTransportExpenseBuilder.
				BuildUpdateAttendanceTransportExpenseModel(
					currentExpense,
					expenseReq,
					sortOrder,
				)
		if buildUpdateResult.Error {
			return buildUpdateResult
		}

		_, saveResult :=
			service.attendanceTransportExpenseRepository.
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

		deletedExpense, buildDeleteResult :=
			service.attendanceTransportExpenseBuilder.
				BuildDeleteAttendanceTransportExpenseModel(currentExpense)
		if buildDeleteResult.Error {
			return buildDeleteResult
		}

		_, saveResult :=
			service.attendanceTransportExpenseRepository.
				SaveAttendanceTransportExpense(deletedExpense)
		if saveResult.Error {
			return saveResult
		}

		savedCount++
	}

	return results.OK(
		types.UpdateAttendanceTransportExpensesByWorkDateResponse{
			TargetUserID: req.TargetUserID,
			WorkDate:     req.WorkDate,

			SavedAttendanceTransportExpenseCount: savedCount,
		},
		"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_SUCCESS",
		"日別交通費を保存しました",
		nil,
	)
}

/*
 * 日別交通費明細入力チェック
 */
func validateAttendanceTransportExpenseRequest(
	req types.UpdateAttendanceTransportExpensesByWorkDateExpenseRequest,
	index int,
) results.Result {
	if strings.TrimSpace(req.TransportFrom) == "" {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_EMPTY_TRANSPORT_FROM",
			"日別交通費の出発地を入力してください",
			map[string]any{
				"index": index,
			},
		)
	}

	if strings.TrimSpace(req.TransportTo) == "" {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_EMPTY_TRANSPORT_TO",
			"日別交通費の目的地を入力してください",
			map[string]any{
				"index": index,
			},
		)
	}

	if strings.TrimSpace(req.TransportMethod) == "" {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_TRANSPORT_EXPENSES_EMPTY_TRANSPORT_METHOD",
			"日別交通費の交通手段を入力してください",
			map[string]any{
				"index": index,
			},
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

	return results.OK(
		nil,
		"VALIDATE_ATTENDANCE_TRANSPORT_EXPENSE_REQUEST_SUCCESS",
		"",
		nil,
	)
}
