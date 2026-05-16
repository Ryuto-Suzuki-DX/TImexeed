/*
 * 従業員 経費 Type
 *
 * 注意：
 * ・従業員APIでは userId / targetUserId を送らない
 * ・ログイン中ユーザーIDはバックエンドがJWTから取得する
 * ・検索、詳細、更新、削除、領収書表示は本人の経費だけが対象
 * ・create/update は multipart/form-data で送信する
 */

export type ApiResponse<TData> = {
  data: TData;
  error: boolean;
  code: string;
  message: string;
  details?: unknown;
};

export type SearchExpensesRequest = {
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

export type CreateExpenseRequest = {
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

export type UpdateExpenseRequest = {
  expenseId: number;

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

export type DeleteExpenseRequest = {
  expenseId: number;
};

export type DeleteExpenseResponse = {
  expenseId: number;
};

export type ViewExpenseReceiptRequest = {
  expenseId: number;
};
