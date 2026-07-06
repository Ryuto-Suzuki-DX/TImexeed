package services

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"timexeed/backend/internal/constants"
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
)

/*
 * 月次勤怠集計CSV出力 Service interface
 *
 * 管理者専用。
 */
type MonthlyAttendanceSummaryExportService interface {
	ExportMonthlyAttendanceSummaryFile(request types.ExportMonthlyAttendanceSummaryCsvRequest) ([]byte, string, string, results.Result)
	ExportMonthlyAttendanceSummaryCsv(request types.ExportMonthlyAttendanceSummaryCsvRequest) ([]byte, string, results.Result)
}

const monthlyAttendanceSummaryExportContentTypeCSV = "text/csv; charset=utf-8"
const monthlyAttendanceSummaryExportContentTypeXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

/*
 * 月次勤怠集計CSV出力 Service
 *
 * 注意：
 * ・給与計算は行わず、勤怠・休暇・交通費・経費のみを集計する
 * ・APPROVED の月だけ勤怠集計値をCSVへ出力する
 * ・APPROVED 以外はステータスのみCSVへ出力する
 * ・残業、週残業、休日出勤、深夜労働は二重計上しない
 * ・変形労働制フラグは持たず、AttendanceDay.ScheduledWorkMinutes の値だけで判断する
 * ・予定区分は AttendanceDay.PlanAttendanceType を見る
 * ・実績状態は AttendanceDay.ActualWorkStatus を見る
 * ・日別交通費は AttendanceTransportExpense を AttendanceDayID 単位で集計する
 * ・ActualAttendanceTypeID / ActualAttendanceType は使わない
 */
type monthlyAttendanceSummaryExportService struct {
	monthlyAttendanceSummaryExportBuilder    builders.MonthlyAttendanceSummaryExportBuilder
	monthlyAttendanceSummaryExportRepository repositories.MonthlyAttendanceSummaryExportRepository
}

/*
 * MonthlyAttendanceSummaryExportService生成
 */
func NewMonthlyAttendanceSummaryExportService(
	monthlyAttendanceSummaryExportBuilder builders.MonthlyAttendanceSummaryExportBuilder,
	monthlyAttendanceSummaryExportRepository repositories.MonthlyAttendanceSummaryExportRepository,
) MonthlyAttendanceSummaryExportService {
	return &monthlyAttendanceSummaryExportService{
		monthlyAttendanceSummaryExportBuilder:    monthlyAttendanceSummaryExportBuilder,
		monthlyAttendanceSummaryExportRepository: monthlyAttendanceSummaryExportRepository,
	}
}

/*
 * 月次勤怠集計ファイル出力
 *
 * format が XLSX の場合は、同じ集計結果を見た目付きExcelとして出力する。
 * 未指定または CSV の場合は、従来通りCSVを出力する。
 */
func (service *monthlyAttendanceSummaryExportService) ExportMonthlyAttendanceSummaryFile(
	request types.ExportMonthlyAttendanceSummaryCsvRequest,
) ([]byte, string, string, results.Result) {
	if request.TargetYear <= 0 {
		return nil, "", "", results.BadRequest(
			"EXPORT_MONTHLY_ATTENDANCE_SUMMARY_CSV_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{
				"targetYear": request.TargetYear,
			},
		)
	}

	if request.TargetMonth < 1 || request.TargetMonth > 12 {
		return nil, "", "", results.BadRequest(
			"EXPORT_MONTHLY_ATTENDANCE_SUMMARY_CSV_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{
				"targetMonth": request.TargetMonth,
			},
		)
	}

	targetLocation := jstLocation()
	targetMonthStart := time.Date(request.TargetYear, time.Month(request.TargetMonth), 1, 0, 0, 0, 0, targetLocation)
	targetMonthEnd := targetMonthStart.AddDate(0, 1, -1)

	/*
	 * 週残業・休日出勤は月をまたぐ週も見る必要があるため、
	 * 対象月初を含む週の月曜から、対象月末を含む週の日曜まで取得する。
	 */
	extendedFromDate := startOfWeek(targetMonthStart)
	extendedToDate := startOfWeek(targetMonthEnd).AddDate(0, 0, 6)

	exportedAt := toJST(time.Now()).Format("2006-01-02 15:04:05")

	users, usersResult := service.monthlyAttendanceSummaryExportRepository.SearchExportTargetUsers(request)
	if usersResult.Error {
		return nil, "", "", usersResult
	}

	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	monthlyAttendanceRequestMap, monthlyRequestResult := service.monthlyAttendanceSummaryExportRepository.FindMonthlyAttendanceRequests(
		userIDs,
		request.TargetYear,
		request.TargetMonth,
	)
	if monthlyRequestResult.Error {
		return nil, "", "", monthlyRequestResult
	}

	attendanceDays, attendanceDaysResult := service.monthlyAttendanceSummaryExportRepository.FindAttendanceDays(
		userIDs,
		extendedFromDate,
		extendedToDate,
	)
	if attendanceDaysResult.Error {
		return nil, "", "", attendanceDaysResult
	}

	attendanceDayIDs := make([]uint, 0, len(attendanceDays))
	for _, attendanceDay := range attendanceDays {
		attendanceDayIDs = append(attendanceDayIDs, attendanceDay.ID)
	}

	attendanceBreakMap, attendanceBreakResult := service.monthlyAttendanceSummaryExportRepository.FindAttendanceBreaks(attendanceDayIDs)
	if attendanceBreakResult.Error {
		return nil, "", "", attendanceBreakResult
	}

	attendanceTransportExpenseMap, attendanceTransportExpenseResult :=
		service.monthlyAttendanceSummaryExportRepository.FindAttendanceTransportExpenses(attendanceDayIDs)
	if attendanceTransportExpenseResult.Error {
		return nil, "", "", attendanceTransportExpenseResult
	}

	monthlyCommuterPassMap, commuterPassResult := service.monthlyAttendanceSummaryExportRepository.FindMonthlyCommuterPasses(
		userIDs,
		request.TargetYear,
		request.TargetMonth,
	)
	if commuterPassResult.Error {
		return nil, "", "", commuterPassResult
	}

	paidLeaveUsageMap, paidLeaveUsageResult := service.monthlyAttendanceSummaryExportRepository.FindPaidLeaveUsages(
		userIDs,
		targetMonthStart,
		targetMonthEnd,
	)
	if paidLeaveUsageResult.Error {
		return nil, "", "", paidLeaveUsageResult
	}

	expenseMap, expenseResult := service.monthlyAttendanceSummaryExportRepository.FindExpenses(
		userIDs,
		targetMonthStart,
	)
	if expenseResult.Error {
		return nil, "", "", expenseResult
	}

	attendanceDaysByUserID := groupAttendanceDaysByUserID(attendanceDays)

	rows := make([]types.MonthlyAttendanceSummaryCsvRow, 0, len(users))
	for _, user := range users {
		monthlyAttendanceRequest, hasMonthlyAttendanceRequest := monthlyAttendanceRequestMap[user.ID]

		row := service.buildBaseCsvRow(
			user,
			monthlyAttendanceRequest,
			hasMonthlyAttendanceRequest,
			request.TargetYear,
			request.TargetMonth,
			exportedAt,
			targetMonthEnd,
		)

		if !hasMonthlyAttendanceRequest || monthlyAttendanceRequest.Status != types.MonthlyAttendanceSummaryMonthlyStatusApproved {
			if request.IncludeNotApproved {
				row.CalculationStatus = types.MonthlyAttendanceSummaryCalculationStatusSkippedNotApproved
				row.HasMonthlyApprovalWarning = true
				row.HasDataWarning = true
				row.Warnings = buildWarningText([]string{
					"月次承認警告: 月次申請が承認済みではないため、勤怠集計値は出力していません",
				})
				row.WarningCount = 1
				rows = append(rows, row)
			}

			continue
		}

		calculatedRow := service.calculateApprovedUserRow(
			row,
			attendanceDaysByUserID[user.ID],
			attendanceBreakMap,
			attendanceTransportExpenseMap,
			monthlyCommuterPassMap[user.ID],
			paidLeaveUsageMap[user.ID],
			expenseMap[user.ID],
			targetMonthStart,
			targetMonthEnd,
		)

		rows = append(rows, calculatedRow)
	}

	format := normalizeMonthlyAttendanceSummaryExportFormat(request.Format)
	if format == types.MonthlyAttendanceSummaryExportFormatXLSX {
		excelBytes, excelResult := service.monthlyAttendanceSummaryExportBuilder.BuildExcel(
			rows,
			request.TargetYear,
			request.TargetMonth,
		)
		if excelResult.Error {
			return nil, "", "", excelResult
		}

		fileName := service.monthlyAttendanceSummaryExportBuilder.BuildExcelFileName(request.TargetYear, request.TargetMonth)

		return excelBytes, fileName, monthlyAttendanceSummaryExportContentTypeXLSX, results.OK(
			types.ExportMonthlyAttendanceSummaryCsvResponse{
				FileName:    fileName,
				TargetYear:  request.TargetYear,
				TargetMonth: request.TargetMonth,
				RowCount:    len(rows),
			},
			"EXPORT_MONTHLY_ATTENDANCE_SUMMARY_EXCEL_SUCCESS",
			"月次勤怠集計Excelを出力しました",
			nil,
		)
	}

	csvBytes, csvResult := service.monthlyAttendanceSummaryExportBuilder.BuildCSV(rows)
	if csvResult.Error {
		return nil, "", "", csvResult
	}

	fileName := service.monthlyAttendanceSummaryExportBuilder.BuildFileName(request.TargetYear, request.TargetMonth)

	return csvBytes, fileName, monthlyAttendanceSummaryExportContentTypeCSV, results.OK(
		types.ExportMonthlyAttendanceSummaryCsvResponse{
			FileName:    fileName,
			TargetYear:  request.TargetYear,
			TargetMonth: request.TargetMonth,
			RowCount:    len(rows),
		},
		"EXPORT_MONTHLY_ATTENDANCE_SUMMARY_CSV_SUCCESS",
		"月次勤怠集計CSVを出力しました",
		nil,
	)
}

