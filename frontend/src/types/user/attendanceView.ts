/*
 * 従業員 勤怠画面用 Type
 *
 * 注意：
 * これはバックエンドAPIのRequest/Response型ではない。
 * page.tsx や component が扱いやすいように整形した画面表示専用型。
 */

export type PageMessageVariant = "info" | "success" | "warning" | "error";

/*
 * 休憩1件分の画面用Row
 *
 * APIでは日時は RFC3339。
 * 画面では input type="time" で扱いやすいよう HH:mm で持つ。
 */
export type AttendanceBreakViewRow = {
  id: number | null;
  breakStartTime: string;
  breakEndTime: string;
  breakMemo: string;
  isNew: boolean;
  isDirty: boolean;
};

/*
 * 勤怠1日分の画面用Row
 *
 * APIでは日時は RFC3339 だが、
 * 画面では input type="time" で扱いやすいよう HH:mm で持つ。
 */
export type AttendanceViewRow = {
  workDate: string;
  dayLabel: string;
  weekday: string;

  attendanceDayId: number | null;

  planAttendanceTypeId: number;
  actualAttendanceTypeId: number | null;

  commonStartTime: string;
  commonEndTime: string;

  planStartTime: string;
  planEndTime: string;

  actualStartTime: string;
  actualEndTime: string;

  lateFlag: boolean;
  earlyLeaveFlag: boolean;
  absenceFlag: boolean;
  sickLeaveFlag: boolean;

  requestStatus: string;
  requestMemo: string;
  rejectedReason: string | null;
  systemMessage: string | null;

  monthlyStatus: string;

  transportFrom: string;
  transportTo: string;
  transportMethod: string;
  transportAmount: string;

  breaks: AttendanceBreakViewRow[];

  isDirty: boolean;
};

/*
 * 月次通勤定期の画面用Form
 *
 * APIでは commuterAmount は number | null。
 * 画面では input value として扱いやすいよう string で持つ。
 */
export type CommuterPassViewForm = {
  commuterFrom: string;
  commuterTo: string;
  commuterMethod: string;
  commuterAmount: string;
  monthlyStatus: string;
};