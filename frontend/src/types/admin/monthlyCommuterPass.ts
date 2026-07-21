/*
 * 管理者 月次通勤定期 Type
 *
 * バックエンドの admin MonthlyCommuterPassResponse に対応する。
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

  monthlyCommuterPasses: MonthlyCommuterPass[];
  totalCommuterAmount: number;
};
