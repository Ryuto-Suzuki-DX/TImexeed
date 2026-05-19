package types

/*
 * 月次勤怠集計CSV出力 Type
 *
 * 管理者専用。
 *
 * 注意：
 * ・給与計算そのものは行わない
 * ・APPROVED の月だけ集計値を出力する
 * ・APPROVED 以外の月はステータスのみ出力する
 * ・変形労働制フラグは持たず、AttendanceDay.ScheduledWorkMinutes の値だけで判断する
 * ・予定区分は attendance_types を参照する
 * ・実績状態は constants/attendance_status_constants.go の固定値を使う
 * ・ActualAttendanceTypeID / ActualAttendanceTypeCode / ActualAttendanceTypeIsWorked は使わない
 */

/*
 * 月次勤怠集計CSV出力 Request
 *
 * TargetUserIDs:
 *   指定がある場合、そのユーザーだけを対象にする。
 *   空の場合は、検索条件に一致するユーザーを対象にする。
 *
 * DepartmentID:
 *   指定がある場合、その所属のユーザーだけを対象にする。
 *
 * Keyword:
 *   指定がある場合、ユーザー名/メールアドレスなどのフリーワード検索に使う。
 *
 * IncludeNotApproved:
 *   true の場合、APPROVED 以外のユーザーもステータスのみCSVへ出力する。
 *   false の場合、APPROVED のユーザーだけ出力する。
 */
type ExportMonthlyAttendanceSummaryCsvRequest struct {
	TargetYear         int    `json:"targetYear"`
	TargetMonth        int    `json:"targetMonth"`
	TargetUserIDs      []uint `json:"targetUserIds"`
	DepartmentID       *uint  `json:"departmentId"`
	Keyword            string `json:"keyword"`
	IncludeNotApproved bool   `json:"includeNotApproved"`
}

/*
 * 月次勤怠集計CSV出力 Response
 *
 * CSVは controller で c.Data として返すため、
 * 正常系のJSONレスポンスでは基本的に使わない。
 *
 * 将来的にプレビューAPIや出力履歴APIを作る場合に備えて残す。
 */
type ExportMonthlyAttendanceSummaryCsvResponse struct {
	FileName    string `json:"fileName"`
	TargetYear  int    `json:"targetYear"`
	TargetMonth int    `json:"targetMonth"`
	RowCount    int    `json:"rowCount"`
}

/*
 * 月次勤怠集計CSV 1行分
 *
 * 基本は「対象年月 × ユーザー」で1行。
 *
 * APPROVED 以外の場合：
 *   ユーザー情報、対象年月、月次申請ステータス、calculationStatus のみセットし、
 *   集計系の値はCSV生成時に空欄で出力する。
 */
