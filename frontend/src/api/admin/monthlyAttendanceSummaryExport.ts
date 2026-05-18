"use client";

import type { ExportMonthlyAttendanceSummaryCsvRequest } from "@/types/admin/monthlyAttendanceSummaryExport";

/*
 * 管理者 月次勤怠集計CSV出力
 *
 * POST /admin/monthly-attendance-summary-exports/export
 *
 * 注意：
 * ・このAPIは通常のJSONレスポンスではなくCSVファイルを返す
 * ・そのため apiPost は使わず、fetchでBlobとして受け取る
 * ・Authorization は localStorage の accessToken を使う
 */
export async function exportMonthlyAttendanceSummaryCsv(
  request: ExportMonthlyAttendanceSummaryCsvRequest
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

  const response = await fetch(
    `${apiBaseUrl}/admin/monthly-attendance-summary-exports/export`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(request),
    }
  );

  if (!response.ok) {
    const errorText = await response.text();

    try {
      const errorJson = JSON.parse(errorText);
      throw new Error(errorJson.message || "月次勤怠集計CSVの出力に失敗しました。");
    } catch {
      throw new Error(errorText || "月次勤怠集計CSVの出力に失敗しました。");
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
      `monthly_attendance_summary_${request.targetYear}_${String(
        request.targetMonth
      ).padStart(2, "0")}.csv`,
  };
}

/*
 * CSVダウンロード実行
 *
 * page.tsx 側では基本これを呼べばよい。
 */
export async function downloadMonthlyAttendanceSummaryCsv(
  request: ExportMonthlyAttendanceSummaryCsvRequest
) {
  const { blob, fileName } = await exportMonthlyAttendanceSummaryCsv(request);

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
