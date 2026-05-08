import { apiPost } from "@/api/client";
import type {
  DeleteMonthlyCommuterPassRequest,
  DeleteMonthlyCommuterPassResponse,
  SearchMonthlyCommuterPassRequest,
  SearchMonthlyCommuterPassResponse,
} from "@/types/user/monthlyCommuterPass";

/*
 * 月次通勤定期検索
 *
 * POST /user/monthly-commuter-passes/search
 */
export function searchMonthlyCommuterPass(request: SearchMonthlyCommuterPassRequest) {
  return apiPost<SearchMonthlyCommuterPassResponse, SearchMonthlyCommuterPassRequest>("/user/monthly-commuter-passes/search", request);
}

/*
 * 月次通勤定期削除
 *
 * POST /user/monthly-commuter-passes/delete
 */
export function deleteMonthlyCommuterPass(request: DeleteMonthlyCommuterPassRequest) {
  return apiPost<DeleteMonthlyCommuterPassResponse, DeleteMonthlyCommuterPassRequest>("/user/monthly-commuter-passes/delete", request);
}