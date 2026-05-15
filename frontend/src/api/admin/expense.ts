import { apiPost } from "@/api/client";
import type {
  ApiResponse,
  CreateExpenseRequest,
  CreateExpenseResponse,
  DeleteExpenseRequest,
  DeleteExpenseResponse,
  ExpenseDetailRequest,
  ExpenseDetailResponse,
  SearchExpensesRequest,
  SearchExpensesResponse,
  UpdateExpenseRequest,
  UpdateExpenseResponse,
  ViewExpenseReceiptRequest,
} from "@/types/admin/expense";

/*
 * 管理者 経費検索
 *
 * POST /admin/expenses/search
 */
export function searchExpenses(request: SearchExpensesRequest) {
  return apiPost<SearchExpensesResponse, SearchExpensesRequest>(
    "/admin/expenses/search",
    request
  );
}

/*
 * 管理者 経費詳細取得
 *
 * POST /admin/expenses/detail
 */
export function getExpenseDetail(request: ExpenseDetailRequest) {
  return apiPost<ExpenseDetailResponse, ExpenseDetailRequest>(
    "/admin/expenses/detail",
    request
  );
}

/*
 * 管理者 経費作成
 *
 * POST /admin/expenses/create
 *
 * バックエンド側が multipart/form-data のため、
 * apiPost(JSON送信)は使わず fetch + FormData で送信する。
 */
export function createExpense(request: CreateExpenseRequest) {
  const formData = buildCreateExpenseFormData(request);

  return apiPostFormData<CreateExpenseResponse>("/admin/expenses/create", formData);
}

/*
 * 管理者 経費更新
 *
 * POST /admin/expenses/update
 *
 * バックエンド側が multipart/form-data のため、
 * apiPost(JSON送信)は使わず fetch + FormData で送信する。
 */
export function updateExpense(request: UpdateExpenseRequest) {
  const formData = buildUpdateExpenseFormData(request);

  return apiPostFormData<UpdateExpenseResponse>("/admin/expenses/update", formData);
}

/*
 * 管理者 経費削除
 *
 * POST /admin/expenses/delete
 */
export function deleteExpense(request: DeleteExpenseRequest) {
  return apiPost<DeleteExpenseResponse, DeleteExpenseRequest>(
    "/admin/expenses/delete",
    request
  );
}

/*
 * 管理者 経費領収書表示
 *
 * POST /admin/expenses/receipt/view
 *
 * このAPIは成功時、共通JSONではなく画像/PDFなどのファイル本体を返す。
 */
export function viewExpenseReceipt(request: ViewExpenseReceiptRequest) {
  return apiPostBlob("/admin/expenses/receipt/view", request);
}

/*
 * 領収書を別タブで開く。
 */
export async function openExpenseReceiptInNewTab(request: ViewExpenseReceiptRequest) {
  const blob = await viewExpenseReceipt(request);
  const objectUrl = URL.createObjectURL(blob);

  window.open(objectUrl, "_blank", "noopener,noreferrer");

  window.setTimeout(() => {
    URL.revokeObjectURL(objectUrl);
  }, 60_000);
}

/*
 * 作成用FormData作成
 */
function buildCreateExpenseFormData(request: CreateExpenseRequest) {
  const formData = new FormData();

  formData.append("targetUserId", String(request.targetUserId));
  formData.append("targetMonth", request.targetMonth);
  formData.append("expenseDate", request.expenseDate);
  formData.append("amount", String(request.amount));
  formData.append("description", request.description);

  if (request.memo !== null && request.memo !== undefined && request.memo.trim() !== "") {
    formData.append("memo", request.memo);
  }

  if (request.receiptFile) {
    formData.append("receiptFile", request.receiptFile);
  }

  return formData;
}

/*
 * 更新用FormData作成
 */
function buildUpdateExpenseFormData(request: UpdateExpenseRequest) {
  const formData = new FormData();

  formData.append("expenseId", String(request.expenseId));
  formData.append("targetUserId", String(request.targetUserId));
  formData.append("targetMonth", request.targetMonth);
  formData.append("expenseDate", request.expenseDate);
  formData.append("amount", String(request.amount));
  formData.append("description", request.description);

  if (request.memo !== null && request.memo !== undefined && request.memo.trim() !== "") {
    formData.append("memo", request.memo);
  }

  if (request.receiptFile) {
    formData.append("receiptFile", request.receiptFile);
  }

  return formData;
}

/*
 * multipart/form-data POST
 *
 * 注意：
 * Content-Type は手動で指定しない。
 * ブラウザが multipart boundary を自動付与するため。
 */
async function apiPostFormData<TData>(path: string, formData: FormData): Promise<ApiResponse<TData>> {
  const response = await fetch(buildApiUrl(path), {
    method: "POST",
    headers: buildAuthHeaders(),
    body: formData,
  });

  return readJsonResponse<TData>(response);
}

/*
 * blob取得用POST
 */
async function apiPostBlob<TRequest extends object>(path: string, request: TRequest): Promise<Blob> {
  const response = await fetch(buildApiUrl(path), {
    method: "POST",
    headers: {
      ...buildAuthHeaders(),
      "Content-Type": "application/json",
    },
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    const errorPayload = await readErrorPayload(response);
    throw new Error(errorPayload);
  }

  return response.blob();
}

/*
 * API URL作成
 */
function buildApiUrl(path: string) {
  const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
  const normalizedBaseUrl = baseUrl.endsWith("/") ? baseUrl.slice(0, -1) : baseUrl;
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  return `${normalizedBaseUrl}${normalizedPath}`;
}

/*
 * 認証ヘッダー作成
 */
function buildAuthHeaders(): HeadersInit {
  const token = getAccessToken();

  if (!token) {
    return {};
  }

  return {
    Authorization: `Bearer ${token}`,
  };
}

/*
 * localStorageからアクセストークン取得
 */
function getAccessToken() {
  if (typeof window === "undefined") {
    return null;
  }

  return window.localStorage.getItem("accessToken");
}

/*
 * 共通JSONレスポンス読み取り
 */
async function readJsonResponse<TData>(response: Response): Promise<ApiResponse<TData>> {
  const payload = (await response.json()) as ApiResponse<TData>;

  if (!response.ok) {
    throw new Error(payload.message || "API request failed");
  }

  return payload;
}

/*
 * エラーレスポンス読み取り
 */
async function readErrorPayload(response: Response) {
  const contentType = response.headers.get("Content-Type") ?? "";

  if (contentType.includes("application/json")) {
    try {
      const payload = (await response.json()) as Partial<ApiResponse<unknown>>;
      return payload.message || payload.code || "API request failed";
    } catch {
      return "API request failed";
    }
  }

  try {
    const text = await response.text();
    return text || "API request failed";
  } catch {
    return "API request failed";
  }
}
