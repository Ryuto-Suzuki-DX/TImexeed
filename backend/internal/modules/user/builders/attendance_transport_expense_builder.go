package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type AttendanceTransportExpenseBuilder interface {
	BuildSearchAttendanceTransportExpensesQuery(
		userID uint,
		req types.SearchAttendanceTransportExpensesRequest,
	) (*gorm.DB, results.Result)
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
 * 従業員用日別交通費Builder
 *
 * 注意：
 * ・DB実行はRepositoryに任せる
 * ・検索対象はJWT由来のログイン中ユーザー本人だけ
 */
type attendanceTransportExpenseBuilder struct {
	db *gorm.DB
}

func NewAttendanceTransportExpenseBuilder(db *gorm.DB) AttendanceTransportExpenseBuilder {
	return &attendanceTransportExpenseBuilder{db: db}
}

/*
 * 対象年月の本人の日別交通費検索用クエリ作成
 */
func (builder *attendanceTransportExpenseBuilder) BuildSearchAttendanceTransportExpensesQuery(
	userID uint,
	req types.SearchAttendanceTransportExpensesRequest,
) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_QUERY_INVALID_USER_ID",
			"日別交通費検索条件の作成に失敗しました",
			map[string]any{"userId": userID},
		)
	}

	if req.TargetYear <= 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_QUERY_INVALID_TARGET_YEAR",
			"日別交通費検索条件の作成に失敗しました",
			map[string]any{"targetYear": req.TargetYear},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_QUERY_INVALID_TARGET_MONTH",
			"日別交通費検索条件の作成に失敗しました",
			map[string]any{"targetMonth": req.TargetMonth},
		)
	}

	startDate := time.Date(req.TargetYear, time.Month(req.TargetMonth), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	query := builder.db.
		Model(&models.AttendanceTransportExpense{}).
		Joins("JOIN attendance_days ON attendance_days.id = attendance_transport_expenses.attendance_day_id").
		Preload("AttendanceDay").
		Where("attendance_days.user_id = ?", userID).
		Where("attendance_days.work_date >= ?", startDate).
		Where("attendance_days.work_date < ?", endDate).
		Where("attendance_days.is_deleted = ?", false).
		Where("attendance_transport_expenses.is_deleted = ?", false).
		Order("attendance_days.work_date ASC").
		Order("attendance_transport_expenses.sort_order ASC").
		Order("attendance_transport_expenses.id ASC")

	return query, results.OK(nil, "BUILD_SEARCH_ATTENDANCE_TRANSPORT_EXPENSES_QUERY_SUCCESS", "", nil)
}

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

	return query, results.OK(nil, "BUILD_FIND_ATTENDANCE_TRANSPORT_EXPENSES_QUERY_SUCCESS", "", nil)
}

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

	return query, results.OK(nil, "BUILD_FIND_ATTENDANCE_TRANSPORT_EXPENSE_QUERY_SUCCESS", "", nil)
}

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

	model := models.AttendanceTransportExpense{
		AttendanceDayID: attendanceDayID,
		SortOrder:       sortOrder,
		TransportFrom:   req.TransportFrom,
		TransportTo:     req.TransportTo,
		TransportMethod: req.TransportMethod,
		TransportAmount: req.TransportAmount,
		TransportMemo:   req.TransportMemo,
		IsDeleted:       false,
	}

	return model, results.OK(nil, "BUILD_CREATE_ATTENDANCE_TRANSPORT_EXPENSE_MODEL_SUCCESS", "", nil)
}

func (builder *attendanceTransportExpenseBuilder) BuildUpdateAttendanceTransportExpenseModel(
	current models.AttendanceTransportExpense,
	req types.UpdateAttendanceTransportExpensesByWorkDateExpenseRequest,
	sortOrder int,
) (models.AttendanceTransportExpense, results.Result) {
	if current.ID == 0 {
		return models.AttendanceTransportExpense{}, results.BadRequest(
			"BUILD_UPDATE_ATTENDANCE_TRANSPORT_EXPENSE_MODEL_EMPTY_CURRENT_DATA",
			"日別交通費更新データの作成に失敗しました",
			nil,
		)
	}

	current.SortOrder = sortOrder
	current.TransportFrom = req.TransportFrom
	current.TransportTo = req.TransportTo
	current.TransportMethod = req.TransportMethod
	current.TransportAmount = req.TransportAmount
	current.TransportMemo = req.TransportMemo
	current.IsDeleted = false
	current.DeletedAt = nil

	return current, results.OK(nil, "BUILD_UPDATE_ATTENDANCE_TRANSPORT_EXPENSE_MODEL_SUCCESS", "", nil)
}

func (builder *attendanceTransportExpenseBuilder) BuildDeleteAttendanceTransportExpenseModel(
	current models.AttendanceTransportExpense,
) (models.AttendanceTransportExpense, results.Result) {
	if current.ID == 0 {
		return models.AttendanceTransportExpense{}, results.BadRequest(
			"BUILD_DELETE_ATTENDANCE_TRANSPORT_EXPENSE_MODEL_EMPTY_CURRENT_DATA",
			"日別交通費削除データの作成に失敗しました",
			nil,
		)
	}

	now := time.Now()
	current.IsDeleted = true
	current.DeletedAt = &now

	return current, results.OK(nil, "BUILD_DELETE_ATTENDANCE_TRANSPORT_EXPENSE_MODEL_SUCCESS", "", nil)
}