/*
 * 月次勤怠集計CSV出力
 *
 * 既存呼び出しとの互換用。
 */
func (service *monthlyAttendanceSummaryExportService) ExportMonthlyAttendanceSummaryCsv(
	request types.ExportMonthlyAttendanceSummaryCsvRequest,
) ([]byte, string, results.Result) {
	request.Format = types.MonthlyAttendanceSummaryExportFormatCSV
	csvBytes, fileName, _, result := service.ExportMonthlyAttendanceSummaryFile(request)
	return csvBytes, fileName, result
}

/*
 * CSV基本行生成
 */
func (service *monthlyAttendanceSummaryExportService) buildBaseCsvRow(
	user repositories.MonthlyAttendanceSummaryExportUserRecord,
	monthlyAttendanceRequest models.MonthlyAttendanceRequest,
	hasMonthlyAttendanceRequest bool,
	targetYear int,
	targetMonth int,
	exportedAt string,
	targetMonthEnd time.Time,
) types.MonthlyAttendanceSummaryCsvRow {
	departmentID := uint(0)
	if user.DepartmentID != nil {
		departmentID = *user.DepartmentID
	}

	departmentName := ""
	if user.DepartmentName != nil {
		departmentName = *user.DepartmentName
	}

	retirementDate := ""
	if user.RetirementDate != nil {
		retirementDate = formatDate(*user.RetirementDate)
	}

	row := types.MonthlyAttendanceSummaryCsvRow{
		ExportTargetYear:                 targetYear,
		ExportTargetMonth:                targetMonth,
		ExportedAt:                       exportedAt,
		ExportStatus:                     types.MonthlyAttendanceSummaryExportStatusRowOutput,
		CalculationStatus:                types.MonthlyAttendanceSummaryCalculationStatusCalculated,
		UserID:                           user.ID,
		UserName:                         user.Name,
		UserEmail:                        user.Email,
		DepartmentID:                     departmentID,
		DepartmentName:                   departmentName,
		Role:                             user.Role,
		HireDate:                         formatDate(user.HireDate),
		RetirementDate:                   retirementDate,
		IsRetiredInTargetMonth:           user.RetirementDate != nil && !user.RetirementDate.After(targetMonthEnd),
		MonthlyStatus:                    types.MonthlyAttendanceSummaryMonthlyStatusNotSubmitted,
		CompanyDailyStandardWorkMinutes:  constants.CompanyDailyStandardWorkMinutes,
		CompanyWeeklyStandardWorkMinutes: constants.CompanyWeeklyStandardWorkMinutes,
	}

	if !hasMonthlyAttendanceRequest {
		return row
	}

	row.MonthlyRequestID = monthlyAttendanceRequest.ID
	row.MonthlyStatus = monthlyAttendanceRequest.Status
	row.RequestMemo = stringPtrValue(monthlyAttendanceRequest.RequestMemo)
	row.RequestedAt = timePtrValue(monthlyAttendanceRequest.RequestedAt)
	row.ApprovedBy = uintPtrValue(monthlyAttendanceRequest.ApprovedBy)
	row.ApprovedAt = timePtrValue(monthlyAttendanceRequest.ApprovedAt)
	row.RejectedReason = stringPtrValue(monthlyAttendanceRequest.RejectedReason)
	row.RejectedAt = timePtrValue(monthlyAttendanceRequest.RejectedAt)
	row.CanceledReason = stringPtrValue(monthlyAttendanceRequest.CanceledReason)
	row.CanceledAt = timePtrValue(monthlyAttendanceRequest.CanceledAt)

	return row
}

/*
 * APPROVEDユーザー行の集計
 */
