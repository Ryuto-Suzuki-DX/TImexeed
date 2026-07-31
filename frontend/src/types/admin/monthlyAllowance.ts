/*
 * 管理者 月次手当
 *
 * ユーザーごと・対象年月ごとの手当明細を管理する。
 */

export type SearchMonthlyAllowancesRequest = {
  targetUserId: number | null;
  targetYear: number | null;
  targetMonth: number | null;
  allowanceTypeId: number | null;
  includeDeleted: boolean;
  offset: number;
  limit: number;
};

export type MonthlyAllowanceDetailRequest = {
  monthlyAllowanceId: number;
};

export type CreateMonthlyAllowanceRequest = {
  targetUserId: number;
  targetYear: number;
  targetMonth: number;
  allowanceTypeId: number;
  amount: number;
  memo: string;
};

export type UpdateMonthlyAllowanceRequest = {
  monthlyAllowanceId: number;
  targetUserId: number;
  targetYear: number;
  targetMonth: number;
  allowanceTypeId: number;
  amount: number;
  memo: string;
};

export type DeleteMonthlyAllowanceRequest = {
  monthlyAllowanceId: number;
};

export type MonthlyAllowanceResponse = {
  id: number;
  userId: number;
  targetYear: number;
  targetMonth: number;
  allowanceTypeId: number;
  allowanceTypeName: string;
  amount: number;
  memo: string;
  isDeleted: boolean;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchMonthlyAllowancesResponse = {
  monthlyAllowances: MonthlyAllowanceResponse[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};

export type MonthlyAllowanceDetailResponse = {
  monthlyAllowance: MonthlyAllowanceResponse;
};

export type CreateMonthlyAllowanceResponse = {
  monthlyAllowance: MonthlyAllowanceResponse;
};

export type UpdateMonthlyAllowanceResponse = {
  monthlyAllowance: MonthlyAllowanceResponse;
};

export type DeleteMonthlyAllowanceResponse = {
  monthlyAllowanceId: number;
};
