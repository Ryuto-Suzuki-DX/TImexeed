import { apiPost } from "@/api/client";
import type {
  SearchMonthlyAttendanceRequestRequest,
  SearchMonthlyAttendanceRequestResponse,
  SubmitMonthlyAttendanceRequestRequest,
  SubmitMonthlyAttendanceRequestResponse,
  WithdrawMonthlyAttendanceRequestRequest,
  WithdrawMonthlyAttendanceRequestResponse,
} from "@/types/user/monthlyAttendanceRequest";

/*
 * 月次勤怠申請状態取得
 *
 * POST /user/monthly-attendance-requests/status
 *
 * 注意：
 * ・バックエンド側のルート名は search ではなく status
 */
export function searchMonthlyAttendanceRequest(request: SearchMonthlyAttendanceRequestRequest) {
  return apiPost<SearchMonthlyAttendanceRequestResponse, SearchMonthlyAttendanceRequestRequest>(
    "/user/monthly-attendance-requests/status",
    request,
  );
}

/*
 * 月次勤怠申請
 *
 * POST /user/monthly-attendance-requests/submit
 */
export function submitMonthlyAttendanceRequest(request: SubmitMonthlyAttendanceRequestRequest) {
  return apiPost<SubmitMonthlyAttendanceRequestResponse, SubmitMonthlyAttendanceRequestRequest>(
    "/user/monthly-attendance-requests/submit",
    request,
  );
}

/*
 * 月次勤怠申請取り下げ
 *
 * POST /user/monthly-attendance-requests/cancel
 *
 * 注意：
 * ・バックエンド側のルート名は withdraw ではなく cancel
 */
export function withdrawMonthlyAttendanceRequest(request: WithdrawMonthlyAttendanceRequestRequest) {
  return apiPost<WithdrawMonthlyAttendanceRequestResponse, WithdrawMonthlyAttendanceRequestRequest>(
    "/user/monthly-attendance-requests/cancel",
    request,
  );
}