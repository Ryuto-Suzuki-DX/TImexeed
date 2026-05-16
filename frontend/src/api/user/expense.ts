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
} from "@/types/user/expense";

/*
 * 従業員 経費検索
 *
 * POST /user/expenses/search
 */
export function searchExpenses(request: SearchExpensesRequest) {
  return apiPost<SearchExpensesResponse, SearchExpensesRequest>(
    "/user/expenses/search",
    request
  );
}

/*
 * 従業員 経費詳細取得
 *
 * POST /user/expenses/detail
 */
export function getExpenseDetail(request: ExpenseDetailRequest) {
  return apiPost<ExpenseDetailResponse, ExpenseDetailRequest>(
    "/user/expenses/detail",
    request
  );
}

/*
 * 従業員 経費作成
 *
 * POST /user/expenses/create
 *
 * バックエンド側が multipart/form-data のため、
 * apiPost(JSON送信)は使わず fetch + FormData で送信する。
 */
export function createExpense(request: CreateExpenseRequest) {
  const formData = buildCreateExpenseFormData(request);

  return apiPostFormData<CreateExpenseResponse>("/user/expenses/create", formData);
}

/*
 * 従業員 経費更新
 *
 * POST /user/expenses/update
 *
 * バックエンド側が multipart/form-data のため、
 * apiPost(JSON送信)は使わず fetch + FormData で送信する。
 */
export function updateExpense(request: UpdateExpenseRequest) {
  const formData = buildUpdateExpenseFormData(request);

  return apiPostFormData<UpdateExpenseResponse>("/user/expenses/update", formData);
}

/*
 * 従業員 経費削除
 *
 * POST /user/expenses/delete
 */
export function deleteExpense(request: DeleteExpenseRequest) {
  return apiPost<DeleteExpenseResponse, DeleteExpenseRequest>(
    "/user/expenses/delete",
    request
  );
}

/*
 * 従業員 経費領収書表示
 *
 * POST /user/expenses/receipt/view
 *
 * このAPIは成功時、共通JSONではなく画像/PDFなどのファイル本体を返す。
 */
export function viewExpenseReceipt(request: ViewExpenseReceiptRequest) {
  return apiPostBlob("/user/expenses/receipt/view", request);
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

function buildCreateExpenseFormData(request: CreateExpenseRequest) {
  const formData = new FormData();

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

function buildUpdateExpenseFormData(request: UpdateExpenseRequest) {
  const formData = new FormData();

  formData.append("expenseId", String(request.expenseId));
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

async function apiPostFormData<TData>(path: string, formData: FormData): Promise<ApiResponse<TData>> {
  const response = await fetch(buildApiUrl(path), {
    method: "POST",
    headers: buildAuthHeaders(),
    body: formData,
  });

  return readJsonResponse<TData>(response);
}

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

function buildApiUrl(path: string) {
  const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
  const normalizedBaseUrl = baseUrl.endsWith("/") ? baseUrl.slice(0, -1) : baseUrl;
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  return `${normalizedBaseUrl}${normalizedPath}`;
}

function buildAuthHeaders(): HeadersInit {
  const token = getAccessToken();

  if (!token) {
    return {};
  }

  return {
    Authorization: `Bearer ${token}`,
  };
}

function getAccessToken() {
  if (typeof window === "undefined") {
    return null;
  }

  return window.localStorage.getItem("accessToken");
}

async function readJsonResponse<TData>(response: Response): Promise<ApiResponse<TData>> {
  const payload = (await response.json()) as ApiResponse<TData>;

  if (!response.ok) {
    throw new Error(payload.message || "API request failed");
  }

  return payload;
}

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

  return response.text();
}
