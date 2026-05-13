import { apiGet } from "@/api/client";
import type { PaidLeaveBalanceResponse } from "@/types/user/paidLeave";

export function getPaidLeaveBalance() {
  return apiGet<PaidLeaveBalanceResponse>("/user/paid-leave/balance");
}