func (service *monthlyAttendanceSummaryExportService) calculateApprovedUserRow(
	row types.MonthlyAttendanceSummaryCsvRow,
	attendanceDays []models.AttendanceDay,
	attendanceBreakMap map[uint][]models.AttendanceBreak,
	attendanceTransportExpenseMap map[uint][]models.AttendanceTransportExpense,
	monthlyCommuterPass models.MonthlyCommuterPass,
	paidLeaveUsages []models.PaidLeaveUsage,
	expenses []models.Expense,
	targetMonthStart time.Time,
	targetMonthEnd time.Time,
) types.MonthlyAttendanceSummaryCsvRow {
	row.CalculationStatus = types.MonthlyAttendanceSummaryCalculationStatusCalculated
	row.CalendarDays = targetMonthEnd.Day()
	row.WorkingDayCount = countWeekdays(targetMonthStart, targetMonthEnd)

	workRows := service.buildWorkRows(
		attendanceDays,
		attendanceBreakMap,
		attendanceTransportExpenseMap,
	)
	service.applyHolidayWorkFlags(workRows)
	service.applyDailyOvertime(workRows)
	weeklyWorks := service.applyWeeklyOvertime(workRows, targetMonthStart, targetMonthEnd)

	paidLeaveMinutesByDate := service.buildPaidLeaveMinutesByDate(paidLeaveUsages, attendanceDays)

	attendanceDaysInMonth := 0
	dateWithAttendance := map[string]bool{}
	warnings := []string{}

	for _, workRow := range workRows {
		workDate, err := parseDate(workRow.WorkDate)
		if err != nil {
			continue
		}

		if workDate.Before(targetMonthStart) || workDate.After(targetMonthEnd) {
			continue
		}

		attendanceDaysInMonth++
		dateWithAttendance[workRow.WorkDate] = true

		row.ScheduledWorkMinutes += workRow.ScheduledWorkMinutes
		if workRow.ScheduledWorkMinutes > 0 {
			row.ScheduledWorkDays++
		}

		row.ActualWorkMinutes += workRow.ActualWorkMinutes
		row.DayWorkMinutes += workRow.DayWorkMinutes
		row.NightWorkMinutes += workRow.NightWorkMinutes
		row.BreakMinutes += workRow.BreakMinutes
		row.RegularWorkMinutes += workRow.RegularWorkMinutes
		row.LateNightWorkMinutes += workRow.LateNightWorkMinutes

		if workRow.IsActualWorkDay {
			row.ActualWorkDays++
			if workRow.NightWorkMinutes > 0 {
				row.NightShiftWorkDays++
			} else {
				row.DayShiftWorkDays++
			}
		}

		if workRow.IsPlannedHoliday {
			row.PlannedHolidayDays++
		}

		if workRow.IsHolidayWorkDay {
			row.HolidayWorkDays++
			row.HolidayWorkMinutes += workRow.ActualWorkMinutes
			row.HolidayLateNightWorkMinutes += workRow.HolidayLateNightWorkMinutes
		}

		if workRow.IsPaidLeaveDay {
			row.PaidLeaveDays++
		}

		if workRow.IsHalfPaidLeaveDay {
			row.HalfPaidLeaveDays++
		}

		if workRow.IsAbsenceDay {
			row.AbsenceDays++
			row.AbsenceMinutes += workRow.AbsenceMinutes
		}

		if workRow.IsSickLeaveDay {
			row.SickLeaveDays++
			row.SickLeaveMinutes += workRow.SickLeaveMinutes
		}

		if workRow.ActualWorkStatus == constants.ActualWorkStatusLate {
			row.LateDays++
		}

		if workRow.ActualWorkStatus == constants.ActualWorkStatusEarlyLeave {
			row.EarlyLeaveDays++
		}

		row.DailyOvertimeThresholdMinutes += workRow.DailyOvertimeThresholdMinutes
		row.DailyOvertimeMinutes += workRow.DailyOvertimeMinutes
		row.DayOvertimeMinutes += workRow.DayOvertimeMinutes
		row.NightOvertimeMinutes += workRow.NightOvertimeMinutes
		row.WorkShortageMinutes += workRow.WorkShortageMinutes
		row.WorkExcessAgainstScheduledMinutes += workRow.WorkExcessAgainstScheduledMinutes
		row.LateMinutes += workRow.LateMinutes
		row.EarlyLeaveMinutes += workRow.EarlyLeaveMinutes
		if workRow.IsScheduledButNoActual {
			row.ScheduledButNoActualDays++
		}
		if workRow.IsActualButNoScheduled {
			row.ActualButNoScheduledDays++
		}
		if workRow.IsMissingScheduledWork {
			row.MissingScheduledWorkDays++
		}

		row.DailyTransportationAmount += workRow.TransportAmount
		row.DailyTransportationCount += workRow.TransportCount

		for _, warning := range workRow.Warnings {
			warnings = append(warnings, warning)
		}
	}

	for _, weeklyWork := range weeklyWorks {
		row.WeeklyScheduledWorkMinutes += weeklyWork.ScheduledWorkMinutes
		row.WeeklyOvertimeThresholdMinutes += weeklyWork.WeeklyOvertimeThresholdMinutes
		row.WeeklyOvertimeMinutes += weeklyWork.WeeklyOvertimeMinutes
	}

	row.OvertimeMinutes = row.DailyOvertimeMinutes + row.WeeklyOvertimeMinutes
	row.DeductionTargetMinutes = row.WorkShortageMinutes

	row.RegisteredAttendanceDays = attendanceDaysInMonth
	row.MissingAttendanceDays = row.CalendarDays - len(dateWithAttendance)
	if row.MissingAttendanceDays < 0 {
		row.MissingAttendanceDays = 0
	}

	service.applyCommuterPassToRow(&row, monthlyCommuterPass)

	paidLeaveUsedDays := 0.0
	paidLeaveUsedMinutes := 0
	for _, paidLeaveUsage := range paidLeaveUsages {
		paidLeaveUsedDays += paidLeaveUsage.UsageDays
		paidLeaveUsedMinutes += paidLeaveMinutesByDate[formatDate(paidLeaveUsage.UsageDate)]
	}

	row.PaidLeaveUsedDays = paidLeaveUsedDays
	row.PaidLeaveUsedMinutes = paidLeaveUsedMinutes
	row.PaidLeaveMinutes = paidLeaveUsedMinutes

	service.applyExpensesToRow(&row, expenses)

	row.TotalTransportationAmount = row.DailyTransportationAmount + row.CommuterPassAmount

	service.applyActualOperationRateToRow(&row)
	service.applyDataWarningFlagsToRow(&row, &warnings)

	row.InvalidBreakCount = countWarningsByPrefix(warnings, "休憩")
	row.InvalidTimeCount = countWarningsByPrefix(warnings, "時刻")
	row.WarningCount = len(warnings)
	row.Warnings = buildWarningText(warnings)
	row.HasDataWarning = row.WarningCount > 0

	return row
}

/*
 * 日別作業行生成
 */
