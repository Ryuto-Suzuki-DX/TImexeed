package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用経費Repository interface
 */
type ExpenseRepository interface {
	FindExpenses(query *gorm.DB) ([]models.Expense, results.Result)
	CountExpenses(query *gorm.DB) (int64, results.Result)
	FindExpense(query *gorm.DB) (models.Expense, results.Result)
	CreateExpense(expense models.Expense) (models.Expense, results.Result)
	SaveExpense(expense models.Expense) (models.Expense, results.Result)
}

/*
 * 管理者用経費Repository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・DBへのCreate / Saveを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 */
type expenseRepository struct {
	db *gorm.DB
}

/*
 * ExpenseRepository生成
 */
func NewExpenseRepository(db *gorm.DB) ExpenseRepository {
	return &expenseRepository{
		db: db,
	}
}

/*
 * 経費一覧取得
 */
func (repository *expenseRepository) FindExpenses(query *gorm.DB) ([]models.Expense, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_EXPENSES_QUERY_IS_NIL",
			"経費一覧の取得に失敗しました",
			nil,
		)
	}

	var expenses []models.Expense

	if err := query.Find(&expenses).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_EXPENSES_FAILED",
			"経費一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return expenses, results.OK(
		nil,
		"FIND_EXPENSES_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費件数取得
 */
func (repository *expenseRepository) CountExpenses(query *gorm.DB) (int64, results.Result) {
	if query == nil {
		return 0, results.InternalServerError(
			"COUNT_EXPENSES_QUERY_IS_NIL",
			"経費件数の取得に失敗しました",
			nil,
		)
	}

	var count int64

	if err := query.Count(&count).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_EXPENSES_FAILED",
			"経費件数の取得に失敗しました",
			err.Error(),
		)
	}

	return count, results.OK(
		nil,
		"COUNT_EXPENSES_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費1件取得
 */
func (repository *expenseRepository) FindExpense(query *gorm.DB) (models.Expense, results.Result) {
	if query == nil {
		return models.Expense{}, results.InternalServerError(
			"FIND_EXPENSE_QUERY_IS_NIL",
			"経費情報の取得に失敗しました",
			nil,
		)
	}

	var expense models.Expense

	if err := query.First(&expense).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Expense{}, results.NotFound(
				"EXPENSE_NOT_FOUND",
				"対象経費が見つかりません",
				nil,
			)
		}

		return models.Expense{}, results.InternalServerError(
			"FIND_EXPENSE_FAILED",
			"経費情報の取得に失敗しました",
			err.Error(),
		)
	}

	return expense, results.OK(
		nil,
		"FIND_EXPENSE_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費作成
 */
func (repository *expenseRepository) CreateExpense(expense models.Expense) (models.Expense, results.Result) {
	if err := repository.db.Create(&expense).Error; err != nil {
		return models.Expense{}, results.InternalServerError(
			"CREATE_EXPENSE_FAILED",
			"経費の作成に失敗しました",
			err.Error(),
		)
	}

	return expense, results.OK(
		nil,
		"CREATE_EXPENSE_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費保存
 *
 * 更新・論理削除で使う。
 */
func (repository *expenseRepository) SaveExpense(expense models.Expense) (models.Expense, results.Result) {
	if expense.ID == 0 {
		return models.Expense{}, results.InternalServerError(
			"SAVE_EXPENSE_EMPTY_ID",
			"経費情報の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&expense).Error; err != nil {
		return models.Expense{}, results.InternalServerError(
			"SAVE_EXPENSE_FAILED",
			"経費情報の保存に失敗しました",
			err.Error(),
		)
	}

	return expense, results.OK(
		nil,
		"SAVE_EXPENSE_SUCCESS",
		"",
		nil,
	)
}
