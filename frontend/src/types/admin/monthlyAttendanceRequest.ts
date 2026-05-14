/*
 * 管理者 月次勤怠申請 Type
 *
 * バックエンドの admin MonthlyAttendanceRequestResponse に対応する。
 *
 * 注意：
 * ・月次申請状態は日別勤怠や月次通勤定期ではなく、この型で管理する
 * ・勤怠日の保存とは別物
 * ・月次勤怠全体保存APIとは別物
 * ・管理者APIでは対象ユーザーIDを targetUserId で送る
 * ・管理者側では月次申請状態による勤怠編集ロックを行わない
 * ・承認者IDはフロントから送らず、バックエンドでJWTから取得する
 */

export type MonthlyAttendanceRequestStatus =
  | "NOT_SUBMITTED"
  | "PENDING"
  | "APPROVED"
  | "REJECTED"
  | "CANCELED";

export type MonthlyAttendanceRequest = {
  id: number | null;

  targetUserId: number;

  targetYear: number;
  targetMonth: number;

  status: MonthlyAttendanceRequestStatus;

  exists: boolean;

  editable: boolean;
  canSubmit: boolean;
  canCancel: boolean;
  canApprove: boolean;
  canReject: boolean;

  requestMemo: string | null;
  requestedAt: string | null;

  approvedBy: number | null;
  approvedAt: string | null;

  rejectedReason: string | null;
  rejectedAt: string | null;

  canceledReason: string | null;
  canceledAt: string | null;

  createdAt: string | null;
  updatedAt: string | null;
};

/*
 * 月次勤怠申請一覧検索 Row
 *
 * POST /admin/monthly-attendance-requests/search
 *
 * 注意：
 * ・申請一覧画面用
 * ・ユーザー情報と対象月の月次申請状態を1行にまとめた型
 * ・未申請の場合、monthlyAttendanceRequest.status は NOT_SUBMITTED
 * ・未申請の場合、monthlyAttendanceRequest.id は null
 */
export type MonthlyAttendanceRequestListRow = {
  targetUserId: number;

  userName: string;
  email: string;

  departmentId: number | null;
  departmentName: string | null;

  targetYear: number;
  targetMonth: number;

  monthlyAttendanceRequest: MonthlyAttendanceRequest;
};

/*
 * 月次勤怠申請一覧検索
 *
 * POST /admin/monthly-attendance-requests/search
 *
 * 注意：
 * ・対象年月、ユーザーのフリーワード、申請状態で絞り込む
 * ・statuses は複数選択可能
 * ・未申請も検索対象にする場合は NOT_SUBMITTED を含める
 * ・includeDeletedUsers は退職者/削除済みユーザーも対象に含める場合に true
 * ・offset / limit / hasMore でページングする
 */
export type SearchMonthlyAttendanceRequestsRequest = {
  keyword: string;

  targetYear: number;
  targetMonth: number;

  statuses: MonthlyAttendanceRequestStatus[];

  includeDeletedUsers: boolean;

  offset: number;
  limit: number;
};

export type SearchMonthlyAttendanceRequestsResponse = {
  monthlyAttendanceRequests: MonthlyAttendanceRequestListRow[];
  hasMore: boolean;
};

/*
 * 月次勤怠申請状態取得
 *
 * POST /admin/monthly-attendance-requests/status
 *
 * 注意：
 * ・バックエンド側のルート名は search ではなく status
 * ・user側の命名に合わせて SearchMonthlyAttendanceRequestRequest / Response とする
 */
export type SearchMonthlyAttendanceRequestRequest = {
  targetUserId: number;
  targetYear: number;
  targetMonth: number;
};

export type SearchMonthlyAttendanceRequestResponse = {
  monthlyAttendanceRequest: MonthlyAttendanceRequest;
};

/*
 * 月次勤怠申請
 *
 * POST /admin/monthly-attendance-requests/submit
 */
export type SubmitMonthlyAttendanceRequestRequest = {
  targetUserId: number;
  targetYear: number;
  targetMonth: number;
  requestMemo: string | null;
};

export type SubmitMonthlyAttendanceRequestResponse = {
  monthlyAttendanceRequest: MonthlyAttendanceRequest;
};

/*
 * 月次勤怠申請取り下げ
 *
 * POST /admin/monthly-attendance-requests/cancel
 *
 * 注意：
 * ・バックエンド側のルート名は withdraw ではなく cancel
 * ・user側の命名に合わせて WithdrawMonthlyAttendanceRequestRequest / Response とする
 */
export type WithdrawMonthlyAttendanceRequestRequest = {
  targetUserId: number;
  targetYear: number;
  targetMonth: number;
  canceledReason: string | null;
};

export type WithdrawMonthlyAttendanceRequestResponse = {
  monthlyAttendanceRequest: MonthlyAttendanceRequest;
};

/*
 * 月次勤怠申請承認
 *
 * POST /admin/monthly-attendance-requests/approve
 */
export type ApproveMonthlyAttendanceRequestRequest = {
  targetRequestId: number;
};

export type ApproveMonthlyAttendanceRequestResponse = {
  monthlyAttendanceRequest: MonthlyAttendanceRequest;
};

/*
 * 月次勤怠申請否認
 *
 * POST /admin/monthly-attendance-requests/reject
 */
export type RejectMonthlyAttendanceRequestRequest = {
  targetRequestId: number;
  rejectedReason: string;
};

export type RejectMonthlyAttendanceRequestResponse = {
  monthlyAttendanceRequest: MonthlyAttendanceRequest;
};
