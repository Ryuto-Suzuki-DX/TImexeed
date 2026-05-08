"use client";

import { getStatusLabel } from "@/utils/attendance/attendanceStatus";
import styles from "@/app/user/attendance/page.module.css";

type AttendanceStatusBadgeProps = {
  status: string;
  requiresRequest?: boolean;
};

function getStatusClass(status: string, requiresRequest: boolean) {
  if (status === "PENDING") {
    return styles.statusPending;
  }

  if (status === "APPROVED") {
    return styles.statusApproved;
  }

  if (status === "REJECTED") {
    return styles.statusRejected;
  }

  if (requiresRequest && status !== "PENDING" && status !== "APPROVED") {
    return styles.statusRequiresRequest;
  }

  if (status === "DRAFT") {
    return styles.statusDraft;
  }

  return styles.statusNone;
}

export default function AttendanceStatusBadge({ status, requiresRequest = false }: AttendanceStatusBadgeProps) {
  return <span className={`${styles.statusBadge} ${getStatusClass(status, requiresRequest)}`}>{getStatusLabel(status)}</span>;
}