package services

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
)

/*
 * 全従業員の日別勤怠明細Excel出力
 *
 * 既存の月次集計CSV/Excelとは別用途。
 * 承認状態に関係なく、対象月時点で入力されている日別データを出力する。
 */
func (service *monthlyAttendanceSummaryExportService) ExportAllUsersDailyAttendanceDetailExcel(
	request types.ExportMonthlyAttendanceSummaryCsvRequest,
) ([]byte, string, results.Result) {
	if request.TargetYear <= 0 {
		return nil, "", results.BadRequest(
			"EXPORT_DAILY_ATTENDANCE_DETAIL_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{"targetYear": request.TargetYear},
		)
	}
	if request.TargetMonth < 1 || request.TargetMonth > 12 {
		return nil, "", results.BadRequest(
			"EXPORT_DAILY_ATTENDANCE_DETAIL_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{"targetMonth": request.TargetMonth},
		)
	}

	request.TargetType = types.MonthlyAttendanceSummaryExportTargetTypeAll
	request.TargetUserID = nil
	request.DepartmentIDs = nil
	request.IncludeUnassignedDepartment = true
	request.IncludeNotApproved = true

	location := jstLocation()
	targetMonthStart := time.Date(request.TargetYear, time.Month(request.TargetMonth), 1, 0, 0, 0, 0, location)
	targetMonthEnd := targetMonthStart.AddDate(0, 1, -1)
	extendedFromDate := startOfWeek(targetMonthStart)
	extendedToDate := startOfWeek(targetMonthEnd).AddDate(0, 0, 6)

	users, usersResult := service.monthlyAttendanceSummaryExportRepository.SearchExportTargetUsers(request)
	if usersResult.Error {
		return nil, "", usersResult
	}

	// 対象月に在籍していた一般ユーザーだけに絞る。
	filteredUsers := users[:0]
	for _, user := range users {
		if user.HireDate.After(targetMonthEnd) {
			continue
		}
		if user.RetirementDate != nil && user.RetirementDate.Before(targetMonthStart) {
			continue
		}
		filteredUsers = append(filteredUsers, user)
	}
	users = filteredUsers

	if len(users) == 0 {
		return nil, "", results.NotFound(
			"DAILY_ATTENDANCE_DETAIL_TARGET_USERS_NOT_FOUND",
			"対象月に在籍する従業員が見つかりません",
			nil,
		)
	}

	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	monthlyRequestMap, monthlyRequestResult :=
		service.monthlyAttendanceSummaryExportRepository.FindMonthlyAttendanceRequests(
			userIDs,
			request.TargetYear,
			request.TargetMonth,
		)
	if monthlyRequestResult.Error {
		return nil, "", monthlyRequestResult
	}

	attendanceDays, attendanceDaysResult :=
		service.monthlyAttendanceSummaryExportRepository.FindAttendanceDays(
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

	breakMap, breakResult :=
		service.monthlyAttendanceSummaryExportRepository.FindAttendanceBreaks(attendanceDayIDs)
	if breakResult.Error {
		return nil, "", breakResult
	}

	transportMap, transportResult :=
		service.monthlyAttendanceSummaryExportRepository.FindAttendanceTransportExpenses(attendanceDayIDs)
	if transportResult.Error {
		return nil, "", transportResult
	}

	attendanceDaysByUserID := groupAttendanceDaysByUserID(attendanceDays)
	userSheets := make([]types.MonthlyAttendanceDailyDetailUserSheet, 0, len(users))

	for _, user := range users {
		userAttendanceDays := attendanceDaysByUserID[user.ID]
		workRows := service.buildWorkRows(userAttendanceDays, breakMap, transportMap)
		service.applyHolidayWorkFlags(workRows)
		service.applyDailyOvertime(workRows)
		service.applyWeeklyOvertime(workRows, targetMonthStart, targetMonthEnd)

		workRowByDate := make(map[string]*types.MonthlyAttendanceSummaryWorkRow, len(workRows))
		for _, workRow := range workRows {
			workRowByDate[workRow.WorkDate] = workRow
		}

		attendanceDayByDate := make(map[string]models.AttendanceDay, len(userAttendanceDays))
		for _, attendanceDay := range userAttendanceDays {
			attendanceDayByDate[formatDate(attendanceDay.WorkDate)] = attendanceDay
		}

		status := types.MonthlyAttendanceSummaryMonthlyStatusNotSubmitted
		if monthlyRequest, exists := monthlyRequestMap[user.ID]; exists {
			status = monthlyRequest.Status
		}

		departmentName := ""
		if user.DepartmentName != nil {
			departmentName = *user.DepartmentName
		}

		sheet := types.MonthlyAttendanceDailyDetailUserSheet{
			UserID:         user.ID,
			UserName:       user.Name,
			UserEmail:      user.Email,
			DepartmentName: departmentName,
			MonthlyStatus:  status,
			Rows:           make([]types.MonthlyAttendanceDailyDetailRow, 0, targetMonthEnd.Day()),
		}

		for currentDate := targetMonthStart; !currentDate.After(targetMonthEnd); currentDate = currentDate.AddDate(0, 0, 1) {
			dateText := formatDate(currentDate)
			attendanceDay, hasAttendanceDay := attendanceDayByDate[dateText]
			workRow := workRowByDate[dateText]

			row := types.MonthlyAttendanceDailyDetailRow{
				WorkDate:   dateText,
				Weekday:    japaneseWeekday(currentDate),
				IsSaturday: currentDate.Weekday() == time.Saturday,
				IsSunday:   currentDate.Weekday() == time.Sunday,
			}

			if hasAttendanceDay {
				row.PlanAttendanceType = attendanceDay.PlanAttendanceType.Name
				if strings.TrimSpace(row.PlanAttendanceType) == "" {
					row.PlanAttendanceType = attendanceDay.PlanAttendanceType.Code
				}
				row.ActualWorkStatus = attendanceDay.ActualWorkStatus
				row.PlanStart = formatAttendanceTime(attendanceDay.PlanStartAt, currentDate)
				row.PlanEnd = formatAttendanceTime(attendanceDay.PlanEndAt, currentDate)
				row.ActualStart = formatAttendanceTime(attendanceDay.ActualStartAt, currentDate)
				row.ActualEnd = formatAttendanceTime(attendanceDay.ActualEndAt, currentDate)
				row.BreakDetails = buildAttendanceBreakDetails(
					breakMap[attendanceDay.ID],
					attendanceDay.ActualStartAt,
					attendanceDay.ActualEndAt,
				)
			}

			if workRow != nil {
				row.ScheduledWorkMinutes = workRow.ScheduledWorkMinutes
				row.BreakMinutes = workRow.BreakMinutes
				row.ActualWorkMinutes = workRow.ActualWorkMinutes
				row.RegularWorkMinutes = workRow.RegularWorkMinutes
				row.OvertimeMinutes = workRow.DayOvertimeMinutes + workRow.NightOvertimeMinutes
				row.LateNightWorkMinutes = workRow.LateNightWorkMinutes
				if workRow.IsHolidayWorkDay {
					row.HolidayWorkMinutes = workRow.ActualWorkMinutes
				}
				row.LateMinutes = workRow.LateMinutes
				row.EarlyLeaveMinutes = workRow.EarlyLeaveMinutes
				row.TransportAmount = workRow.TransportAmount
				row.Warnings = buildWarningText(workRow.Warnings)
			}

			if hasAttendanceDay && attendanceDay.ActualStartAt != nil && attendanceDay.ActualEndAt != nil {
				startAt := toJST(*attendanceDay.ActualStartAt)
				endAt := toJST(*attendanceDay.ActualEndAt)
				if endAt.After(startAt) {
					row.GrossWorkMinutes = minutesBetween(startAt, endAt)
				}
			}

			sheet.Rows = append(sheet.Rows, row)
		}

		userSheets = append(userSheets, sheet)
	}

	sort.SliceStable(userSheets, func(i, j int) bool {
		if userSheets[i].DepartmentName == userSheets[j].DepartmentName {
			return userSheets[i].UserID < userSheets[j].UserID
		}
		return userSheets[i].DepartmentName < userSheets[j].DepartmentName
	})

	fileBytes, buildResult := service.monthlyAttendanceSummaryExportBuilder.BuildDailyDetailExcel(
		userSheets,
		request.TargetYear,
		request.TargetMonth,
	)
	if buildResult.Error {
		return nil, "", buildResult
	}

	fileName := service.monthlyAttendanceSummaryExportBuilder.BuildDailyDetailExcelFileName(
		request.TargetYear,
		request.TargetMonth,
	)

	return fileBytes, fileName, results.OK(
		types.ExportMonthlyAttendanceSummaryCsvResponse{
			FileName:    fileName,
			TargetYear:  request.TargetYear,
			TargetMonth: request.TargetMonth,
			RowCount:    len(userSheets),
		},
		"EXPORT_ALL_USERS_DAILY_ATTENDANCE_DETAIL_EXCEL_SUCCESS",
		"全従業員の日別勤怠明細Excelを出力しました",
		nil,
	)
}

func japaneseWeekday(value time.Time) string {
	weekdays := []string{"日", "月", "火", "水", "木", "金", "土"}
	return weekdays[int(value.Weekday())]
}

func formatAttendanceTime(value *time.Time, baseDate time.Time) string {
	if value == nil {
		return ""
	}
	target := toJST(*value)
	prefix := ""
	if target.Year() != baseDate.Year() || target.YearDay() != baseDate.YearDay() {
		prefix = "翌"
	}
	return prefix + target.Format("15:04")
}

func buildAttendanceBreakDetails(
	breaks []models.AttendanceBreak,
	actualStartAt *time.Time,
	actualEndAt *time.Time,
) string {
	if len(breaks) == 0 {
		return ""
	}

	values := make([]string, 0, len(breaks))
	for _, attendanceBreak := range breaks {
		if attendanceBreak.IsDeleted {
			continue
		}

		startAt := toJST(attendanceBreak.BreakStartAt)
		endAt := toJST(attendanceBreak.BreakEndAt)

		if actualStartAt != nil && actualEndAt != nil {
			normalizedStartAt, normalizedEndAt, normalized := normalizeBreakPeriodToActualWork(
				startAt,
				endAt,
				toJST(*actualStartAt),
				toJST(*actualEndAt),
			)
			if normalized {
				startAt = normalizedStartAt
				endAt = normalizedEndAt
			}
		}

		values = append(values, fmt.Sprintf("%s～%s", startAt.Format("15:04"), endAt.Format("15:04")))
	}

	return strings.Join(values, " / ")
}
