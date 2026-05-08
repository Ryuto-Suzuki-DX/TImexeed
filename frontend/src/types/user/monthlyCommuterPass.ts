/*
 * 従業員 月次通勤定期 Type
 *
 * バックエンドの MonthlyCommuterPassResponse / Request に対応する。
 *
 * 注意：
 * ・月次通勤定期の更新は月次勤怠全体保存APIへ集約する
 * ・このファイルでは検索、削除、レスポンス型を中心に扱う
 */

export type MonthlyCommuterPass = {
  id: number;

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
  targetYear: number;
  targetMonth: number;
};

export type SearchMonthlyCommuterPassResponse = {
  targetYear: number;
  targetMonth: number;
  monthlyCommuterPass: MonthlyCommuterPass | null;
};

export type DeleteMonthlyCommuterPassRequest = {
  targetYear: number;
  targetMonth: number;
};

export type DeleteMonthlyCommuterPassResponse = {
  targetYear: number;
  targetMonth: number;
};