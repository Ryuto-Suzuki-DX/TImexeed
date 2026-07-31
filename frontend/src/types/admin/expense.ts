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


/*
 * 経費検索結果一式出力Request
 */
export type ExportExpensesRequest = {
  keyword: string;
  targetMonthFrom: string;
  targetMonthTo: string;
};
