/*
 * 勤怠 ステータス表示 Utility
 *
 * USER/ADMIN共通で使える表示変換。
 *
 * 主な月次申請状態：
 * ・DRAFT    未申請
 * ・PENDING  申請中
 * ・APPROVED 承認済み
 * ・REJECTED 否認
 *
 * NONE は既存画面や日別表示の互換用として残す。
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