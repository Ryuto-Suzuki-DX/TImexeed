package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type AttendanceTransportExpenseBuilder interface {
	BuildSearchAttendanceTransportExpensesQuery(req types.SearchAttendanceTransportExpensesRequest) (*gorm.DB, results.Result)
	BuildFindAttendanceTransportExpensesByAttendanceDayIDQuery(attendanceDayID uint) (*gorm.DB, results.Result)
	BuildFindAttendanceTransportExpenseByIDQuery(attendanceTransportExpenseID uint) (*gorm.DB, results.Result)
	BuildCreateAttendanceTransportExpenseModel(
		attendanceDayID uint,
		req types.UpdateAttendanceTransportExpensesByWorkDateExpenseRequest,
		sortOrder int,
	) (models.AttendanceTransportExpense, results.Result)
	BuildUpdateAttendanceTransportExpenseModel(
		currentAttendanceTransportExpense models.AttendanceTransportExpense,
		req types.UpdateAttendanceTransportExpensesByWorkDateExpenseRequest,
		sortOrder int,
	) (models.AttendanceTransportExpense, results.Result)
	BuildDeleteAttendanceTransportExpenseModel(
		currentAttendanceTransportExpense models.AttendanceTransportExpense,
	) (models.AttendanceTransportExpense, results.Result)
}

/*
 * 管理者用日別交通費Builder
 *
 * 役割：
 * ・検索用GORMクエリを作成する
 * ・保存用Modelを作成する
 *
 * 注意：
 * ・DB実行はRepositoryに任せる
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
type attendanceTransportExpenseBuilder struct {
	db *gorm.DB
}

/*
 * AttendanceTransportExpenseBuilder生成
 */
func NewAttendanceTransportExpenseBuilder(db *gorm.DB) AttendanceTransportExpenseBuilder {
	return &attendanceTransportExpenseBuilder{db: db}
}

/*
 * 対象ユーザー・対象年月の日別交通費検索用クエリ作成
 */
