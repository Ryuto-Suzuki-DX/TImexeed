/*
 * 管理者 勤怠日 Type
 *
 * バックエンドの admin AttendanceDayResponse に対応する。
 *
 * 注意：
 * ・日別勤怠の単体更新/削除は画面から直接行わない
 * ・登録/更新/初期値戻しは月次勤怠全体保存APIへ集約する
 * ・月次申請状態は MonthlyAttendanceRequest 側で管理する
 * ・管理者側では月次申請状態による編集ロックを行わない
 * ・systemMessage は保存せず、画面側で計算して表示する
 * ・日別勤怠に申請メモは持たせない
 * ・管理者APIでは対象ユーザーIDを targetUserId で送る
 */

export type AttendanceDay = {
  id: number;

  userId: number;

  workDate: string;

  planAttendanceTypeId: number;
  actualAttendanceTypeId: number;

  planStartAt: string | null;
  planEndAt: string | null;

  actualStartAt: string | null;
  actualEndAt: string | null;

  lateFlag: boolean;
  earlyLeaveFlag: boolean;
  absenceFlag: boolean;
  sickLeaveFlag: boolean;

  remoteWorkAllowanceFlag: boolean;

  transportFrom: string | null;
  transportTo: string | null;
  transportMethod: string | null;
  transportAmount: number | null;

  isDeleted: boolean;

  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchAttendanceDaysRequest = {
  targetUserId: number;
  targetYear: number;
  targetMonth: number;
};

export type SearchAttendanceDaysResponse = {
  targetUserId: number;
  targetYear: number;
  targetMonth: number;
  attendanceDays: AttendanceDay[];
};
