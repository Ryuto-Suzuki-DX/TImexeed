/*
 * 有給
 */

export type PaidLeaveBalanceResponse = {
  userId: number;

  totalGrantedDays: number;
  usedDays: number;
  remainingDays: number;

  nextGrantDate: string | null;
  nextGrantDays: number;

  requiredUseDays: number;
  requiredUseDeadline: string | null;
  requiredUseRemainingDays: number;
};