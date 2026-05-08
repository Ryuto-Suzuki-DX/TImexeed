import { apiPost } from "@/api/client";
import type {
  SearchAttendanceTypesRequest,
  SearchAttendanceTypesResponse,
} from "@/types/user/attendanceType";

/*
 * 勤務区分マスタ検索
 *
 * POST /user/attendance-types/search
 */
export function searchAttendanceTypes(request: SearchAttendanceTypesRequest) {
  return apiPost<SearchAttendanceTypesResponse, SearchAttendanceTypesRequest>("/user/attendance-types/search", request);
}