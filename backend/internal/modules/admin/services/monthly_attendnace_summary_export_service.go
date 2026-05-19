package services

import (
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
	ExportMonthlyAttendanceSummaryCsv(request types.ExportMonthlyAttendanceSummaryCsvRequest) ([]byte, string, results.Result)
}

/*
 * 月次勤怠集計CSV出力 Service
 *
 * 注意：
 * ・給与計算そのものは行わない
 * ・APPROVED の月だけ集計値をCSVへ出力する
 * ・APPROVED 以外はステータスのみCSVへ出力する
 * ・残業、週残業、休日出勤、深夜労働は二重計上しない
 * ・変形労働制フラグは持たず、AttendanceDay.ScheduledWorkMinutes の値だけで判断する
 * ・予定区分は AttendanceDay.PlanAttendanceType を見る
 * ・実績状態は AttendanceDay.ActualWorkStatus を見る
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
 * 月次勤怠集計CSV出力
 */
func (service *monthlyAttendanceSummaryExportService) ExportMonthlyAttendanceSummaryCsv(
	request types.ExportMonthlyAttendanceSummaryCsvRequest,
) ([]byte, string, results.Result) {
	if request.TargetYear <= 0 {
		return nil, "", results.BadRequest(
			"EXPORT_MONTHLY_ATTENDANCE_SUMMARY_CSV_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{
				"targetYear": request.TargetYear,
			},
		)
	}

	if request.TargetMonth < 1 || request.TargetMonth > 12 {
		return nil, "", results.BadRequest(
			"EXPORT_MONTHLY_ATTENDANCE_SUMMARY_CSV_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{
				"targetMonth": request.TargetMonth,
			},
		)
	}

	targetMonthStart := time.Date(request.TargetYear, time.Month(request.TargetMonth), 1, 0, 0, 0, 0, time.Local)
	targetMonthEnd := targetMonthStart.AddDate(0, 1, -1)

	/*
	 * 週残業・休日出勤は月をまたぐ週も見る必要があるため、
	 * 対象月初を含む週の月曜から、対象月末を含む週の日曜まで取得する。
	 */
	extendedFromDate := startOfWeek(targetMonthStart)
	extendedToDate := startOfWeek(targetMonthEnd).AddDate(0, 0, 6)

	exportedAt := time.Now().Format("2006-01-02 15:04:05")

	users, usersResult := service.monthlyAttendanceSummaryExportRepository.SearchExportTargetUsers(request)
	if usersResult.Error {
		return nil, "", usersResult
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
		return nil, "", monthlyRequestResult
	}

	attendanceDays, attendanceDaysResult := service.monthlyAttendanceSummaryExportRepository.FindAttendanceDays(
		userIDs,
		extendedFromDate,
		extendedToDate,
	)
	if attendanceDaysResult.Error {
		return nil, "", attendanceDaysResult
	}

	attendanceDayIDs := make([]uint, 0, len(attendanceDays))
	for _, attendanceDay := range attendanceDays {
		attendanceDayIDs = append(attendanceDayIDs, attendanceDay.ID)
	}

	attendanceBreakMap, attendanceBreakResult := service.monthlyAttendanceSummaryExportRepository.FindAttendanceBreaks(attendanceDayIDs)
	if attendanceBreakResult.Error {
		return nil, "", attendanceBreakResult
	}

	monthlyCommuterPassMap, commuterPassResult := service.monthlyAttendanceSummaryExportRepository.FindMonthlyCommuterPasses(
		userIDs,
		request.TargetYear,
		request.TargetMonth,
	)
	if commuterPassResult.Error {
		return nil, "", commuterPassResult
	}

	userSalaryDetailMap, userSalaryDetailResult := service.monthlyAttendanceSummaryExportRepository.FindUserSalaryDetails(
		userIDs,
		targetMonthStart,
		targetMonthEnd,
	)
	if userSalaryDetailResult.Error {
		return nil, "", userSalaryDetailResult
	}

	paidLeaveUsageMap, paidLeaveUsageResult := service.monthlyAttendanceSummaryExportRepository.FindPaidLeaveUsages(
		userIDs,
		targetMonthStart,
		targetMonthEnd,
	)
	if paidLeaveUsageResult.Error {
		return nil, "", paidLeaveUsageResult
	}

	expenseMap, expenseResult := service.monthlyAttendanceSummaryExportRepository.FindExpenses(
		userIDs,
		targetMonthStart,
	)
	if expenseResult.Error {
		return nil, "", expenseResult
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
				rows = append(rows, row)
			}

			continue
		}

		calculatedRow := service.calculateApprovedUserRow(
			row,
			attendanceDaysByUserID[user.ID],
			attendanceBreakMap,
			monthlyCommuterPassMap[user.ID],
			userSalaryDetailMap[user.ID],
			paidLeaveUsageMap[user.ID],
			expenseMap[user.ID],
			targetMonthStart,
			targetMonthEnd,
		)

		rows = append(rows, calculatedRow)
	}

	csvBytes, csvResult := service.monthlyAttendanceSummaryExportBuilder.BuildCSV(rows)
	if csvResult.Error {
		return nil, "", csvResult
	}

	fileName := service.monthlyAttendanceSummaryExportBuilder.BuildFileName(request.TargetYear, request.TargetMonth)

	return csvBytes, fileName, results.OK(
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
	monthlyCommuterPass models.MonthlyCommuterPass,
	userSalaryDetail models.UserSalaryDetail,
	paidLeaveUsages []models.PaidLeaveUsage,
	expenses []models.Expense,
	targetMonthStart time.Time,
	targetMonthEnd time.Time,
) types.MonthlyAttendanceSummaryCsvRow {
	row.CalculationStatus = types.MonthlyAttendanceSummaryCalculationStatusCalculated
	row.CalendarDays = targetMonthEnd.Day()
	row.WorkingDayCount = countWeekdays(targetMonthStart, targetMonthEnd)

	workRows := service.buildWorkRows(attendanceDays, attendanceBreakMap)
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
		row.BreakMinutes += workRow.BreakMinutes
		row.LateNightWorkMinutes += workRow.LateNightWorkMinutes

		if workRow.IsActualWorkDay {
			row.ActualWorkDays++
		}

		if workRow.IsHolidayWorkDay {
			row.HolidayWorkDays++
			row.HolidayWorkMinutes += workRow.ActualWorkMinutes
		}

		if workRow.IsPaidLeaveDay {
			row.PaidLeaveDays++
		}

		if workRow.IsHalfPaidLeaveDay {
			row.HalfPaidLeaveDays++
		}

		if workRow.IsAbsenceDay {
			row.AbsenceDays++
			row.AbsenceMinutes += workRow.ScheduledWorkMinutes
		}

		if workRow.IsSickLeaveDay {
			row.SickLeaveDays++
		}

		if workRow.ActualWorkStatus == constants.ActualWorkStatusLate {
			row.LateDays++
		}

		if workRow.ActualWorkStatus == constants.ActualWorkStatusEarlyLeave {
			row.EarlyLeaveDays++
		}

		row.DailyOvertimeThresholdMinutes += workRow.DailyOvertimeThresholdMinutes
		row.DailyOvertimeMinutes += workRow.DailyOvertimeMinutes
		row.WorkShortageMinutes += workRow.WorkShortageMinutes
		row.WorkExcessAgainstScheduledMinutes += workRow.WorkExcessAgainstScheduledMinutes

		row.DailyTransportationAmount += workRow.TransportAmount
		if workRow.TransportAmount > 0 {
			row.DailyTransportationCount++
		}

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

	row.RegisteredAttendanceDays = attendanceDaysInMonth
	row.MissingAttendanceDays = row.CalendarDays - len(dateWithAttendance)
	if row.MissingAttendanceDays < 0 {
		row.MissingAttendanceDays = 0
	}

	service.applyCommuterPassToRow(&row, monthlyCommuterPass)
	service.applyUserSalaryDetailToRow(&row, userSalaryDetail)

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

	row.InvalidBreakCount = countWarningsByPrefix(warnings, "休憩")
	row.InvalidTimeCount = countWarningsByPrefix(warnings, "時刻")
	row.WarningCount = len(warnings)
	row.Warnings = strings.Join(warnings, "; ")
	row.HasDataWarning = row.WarningCount > 0

	return row
}

