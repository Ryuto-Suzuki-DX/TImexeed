/*
 * 管理者 月次通勤定期 Type
 *
 * バックエンドの admin MonthlyCommuterPassResponse に対応する。
 *
 * 注意：
 * ・月次通勤定期の更新/削除は月次勤怠全体保存APIへ集約する
 * ・このファイルでは検索、レスポンス型を中心に扱う
 * ・月次申請状態は MonthlyAttendanceRequest 側で管理する
 * ・管理者APIでは対象ユーザーIDを targetUserId で送る
 * ・管理者側では月次申請状態による編集ロックを行わない
 */

export type MonthlyCommuterPass = {
  id: number;

  userId: number;

  targetYear: number;
  targetMonth: number;

  commuterFrom: string | null;
  commuterTo: string | null;
  commuterMethod: string | null;
  commuterAmount: number | null;

  monthlyStatus: string;

  isDeleted: boolean;

  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchMonthlyCommuterPassRequest = {
  targetUserId: number;
  targetYear: number;
  targetMonth: number;
};

export type SearchMonthlyCommuterPassResponse = {
  targetUserId: number;
  targetYear: number;
  targetMonth: number;
  monthlyCommuterPass: MonthlyCommuterPass | null;
};
