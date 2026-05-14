/*
 * 管理者 祝日 Type
 *
 * バックエンドの admin/types/holiday_date.go に対応する。
 *
 * 役割：
 * ・祝日CSVインポート
 * ・対象年月の祝日検索
 * ・祝日一覧表示
 */

export type ImportHolidayDatesRequest = {
  csvText: string;
};

export type SearchHolidayDatesRequest = {
  targetYear: number;
  targetMonth: number;
};

export type HolidayDate = {
  id: number;
  holidayDate: string;
  holidayName: string;
  createdAt: string;
  updatedAt: string;
};

export type ImportHolidayDatesResponse = {
  deletedCount: number;
  importedCount: number;
  skippedCount: number;
};

export type SearchHolidayDatesResponse = {
  holidays: HolidayDate[];
};