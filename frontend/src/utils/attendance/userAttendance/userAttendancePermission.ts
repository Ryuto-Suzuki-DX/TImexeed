import type { AttendanceViewRow } from "@/types/user/attendanceView";

/*
 * 従業員勤怠 権限制御 Utility
 *
 * USER側専用。
 *
 * 管理者側では、申請中・承認済みでも編集できる可能性があるため、
 * ここはADMINと共通化しない。
 */

/*
 * 従業員側で勤怠行を編集できない状態か判定する
 *
 * バックエンド側も同じ思想でブロックしている。
 */
export function isUserAttendanceRowLocked(row: AttendanceViewRow) {
  return (
    row.monthlyStatus === "PENDING" ||
    row.monthlyStatus === "APPROVED" ||
    row.requestStatus === "PENDING" ||
    row.requestStatus === "APPROVED"
  );
}

/*
 * 従業員側で月次通勤定期を編集できない状態か判定する
 */
export function isUserMonthlyCommuterPassLocked(monthlyStatus: string) {
  return monthlyStatus === "PENDING" || monthlyStatus === "APPROVED";
}

/*
 * 従業員側で月次申請ボタンを押せない状態か判定する
 */
export function isUserMonthlySubmitDisabled(hasUnsubmittedRequest: boolean, monthlyStatus: string) {
  return hasUnsubmittedRequest || monthlyStatus === "PENDING" || monthlyStatus === "APPROVED";
}