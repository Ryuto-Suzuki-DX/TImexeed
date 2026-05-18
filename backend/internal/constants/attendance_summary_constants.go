package constants

import "time"

/*
 * 月次勤怠集計CSV用 Constants
 *
 * 注意：
 * ・給与計算そのものは行わない
 * ・CSV集計で使う固定基準値のみをここにまとめる
 * ・派遣先所定労働時間は AttendanceDay.ScheduledWorkMinutes を使う
 * ・変形労働制フラグは持たない
 */

/*
 * 社内の日次所定労働時間（分）
 *
 * 8時間 = 480分。
 *
 * 日別残業の判定では、
 * max(CompanyDailyStandardWorkMinutes, AttendanceDay.ScheduledWorkMinutes)
 * をその日の残業判定基準にする。
 */
const CompanyDailyStandardWorkMinutes = 8 * 60

/*
 * 社内の週次所定労働時間（分）
 *
 * 40時間 = 2400分。
 *
 * 週別残業の判定では、
 * max(CompanyWeeklyStandardWorkMinutes, その週の ScheduledWorkMinutes 合計)
 * をその週の残業判定基準にする。
 */
const CompanyWeeklyStandardWorkMinutes = 40 * 60

/*
 * 深夜労働開始時刻
 *
 * 22:00。
 */
const LateNightWorkStartHour = 22

/*
 * 深夜労働終了時刻
 *
 * 翌 05:00。
 */
const LateNightWorkEndHour = 5

/*
 * 深夜労働開始時刻（1日の開始からの分）
 *
 * 22:00 = 1320分。
 */
const LateNightWorkStartMinuteOfDay = LateNightWorkStartHour * 60

/*
 * 深夜労働終了時刻（1日の開始からの分）
 *
 * 05:00 = 300分。
 */
const LateNightWorkEndMinuteOfDay = LateNightWorkEndHour * 60

/*
 * 週起算日
 *
 * 月曜起算。
 *
 * 月をまたぐ週でも、月曜〜日曜の1週間単位で
 * 週別残業・休日出勤判定を行う。
 */
const AttendanceSummaryWeekStartDay = time.Monday
