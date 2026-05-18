/*
 * 管理者 月次勤怠集計CSV出力 Type
 *
 * バックエンド：
 * POST /admin/monthly-attendance-summary-exports/export
 *
 * 注意：
 * ・レスポンスはJSONではなくCSVファイル
 * ・通常の ApiResponse<T> は使わない
 */

/*
 * 月次勤怠集計CSV出力 Request
 *
 * targetUserIds:
 *   指定がある場合、そのユーザーだけを対象にする。
 *   空配列または未指定の場合は検索条件に一致するユーザーを対象にする。
 *
 * departmentId:
 *   指定がある場合、その所属のユーザーだけを対象にする。
 *
 * keyword:
 *   ユーザー名/メールアドレスのフリーワード検索。
 *
 * includeNotApproved:
 *   true の場合、APPROVED以外もステータスだけCSVへ出力する。
 *   false の場合、APPROVEDのみCSVへ出力する。
 */
export type ExportMonthlyAttendanceSummaryCsvRequest = {
  targetYear: number;
  targetMonth: number;
  targetUserIds?: number[];
  departmentId?: number | null;
  keyword?: string;
  includeNotApproved: boolean;
};

/*
 * CSV出力結果
 *
 * API関数内でBlobとファイル名を返すためのフロント専用型。
 */
export type ExportMonthlyAttendanceSummaryCsvResult = {
  blob: Blob;
  fileName: string;
};

/*
 * 月次勤怠集計CSV 出力状態
 */
export type MonthlyAttendanceSummaryCalculationStatus =
  | "CALCULATED"
  | "SKIPPED_NOT_APPROVED"
  | "ERROR";

/*
 * 月次勤怠ステータス
 */
export type MonthlyAttendanceSummaryMonthlyStatus =
  | "NOT_SUBMITTED"
  | "PENDING"
  | "APPROVED"
  | "REJECTED"
  | "CANCELED";
