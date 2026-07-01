package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type AttendanceTransportExpenseRepository interface {
	FindAttendanceTransportExpenses(query *gorm.DB) ([]models.AttendanceTransportExpense, results.Result)
	FindAttendanceTransportExpense(query *gorm.DB) (models.AttendanceTransportExpense, results.Result)
	CreateAttendanceTransportExpense(
		attendanceTransportExpense models.AttendanceTransportExpense,
	) (models.AttendanceTransportExpense, results.Result)
	SaveAttendanceTransportExpense(
		attendanceTransportExpense models.AttendanceTransportExpense,
	) (models.AttendanceTransportExpense, results.Result)
}

/*
 * 従業員用日別交通費Repository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・DBへのCreate / Saveを実行する
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・月次申請状態による編集可否はServiceで判定する
 */
type attendanceTransportExpenseRepository struct {
	db *gorm.DB
}

/*
 * AttendanceTransportExpenseRepository生成
 */
func NewAttendanceTransportExpenseRepository(
	db *gorm.DB,
) AttendanceTransportExpenseRepository {
	return &attendanceTransportExpenseRepository{db: db}
}

/*
 * 日別交通費一覧取得
 */
func (repository *attendanceTransportExpenseRepository) FindAttendanceTransportExpenses(
	query *gorm.DB,
) ([]models.AttendanceTransportExpense, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_ATTENDANCE_TRANSPORT_EXPENSES_QUERY_IS_NIL",
			"日別交通費一覧の取得に失敗しました",
			nil,
		)
	}

	var attendanceTransportExpenses []models.AttendanceTransportExpense

	if err := query.Find(&attendanceTransportExpenses).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_ATTENDANCE_TRANSPORT_EXPENSES_FAILED",
			"日別交通費一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return attendanceTransportExpenses, results.OK(
		nil,
		"FIND_ATTENDANCE_TRANSPORT_EXPENSES_SUCCESS",
		"",
		nil,
	)
}

/*
 * 日別交通費1件取得
 */
func (repository *attendanceTransportExpenseRepository) FindAttendanceTransportExpense(
	query *gorm.DB,
) (models.AttendanceTransportExpense, results.Result) {
	if query == nil {
		return models.AttendanceTransportExpense{}, results.InternalServerError(
			"FIND_ATTENDANCE_TRANSPORT_EXPENSE_QUERY_IS_NIL",
			"日別交通費の取得に失敗しました",
			nil,
		)
	}

	var attendanceTransportExpense models.AttendanceTransportExpense

	if err := query.First(&attendanceTransportExpense).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AttendanceTransportExpense{}, results.NotFound(
				"ATTENDANCE_TRANSPORT_EXPENSE_NOT_FOUND",
				"対象の日別交通費が見つかりません",
				nil,
			)
		}

		return models.AttendanceTransportExpense{}, results.InternalServerError(
			"FIND_ATTENDANCE_TRANSPORT_EXPENSE_FAILED",
			"日別交通費の取得に失敗しました",
			err.Error(),
		)
	}

	return attendanceTransportExpense, results.OK(
		nil,
		"FIND_ATTENDANCE_TRANSPORT_EXPENSE_SUCCESS",
		"",
		nil,
	)
}

/*
 * 日別交通費作成
 */
func (repository *attendanceTransportExpenseRepository) CreateAttendanceTransportExpense(
	attendanceTransportExpense models.AttendanceTransportExpense,
) (models.AttendanceTransportExpense, results.Result) {
	if err := repository.db.Create(&attendanceTransportExpense).Error; err != nil {
		return models.AttendanceTransportExpense{}, results.InternalServerError(
			"CREATE_ATTENDANCE_TRANSPORT_EXPENSE_FAILED",
			"日別交通費の作成に失敗しました",
			err.Error(),
		)
	}

	return attendanceTransportExpense, results.OK(
		nil,
		"CREATE_ATTENDANCE_TRANSPORT_EXPENSE_SUCCESS",
		"",
		nil,
	)
}

/*
 * 日別交通費保存
 *
 * 更新・論理削除で使う。
 */
func (repository *attendanceTransportExpenseRepository) SaveAttendanceTransportExpense(
	attendanceTransportExpense models.AttendanceTransportExpense,
) (models.AttendanceTransportExpense, results.Result) {
	if attendanceTransportExpense.ID == 0 {
		return models.AttendanceTransportExpense{}, results.InternalServerError(
			"SAVE_ATTENDANCE_TRANSPORT_EXPENSE_EMPTY_ID",
			"日別交通費の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&attendanceTransportExpense).Error; err != nil {
		return models.AttendanceTransportExpense{}, results.InternalServerError(
			"SAVE_ATTENDANCE_TRANSPORT_EXPENSE_FAILED",
			"日別交通費の保存に失敗しました",
			err.Error(),
		)
	}

	return attendanceTransportExpense, results.OK(
		nil,
		"SAVE_ATTENDANCE_TRANSPORT_EXPENSE_SUCCESS",
		"",
		nil,
	)
}
