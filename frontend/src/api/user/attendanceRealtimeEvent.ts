import { apiPost } from "@/api/client";
import type {
  CreateAttendanceRealtimeEventRequest,
  CreateAttendanceRealtimeEventResponse,
  GetTodayAttendanceRealtimeEventsRequest,
  GetTodayAttendanceRealtimeEventsResponse,
} from "@/types/user/attendanceRealtimeEvent";

/*
 * 従業員 勤怠リアルタイムイベント作成
 *
 * POST /user/attendance-realtime-events/create
 */
export function createAttendanceRealtimeEvent(
  request: CreateAttendanceRealtimeEventRequest
) {
  return apiPost<
    CreateAttendanceRealtimeEventResponse,
    CreateAttendanceRealtimeEventRequest
  >("/user/attendance-realtime-events/create", request);
}

/*
 * 従業員 本日の勤怠リアルタイムイベント状態取得
 *
 * POST /user/attendance-realtime-events/today
 */
export function getTodayAttendanceRealtimeEvents(
  request: GetTodayAttendanceRealtimeEventsRequest
) {
  return apiPost<
    GetTodayAttendanceRealtimeEventsResponse,
    GetTodayAttendanceRealtimeEventsRequest
  >("/user/attendance-realtime-events/today", request);
}
