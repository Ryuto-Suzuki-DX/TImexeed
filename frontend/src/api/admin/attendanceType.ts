import { apiPost } from "@/api/client";
import type {
  SearchAttendanceTypesRequest,
  SearchAttendanceTypesResponse,
} from "@/types/admin/attendanceType";

/*
 * 管理者 勤務区分マスタ検索
 *
 * POST /admin/attendance-types/search
 */
export function searchAttendanceTypes(request: SearchAttendanceTypesRequest) {
  return apiPost<SearchAttendanceTypesResponse, SearchAttendanceTypesRequest>(
    "/admin/attendance-types/search",
    request
  );
}
