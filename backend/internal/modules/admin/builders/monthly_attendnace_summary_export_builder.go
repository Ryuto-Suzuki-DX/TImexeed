package builders

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"

	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 月次勤怠集計CSV出力 Builder interface
 *
 * 注意：
 * ・CSVヘッダーとCSVレコードの生成を担当する
 * ・勤怠計算そのものはServiceで行う
 */
type MonthlyAttendanceSummaryExportBuilder interface {
	BuildCSV(rows []types.MonthlyAttendanceSummaryCsvRow) ([]byte, results.Result)
	BuildFileName(targetYear int, targetMonth int) string
}

/*
 * 月次勤怠集計CSV出力 Builder
 */
type monthlyAttendanceSummaryExportBuilder struct {
	db *gorm.DB
}

/*
 * MonthlyAttendanceSummaryExportBuilder生成
 */
func NewMonthlyAttendanceSummaryExportBuilder(db *gorm.DB) MonthlyAttendanceSummaryExportBuilder {
	return &monthlyAttendanceSummaryExportBuilder{
		db: db,
	}
}

/*
 * CSVファイル名生成
 */
func (builder *monthlyAttendanceSummaryExportBuilder) BuildFileName(targetYear int, targetMonth int) string {
	return fmt.Sprintf("monthly_attendance_summary_%04d_%02d.csv", targetYear, targetMonth)
}

/*
 * CSV生成
 *
 * Excelでも文字化けしにくいようにUTF-8 BOMを付与する。
 */