func (service *monthlyAttendanceSummaryExportService) buildWorkRows(
	attendanceDays []models.AttendanceDay,
	attendanceBreakMap map[uint][]models.AttendanceBreak,
	attendanceTransportExpenseMap map[uint][]models.AttendanceTransportExpense,
) []*types.MonthlyAttendanceSummaryWorkRow {
	workRows := make([]*types.MonthlyAttendanceSummaryWorkRow, 0, len(attendanceDays))

	for _, attendanceDay := range attendanceDays {
		workDate := formatDate(attendanceDay.WorkDate)
		scheduledWorkMinutes := intPtrValue(attendanceDay.ScheduledWorkMinutes)

		transportAmount := 0
		transportCount := 0

		for _, attendanceTransportExpense := range attendanceTransportExpenseMap[attendanceDay.ID] {
			if attendanceTransportExpense.IsDeleted {
				continue
			}

			transportAmount += attendanceTransportExpense.TransportAmount
			transportCount++
		}

		workRow := &types.MonthlyAttendanceSummaryWorkRow{
			UserID:                     attendanceDay.UserID,
			WorkDate:                   workDate,
			AttendanceDayID:            attendanceDay.ID,
			PlanAttendanceTypeID:       attendanceDay.PlanAttendanceTypeID,
			PlanAttendanceTypeCode:     attendanceDay.PlanAttendanceType.Code,
			PlanAttendanceTypeCategory: attendanceDay.PlanAttendanceType.Category,
			ActualWorkStatus:           attendanceDay.ActualWorkStatus,
			ScheduledWorkMinutes:       scheduledWorkMinutes,
			TransportAmount:            transportAmount,
			TransportCount:             transportCount,
			IsPlannedHoliday:           isHolidayAttendanceType(attendanceDay.PlanAttendanceType),
			IsPaidLeaveDay:             isPaidLeaveAttendanceType(attendanceDay.PlanAttendanceType),
			IsAbsenceDay:               attendanceDay.ActualWorkStatus == constants.ActualWorkStatusAbsence,
			IsSickLeaveDay:             attendanceDay.ActualWorkStatus == constants.ActualWorkStatusSickLeave,
		}

		if workRow.IsPaidLeaveDay && scheduledWorkMinutes > 0 {
			if scheduledWorkMinutes <= constants.CompanyDailyStandardWorkMinutes/2 {
				workRow.IsHalfPaidLeaveDay = true
			}
		}

		if !workRow.IsPlannedHoliday && !workRow.IsPaidLeaveDay && scheduledWorkMinutes == 0 {
			workRow.IsMissingScheduledWork = true
			workRow.Warnings = append(workRow.Warnings, workDate+" 予定不整合: 予定労働時間が未設定です")
		}

		if workRow.IsAbsenceDay {
			workRow.AbsenceMinutes = scheduledWorkMinutes
		}

		if workRow.IsSickLeaveDay {
			workRow.SickLeaveMinutes = scheduledWorkMinutes
		}

		if attendanceDay.ActualStartAt == nil && attendanceDay.ActualEndAt == nil {
			if scheduledWorkMinutes > 0 && !workRow.IsAbsenceDay && !workRow.IsSickLeaveDay && !workRow.IsPaidLeaveDay {
				workRow.IsScheduledButNoActual = true
				workRow.WorkShortageMinutes = scheduledWorkMinutes
				workRow.Warnings = append(workRow.Warnings, workDate+" 実績未入力: 予定労働時間がありますが実績時刻が未入力です")
			}

			workRows = append(workRows, workRow)
			continue
		}

		if attendanceDay.ActualStartAt == nil || attendanceDay.ActualEndAt == nil {
			workRow.Warnings = append(workRow.Warnings, workDate+" 時刻不整合: 実績開始または実績終了が未入力です")
			workRows = append(workRows, workRow)
			continue
		}

		actualStartAt := toJST(*attendanceDay.ActualStartAt)
		actualEndAt := toJST(*attendanceDay.ActualEndAt)

		if !actualEndAt.After(actualStartAt) {
			workRow.Warnings = append(workRow.Warnings, workDate+" 時刻不整合: 実績終了が実績開始以前です")
			workRows = append(workRows, workRow)
			continue
		}

		validBreaks := make([]models.AttendanceBreak, 0, len(attendanceBreakMap[attendanceDay.ID]))
		validBreakMinutes := 0

		for _, attendanceBreak := range attendanceBreakMap[attendanceDay.ID] {
			breakStartAt, breakEndAt, normalized := normalizeBreakPeriodToActualWork(
				toJST(attendanceBreak.BreakStartAt),
				toJST(attendanceBreak.BreakEndAt),
				actualStartAt,
				actualEndAt,
			)

			if !normalized {
				workRow.Warnings = append(workRow.Warnings, workDate+" 休憩不整合: 休憩が実績勤務時間内に収まりません")
				continue
			}

			convertedBreak := attendanceBreak
			convertedBreak.BreakStartAt = breakStartAt
			convertedBreak.BreakEndAt = breakEndAt

			validBreaks = append(validBreaks, convertedBreak)
			validBreakMinutes += minutesBetween(breakStartAt, breakEndAt)
		}

		grossActualMinutes := minutesBetween(actualStartAt, actualEndAt)
		actualWorkMinutes := grossActualMinutes - validBreakMinutes
		if actualWorkMinutes < 0 {
			actualWorkMinutes = 0
			workRow.Warnings = append(workRow.Warnings, workDate+" 時刻不整合: 休憩時間が実績勤務時間を超えています")
		}

		workRow.BreakMinutes = validBreakMinutes
		workRow.ActualWorkMinutes = actualWorkMinutes

		/*
		 * 実勤務日判定
		 *
		 * 実績状態が欠勤・病欠の場合は実勤務日にしない。
		 * 遅刻・早退は実勤務ありとして扱う。
		 */
		workRow.IsActualWorkDay = actualWorkMinutes > 0 &&
			attendanceDay.ActualWorkStatus != constants.ActualWorkStatusAbsence &&
			attendanceDay.ActualWorkStatus != constants.ActualWorkStatusSickLeave

		if actualWorkMinutes > 0 && scheduledWorkMinutes == 0 && !workRow.IsPlannedHoliday {
			workRow.IsActualButNoScheduled = true
			workRow.Warnings = append(workRow.Warnings, workDate+" 予定不整合: 実績がありますが予定労働時間が未設定です")
		}

		if scheduledWorkMinutes > 0 && actualWorkMinutes == 0 && !workRow.IsAbsenceDay && !workRow.IsSickLeaveDay && !workRow.IsPaidLeaveDay {
			workRow.IsScheduledButNoActual = true
			workRow.Warnings = append(workRow.Warnings, workDate+" 実績未入力: 予定労働時間がありますが有効な実績労働時間がありません")
		}

		if attendanceDay.ActualWorkStatus == constants.ActualWorkStatusLate {
			workRow.LateMinutes = calculateLateMinutes(attendanceDay.PlanStartAt, attendanceDay.ActualStartAt)
			if workRow.LateMinutes == 0 && scheduledWorkMinutes > actualWorkMinutes {
				workRow.LateMinutes = scheduledWorkMinutes - actualWorkMinutes
			}
		}

		if attendanceDay.ActualWorkStatus == constants.ActualWorkStatusEarlyLeave {
			workRow.EarlyLeaveMinutes = calculateEarlyLeaveMinutes(attendanceDay.PlanEndAt, attendanceDay.ActualEndAt)
			if workRow.EarlyLeaveMinutes == 0 && scheduledWorkMinutes > actualWorkMinutes {
				workRow.EarlyLeaveMinutes = scheduledWorkMinutes - actualWorkMinutes
			}
		}

		workSegments := buildWorkSegments(actualStartAt, actualEndAt, validBreaks)
		service.applyDayNightWorkMinutes(workRow, workSegments)

		workRows = append(workRows, workRow)
	}

	return workRows
}

