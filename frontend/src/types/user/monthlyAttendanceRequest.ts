/*
 * 従業員 月次勤怠申請 Type
 *
 * バックエンドの MonthlyAttendanceRequestResponse に対応する。
 *
 * 注意：
 * ・月次申請状態は日別勤怠や月次通勤定期ではなく、この型で管理する
 * ・勤怠日の保存とは別物
 * ・月次勤怠全体保存APIとは別物
 */

export type MonthlyAttendanceRequestStatus = "DRAFT" | "PENDING" | "APPROVED" | "REJECTED";

export type MonthlyAttendanceRequest = {
  id: number;

  targetYear: number;
  targetMonth: number;

  status: MonthlyAttendanceRequestStatus;

  requestedAt: string | null;
  approvedAt: string | null;
  rejectedAt: string | null;

  approvedBy: number | null;
  rejectedBy: number | null;

  adminMessage: string | null;

  isDeleted: boolean;

  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

/*
 * 月次勤怠申請状態検索
 */
export type SearchMonthlyAttendanceRequestRequest = {
  targetYear: number;
  targetMonth: number;
};

export type SearchMonthlyAttendanceRequestResponse = {
  targetYear: number;
  targetMonth: number;
  monthlyAttendanceRequest: MonthlyAttendanceRequest | null;
};

/*
 * 月次勤怠申請
 */
export type SubmitMonthlyAttendanceRequestRequest = {
  targetYear: number;
  targetMonth: number;
};

export type SubmitMonthlyAttendanceRequestResponse = {
  targetYear: number;
  targetMonth: number;
  monthlyAttendanceRequest: MonthlyAttendanceRequest;
};

/*
 * 月次勤怠申請取り下げ
 */
export type WithdrawMonthlyAttendanceRequestRequest = {
  targetYear: number;
  targetMonth: number;
};

export type WithdrawMonthlyAttendanceRequestResponse = {
  targetYear: number;
  targetMonth: number;
  monthlyAttendanceRequest: MonthlyAttendanceRequest;
};