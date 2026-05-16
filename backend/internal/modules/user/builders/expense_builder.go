package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

const ExpenseReceiptExternalStorageLinkType = "EXPENSE_RECEIPT_BOX"

type ExpenseBuilder interface {
	BuildSearchExpensesQuery(userID uint, req types.SearchExpensesRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindExpenseByIDQuery(userID uint, expenseID uint) (*gorm.DB, results.Result)
	BuildFindExpenseReceiptStorageLinkQuery() (*gorm.DB, results.Result)
	BuildCreateExpenseModel(userID uint, req types.CreateExpenseRequest) (models.Expense, results.Result)
	BuildUpdateExpenseModel(currentExpense models.Expense, req types.UpdateExpenseRequest) (models.Expense, results.Result)
	BuildApplyReceiptFileModel(currentExpense models.Expense, receipt ExpenseReceiptFileModel) (models.Expense, results.Result)
	BuildDeleteExpenseModel(currentExpense models.Expense) (models.Expense, results.Result)
}

type ExpenseReceiptFileModel struct {
	OriginalFileName      string
	StoredFileName        string
	FileURL               string
	DriveFileID           string
	ExternalStorageLinkID uint
	MimeType              string
	SizeBytes             int64
}

type expenseBuilder struct {
	db *gorm.DB
}

func NewExpenseBuilder(db *gorm.DB) ExpenseBuilder {
	return &expenseBuilder{db: db}
}

func (builder *expenseBuilder) BuildSearchExpensesQuery(userID uint, req types.SearchExpensesRequest) (*gorm.DB, *gorm.DB, results.Result) {
	if userID == 0 {
		return nil, nil, results.BadRequest("BUILD_SEARCH_EXPENSES_QUERY_INVALID_USER_ID", "経費検索条件の作成に失敗しました", map[string]any{"userId": userID})
	}

	if req.Offset < 0 {
		return nil, nil, results.BadRequest("BUILD_SEARCH_EXPENSES_QUERY_INVALID_OFFSET", "経費検索条件の作成に失敗しました", map[string]any{"offset": req.Offset})
	}

	if req.Limit <= 0 {
		return nil, nil, results.BadRequest("BUILD_SEARCH_EXPENSES_QUERY_INVALID_LIMIT", "経費検索条件の作成に失敗しました", map[string]any{"limit": req.Limit})
	}

	if req.TargetMonthFrom == "" {
		return nil, nil, results.BadRequest("BUILD_SEARCH_EXPENSES_QUERY_EMPTY_TARGET_MONTH_FROM", "経費検索条件の作成に失敗しました", nil)
	}

	if req.TargetMonthTo == "" {
		return nil, nil, results.BadRequest("BUILD_SEARCH_EXPENSES_QUERY_EMPTY_TARGET_MONTH_TO", "経費検索条件の作成に失敗しました", nil)
	}

	targetMonthFrom, parseFromResult := parseExpenseTargetMonth(req.TargetMonthFrom, "BUILD_SEARCH_EXPENSES_QUERY_INVALID_TARGET_MONTH_FROM")
	if parseFromResult.Error {
		return nil, nil, parseFromResult
	}

	targetMonthTo, parseToResult := parseExpenseTargetMonth(req.TargetMonthTo, "BUILD_SEARCH_EXPENSES_QUERY_INVALID_TARGET_MONTH_TO")
	if parseToResult.Error {
		return nil, nil, parseToResult
	}

	searchQuery := builder.db.Model(&models.Expense{}).Preload("User")
	countQuery := builder.db.Model(&models.Expense{})

	searchQuery = applySearchExpensesCondition(searchQuery, userID, targetMonthFrom, targetMonthTo)
	countQuery = applySearchExpensesCondition(countQuery, userID, targetMonthFrom, targetMonthTo)

	searchQuery = searchQuery.
		Order("expenses.target_month DESC").
		Order("expenses.expense_date DESC").
		Order("expenses.id DESC").
		Offset(req.Offset).
		Limit(req.Limit)

	return searchQuery, countQuery, results.OK(nil, "BUILD_SEARCH_EXPENSES_QUERY_SUCCESS", "", nil)
}

func (builder *expenseBuilder) BuildFindExpenseByIDQuery(userID uint, expenseID uint) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest("BUILD_FIND_EXPENSE_BY_ID_QUERY_INVALID_USER_ID", "経費取得条件の作成に失敗しました", map[string]any{"userId": userID})
	}

	if expenseID == 0 {
		return nil, results.BadRequest("BUILD_FIND_EXPENSE_BY_ID_QUERY_INVALID_EXPENSE_ID", "経費取得条件の作成に失敗しました", map[string]any{"expenseId": expenseID})
	}

	query := builder.db.
		Model(&models.Expense{}).
		Preload("User").
		Where("expenses.id = ?", expenseID).
		Where("expenses.user_id = ?", userID).
		Where("expenses.is_deleted = ?", false)

	return query, results.OK(nil, "BUILD_FIND_EXPENSE_BY_ID_QUERY_SUCCESS", "", nil)
}

