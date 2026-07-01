package services

import (
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
)

/*
 * 従業員用月次勤怠全体保存Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type MonthlyAttendanceSaveService interface {
	UpdateMonthlyAttendance(userID uint, req types.UpdateMonthlyAttendanceRequest) results.Result
}

/*
 * 従業員用月次勤怠全体保存Service
 *
 * 役割：
 * ・月次勤怠画面の全体保存をまとめて処理する
 * ・月次通勤定期、勤怠日、日別交通費、休憩を分解して既存Serviceへ渡す
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・既存Serviceから返されたエラーはそのまま返す
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはしない
 * ・まずは既存Serviceを呼び出して保存処理を統一する
 * ・月次申請中、月次承認済みの保存可否は各Service側で MonthlyAttendanceRequest を見て判定する
 * ・予定区分は PlanAttendanceTypeID
 * ・実績状態は ActualWorkStatus
 * ・ActualAttendanceTypeID は使わない
 *
 * 画面表示用メッセージ方針：
 * ・SystemMessage はDB保存しない
 * ・月次勤怠全体保存でも SystemMessage は受け渡ししない
 * ・残業、深夜勤務、有給申請中、承認済みなどは表示時に組み立てる
 */
type monthlyAttendanceSaveService struct {
	attendanceDayService              AttendanceDayService
	attendanceTransportExpenseService AttendanceTransportExpenseService
	attendanceBreakService            AttendanceBreakService
	monthlyCommuterPassService        MonthlyCommuterPassService
	attendanceTypeService             AttendanceTypeService
	paidLeaveService                  PaidLeaveService
}

/*
 * MonthlyAttendanceSaveService生成
 */
func NewMonthlyAttendanceSaveService(
	attendanceDayService AttendanceDayService,
	attendanceTransportExpenseService AttendanceTransportExpenseService,
	attendanceBreakService AttendanceBreakService,
	monthlyCommuterPassService MonthlyCommuterPassService,
	attendanceTypeService AttendanceTypeService,
	paidLeaveService PaidLeaveService,
) *monthlyAttendanceSaveService {
	return &monthlyAttendanceSaveService{
		attendanceDayService:              attendanceDayService,
		attendanceTransportExpenseService: attendanceTransportExpenseService,
		attendanceBreakService:            attendanceBreakService,
		monthlyCommuterPassService:        monthlyCommuterPassService,
		attendanceTypeService:             attendanceTypeService,
		paidLeaveService:                  paidLeaveService,
	}
}

/*
 * 月次勤怠全体保存
 *
 * 保存順：
 * 1. 月次通勤定期
 * 2. 日別勤怠
 * 3. 日別交通費
 * 4. 既存休憩削除
 * 5. 休憩作成
 *
 * 現時点の休憩方針：
 * ・画面に残っている休憩だけを正とする
 * ・保存時に既存休憩を一旦削除する
 * ・その後、送られてきた休憩を作り直す
 *
 * 注意：
 * ・AttendanceDay は申請状態を持たない
 * ・MonthlyCommuterPass も申請状態を持たない
 * ・月次申請状態は MonthlyAttendanceRequest 側で管理する
 * ・SystemMessage は保存しない
 */
