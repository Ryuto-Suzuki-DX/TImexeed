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
 *
 * 注意：
 * ・月次申請状態は MonthlyAttendanceRequest 側で管理する
 * ・systemMessage は保存せず、画面側で計算して表示する
 * ・日別勤怠に申請メモは持たせない
 * ・日別勤怠の削除はAPIでは行わず、画面stateを初期値に戻して全体保存する
 * ・祝日はHolidayDate APIから取得し、画面表示用に保持する
 */
export type AttendanceViewRow = {
  workDate: string;
  dayLabel: string;
  weekday: string;

  isHoliday: boolean;
  holidayName: string | null;

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

  remoteWorkAllowanceFlag: boolean;

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
 *
 * 注意：
 * ・月次申請状態は MonthlyAttendanceRequest 側で管理する
 */
export type CommuterPassViewForm = {
  commuterFrom: string;
  commuterTo: string;
  commuterMethod: string;
  commuterAmount: string;
};
