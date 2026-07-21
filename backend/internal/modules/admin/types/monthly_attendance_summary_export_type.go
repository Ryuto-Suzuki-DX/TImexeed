package types

/*
 * 月次勤怠集計CSV出力 Type
 *
 * 管理者専用。
 *
 * 注意：
 * ・給与計算そのものは行わない
 * ・給与計算に必要な、Timexeed内に存在する月次情報をCSVへ出力する
 * ・APPROVED の月だけ集計値を出力する
 * ・APPROVED 以外の月はステータスと警告のみ出力する
 * ・変形労働制フラグは持たず、AttendanceDay.ScheduledWorkMinutes の値だけで判断する
 * ・予定区分は attendance_types を参照する
 * ・実績状態は constants/attendance_status_constants.go の固定値を使う
 * ・ActualAttendanceTypeID / ActualAttendanceTypeCode / ActualAttendanceTypeIsWorked は使わない
 */

/*
 * 月次勤怠集計ファイル出力対象種別
 *
 * USER:
 * ・指定したユーザー1人だけを出力する
 *
 * DEPARTMENT:
 * ・指定した複数所属に該当する一般ユーザーを出力する
 * ・IncludeUnassignedDepartment が true の場合は所属なしユーザーも含める
 */
const (
	MonthlyAttendanceSummaryExportTargetTypeUser       = "USER"
	MonthlyAttendanceSummaryExportTargetTypeDepartment = "DEPARTMENT"
)

/*
 * 月次勤怠集計CSV出力 Request
 *
 * TargetType:
 * ・USER       ：ユーザー単体で出力する
 * ・DEPARTMENT ：複数所属単位で出力する
 *
 * TargetUserID:
 * ・TargetType が USER の場合に必須
 * ・選択したユーザー1人のIDを指定する
 *
 * DepartmentIDs:
 * ・TargetType が DEPARTMENT の場合に使用する
 * ・複数の所属IDを指定できる
 *
 * IncludeUnassignedDepartment:
 * ・TargetType が DEPARTMENT の場合に使用する
 * ・true の場合は department_id が NULL の所属なしユーザーも対象にする
 *
 * IncludeNotApproved:
 * ・true の場合、APPROVED 以外のユーザーもステータスと警告のみ出力する
 * ・false の場合、APPROVED のユーザーだけ出力する
 *
 * 注意：
 * ・CSV出力APIではフリーワード検索を行わない
 * ・ユーザー検索は /admin/users/search-business-targets を使用する
 * ・所属一覧取得は /admin/departments/search を使用する
 */