func (service *monthlyAttendanceSaveService) UpdateMonthlyAttendance(
	userID uint,
	req types.UpdateMonthlyAttendanceRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"UPDATE_MONTHLY_ATTENDANCE_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	if req.TargetYear <= 0 {
		return results.BadRequest(
			"UPDATE_MONTHLY_ATTENDANCE_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return results.BadRequest(
			"UPDATE_MONTHLY_ATTENDANCE_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	paidLeaveCheckResult := service.validatePaidLeaveBalanceBeforeMonthlySave(userID, req)
	if paidLeaveCheckResult.Error {
		return paidLeaveCheckResult
	}

	savedMonthlyCommuterPass := false
	savedAttendanceDayCount := 0
	savedAttendanceTransportExpenseCount := 0
	savedAttendanceBreakCount := 0

	if req.CommuterPass != nil {
		updateMonthlyCommuterPassResult := service.monthlyCommuterPassService.UpdateMonthlyCommuterPass(
			userID,
			types.UpdateMonthlyCommuterPassRequest{
				TargetYear:     req.TargetYear,
				TargetMonth:    req.TargetMonth,
				CommuterFrom:   req.CommuterPass.CommuterFrom,
				CommuterTo:     req.CommuterPass.CommuterTo,
				CommuterMethod: req.CommuterPass.CommuterMethod,
				CommuterAmount: req.CommuterPass.CommuterAmount,
			},
		)

		if updateMonthlyCommuterPassResult.Error {
			return updateMonthlyCommuterPassResult
		}

		savedMonthlyCommuterPass = true
	}

	for _, attendanceDayReq := range req.AttendanceDays {
		/*
		 * 初期値戻し
		 *
		 * planAttendanceTypeId = 0 は勤務区分マスタIDではなく、
		 * 対象日の勤怠を初期値へ戻すための特別値として扱う。
		 *
		 * 処理順：
		 * 1. 日別交通費を空配列で差分保存して既存明細を論理削除
		 * 2. 既存休憩を検索して削除
		 * 3. 勤怠日を論理削除
		 */
		if attendanceDayReq.PlanAttendanceTypeID == 0 {
			resetTransportExpensesResult :=
				service.attendanceTransportExpenseService.UpdateAttendanceTransportExpensesByWorkDate(
					userID,
					types.UpdateAttendanceTransportExpensesByWorkDateRequest{
						WorkDate:          attendanceDayReq.WorkDate,
						TransportExpenses: []types.UpdateAttendanceTransportExpensesByWorkDateExpenseRequest{},
					},
				)

			if resetTransportExpensesResult.Error {
				if resetTransportExpensesResult.Code == "ATTENDANCE_DAY_NOT_FOUND" {
					continue
				}

				return resetTransportExpensesResult
			}

			resetTransportExpensesResponse, ok :=
				resetTransportExpensesResult.Data.(types.UpdateAttendanceTransportExpensesByWorkDateResponse)
			if !ok {
				return results.InternalServerError(
					"UPDATE_MONTHLY_ATTENDANCE_INVALID_TRANSPORT_EXPENSE_RESET_RESPONSE",
					"日別交通費初期値戻し結果の形式が正しくありません",
					map[string]any{
						"workDate": attendanceDayReq.WorkDate,
					},
				)
			}

			savedAttendanceTransportExpenseCount +=
				resetTransportExpensesResponse.SavedAttendanceTransportExpenseCount

			searchAttendanceBreaksResult := service.attendanceBreakService.SearchAttendanceBreaks(
				userID,
				types.SearchAttendanceBreaksRequest{
					WorkDate: attendanceDayReq.WorkDate,
				},
			)

			if searchAttendanceBreaksResult.Error {
				return searchAttendanceBreaksResult
			}

			searchAttendanceBreaksResponse, ok :=
				searchAttendanceBreaksResult.Data.(types.SearchAttendanceBreaksResponse)
			if !ok {
				return results.InternalServerError(
					"UPDATE_MONTHLY_ATTENDANCE_INVALID_BREAK_RESET_SEARCH_RESPONSE",
					"休憩初期値戻し検索結果の形式が正しくありません",
					map[string]any{
						"workDate": attendanceDayReq.WorkDate,
					},
				)
			}

			for _, attendanceBreak := range searchAttendanceBreaksResponse.AttendanceBreaks {
				deleteAttendanceBreakResult := service.attendanceBreakService.DeleteAttendanceBreak(
					userID,
					types.DeleteAttendanceBreakRequest{
						WorkDate:          attendanceDayReq.WorkDate,
						AttendanceBreakID: attendanceBreak.ID,
					},
				)

				if deleteAttendanceBreakResult.Error {
					return deleteAttendanceBreakResult
				}

				savedAttendanceBreakCount++
			}

			deleteAttendanceDayResult := service.attendanceDayService.DeleteAttendanceDay(
				userID,
				types.DeleteAttendanceDayRequest{
					WorkDate: attendanceDayReq.WorkDate,
				},
			)

			if deleteAttendanceDayResult.Error {
				return deleteAttendanceDayResult
			}

			savedAttendanceDayCount++
			continue
		}

		updateAttendanceDayResult := service.attendanceDayService.UpdateAttendanceDay(
			userID,
			types.UpdateAttendanceDayRequest{
				WorkDate: attendanceDayReq.WorkDate,

				PlanAttendanceTypeID: attendanceDayReq.PlanAttendanceTypeID,
				ActualWorkStatus:     attendanceDayReq.ActualWorkStatus,

				CommonStartAt: attendanceDayReq.CommonStartAt,
				CommonEndAt:   attendanceDayReq.CommonEndAt,

				PlanStartAt: attendanceDayReq.PlanStartAt,
				PlanEndAt:   attendanceDayReq.PlanEndAt,

				ActualStartAt: attendanceDayReq.ActualStartAt,
				ActualEndAt:   attendanceDayReq.ActualEndAt,

				ScheduledWorkMinutes: attendanceDayReq.ScheduledWorkMinutes,

				RemoteWorkAllowanceFlag: attendanceDayReq.RemoteWorkAllowanceFlag,
			},
		)

		if updateAttendanceDayResult.Error {
			return updateAttendanceDayResult
		}

		savedAttendanceDayCount++

		transportExpenseRequests := make(
			[]types.UpdateAttendanceTransportExpensesByWorkDateExpenseRequest,
			0,
			len(attendanceDayReq.TransportExpenses),
		)

		for index, transportExpenseReq := range attendanceDayReq.TransportExpenses {
			sortOrder := transportExpenseReq.SortOrder
			if sortOrder <= 0 {
				sortOrder = index + 1
			}

			transportExpenseRequests = append(
				transportExpenseRequests,
				types.UpdateAttendanceTransportExpensesByWorkDateExpenseRequest{
					AttendanceTransportExpenseID: transportExpenseReq.AttendanceTransportExpenseID,
					SortOrder:                    sortOrder,
					TransportFrom:                transportExpenseReq.TransportFrom,
					TransportTo:                  transportExpenseReq.TransportTo,
					TransportMethod:              transportExpenseReq.TransportMethod,
					TransportAmount:              transportExpenseReq.TransportAmount,
					TransportMemo:                transportExpenseReq.TransportMemo,
				},
			)
		}

		updateAttendanceTransportExpensesResult :=
			service.attendanceTransportExpenseService.UpdateAttendanceTransportExpensesByWorkDate(
				userID,
				types.UpdateAttendanceTransportExpensesByWorkDateRequest{
					WorkDate:          attendanceDayReq.WorkDate,
					TransportExpenses: transportExpenseRequests,
				},
			)

		if updateAttendanceTransportExpensesResult.Error {
			return updateAttendanceTransportExpensesResult
		}

		updateAttendanceTransportExpensesResponse, ok :=
			updateAttendanceTransportExpensesResult.Data.(types.UpdateAttendanceTransportExpensesByWorkDateResponse)
		if !ok {
			return results.InternalServerError(
				"UPDATE_MONTHLY_ATTENDANCE_INVALID_TRANSPORT_EXPENSE_UPDATE_RESPONSE",
				"日別交通費保存結果の形式が正しくありません",
				map[string]any{
					"workDate": attendanceDayReq.WorkDate,
				},
			)
		}

		savedAttendanceTransportExpenseCount +=
			updateAttendanceTransportExpensesResponse.SavedAttendanceTransportExpenseCount

		searchAttendanceBreaksResult := service.attendanceBreakService.SearchAttendanceBreaks(
			userID,
			types.SearchAttendanceBreaksRequest{
				WorkDate: attendanceDayReq.WorkDate,
			},
		)

		if searchAttendanceBreaksResult.Error {
			return searchAttendanceBreaksResult
		}

		searchAttendanceBreaksResponse, ok := searchAttendanceBreaksResult.Data.(types.SearchAttendanceBreaksResponse)
		if !ok {
			return results.InternalServerError(
				"UPDATE_MONTHLY_ATTENDANCE_INVALID_BREAK_SEARCH_RESPONSE",
				"休憩検索結果の形式が正しくありません",
				map[string]any{
					"workDate": attendanceDayReq.WorkDate,
				},
			)
		}

		for _, attendanceBreak := range searchAttendanceBreaksResponse.AttendanceBreaks {
			deleteAttendanceBreakResult := service.attendanceBreakService.DeleteAttendanceBreak(
				userID,
				types.DeleteAttendanceBreakRequest{
					WorkDate:          attendanceDayReq.WorkDate,
					AttendanceBreakID: attendanceBreak.ID,
				},
			)

			if deleteAttendanceBreakResult.Error {
				return deleteAttendanceBreakResult
			}
		}

		for _, attendanceBreakReq := range attendanceDayReq.Breaks {
			createAttendanceBreakResult := service.attendanceBreakService.CreateAttendanceBreak(
				userID,
				types.CreateAttendanceBreakRequest{
					WorkDate:     attendanceDayReq.WorkDate,
					BreakStartAt: attendanceBreakReq.BreakStartAt,
					BreakEndAt:   attendanceBreakReq.BreakEndAt,
					BreakMemo:    attendanceBreakReq.BreakMemo,
				},
			)

			if createAttendanceBreakResult.Error {
				return createAttendanceBreakResult
			}

			savedAttendanceBreakCount++
		}
	}

	return results.OK(
		types.UpdateMonthlyAttendanceResponse{
			TargetYear:                           req.TargetYear,
			TargetMonth:                          req.TargetMonth,
			SavedMonthlyCommuterPass:             savedMonthlyCommuterPass,
			SavedAttendanceDayCount:              savedAttendanceDayCount,
			SavedAttendanceTransportExpenseCount: savedAttendanceTransportExpenseCount,
			SavedAttendanceBreakCount:            savedAttendanceBreakCount,
		},
		"UPDATE_MONTHLY_ATTENDANCE_SUCCESS",
		"月次勤怠を全体保存しました",
		nil,
	)
}

func (service *monthlyAttendanceSaveService) validatePaidLeaveBalanceBeforeMonthlySave(
	userID uint,
	req types.UpdateMonthlyAttendanceRequest,
) results.Result {
	searchAttendanceTypesResult := service.attendanceTypeService.SearchAttendanceTypes(types.SearchAttendanceTypesRequest{})
	if searchAttendanceTypesResult.Error {
		return searchAttendanceTypesResult
	}

	searchAttendanceTypesResponse, ok := searchAttendanceTypesResult.Data.(types.SearchAttendanceTypesResponse)
	if !ok {
		return results.InternalServerError(
			"UPDATE_MONTHLY_ATTENDANCE_INVALID_ATTENDANCE_TYPE_SEARCH_RESPONSE",
			"勤務区分マスタ検索結果の形式が正しくありません",
			nil,
		)
	}

	paidLeaveAttendanceTypeIDs := buildPaidLeaveAttendanceTypeIDMap(searchAttendanceTypesResponse.AttendanceTypes)

	if len(paidLeaveAttendanceTypeIDs) == 0 {
		return results.OK(
			nil,
			"UPDATE_MONTHLY_ATTENDANCE_PAID_LEAVE_TYPE_NOT_FOUND",
			"",
			nil,
		)
	}

	hasPaidLeave := hasPaidLeaveAttendanceDay(req, paidLeaveAttendanceTypeIDs)

	if !hasPaidLeave {
		return results.OK(
			nil,
			"UPDATE_MONTHLY_ATTENDANCE_PAID_LEAVE_NOT_INCLUDED",
			"",
			nil,
		)
	}

	getPaidLeaveBalanceResult := service.paidLeaveService.GetPaidLeaveBalance(userID)
	if getPaidLeaveBalanceResult.Error {
		return getPaidLeaveBalanceResult
	}

	paidLeaveBalanceResponse, ok := getPaidLeaveBalanceResult.Data.(types.PaidLeaveBalanceResponse)
	if !ok {
		return results.InternalServerError(
			"UPDATE_MONTHLY_ATTENDANCE_INVALID_PAID_LEAVE_BALANCE_RESPONSE",
			"有給残数取得結果の形式が正しくありません",
			nil,
		)
	}

	if paidLeaveBalanceResponse.RemainingDays <= 0 {
		return results.BadRequest(
			"UPDATE_MONTHLY_ATTENDANCE_PAID_LEAVE_BALANCE_NOT_ENOUGH",
			"有給残数がないため、有給を登録できません",
			map[string]any{
				"remainingDays": paidLeaveBalanceResponse.RemainingDays,
			},
		)
	}

	return results.OK(
		nil,
		"UPDATE_MONTHLY_ATTENDANCE_PAID_LEAVE_BALANCE_CHECK_SUCCESS",
		"",
		nil,
	)
}

func buildPaidLeaveAttendanceTypeIDMap(attendanceTypes []types.AttendanceTypeResponse) map[uint]bool {
	paidLeaveAttendanceTypeIDs := make(map[uint]bool)

	for _, attendanceType := range attendanceTypes {
		if attendanceType.Code == "PAID_LEAVE" || attendanceType.Name == "有給" {
			paidLeaveAttendanceTypeIDs[attendanceType.ID] = true
		}
	}

	return paidLeaveAttendanceTypeIDs
}

func hasPaidLeaveAttendanceDay(
	req types.UpdateMonthlyAttendanceRequest,
	paidLeaveAttendanceTypeIDs map[uint]bool,
) bool {
	for _, attendanceDayReq := range req.AttendanceDays {
		if paidLeaveAttendanceTypeIDs[attendanceDayReq.PlanAttendanceTypeID] {
			return true
		}
	}

	return false
}
