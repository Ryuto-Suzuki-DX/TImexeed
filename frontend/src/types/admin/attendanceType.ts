/*
 * 管理者 勤務区分マスタ Type
 *
 * バックエンドの admin AttendanceTypeResponse に対応する。
 *
 * 注意：
 * ・管理者勤怠画面の勤務区分プルダウンで使用する
 * ・管理者側でも勤務区分マスタの作成・更新・削除はここでは扱わない
 * ・user側と同じレスポンス構造
 */

export type AttendanceType = {
  id: number;
  code: string;
  name: string;
  category: string;

  syncPlanActual: boolean;

  allowActualTimeInput: boolean;
  allowBreakInput: boolean;
  allowTransportInput: boolean;

  allowLateFlag: boolean;
  allowEarlyLeaveFlag: boolean;
  allowAbsenceFlag: boolean;
  allowSickLeaveFlag: boolean;

  requiresRequest: boolean;

  displayOrder: number;
};

export type SearchAttendanceTypesRequest = Record<string, never>;

export type SearchAttendanceTypesResponse = {
  attendanceTypes: AttendanceType[];
};
