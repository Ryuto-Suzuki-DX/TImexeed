/*
 * 勤怠 日付・時刻変換 Utility
 *
 * USER/ADMIN共通で使う想定。
 */

const WEEKDAYS = ["日", "月", "火", "水", "木", "金", "土"];

/*
 * 現在年月を input type="month" 用の yyyy-MM で返す
 */
export function getCurrentMonth() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");

  return `${year}-${month}`;
}

/*
 * yyyy-MM を targetYear / targetMonth に分解する
 */
export function parseTargetMonth(targetMonth: string) {
  const [yearText, monthText] = targetMonth.split("-");

  return {
    targetYear: Number(yearText),
    targetMonthValue: Number(monthText),
  };
}

/*
 * targetYear / targetMonth から yyyy-MM を作る
 */
export function buildTargetMonth(targetYear: number, targetMonth: number) {
  return `${targetYear}-${String(targetMonth).padStart(2, "0")}`;
}

/*
 * 対象月の日数を返す
 */
export function getDaysInMonth(targetYear: number, targetMonth: number) {
  return new Date(targetYear, targetMonth, 0).getDate();
}

/*
 * Date / RFC3339文字列から yyyy-MM-dd だけ取り出す
 */
export function toDateOnly(value: string) {
  return value.slice(0, 10);
}

/*
 * RFC3339文字列から HH:mm を作る
 */
export function toTimeText(value: string | null) {
  if (!value) {
    return "";
  }

  const date = new Date(value);
  const hours = String(date.getHours()).padStart(2, "0");
  const minutes = String(date.getMinutes()).padStart(2, "0");

  return `${hours}:${minutes}`;
}

/*
 * 画面表示用の日付ラベルを作る
 */
export function buildDayLabel(targetMonth: number, day: number) {
  return `${targetMonth}/${day}`;
}

/*
 * 曜日ラベルを作る
 */
export function buildWeekdayLabel(targetYear: number, targetMonth: number, day: number) {
  const date = new Date(targetYear, targetMonth - 1, day);

  return WEEKDAYS[date.getDay()];
}

/*
 * yyyy-MM-dd の翌日を返す
 */
export function addOneDay(dateText: string) {
  const [year, month, day] = dateText.split("-").map(Number);
  const date = new Date(year, month - 1, day + 1);

  const nextYear = date.getFullYear();
  const nextMonth = String(date.getMonth() + 1).padStart(2, "0");
  const nextDay = String(date.getDate()).padStart(2, "0");

  return `${nextYear}-${nextMonth}-${nextDay}`;
}

/*
 * 終了時刻が開始時刻以下なら日跨ぎとみなす
 *
 * 例：
 * 22:00 - 07:00 は翌日終了
 */
export function shouldUseNextDay(startTime: string, endTime: string) {
  if (!startTime || !endTime) {
    return false;
  }

  return endTime <= startTime;
}

/*
 * workDate + HH:mm から RFC3339 文字列を作る
 *
 * バックエンドは RFC3339 を要求しているため、
 * 保存時は必ずこの形式にする。
 */
export function toRfc3339(workDate: string, timeText: string, useNextDay: boolean) {
  if (!timeText) {
    return null;
  }

  const dateText = useNextDay ? addOneDay(workDate) : workDate;

  return `${dateText}T${timeText}:00+09:00`;
}

/*
 * 空文字を null にする
 */
export function toNullableString(value: string) {
  const trimmed = value.trim();

  return trimmed === "" ? null : trimmed;
}

/*
 * 空文字を null にする。
 * 数字でなければ null にする。
 */
export function toNullableNumber(value: string) {
  const trimmed = value.trim();

  if (trimmed === "") {
    return null;
  }

  const numberValue = Number(trimmed);

  return Number.isNaN(numberValue) ? null : numberValue;
}