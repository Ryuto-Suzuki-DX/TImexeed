import { apiPost } from "@/api/client";
import type {
  ApproveMonthlyAttendanceRequestRequest,
  ApproveMonthlyAttendanceRequestResponse,
  RejectMonthlyAttendanceRequestRequest,
  RejectMonthlyAttendanceRequestResponse,
  SearchMonthlyAttendanceRequestRequest,
  SearchMonthlyAttendanceRequestResponse,
  SubmitMonthlyAttendanceRequestRequest,
  SubmitMonthlyAttendanceRequestResponse,
  WithdrawMonthlyAttendanceRequestRequest,
  WithdrawMonthlyAttendanceRequestResponse,
} from "@/types/admin/monthlyAttendanceRequest";

/*
 * 管理者 月次勤怠申請状態取得
 *
 * POST /admin/monthly-attendance-requests/status
 *
 * 注意：
 * ・バックエンド側のルート名は search ではなく status
 * ・管理者APIでは targetUserId を送る
 */
export function searchMonthlyAttendanceRequest(request: SearchMonthlyAttendanceRequestRequest) {
  return apiPost<SearchMonthlyAttendanceRequestResponse, SearchMonthlyAttendanceRequestRequest>(
    "/admin/monthly-attendance-requests/status",
    request
  );
}

/*
 * 管理者 月次勤怠申請
 *
 * POST /admin/monthly-attendance-requests/submit
 *
 * 注意：
 * ・管理者が対象ユーザーの月次勤怠を代理申請する
 */
export function submitMonthlyAttendanceRequest(request: SubmitMonthlyAttendanceRequestRequest) {
  return apiPost<SubmitMonthlyAttendanceRequestResponse, SubmitMonthlyAttendanceRequestRequest>(
    "/admin/monthly-attendance-requests/submit",
    request
  );
}

/*
 * 管理者 月次勤怠申請取り下げ
 *
 * POST /admin/monthly-attendance-requests/cancel
 *
 * 注意：
 * ・バックエンド側のルート名は withdraw ではなく cancel
 * ・管理者が対象ユーザーの月次勤怠申請を代理で取り下げる
 */
export function withdrawMonthlyAttendanceRequest(request: WithdrawMonthlyAttendanceRequestRequest) {
  return apiPost<WithdrawMonthlyAttendanceRequestResponse, WithdrawMonthlyAttendanceRequestRequest>(
    "/admin/monthly-attendance-requests/cancel",
    request
  );
}

/*
 * 管理者 月次勤怠申請承認
 *
 * POST /admin/monthly-attendance-requests/approve
 *
 * 注意：
 * ・承認者IDはフロントから送らない
 * ・バックエンドでJWTから管理者IDを取得する
 */
export function approveMonthlyAttendanceRequest(request: ApproveMonthlyAttendanceRequestRequest) {
  return apiPost<ApproveMonthlyAttendanceRequestResponse, ApproveMonthlyAttendanceRequestRequest>(
    "/admin/monthly-attendance-requests/approve",
    request
  );
}

/*
 * 管理者 月次勤怠申請否認
 *
 * POST /admin/monthly-attendance-requests/reject
 *
 * 注意：
 * ・否認理由 rejectedReason は必須
 */
export function rejectMonthlyAttendanceRequest(request: RejectMonthlyAttendanceRequestRequest) {
  return apiPost<RejectMonthlyAttendanceRequestResponse, RejectMonthlyAttendanceRequestRequest>(
    "/admin/monthly-attendance-requests/reject",
    request
  );
}