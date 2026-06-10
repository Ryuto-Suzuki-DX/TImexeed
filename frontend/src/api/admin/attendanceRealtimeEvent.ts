import { apiPost } from "@/api/client";
import type {
  SearchAttendanceRealtimeEventsRequest,
  SearchAttendanceRealtimeEventsResponse,
} from "@/types/admin/attendanceRealtimeEvent";

/*
 * 管理者 勤怠リアルタイムイベント検索
 *
 * POST /admin/attendance-realtime-events/search
 */
export function searchAttendanceRealtimeEvents(
  request: SearchAttendanceRealtimeEventsRequest
) {
  return apiPost<
    SearchAttendanceRealtimeEventsResponse,
    SearchAttendanceRealtimeEventsRequest
  >("/admin/attendance-realtime-events/search", request);
}
