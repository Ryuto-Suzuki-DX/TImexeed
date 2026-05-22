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
 *
 * 給与計算担当者がそのまま見られるように、日本語ヘッダーで出力する。
 */
func (builder *monthlyAttendanceSummaryExportBuilder) buildHeader() []string {
	return []string{
		"対象年",
		"対象月",
		"出力日時",
		"出力状態",
		"集計状態",

		"従業員ID",
		"従業員名",
		"メールアドレス",
		"部署ID",
		"部署名",
		"権限",
		"入社日",
		"退職日",
		"対象月退職済み",

		"月次申請ID",
		"月次申請状態",
		"申請メモ",
		"申請日時",
		"承認者ID",
		"承認日時",
		"否認理由",
		"否認日時",
		"取下理由",
		"取下日時",

		"給与詳細ID",
		"給与区分",
		"月給",
		"時給",
		"日給",
		"追加手当",
		"追加手当メモ",
		"固定控除",
		"固定控除メモ",
		"給与計算対象",
		"給与設定適用開始日",
		"給与設定適用終了日",

		"暦日数",
		"勤怠登録日数",
		"勤怠未登録日数",
		"予定出勤日数",
		"実出勤日数",
		"日勤出勤日数",
		"夜勤出勤日数",
		"有給日数",
		"半日有給回数",
		"欠勤日数",
		"病欠日数",
		"休日出勤日数",
		"遅刻回数",
		"早退回数",
		"予定あり実績なし日数",
		"実績あり予定なし日数",
		"予定労働時間未設定日数",
		"平日数",
		"祝日数",

		"会社1日標準労働時間_分",
		"会社週標準労働時間_分",
		"予定労働時間_分",
		"総労働時間_分",
		"日中労働時間_分",
		"夜勤労働時間_分",
		"休憩時間_分",
		"所定内労働時間_分",
		"控除対象不足時間_分",
		"予定超過時間_分",
		"日別残業基準時間_分",
		"週予定労働時間_分",
		"週残業基準時間_分",
		"日別残業時間_分",
		"週残業時間_分",
		"総残業時間_分",
		"日中残業時間_分",
		"夜勤残業時間_分",
		"深夜労働時間_分",
		"休日労働時間_分",
		"休日深夜労働時間_分",
		"有給換算時間_分",
		"欠勤控除時間_分",
		"病欠控除時間_分",
		"遅刻控除時間_分",
		"早退控除時間_分",

		"実労働稼働率_％",
		"給与対象稼働率_％",
		"稼働率判定",

		"日別交通費合計",
		"月次定期代",
		"交通費合計",
		"通勤区間From",
		"通勤区間To",
		"通勤方法",
		"日別交通費登録回数",

		"有給使用日数",
		"有給使用換算時間_分",

		"経費合計",
		"給与含め経費合計",
		"経費件数",
		"交通費系経費",
		"備品系経費",
		"通信費系経費",
		"その他経費",

		"警告件数",
		"警告内容",
		"休憩不整合件数",
		"時刻不整合件数",
		"データ警告あり",
		"給与設定警告あり",
		"給与計算対象外警告あり",
		"月次承認警告あり",
		"勤怠未登録警告あり",
		"予定実績不整合警告あり",
		"経費カテゴリ警告あり",
	}
}

/*
 * CSVレコード生成
 */