/*
 * 休日出勤判定
 *
 * 1. 予定区分が休日の日に実績勤務がある場合は休日出勤
 * 2. 予定休日が明示されていない週で、月曜起算の1週間に1日も休みがない場合、
 *    週の最後の勤務日を休日出勤扱いにする
 */
func (service *monthlyAttendanceSummaryExportService) applyHolidayWorkFlags(
	workRows []*types.MonthlyAttendanceSummaryWorkRow,
) {
	weeks := groupWorkRowsByWeek(workRows)

	for _, weekRows := range weeks {
		sort.Slice(weekRows, func(i int, j int) bool {
			return weekRows[i].WorkDate < weekRows[j].WorkDate
		})

		hasHolidayWorkByPlan := false
		actualWorkCount := 0
		var lastActualWorkRow *types.MonthlyAttendanceSummaryWorkRow

		for _, workRow := range weekRows {
			if workRow.IsActualWorkDay {
				actualWorkCount++
				lastActualWorkRow = workRow
			}

			if workRow.IsPlannedHoliday && workRow.IsActualWorkDay {
				workRow.IsHolidayWorkDay = true
				hasHolidayWorkByPlan = true
			}
		}

		if hasHolidayWorkByPlan {
			continue
		}

		if actualWorkCount >= 7 && lastActualWorkRow != nil {
			lastActualWorkRow.IsHolidayWorkDay = true
		}
	}
}

/*
 * 日別残業計算
 */
func (service *monthlyAttendanceSummaryExportService) applyDailyOvertime(
	workRows []*types.MonthlyAttendanceSummaryWorkRow,
) {
	for _, workRow := range workRows {
		if !workRow.IsActualWorkDay {
			continue
		}

		workRow.DailyOvertimeThresholdMinutes = maxInt(
			constants.CompanyDailyStandardWorkMinutes,
			workRow.ScheduledWorkMinutes,
		)

		workRow.WorkShortageMinutes = maxInt(
			workRow.ScheduledWorkMinutes-workRow.ActualWorkMinutes,
			0,
		)

		workRow.WorkExcessAgainstScheduledMinutes = maxInt(
			workRow.ActualWorkMinutes-workRow.ScheduledWorkMinutes,
			0,
		)

		if workRow.IsHolidayWorkDay {
			workRow.RegularWorkMinutes = 0
			workRow.DailyOvertimeMinutes = 0
			workRow.DayOvertimeMinutes = 0
			workRow.NightOvertimeMinutes = 0
			workRow.HolidayLateNightWorkMinutes = workRow.NightWorkMinutes
			continue
		}

		workRow.RegularWorkMinutes = minInt(workRow.ActualWorkMinutes, workRow.DailyOvertimeThresholdMinutes)
		workRow.DailyOvertimeMinutes = maxInt(
			workRow.ActualWorkMinutes-workRow.DailyOvertimeThresholdMinutes,
			0,
		)

		service.applyDayNightOvertimeMinutes(workRow)
	}
}

/*
 * 週別残業計算
 *
 * 日別残業・休日出勤を除外したうえで週基準超過分だけを週残業にする。
 * 週をまたいだ月境界でも二重計上しないため、日ごとに超過発生分を割り当てる。
 */
func (service *monthlyAttendanceSummaryExportService) applyWeeklyOvertime(
	workRows []*types.MonthlyAttendanceSummaryWorkRow,
	targetMonthStart time.Time,
	targetMonthEnd time.Time,
) []types.MonthlyAttendanceSummaryWeekWork {
	weeks := groupWorkRowsByWeek(workRows)
	weekWorks := make([]types.MonthlyAttendanceSummaryWeekWork, 0, len(weeks))

	for weekStartDate, weekRows := range weeks {
		sort.Slice(weekRows, func(i int, j int) bool {
			return weekRows[i].WorkDate < weekRows[j].WorkDate
		})

		weeklyScheduledWorkMinutes := 0
		weeklyActualWorkMinutes := 0
		weeklyDailyOvertimeMinutes := 0
		weeklyHolidayWorkMinutes := 0

		for _, workRow := range weekRows {
			weeklyScheduledWorkMinutes += workRow.ScheduledWorkMinutes
			weeklyActualWorkMinutes += workRow.ActualWorkMinutes
			weeklyDailyOvertimeMinutes += workRow.DailyOvertimeMinutes
			if workRow.IsHolidayWorkDay {
				weeklyHolidayWorkMinutes += workRow.ActualWorkMinutes
			}
		}

		weeklyOvertimeThresholdMinutes := maxInt(
			constants.CompanyWeeklyStandardWorkMinutes,
			weeklyScheduledWorkMinutes,
		)

		cumulativeBaseWorkMinutes := 0
		weeklyOvertimeMinutesInTargetMonth := 0

		for _, workRow := range weekRows {
			workDate, err := parseDate(workRow.WorkDate)
			if err != nil {
				continue
			}

			dayBaseWorkMinutes := workRow.ActualWorkMinutes - workRow.DailyOvertimeMinutes
			if workRow.IsHolidayWorkDay {
				dayBaseWorkMinutes -= workRow.ActualWorkMinutes
			}

			if dayBaseWorkMinutes < 0 {
				dayBaseWorkMinutes = 0
			}

			beforeExceededMinutes := maxInt(cumulativeBaseWorkMinutes-weeklyOvertimeThresholdMinutes, 0)
			cumulativeBaseWorkMinutes += dayBaseWorkMinutes
			afterExceededMinutes := maxInt(cumulativeBaseWorkMinutes-weeklyOvertimeThresholdMinutes, 0)

			dayWeeklyOvertimeMinutes := afterExceededMinutes - beforeExceededMinutes
			if dayWeeklyOvertimeMinutes < 0 {
				dayWeeklyOvertimeMinutes = 0
			}

			// 週40時間超過分も、その日の勤務区間に沿って日中／深夜へ配賦する。
			// 日別残業部分はすでに除外されているため、二重計上しない。
			allocateWeeklyOvertimeByDayNight(workRow, dayWeeklyOvertimeMinutes)

			if !workDate.Before(targetMonthStart) && !workDate.After(targetMonthEnd) {
				weeklyOvertimeMinutesInTargetMonth += dayWeeklyOvertimeMinutes
			}
		}

		weekStart, _ := parseDate(weekStartDate)
		weekWorks = append(weekWorks, types.MonthlyAttendanceSummaryWeekWork{
			WeekStartDate:                  weekStartDate,
			WeekEndDate:                    formatDate(weekStart.AddDate(0, 0, 6)),
			ScheduledWorkMinutes:           weeklyScheduledWorkMinutes,
			ActualWorkMinutes:              weeklyActualWorkMinutes,
			DailyOvertimeMinutes:           weeklyDailyOvertimeMinutes,
			HolidayWorkMinutes:             weeklyHolidayWorkMinutes,
			WeeklyOvertimeThresholdMinutes: weeklyOvertimeThresholdMinutes,
			WeeklyOvertimeMinutes:          weeklyOvertimeMinutesInTargetMonth,
		})
	}

	sort.Slice(weekWorks, func(i int, j int) bool {
		return weekWorks[i].WeekStartDate < weekWorks[j].WeekStartDate
	})

	return weekWorks
}

