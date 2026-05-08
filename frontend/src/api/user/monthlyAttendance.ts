import { apiPost } from "@/api/client";
import type {
  UpdateMonthlyAttendanceRequest,
  UpdateMonthlyAttendanceResponse,
} from "@/types/user/monthlyAttendance";

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
export function updateMonthlyAttendance(request: UpdateMonthlyAttendanceRequest) {
  return apiPost<UpdateMonthlyAttendanceResponse, UpdateMonthlyAttendanceRequest>("/user/monthly-attendances/update", request);
}