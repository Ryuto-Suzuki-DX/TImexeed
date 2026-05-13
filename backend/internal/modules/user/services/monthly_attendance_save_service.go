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
type MonthlyAttendanceService interface {
	UpdateMonthlyAttendance(userID uint, req types.UpdateMonthlyAttendanceRequest) results.Result
}

/*
 * 従業員用月次勤怠全体保存Service
 *
 * 役割：
 * ・月次勤怠画面の全体保存をまとめて処理する
 * ・月次通勤定期、勤怠日、休憩を分解して既存Serviceへ渡す
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・既存Serviceから返されたエラーはそのまま返す
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはしない
 * ・まずは既存Serviceを呼び出して保存処理を統一する
 * ・月次申請中、月次承認済みの保存可否は各Service側で MonthlyAttendanceRequest を見て判定する
 *
 * 画面表示用メッセージ方針：
 * ・SystemMessage はDB保存しない
 * ・月次勤怠全体保存でも SystemMessage は受け渡ししない
 * ・残業、深夜勤務、有給申請中、承認済みなどは表示時に組み立てる
 */
type monthlyAttendanceService struct {
	attendanceDayService       AttendanceDayService
	attendanceBreakService     AttendanceBreakService
	monthlyCommuterPassService MonthlyCommuterPassService
	attendanceTypeService      AttendanceTypeService
	paidLeaveService           PaidLeaveService
}

/*
 * MonthlyAttendanceService生成
 */
func NewMonthlyAttendanceService(
	attendanceDayService AttendanceDayService,
	attendanceBreakService AttendanceBreakService,
	monthlyCommuterPassService MonthlyCommuterPassService,
	attendanceTypeService AttendanceTypeService,
	paidLeaveService PaidLeaveService,
) *monthlyAttendanceService {
	return &monthlyAttendanceService{
		attendanceDayService:       attendanceDayService,
		attendanceBreakService:     attendanceBreakService,
		monthlyCommuterPassService: monthlyCommuterPassService,
		attendanceTypeService:      attendanceTypeService,
		paidLeaveService:           paidLeaveService,
	}
}

/*
 * 月次勤怠全体保存
 *
 * 保存順：
 * 1. 月次通勤定期
 * 2. 日別勤怠
 * 3. 既存休憩削除
 * 4. 休憩作成
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
 *
 * 今後の改善：
 * ・休憩も削除→新規作成ではなく、attendanceBreakId を使った update / create / delete に変える
 * ・その場合は AttendanceBreakService 側に更新用メソッドを用意してからこのServiceを差し替える
 */
func (service *monthlyAttendanceService) UpdateMonthlyAttendance(
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

	/*
	 * 有給残数チェック
	 *
	 * フロントでも有給残数0以下の場合は止めているが、
	 * バックエンドでも保存前に必ず止める。
	 */
	paidLeaveCheckResult := service.validatePaidLeaveBalanceBeforeMonthlySave(userID, req)
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

	/*
	 * 2. 日別勤怠と休憩を保存する
	 */
	for _, attendanceDayReq := range req.AttendanceDays {
		updateAttendanceDayResult := service.attendanceDayService.UpdateAttendanceDay(
			userID,
			types.UpdateAttendanceDayRequest{
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
		 * 3. 既存休憩を検索して削除する
		 *
		 * 月次勤怠画面では、画面に残っている休憩だけを正とする。
		 * そのため、対象日の休憩は一旦すべて削除し、後続で作り直す。
		 *
		 * 注意：
		 * ・ここはまだ update 方式ではない
		 * ・休憩を update 方式にするには、AttendanceBreak 側の Request にIDを持たせ、
		 *   UpdateAttendanceBreak を実装してからこの処理を変更する
		 */
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

		/*
		 * 4. リクエストで送られた休憩を作成する
		 */
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
			TargetYear:                req.TargetYear,
			TargetMonth:               req.TargetMonth,
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
 */
func (service *monthlyAttendanceService) validatePaidLeaveBalanceBeforeMonthlySave(
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
