/*
 * 従業員 祝日 Type
 *
 * バックエンドの SearchHolidayDatesRequest / SearchHolidayDatesResponse に対応する。
 *
 * 注意：
 * ・従業員側では祝日の登録、更新、削除は行わない
 * ・CSV取り込みは管理者側APIで行う
 * ・祝日マスタ自体は全ユーザー共通
 */

export type SearchHolidayDatesRequest = {
  targetYear: number;
  targetMonth: number;
};

export type HolidayDate = {
  id: number;
  holidayDate: string;
  holidayName: string;
};

export type SearchHolidayDatesResponse = {
  holidays: HolidayDate[];
};
