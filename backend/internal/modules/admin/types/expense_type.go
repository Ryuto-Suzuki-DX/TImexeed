package types

import (
	"io"
	"mime/multipart"
	"time"
)

/*
 * 〇 管理者 経費 Type
 *
 * 注意：
 * ・経費は申請ではなく、上長確認済みの経費登録として扱う
 * ・承認/否認ステータスは持たない
 * ・対象月は year / month に分けず、"2026-05" のような文字列で受ける
 * ・DBには月初日 date として保存する
 * ・領収書ファイルは multipart/form-data の receiptFile で受ける
 */

/*
 * 経費検索Request
 *
 * 管理者画面では、対象月の期間検索を必須にする。
 * ユーザー検索は keyword で name / email などをフリーワード検索する想定。
 */
type SearchExpensesRequest struct {
	Keyword string `json:"keyword"`

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

/*
 * 経費詳細Request
 */
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

/*
 * 経費作成Request
 *
 * multipart/form-data からControllerで組み立てる。
 */
type CreateExpenseRequest struct {
	TargetUserID uint

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

/*
 * 経費更新Request
 *
 * ReceiptFile が nil の場合は、既存の領収書情報を維持する。
 * ReceiptFile がある場合は、Google Driveへ新規アップロードし、DB上の領収書情報を差し替える。
 */
type UpdateExpenseRequest struct {
	ExpenseID uint

	TargetUserID uint

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

/*
 * 経費削除Request
 */
type DeleteExpenseRequest struct {
	ExpenseID uint `json:"expenseId"`
}

type DeleteExpenseResponse struct {
	ExpenseID uint `json:"expenseId"`
}

/*
 * 領収書表示Request
 *
 * 経費1件に領収書1つの設計なので、expenseId で受ける。
 * URLにIDを載せないTimexeed方針に合わせてPOSTで受ける。
 */
type ViewExpenseReceiptRequest struct {
	ExpenseID uint `json:"expenseId"`
}

/*
 * 領収書ファイルResponse
 *
 * Controllerで DataFromReader するための内部受け渡し用。
 * JSONレスポンス用ではない。
 */
type ExpenseReceiptFileResponse struct {
	Body      io.ReadCloser
	FileName  string
	MimeType  string
	SizeBytes int64
}
