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
 * 日別交通費1件分の画面用Row
 *
 * APIでは金額は number。
 * 画面では input value として扱いやすいよう string で持つ。
 */
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

/*
 * 勤怠1日分の画面用Row
 *
 * APIでは日時は RFC3339 だが、
 * 画面では input type="time" で扱いやすいよう HH:mm で持つ。
 *
 * 注意：
 * ・月次申請状態は MonthlyAttendanceRequest 側で管理する
 * ・管理者側では月次申請状態による編集ロックを行わない
 * ・systemMessage は保存せず、画面側で計算して表示する
 * ・日別勤怠に申請メモは持たせない
 * ・日別勤怠の削除はAPIでは行わず、画面stateを初期値に戻して全体保存する
 * ・祝日は HolidayDate API から取得し、画面表示用に保持する
 * ・予定区分は planAttendanceTypeId
 * ・実績状態は actualWorkStatus
 * ・actualAttendanceTypeId は使わない
 */
export type AttendanceViewRow = {
  workDate: string;
  dayLabel: string;
  weekday: string;

  isHoliday: boolean;
  holidayName: string | null;

  attendanceDayId: number | null;

  planAttendanceTypeId: number;

  /*
   * 実績状態
   *
   * バックエンド constants/attendance_status_constants.go の固定値。
   * 例: NORMAL, ABSENCE, SICK_LEAVE, LATE, EARLY_LEAVE
   */
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
 * 月次通勤定期の画面用Form
 *
 * APIでは commuterAmount は number | null。
 * 画面では input value として扱いやすいよう string で持つ。
 *
 * 注意：
 * ・月次申請状態は MonthlyAttendanceRequest 側で管理する
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
export type CommuterPassViewForm = {
  commuterFrom: string;
  commuterTo: string;
  commuterMethod: string;
  commuterAmount: string;
};
