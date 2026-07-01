/*
 * 従業員 日別交通費 Type
 *
 * バックエンドの user AttendanceTransportExpenseResponse に対応する。
 *
 * 注意：
 * ・ログイン中ユーザー本人の対象年月の交通費明細を取得する
 * ・1つの勤怠日に対して複数件の交通費明細を持てる
 * ・保存は月次勤怠全体保存APIへ集約する
 */

export type AttendanceTransportExpense = {
  id: number;

  attendanceDayId: number;
  workDate: string;

  sortOrder: number;

  transportFrom: string;
  transportTo: string;
  transportMethod: string;
  transportAmount: number;
  transportMemo: string | null;

  isDeleted: boolean;

  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchAttendanceTransportExpensesRequest = {
  targetYear: number;
  targetMonth: number;
};

export type SearchAttendanceTransportExpensesResponse = {
  targetYear: number;
  targetMonth: number;

  attendanceTransportExpenses: AttendanceTransportExpense[];
};