func (builder *expenseBuilder) BuildFindExpenseReceiptStorageLinkQuery() (*gorm.DB, results.Result) {
	query := builder.db.
		Model(&models.ExternalStorageLink{}).
		Where("link_type = ?", ExpenseReceiptExternalStorageLinkType).
		Where("is_deleted = ?", false)

	return query, results.OK(nil, "BUILD_FIND_EXPENSE_RECEIPT_STORAGE_LINK_QUERY_SUCCESS", "", nil)
}

func (builder *expenseBuilder) BuildCreateExpenseModel(userID uint, req types.CreateExpenseRequest) (models.Expense, results.Result) {
	if userID == 0 {
		return models.Expense{}, results.BadRequest("BUILD_CREATE_EXPENSE_MODEL_INVALID_USER_ID", "経費作成データの作成に失敗しました", map[string]any{"userId": userID})
	}

	targetMonth, parseTargetMonthResult := parseExpenseTargetMonth(req.TargetMonth, "BUILD_CREATE_EXPENSE_MODEL_INVALID_TARGET_MONTH")
	if parseTargetMonthResult.Error {
		return models.Expense{}, parseTargetMonthResult
	}

	expenseDate, parseExpenseDateResult := parseExpenseDate(req.ExpenseDate, "BUILD_CREATE_EXPENSE_MODEL_INVALID_EXPENSE_DATE")
	if parseExpenseDateResult.Error {
		return models.Expense{}, parseExpenseDateResult
	}

	if req.Amount <= 0 {
		return models.Expense{}, results.BadRequest("BUILD_CREATE_EXPENSE_MODEL_INVALID_AMOUNT", "経費作成データの作成に失敗しました", map[string]any{"amount": req.Amount})
	}

	if req.Description == "" {
		return models.Expense{}, results.BadRequest("BUILD_CREATE_EXPENSE_MODEL_EMPTY_DESCRIPTION", "経費作成データの作成に失敗しました", nil)
	}

	expense := models.Expense{
		UserID:      userID,
		TargetMonth: targetMonth,
		ExpenseDate: expenseDate,
		Amount:      req.Amount,
		Description: req.Description,
		Memo:        req.Memo,
		IsDeleted:   false,
	}

	return expense, results.OK(nil, "BUILD_CREATE_EXPENSE_MODEL_SUCCESS", "", nil)
}

func (builder *expenseBuilder) BuildUpdateExpenseModel(currentExpense models.Expense, req types.UpdateExpenseRequest) (models.Expense, results.Result) {
	if currentExpense.ID == 0 {
		return models.Expense{}, results.BadRequest("BUILD_UPDATE_EXPENSE_MODEL_EMPTY_CURRENT_EXPENSE", "経費更新データの作成に失敗しました", nil)
	}

	targetMonth, parseTargetMonthResult := parseExpenseTargetMonth(req.TargetMonth, "BUILD_UPDATE_EXPENSE_MODEL_INVALID_TARGET_MONTH")
	if parseTargetMonthResult.Error {
		return models.Expense{}, parseTargetMonthResult
	}

	expenseDate, parseExpenseDateResult := parseExpenseDate(req.ExpenseDate, "BUILD_UPDATE_EXPENSE_MODEL_INVALID_EXPENSE_DATE")
	if parseExpenseDateResult.Error {
		return models.Expense{}, parseExpenseDateResult
	}

	if req.Amount <= 0 {
		return models.Expense{}, results.BadRequest("BUILD_UPDATE_EXPENSE_MODEL_INVALID_AMOUNT", "経費更新データの作成に失敗しました", map[string]any{"amount": req.Amount})
	}

	if req.Description == "" {
		return models.Expense{}, results.BadRequest("BUILD_UPDATE_EXPENSE_MODEL_EMPTY_DESCRIPTION", "経費更新データの作成に失敗しました", nil)
	}

	currentExpense.TargetMonth = targetMonth
	currentExpense.ExpenseDate = expenseDate
	currentExpense.Amount = req.Amount
	currentExpense.Description = req.Description
	currentExpense.Memo = req.Memo

	return currentExpense, results.OK(nil, "BUILD_UPDATE_EXPENSE_MODEL_SUCCESS", "", nil)
}