func (builder *monthlyAttendanceSummaryExportBuilder) BuildCSV(
	rows []types.MonthlyAttendanceSummaryCsvRow,
) ([]byte, results.Result) {
	buffer := &bytes.Buffer{}

	// UTF-8 BOM
	buffer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buffer)

	if err := writer.Write(builder.buildHeader()); err != nil {
		return nil, results.BadRequest(
			"BUILD_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_CSV_HEADER_FAILED",
			"月次勤怠集計CSVのヘッダー生成に失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	for _, row := range rows {
		if err := writer.Write(builder.buildRecord(row)); err != nil {
			return nil, results.BadRequest(
				"BUILD_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_CSV_RECORD_FAILED",
				"月次勤怠集計CSVの行生成に失敗しました",
				map[string]any{
					"error": err.Error(),
				},
			)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, results.BadRequest(
			"BUILD_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_CSV_FLUSH_FAILED",
			"月次勤怠集計CSVの書き込みに失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	return buffer.Bytes(), results.OK(
		nil,
		"BUILD_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_CSV_SUCCESS",
		"",
		nil,
	)
}

/*
 * CSVヘッダー生成
 */
func (builder *monthlyAttendanceSummaryExportBuilder) buildHeader() []string {
	return []string{
		"export_target_year",
		"export_target_month",
		"exported_at",
		"export_status",
		"calculation_status",

		"user_id",
		"user_name",
		"user_email",
		"department_id",
		"department_name",
		"role",
		"hire_date",
		"retirement_date",
		"is_retired_in_target_month",

		"monthly_request_id",
		"monthly_status",
		"request_memo",
		"requested_at",
		"approved_by",
		"approved_at",
		"rejected_reason",
		"rejected_at",
		"canceled_reason",
		"canceled_at",

		"user_salary_detail_id",
		"salary_type",
		"base_salary",
		"hourly_wage",
		"daily_wage",
		"extra_allowance_amount",
		"extra_allowance_memo",
		"fixed_deduction_amount",
		"fixed_deduction_memo",
		"is_payroll_target",
		"salary_effective_from",
		"salary_effective_to",

		"calendar_days",
		"registered_attendance_days",
		"scheduled_work_days",
		"actual_work_days",
		"paid_leave_days",
		"half_paid_leave_days",
		"absence_days",
		"sick_leave_days",
		"holiday_work_days",
		"late_days",
		"early_leave_days",

		"company_daily_standard_work_minutes",
		"company_weekly_standard_work_minutes",
		"scheduled_work_minutes",
		"actual_work_minutes",
		"break_minutes",
		"work_shortage_minutes",
		"work_excess_against_scheduled_minutes",
		"daily_overtime_threshold_minutes",
		"weekly_scheduled_work_minutes",
		"weekly_overtime_threshold_minutes",
		"daily_overtime_minutes",
		"weekly_overtime_minutes",
		"overtime_minutes",
		"late_night_work_minutes",
		"holiday_work_minutes",
		"paid_leave_minutes",
		"absence_minutes",
		"late_minutes",
		"early_leave_minutes",

		"daily_transportation_amount",
		"commuter_pass_amount",
		"total_transportation_amount",
		"commuter_pass_from",
		"commuter_pass_to",
		"commuter_pass_method",
		"daily_transportation_count",

		"paid_leave_used_days",
		"paid_leave_used_minutes",

		"expense_total_amount",
		"salary_included_expense_amount",
		"expense_count",
		"transportation_expense_amount",
		"supplies_expense_amount",
		"communication_expense_amount",
		"other_expense_amount",

		"holiday_count",
		"working_day_count",

		"warning_count",
		"warnings",
		"missing_attendance_days",
		"invalid_break_count",
		"invalid_time_count",
		"has_data_warning",
	}
}

/*
 * CSVレコード生成
 */
func (builder *monthlyAttendanceSummaryExportBuilder) buildRecord(row types.MonthlyAttendanceSummaryCsvRow) []string {
	baseColumns := []string{
		strconv.Itoa(row.ExportTargetYear),
		strconv.Itoa(row.ExportTargetMonth),
		row.ExportedAt,
		row.ExportStatus,
		row.CalculationStatus,

		uintToString(row.UserID),
		row.UserName,
		row.UserEmail,
		uintToString(row.DepartmentID),
		row.DepartmentName,
		row.Role,
		row.HireDate,
		row.RetirementDate,
		boolToString(row.IsRetiredInTargetMonth),

		uintToString(row.MonthlyRequestID),
		row.MonthlyStatus,
		row.RequestMemo,
		row.RequestedAt,
		uintToString(row.ApprovedBy),
		row.ApprovedAt,
		row.RejectedReason,
		row.RejectedAt,
		row.CanceledReason,
		row.CanceledAt,
	}

	/*
	 * APPROVED以外、またはERRORの場合は、
	 * 従業員情報・月次ステータス・警告系だけを出し、
	 * 集計系は空欄にする。
	 */
	if row.CalculationStatus != types.MonthlyAttendanceSummaryCalculationStatusCalculated {
		return append(baseColumns, builder.buildBlankCalculatedColumns(row)...)
	}

	return append(baseColumns,
		uintToString(row.UserSalaryDetailID),
		row.SalaryType,
		intToString(row.BaseSalary),
		intToString(row.HourlyWage),
		intToString(row.DailyWage),
		intToString(row.ExtraAllowanceAmount),
		row.ExtraAllowanceMemo,
		intToString(row.FixedDeductionAmount),
		row.FixedDeductionMemo,
		boolToString(row.IsPayrollTarget),
		row.SalaryEffectiveFrom,
		row.SalaryEffectiveTo,

		intToString(row.CalendarDays),
		intToString(row.RegisteredAttendanceDays),
		intToString(row.ScheduledWorkDays),
		intToString(row.ActualWorkDays),
		intToString(row.PaidLeaveDays),
		intToString(row.HalfPaidLeaveDays),
		intToString(row.AbsenceDays),
		intToString(row.SickLeaveDays),
		intToString(row.HolidayWorkDays),
		intToString(row.LateDays),
		intToString(row.EarlyLeaveDays),

		intToString(row.CompanyDailyStandardWorkMinutes),
		intToString(row.CompanyWeeklyStandardWorkMinutes),
		intToString(row.ScheduledWorkMinutes),
		intToString(row.ActualWorkMinutes),
		intToString(row.BreakMinutes),
		intToString(row.WorkShortageMinutes),
		intToString(row.WorkExcessAgainstScheduledMinutes),
		intToString(row.DailyOvertimeThresholdMinutes),
		intToString(row.WeeklyScheduledWorkMinutes),
		intToString(row.WeeklyOvertimeThresholdMinutes),
		intToString(row.DailyOvertimeMinutes),
		intToString(row.WeeklyOvertimeMinutes),
		intToString(row.OvertimeMinutes),
		intToString(row.LateNightWorkMinutes),
		intToString(row.HolidayWorkMinutes),
		intToString(row.PaidLeaveMinutes),
		intToString(row.AbsenceMinutes),
		intToString(row.LateMinutes),
		intToString(row.EarlyLeaveMinutes),

		intToString(row.DailyTransportationAmount),
		intToString(row.CommuterPassAmount),
		intToString(row.TotalTransportationAmount),
		row.CommuterPassFrom,
		row.CommuterPassTo,
		row.CommuterPassMethod,
		intToString(row.DailyTransportationCount),

		floatToString(row.PaidLeaveUsedDays),
		intToString(row.PaidLeaveUsedMinutes),

		intToString(row.ExpenseTotalAmount),
		intToString(row.SalaryIncludedExpenseAmount),
		intToString(row.ExpenseCount),
		intToString(row.TransportationExpenseAmount),
		intToString(row.SuppliesExpenseAmount),
		intToString(row.CommunicationExpenseAmount),
		intToString(row.OtherExpenseAmount),

		intToString(row.HolidayCount),
		intToString(row.WorkingDayCount),

		intToString(row.WarningCount),
		row.Warnings,
		intToString(row.MissingAttendanceDays),
		intToString(row.InvalidBreakCount),
		intToString(row.InvalidTimeCount),
		boolToString(row.HasDataWarning),
	)
}

/*
 * APPROVED以外/ERROR行の集計系空欄生成
 */
func (builder *monthlyAttendanceSummaryExportBuilder) buildBlankCalculatedColumns(
	row types.MonthlyAttendanceSummaryCsvRow,
) []string {
	return []string{
		"", "", "", "", "", "", "", "", "", "", "", "",

		"", "", "", "", "", "", "", "", "", "", "",

		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",

		"", "", "", "", "", "", "",

		"", "",

		"", "", "", "", "", "", "",

		"", "",

		intToString(row.WarningCount),
		row.Warnings,
		intToString(row.MissingAttendanceDays),
		intToString(row.InvalidBreakCount),
		intToString(row.InvalidTimeCount),
		boolToString(row.HasDataWarning),
	}
}

func uintToString(value uint) string {
	if value == 0 {
		return ""
	}

	return strconv.FormatUint(uint64(value), 10)
}

func intToString(value int) string {
	return strconv.Itoa(value)
}

func floatToString(value float64) string {
	return strconv.FormatFloat(value, 'f', 1, 64)
}

func boolToString(value bool) string {
	if value {
		return "true"
	}

	return "false"
}
