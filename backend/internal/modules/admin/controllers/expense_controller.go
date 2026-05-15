package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用経費Controller
 *
 * 役割：
 * ・リクエストをRequest型に変換する
 * ・bind/parse失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * 注意：
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・c.JSONは直接使わず responses.JSON を使う
 * ・領収書ファイルを扱う create/update は multipart/form-data で受ける
 */
type ExpenseController struct {
	expenseService services.ExpenseService
}

/*
 * ExpenseController生成
 */
func NewExpenseController(expenseService services.ExpenseService) *ExpenseController {
	return &ExpenseController{
		expenseService: expenseService,
	}
}

/*
 * 検索
 *
 * POST /admin/expenses/search
 */
func (controller *ExpenseController) SearchExpenses(c *gin.Context) {
	var req types.SearchExpensesRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_EXPENSES_INVALID_REQUEST",
			"経費検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.expenseService.SearchExpenses(req)

	responses.JSON(c, result)
}

/*
 * 取得
 *
 * POST /admin/expenses/detail
 */
func (controller *ExpenseController) GetExpenseDetail(c *gin.Context) {
	var req types.ExpenseDetailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"GET_EXPENSE_DETAIL_INVALID_REQUEST",
			"経費詳細取得のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.expenseService.GetExpenseDetail(req)

	responses.JSON(c, result)
}

/*
 * 新規作成
 *
 * POST /admin/expenses/create
 *
 * Content-Type:
 * multipart/form-data
 *
 * form fields:
 * - targetUserId
 * - targetMonth
 * - expenseDate
 * - amount
 * - description
 * - memo
 * - receiptFile
 */
func (controller *ExpenseController) CreateExpense(c *gin.Context) {
	req, parseResult := buildCreateExpenseRequestFromMultipart(c)
	if parseResult.Error {
		responses.JSON(c, parseResult)
		return
	}

	result := controller.expenseService.CreateExpense(c.Request.Context(), req)

	responses.JSON(c, result)
}

/*
 * 更新
 *
 * POST /admin/expenses/update
 *
 * Content-Type:
 * multipart/form-data
 *
 * form fields:
 * - expenseId
 * - targetUserId
 * - targetMonth
 * - expenseDate
 * - amount
 * - description
 * - memo
 * - receiptFile
 */
func (controller *ExpenseController) UpdateExpense(c *gin.Context) {
	req, parseResult := buildUpdateExpenseRequestFromMultipart(c)
	if parseResult.Error {
		responses.JSON(c, parseResult)
		return
	}

	result := controller.expenseService.UpdateExpense(c.Request.Context(), req)

	responses.JSON(c, result)
}

/*
 * 論理削除
 *
 * POST /admin/expenses/delete
 */
func (controller *ExpenseController) DeleteExpense(c *gin.Context) {
	var req types.DeleteExpenseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"DELETE_EXPENSE_INVALID_REQUEST",
			"経費削除のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.expenseService.DeleteExpense(req)

	responses.JSON(c, result)
}

/*
 * 領収書表示
 *
 * POST /admin/expenses/receipt/view
 *
 * 注意：
 * ・成功時は共通JSONではなく、ファイル本体を返す
 * ・失敗時は共通JSONで返す
 */
