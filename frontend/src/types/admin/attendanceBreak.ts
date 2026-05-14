/*
 * 管理者 休憩 Type
 *
 * バックエンドの admin AttendanceBreakResponse に対応する。
 *
 * 注意：
 * ・休憩の作成、更新、削除は月次勤怠全体保存APIへ集約する
 * ・このファイルでは検索、レスポンス型を中心に扱う
 * ・休憩は AttendanceDay に紐づく子データ
 * ・休憩検索APIは月単位ではなく、1日単位で検索する
 * ・管理者APIでは対象ユーザーIDを targetUserId で送る
 * ・管理者側では月次申請状態による編集ロックを行わない
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
  targetUserId: number;
  workDate: string;
};

export type SearchAttendanceBreaksResponse = {
  targetUserId: number;
  workDate: string;
  attendanceBreaks: AttendanceBreak[];
};
