import { apiPost } from "@/api/client";
import type {
  SearchAttendanceTransportExpensesRequest,
  SearchAttendanceTransportExpensesResponse,
} from "@/types/user/attendanceTransportExpense";

/*
 * 従業員 日別交通費検索
 *
 * POST /user/attendance-transport-expenses/search
 *
 * 注意：
 * ・ログイン中ユーザー本人の対象年月の日別交通費明細を取得する
 * ・1つの勤怠日に対して複数件の交通費明細が返る
 */
export function searchAttendanceTransportExpenses(
  request: SearchAttendanceTransportExpensesRequest,
) {
  return apiPost<
    SearchAttendanceTransportExpensesResponse,
    SearchAttendanceTransportExpensesRequest
  >("/user/attendance-transport-expenses/search", request);
}
