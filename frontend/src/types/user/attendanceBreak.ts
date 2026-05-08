/*
 * 従業員 休憩 Type
 *
 * バックエンドの AttendanceBreakResponse / Request に対応する。
 *
 * 注意：
 * ・休憩の作成、更新、削除は月次勤怠全体保存APIへ集約する
 * ・このファイルでは検索、レスポンス型を中心に扱う
 */

export type AttendanceBreak = {
  id: number;

  attendanceDayId: number;

  breakStartAt: string;
  breakEndAt: string;

  breakMemo: string | null;

  isDeleted: boolean;

  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchAttendanceBreaksRequest = {
  workDate: string;
};

export type SearchAttendanceBreaksResponse = {
  workDate: string;
  attendanceBreaks: AttendanceBreak[];
};