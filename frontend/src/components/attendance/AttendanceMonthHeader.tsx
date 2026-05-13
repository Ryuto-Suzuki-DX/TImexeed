"use client";

import Button from "@/components/atoms/Button";
import Input from "@/components/atoms/Input";
import PageTitle from "@/components/atoms/PageTitle";
import styles from "@/app/user/attendance/page.module.css";

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
        description="対象月の予定、実績、休憩、交通費、申請状態を管理します。"
      />

      <div className={styles.monthControl}>
        <Input
          label="対象月"
          type="month"
          value={targetMonth}
          onChange={(event) => onChangeMonth(event.target.value)}
          className={styles.monthInput}
        />

        <Button type="button" variant="secondary" onClick={onPreviousMonth}>
          前月
        </Button>

        <Button type="button" variant="secondary" onClick={onNextMonth}>
          次月
        </Button>

        <Button type="button" variant="secondary" onClick={onSaveAll} disabled={saveDisabled}>
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
          <Button type="button" onClick={onMonthlySubmit} disabled={monthlySubmitDisabled}>
            月次申請する
          </Button>
        )}
      </div>
    </div>
  );
}