/*
 * 日別作業行生成
 */
func (service *monthlyAttendanceSummaryExportService) buildWorkRows(
	attendanceDays []models.AttendanceDay,
	attendanceBreakMap map[uint][]models.AttendanceBreak,
) []*types.MonthlyAttendanceSummaryWorkRow {
	workRows := make([]*types.MonthlyAttendanceSummaryWorkRow, 0, len(attendanceDays))

	for _, attendanceDay := range attendanceDays {
		workDate := formatDate(attendanceDay.WorkDate)
		scheduledWorkMinutes := intPtrValue(attendanceDay.ScheduledWorkMinutes)

		workRow := &types.MonthlyAttendanceSummaryWorkRow{
			UserID:                     attendanceDay.UserID,
			WorkDate:                   workDate,
			AttendanceDayID:            attendanceDay.ID,
			PlanAttendanceTypeID:       attendanceDay.PlanAttendanceTypeID,
			PlanAttendanceTypeCode:     attendanceDay.PlanAttendanceType.Code,
			PlanAttendanceTypeCategory: attendanceDay.PlanAttendanceType.Category,
			ActualWorkStatus:           attendanceDay.ActualWorkStatus,
			ScheduledWorkMinutes:       scheduledWorkMinutes,
			TransportAmount:            intPtrValue(attendanceDay.TransportAmount),
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

		if attendanceDay.ActualStartAt == nil && attendanceDay.ActualEndAt == nil {
			workRows = append(workRows, workRow)
			continue
		}

		if attendanceDay.ActualStartAt == nil || attendanceDay.ActualEndAt == nil {
			workRow.Warnings = append(workRow.Warnings, workDate+" 時刻不整合: 実績開始または実績終了が未入力です")
			workRows = append(workRows, workRow)
			continue
		}

		if !attendanceDay.ActualEndAt.After(*attendanceDay.ActualStartAt) {
			workRow.Warnings = append(workRow.Warnings, workDate+" 時刻不整合: 実績終了が実績開始以前です")
			workRows = append(workRows, workRow)
			continue
		}

		actualStartAt := *attendanceDay.ActualStartAt
		actualEndAt := *attendanceDay.ActualEndAt

		validBreakMinutes := 0
		lateNightBreakMinutes := 0

		for _, attendanceBreak := range attendanceBreakMap[attendanceDay.ID] {
			if !attendanceBreak.BreakEndAt.After(attendanceBreak.BreakStartAt) {
				workRow.Warnings = append(workRow.Warnings, workDate+" 休憩不整合: 休憩終了が休憩開始以前です")
				continue
			}

			if attendanceBreak.BreakStartAt.Before(actualStartAt) || attendanceBreak.BreakEndAt.After(actualEndAt) {
				workRow.Warnings = append(workRow.Warnings, workDate+" 休憩不整合: 休憩が実績勤務時間外です")
				continue
			}

			validBreakMinutes += minutesBetween(attendanceBreak.BreakStartAt, attendanceBreak.BreakEndAt)
			lateNightBreakMinutes += calculateLateNightOverlapMinutes(attendanceBreak.BreakStartAt, attendanceBreak.BreakEndAt)
		}

		grossActualMinutes := minutesBetween(actualStartAt, actualEndAt)
		actualWorkMinutes := grossActualMinutes - validBreakMinutes
		if actualWorkMinutes < 0 {
			actualWorkMinutes = 0
			workRow.Warnings = append(workRow.Warnings, workDate+" 時刻不整合: 休憩時間が実績勤務時間を超えています")
		}

		lateNightWorkMinutes := calculateLateNightOverlapMinutes(actualStartAt, actualEndAt) - lateNightBreakMinutes
		if lateNightWorkMinutes < 0 {
			lateNightWorkMinutes = 0
		}

		workRow.BreakMinutes = validBreakMinutes
		workRow.ActualWorkMinutes = actualWorkMinutes
		workRow.LateNightWorkMinutes = lateNightWorkMinutes

		/*
		 * 実勤務日判定
		 *
		 * 実績状態が欠勤・病欠の場合は実勤務日にしない。
		 * 遅刻・早退は実勤務ありとして扱う。
		 */
		workRow.IsActualWorkDay = actualWorkMinutes > 0 &&
			attendanceDay.ActualWorkStatus != constants.ActualWorkStatusAbsence &&
			attendanceDay.ActualWorkStatus != constants.ActualWorkStatusSickLeave

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

		if workRow.IsHolidayWorkDay {
			workRow.DailyOvertimeMinutes = 0
			continue
		}

		workRow.DailyOvertimeMinutes = maxInt(
			workRow.ActualWorkMinutes-workRow.DailyOvertimeThresholdMinutes,
			0,
		)

		workRow.WorkShortageMinutes = maxInt(
			workRow.ScheduledWorkMinutes-workRow.ActualWorkMinutes,
			0,
		)

		workRow.WorkExcessAgainstScheduledMinutes = maxInt(
			workRow.ActualWorkMinutes-workRow.ScheduledWorkMinutes,
			0,
		)
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
 * ユーザー給与詳細をCSV行へ反映
 */
func (service *monthlyAttendanceSummaryExportService) applyUserSalaryDetailToRow(
	row *types.MonthlyAttendanceSummaryCsvRow,
	userSalaryDetail models.UserSalaryDetail,
) {
	row.UserSalaryDetailID = userSalaryDetail.ID
	row.SalaryType = userSalaryDetail.SalaryType
	row.ExtraAllowanceAmount = userSalaryDetail.ExtraAllowanceAmount
	row.ExtraAllowanceMemo = userSalaryDetail.ExtraAllowanceMemo
	row.FixedDeductionAmount = userSalaryDetail.FixedDeductionAmount
	row.FixedDeductionMemo = userSalaryDetail.FixedDeductionMemo
	row.IsPayrollTarget = userSalaryDetail.IsPayrollTarget
	row.SalaryEffectiveFrom = formatDate(userSalaryDetail.EffectiveFrom)

	if userSalaryDetail.EffectiveTo != nil {
		row.SalaryEffectiveTo = formatDate(*userSalaryDetail.EffectiveTo)
	}

	switch userSalaryDetail.SalaryType {
	case "MONTHLY":
		row.BaseSalary = userSalaryDetail.BaseAmount
	case "HOURLY":
		row.HourlyWage = userSalaryDetail.BaseAmount
	case "DAILY":
		row.DailyWage = userSalaryDetail.BaseAmount
	default:
		row.BaseSalary = userSalaryDetail.BaseAmount
	}
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
		row.SalaryIncludedExpenseAmount += expense.Amount
		row.ExpenseCount++
		row.OtherExpenseAmount += expense.Amount
	}
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
	return time.ParseInLocation("2006-01-02", value, time.Local)
}

func timePtrValue(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.Format("2006-01-02 15:04:05")
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
