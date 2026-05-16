package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用経費Controller
 *
 * 注意：
 * ・従業員APIでは userId / targetUserId を request body で受け取らない
 * ・AuthMiddlewareでJWTから取得したログイン中ユーザーIDを使う
 * ・検索、詳細、更新、削除、領収書表示はログイン中ユーザー本人の経費だけを対象にする
 */
type ExpenseController struct {
	expenseService services.ExpenseService
}

func NewExpenseController(expenseService services.ExpenseService) *ExpenseController {
	return &ExpenseController{
		expenseService: expenseService,
	}
}

/*
 * POST /user/expenses/search
 */
func (controller *ExpenseController) SearchExpenses(c *gin.Context) {
	loginUserID, userIDResult := getLoginUserIDForExpense(c, "SEARCH_EXPENSES")
	if userIDResult.Error {
		responses.JSON(c, userIDResult)
		return
	}

	var req types.SearchExpensesRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_EXPENSES_INVALID_REQUEST",
			"経費検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.expenseService.SearchExpenses(loginUserID, req)

	responses.JSON(c, result)
}

/*
 * POST /user/expenses/detail
 */
func (controller *ExpenseController) GetExpenseDetail(c *gin.Context) {
	loginUserID, userIDResult := getLoginUserIDForExpense(c, "GET_EXPENSE_DETAIL")
	if userIDResult.Error {
		responses.JSON(c, userIDResult)
		return
	}

	var req types.ExpenseDetailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"GET_EXPENSE_DETAIL_INVALID_REQUEST",
			"経費詳細取得のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.expenseService.GetExpenseDetail(loginUserID, req)

	responses.JSON(c, result)
}

/*
 * POST /user/expenses/create
 */
func (controller *ExpenseController) CreateExpense(c *gin.Context) {
	loginUserID, userIDResult := getLoginUserIDForExpense(c, "CREATE_EXPENSE")
	if userIDResult.Error {
		responses.JSON(c, userIDResult)
		return
	}

	req, parseResult := buildCreateExpenseRequestFromMultipart(c)
	if parseResult.Error {
		responses.JSON(c, parseResult)
		return
	}

	result := controller.expenseService.CreateExpense(c.Request.Context(), loginUserID, req)

	responses.JSON(c, result)
}

/*
 * POST /user/expenses/update
 */
func (controller *ExpenseController) UpdateExpense(c *gin.Context) {
	loginUserID, userIDResult := getLoginUserIDForExpense(c, "UPDATE_EXPENSE")
	if userIDResult.Error {
		responses.JSON(c, userIDResult)
		return
	}

	req, parseResult := buildUpdateExpenseRequestFromMultipart(c)
	if parseResult.Error {
		responses.JSON(c, parseResult)
		return
	}

	result := controller.expenseService.UpdateExpense(c.Request.Context(), loginUserID, req)

	responses.JSON(c, result)
}

/*
 * POST /user/expenses/delete
 */
func (controller *ExpenseController) DeleteExpense(c *gin.Context) {
	loginUserID, userIDResult := getLoginUserIDForExpense(c, "DELETE_EXPENSE")
	if userIDResult.Error {
		responses.JSON(c, userIDResult)
		return
	}

	var req types.DeleteExpenseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"DELETE_EXPENSE_INVALID_REQUEST",
			"経費削除のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	result := controller.expenseService.DeleteExpense(loginUserID, req)

	responses.JSON(c, result)
}

/*
 * POST /user/expenses/receipt/view
 */
func (controller *ExpenseController) ViewExpenseReceipt(c *gin.Context) {
	loginUserID, userIDResult := getLoginUserIDForExpense(c, "VIEW_EXPENSE_RECEIPT")
	if userIDResult.Error {
		responses.JSON(c, userIDResult)
		return
	}

	var req types.ViewExpenseReceiptRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"VIEW_EXPENSE_RECEIPT_INVALID_REQUEST",
			"経費領収書表示のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	fileResponse, result := controller.expenseService.DownloadExpenseReceipt(c.Request.Context(), loginUserID, req)
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

func getLoginUserIDForExpense(c *gin.Context, actionCode string) (uint, results.Result) {
	userIDValue, exists := c.Get("userId")
	if !exists {
		return 0, results.Unauthorized(
			actionCode+"_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		)
	}

	loginUserID, ok := userIDValue.(uint)
	if !ok || loginUserID == 0 {
		return 0, results.Unauthorized(
			actionCode+"_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	return loginUserID, results.OK(nil, actionCode+"_USER_ID_FOUND", "", nil)
}

func buildCreateExpenseRequestFromMultipart(c *gin.Context) (types.CreateExpenseRequest, results.Result) {
	amount, amountResult := parseRequiredIntForm(c, "amount", "CREATE_EXPENSE_INVALID_AMOUNT", "金額が正しくありません")
	if amountResult.Error {
		return types.CreateExpenseRequest{}, amountResult
	}

	receiptFile, err := c.FormFile("receiptFile")
	if err != nil {
		receiptFile = nil
	}

	req := types.CreateExpenseRequest{
		TargetMonth: strings.TrimSpace(c.PostForm("targetMonth")),
		ExpenseDate: strings.TrimSpace(c.PostForm("expenseDate")),
		Amount:      amount,
		Description: strings.TrimSpace(c.PostForm("description")),
		Memo:        optionalStringPointer(c.PostForm("memo")),
		ReceiptFile: receiptFile,
	}

	return req, results.OK(nil, "CREATE_EXPENSE_REQUEST_PARSED", "", nil)
}

func buildUpdateExpenseRequestFromMultipart(c *gin.Context) (types.UpdateExpenseRequest, results.Result) {
	expenseID, expenseIDResult := parseRequiredUintForm(c, "expenseId", "UPDATE_EXPENSE_INVALID_EXPENSE_ID", "経費IDが正しくありません")
	if expenseIDResult.Error {
		return types.UpdateExpenseRequest{}, expenseIDResult
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
		ExpenseID:   expenseID,
		TargetMonth: strings.TrimSpace(c.PostForm("targetMonth")),
		ExpenseDate: strings.TrimSpace(c.PostForm("expenseDate")),
		Amount:      amount,
		Description: strings.TrimSpace(c.PostForm("description")),
		Memo:        optionalStringPointer(c.PostForm("memo")),
		ReceiptFile: receiptFile,
	}

	return req, results.OK(nil, "UPDATE_EXPENSE_REQUEST_PARSED", "", nil)
}

func parseRequiredUintForm(c *gin.Context, fieldName string, code string, message string) (uint, results.Result) {
	value := strings.TrimSpace(c.PostForm(fieldName))
	if value == "" {
		return 0, results.BadRequest(code, message, map[string]any{"field": fieldName})
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, results.BadRequest(code, message, map[string]any{"field": fieldName, "value": value})
	}

	return uint(parsed), results.OK(nil, code+"_SUCCESS", "", nil)
}

func parseRequiredIntForm(c *gin.Context, fieldName string, code string, message string) (int, results.Result) {
	value := strings.TrimSpace(c.PostForm(fieldName))
	if value == "" {
		return 0, results.BadRequest(code, message, map[string]any{"field": fieldName})
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, results.BadRequest(code, message, map[string]any{"field": fieldName, "value": value})
	}

	return parsed, results.OK(nil, code+"_SUCCESS", "", nil)
}

func optionalStringPointer(value string) *string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil
	}

	return &trimmedValue
}
