import { apiPost } from "@/api/client";
import type {
  DeleteAttendanceDayRequest,
  DeleteAttendanceDayResponse,
  SearchAttendanceDaysRequest,
  SearchAttendanceDaysResponse,
} from "@/types/user/attendanceDay";

/*
 * 勤怠日検索
 *
 * POST /user/attendance-days/search
 */
export function searchAttendanceDays(request: SearchAttendanceDaysRequest) {
  return apiPost<SearchAttendanceDaysResponse, SearchAttendanceDaysRequest>("/user/attendance-days/search", request);
}

/*
 * 勤怠日削除
 *
 * POST /user/attendance-days/delete
 */
export function deleteAttendanceDay(request: DeleteAttendanceDayRequest) {
  return apiPost<DeleteAttendanceDayResponse, DeleteAttendanceDayRequest>("/user/attendance-days/delete", request);
}