func (builder *attendanceTransportExpenseBuilder) BuildSearchAttendanceTransportExpensesQuery(
	req types.SearchAttendanceTransportExpensesRequest,
) (*gorm.DB, results.Result) {
	if req.TargetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_QUERY_INVALID_TARGET_USER_ID",
			"日別交通費検索条件の作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

	if req.TargetYear <= 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_QUERY_INVALID_TARGET_YEAR",
			"日別交通費検索条件の作成に失敗しました",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_QUERY_INVALID_TARGET_MONTH",
			"日別交通費検索条件の作成に失敗しました",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	startDate := time.Date(req.TargetYear, time.Month(req.TargetMonth), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	query := builder.db.
		Model(&models.AttendanceTransportExpense{}).
		Joins("JOIN attendance_days ON attendance_days.id = attendance_transport_expenses.attendance_day_id").
		Preload("AttendanceDay").
		Where("attendance_days.user_id = ?", req.TargetUserID).
		Where("attendance_days.work_date >= ?", startDate).
		Where("attendance_days.work_date < ?", endDate).
		Where("attendance_days.is_deleted = ?", false).
		Where("attendance_transport_expenses.is_deleted = ?", false).
		Order("attendance_days.work_date ASC").
		Order("attendance_transport_expenses.sort_order ASC").
		Order("attendance_transport_expenses.id ASC")

	return query, results.OK(
		nil,
		"BUILD_SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠日IDに紐づく日別交通費一覧取得用クエリ作成
 */
func (builder *attendanceTransportExpenseBuilder) BuildFindAttendanceTransportExpensesByAttendanceDayIDQuery(
	attendanceDayID uint,
) (*gorm.DB, results.Result) {
	if attendanceDayID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ATTENDANCE_TRANSPORT_EXPENSES_QUERY_INVALID_ATTENDANCE_DAY_ID",
			"日別交通費取得条件の作成に失敗しました",
			nil,
		)
	}

	query := builder.db.
		Model(&models.AttendanceTransportExpense{}).
		Where("attendance_day_id = ?", attendanceDayID).
		Where("is_deleted = ?", false).
		Order("sort_order ASC").
		Order("id ASC")

	return query, results.OK(
		nil,
		"BUILD_FIND_ATTENDANCE_TRANSPORT_EXPENSES_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 日別交通費IDによる1件取得用クエリ作成
 */
func (builder *attendanceTransportExpenseBuilder) BuildFindAttendanceTransportExpenseByIDQuery(
	attendanceTransportExpenseID uint,
) (*gorm.DB, results.Result) {
	if attendanceTransportExpenseID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ATTENDANCE_TRANSPORT_EXPENSE_QUERY_INVALID_ID",
			"日別交通費取得条件の作成に失敗しました",
			nil,
		)
	}

	query := builder.db.
		Model(&models.AttendanceTransportExpense{}).
		Where("id = ?", attendanceTransportExpenseID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_ATTENDANCE_TRANSPORT_EXPENSE_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 日別交通費作成用Model作成
 */
func (builder *attendanceTransportExpenseBuilder) BuildCreateAttendanceTransportExpenseModel(
	attendanceDayID uint,
	req types.UpdateAttendanceTransportExpensesByWorkDateExpenseRequest,
	sortOrder int,
) (models.AttendanceTransportExpense, results.Result) {
	if attendanceDayID == 0 {
		return models.AttendanceTransportExpense{}, results.BadRequest(
			"BUILD_CREATE_ATTENDANCE_TRANSPORT_EXPENSE_MODEL_INVALID_ATTENDANCE_DAY_ID",
			"日別交通費作成データの作成に失敗しました",
			nil,
		)
	}

	attendanceTransportExpense := models.AttendanceTransportExpense{
		AttendanceDayID: attendanceDayID,
		SortOrder:       sortOrder,
		TransportFrom:   req.TransportFrom,
		TransportTo:     req.TransportTo,
		TransportMethod: req.TransportMethod,
		TransportAmount: req.TransportAmount,
		TransportMemo:   req.TransportMemo,
		IsDeleted:       false,
	}

	return attendanceTransportExpense, results.OK(
		nil,
		"BUILD_CREATE_ATTENDANCE_TRANSPORT_EXPENSE_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 日別交通費更新用Model作成
 */
func (builder *attendanceTransportExpenseBuilder) BuildUpdateAttendanceTransportExpenseModel(
	currentAttendanceTransportExpense models.AttendanceTransportExpense,
	req types.UpdateAttendanceTransportExpensesByWorkDateExpenseRequest,
	sortOrder int,
) (models.AttendanceTransportExpense, results.Result) {
	if currentAttendanceTransportExpense.ID == 0 {
		return models.AttendanceTransportExpense{}, results.BadRequest(
			"BUILD_UPDATE_ATTENDANCE_TRANSPORT_EXPENSE_MODEL_EMPTY_CURRENT_DATA",
			"日別交通費更新データの作成に失敗しました",
			nil,
		)
	}

	currentAttendanceTransportExpense.SortOrder = sortOrder
	currentAttendanceTransportExpense.TransportFrom = req.TransportFrom
	currentAttendanceTransportExpense.TransportTo = req.TransportTo
	currentAttendanceTransportExpense.TransportMethod = req.TransportMethod
	currentAttendanceTransportExpense.TransportAmount = req.TransportAmount
	currentAttendanceTransportExpense.TransportMemo = req.TransportMemo
	currentAttendanceTransportExpense.IsDeleted = false
	currentAttendanceTransportExpense.DeletedAt = nil

	return currentAttendanceTransportExpense, results.OK(
		nil,
		"BUILD_UPDATE_ATTENDANCE_TRANSPORT_EXPENSE_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 日別交通費論理削除用Model作成
 */
func (builder *attendanceTransportExpenseBuilder) BuildDeleteAttendanceTransportExpenseModel(
	currentAttendanceTransportExpense models.AttendanceTransportExpense,
) (models.AttendanceTransportExpense, results.Result) {
	if currentAttendanceTransportExpense.ID == 0 {
		return models.AttendanceTransportExpense{}, results.BadRequest(
			"BUILD_DELETE_ATTENDANCE_TRANSPORT_EXPENSE_MODEL_EMPTY_CURRENT_DATA",
			"日別交通費削除データの作成に失敗しました",
			nil,
		)
	}

	now := time.Now()
	currentAttendanceTransportExpense.IsDeleted = true
	currentAttendanceTransportExpense.DeletedAt = &now

	return currentAttendanceTransportExpense, results.OK(
		nil,
		"BUILD_DELETE_ATTENDANCE_TRANSPORT_EXPENSE_MODEL_SUCCESS",
		"",
		nil,
	)
}