/*
 * 月次通勤定期をCSV行へ反映
 */
func (service *monthlyAttendanceSummaryExportService) applyCommuterPassToRow(
	row *types.MonthlyAttendanceSummaryCsvRow,
	monthlyCommuterPass models.MonthlyCommuterPass,
) {
	row.CommuterPassFrom = stringPtrValue(monthlyCommuterPass.CommuterFrom)
	row.CommuterPassTo = stringPtrValue(monthlyCommuterPass.CommuterTo)
	row.CommuterPassMethod = stringPtrValue(monthlyCommuterPass.CommuterMethod)
	row.CommuterPassAmount = intPtrValue(monthlyCommuterPass.CommuterAmount)
}

/*
 * 有給使用日ごとの有給換算時間生成
 */
func (service *monthlyAttendanceSummaryExportService) buildPaidLeaveMinutesByDate(
	paidLeaveUsages []models.PaidLeaveUsage,
	attendanceDays []models.AttendanceDay,
) map[string]int {
	scheduledWorkMinutesByDate := map[string]int{}
	for _, attendanceDay := range attendanceDays {
		scheduledWorkMinutesByDate[formatDate(attendanceDay.WorkDate)] = intPtrValue(attendanceDay.ScheduledWorkMinutes)
	}

	paidLeaveMinutesByDate := map[string]int{}
	for _, paidLeaveUsage := range paidLeaveUsages {
		usageDate := formatDate(paidLeaveUsage.UsageDate)
		scheduledWorkMinutes := scheduledWorkMinutesByDate[usageDate]
		if scheduledWorkMinutes == 0 {
			scheduledWorkMinutes = constants.CompanyDailyStandardWorkMinutes
		}

		paidLeaveMinutesByDate[usageDate] += int(math.Round(float64(scheduledWorkMinutes) * paidLeaveUsage.UsageDays))
	}

	return paidLeaveMinutesByDate
}

/*
 * 実労働稼働率をCSV行へ反映
 *
 * 実績労働時間 ÷ 予定労働時間 × 100
 *
 * 給与計算は行わないため、給与対象稼働率・給与判定は設定しない。
 */
func (service *monthlyAttendanceSummaryExportService) applyActualOperationRateToRow(
	row *types.MonthlyAttendanceSummaryCsvRow,
) {
	if row.ScheduledWorkMinutes <= 0 {
		return
	}

	row.ActualOperationRate = roundToOneDecimal(
		float64(row.ActualWorkMinutes) / float64(row.ScheduledWorkMinutes) * 100,
	)
}

/*
 * 勤怠集計向けのデータ警告をCSV行へ反映
 */
func (service *monthlyAttendanceSummaryExportService) applyDataWarningFlagsToRow(
	row *types.MonthlyAttendanceSummaryCsvRow,
	warnings *[]string,
) {
	if row.MissingAttendanceDays > 0 {
		row.HasAttendanceMissingWarning = true
		*warnings = append(*warnings, fmt.Sprintf("勤怠未登録警告: 対象月に勤怠未登録日が%d日あります", row.MissingAttendanceDays))
	}

	if row.ScheduledButNoActualDays > 0 || row.ActualButNoScheduledDays > 0 || row.MissingScheduledWorkDays > 0 {
		row.HasScheduleActualMismatchWarning = true
	}

	if row.ScheduledButNoActualDays > 0 {
		*warnings = append(*warnings, fmt.Sprintf("予定実績不整合: 予定あり実績なしが%d日あります", row.ScheduledButNoActualDays))
	}

	if row.ActualButNoScheduledDays > 0 {
		*warnings = append(*warnings, fmt.Sprintf("予定実績不整合: 実績あり予定なしが%d日あります", row.ActualButNoScheduledDays))
	}

	if row.MissingScheduledWorkDays > 0 {
		*warnings = append(*warnings, fmt.Sprintf("予定不整合: 予定労働時間未設定日が%d日あります", row.MissingScheduledWorkDays))
	}

	if row.ExpenseCount > 0 && row.OtherExpenseAmount == row.ExpenseTotalAmount {
		row.HasExpenseCategoryWarning = true
		*warnings = append(*warnings, "経費カテゴリ警告: Expenseにカテゴリがないため、経費はすべてその他経費として出力しています")
	}
}

/*
 * 経費をCSV行へ反映
 *
 * 現時点の Expense model にはカテゴリカラムが存在しない。
 * そのため、カテゴリ別集計は全額 other_expense_amount に寄せる。
 */
func (service *monthlyAttendanceSummaryExportService) applyExpensesToRow(
	row *types.MonthlyAttendanceSummaryCsvRow,
	expenses []models.Expense,
) {
	for _, expense := range expenses {
		row.ExpenseTotalAmount += expense.Amount
		row.ExpenseCount++
		row.OtherExpenseAmount += expense.Amount
	}
}

type monthlyAttendanceSummaryWorkSegment struct {
	StartAt time.Time
	EndAt   time.Time
}

type monthlyAttendanceSummaryClassifiedSegment struct {
	Minutes     int
	IsLateNight bool
}

/*
 * 有効な勤務区間を生成する
 *
 * 実績時間から有効な休憩時間を差し引き、残った勤務区間を時系列で返す。
 */
func buildWorkSegments(
	actualStartAt time.Time,
	actualEndAt time.Time,
	validBreaks []models.AttendanceBreak,
) []monthlyAttendanceSummaryWorkSegment {
	if !actualEndAt.After(actualStartAt) {
		return []monthlyAttendanceSummaryWorkSegment{}
	}

	sort.Slice(validBreaks, func(i int, j int) bool {
		return validBreaks[i].BreakStartAt.Before(validBreaks[j].BreakStartAt)
	})

	segments := []monthlyAttendanceSummaryWorkSegment{}
	currentStartAt := actualStartAt

	for _, attendanceBreak := range validBreaks {
		if attendanceBreak.BreakStartAt.After(currentStartAt) {
			segments = append(segments, monthlyAttendanceSummaryWorkSegment{
				StartAt: currentStartAt,
				EndAt:   attendanceBreak.BreakStartAt,
			})
		}

		if attendanceBreak.BreakEndAt.After(currentStartAt) {
			currentStartAt = attendanceBreak.BreakEndAt
		}
	}

	if actualEndAt.After(currentStartAt) {
		segments = append(segments, monthlyAttendanceSummaryWorkSegment{
			StartAt: currentStartAt,
			EndAt:   actualEndAt,
		})
	}

	return segments
}

/*
 * 勤務区間を日中/深夜に分解し、日別行へ反映する
 */
