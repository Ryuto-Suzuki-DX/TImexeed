import { apiPost } from "@/api/client";
import type {
  UpdateMonthlyAttendanceSaveRequest,
  UpdateMonthlyAttendanceSaveResponse,
} from "@/types/admin/monthlyAttendanceSave";

/*
 * 管理者 月次勤怠全体保存
 *
 * POST /admin/monthly-attendances/update
 *
 * 保存対象：
 * ・日別勤怠
 * ・日別交通費
 * ・日別休憩
 * ・月次通勤定期
 *
 * 注意：
 * ・管理者APIでは targetUserId を送る
 * ・月次申請状態に関係なく保存できる
 * ・日別交通費は各勤怠日の transportExpenses 配列で送る
 */
export function updateMonthlyAttendanceSave(
  request: UpdateMonthlyAttendanceSaveRequest
) {
  return apiPost<
    UpdateMonthlyAttendanceSaveResponse,
    UpdateMonthlyAttendanceSaveRequest
  >("/admin/monthly-attendances/update", request);
}
