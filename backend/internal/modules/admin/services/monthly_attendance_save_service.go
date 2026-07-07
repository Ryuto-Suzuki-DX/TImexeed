package services

import (
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
)

/*
 * 管理者用月次勤怠全体保存Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type MonthlyAttendanceSaveService interface {
	UpdateMonthlyAttendance(req types.UpdateMonthlyAttendanceRequest) results.Result
}

/*
 * 管理者用月次勤怠全体保存Service
 *
 * 役割：
 * ・管理者用月次勤怠画面の全体保存をまとめて処理する
 * ・月次通勤定期、勤怠日、日別交通費、休憩を分解して既存Serviceへ渡す
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・既存Serviceから返されたエラーはそのまま返す
 *
 * 重要：
 * ・DBへの直接アクセスはしない
 * ・Repository / Builder は基本的に使わない
 * ・ここでは、今まで作成したadmin用Serviceを呼び出すだけにする
 * ・管理者APIでは対象ユーザーIDを targetUserId としてRequestで受け取る
 * ・管理者側では月次申請状態による編集ロックを行わない
 *
 * 保存対象：
 * ・月次通勤定期
 * ・日別勤怠
 * ・日別交通費
 * ・日別休憩
 *
 * 保存順：
 * 1. 有給残数チェック
 * 2. 月次通勤定期
 * 3. 日別勤怠
 * 4. 対象日の日別交通費差分保存
 * 5. 対象日の休憩差分保存
 *
 * 注意：
 * ・予定区分は attendance_types を使う
 * ・実績状態は ActualWorkStatus を使う
 * ・ActualAttendanceTypeID は使わない
 */
type monthlyAttendanceSaveService struct {
	attendanceDayService              AttendanceDayService
	attendanceTransportExpenseService AttendanceTransportExpenseService
	attendanceBreakService            AttendanceBreakService
	monthlyCommuterPassService        MonthlyCommuterPassService
	attendanceTypeService             AttendanceTypeService
	paidLeaveUsageService             PaidLeaveUsageService
}

/*
 * MonthlyAttendanceService生成
 *
 * 注意：
 * ・有給残数取得は既存の PaidLeaveUsageService.GetPaidLeaveBalance を使う
 * ・別の PaidLeaveService は作らない
 */
func NewMonthlyAttendanceSaveService(
	attendanceDayService AttendanceDayService,
	attendanceTransportExpenseService AttendanceTransportExpenseService,
	attendanceBreakService AttendanceBreakService,
	monthlyCommuterPassService MonthlyCommuterPassService,
	attendanceTypeService AttendanceTypeService,
	paidLeaveUsageService PaidLeaveUsageService,
) *monthlyAttendanceSaveService {
	return &monthlyAttendanceSaveService{
		attendanceDayService:              attendanceDayService,
		attendanceTransportExpenseService: attendanceTransportExpenseService,
		attendanceBreakService:            attendanceBreakService,
		monthlyCommuterPassService:        monthlyCommuterPassService,
		attendanceTypeService:             attendanceTypeService,
		paidLeaveUsageService:             paidLeaveUsageService,
	}
}

/*
 * 月次勤怠全体保存
 *
 * 管理者が対象ユーザーの対象年月の勤怠をまとめて保存する。
 *
 * 注意：
 * ・管理者側では月次申請状態による編集ロックを行わない
 * ・ロック解除済みのadmin用Serviceだけを呼び出す
 */
