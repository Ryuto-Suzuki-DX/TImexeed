/*
 * 管理者 勤怠画面用 Type
 *
 * 注意：
 * これはバックエンドAPIのRequest/Response型ではない。
 * page.tsx や component が扱いやすいように整形した画面表示専用型。
 *
 * 管理者側では、対象ユーザーの勤怠を編集するため、
 * API送信時には mapper 側で targetUserId を付与する。
 */

export type PageMessageVariant = "info" | "success" | "warning" | "error";

export type AttendanceBreakViewRow = {
  id: number | null;
  breakStartTime: string;
  breakEndTime: string;
  breakMemo: string;
  isNew: boolean;
  isDirty: boolean;
};

export type AttendanceTransportExpenseViewRow = {
  id: number | null;
  sortOrder: number;
  transportFrom: string;
  transportTo: string;
  transportMethod: string;
  transportAmount: string;
  transportMemo: string;
  isNew: boolean;
  isDirty: boolean;
};

export type AttendanceViewRow = {
  workDate: string;
  dayLabel: string;
  weekday: string;

  isHoliday: boolean;
  holidayName: string | null;

  attendanceDayId: number | null;

  planAttendanceTypeId: number;
  actualWorkStatus: string;

  commonStartTime: string;
  commonEndTime: string;

  planStartTime: string;
  planEndTime: string;

  actualStartTime: string;
  actualEndTime: string;

  scheduledWorkMinutes: string;

  lateFlag: boolean;
  earlyLeaveFlag: boolean;
  absenceFlag: boolean;
  sickLeaveFlag: boolean;

  remoteWorkAllowanceFlag: boolean;

  transportExpenses: AttendanceTransportExpenseViewRow[];
  breaks: AttendanceBreakViewRow[];

  isDirty: boolean;
};

/*
 * 月次通勤定期1件分の画面用Row
 *
 * 注意：
 * ・同じ対象年月に複数件登録できる
 * ・monthlyCommuterPassId が null の場合は新規登録
 * ・IDがある場合は既存レコード更新
 * ・画面から削除した行は全体保存時にバックエンド側で論理削除される
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
export type CommuterPassViewForm = {
  monthlyCommuterPassId: number | null;

  commuterFrom: string;
  commuterTo: string;
  commuterMethod: string;
  commuterAmount: string;

  isNew: boolean;
  isDirty: boolean;
};
