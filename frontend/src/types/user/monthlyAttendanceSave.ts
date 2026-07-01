/*
 * 従業員 月次勤怠全体保存 Type
 *
 * バックエンドの月次勤怠全体保存APIに対応する。
 *
 * 対象：
 * ・月次通勤定期
 * ・日別勤怠
 * ・日別休憩
 * ・日別交通費
 *
 * 注意：
 * ・画面用stateではない
 * ・APIに送るRequest型
 * ・画面用の AttendanceViewRow / CommuterPassViewForm から mapper で変換して作る
 * ・日別勤怠に申請メモは送らない
 * ・予定区分は planAttendanceTypeId
 * ・実績状態は actualWorkStatus
 * ・actualAttendanceTypeId は使わない
 */

export type UpdateMonthlyAttendanceSaveRequest = {
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

  transportExpenses: UpdateMonthlyAttendanceSaveTransportExpenseRequest[];

  breaks: UpdateMonthlyAttendanceSaveBreakRequest[];
};

export type UpdateMonthlyAttendanceSaveTransportExpenseRequest = {
  attendanceTransportExpenseId: number | null;
  sortOrder: number;
  transportFrom: string;
  transportTo: string;
  transportMethod: string;
  transportAmount: number;
  transportMemo: string | null;
};

export type UpdateMonthlyAttendanceSaveBreakRequest = {
  breakStartAt: string;
  breakEndAt: string;
  breakMemo: string | null;
};

export type UpdateMonthlyAttendanceSaveResponse = {
  targetYear: number;
  targetMonth: number;

  savedMonthlyCommuterPass: boolean;
  savedAttendanceDayCount: number;
  savedAttendanceBreakCount: number;
  savedAttendanceTransportExpenseCount: number;
};
