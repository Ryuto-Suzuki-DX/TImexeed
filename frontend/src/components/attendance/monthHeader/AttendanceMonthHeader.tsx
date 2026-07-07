"use client";

import Button from "@/components/atoms/Button";
import PageTitle from "@/components/atoms/PageTitle";
import styles from "./AttendanceMonthHeader.module.css";

type AttendanceMonthHeaderProps = {
  targetMonth: string;
  monthlyStatus: string;
  monthlySubmitDisabled: boolean;
  monthlyWithdrawDisabled: boolean;
  saveDisabled: boolean;
  onChangeMonth: (value: string) => void;
  onPreviousMonth: () => void;
  onNextMonth: () => void;
  onSaveAll: () => void;
  onMonthlySubmit: () => void;
  onMonthlyWithdraw: () => void;
};

export default function AttendanceMonthHeader({
  targetMonth,
  monthlyStatus,
  monthlySubmitDisabled,
  monthlyWithdrawDisabled,
  saveDisabled,
  onChangeMonth,
  onPreviousMonth,
  onNextMonth,
  onSaveAll,
  onMonthlySubmit,
  onMonthlyWithdraw,
}: AttendanceMonthHeaderProps) {
  const showWithdrawButton = monthlyStatus === "PENDING";

  return (
    <div className={styles.header}>
      <PageTitle
        title="勤怠入力"
        description="対象月の予定、実績、休憩、交通費を管理します。"
      />

      <div className={styles.monthControl}>
        <label className={styles.monthField}>
          <span className={styles.monthLabel}>対象月</span>

          <span className={styles.monthPicker}>
            <span className={styles.monthPickerValue}>
              {formatMonthPickerLabel(targetMonth)}
            </span>

            <span className={styles.monthPickerIcon} aria-hidden="true">
              ▾
            </span>

            <input
              className={styles.monthPickerInput}
              type="month"
              value={targetMonth}
              onChange={(event) => onChangeMonth(event.target.value)}
              aria-label="対象月を選択"
            />
          </span>
        </label>

        <div className={styles.monthNavigation}>
          <Button
            type="button"
            variant="secondary"
            onClick={onPreviousMonth}
          >
            前月
          </Button>

          <Button
            type="button"
            variant="secondary"
            onClick={onNextMonth}
          >
            次月
          </Button>
        </div>

        <div className={styles.monthActions}>
          <Button
            type="button"
            variant="secondary"
            onClick={onSaveAll}
            disabled={saveDisabled}
          >
            全体保存
          </Button>

          {showWithdrawButton ? (
            <Button
              type="button"
              variant="danger"
              onClick={onMonthlyWithdraw}
              disabled={monthlyWithdrawDisabled}
            >
              申請取り下げ
            </Button>
          ) : (
            <Button
              type="button"
              onClick={onMonthlySubmit}
              disabled={monthlySubmitDisabled}
            >
              月次申請する
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}

function formatMonthPickerLabel(value: string) {
  const [yearText, monthText] = value.split("-");
  const year = Number(yearText);
  const month = Number(monthText);

  if (!year || !month) {
    return "月を選択";
  }

  return `${year}年${month}月`;
}
