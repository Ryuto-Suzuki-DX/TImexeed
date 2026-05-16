package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type ExpenseRepository interface {
	FindExpenses(query *gorm.DB) ([]models.Expense, results.Result)
	CountExpenses(query *gorm.DB) (int64, results.Result)
	FindExpense(query *gorm.DB) (models.Expense, results.Result)
	FindExternalStorageLink(query *gorm.DB) (models.ExternalStorageLink, results.Result)
	CreateExpense(expense models.Expense) (models.Expense, results.Result)
	SaveExpense(expense models.Expense) (models.Expense, results.Result)
}

type expenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) ExpenseRepository {
	return &expenseRepository{db: db}
}

func (repository *expenseRepository) FindExpenses(query *gorm.DB) ([]models.Expense, results.Result) {
	if query == nil {
		return nil, results.InternalServerError("FIND_EXPENSES_QUERY_IS_NIL", "経費一覧の取得に失敗しました", nil)
	}

	var expenses []models.Expense

	if err := query.Find(&expenses).Error; err != nil {
		return nil, results.InternalServerError("FIND_EXPENSES_FAILED", "経費一覧の取得に失敗しました", err.Error())
	}

	return expenses, results.OK(nil, "FIND_EXPENSES_SUCCESS", "", nil)
}

func (repository *expenseRepository) CountExpenses(query *gorm.DB) (int64, results.Result) {
	if query == nil {
		return 0, results.InternalServerError("COUNT_EXPENSES_QUERY_IS_NIL", "経費件数の取得に失敗しました", nil)
	}

	var count int64

	if err := query.Count(&count).Error; err != nil {
		return 0, results.InternalServerError("COUNT_EXPENSES_FAILED", "経費件数の取得に失敗しました", err.Error())
	}

	return count, results.OK(nil, "COUNT_EXPENSES_SUCCESS", "", nil)
}

func (repository *expenseRepository) FindExpense(query *gorm.DB) (models.Expense, results.Result) {
	if query == nil {
		return models.Expense{}, results.InternalServerError("FIND_EXPENSE_QUERY_IS_NIL", "経費情報の取得に失敗しました", nil)
	}

	var expense models.Expense

	if err := query.First(&expense).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Expense{}, results.NotFound("EXPENSE_NOT_FOUND", "対象経費が見つかりません", nil)
		}

		return models.Expense{}, results.InternalServerError("FIND_EXPENSE_FAILED", "経費情報の取得に失敗しました", err.Error())
	}

	return expense, results.OK(nil, "FIND_EXPENSE_SUCCESS", "", nil)
}

func (repository *expenseRepository) FindExternalStorageLink(query *gorm.DB) (models.ExternalStorageLink, results.Result) {
	if query == nil {
		return models.ExternalStorageLink{}, results.InternalServerError("FIND_EXTERNAL_STORAGE_LINK_QUERY_IS_NIL", "外部ストレージリンクの取得に失敗しました", nil)
	}

	var externalStorageLink models.ExternalStorageLink

	if err := query.First(&externalStorageLink).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.ExternalStorageLink{}, results.NotFound("EXTERNAL_STORAGE_LINK_NOT_FOUND", "外部ストレージリンクが見つかりません", nil)
		}

		return models.ExternalStorageLink{}, results.InternalServerError("FIND_EXTERNAL_STORAGE_LINK_FAILED", "外部ストレージリンクの取得に失敗しました", err.Error())
	}

	return externalStorageLink, results.OK(nil, "FIND_EXTERNAL_STORAGE_LINK_SUCCESS", "", nil)
}

func (repository *expenseRepository) CreateExpense(expense models.Expense) (models.Expense, results.Result) {
	if err := repository.db.Create(&expense).Error; err != nil {
		return models.Expense{}, results.InternalServerError("CREATE_EXPENSE_FAILED", "経費の作成に失敗しました", err.Error())
	}

	return expense, results.OK(nil, "CREATE_EXPENSE_SUCCESS", "", nil)
}

func (repository *expenseRepository) SaveExpense(expense models.Expense) (models.Expense, results.Result) {
	if expense.ID == 0 {
		return models.Expense{}, results.InternalServerError("SAVE_EXPENSE_EMPTY_ID", "経費情報の保存に失敗しました", nil)
	}

	if err := repository.db.Save(&expense).Error; err != nil {
		return models.Expense{}, results.InternalServerError("SAVE_EXPENSE_FAILED", "経費情報の保存に失敗しました", err.Error())
	}

	return expense, results.OK(nil, "SAVE_EXPENSE_SUCCESS", "", nil)
}
