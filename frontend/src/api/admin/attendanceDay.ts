import { apiPost } from "@/api/client";
import type {
  SearchAttendanceDaysRequest,
  SearchAttendanceDaysResponse,
} from "@/types/admin/attendanceDay";

/*
 * 管理者 勤怠日検索
 *
 * POST /admin/attendance-days/search
 */
export function searchAttendanceDays(request: SearchAttendanceDaysRequest) {
  return apiPost<SearchAttendanceDaysResponse, SearchAttendanceDaysRequest>(
    "/admin/attendance-days/search",
    request
  );
}
