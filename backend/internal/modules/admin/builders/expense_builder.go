package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

const ExpenseReceiptExternalStorageLinkType = "EXPENSE_RECEIPT_BOX"

/*
 * 管理者用経費Builder interface
 */
type ExpenseBuilder interface {
	BuildSearchExpensesQuery(req types.SearchExpensesRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindExpenseByIDQuery(expenseID uint) (*gorm.DB, results.Result)
	BuildFindExpenseReceiptStorageLinkQuery() (*gorm.DB, results.Result)
	BuildCreateExpenseModel(req types.CreateExpenseRequest) (models.Expense, results.Result)
	BuildUpdateExpenseModel(currentExpense models.Expense, req types.UpdateExpenseRequest) (models.Expense, results.Result)
	BuildApplyReceiptFileModel(currentExpense models.Expense, receipt ExpenseReceiptFileModel) (models.Expense, results.Result)
	BuildDeleteExpenseModel(currentExpense models.Expense) (models.Expense, results.Result)
}

/*
 * 領収書ファイル情報をExpenseへ反映するための内部Model
 */
type ExpenseReceiptFileModel struct {
	OriginalFileName      string
	StoredFileName        string
	FileURL               string
	DriveFileID           string
	ExternalStorageLinkID uint
	MimeType              string
	SizeBytes             int64
}

/*
 * 管理者用経費Builder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Serviceから受け取ったRequestをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Count / Create / Save はRepositoryに任せる
 */
type expenseBuilder struct {
	db *gorm.DB
}

/*
 * ExpenseBuilder生成
 */
func NewExpenseBuilder(db *gorm.DB) ExpenseBuilder {
	return &expenseBuilder{
		db: db,
	}
}

/*
 * 経費検索用クエリ作成
 *
 * searchQuery：
 * ・一覧取得用
 * ・offset / limit / order を含む
 *
 * countQuery：
 * ・総件数取得用
 * ・offset / limit は含めない
 */
func (builder *expenseBuilder) BuildSearchExpensesQuery(req types.SearchExpensesRequest) (*gorm.DB, *gorm.DB, results.Result) {
	if req.Offset < 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_EXPENSES_QUERY_INVALID_OFFSET",
			"経費検索条件の作成に失敗しました",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	if req.Limit <= 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_EXPENSES_QUERY_INVALID_LIMIT",
			"経費検索条件の作成に失敗しました",
			map[string]any{
				"limit": req.Limit,
			},
		)
	}

	if req.TargetMonthFrom == "" {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_EXPENSES_QUERY_EMPTY_TARGET_MONTH_FROM",
			"経費検索条件の作成に失敗しました",
			nil,
		)
	}

	if req.TargetMonthTo == "" {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_EXPENSES_QUERY_EMPTY_TARGET_MONTH_TO",
			"経費検索条件の作成に失敗しました",
			nil,
		)
	}

	targetMonthFrom, parseFromResult := parseExpenseTargetMonth(req.TargetMonthFrom, "BUILD_SEARCH_EXPENSES_QUERY_INVALID_TARGET_MONTH_FROM")
	if parseFromResult.Error {
		return nil, nil, parseFromResult
	}

	targetMonthTo, parseToResult := parseExpenseTargetMonth(req.TargetMonthTo, "BUILD_SEARCH_EXPENSES_QUERY_INVALID_TARGET_MONTH_TO")
	if parseToResult.Error {
		return nil, nil, parseToResult
	}

	searchQuery := builder.db.
		Model(&models.Expense{}).
		Preload("User").
		Joins("JOIN users ON users.id = expenses.user_id")

	countQuery := builder.db.
		Model(&models.Expense{}).
		Joins("JOIN users ON users.id = expenses.user_id")

	searchQuery = applySearchExpensesCondition(searchQuery, req, targetMonthFrom, targetMonthTo)
	countQuery = applySearchExpensesCondition(countQuery, req, targetMonthFrom, targetMonthTo)

	searchQuery = searchQuery.
		Order("expenses.target_month DESC").
		Order("expenses.expense_date DESC").
		Order("expenses.id DESC").
		Offset(req.Offset).
		Limit(req.Limit)

	return searchQuery, countQuery, results.OK(
		nil,
		"BUILD_SEARCH_EXPENSES_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費ID検索用クエリ作成
 *
 * 論理削除済み経費は対象外にする。
 */
func (builder *expenseBuilder) BuildFindExpenseByIDQuery(expenseID uint) (*gorm.DB, results.Result) {
	if expenseID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_EXPENSE_BY_ID_QUERY_INVALID_EXPENSE_ID",
			"経費取得条件の作成に失敗しました",
			map[string]any{
				"expenseId": expenseID,
			},
		)
	}

	query := builder.db.
		Model(&models.Expense{}).
		Preload("User").
		Where("expenses.id = ?", expenseID).
		Where("expenses.is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_EXPENSE_BY_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費領収書の保存先リンク取得Query作成
 */
func (builder *expenseBuilder) BuildFindExpenseReceiptStorageLinkQuery() (*gorm.DB, results.Result) {
	query := builder.db.
		Model(&models.ExternalStorageLink{}).
		Where("link_type = ?", ExpenseReceiptExternalStorageLinkType).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_EXPENSE_RECEIPT_STORAGE_LINK_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費作成用Model作成
 */
func (builder *expenseBuilder) BuildCreateExpenseModel(req types.CreateExpenseRequest) (models.Expense, results.Result) {
	if req.TargetUserID == 0 {
		return models.Expense{}, results.BadRequest(
			"BUILD_CREATE_EXPENSE_MODEL_INVALID_TARGET_USER_ID",
			"経費作成データの作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
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
		return models.Expense{}, results.BadRequest(
			"BUILD_CREATE_EXPENSE_MODEL_INVALID_AMOUNT",
			"経費作成データの作成に失敗しました",
			map[string]any{
				"amount": req.Amount,
			},
		)
	}

	if req.Description == "" {
		return models.Expense{}, results.BadRequest(
			"BUILD_CREATE_EXPENSE_MODEL_EMPTY_DESCRIPTION",
			"経費作成データの作成に失敗しました",
			nil,
		)
	}

	expense := models.Expense{
		UserID:      req.TargetUserID,
		TargetMonth: targetMonth,
		ExpenseDate: expenseDate,
		Amount:      req.Amount,
		Description: req.Description,
		Memo:        req.Memo,
		IsDeleted:   false,
	}

	return expense, results.OK(
		nil,
		"BUILD_CREATE_EXPENSE_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費更新用Model作成
 */
func (builder *expenseBuilder) BuildUpdateExpenseModel(
	currentExpense models.Expense,
	req types.UpdateExpenseRequest,
) (models.Expense, results.Result) {
	if currentExpense.ID == 0 {
		return models.Expense{}, results.BadRequest(
			"BUILD_UPDATE_EXPENSE_MODEL_EMPTY_CURRENT_EXPENSE",
			"経費更新データの作成に失敗しました",
			nil,
		)
	}

	if req.TargetUserID == 0 {
		return models.Expense{}, results.BadRequest(
			"BUILD_UPDATE_EXPENSE_MODEL_INVALID_TARGET_USER_ID",
			"経費更新データの作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
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
		return models.Expense{}, results.BadRequest(
			"BUILD_UPDATE_EXPENSE_MODEL_INVALID_AMOUNT",
			"経費更新データの作成に失敗しました",
			map[string]any{
				"amount": req.Amount,
			},
		)
	}

	if req.Description == "" {
		return models.Expense{}, results.BadRequest(
			"BUILD_UPDATE_EXPENSE_MODEL_EMPTY_DESCRIPTION",
			"経費更新データの作成に失敗しました",
			nil,
		)
	}

	currentExpense.UserID = req.TargetUserID
	currentExpense.TargetMonth = targetMonth
	currentExpense.ExpenseDate = expenseDate
	currentExpense.Amount = req.Amount
	currentExpense.Description = req.Description
	currentExpense.Memo = req.Memo

	return currentExpense, results.OK(
		nil,
		"BUILD_UPDATE_EXPENSE_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 領収書ファイル情報をExpenseへ反映する
 */
func (builder *expenseBuilder) BuildApplyReceiptFileModel(currentExpense models.Expense, receipt ExpenseReceiptFileModel) (models.Expense, results.Result) {
	if currentExpense.ID == 0 {
		return models.Expense{}, results.BadRequest(
			"BUILD_APPLY_RECEIPT_FILE_MODEL_EMPTY_CURRENT_EXPENSE",
			"領収書ファイル情報の反映に失敗しました",
			nil,
		)
	}

	if receipt.DriveFileID == "" {
		return models.Expense{}, results.BadRequest(
			"BUILD_APPLY_RECEIPT_FILE_MODEL_EMPTY_DRIVE_FILE_ID",
			"領収書ファイル情報の反映に失敗しました",
			nil,
		)
	}

	currentExpense.OriginalFileName = stringPointer(receipt.OriginalFileName)
	currentExpense.StoredFileName = stringPointer(receipt.StoredFileName)
	currentExpense.FileURL = stringPointer(receipt.FileURL)
	currentExpense.DriveFileID = stringPointer(receipt.DriveFileID)
	currentExpense.ExternalStorageLinkID = uintPointer(receipt.ExternalStorageLinkID)
	currentExpense.MimeType = stringPointer(receipt.MimeType)
	currentExpense.SizeBytes = int64Pointer(receipt.SizeBytes)

	return currentExpense, results.OK(
		nil,
		"BUILD_APPLY_RECEIPT_FILE_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費論理削除用Model作成
 */
func (builder *expenseBuilder) BuildDeleteExpenseModel(currentExpense models.Expense) (models.Expense, results.Result) {
	if currentExpense.ID == 0 {
		return models.Expense{}, results.BadRequest(
			"BUILD_DELETE_EXPENSE_MODEL_EMPTY_CURRENT_EXPENSE",
			"経費削除データの作成に失敗しました",
			nil,
		)
	}

	now := time.Now()

	currentExpense.IsDeleted = true
	currentExpense.DeletedAt = &now

	return currentExpense, results.OK(
		nil,
		"BUILD_DELETE_EXPENSE_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費検索条件をGORMクエリへ適用する
 */
func applySearchExpensesCondition(query *gorm.DB, req types.SearchExpensesRequest, targetMonthFrom time.Time, targetMonthTo time.Time) *gorm.DB {
	query = query.
		Where("expenses.is_deleted = ?", false).
		Where("users.is_deleted = ?", false).
		Where("expenses.target_month >= ?", targetMonthFrom).
		Where("expenses.target_month <= ?", targetMonthTo)

	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where(
			"(users.name ILIKE ? OR users.email ILIKE ?)",
			keyword,
			keyword,
		)
	}

	return query
}

/*
 * 対象月文字列を月初日に変換する
 *
 * 入力例：2026-05
 * DB保存値：2026-05-01
 */
func parseExpenseTargetMonth(value string, code string) (time.Time, results.Result) {
	parsed, err := time.Parse("2006-01", value)
	if err != nil {
		return time.Time{}, results.BadRequest(
			code,
			"対象月の形式が正しくありません",
			map[string]any{
				"targetMonth": value,
				"expected":    "YYYY-MM",
			},
		)
	}

	return parsed, results.OK(
		nil,
		code+"_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費発生日文字列をdateに変換する
 *
 * 入力例：2026-05-16
 */
func parseExpenseDate(value string, code string) (time.Time, results.Result) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, results.BadRequest(
			code,
			"経費発生日の形式が正しくありません",
			map[string]any{
				"expenseDate": value,
				"expected":    "YYYY-MM-DD",
			},
		)
	}

	return parsed, results.OK(
		nil,
		code+"_SUCCESS",
		"",
		nil,
	)
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