func (service *monthlyAttendanceSummaryExportService) applyDayNightWorkMinutes(
	workRow *types.MonthlyAttendanceSummaryWorkRow,
	workSegments []monthlyAttendanceSummaryWorkSegment,
) {
	workRow.WorkMinuteSegments = []types.MonthlyAttendanceSummaryWorkMinuteSegment{}

	for _, segment := range workSegments {
		for _, classifiedSegment := range splitSegmentByLateNight(segment.StartAt, segment.EndAt) {
			if classifiedSegment.Minutes <= 0 {
				continue
			}

			workRow.WorkMinuteSegments = append(workRow.WorkMinuteSegments, types.MonthlyAttendanceSummaryWorkMinuteSegment{
				Minutes:     classifiedSegment.Minutes,
				IsLateNight: classifiedSegment.IsLateNight,
			})

			if classifiedSegment.IsLateNight {
				workRow.NightWorkMinutes += classifiedSegment.Minutes
				workRow.LateNightWorkMinutes += classifiedSegment.Minutes
			} else {
				workRow.DayWorkMinutes += classifiedSegment.Minutes
			}
		}
	}
}

/*
 * 日別残業時間を日中/夜勤に分解する
 *
 * 勤務開始から日別残業基準時間までは所定内、
 * それを超えた部分を日別残業として扱う。
 */
func (service *monthlyAttendanceSummaryExportService) applyDayNightOvertimeMinutes(
	workRow *types.MonthlyAttendanceSummaryWorkRow,
) {
	if workRow.DailyOvertimeMinutes <= 0 {
		return
	}

	workedMinutesBeforeCurrentSegment := 0
	for _, segment := range workRow.WorkMinuteSegments {
		segmentStartWorkedMinutes := workedMinutesBeforeCurrentSegment
		segmentEndWorkedMinutes := workedMinutesBeforeCurrentSegment + segment.Minutes

		overtimeStartInSegment := maxInt(workRow.DailyOvertimeThresholdMinutes, segmentStartWorkedMinutes)
		overtimeEndInSegment := segmentEndWorkedMinutes
		overtimeMinutes := overtimeEndInSegment - overtimeStartInSegment
		if overtimeMinutes < 0 {
			overtimeMinutes = 0
		}

		if segment.IsLateNight {
			workRow.NightOvertimeMinutes += overtimeMinutes
		} else {
			workRow.DayOvertimeMinutes += overtimeMinutes
		}

		workedMinutesBeforeCurrentSegment = segmentEndWorkedMinutes
	}

	allocatedOvertimeMinutes := workRow.DayOvertimeMinutes + workRow.NightOvertimeMinutes
	if allocatedOvertimeMinutes != workRow.DailyOvertimeMinutes {
		gapMinutes := workRow.DailyOvertimeMinutes - allocatedOvertimeMinutes
		if gapMinutes > 0 {
			workRow.DayOvertimeMinutes += gapMinutes
		}
	}
}

/*
 * 1つの勤務区間を日中/深夜に分解する
 */
func splitSegmentByLateNight(
	startAt time.Time,
	endAt time.Time,
) []monthlyAttendanceSummaryClassifiedSegment {
	classifiedSegments := []monthlyAttendanceSummaryClassifiedSegment{}

	if !endAt.After(startAt) {
		return classifiedSegments
	}

	currentAt := startAt
	for currentAt.Before(endAt) {
		nextBoundary := nextLateNightBoundary(currentAt)
		segmentEndAt := minTime(nextBoundary, endAt)
		minutes := minutesBetween(currentAt, segmentEndAt)

		if minutes > 0 {
			classifiedSegments = append(classifiedSegments, monthlyAttendanceSummaryClassifiedSegment{
				Minutes:     minutes,
				IsLateNight: isLateNightTime(currentAt),
			})
		}

		currentAt = segmentEndAt
	}

	return classifiedSegments
}

func nextLateNightBoundary(currentAt time.Time) time.Time {
	currentDate := truncateDate(currentAt)

	boundaries := []time.Time{
		time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), constants.LateNightWorkEndHour, 0, 0, 0, currentAt.Location()),
		time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), constants.LateNightWorkStartHour, 0, 0, 0, currentAt.Location()),
		time.Date(currentDate.AddDate(0, 0, 1).Year(), currentDate.AddDate(0, 0, 1).Month(), currentDate.AddDate(0, 0, 1).Day(), constants.LateNightWorkEndHour, 0, 0, 0, currentAt.Location()),
		time.Date(currentDate.AddDate(0, 0, 1).Year(), currentDate.AddDate(0, 0, 1).Month(), currentDate.AddDate(0, 0, 1).Day(), constants.LateNightWorkStartHour, 0, 0, 0, currentAt.Location()),
	}

	sort.Slice(boundaries, func(i int, j int) bool {
		return boundaries[i].Before(boundaries[j])
	})

	for _, boundary := range boundaries {
		if boundary.After(currentAt) {
			return boundary
		}
	}

	return currentAt.Add(24 * time.Hour)
}

func isLateNightTime(currentAt time.Time) bool {
	hour := currentAt.Hour()
	return hour >= constants.LateNightWorkStartHour || hour < constants.LateNightWorkEndHour
}

func calculateLateMinutes(planStartAt *time.Time, actualStartAt *time.Time) int {
	if planStartAt == nil || actualStartAt == nil {
		return 0
	}

	planStartAtJST := toJST(*planStartAt)
	actualStartAtJST := toJST(*actualStartAt)

	if !actualStartAtJST.After(planStartAtJST) {
		return 0
	}

	return minutesBetween(planStartAtJST, actualStartAtJST)
}

func calculateEarlyLeaveMinutes(planEndAt *time.Time, actualEndAt *time.Time) int {
	if planEndAt == nil || actualEndAt == nil {
		return 0
	}

	planEndAtJST := toJST(*planEndAt)
	actualEndAtJST := toJST(*actualEndAt)

	if !planEndAtJST.After(actualEndAtJST) {
		return 0
	}

	return minutesBetween(actualEndAtJST, planEndAtJST)
}

func roundToOneDecimal(value float64) float64 {
	return math.Round(value*10) / 10
}

func buildWarningText(warnings []string) string {
	return strings.Join(warnings, "; ")
}

/*
 * 勤怠日をユーザーIDでグループ化
 */
func groupAttendanceDaysByUserID(
	attendanceDays []models.AttendanceDay,
) map[uint][]models.AttendanceDay {
	attendanceDaysByUserID := map[uint][]models.AttendanceDay{}

	for _, attendanceDay := range attendanceDays {
		attendanceDaysByUserID[attendanceDay.UserID] = append(attendanceDaysByUserID[attendanceDay.UserID], attendanceDay)
	}

	return attendanceDaysByUserID
}

/*
 * 作業行を週開始日でグループ化
 */
func groupWorkRowsByWeek(
	workRows []*types.MonthlyAttendanceSummaryWorkRow,
) map[string][]*types.MonthlyAttendanceSummaryWorkRow {
	weeks := map[string][]*types.MonthlyAttendanceSummaryWorkRow{}

	for _, workRow := range workRows {
		workDate, err := parseDate(workRow.WorkDate)
		if err != nil {
			continue
		}

		weekStartDate := formatDate(startOfWeek(workDate))
		weeks[weekStartDate] = append(weeks[weekStartDate], workRow)
	}

	return weeks
}

/*
 * 週開始日取得
 */
func startOfWeek(date time.Time) time.Time {
	weekday := int(date.Weekday())
	weekStartDay := int(constants.AttendanceSummaryWeekStartDay)

	diff := weekday - weekStartDay
	if diff < 0 {
		diff += 7
	}

	return truncateDate(date).AddDate(0, 0, -diff)
}