type ExportMonthlyAttendanceSummaryCsvRequest struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	// USER または DEPARTMENT
	TargetType string `json:"targetType"`

	// USERの場合のみ使用する
	TargetUserID *uint `json:"targetUserId"`

	// DEPARTMENTの場合のみ使用する
	DepartmentIDs []uint `json:"departmentIds"`

	// DEPARTMENTの場合のみ使用する
	IncludeUnassignedDepartment bool `json:"includeUnassignedDepartment"`

	IncludeNotApproved bool   `json:"includeNotApproved"`
	Format             string `json:"format"`
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
 *   ユーザー情報、対象年月、月次申請ステータス、calculationStatus、警告のみセットし、
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
	MissingAttendanceDays    int `json:"missingAttendanceDays"`
	ScheduledWorkDays        int `json:"scheduledWorkDays"`
	ActualWorkDays           int `json:"actualWorkDays"`
	DayShiftWorkDays         int `json:"dayShiftWorkDays"`
	NightShiftWorkDays       int `json:"nightShiftWorkDays"`
	PlannedHolidayDays       int `json:"plannedHolidayDays"`
	PaidLeaveDays            int `json:"paidLeaveDays"`
	HalfPaidLeaveDays        int `json:"halfPaidLeaveDays"`
	AbsenceDays              int `json:"absenceDays"`
	SickLeaveDays            int `json:"sickLeaveDays"`
	HolidayWorkDays          int `json:"holidayWorkDays"`
	LateDays                 int `json:"lateDays"`
	EarlyLeaveDays           int `json:"earlyLeaveDays"`
	ScheduledButNoActualDays int `json:"scheduledButNoActualDays"`
	ActualButNoScheduledDays int `json:"actualButNoScheduledDays"`
	MissingScheduledWorkDays int `json:"missingScheduledWorkDays"`
	WorkingDayCount          int `json:"workingDayCount"`
	HolidayCount             int `json:"holidayCount"`

	// 勤怠時間集計
	CompanyDailyStandardWorkMinutes  int `json:"companyDailyStandardWorkMinutes"`
	CompanyWeeklyStandardWorkMinutes int `json:"companyWeeklyStandardWorkMinutes"`

	ScheduledWorkMinutes              int `json:"scheduledWorkMinutes"`
	ActualWorkMinutes                 int `json:"actualWorkMinutes"`
	DayWorkMinutes                    int `json:"dayWorkMinutes"`
	NightWorkMinutes                  int `json:"nightWorkMinutes"`
	BreakMinutes                      int `json:"breakMinutes"`
	RegularWorkMinutes                int `json:"regularWorkMinutes"`
	WorkShortageMinutes               int `json:"workShortageMinutes"`
	WorkExcessAgainstScheduledMinutes int `json:"workExcessAgainstScheduledMinutes"`

	DailyOvertimeThresholdMinutes  int `json:"dailyOvertimeThresholdMinutes"`
	WeeklyScheduledWorkMinutes     int `json:"weeklyScheduledWorkMinutes"`
	WeeklyOvertimeThresholdMinutes int `json:"weeklyOvertimeThresholdMinutes"`
	DailyOvertimeMinutes           int `json:"dailyOvertimeMinutes"`
	WeeklyOvertimeMinutes          int `json:"weeklyOvertimeMinutes"`
	OvertimeMinutes                int `json:"overtimeMinutes"`
	DayOvertimeMinutes             int `json:"dayOvertimeMinutes"`
	NightOvertimeMinutes           int `json:"nightOvertimeMinutes"`

	LateNightWorkMinutes        int `json:"lateNightWorkMinutes"`
	HolidayWorkMinutes          int `json:"holidayWorkMinutes"`
	HolidayLateNightWorkMinutes int `json:"holidayLateNightWorkMinutes"`
	PaidLeaveMinutes            int `json:"paidLeaveMinutes"`
	AbsenceMinutes              int `json:"absenceMinutes"`
	SickLeaveMinutes            int `json:"sickLeaveMinutes"`
	LateMinutes                 int `json:"lateMinutes"`
	EarlyLeaveMinutes           int `json:"earlyLeaveMinutes"`
	DeductionTargetMinutes      int `json:"deductionTargetMinutes"`

	// 稼働率
	ActualOperationRate        float64 `json:"actualOperationRate"`
	PayrollTargetOperationRate float64 `json:"payrollTargetOperationRate"`
	OperationRateJudge         string  `json:"operationRateJudge"`

	// 交通費集計
	DailyTransportationAmount int `json:"dailyTransportationAmount"`

	// 対象年月に登録された月次通勤定期の金額合計
	CommuterPassAmount int `json:"commuterPassAmount"`

	// 日別交通費合計 + 月次通勤定期合計
	TotalTransportationAmount int `json:"totalTransportationAmount"`

	// 月次通勤定期が複数件ある場合は「 / 」で連結する
	CommuterPassFrom   string `json:"commuterPassFrom"`
	CommuterPassTo     string `json:"commuterPassTo"`
	CommuterPassMethod string `json:"commuterPassMethod"`

	DailyTransportationCount int `json:"dailyTransportationCount"`

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

	// 警告・不整合
	WarningCount                     int    `json:"warningCount"`
	Warnings                         string `json:"warnings"`
	InvalidBreakCount                int    `json:"invalidBreakCount"`
	InvalidTimeCount                 int    `json:"invalidTimeCount"`
	HasDataWarning                   bool   `json:"hasDataWarning"`
	HasSalarySettingWarning          bool   `json:"hasSalarySettingWarning"`
	HasPayrollExcludedWarning        bool   `json:"hasPayrollExcludedWarning"`
	HasMonthlyApprovalWarning        bool   `json:"hasMonthlyApprovalWarning"`
	HasAttendanceMissingWarning      bool   `json:"hasAttendanceMissingWarning"`
	HasScheduleActualMismatchWarning bool   `json:"hasScheduleActualMismatchWarning"`
	HasExpenseCategoryWarning        bool   `json:"hasExpenseCategoryWarning"`
}

/*
 * 月次勤怠集計用の日別内部データ
 *
 * service 内で日別計算・週別計算を行うための作業用。
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
	DayWorkMinutes    int
	NightWorkMinutes  int
	BreakMinutes      int

	WorkMinuteSegments []MonthlyAttendanceSummaryWorkMinuteSegment

	RegularWorkMinutes int

	LateNightWorkMinutes        int
	HolidayLateNightWorkMinutes int

	IsActualWorkDay    bool
	IsPlannedHoliday   bool
	IsHolidayWorkDay   bool
	IsPaidLeaveDay     bool
	IsHalfPaidLeaveDay bool
	IsAbsenceDay       bool
	IsSickLeaveDay     bool

	IsScheduledButNoActual bool
	IsActualButNoScheduled bool
	IsMissingScheduledWork bool

	DailyOvertimeThresholdMinutes int
	DailyOvertimeMinutes          int
	DayOvertimeMinutes            int
	NightOvertimeMinutes          int

	WorkShortageMinutes               int
	WorkExcessAgainstScheduledMinutes int

	AbsenceMinutes    int
	SickLeaveMinutes  int
	LateMinutes       int
	EarlyLeaveMinutes int

	TransportAmount int
	TransportCount  int

	Warnings []string
}

/*
 * 月次勤怠集計用の日中/深夜分解済み勤務区間
 */
type MonthlyAttendanceSummaryWorkMinuteSegment struct {
	Minutes     int
	IsLateNight bool
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
 * 出力形式
 *
 * 未指定の場合はCSVとして扱う。
 */
const MonthlyAttendanceSummaryExportFormatCSV = "CSV"
const MonthlyAttendanceSummaryExportFormatXLSX = "XLSX"

/*
 * 出力ステータス
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

/*
 * 稼働率判定
 */
const MonthlyAttendanceSummaryOperationRateJudgeNotAvailable = "判定不可"
const MonthlyAttendanceSummaryOperationRateJudgeOver80 = "80%以上"
const MonthlyAttendanceSummaryOperationRateJudgeUnder80 = "80%未満"
