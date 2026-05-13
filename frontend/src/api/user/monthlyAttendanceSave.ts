import { apiPost } from "@/api/client";
import type {
  UpdateMonthlyAttendanceSaveRequest,
  UpdateMonthlyAttendanceSaveResponse,
} from "@/types/user/monthlyAttendanceSave";

/*
 * 月次勤怠全体保存
 *
 * POST /user/monthly-attendances/update
 *
 * 保存対象：
 * ・日別勤怠
 * ・日別休憩
 * ・月次通勤定期
 */
export function updateMonthlyAttendanceSave(request: UpdateMonthlyAttendanceSaveRequest) {
  return apiPost<UpdateMonthlyAttendanceSaveResponse, UpdateMonthlyAttendanceSaveRequest>(
    "/user/monthly-attendances/update",
    request
  );
}