/*
 * 深夜時間との重複分を分で計算する
 */
func calculateLateNightOverlapMinutes(
	startAt time.Time,
	endAt time.Time,
) int {
	if !endAt.After(startAt) {
		return 0
	}

	totalMinutes := 0

	startDate := truncateDate(startAt).AddDate(0, 0, -1)
	endDate := truncateDate(endAt).AddDate(0, 0, 1)

	for currentDate := startDate; !currentDate.After(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
		lateNightStartAt := time.Date(
			currentDate.Year(),
			currentDate.Month(),
			currentDate.Day(),
			constants.LateNightWorkStartHour,
			0,
			0,
			0,
			startAt.Location(),
		)

		lateNightEndAt := time.Date(
			currentDate.AddDate(0, 0, 1).Year(),
			currentDate.AddDate(0, 0, 1).Month(),
			currentDate.AddDate(0, 0, 1).Day(),
			constants.LateNightWorkEndHour,
			0,
			0,
			0,
			startAt.Location(),
		)

		totalMinutes += overlapMinutes(startAt, endAt, lateNightStartAt, lateNightEndAt)
	}

	return totalMinutes
}

/*
 * 2つの期間の重複分を分で返す
 */
func overlapMinutes(
	startA time.Time,
	endA time.Time,
	startB time.Time,
	endB time.Time,
) int {
	start := maxTime(startA, startB)
	end := minTime(endA, endB)

	if !end.After(start) {
		return 0
	}

	return minutesBetween(start, end)
}

/*
 * 夜勤など日を跨ぐ勤務の休憩日時を、実績勤務区間へ正規化する。
 *
 * AttendanceBreak は AttendanceDayID で当日の勤怠に正しく紐づいているため、
 * 時刻部分が 00:30 など勤務開始時刻より小さい場合でも、
 * 実績勤務区間内に入る日付へ最大2日分シフトして判定する。
 *
 * また、23:50〜00:10 のように終了時刻が開始時刻以前の場合は、
 * 終了時刻を翌日として扱う。
 */
func normalizeBreakPeriodToActualWork(
	breakStartAt time.Time,
	breakEndAt time.Time,
	actualStartAt time.Time,
	actualEndAt time.Time,
) (time.Time, time.Time, bool) {
	if !actualEndAt.After(actualStartAt) {
		return breakStartAt, breakEndAt, false
	}

	if !breakEndAt.After(breakStartAt) {
		breakEndAt = breakEndAt.AddDate(0, 0, 1)
	}

	// 保存された日付を基準に、前日〜翌々日まで候補を確認する。
	for dayOffset := -1; dayOffset <= 2; dayOffset++ {
		candidateStartAt := breakStartAt.AddDate(0, 0, dayOffset)
		candidateEndAt := breakEndAt.AddDate(0, 0, dayOffset)

		if candidateStartAt.Before(actualStartAt) {
			continue
		}
		if candidateEndAt.After(actualEndAt) {
			continue
		}
		if !candidateEndAt.After(candidateStartAt) {
			continue
		}

		return candidateStartAt, candidateEndAt, true
	}

	return breakStartAt, breakEndAt, false
}

/*
 * 週残業時間を、その日の「日別残業を除いた勤務区間」の末尾から
 * 日中残業・深夜残業へ配賦する。
 */
func allocateWeeklyOvertimeByDayNight(
	workRow *types.MonthlyAttendanceSummaryWorkRow,
	weeklyOvertimeMinutes int,
) {
	if weeklyOvertimeMinutes <= 0 || len(workRow.WorkMinuteSegments) == 0 {
		return
	}

	baseWorkMinutes := workRow.ActualWorkMinutes - workRow.DailyOvertimeMinutes
	if baseWorkMinutes <= 0 {
		return
	}

	remainingBaseMinutes := baseWorkMinutes
	baseSegments := make([]types.MonthlyAttendanceSummaryWorkMinuteSegment, 0, len(workRow.WorkMinuteSegments))

	for _, segment := range workRow.WorkMinuteSegments {
		if remainingBaseMinutes <= 0 {
			break
		}

		segmentMinutes := minInt(segment.Minutes, remainingBaseMinutes)
		if segmentMinutes > 0 {
			baseSegments = append(baseSegments, types.MonthlyAttendanceSummaryWorkMinuteSegment{
				Minutes:     segmentMinutes,
				IsLateNight: segment.IsLateNight,
			})
			remainingBaseMinutes -= segmentMinutes
		}
	}

	remainingOvertimeMinutes := minInt(weeklyOvertimeMinutes, baseWorkMinutes)
	for index := len(baseSegments) - 1; index >= 0 && remainingOvertimeMinutes > 0; index-- {
		segment := baseSegments[index]
		allocatedMinutes := minInt(segment.Minutes, remainingOvertimeMinutes)

		if segment.IsLateNight {
			workRow.NightOvertimeMinutes += allocatedMinutes
		} else {
			workRow.DayOvertimeMinutes += allocatedMinutes
		}

		remainingOvertimeMinutes -= allocatedMinutes
	}
}

func normalizeMonthlyAttendanceSummaryExportFormat(format string) string {
	normalizedFormat := strings.ToUpper(strings.TrimSpace(format))
	if normalizedFormat == types.MonthlyAttendanceSummaryExportFormatXLSX {
		return types.MonthlyAttendanceSummaryExportFormatXLSX
	}

	return types.MonthlyAttendanceSummaryExportFormatCSV
}

func jstLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Local
	}

	return location
}

func toJST(value time.Time) time.Time {
	return value.In(jstLocation())
}

func minutesBetween(startAt time.Time, endAt time.Time) int {
	return int(endAt.Sub(startAt).Minutes())
}

func truncateDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

func formatDate(date time.Time) string {
	return date.Format("2006-01-02")
}

func parseDate(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", value, jstLocation())
}

func timePtrValue(value *time.Time) string {
	if value == nil {
		return ""
	}

	return toJST(*value).Format("2006-01-02 15:04:05")
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func uintPtrValue(value *uint) uint {
	if value == nil {
		return 0
	}

	return *value
}

func intPtrValue(value *int) int {
	if value == nil {
		return 0
	}

	return *value
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}

	return b
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}

	return b
}

func maxTime(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	}

	return b
}

func minTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}

	return b
}

func countWeekdays(startDate time.Time, endDate time.Time) int {
	count := 0

	for currentDate := startDate; !currentDate.After(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
		if currentDate.Weekday() != time.Saturday && currentDate.Weekday() != time.Sunday {
			count++
		}
	}

	return count
}

func countWarningsByPrefix(warnings []string, prefix string) int {
	count := 0

	for _, warning := range warnings {
		if strings.Contains(warning, prefix) {
			count++
		}
	}

	return count
}

func isHolidayAttendanceType(attendanceType models.AttendanceType) bool {
	return attendanceType.Code == "HOLIDAY" || attendanceType.Category == "HOLIDAY"
}

func isPaidLeaveAttendanceType(attendanceType models.AttendanceType) bool {
	return attendanceType.Code == "PAID_LEAVE"
}
