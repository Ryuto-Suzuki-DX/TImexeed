import { apiPost } from "@/api/client";
import type {
  SearchMonthlyCommuterPassRequest,
  SearchMonthlyCommuterPassResponse,
} from "@/types/user/monthlyCommuterPass";

/*
 * 月次通勤定期検索
 *
 * POST /user/monthly-commuter-passes/search
 */
export function searchMonthlyCommuterPass(request: SearchMonthlyCommuterPassRequest) {
  return apiPost<SearchMonthlyCommuterPassResponse, SearchMonthlyCommuterPassRequest>(
    "/user/monthly-commuter-passes/search",
    request
  );
}