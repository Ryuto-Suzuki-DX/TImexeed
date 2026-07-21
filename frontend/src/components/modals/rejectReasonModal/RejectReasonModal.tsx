"use client";

import type { MouseEvent } from "react";
import Button from "@/components/atoms/Button";
import styles from "./RejectReasonModal.module.css";

type RejectReasonModalProps = {
  open: boolean;
  reason: string;
  isSubmitting?: boolean;
  title?: string;
  description?: string;
  confirmLabel?: string;
  onChangeReason: (value: string) => void;
  onCancel: () => void;
  onConfirm: () => void;
};

export default function RejectReasonModal({
  open,
  reason,
  isSubmitting = false,
  title = "月次申請を否認",
  description = "否認理由を入力してください。",
  confirmLabel = "否認する",
  onChangeReason,
  onCancel,
  onConfirm,
}: RejectReasonModalProps) {
  if (!open) {
    return null;
  }

  const handleBackdropMouseDown = () => {
    if (!isSubmitting) {
      onCancel();
    }
  };

  const handleDialogMouseDown = (event: MouseEvent<HTMLDivElement>) => {
    event.stopPropagation();
  };

  const handleConfirm = () => {
    if (isSubmitting || reason.trim() === "") {
      return;
    }

    onConfirm();
  };

  return (
    <div
      className={styles.backdrop}
      role="presentation"
      onMouseDown={handleBackdropMouseDown}
    >
      <div
        className={styles.dialog}
        role="dialog"
        aria-modal="true"
        aria-labelledby="reject-reason-modal-title"
        onMouseDown={handleDialogMouseDown}
      >
        <div className={styles.header}>
          <h2 id="reject-reason-modal-title" className={styles.title}>
            {title}
          </h2>
          <p className={styles.description}>{description}</p>
        </div>

        <label className={styles.field}>
          <span className={styles.label}>否認理由</span>

          <textarea
            className={styles.textarea}
            value={reason}
            rows={6}
            maxLength={1000}
            placeholder="否認理由を入力してください"
            disabled={isSubmitting}
            autoFocus
            onChange={(event) => onChangeReason(event.target.value)}
          />

          <span className={styles.counter}>{reason.length} / 1000</span>
        </label>

        <div className={styles.actions}>
          <Button
            type="button"
            variant="secondary"
            disabled={isSubmitting}
            onClick={onCancel}
          >
            キャンセル
          </Button>

          <Button
            type="button"
            variant="danger"
            disabled={isSubmitting || reason.trim() === ""}
            onClick={handleConfirm}
          >
            {isSubmitting ? "処理中..." : confirmLabel}
          </Button>
        </div>
      </div>
    </div>
  );
}