func (service *monthlyAttendanceSaveService) UpdateMonthlyAttendance(
	req types.UpdateMonthlyAttendanceRequest,
) results.Result {
	if req.TargetUserID == 0 {
		return results.BadRequest(
			"UPDATE_MONTHLY_ATTENDANCE_INVALID_TARGET_USER_ID",
			"対象ユーザーIDが正しくありません",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
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

	/*
	 * 有給残数チェック
	 *
	 * フロントでも有給残数0以下の場合は止める想定だが、
	 * バックエンドでも保存前に確認する。
	 *
	 * 注意：
	 * ・有給判定は予定区分 PlanAttendanceTypeID で行う
	 * ・ActualWorkStatus は 通常/欠勤/病欠/遅刻/早退 なので有給判定には使わない
	 */
	paidLeaveAttendanceTypeIDs, paidLeaveTypeResult := service.loadPaidLeaveAttendanceTypeIDMap()
	if paidLeaveTypeResult.Error {
		return paidLeaveTypeResult
	}

	paidLeaveCheckResult := service.paidLeaveUsageService.ValidateMonthlyAttendancePaidLeaveBalance(
		req,
		paidLeaveAttendanceTypeIDs,
	)
	if paidLeaveCheckResult.Error {
		return paidLeaveCheckResult
	}

	savedMonthlyCommuterPass := false
	savedAttendanceDayCount := 0
	savedAttendanceTransportExpenseCount := 0
	savedAttendanceBreakCount := 0

	/*
	 * 1. 月次通勤定期を保存する
	 *
	 * commuterPass が nil の場合は保存しない。
	 */
	if req.CommuterPass != nil {
		updateMonthlyCommuterPassResult := service.monthlyCommuterPassService.UpdateMonthlyCommuterPass(
			types.UpdateMonthlyCommuterPassRequest{
				TargetUserID: req.TargetUserID,
				TargetYear:   req.TargetYear,
				TargetMonth:  req.TargetMonth,

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

	/*
	 * 2. 日別勤怠・日別交通費・休憩を保存する
	 */
	for _, attendanceDayReq := range req.AttendanceDays {
		/*
		 * 初期値戻し
		 *
		 * planAttendanceTypeId = 0 は勤務区分マスタIDではなく、
		 * 対象日の勤怠を初期値へ戻すための画面上の特別値として扱う。
		 *
		 * 処理順：
		 * 1. 日別交通費を空配列で差分保存し、既存明細を論理削除する
		 * 2. 休憩を空配列で差分保存し、既存明細を論理削除する
		 * 3. 勤怠日を論理削除する
		 *
		 * 対象日の勤怠がまだDBに存在しない場合は、
		 * 日別交通費差分保存が ATTENDANCE_DAY_NOT_FOUND を返すため、
		 * 保存対象なしとしてそのまま次の日へ進む。
		 */
		if attendanceDayReq.PlanAttendanceTypeID == 0 {
			resetTransportExpensesResult :=
				service.attendanceTransportExpenseService.
					UpdateAttendanceTransportExpensesByWorkDate(
						types.UpdateAttendanceTransportExpensesByWorkDateRequest{
							TargetUserID:      req.TargetUserID,
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
						"targetUserId": req.TargetUserID,
						"workDate":     attendanceDayReq.WorkDate,
					},
				)
			}

			savedAttendanceTransportExpenseCount +=
				resetTransportExpensesResponse.SavedAttendanceTransportExpenseCount

			resetBreaksResult :=
				service.attendanceBreakService.UpdateAttendanceBreaksByWorkDate(
					types.UpdateAttendanceBreaksByWorkDateRequest{
						TargetUserID: req.TargetUserID,
						WorkDate:     attendanceDayReq.WorkDate,
						Breaks:       []types.UpdateAttendanceBreaksByWorkDateBreakRequest{},
					},
				)

			if resetBreaksResult.Error {
				return resetBreaksResult
			}

			resetBreaksResponse, ok :=
				resetBreaksResult.Data.(types.UpdateAttendanceBreaksByWorkDateResponse)
			if !ok {
				return results.InternalServerError(
					"UPDATE_MONTHLY_ATTENDANCE_INVALID_BREAK_RESET_RESPONSE",
					"休憩初期値戻し結果の形式が正しくありません",
					map[string]any{
						"targetUserId": req.TargetUserID,
						"workDate":     attendanceDayReq.WorkDate,
					},
				)
			}

			savedAttendanceBreakCount +=
				resetBreaksResponse.SavedAttendanceBreakCount

			deleteAttendanceDayResult :=
				service.attendanceDayService.DeleteAttendanceDay(
					types.DeleteAttendanceDayRequest{
						TargetUserID: req.TargetUserID,
						WorkDate:     attendanceDayReq.WorkDate,
					},
				)

			if deleteAttendanceDayResult.Error {
				return deleteAttendanceDayResult
			}

			syncPaidLeaveUsageResult := service.paidLeaveUsageService.SyncAutomaticPaidLeaveUsage(
				req.TargetUserID,
				attendanceDayReq.WorkDate,
				false,
			)
			if syncPaidLeaveUsageResult.Error {
				return syncPaidLeaveUsageResult
			}

			savedAttendanceDayCount++
			continue
		}

		updateAttendanceDayResult := service.attendanceDayService.UpdateAttendanceDay(
			types.UpdateAttendanceDayRequest{
				TargetUserID: req.TargetUserID,

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

		shouldUsePaidLeave := paidLeaveAttendanceTypeIDs[attendanceDayReq.PlanAttendanceTypeID]
		syncPaidLeaveUsageResult := service.paidLeaveUsageService.SyncAutomaticPaidLeaveUsage(
			req.TargetUserID,
			attendanceDayReq.WorkDate,
			shouldUsePaidLeave,
		)
		if syncPaidLeaveUsageResult.Error {
			return syncPaidLeaveUsageResult
		}

		savedAttendanceDayCount++

		/*
		 * 3. 対象日の日別交通費を差分保存する
		 *
		 * 日別交通費の保存は AttendanceTransportExpenseService に任せる。
		 *
		 * 保存方針：
		 * ・IDあり：既存明細を更新
		 * ・IDなし：新規作成
		 * ・DBに存在するがRequestから消えた明細：論理削除
		 *
		 * 注意：
		 * ・勤怠日の保存後に実行する
		 * ・これにより、対象日のAttendanceDayが未登録だった場合でも、
		 *   先にAttendanceDayを作成してから日別交通費を紐づけられる
		 * ・管理者側では月次申請状態による編集ロックを行わない
		 */
		transportExpenseRequests := make(
			[]types.UpdateAttendanceTransportExpensesByWorkDateExpenseRequest,
			0,
			len(attendanceDayReq.TransportExpenses),
		)

		for _, transportExpenseReq := range attendanceDayReq.TransportExpenses {
			transportExpenseRequests = append(
				transportExpenseRequests,
				types.UpdateAttendanceTransportExpensesByWorkDateExpenseRequest{
					AttendanceTransportExpenseID: transportExpenseReq.AttendanceTransportExpenseID,
					SortOrder:                    transportExpenseReq.SortOrder,
					TransportFrom:                transportExpenseReq.TransportFrom,
					TransportTo:                  transportExpenseReq.TransportTo,
					TransportMethod:              transportExpenseReq.TransportMethod,
					TransportAmount:              transportExpenseReq.TransportAmount,
					TransportMemo:                transportExpenseReq.TransportMemo,
				},
			)
		}

		updateAttendanceTransportExpensesResult :=
			service.attendanceTransportExpenseService.
				UpdateAttendanceTransportExpensesByWorkDate(
					types.UpdateAttendanceTransportExpensesByWorkDateRequest{
						TargetUserID: req.TargetUserID,
						WorkDate:     attendanceDayReq.WorkDate,

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
					"targetUserId": req.TargetUserID,
					"workDate":     attendanceDayReq.WorkDate,
				},
			)
		}

		savedAttendanceTransportExpenseCount +=
			updateAttendanceTransportExpensesResponse.
				SavedAttendanceTransportExpenseCount

		/*
		 * 4. 対象日の休憩を差分保存する
		 *
		 * 休憩の保存は AttendanceBreakService に任せる。
		 * 管理者側の AttendanceBreakService は月次申請状態による編集ロックを行わない。
		 */
		breakRequests := make([]types.UpdateAttendanceBreaksByWorkDateBreakRequest, 0, len(attendanceDayReq.Breaks))
		for _, attendanceBreakReq := range attendanceDayReq.Breaks {
			breakRequests = append(breakRequests, types.UpdateAttendanceBreaksByWorkDateBreakRequest{
				AttendanceBreakID: attendanceBreakReq.AttendanceBreakID,
				BreakStartAt:      attendanceBreakReq.BreakStartAt,
				BreakEndAt:        attendanceBreakReq.BreakEndAt,
				BreakMemo:         attendanceBreakReq.BreakMemo,
			})
		}

		updateAttendanceBreaksResult := service.attendanceBreakService.UpdateAttendanceBreaksByWorkDate(
			types.UpdateAttendanceBreaksByWorkDateRequest{
				TargetUserID: req.TargetUserID,
				WorkDate:     attendanceDayReq.WorkDate,
				Breaks:       breakRequests,
			},
		)

		if updateAttendanceBreaksResult.Error {
			return updateAttendanceBreaksResult
		}

		updateAttendanceBreaksResponse, ok := updateAttendanceBreaksResult.Data.(types.UpdateAttendanceBreaksByWorkDateResponse)
		if !ok {
			return results.InternalServerError(
				"UPDATE_MONTHLY_ATTENDANCE_INVALID_BREAK_UPDATE_RESPONSE",
				"休憩保存結果の形式が正しくありません",
				map[string]any{
					"targetUserId": req.TargetUserID,
					"workDate":     attendanceDayReq.WorkDate,
				},
			)
		}

		savedAttendanceBreakCount += updateAttendanceBreaksResponse.SavedAttendanceBreakCount
	}

	return results.OK(
		types.UpdateMonthlyAttendanceResponse{
			TargetUserID: req.TargetUserID,

			TargetYear:  req.TargetYear,
			TargetMonth: req.TargetMonth,

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

/*
 * 月次勤怠保存前の有給残数チェック
 *
 * 保存対象に有給区分が含まれている場合だけ、有給残数を確認する。
 *
 * 注意：
 * ・admin側では既存の PaidLeaveUsageService.GetPaidLeaveBalance を使う
 * ・ここでは有給残数だけを確認する
 * ・月次申請状態による編集ロックは行わない
 * ・有給判定は予定区分 PlanAttendanceTypeID だけを見る
 */
func (service *monthlyAttendanceSaveService) loadPaidLeaveAttendanceTypeIDMap() (map[uint]bool, results.Result) {
	searchAttendanceTypesResult := service.attendanceTypeService.SearchAttendanceTypes(types.SearchAttendanceTypesRequest{})
	if searchAttendanceTypesResult.Error {
		return nil, searchAttendanceTypesResult
	}

	searchAttendanceTypesResponse, ok := searchAttendanceTypesResult.Data.(types.SearchAttendanceTypesResponse)
	if !ok {
		return nil, results.InternalServerError(
			"UPDATE_MONTHLY_ATTENDANCE_INVALID_ATTENDANCE_TYPE_SEARCH_RESPONSE",
			"勤務区分マスタ検索結果の形式が正しくありません",
			nil,
		)
	}

	paidLeaveAttendanceTypeIDs := buildPaidLeaveAttendanceTypeIDMap(searchAttendanceTypesResponse.AttendanceTypes)
	if len(paidLeaveAttendanceTypeIDs) == 0 {
		return nil, results.InternalServerError(
			"UPDATE_MONTHLY_ATTENDANCE_PAID_LEAVE_TYPE_NOT_FOUND",
			"有給の勤務区分が見つかりません",
			nil,
		)
	}

	return paidLeaveAttendanceTypeIDs, results.OK(
		nil,
		"LOAD_PAID_LEAVE_ATTENDANCE_TYPE_ID_MAP_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給区分IDマップ作成
 *
 * 基本は code = PAID_LEAVE をシステム判定に使う。
 * 念のため name = 有給 も対象にする。
 */
func buildPaidLeaveAttendanceTypeIDMap(attendanceTypes []types.AttendanceTypeResponse) map[uint]bool {
	paidLeaveAttendanceTypeIDs := make(map[uint]bool)

	for _, attendanceType := range attendanceTypes {
		if attendanceType.Code == "PAID_LEAVE" || attendanceType.Name == "有給" {
			paidLeaveAttendanceTypeIDs[attendanceType.ID] = true
		}
	}

	return paidLeaveAttendanceTypeIDs
}
