/*
 * 従業員 勤怠日 Type
 *
 * バックエンドの AttendanceDayResponse / Request に対応する。
 *
 * 注意：
 * ・日別勤怠の単体更新は月次勤怠全体保存APIへ集約する
 * ・このファイルでは検索、削除、レスポンス型を中心に扱う
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

  requestStatus: string;
  requestMemo: string | null;

  approvedBy: number | null;
  approvedAt: string | null;
  rejectedReason: string | null;

  lateFlag: boolean;
  earlyLeaveFlag: boolean;
  absenceFlag: boolean;
  sickLeaveFlag: boolean;

  systemMessage: string | null;

  transportFrom: string | null;
  transportTo: string | null;
  transportMethod: string | null;
  transportAmount: number | null;

  monthlyStatus: string;

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

export type DeleteAttendanceDayRequest = {
  workDate: string;
};

export type DeleteAttendanceDayResponse = {
  workDate: string;
};