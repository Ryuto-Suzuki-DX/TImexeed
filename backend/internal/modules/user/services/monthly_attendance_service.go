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
 */
type monthlyAttendanceService struct {
	attendanceDayService       AttendanceDayService
	attendanceBreakService     AttendanceBreakService
	monthlyCommuterPassService MonthlyCommuterPassService
}

/*
 * MonthlyAttendanceService生成
 */
func NewMonthlyAttendanceService(
	attendanceDayService AttendanceDayService,
	attendanceBreakService AttendanceBreakService,
	monthlyCommuterPassService MonthlyCommuterPassService,
) *monthlyAttendanceService {
	return &monthlyAttendanceService{
		attendanceDayService:       attendanceDayService,
		attendanceBreakService:     attendanceBreakService,
		monthlyCommuterPassService: monthlyCommuterPassService,
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
 * 休憩方針：
 * ・画面に残っている休憩だけを正とする
 * ・保存時に既存休憩を一旦削除する
 * ・その後、送られてきた休憩を作り直す
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

				RequestMemo: attendanceDayReq.RequestMemo,

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
