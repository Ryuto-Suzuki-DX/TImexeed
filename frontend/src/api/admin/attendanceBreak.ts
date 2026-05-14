import { apiPost } from "@/api/client";
import type {
  SearchAttendanceBreaksRequest,
  SearchAttendanceBreaksResponse,
} from "@/types/admin/attendanceBreak";

/*
 * 管理者 休憩検索
 *
 * POST /admin/attendance-breaks/search
 */
export function searchAttendanceBreaks(request: SearchAttendanceBreaksRequest) {
  return apiPost<SearchAttendanceBreaksResponse, SearchAttendanceBreaksRequest>(
    "/admin/attendance-breaks/search",
    request
  );
}
