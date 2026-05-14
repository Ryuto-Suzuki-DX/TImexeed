import { apiPost } from "@/api/client";
import type {
  ImportHolidayDatesRequest,
  ImportHolidayDatesResponse,
  SearchHolidayDatesRequest,
  SearchHolidayDatesResponse,
} from "@/types/admin/holidayDate";

/*
 * 管理者 祝日CSVインポート
 *
 * POST /admin/holiday-dates/import
 *
 * CSVファイルの中身をフロント側で文字列として読み取り、
 * csvText に入れてバックエンドへ送信する。
 *
 * バックエンド側では、
 * ・既存の祝日データを物理削除
 * ・CSV内容から祝日データを全件登録
 * する。
 */
export function importHolidayDates(request: ImportHolidayDatesRequest) {
  return apiPost<ImportHolidayDatesResponse, ImportHolidayDatesRequest>(
    "/admin/holiday-dates/import",
    request
  );
}

/*
 * 管理者 祝日検索
 *
 * POST /admin/holiday-dates/search
 *
 * 管理者画面で、登録済み祝日を対象年月ごとに取得する。
 */
export function searchHolidayDates(request: SearchHolidayDatesRequest) {
  return apiPost<SearchHolidayDatesResponse, SearchHolidayDatesRequest>(
    "/admin/holiday-dates/search",
    request
  );
}