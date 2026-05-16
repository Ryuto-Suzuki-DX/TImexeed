package types

import (
	"io"
	"mime/multipart"
	"time"
)

/*
 * 〇 従業員 経費 Type
 *
 * 注意：
 * ・従業員APIでは userId / targetUserId をRequestで受け取らない
 * ・ログイン中ユーザーIDはAuthMiddleware由来のuserIdをControllerで取得する
 * ・検索、詳細、更新、削除、領収書表示は本人の経費だけを対象にする
 */
type SearchExpensesRequest struct {
	TargetMonthFrom string `json:"targetMonthFrom"`
	TargetMonthTo   string `json:"targetMonthTo"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type SearchExpensesResponse struct {
	Expenses []ExpenseListItemResponse `json:"expenses"`
	Total    int64                     `json:"total"`
	Offset   int                       `json:"offset"`
	Limit    int                       `json:"limit"`
	HasMore  bool                      `json:"hasMore"`
}

type ExpenseListItemResponse struct {
	ID uint `json:"id"`

	UserID   uint   `json:"userId"`
	UserName string `json:"userName"`
	Email    string `json:"email"`

	TargetMonth string `json:"targetMonth"`
	ExpenseDate string `json:"expenseDate"`

	Amount      int     `json:"amount"`
	Description string  `json:"description"`
	Memo        *string `json:"memo"`

	HasReceiptFile   bool    `json:"hasReceiptFile"`
	OriginalFileName *string `json:"originalFileName"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ExpenseDetailRequest struct {
	ExpenseID uint `json:"expenseId"`
}

type ExpenseDetailResponse struct {
	Expense ExpenseResponse `json:"expense"`
}

type ExpenseResponse struct {
	ID uint `json:"id"`

	UserID   uint   `json:"userId"`
	UserName string `json:"userName"`
	Email    string `json:"email"`

	TargetMonth string `json:"targetMonth"`
	ExpenseDate string `json:"expenseDate"`

	Amount      int     `json:"amount"`
	Description string  `json:"description"`
	Memo        *string `json:"memo"`

	HasReceiptFile   bool    `json:"hasReceiptFile"`
	OriginalFileName *string `json:"originalFileName"`
	StoredFileName   *string `json:"storedFileName"`
	MimeType         *string `json:"mimeType"`
	SizeBytes        *int64  `json:"sizeBytes"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

type CreateExpenseRequest struct {
	TargetMonth string
	ExpenseDate string

	Amount      int
	Description string
	Memo        *string

	ReceiptFile *multipart.FileHeader
}

type CreateExpenseResponse struct {
	Expense ExpenseResponse `json:"expense"`
}

type UpdateExpenseRequest struct {
	ExpenseID uint

	TargetMonth string
	ExpenseDate string

	Amount      int
	Description string
	Memo        *string

	ReceiptFile *multipart.FileHeader
}

type UpdateExpenseResponse struct {
	Expense ExpenseResponse `json:"expense"`
}

type DeleteExpenseRequest struct {
	ExpenseID uint `json:"expenseId"`
}

type DeleteExpenseResponse struct {
	ExpenseID uint `json:"expenseId"`
}

type ViewExpenseReceiptRequest struct {
	ExpenseID uint `json:"expenseId"`
}

type ExpenseReceiptFileResponse struct {
	Body      io.ReadCloser
	FileName  string
	MimeType  string
	SizeBytes int64
}
