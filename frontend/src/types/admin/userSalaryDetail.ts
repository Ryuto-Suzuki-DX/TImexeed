/*
 * 管理者 ユーザー給与詳細
 *
 * バックエンドの UserSalaryDetail Request / Response に対応する。
 * 管理者だけがユーザーごとの給与詳細を操作する。
 */

export type SalaryType = "MONTHLY" | "HOURLY" | "DAILY";

export type SearchUserSalaryDetailsRequest = {
  targetUserId: number;
  includeDeleted: boolean;
  offset: number;
  limit: number;
};

export type GetUserSalaryDetailRequest = {
  userSalaryDetailId: number;
};

export type CreateUserSalaryDetailRequest = {
  targetUserId: number;

  salaryType: SalaryType;
  baseAmount: number;

  extraAllowanceAmount: number;
  extraAllowanceMemo: string;

  fixedDeductionAmount: number;
  fixedDeductionMemo: string;

  isPayrollTarget: boolean;

  effectiveFrom: string;
  effectiveTo: string | null;

  memo: string;
};

export type UpdateUserSalaryDetailRequest = {
  userSalaryDetailId: number;

  salaryType: SalaryType;
  baseAmount: number;

  extraAllowanceAmount: number;
  extraAllowanceMemo: string;

  fixedDeductionAmount: number;
  fixedDeductionMemo: string;

  isPayrollTarget: boolean;

  effectiveFrom: string;
  effectiveTo: string | null;

  memo: string;
};

export type DeleteUserSalaryDetailRequest = {
  userSalaryDetailId: number;
};

export type UserSalaryDetailResponse = {
  id: number;

  userId: number;

  salaryType: SalaryType;
  baseAmount: number;

  extraAllowanceAmount: number;
  extraAllowanceMemo: string;

  fixedDeductionAmount: number;
  fixedDeductionMemo: string;

  isPayrollTarget: boolean;

  effectiveFrom: string;
  effectiveTo: string | null;

  memo: string;

  isDeleted: boolean;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchUserSalaryDetailsResponse = {
  userSalaryDetails: UserSalaryDetailResponse[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};

export type GetUserSalaryDetailResponse = {
  userSalaryDetail: UserSalaryDetailResponse;
};

export type CreateUserSalaryDetailResponse = {
  userSalaryDetail: UserSalaryDetailResponse;
};

export type UpdateUserSalaryDetailResponse = {
  userSalaryDetail: UserSalaryDetailResponse;
};

export type DeleteUserSalaryDetailResponse = {
  userSalaryDetailId: number;
};
