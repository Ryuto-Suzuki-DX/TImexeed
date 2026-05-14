import { apiPost } from "@/api/client";
import type {
  SearchMonthlyCommuterPassRequest,
  SearchMonthlyCommuterPassResponse,
} from "@/types/admin/monthlyCommuterPass";

/*
 * 管理者 月次通勤定期検索
 *
 * POST /admin/monthly-commuter-passes/search
 *
 * 注意：
 * ・管理者APIでは targetUserId を送る
 */
export function searchMonthlyCommuterPass(request: SearchMonthlyCommuterPassRequest) {
  return apiPost<SearchMonthlyCommuterPassResponse, SearchMonthlyCommuterPassRequest>(
    "/admin/monthly-commuter-passes/search",
    request
  );
}
