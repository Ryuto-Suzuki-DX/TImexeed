/*
 * 管理者 月次勤怠全体保存 Type
 *
 * バックエンドの管理者用 月次勤怠全体保存APIに対応する。
 *
 * 対象：
 * ・月次通勤定期
 * ・日別勤怠
 * ・日別休憩
 *
 * 注意：
 * ・画面用stateではない
 * ・APIに送るRequest型
 * ・画面用の AttendanceViewRow / CommuterPassViewForm から mapper で変換して作る
 * ・日別勤怠に申請メモは送らない
 * ・管理者APIなので操作対象ユーザーIDを targetUserId で送る
 * ・管理者側では月次申請状態による編集ロックは行わない
 */

export type UpdateMonthlyAttendanceSaveRequest = {
  targetUserId: number;

  targetYear: number;
  targetMonth: number;

  commuterPass: UpdateMonthlyAttendanceSaveCommuterPassRequest | null;

  attendanceDays: UpdateMonthlyAttendanceSaveDayRequest[];
};

export type UpdateMonthlyAttendanceSaveCommuterPassRequest = {
  commuterFrom: string | null;
  commuterTo: string | null;
  commuterMethod: string | null;
  commuterAmount: number | null;
};

export type UpdateMonthlyAttendanceSaveDayRequest = {
  workDate: string;

  planAttendanceTypeId: number;
  actualAttendanceTypeId: number | null;

  commonStartAt: string | null;
  commonEndAt: string | null;

  planStartAt: string | null;
  planEndAt: string | null;

  actualStartAt: string | null;
  actualEndAt: string | null;

  lateFlag: boolean;
  earlyLeaveFlag: boolean;
  absenceFlag: boolean;
  sickLeaveFlag: boolean;

  remoteWorkAllowanceFlag: boolean;

  transportFrom: string | null;
  transportTo: string | null;
  transportMethod: string | null;
  transportAmount: number | null;

  breaks: UpdateMonthlyAttendanceSaveBreakRequest[];
};

export type UpdateMonthlyAttendanceSaveBreakRequest = {
  breakStartAt: string;
  breakEndAt: string;
  breakMemo: string | null;
};

export type UpdateMonthlyAttendanceSaveResponse = {
  targetUserId: number;

  targetYear: number;
  targetMonth: number;

  savedMonthlyCommuterPass: boolean;
  savedAttendanceDayCount: number;
  savedAttendanceBreakCount: number;
};