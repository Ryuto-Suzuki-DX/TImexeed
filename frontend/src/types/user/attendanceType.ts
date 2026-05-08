/*
 * 従業員 勤務区分マスタ Type
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