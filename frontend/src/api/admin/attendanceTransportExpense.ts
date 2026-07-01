import { apiPost } from "@/api/client";
import type {
  SearchAttendanceTransportExpensesRequest,
  SearchAttendanceTransportExpensesResponse,
} from "@/types/admin/attendanceTransportExpense";

/*
 * 管理者 日別交通費検索
 *
 * POST /admin/attendance-transport-expenses/search
 *
 * 注意：
 * ・対象ユーザー + 対象年月の日別交通費明細を取得する
 * ・管理者APIでは targetUserId を送る
 * ・1つの勤怠日に対して複数件の交通費明細が返る
 */
export function searchAttendanceTransportExpenses(
  request: SearchAttendanceTransportExpensesRequest
) {
  return apiPost<
    SearchAttendanceTransportExpensesResponse,
    SearchAttendanceTransportExpensesRequest
  >("/admin/attendance-transport-expenses/search", request);
}
