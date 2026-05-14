import { apiPost } from "@/api/client";
import type {
  SearchHolidayDatesRequest,
  SearchHolidayDatesResponse,
} from "@/types/user/holidayDate";

/*
 * 祝日検索
 *
 * POST /user/holiday-dates/search
 */
export function searchHolidayDates(request: SearchHolidayDatesRequest) {
  return apiPost<SearchHolidayDatesResponse, SearchHolidayDatesRequest>(
    "/user/holiday-dates/search",
    request
  );
}
