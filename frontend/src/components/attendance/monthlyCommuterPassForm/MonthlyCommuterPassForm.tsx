"use client";

import Button from "@/components/atoms/Button";
import Input from "@/components/atoms/Input";
import styles from "./MonthlyCommuterPassForm.module.css";

/*
 * Admin/User共用コンポーネントのため、
 * Admin側・User側どちらか一方のattendanceView型には依存させない。
 *
 * 両画面の親コンポーネントが、この構造と同じ通勤定期Rowを渡す。
 */
export type MonthlyCommuterPassFormRow = {
  monthlyCommuterPassId: number | null;

  commuterFrom: string;
  commuterTo: string;
  commuterMethod: string;
  commuterAmount: string;

  isNew: boolean;
  isDirty: boolean;
};

type MonthlyCommuterPassFormProps = {
  commuterPasses: MonthlyCommuterPassFormRow[];
  disabled: boolean;

  onChange: <K extends keyof MonthlyCommuterPassFormRow>(
    index: number,
    key: K,
    value: MonthlyCommuterPassFormRow[K],
  ) => void;

  onAdd: () => void;
  onRemove: (index: number) => void;
  onReset: () => void;
};

function parseAmount(value: string): number {
  const normalized = value.trim();

  if (normalized === "") {
    return 0;
  }

  const parsed = Number(normalized);

  if (!Number.isFinite(parsed)) {
    return 0;
  }

  return parsed;
}

export default function MonthlyCommuterPassForm({
  commuterPasses,
  disabled,
  onChange,
  onAdd,
  onRemove,
  onReset,
}: MonthlyCommuterPassFormProps) {
  const totalCommuterAmount = commuterPasses.reduce(
    (total, commuterPass) =>
      total + parseAmount(commuterPass.commuterAmount),
    0,
  );

  return (
    <section className={styles.commuterPassSection}>
      <div className={styles.sectionHeader}>
        <div className={styles.sectionHeadingArea}>
          <h2 className={styles.sectionTitle}>月次通勤定期</h2>

          <p className={styles.sectionDescription}>
            対象月ごとの通勤定期情報です。複数件登録でき、全体保存でまとめて保存されます。
          </p>
        </div>

        <div className={styles.headerActions}>
          <Button
            type="button"
            variant="secondary"
            onClick={onAdd}
            disabled={disabled}
          >
            定期を追加
          </Button>

          <Button
            type="button"
            variant="secondary"
            onClick={onReset}
            disabled={disabled}
          >
            リセット
          </Button>
        </div>
      </div>

      <div className={styles.summaryRow}>
        <span className={styles.summaryLabel}>登録件数</span>

        <strong className={styles.summaryValue}>
          {commuterPasses.length}件
        </strong>

        <span className={styles.summaryLabel}>定期代合計</span>

        <strong className={styles.summaryValue}>
          {totalCommuterAmount.toLocaleString("ja-JP")}円
        </strong>
      </div>

      {commuterPasses.length === 0 ? (
        <div className={styles.emptyState}>
          <p className={styles.emptyStateText}>
            月次通勤定期は登録されていません。
          </p>

          <Button
            type="button"
            variant="secondary"
            onClick={onAdd}
            disabled={disabled}
          >
            最初の定期を追加
          </Button>
        </div>
      ) : (
        <div className={styles.commuterPassList}>
          {commuterPasses.map((commuterPass, index) => (
            <article
              key={
                commuterPass.monthlyCommuterPassId !== null
                  ? `commuter-pass-${commuterPass.monthlyCommuterPassId}`
                  : `new-commuter-pass-${index}`
              }
              className={styles.commuterPassCard}
            >
              <div className={styles.cardHeader}>
                <div className={styles.cardTitleArea}>
                  <h3 className={styles.cardTitle}>
                    定期 {index + 1}
                  </h3>

                  {commuterPass.monthlyCommuterPassId !== null ? (
                    <span className={styles.savedBadge}>保存済み</span>
                  ) : (
                    <span className={styles.newBadge}>新規</span>
                  )}
                </div>

                <Button
                  type="button"
                  variant="danger"
                  onClick={() => onRemove(index)}
                  disabled={disabled}
                >
                  削除
                </Button>
              </div>

              <div className={styles.commuterPassGrid}>
                <Input
                  label="出発地"
                  value={commuterPass.commuterFrom}
                  onChange={(event) =>
                    onChange(
                      index,
                      "commuterFrom",
                      event.target.value,
                    )
                  }
                  disabled={disabled}
                />

                <Input
                  label="目的地"
                  value={commuterPass.commuterTo}
                  onChange={(event) =>
                    onChange(
                      index,
                      "commuterTo",
                      event.target.value,
                    )
                  }
                  disabled={disabled}
                />

                <label className={styles.fieldLabel}>
                  <span className={styles.fieldLabelText}>
                    手段
                  </span>

                  <select
                    aria-label={`月次通勤定期${index + 1}の交通手段`}
                    value={commuterPass.commuterMethod}
                    onChange={(event) =>
                      onChange(
                        index,
                        "commuterMethod",
                        event.target.value,
                      )
                    }
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
                  min="0"
                  step="1"
                  value={commuterPass.commuterAmount}
                  onChange={(event) =>
                    onChange(
                      index,
                      "commuterAmount",
                      event.target.value,
                    )
                  }
                  disabled={disabled}
                />
              </div>

              <div className={styles.cardAmountRow}>
                <span className={styles.cardAmountLabel}>
                  この定期の金額
                </span>

                <strong className={styles.cardAmountValue}>
                  {parseAmount(
                    commuterPass.commuterAmount,
                  ).toLocaleString("ja-JP")}
                  円
                </strong>
              </div>
            </article>
          ))}
        </div>
      )}

      {commuterPasses.length > 0 ? (
        <div className={styles.footerActionArea}>
          <Button
            type="button"
            variant="secondary"
            onClick={onAdd}
            disabled={disabled}
          >
            定期を追加
          </Button>
        </div>
      ) : null}
    </section>
  );
}