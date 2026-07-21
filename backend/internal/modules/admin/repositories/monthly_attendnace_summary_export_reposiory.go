package repositories

import (
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 月次勤怠集計CSV出力 Repository interface
 *
 * 注意：
 * ・CSV集計に必要なデータをまとめて取得する
 * ・集計計算そのものはServiceで行う
 * ・CSV生成そのものはBuilderで行う
 */
type MonthlyAttendanceSummaryExportRepository interface {
	SearchExportTargetUsers(req types.ExportMonthlyAttendanceSummaryCsvRequest) ([]MonthlyAttendanceSummaryExportUserRecord, results.Result)
	FindMonthlyAttendanceRequests(userIDs []uint, targetYear int, targetMonth int) (map[uint]models.MonthlyAttendanceRequest, results.Result)
	FindAttendanceDays(userIDs []uint, fromDate time.Time, toDate time.Time) ([]models.AttendanceDay, results.Result)
	FindAttendanceBreaks(attendanceDayIDs []uint) (map[uint][]models.AttendanceBreak, results.Result)
	FindAttendanceTransportExpenses(attendanceDayIDs []uint) (map[uint][]models.AttendanceTransportExpense, results.Result)
	FindMonthlyCommuterPasses(userIDs []uint, targetYear int, targetMonth int) (map[uint][]models.MonthlyCommuterPass, results.Result)
	FindUserSalaryDetails(userIDs []uint, targetMonthStart time.Time, targetMonthEnd time.Time) (map[uint]models.UserSalaryDetail, results.Result)
	FindPaidLeaveUsages(userIDs []uint, targetMonthStart time.Time, targetMonthEnd time.Time) (map[uint][]models.PaidLeaveUsage, results.Result)
	FindExpenses(userIDs []uint, targetMonthStart time.Time) (map[uint][]models.Expense, results.Result)
}

/*
 * CSV出力対象ユーザー取得用Record
 *
 * User model には Department のrelationがないため、
 * users と departments をJOINしてこのRecordへScanする。
 */
type MonthlyAttendanceSummaryExportUserRecord struct {
	ID             uint
	Name           string
	Email          string
	Role           string
	DepartmentID   *uint
	DepartmentName *string
	HireDate       time.Time
	RetirementDate *time.Time
	IsDeleted      bool
}

/*
 * 月次勤怠集計CSV出力 Repository
 */
type monthlyAttendanceSummaryExportRepository struct {
	db *gorm.DB
}

/*
 * MonthlyAttendanceSummaryExportRepository生成
 */
func NewMonthlyAttendanceSummaryExportRepository(db *gorm.DB) MonthlyAttendanceSummaryExportRepository {
	return &monthlyAttendanceSummaryExportRepository{
		db: db,
	}
}

/*
 * CSV出力対象ユーザー検索
 */
func (repository *monthlyAttendanceSummaryExportRepository) SearchExportTargetUsers(
	req types.ExportMonthlyAttendanceSummaryCsvRequest,
) ([]MonthlyAttendanceSummaryExportUserRecord, results.Result) {
	query := repository.db.
		Table("users").
		Select(`
			users.id,
			users.name,
			users.email,
			users.role,
			users.department_id,
			departments.name AS department_name,
			users.hire_date,
			users.retirement_date,
			users.is_deleted
		`).
		Joins("LEFT JOIN departments ON departments.id = users.department_id AND departments.is_deleted = false").
		Where("users.is_deleted = false").
		Where("users.role = ?", "USER")

	if len(req.TargetUserIDs) > 0 {
		query = query.Where("users.id IN ?", req.TargetUserIDs)
	}

	if req.DepartmentID != nil && *req.DepartmentID != 0 {
		query = query.Where("users.department_id = ?", *req.DepartmentID)
	}

	if strings.TrimSpace(req.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(req.Keyword) + "%"
		query = query.Where("(users.name LIKE ? OR users.email LIKE ?)", keyword, keyword)
	}

	var users []MonthlyAttendanceSummaryExportUserRecord
	if err := query.Order("users.id ASC").Scan(&users).Error; err != nil {
		return nil, results.BadRequest(
			"SEARCH_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_USERS_FAILED",
			"月次勤怠集計CSV出力対象ユーザーの取得に失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	return users, results.OK(
		nil,
		"SEARCH_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_USERS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次勤怠申請取得
 */
func (repository *monthlyAttendanceSummaryExportRepository) FindMonthlyAttendanceRequests(
	userIDs []uint,
	targetYear int,
	targetMonth int,
) (map[uint]models.MonthlyAttendanceRequest, results.Result) {
	monthlyAttendanceRequestMap := map[uint]models.MonthlyAttendanceRequest{}

	if len(userIDs) == 0 {
		return monthlyAttendanceRequestMap, results.OK(
			nil,
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_MONTHLY_REQUESTS_EMPTY",
			"",
			nil,
		)
	}

	var monthlyAttendanceRequests []models.MonthlyAttendanceRequest
	if err := repository.db.
		Where("is_deleted = false").
		Where("user_id IN ?", userIDs).
		Where("target_year = ?", targetYear).
		Where("target_month = ?", targetMonth).
		Find(&monthlyAttendanceRequests).Error; err != nil {
		return nil, results.BadRequest(
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_MONTHLY_REQUESTS_FAILED",
			"月次勤怠申請の取得に失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	for _, monthlyAttendanceRequest := range monthlyAttendanceRequests {
		monthlyAttendanceRequestMap[monthlyAttendanceRequest.UserID] = monthlyAttendanceRequest
	}

	return monthlyAttendanceRequestMap, results.OK(
		nil,
		"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_MONTHLY_REQUESTS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠日取得
 *
 * 月をまたぐ週残業・休日出勤判定があるため、対象月だけではなく、
 * Serviceで算出した拡張期間で取得する。
 */
func (repository *monthlyAttendanceSummaryExportRepository) FindAttendanceDays(
	userIDs []uint,
	fromDate time.Time,
	toDate time.Time,
) ([]models.AttendanceDay, results.Result) {
	if len(userIDs) == 0 {
		return []models.AttendanceDay{}, results.OK(
			nil,
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_ATTENDANCE_DAYS_EMPTY",
			"",
			nil,
		)
	}

	var attendanceDays []models.AttendanceDay
	if err := repository.db.
		Preload("PlanAttendanceType").
		Where("is_deleted = false").
		Where("user_id IN ?", userIDs).
		Where("work_date >= ?", fromDate).
		Where("work_date <= ?", toDate).
		Order("user_id ASC, work_date ASC").
		Find(&attendanceDays).Error; err != nil {
		return nil, results.BadRequest(
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_ATTENDANCE_DAYS_FAILED",
			"勤怠日の取得に失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	return attendanceDays, results.OK(
		nil,
		"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_ATTENDANCE_DAYS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 休憩取得
 */
func (repository *monthlyAttendanceSummaryExportRepository) FindAttendanceBreaks(
	attendanceDayIDs []uint,
) (map[uint][]models.AttendanceBreak, results.Result) {
	attendanceBreakMap := map[uint][]models.AttendanceBreak{}

	if len(attendanceDayIDs) == 0 {
		return attendanceBreakMap, results.OK(
			nil,
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_ATTENDANCE_BREAKS_EMPTY",
			"",
			nil,
		)
	}

	var attendanceBreaks []models.AttendanceBreak
	if err := repository.db.
		Where("is_deleted = false").
		Where("attendance_day_id IN ?", attendanceDayIDs).
		Order("attendance_day_id ASC, break_start_at ASC").
		Find(&attendanceBreaks).Error; err != nil {
		return nil, results.BadRequest(
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_ATTENDANCE_BREAKS_FAILED",
			"休憩の取得に失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	for _, attendanceBreak := range attendanceBreaks {
		attendanceBreakMap[attendanceBreak.AttendanceDayID] = append(attendanceBreakMap[attendanceBreak.AttendanceDayID], attendanceBreak)
	}

	return attendanceBreakMap, results.OK(
		nil,
		"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_ATTENDANCE_BREAKS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 日別交通費取得
 *
 * AttendanceDayIDごとに日別交通費明細をまとめて返す。
 *
 * 注意：
 * ・論理削除済みの明細は対象外
 * ・金額や件数の集計はServiceで行う
 */
func (repository *monthlyAttendanceSummaryExportRepository) FindAttendanceTransportExpenses(
	attendanceDayIDs []uint,
) (map[uint][]models.AttendanceTransportExpense, results.Result) {
	attendanceTransportExpenseMap := map[uint][]models.AttendanceTransportExpense{}

	if len(attendanceDayIDs) == 0 {
		return attendanceTransportExpenseMap, results.OK(
			nil,
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_TRANSPORT_EXPENSES_EMPTY",
			"",
			nil,
		)
	}

	var attendanceTransportExpenses []models.AttendanceTransportExpense
	if err := repository.db.
		Where("is_deleted = false").
		Where("attendance_day_id IN ?", attendanceDayIDs).
		Order("attendance_day_id ASC, sort_order ASC, id ASC").
		Find(&attendanceTransportExpenses).Error; err != nil {
		return nil, results.BadRequest(
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_TRANSPORT_EXPENSES_FAILED",
			"日別交通費の取得に失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	for _, attendanceTransportExpense := range attendanceTransportExpenses {
		attendanceTransportExpenseMap[attendanceTransportExpense.AttendanceDayID] = append(
			attendanceTransportExpenseMap[attendanceTransportExpense.AttendanceDayID],
			attendanceTransportExpense,
		)
	}

	return attendanceTransportExpenseMap, results.OK(
		nil,
		"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_TRANSPORT_EXPENSES_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次通勤定期取得
 *
 * ユーザーIDごとに、対象年月の有効な月次通勤定期を複数件まとめて返す。
 *
 * 注意：
 * ・同じユーザー・対象年月に複数件登録できる
 * ・論理削除済みの定期は対象外
 * ・金額の合計と表示文字列の生成はServiceで行う
 */
func (repository *monthlyAttendanceSummaryExportRepository) FindMonthlyCommuterPasses(
	userIDs []uint,
	targetYear int,
	targetMonth int,
) (map[uint][]models.MonthlyCommuterPass, results.Result) {
	monthlyCommuterPassMap := map[uint][]models.MonthlyCommuterPass{}

	if len(userIDs) == 0 {
		return monthlyCommuterPassMap, results.OK(
			nil,
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_COMMUTER_PASSES_EMPTY",
			"",
			nil,
		)
	}

	var monthlyCommuterPasses []models.MonthlyCommuterPass
	if err := repository.db.
		Where("is_deleted = false").
		Where("user_id IN ?", userIDs).
		Where("target_year = ?", targetYear).
		Where("target_month = ?", targetMonth).
		Order("user_id ASC, id ASC").
		Find(&monthlyCommuterPasses).Error; err != nil {
		return nil, results.BadRequest(
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_COMMUTER_PASSES_FAILED",
			"月次通勤定期の取得に失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	for _, monthlyCommuterPass := range monthlyCommuterPasses {
		monthlyCommuterPassMap[monthlyCommuterPass.UserID] = append(
			monthlyCommuterPassMap[monthlyCommuterPass.UserID],
			monthlyCommuterPass,
		)
	}

	return monthlyCommuterPassMap, results.OK(
		nil,
		"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_COMMUTER_PASSES_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー給与詳細取得
 *
 * 対象月に有効な給与詳細を取得する。
 * 複数件ある場合は EffectiveFrom が一番新しいものを採用する。
 */
func (repository *monthlyAttendanceSummaryExportRepository) FindUserSalaryDetails(
	userIDs []uint,
	targetMonthStart time.Time,
	targetMonthEnd time.Time,
) (map[uint]models.UserSalaryDetail, results.Result) {
	userSalaryDetailMap := map[uint]models.UserSalaryDetail{}

	if len(userIDs) == 0 {
		return userSalaryDetailMap, results.OK(
			nil,
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_USER_SALARY_DETAILS_EMPTY",
			"",
			nil,
		)
	}

	var userSalaryDetails []models.UserSalaryDetail
	if err := repository.db.
		Where("is_deleted = false").
		Where("user_id IN ?", userIDs).
		Where("effective_from <= ?", targetMonthEnd).
		Where("(effective_to IS NULL OR effective_to >= ?)", targetMonthStart).
		Order("user_id ASC, effective_from DESC, id DESC").
		Find(&userSalaryDetails).Error; err != nil {
		return nil, results.BadRequest(
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_USER_SALARY_DETAILS_FAILED",
			"ユーザー給与詳細の取得に失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	for _, userSalaryDetail := range userSalaryDetails {
		if _, exists := userSalaryDetailMap[userSalaryDetail.UserID]; exists {
			continue
		}

		userSalaryDetailMap[userSalaryDetail.UserID] = userSalaryDetail
	}

	return userSalaryDetailMap, results.OK(
		nil,
		"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_USER_SALARY_DETAILS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給使用履歴取得
 */
func (repository *monthlyAttendanceSummaryExportRepository) FindPaidLeaveUsages(
	userIDs []uint,
	targetMonthStart time.Time,
	targetMonthEnd time.Time,
) (map[uint][]models.PaidLeaveUsage, results.Result) {
	paidLeaveUsageMap := map[uint][]models.PaidLeaveUsage{}

	if len(userIDs) == 0 {
		return paidLeaveUsageMap, results.OK(
			nil,
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_PAID_LEAVE_USAGES_EMPTY",
			"",
			nil,
		)
	}

	var paidLeaveUsages []models.PaidLeaveUsage
	if err := repository.db.
		Where("is_deleted = false").
		Where("user_id IN ?", userIDs).
		Where("usage_date >= ?", targetMonthStart).
		Where("usage_date <= ?", targetMonthEnd).
		Order("user_id ASC, usage_date ASC").
		Find(&paidLeaveUsages).Error; err != nil {
		return nil, results.BadRequest(
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_PAID_LEAVE_USAGES_FAILED",
			"有給使用履歴の取得に失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	for _, paidLeaveUsage := range paidLeaveUsages {
		paidLeaveUsageMap[paidLeaveUsage.UserID] = append(paidLeaveUsageMap[paidLeaveUsage.UserID], paidLeaveUsage)
	}

	return paidLeaveUsageMap, results.OK(
		nil,
		"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_PAID_LEAVE_USAGES_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費取得
 *
 * Expense.TargetMonth は月初日をdate型で保持する。
 */
func (repository *monthlyAttendanceSummaryExportRepository) FindExpenses(
	userIDs []uint,
	targetMonthStart time.Time,
) (map[uint][]models.Expense, results.Result) {
	expenseMap := map[uint][]models.Expense{}

	if len(userIDs) == 0 {
		return expenseMap, results.OK(
			nil,
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_EXPENSES_EMPTY",
			"",
			nil,
		)
	}

	var expenses []models.Expense
	if err := repository.db.
		Where("is_deleted = false").
		Where("user_id IN ?", userIDs).
		Where("target_month = ?", targetMonthStart).
		Order("user_id ASC, expense_date ASC, id ASC").
		Find(&expenses).Error; err != nil {
		return nil, results.BadRequest(
			"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_EXPENSES_FAILED",
			"経費の取得に失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	for _, expense := range expenses {
		expenseMap[expense.UserID] = append(expenseMap[expense.UserID], expense)
	}

	return expenseMap, results.OK(
		nil,
		"FIND_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_EXPENSES_SUCCESS",
		"",
		nil,
	)
}
