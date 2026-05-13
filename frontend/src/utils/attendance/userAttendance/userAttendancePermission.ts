/*
 * 従業員勤怠 権限制御 Utility
 *
 * USER側専用。
 *
 * 管理者側では、申請中・承認済みでも編集できる可能性があるため、
 * ここはADMINと共通化しない。
 */

/*
 * 従業員側で月次勤怠を編集できない状態か判定する
 *
 * 月次申請状態は MonthlyAttendanceRequest 側で管理する。
 * 日別勤怠や月次通勤定期のレコード状態では判定しない。
 *
 * バックエンド側も同じ思想でブロックしている。
 */
export function isUserAttendanceRowLocked(monthlyStatus: string) {
  return monthlyStatus === "PENDING" || monthlyStatus === "APPROVED";
}

/*
 * 従業員側で月次通勤定期を編集できない状態か判定する
 */
export function isUserMonthlyCommuterPassLocked(monthlyStatus: string) {
  return monthlyStatus === "PENDING" || monthlyStatus === "APPROVED";
}

/*
 * 従業員側で月次申請ボタンを押せない状態か判定する
 *
 * disabledになる条件：
 * ・未保存の変更がある
 * ・すでに月次申請中
 * ・すでに月次承認済み
 */
export function isUserMonthlySubmitDisabled(
  monthlyStatus: string,
  hasUnsavedChanges: boolean,
) {
  return hasUnsavedChanges || monthlyStatus === "PENDING" || monthlyStatus === "APPROVED";
}

/*
 * 従業員側で月次申請取り下げボタンを押せない状態か判定する
 *
 * disabledになる条件：
 * ・未保存の変更がある
 * ・月次申請中ではない
 */
export function isUserMonthlyWithdrawDisabled(
  monthlyStatus: string,
  hasUnsavedChanges: boolean,
) {
  return hasUnsavedChanges || monthlyStatus !== "PENDING";
}