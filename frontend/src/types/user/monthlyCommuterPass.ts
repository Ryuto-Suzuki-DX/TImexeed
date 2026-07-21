/*
 * 従業員 月次通勤定期 Type
 *
 * バックエンドの MonthlyCommuterPassResponse に対応する。
 */

export type MonthlyCommuterPass = {
  id: number;

  targetYear: number;
  targetMonth: number;

  commuterFrom: string | null;
  commuterTo: string | null;
  commuterMethod: string | null;
  commuterAmount: number | null;

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

  monthlyCommuterPasses: MonthlyCommuterPass[];
  totalCommuterAmount: number;
};