func (controller *ExpenseController) ViewExpenseReceipt(c *gin.Context) {
	var req types.ViewExpenseReceiptRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"VIEW_EXPENSE_RECEIPT_INVALID_REQUEST",
			"経費領収書表示のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	fileResponse, result := controller.expenseService.DownloadExpenseReceipt(c.Request.Context(), req)
	if result.Error {
		responses.JSON(c, result)
		return
	}
	defer fileResponse.Body.Close()

	mimeType := strings.TrimSpace(fileResponse.MimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	c.Header("Content-Disposition", "inline; filename="+strconv.Quote(fileResponse.FileName))
	c.DataFromReader(
		http.StatusOK,
		fileResponse.SizeBytes,
		mimeType,
		fileResponse.Body,
		nil,
	)
}

/*
 * multipart/form-data からCreateExpenseRequestを作成する。
 */
func buildCreateExpenseRequestFromMultipart(c *gin.Context) (types.CreateExpenseRequest, results.Result) {
	targetUserID, targetUserIDResult := parseRequiredUintForm(c, "targetUserId", "CREATE_EXPENSE_INVALID_TARGET_USER_ID", "登録対象ユーザーIDが正しくありません")
	if targetUserIDResult.Error {
		return types.CreateExpenseRequest{}, targetUserIDResult
	}

	amount, amountResult := parseRequiredIntForm(c, "amount", "CREATE_EXPENSE_INVALID_AMOUNT", "金額が正しくありません")
	if amountResult.Error {
		return types.CreateExpenseRequest{}, amountResult
	}

	receiptFile, err := c.FormFile("receiptFile")
	if err != nil {
		receiptFile = nil
	}

	req := types.CreateExpenseRequest{
		TargetUserID: targetUserID,
		TargetMonth:  strings.TrimSpace(c.PostForm("targetMonth")),
		ExpenseDate:  strings.TrimSpace(c.PostForm("expenseDate")),
		Amount:       amount,
		Description:  strings.TrimSpace(c.PostForm("description")),
		Memo:         optionalStringPointer(c.PostForm("memo")),
		ReceiptFile:  receiptFile,
	}

	return req, results.OK(
		nil,
		"CREATE_EXPENSE_REQUEST_PARSED",
		"",
		nil,
	)
}

/*
 * multipart/form-data からUpdateExpenseRequestを作成する。
 */
func buildUpdateExpenseRequestFromMultipart(c *gin.Context) (types.UpdateExpenseRequest, results.Result) {
	expenseID, expenseIDResult := parseRequiredUintForm(c, "expenseId", "UPDATE_EXPENSE_INVALID_EXPENSE_ID", "経費IDが正しくありません")
	if expenseIDResult.Error {
		return types.UpdateExpenseRequest{}, expenseIDResult
	}

	targetUserID, targetUserIDResult := parseRequiredUintForm(c, "targetUserId", "UPDATE_EXPENSE_INVALID_TARGET_USER_ID", "登録対象ユーザーIDが正しくありません")
	if targetUserIDResult.Error {
		return types.UpdateExpenseRequest{}, targetUserIDResult
	}

	amount, amountResult := parseRequiredIntForm(c, "amount", "UPDATE_EXPENSE_INVALID_AMOUNT", "金額が正しくありません")
	if amountResult.Error {
		return types.UpdateExpenseRequest{}, amountResult
	}

	receiptFile, err := c.FormFile("receiptFile")
	if err != nil {
		receiptFile = nil
	}

	req := types.UpdateExpenseRequest{
		ExpenseID:    expenseID,
		TargetUserID: targetUserID,
		TargetMonth:  strings.TrimSpace(c.PostForm("targetMonth")),
		ExpenseDate:  strings.TrimSpace(c.PostForm("expenseDate")),
		Amount:       amount,
		Description:  strings.TrimSpace(c.PostForm("description")),
		Memo:         optionalStringPointer(c.PostForm("memo")),
		ReceiptFile:  receiptFile,
	}

	return req, results.OK(
		nil,
		"UPDATE_EXPENSE_REQUEST_PARSED",
		"",
		nil,
	)
}

/*
 * 必須uint form値を取得する。
 */
func parseRequiredUintForm(c *gin.Context, fieldName string, code string, message string) (uint, results.Result) {
	value := strings.TrimSpace(c.PostForm(fieldName))
	if value == "" {
		return 0, results.BadRequest(
			code,
			message,
			map[string]any{
				"field": fieldName,
			},
		)
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, results.BadRequest(
			code,
			message,
			map[string]any{
				"field": fieldName,
				"value": value,
			},
		)
	}

	return uint(parsed), results.OK(nil, code+"_SUCCESS", "", nil)
}

/*
 * 必須int form値を取得する。
 */
func parseRequiredIntForm(c *gin.Context, fieldName string, code string, message string) (int, results.Result) {
	value := strings.TrimSpace(c.PostForm(fieldName))
	if value == "" {
		return 0, results.BadRequest(
			code,
			message,
			map[string]any{
				"field": fieldName,
			},
		)
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, results.BadRequest(
			code,
			message,
			map[string]any{
				"field": fieldName,
				"value": value,
			},
		)
	}

	return parsed, results.OK(nil, code+"_SUCCESS", "", nil)
}

/*
 * 空文字はnil、それ以外はtrimしてポインタ化する。
 */
func optionalStringPointer(value string) *string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil
	}

	return &trimmedValue
}