type MonthlyAttendanceSummaryCsvRow struct {
	// CSV管理情報
	ExportTargetYear  int    `json:"exportTargetYear"`
	ExportTargetMonth int    `json:"exportTargetMonth"`
	ExportedAt        string `json:"exportedAt"`
	ExportStatus      string `json:"exportStatus"`
	CalculationStatus string `json:"calculationStatus"`

	// 従業員情報
	UserID                 uint   `json:"userId"`
	UserName               string `json:"userName"`
	UserEmail              string `json:"userEmail"`
	DepartmentID           uint   `json:"departmentId"`
	DepartmentName         string `json:"departmentName"`
	Role                   string `json:"role"`
	HireDate               string `json:"hireDate"`
	RetirementDate         string `json:"retirementDate"`
	IsRetiredInTargetMonth bool   `json:"isRetiredInTargetMonth"`

	// 月次申請・承認情報
	MonthlyRequestID uint   `json:"monthlyRequestId"`
	MonthlyStatus    string `json:"monthlyStatus"`
	RequestMemo      string `json:"requestMemo"`
	RequestedAt      string `json:"requestedAt"`
	ApprovedBy       uint   `json:"approvedBy"`
	ApprovedAt       string `json:"approvedAt"`
	RejectedReason   string `json:"rejectedReason"`
	RejectedAt       string `json:"rejectedAt"`
	CanceledReason   string `json:"canceledReason"`
	CanceledAt       string `json:"canceledAt"`

	// 給与設定情報
	UserSalaryDetailID   uint   `json:"userSalaryDetailId"`
	SalaryType           string `json:"salaryType"`
	BaseSalary           int    `json:"baseSalary"`
	HourlyWage           int    `json:"hourlyWage"`
	DailyWage            int    `json:"dailyWage"`
	ExtraAllowanceAmount int    `json:"extraAllowanceAmount"`
	ExtraAllowanceMemo   string `json:"extraAllowanceMemo"`
	FixedDeductionAmount int    `json:"fixedDeductionAmount"`
	FixedDeductionMemo   string `json:"fixedDeductionMemo"`
	IsPayrollTarget      bool   `json:"isPayrollTarget"`
	SalaryEffectiveFrom  string `json:"salaryEffectiveFrom"`
	SalaryEffectiveTo    string `json:"salaryEffectiveTo"`

	// 勤怠日数集計
	CalendarDays             int `json:"calendarDays"`
	RegisteredAttendanceDays int `json:"registeredAttendanceDays"`
	ScheduledWorkDays        int `json:"scheduledWorkDays"`
	ActualWorkDays           int `json:"actualWorkDays"`
	PaidLeaveDays            int `json:"paidLeaveDays"`
	HalfPaidLeaveDays        int `json:"halfPaidLeaveDays"`
	AbsenceDays              int `json:"absenceDays"`
	SickLeaveDays            int `json:"sickLeaveDays"`
	HolidayWorkDays          int `json:"holidayWorkDays"`
	LateDays                 int `json:"lateDays"`
	EarlyLeaveDays           int `json:"earlyLeaveDays"`

	// 勤怠時間集計
	CompanyDailyStandardWorkMinutes  int `json:"companyDailyStandardWorkMinutes"`
	CompanyWeeklyStandardWorkMinutes int `json:"companyWeeklyStandardWorkMinutes"`

	ScheduledWorkMinutes              int `json:"scheduledWorkMinutes"`
	ActualWorkMinutes                 int `json:"actualWorkMinutes"`
	BreakMinutes                      int `json:"breakMinutes"`
	WorkShortageMinutes               int `json:"workShortageMinutes"`
	WorkExcessAgainstScheduledMinutes int `json:"workExcessAgainstScheduledMinutes"`

	DailyOvertimeThresholdMinutes  int `json:"dailyOvertimeThresholdMinutes"`
	WeeklyScheduledWorkMinutes     int `json:"weeklyScheduledWorkMinutes"`
	WeeklyOvertimeThresholdMinutes int `json:"weeklyOvertimeThresholdMinutes"`
	DailyOvertimeMinutes           int `json:"dailyOvertimeMinutes"`
	WeeklyOvertimeMinutes          int `json:"weeklyOvertimeMinutes"`
	OvertimeMinutes                int `json:"overtimeMinutes"`

	LateNightWorkMinutes int `json:"lateNightWorkMinutes"`
	HolidayWorkMinutes   int `json:"holidayWorkMinutes"`
	PaidLeaveMinutes     int `json:"paidLeaveMinutes"`
	AbsenceMinutes       int `json:"absenceMinutes"`
	LateMinutes          int `json:"lateMinutes"`
	EarlyLeaveMinutes    int `json:"earlyLeaveMinutes"`

	// 交通費集計
	DailyTransportationAmount int    `json:"dailyTransportationAmount"`
	CommuterPassAmount        int    `json:"commuterPassAmount"`
	TotalTransportationAmount int    `json:"totalTransportationAmount"`
	CommuterPassFrom          string `json:"commuterPassFrom"`
	CommuterPassTo            string `json:"commuterPassTo"`
	CommuterPassMethod        string `json:"commuterPassMethod"`
	DailyTransportationCount  int    `json:"dailyTransportationCount"`

	// 有給集計
	PaidLeaveUsedDays    float64 `json:"paidLeaveUsedDays"`
	PaidLeaveUsedMinutes int     `json:"paidLeaveUsedMinutes"`

	// 経費集計
	ExpenseTotalAmount          int `json:"expenseTotalAmount"`
	SalaryIncludedExpenseAmount int `json:"salaryIncludedExpenseAmount"`
	ExpenseCount                int `json:"expenseCount"`
	TransportationExpenseAmount int `json:"transportationExpenseAmount"`
	SuppliesExpenseAmount       int `json:"suppliesExpenseAmount"`
	CommunicationExpenseAmount  int `json:"communicationExpenseAmount"`
	OtherExpenseAmount          int `json:"otherExpenseAmount"`

	// 祝日・営業日補助
	HolidayCount    int `json:"holidayCount"`
	WorkingDayCount int `json:"workingDayCount"`

	// 警告・不整合
	WarningCount          int    `json:"warningCount"`
	Warnings              string `json:"warnings"`
	MissingAttendanceDays int    `json:"missingAttendanceDays"`
	InvalidBreakCount     int    `json:"invalidBreakCount"`
	InvalidTimeCount      int    `json:"invalidTimeCount"`
	HasDataWarning        bool   `json:"hasDataWarning"`
}

