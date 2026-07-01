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
 * ・日別勤怠に申請メモは持たせない
 * ・予定区分は planAttendanceTypeId
 * ・実績状態は actualWorkStatus
 * ・actualAttendanceTypeId は使わない
 */

export type AttendanceDay = {
  id: number;

  workDate: string;

  planAttendanceTypeId: number;

  /*
   * 実績状態
   *
   * バックエンド constants/attendance_status_constants.go の固定値。
   * 例: NORMAL, ABSENCE, SICK_LEAVE, LATE, EARLY_LEAVE
   */
  actualWorkStatus: string;

  planStartAt: string | null;
  planEndAt: string | null;

  actualStartAt: string | null;
  actualEndAt: string | null;

  scheduledWorkMinutes: number | null;

  remoteWorkAllowanceFlag: boolean;

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
