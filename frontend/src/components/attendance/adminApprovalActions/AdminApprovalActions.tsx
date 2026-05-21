"use client";

import Button from "@/components/atoms/Button";
import { getStatusLabel } from "@/utils/attendance/attendanceStatus";
import styles from "./AdminApprovalActions.module.css";

type AdminApprovalActionsProps = {
  monthlyStatus: string;
  monthlyRequestId: number | null;
  disabled: boolean;
  canApprove: boolean;
  canReject: boolean;
  onApprove: () => void;
  onReject: () => void;
};

export default function AdminApprovalActions({
  monthlyStatus,
  monthlyRequestId,
  disabled,
  canApprove,
  canReject,
  onApprove,
  onReject,
}: AdminApprovalActionsProps) {
  const approveDisabled = disabled || monthlyRequestId === null || !canApprove;
  const rejectDisabled = disabled || monthlyRequestId === null || !canReject;
  const rejectButtonLabel = monthlyStatus === "APPROVED" ? "承認取消して否認" : "否認";

  return (
    <section className={styles.actionSection}>
      <div className={styles.sectionHeader}>
        <div>
          <h2 className={styles.sectionTitle}>月次申請 承認操作</h2>
          <p className={styles.sectionDescription}>
            申請中の月次申請は承認または否認できます。承認済みの月次申請は、承認情報を取り消して否認へ戻せます。
          </p>
        </div>

        <div className={styles.statusBox}>
          <p className={styles.statusLabel}>現在の状態</p>
          <p className={styles.statusValue}>{getStatusLabel(monthlyStatus)}</p>
        </div>
      </div>

      <div className={styles.actionControl}>
        <Button type="button" onClick={onApprove} disabled={approveDisabled}>
          承認
        </Button>

        <Button type="button" variant="danger" onClick={onReject} disabled={rejectDisabled}>
          {rejectButtonLabel}
        </Button>
      </div>

      <p className={styles.helpText}>
        承認・否認は、月次申請状態に応じて実行できます。
      </p>
    </section>
  );
}
