/*
 * 管理者 経費 Type
 *
 * 注意：
 * ・経費は申請ではなく、上長確認済みの経費登録として扱う
 * ・承認/否認ステータスは持たない
 * ・対象月は year / month に分けず、"2026-05" のような文字列で扱う
 * ・DBには月初日 date として保存する
 * ・画像本体は一覧検索では取得しない
 * ・領収書はクリック時に /admin/expenses/receipt/view で取得する
 * ・create/update は multipart/form-data で送信する
 */

/*
 * 共通APIレスポンス
 */
export type ApiResponse<TData> = {
  data: TData;
  error: boolean;
  code: string;
  message: string;
  details?: unknown;
};

/*
 * 経費検索Request
 *
 * 管理者画面では、対象月の期間検索を必須にする。
 * ユーザー検索は keyword で name / email などをフリーワード検索する想定。
 */
export type SearchExpensesRequest = {
  keyword: string;
  targetMonthFrom: string;
  targetMonthTo: string;
  offset: number;
  limit: number;
};

export type SearchExpensesResponse = {
  expenses: ExpenseListItemResponse[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};

export type ExpenseListItemResponse = {
  id: number;

  userId: number;
  userName: string;
  email: string;

  targetMonth: string;
  expenseDate: string;

  amount: number;
  description: string;
  memo: string | null;

  hasReceiptFile: boolean;
  originalFileName: string | null;

  createdAt: string;
  updatedAt: string;
};

/*
 * 経費詳細Request
 */
export type ExpenseDetailRequest = {
  expenseId: number;
};

export type ExpenseDetailResponse = {
  expense: ExpenseResponse;
};

export type ExpenseResponse = {
  id: number;

  userId: number;
  userName: string;
  email: string;

  targetMonth: string;
  expenseDate: string;

  amount: number;
  description: string;
  memo: string | null;

  hasReceiptFile: boolean;
  originalFileName: string | null;
  storedFileName: string | null;
  mimeType: string | null;
  sizeBytes: number | null;

  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

/*
 * 経費作成Request
 *
 * バックエンドは multipart/form-data で受け取る。
 * API関数側で FormData に変換する。
 */
export type CreateExpenseRequest = {
  targetUserId: number;

  targetMonth: string;
  expenseDate: string;

  amount: number;
  description: string;
  memo: string | null;

  receiptFile?: File | null;
};

export type CreateExpenseResponse = {
  expense: ExpenseResponse;
};

/*
 * 経費更新Request
 *
 * receiptFile が null/undefined の場合、既存領収書は差し替えない。
 */
export type UpdateExpenseRequest = {
  expenseId: number;

  targetUserId: number;

  targetMonth: string;
  expenseDate: string;

  amount: number;
  description: string;
  memo: string | null;

  receiptFile?: File | null;
};

export type UpdateExpenseResponse = {
  expense: ExpenseResponse;
};

/*
 * 経費削除Request
 */
export type DeleteExpenseRequest = {
  expenseId: number;
};

export type DeleteExpenseResponse = {
  expenseId: number;
};

/*
 * 領収書表示Request
 */
export type ViewExpenseReceiptRequest = {
  expenseId: number;
};
