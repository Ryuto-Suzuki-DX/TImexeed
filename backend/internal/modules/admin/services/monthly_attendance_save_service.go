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
 * ・月次通勤定期、勤怠日、休憩を分解して既存Serviceへ渡す
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
 * ・日別休憩
 *
 * 保存順：
 * 1. 有給残数チェック
 * 2. 月次通勤定期
 * 3. 日別勤怠
 * 4. 対象日の休憩差分保存
 *
 * 休憩保存方針：
 * ・削除 → 全新規作成ではない
 * ・attendanceBreakId がある休憩は更新する
 * ・attendanceBreakId がない休憩は新規作成する
 * ・DBに存在するがリクエストから消えた休憩は論理削除する
 * ・この差分保存は AttendanceBreakService.UpdateAttendanceBreaksByWorkDate に任せる
 *
 * 画面表示用メッセージ方針：
 * ・SystemMessage はDB保存しない
 * ・月次勤怠全体保存でも SystemMessage は受け渡ししない
 * ・残業、深夜勤務、有給申請中、承認済みなどは表示時に組み立てる
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・各Serviceのエラー文言をここで作り直さない
 */
type monthlyAttendanceSaveService struct {
	attendanceDayService       AttendanceDayService
	attendanceBreakService     AttendanceBreakService
	monthlyCommuterPassService MonthlyCommuterPassService
	attendanceTypeService      AttendanceTypeService
	paidLeaveUsageService      PaidLeaveUsageService
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
	attendanceBreakService AttendanceBreakService,
	monthlyCommuterPassService MonthlyCommuterPassService,
	attendanceTypeService AttendanceTypeService,
	paidLeaveUsageService PaidLeaveUsageService,
) *monthlyAttendanceSaveService {
	return &monthlyAttendanceSaveService{
		attendanceDayService:       attendanceDayService,
		attendanceBreakService:     attendanceBreakService,
		monthlyCommuterPassService: monthlyCommuterPassService,
		attendanceTypeService:      attendanceTypeService,
		paidLeaveUsageService:      paidLeaveUsageService,
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
	 * ・admin側では paid_leave_usage 側の GetPaidLeaveBalance を使う
	 * ・このチェックは月次申請状態の編集ロックとは別物
	 */
	paidLeaveCheckResult := service.validatePaidLeaveBalanceBeforeMonthlySave(req)
	if paidLeaveCheckResult.Error {
		return paidLeaveCheckResult
	}

	savedMonthlyCommuterPass := false
	savedAttendanceDayCount := 0
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
	 * 2. 日別勤怠と休憩を保存する
	 */
	for _, attendanceDayReq := range req.AttendanceDays {
		updateAttendanceDayResult := service.attendanceDayService.UpdateAttendanceDay(
			types.UpdateAttendanceDayRequest{
				TargetUserID: req.TargetUserID,

				WorkDate: attendanceDayReq.WorkDate,

				PlanAttendanceTypeID:   attendanceDayReq.PlanAttendanceTypeID,
				ActualAttendanceTypeID: attendanceDayReq.ActualAttendanceTypeID,

				CommonStartAt: attendanceDayReq.CommonStartAt,
				CommonEndAt:   attendanceDayReq.CommonEndAt,

				PlanStartAt: attendanceDayReq.PlanStartAt,
				PlanEndAt:   attendanceDayReq.PlanEndAt,

				ActualStartAt: attendanceDayReq.ActualStartAt,
				ActualEndAt:   attendanceDayReq.ActualEndAt,

				LateFlag:       attendanceDayReq.LateFlag,
				EarlyLeaveFlag: attendanceDayReq.EarlyLeaveFlag,
				AbsenceFlag:    attendanceDayReq.AbsenceFlag,
				SickLeaveFlag:  attendanceDayReq.SickLeaveFlag,

				RemoteWorkAllowanceFlag: attendanceDayReq.RemoteWorkAllowanceFlag,

				TransportFrom:   attendanceDayReq.TransportFrom,
				TransportTo:     attendanceDayReq.TransportTo,
				TransportMethod: attendanceDayReq.TransportMethod,
				TransportAmount: attendanceDayReq.TransportAmount,
			},
		)

		if updateAttendanceDayResult.Error {
			return updateAttendanceDayResult
		}

		savedAttendanceDayCount++

		/*
		 * 3. 対象日の休憩を差分保存する
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

			SavedMonthlyCommuterPass:  savedMonthlyCommuterPass,
			SavedAttendanceDayCount:   savedAttendanceDayCount,
			SavedAttendanceBreakCount: savedAttendanceBreakCount,
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
 */
func (service *monthlyAttendanceSaveService) validatePaidLeaveBalanceBeforeMonthlySave(
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

	getPaidLeaveBalanceResult := service.paidLeaveUsageService.GetPaidLeaveBalance(
		types.GetPaidLeaveBalanceRequest{
			TargetUserID: req.TargetUserID,
		},
	)
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
				"targetUserId":  req.TargetUserID,
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

/*
 * 月次保存対象に有給区分が含まれているか判定する
 *
 * 予定区分・実績区分のどちらかが有給なら、有給ありと判定する。
 *
 * 注意：
 * ・PlanAttendanceTypeID は uint
 * ・ActualAttendanceTypeID は *uint のため nil チェックしてから判定する
 */
func hasPaidLeaveAttendanceDay(
	req types.UpdateMonthlyAttendanceRequest,
	paidLeaveAttendanceTypeIDs map[uint]bool,
) bool {
	for _, attendanceDayReq := range req.AttendanceDays {
		if paidLeaveAttendanceTypeIDs[attendanceDayReq.PlanAttendanceTypeID] {
			return true
		}

		if attendanceDayReq.ActualAttendanceTypeID != nil {
			if paidLeaveAttendanceTypeIDs[*attendanceDayReq.ActualAttendanceTypeID] {
				return true
			}
		}
	}

	return false
}