func (builder *expenseBuilder) BuildApplyReceiptFileModel(currentExpense models.Expense, receipt ExpenseReceiptFileModel) (models.Expense, results.Result) {
	if currentExpense.ID == 0 {
		return models.Expense{}, results.BadRequest("BUILD_APPLY_RECEIPT_FILE_MODEL_EMPTY_CURRENT_EXPENSE", "領収書ファイル情報の反映に失敗しました", nil)
	}

	if receipt.DriveFileID == "" {
		return models.Expense{}, results.BadRequest("BUILD_APPLY_RECEIPT_FILE_MODEL_EMPTY_DRIVE_FILE_ID", "領収書ファイル情報の反映に失敗しました", nil)
	}

	currentExpense.OriginalFileName = stringPointer(receipt.OriginalFileName)
	currentExpense.StoredFileName = stringPointer(receipt.StoredFileName)
	currentExpense.FileURL = stringPointer(receipt.FileURL)
	currentExpense.DriveFileID = stringPointer(receipt.DriveFileID)
	currentExpense.ExternalStorageLinkID = uintPointer(receipt.ExternalStorageLinkID)
	currentExpense.MimeType = stringPointer(receipt.MimeType)
	currentExpense.SizeBytes = int64Pointer(receipt.SizeBytes)

	return currentExpense, results.OK(nil, "BUILD_APPLY_RECEIPT_FILE_MODEL_SUCCESS", "", nil)
}

func (builder *expenseBuilder) BuildDeleteExpenseModel(currentExpense models.Expense) (models.Expense, results.Result) {
	if currentExpense.ID == 0 {
		return models.Expense{}, results.BadRequest("BUILD_DELETE_EXPENSE_MODEL_EMPTY_CURRENT_EXPENSE", "経費削除データの作成に失敗しました", nil)
	}

	now := time.Now()

	currentExpense.IsDeleted = true
	currentExpense.DeletedAt = &now

	return currentExpense, results.OK(nil, "BUILD_DELETE_EXPENSE_MODEL_SUCCESS", "", nil)
}

func applySearchExpensesCondition(query *gorm.DB, userID uint, targetMonthFrom time.Time, targetMonthTo time.Time) *gorm.DB {
	return query.
		Where("expenses.user_id = ?", userID).
		Where("expenses.is_deleted = ?", false).
		Where("expenses.target_month >= ?", targetMonthFrom).
		Where("expenses.target_month <= ?", targetMonthTo)
}

func parseExpenseTargetMonth(value string, code string) (time.Time, results.Result) {
	parsed, err := time.Parse("2006-01", value)
	if err != nil {
		return time.Time{}, results.BadRequest(code, "対象月の形式が正しくありません", map[string]any{"targetMonth": value, "expected": "YYYY-MM"})
	}

	return parsed, results.OK(nil, code+"_SUCCESS", "", nil)
}

func parseExpenseDate(value string, code string) (time.Time, results.Result) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, results.BadRequest(code, "経費発生日の形式が正しくありません", map[string]any{"expenseDate": value, "expected": "YYYY-MM-DD"})
	}

	return parsed, results.OK(nil, code+"_SUCCESS", "", nil)
}

func stringPointer(value string) *string {
	return &value
}

func uintPointer(value uint) *uint {
	return &value
}

func int64Pointer(value int64) *int64 {
	return &value
}
