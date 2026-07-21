/*
 * 管理者 月次勤怠集計CSV / Excel出力 Type
 *
 * バックエンド：
 * POST /admin/monthly-attendance-summary-exports/export
 *
 * 注意：
 * ・レスポンスはJSONではなくファイル
 * ・通常の ApiResponse<T> は使わない
 */

export type MonthlyAttendanceSummaryExportTargetType =
  | "USER"
  | "DEPARTMENT";

export type MonthlyAttendanceSummaryExportFormat =
  | "CSV"
  | "XLSX";

/*
 * 月次勤怠集計出力 Request
 *
 * USER:
 * ・targetUserId に選択したユーザーIDを設定する
 *
 * DEPARTMENT:
 * ・departmentIds に選択した所属IDを複数設定できる
 * ・所属なしを含める場合は includeUnassignedDepartment を true にする
 */
export type ExportMonthlyAttendanceSummaryCsvRequest = {
  targetYear: number;
  targetMonth: number;

  targetType: MonthlyAttendanceSummaryExportTargetType;

  targetUserId?: number | null;

  departmentIds: number[];

  includeUnassignedDepartment: boolean;

  includeNotApproved: boolean;

  format: MonthlyAttendanceSummaryExportFormat;
};

/*
 * ファイル出力結果
 */
export type ExportMonthlyAttendanceSummaryCsvResult = {
  blob: Blob;
  fileName: string;
};

/*
 * 月次勤怠集計 出力状態
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
