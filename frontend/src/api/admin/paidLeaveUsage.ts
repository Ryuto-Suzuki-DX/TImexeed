import { apiPost } from "@/api/client";
import type {
  CreatePaidLeaveUsageRequest,
  CreatePaidLeaveUsageResponse,
  DeletePaidLeaveUsageRequest,
  DeletePaidLeaveUsageResponse,
  GetPaidLeaveBalanceRequest,
  PaidLeaveBalanceResponse,
  SearchPaidLeaveUsagesRequest,
  SearchPaidLeaveUsagesResponse,
  UpdatePaidLeaveUsageRequest,
  UpdatePaidLeaveUsageResponse,
} from "@/types/admin/paidLeaveUsage";

/*
 * 管理者 有給使用日一覧取得
 *
 * POST /admin/paid-leave-usages/search
 */
export function searchPaidLeaveUsages(request: SearchPaidLeaveUsagesRequest) {
  return apiPost<SearchPaidLeaveUsagesResponse, SearchPaidLeaveUsagesRequest>(
    "/admin/paid-leave-usages/search",
    request
  );
}

/*
 * 管理者 有給残数取得
 *
 * POST /admin/paid-leave-usages/balance
 */
export function getPaidLeaveBalance(request: GetPaidLeaveBalanceRequest) {
  return apiPost<PaidLeaveBalanceResponse, GetPaidLeaveBalanceRequest>(
    "/admin/paid-leave-usages/balance",
    request
  );
}

/*
 * 管理者 過去有給使用日追加
 *
 * POST /admin/paid-leave-usages/create
 *
 * 注意：
 * ・isManual はフロントから送らない
 * ・バックエンドService側で true にする
 */
export function createPaidLeaveUsage(request: CreatePaidLeaveUsageRequest) {
  return apiPost<CreatePaidLeaveUsageResponse, CreatePaidLeaveUsageRequest>(
    "/admin/paid-leave-usages/create",
    request
  );
}

/*
 * 管理者 過去有給使用日更新
 *
 * POST /admin/paid-leave-usages/update
 */
export function updatePaidLeaveUsage(request: UpdatePaidLeaveUsageRequest) {
  return apiPost<UpdatePaidLeaveUsageResponse, UpdatePaidLeaveUsageRequest>(
    "/admin/paid-leave-usages/update",
    request
  );
}

/*
 * 管理者 過去有給使用日削除
 *
 * POST /admin/paid-leave-usages/delete
 */
export function deletePaidLeaveUsage(request: DeletePaidLeaveUsageRequest) {
  return apiPost<DeletePaidLeaveUsageResponse, DeletePaidLeaveUsageRequest>(
    "/admin/paid-leave-usages/delete",
    request
  );
}