/*
 * 月次勤怠集計用の日別内部データ
 *
 * service 内で日別計算・週別計算を行うための作業用。
 *
 * 注意：
 * ・予定区分は attendance_types を参照する
 * ・実績状態は constants/attendance_status_constants.go の固定値を使う
 * ・ActualAttendanceTypeID / ActualAttendanceTypeCode / ActualAttendanceTypeIsWorked は使わない
 */
type MonthlyAttendanceSummaryWorkRow struct {
	UserID   uint
	WorkDate string

	AttendanceDayID uint

	// 予定区分
	PlanAttendanceTypeID       uint
	PlanAttendanceTypeCode     string
	PlanAttendanceTypeCategory string

	// 実績状態
	// 例：NORMAL, ABSENCE, SICK_LEAVE, LATE, EARLY_LEAVE
	ActualWorkStatus string

	ScheduledWorkMinutes int

	ActualWorkMinutes int
	BreakMinutes      int

	LateNightWorkMinutes int

	IsActualWorkDay    bool
	IsPlannedHoliday   bool
	IsHolidayWorkDay   bool
	IsPaidLeaveDay     bool
	IsHalfPaidLeaveDay bool
	IsAbsenceDay       bool
	IsSickLeaveDay     bool

	DailyOvertimeThresholdMinutes int
	DailyOvertimeMinutes          int

	WorkShortageMinutes               int
	WorkExcessAgainstScheduledMinutes int

	TransportAmount int

	Warnings []string
}

/*
 * 月次勤怠集計用の週単位内部データ
 *
 * 週起算日は constants.AttendanceSummaryWeekStartDay を使う。
 * 月をまたぐ週もこの単位で計算する。
 */
type MonthlyAttendanceSummaryWeekWork struct {
	WeekStartDate string
	WeekEndDate   string

	ScheduledWorkMinutes int
	ActualWorkMinutes    int
	DailyOvertimeMinutes int
	HolidayWorkMinutes   int

	WeeklyOvertimeThresholdMinutes int
	WeeklyOvertimeMinutes          int
}

/*
 * 月次勤怠集計CSV 警告
 */
type MonthlyAttendanceSummaryWarning struct {
	UserID   uint   `json:"userId"`
	WorkDate string `json:"workDate"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

/*
 * CSV出力ステータス
 */
const MonthlyAttendanceSummaryExportStatusRowOutput = "ROW_OUTPUT"

/*
 * 集計ステータス
 */
const MonthlyAttendanceSummaryCalculationStatusCalculated = "CALCULATED"
const MonthlyAttendanceSummaryCalculationStatusSkippedNotApproved = "SKIPPED_NOT_APPROVED"
const MonthlyAttendanceSummaryCalculationStatusError = "ERROR"

/*
 * 月次勤怠ステータス
 *
 * NOT_SUBMITTED はDBには保存しない。
 * 月次申請レコードが存在しない場合のCSV表示用。
 */
const MonthlyAttendanceSummaryMonthlyStatusNotSubmitted = "NOT_SUBMITTED"
const MonthlyAttendanceSummaryMonthlyStatusApproved = "APPROVED"
