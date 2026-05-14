"use client";

import styles from "./AttendanceLockedText.module.css";

type AttendanceLockedTextProps = {
  message?: string;
};

export default function AttendanceLockedText({
  message = "申請中または承認済みのため変更できません。",
}: AttendanceLockedTextProps) {
  return <p className={styles.lockedText}>{message}</p>;
}