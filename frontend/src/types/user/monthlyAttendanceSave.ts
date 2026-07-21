/*
 * 従業員 月次勤怠全体保存 Type
 *
 * バックエンドの月次勤怠全体保存APIに対応する。
 */

export type UpdateMonthlyAttendanceSaveRequest = {
  targetYear: number;
  targetMonth: number;

  commuterPasses: UpdateMonthlyAttendanceSaveCommuterPassRequest[];

  attendanceDays: UpdateMonthlyAttendanceSaveDayRequest[];
};

export type UpdateMonthlyAttendanceSaveCommuterPassRequest = {
  monthlyCommuterPassId: number | null;

  commuterFrom: string | null;
  commuterTo: string | null;
  commuterMethod: string | null;
  commuterAmount: number | null;
};

export type UpdateMonthlyAttendanceSaveDayRequest = {
  workDate: string;

  planAttendanceTypeId: number;
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

  savedMonthlyCommuterPassCount: number;
  savedAttendanceDayCount: number;
  savedAttendanceBreakCount: number;
  savedAttendanceTransportExpenseCount: number;
};
