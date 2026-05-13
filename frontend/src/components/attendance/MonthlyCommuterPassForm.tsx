"use client";

import Button from "@/components/atoms/Button";
import Input from "@/components/atoms/Input";
import type { CommuterPassViewForm } from "@/types/user/attendanceView";
import styles from "@/app/user/attendance/page.module.css";

type MonthlyCommuterPassFormProps = {
  commuterPass: CommuterPassViewForm;
  disabled: boolean;
  onChange: <K extends keyof CommuterPassViewForm>(
    key: K,
    value: CommuterPassViewForm[K],
  ) => void;
  onReset: () => void;
};

export default function MonthlyCommuterPassForm({
  commuterPass,
  disabled,
  onChange,
  onReset,
}: MonthlyCommuterPassFormProps) {
  return (
    <section className={styles.commuterPassSection}>
      <div className={styles.sectionHeader}>
        <div>
          <h2 className={styles.sectionTitle}>月次通勤定期</h2>
          <p className={styles.sectionDescription}>
            対象月ごとの通勤定期情報です。全体保存でまとめて保存されます。
          </p>
        </div>

        <Button type="button" variant="secondary" onClick={onReset} disabled={disabled}>
          リセット
        </Button>
      </div>

      <div className={styles.commuterPassGrid}>
        <Input
          label="出発地"
          value={commuterPass.commuterFrom}
          onChange={(event) => onChange("commuterFrom", event.target.value)}
          disabled={disabled}
        />

        <Input
          label="目的地"
          value={commuterPass.commuterTo}
          onChange={(event) => onChange("commuterTo", event.target.value)}
          disabled={disabled}
        />

        <label className={styles.fieldLabel}>
          <span className={styles.fieldLabelText}>手段</span>
          <select
            aria-label="月次通勤定期の交通手段"
            value={commuterPass.commuterMethod}
            onChange={(event) => onChange("commuterMethod", event.target.value)}
            className={styles.select}
            disabled={disabled}
          >
            <option value="">選択してください</option>
            <option value="電車">電車</option>
            <option value="バス">バス</option>
            <option value="車">車</option>
            <option value="徒歩">徒歩</option>
            <option value="その他">その他</option>
          </select>
        </label>

        <Input
          label="金額"
          type="number"
          value={commuterPass.commuterAmount}
          onChange={(event) => onChange("commuterAmount", event.target.value)}
          disabled={disabled}
        />
      </div>
    </section>
  );
}