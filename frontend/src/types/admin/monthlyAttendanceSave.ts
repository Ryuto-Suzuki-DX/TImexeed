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
 * ・予定区分は planAttendanceTypeId
 * ・実績状態は actualWorkStatus
 * ・actualAttendanceTypeId は使わない
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

  /*
   * 実績状態
   *
   * バックエンド constants/attendance_status_constants.go の固定値。
   * 例: NORMAL, ABSENCE, SICK_LEAVE, LATE, EARLY_LEAVE
   */
  actualWorkStatus: string | null;

  commonStartAt: string | null;
  commonEndAt: string | null;

  planStartAt: string | null;
  planEndAt: string | null;

  actualStartAt: string | null;
  actualEndAt: string | null;

  scheduledWorkMinutes: number | null;

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
