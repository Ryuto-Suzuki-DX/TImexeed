/*
 * 従業員 勤怠日 Type
 *
 * バックエンドの AttendanceDayResponse に対応する。
 *
 * 注意：
 * ・日別勤怠の単体更新/削除は行わない
 * ・登録/更新/初期値戻しは月次勤怠全体保存APIへ集約する
 * ・月次申請状態は MonthlyAttendanceRequest 側で管理する
 * ・systemMessage は保存せず、画面側で計算して表示する
 */

export type AttendanceDay = {
  id: number;

  workDate: string;

  planAttendanceTypeId: number;
  actualAttendanceTypeId: number;

  planStartAt: string | null;
  planEndAt: string | null;

  actualStartAt: string | null;
  actualEndAt: string | null;

  requestMemo: string | null;

  lateFlag: boolean;
  earlyLeaveFlag: boolean;
  absenceFlag: boolean;
  sickLeaveFlag: boolean;

  remoteWorkAllowanceFlag: boolean;

  transportFrom: string | null;
  transportTo: string | null;
  transportMethod: string | null;
  transportAmount: number | null;

  isDeleted: boolean;

  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchAttendanceDaysRequest = {
  targetYear: number;
  targetMonth: number;
};

export type SearchAttendanceDaysResponse = {
  targetYear: number;
  targetMonth: number;
  attendanceDays: AttendanceDay[];
};