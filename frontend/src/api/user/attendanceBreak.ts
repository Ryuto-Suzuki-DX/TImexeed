import { apiPost } from "@/api/client";
import type { SearchAttendanceBreaksRequest, SearchAttendanceBreaksResponse } from "@/types/user/attendanceBreak";

/*
 * 休憩検索
 *
 * POST /user/attendance-breaks/search
 */
export function searchAttendanceBreaks(request: SearchAttendanceBreaksRequest) {
  return apiPost<SearchAttendanceBreaksResponse, SearchAttendanceBreaksRequest>("/user/attendance-breaks/search", request);
}