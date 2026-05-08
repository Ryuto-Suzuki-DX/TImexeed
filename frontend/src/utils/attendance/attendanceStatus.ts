/*
 * 勤怠 ステータス表示 Utility
 *
 * USER/ADMIN共通で使える表示変換。
 */

export function getStatusLabel(status: string) {
  if (status === "PENDING") {
    return "申請中";
  }

  if (status === "APPROVED") {
    return "承認済み";
  }

  if (status === "REJECTED") {
    return "否認";
  }

  if (status === "DRAFT") {
    return "未申請";
  }

  if (status === "NONE") {
    return "なし";
  }

  return status || "なし";
}

export function isPendingStatus(status: string) {
  return status === "PENDING";
}

export function isApprovedStatus(status: string) {
  return status === "APPROVED";
}

export function isRejectedStatus(status: string) {
  return status === "REJECTED";
}

export function isDraftStatus(status: string) {
  return status === "DRAFT";
}

export function isNoneStatus(status: string) {
  return status === "NONE" || status === "";
}