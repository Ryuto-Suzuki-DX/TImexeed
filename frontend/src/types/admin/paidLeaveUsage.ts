/*
 * 有給使用日
 *
 * 管理者用の有給使用日管理で使用する型定義。
 *
 * 対応API：
 * ・有給使用日一覧取得
 * ・有給残数取得
 * ・過去有給使用日追加
 * ・過去有給使用日更新
 * ・過去有給使用日削除
 *
 * 注意：
 * ・targetUserId は管理者が操作対象にする一般ユーザーID
 * ・targetPaidLeaveUsageId は操作対象の有給使用日ID
 * ・日付はバックエンドから string として受け取る
 * ・表示形式 yyyy-MM-dd などはフロント側で整形する
 * ・isManual は追加Requestでは送らない
 *   → バックエンド側で true にする
 */

/*
 * =========================================================
 * Request
 * =========================================================
 */

export type SearchPaidLeaveUsagesRequest = {
  targetUserId: number;
  includeDeleted: boolean;
  offset: number;
  limit: number;
};

export type GetPaidLeaveBalanceRequest = {
  targetUserId: number;
};

export type CreatePaidLeaveUsageRequest = {
  targetUserId: number;
  usageDate: string;
  usageDays: number;
  memo: string;
};

export type UpdatePaidLeaveUsageRequest = {
  targetUserId: number;
  targetPaidLeaveUsageId: number;
  usageDate: string;
  usageDays: number;
  memo: string;
};

export type DeletePaidLeaveUsageRequest = {
  targetUserId: number;
  targetPaidLeaveUsageId: number;
};

/*
 * =========================================================
 * Response
 * =========================================================
 */

export type PaidLeaveUsageResponse = {
  id: number;
  userId: number;
  usageDate: string;
  usageDays: number;
  isManual: boolean;
  memo: string;
  isDeleted: boolean;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
};

export type SearchPaidLeaveUsagesResponse = {
  paidLeaveUsages: PaidLeaveUsageResponse[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};

export type PaidLeaveBalanceResponse = {
  targetUserId: number;

  totalGrantedDays: number;
  usedDays: number;
  remainingDays: number;

  nextGrantDate: string | null;
  nextGrantDays: number;

  requiredUseDays: number;
  requiredUseDeadline: string | null;
  requiredUseRemainingDays: number;
};

export type CreatePaidLeaveUsageResponse = {
  paidLeaveUsage: PaidLeaveUsageResponse;
};

export type UpdatePaidLeaveUsageResponse = {
  paidLeaveUsage: PaidLeaveUsageResponse;
};

export type DeletePaidLeaveUsageResponse = {
  targetUserId: number;
  targetPaidLeaveUsageId: number;
};