func (builder *monthlyAttendanceSummaryExportBuilder) buildRecord(row types.MonthlyAttendanceSummaryCsvRow) []string {
	calculated := row.CalculationStatus == types.MonthlyAttendanceSummaryCalculationStatusCalculated

	return []string{
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

		calcUintToString(calculated, row.UserSalaryDetailID),
		calcString(calculated, row.SalaryType),
		calcIntToString(calculated, row.BaseSalary),
		calcIntToString(calculated, row.HourlyWage),
		calcIntToString(calculated, row.DailyWage),
		calcIntToString(calculated, row.ExtraAllowanceAmount),
		calcString(calculated, row.ExtraAllowanceMemo),
		calcIntToString(calculated, row.FixedDeductionAmount),
		calcString(calculated, row.FixedDeductionMemo),
		calcBoolToString(calculated, row.IsPayrollTarget),
		calcString(calculated, row.SalaryEffectiveFrom),
		calcString(calculated, row.SalaryEffectiveTo),

		calcIntToString(calculated, row.CalendarDays),
		calcIntToString(calculated, row.RegisteredAttendanceDays),
		calcIntToString(calculated, row.MissingAttendanceDays),
		calcIntToString(calculated, row.ScheduledWorkDays),
		calcIntToString(calculated, row.ActualWorkDays),
		calcIntToString(calculated, row.DayShiftWorkDays),
		calcIntToString(calculated, row.NightShiftWorkDays),
		calcIntToString(calculated, row.PaidLeaveDays),
		calcIntToString(calculated, row.HalfPaidLeaveDays),
		calcIntToString(calculated, row.AbsenceDays),
		calcIntToString(calculated, row.SickLeaveDays),
		calcIntToString(calculated, row.HolidayWorkDays),
		calcIntToString(calculated, row.LateDays),
		calcIntToString(calculated, row.EarlyLeaveDays),
		calcIntToString(calculated, row.ScheduledButNoActualDays),
		calcIntToString(calculated, row.ActualButNoScheduledDays),
		calcIntToString(calculated, row.MissingScheduledWorkDays),
		calcIntToString(calculated, row.WorkingDayCount),
		calcIntToString(calculated, row.HolidayCount),

		calcIntToString(calculated, row.CompanyDailyStandardWorkMinutes),
		calcIntToString(calculated, row.CompanyWeeklyStandardWorkMinutes),
		calcIntToString(calculated, row.ScheduledWorkMinutes),
		calcIntToString(calculated, row.ActualWorkMinutes),
		calcIntToString(calculated, row.DayWorkMinutes),
		calcIntToString(calculated, row.NightWorkMinutes),
		calcIntToString(calculated, row.BreakMinutes),
		calcIntToString(calculated, row.RegularWorkMinutes),
		calcIntToString(calculated, row.WorkShortageMinutes),
		calcIntToString(calculated, row.WorkExcessAgainstScheduledMinutes),
		calcIntToString(calculated, row.DailyOvertimeThresholdMinutes),
		calcIntToString(calculated, row.WeeklyScheduledWorkMinutes),
		calcIntToString(calculated, row.WeeklyOvertimeThresholdMinutes),
		calcIntToString(calculated, row.DailyOvertimeMinutes),
		calcIntToString(calculated, row.WeeklyOvertimeMinutes),
		calcIntToString(calculated, row.OvertimeMinutes),
		calcIntToString(calculated, row.DayOvertimeMinutes),
		calcIntToString(calculated, row.NightOvertimeMinutes),
		calcIntToString(calculated, row.LateNightWorkMinutes),
		calcIntToString(calculated, row.HolidayWorkMinutes),
		calcIntToString(calculated, row.HolidayLateNightWorkMinutes),
		calcIntToString(calculated, row.PaidLeaveMinutes),
		calcIntToString(calculated, row.AbsenceMinutes),
		calcIntToString(calculated, row.SickLeaveMinutes),
		calcIntToString(calculated, row.LateMinutes),
		calcIntToString(calculated, row.EarlyLeaveMinutes),

		calcFloatToString(calculated, row.ActualOperationRate),
		calcFloatToString(calculated, row.PayrollTargetOperationRate),
		calcString(calculated, row.OperationRateJudge),

		calcIntToString(calculated, row.DailyTransportationAmount),
		calcIntToString(calculated, row.CommuterPassAmount),
		calcIntToString(calculated, row.TotalTransportationAmount),
		calcString(calculated, row.CommuterPassFrom),
		calcString(calculated, row.CommuterPassTo),
		calcString(calculated, row.CommuterPassMethod),
		calcIntToString(calculated, row.DailyTransportationCount),

		calcFloatToString(calculated, row.PaidLeaveUsedDays),
		calcIntToString(calculated, row.PaidLeaveUsedMinutes),

		calcIntToString(calculated, row.ExpenseTotalAmount),
		calcIntToString(calculated, row.SalaryIncludedExpenseAmount),
		calcIntToString(calculated, row.ExpenseCount),
		calcIntToString(calculated, row.TransportationExpenseAmount),
		calcIntToString(calculated, row.SuppliesExpenseAmount),
		calcIntToString(calculated, row.CommunicationExpenseAmount),
		calcIntToString(calculated, row.OtherExpenseAmount),

		intToString(row.WarningCount),
		row.Warnings,
		intToString(row.InvalidBreakCount),
		intToString(row.InvalidTimeCount),
		boolToString(row.HasDataWarning),
		boolToString(row.HasSalarySettingWarning),
		boolToString(row.HasPayrollExcludedWarning),
		boolToString(row.HasMonthlyApprovalWarning),
		boolToString(row.HasAttendanceMissingWarning),
		boolToString(row.HasScheduleActualMismatchWarning),
		boolToString(row.HasExpenseCategoryWarning),
	}
}

func calcString(calculated bool, value string) string {
	if !calculated {
		return ""
	}

	return value
}

func calcUintToString(calculated bool, value uint) string {
	if !calculated {
		return ""
	}

	return uintToString(value)
}

func calcIntToString(calculated bool, value int) string {
	if !calculated {
		return ""
	}

	return intToString(value)
}

func calcFloatToString(calculated bool, value float64) string {
	if !calculated {
		return ""
	}

	return floatToString(value)
}

func calcBoolToString(calculated bool, value bool) string {
	if !calculated {
		return ""
	}

	return boolToString(value)
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
