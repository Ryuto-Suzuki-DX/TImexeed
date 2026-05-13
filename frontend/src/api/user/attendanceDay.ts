import { apiPost } from "@/api/client";
import type {
  SearchAttendanceDaysRequest,
  SearchAttendanceDaysResponse,
} from "@/types/user/attendanceDay";

/*
 * 勤怠日検索
 *
 * POST /user/attendance-days/search
 */
export function searchAttendanceDays(request: SearchAttendanceDaysRequest) {
  return apiPost<SearchAttendanceDaysResponse, SearchAttendanceDaysRequest>(
    "/user/attendance-days/search",
    request
  );
}