"use client";

import type { ExportMonthlyAttendanceSummaryCsvRequest } from "@/types/admin/monthlyAttendanceSummaryExport";

export type MonthlyAttendanceSummaryExportFormat = "CSV" | "XLSX";

export type ExportMonthlyAttendanceSummaryRequest = ExportMonthlyAttendanceSummaryCsvRequest & {
  format?: MonthlyAttendanceSummaryExportFormat;
};

/*
 * 管理者 月次勤怠集計出力
 *
 * POST /admin/monthly-attendance-summary-exports/export
 *
 * 注意：
 * ・このAPIは通常のJSONレスポンスではなくファイルを返す
 * ・そのため apiPost は使わず、fetchでBlobとして受け取る
 * ・Authorization は localStorage の accessToken を使う
 * ・format が CSV の場合はCSVファイルを返す
 * ・format が XLSX の場合はExcelファイルを返す
 * ・format 未指定時は既存互換のためCSV扱い
 */
export async function exportMonthlyAttendanceSummary(
  request: ExportMonthlyAttendanceSummaryRequest
) {
  const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;

  if (!apiBaseUrl) {
    throw new Error("NEXT_PUBLIC_API_BASE_URL が設定されていません。");
  }

  const accessToken =
    typeof window !== "undefined" ? localStorage.getItem("accessToken") : null;

  if (!accessToken) {
    throw new Error("アクセストークンがありません。ログインし直してください。");
  }

  const exportFormat = normalizeExportFormat(request.format);

  const response = await fetch(
    `${apiBaseUrl}/admin/monthly-attendance-summary-exports/export`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify({
        ...request,
        format: exportFormat,
      }),
    }
  );

  if (!response.ok) {
    const errorText = await response.text();

    try {
      const errorJson = JSON.parse(errorText);
      throw new Error(
        errorJson.message || getDefaultExportErrorMessage(exportFormat)
      );
    } catch {
      throw new Error(errorText || getDefaultExportErrorMessage(exportFormat));
    }
  }

  const blob = await response.blob();
  const fileName = getFileNameFromContentDisposition(
    response.headers.get("Content-Disposition")
  );

  return {
    blob,
    fileName:
      fileName ||
      buildFallbackFileName({
        targetYear: request.targetYear,
        targetMonth: request.targetMonth,
        format: exportFormat,
      }),
  };
}

/*
 * ファイルダウンロード実行
 *
 * page.tsx 側では基本これを呼ぶ。
 * format に CSV / XLSX を指定する。
 */
export async function downloadMonthlyAttendanceSummaryExport(
  request: ExportMonthlyAttendanceSummaryRequest
) {
  const { blob, fileName } = await exportMonthlyAttendanceSummary(request);

  const url = window.URL.createObjectURL(blob);
  const link = document.createElement("a");

  link.href = url;
  link.download = fileName;
  document.body.appendChild(link);
  link.click();

  link.remove();
  window.URL.revokeObjectURL(url);
}

/*
 * 管理者 月次勤怠集計CSV出力
 *
 * 既存互換用。
 * 既に page.tsx 側などでこの関数を呼んでいる場合でも壊さない。
 */
export async function exportMonthlyAttendanceSummaryCsv(
  request: ExportMonthlyAttendanceSummaryCsvRequest
) {
  return exportMonthlyAttendanceSummary({
    ...request,
    format: "CSV",
  });
}

/*
 * CSVダウンロード実行
 *
 * 既存互換用。
 */
export async function downloadMonthlyAttendanceSummaryCsv(
  request: ExportMonthlyAttendanceSummaryCsvRequest
) {
  return downloadMonthlyAttendanceSummaryExport({
    ...request,
    format: "CSV",
  });
}

/*
 * Excelダウンロード実行
 *
 * 提出用の色付き・罫線付きExcelを出す場合はこれを呼ぶ。
 */
export async function downloadMonthlyAttendanceSummaryExcel(
  request: ExportMonthlyAttendanceSummaryCsvRequest
) {
  return downloadMonthlyAttendanceSummaryExport({
    ...request,
    format: "XLSX",
  });
}

/*
 * format 正規化
 *
 * 未指定・不正値はCSV扱いにする。
 */
function normalizeExportFormat(
  format: ExportMonthlyAttendanceSummaryRequest["format"]
): MonthlyAttendanceSummaryExportFormat {
  if (format === "XLSX") {
    return "XLSX";
  }

  return "CSV";
}

/*
 * Content-Disposition から filename を取り出す
 */
function getFileNameFromContentDisposition(contentDisposition: string | null) {
  if (!contentDisposition) {
    return "";
  }

  const utf8FileNameMatch = contentDisposition.match(/filename\*=UTF-8''([^;]+)/);
  if (utf8FileNameMatch?.[1]) {
    return decodeURIComponent(utf8FileNameMatch[1]);
  }

  const fileNameMatch = contentDisposition.match(/filename="?([^"]+)"?/);
  if (fileNameMatch?.[1]) {
    return fileNameMatch[1];
  }

  return "";
}

/*
 * Content-Disposition が取れなかった場合の保険ファイル名
 */
function buildFallbackFileName(params: {
  targetYear: number;
  targetMonth: number;
  format: MonthlyAttendanceSummaryExportFormat;
}) {
  const paddedMonth = String(params.targetMonth).padStart(2, "0");

  if (params.format === "XLSX") {
    return `monthly_attendance_summary_${params.targetYear}_${paddedMonth}.xlsx`;
  }

  return `monthly_attendance_summary_${params.targetYear}_${paddedMonth}.csv`;
}

/*
 * エラーメッセージ
 */
function getDefaultExportErrorMessage(format: MonthlyAttendanceSummaryExportFormat) {
  if (format === "XLSX") {
    return "月次勤怠集計Excelの出力に失敗しました。";
  }

  return "月次勤怠集計CSVの出力に失敗しました。";
}