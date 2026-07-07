"use client";

import { FormEvent, useMemo, useState } from "react";
import { downloadMonthlyAttendanceSummaryExport } from "@/api/admin/monthlyAttendanceSummaryExport";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import styles from "./page.module.css";

type PageMessage = {
  variant: "info" | "success" | "warning" | "error";
  text: string;
};

type ExportFormat = "CSV" | "XLSX";

type ExportFormState = {
  targetMonth: string;
  keyword: string;
  includeNotApproved: boolean;
};

const initialExportForm: ExportFormState = {
  targetMonth: getCurrentMonthText(),
  keyword: "",
  includeNotApproved: true,
};

export default function AdminMonthlyAttendanceSummaryExportsPage() {
  const { user, isLoading, message: authMessage } = useRequireRole("ADMIN");

  const [exportForm, setExportForm] = useState<ExportFormState>(initialExportForm);
  const [isExporting, setIsExporting] = useState(false);
  const [pageMessage, setPageMessage] = useState<PageMessage>({
    variant: "info",
    text: "対象月を選択して、月次勤怠集計をCSVまたはExcelで出力できます。",
  });

  const targetMonthText = useMemo(() => {
    if (!exportForm.targetMonth) {
      return "未選択";
    }

    const [year, month] = exportForm.targetMonth.split("-");
    return `${year}年${Number(month)}月`;
  }, [exportForm.targetMonth]);

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="月次勤怠集計出力" description="ログイン情報を確認しています。" />
          <MessageBox variant="info">{authMessage}</MessageBox>
        </section>
      </PageContainer>
    );
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await handleExport("XLSX");
  }

  async function handleExport(format: ExportFormat) {
    const validationMessage = validateExportForm(exportForm);
    if (validationMessage) {
      setPageMessage({
        variant: "warning",
        text: validationMessage,
      });
      return;
    }

    const [targetYearText, targetMonthTextValue] = exportForm.targetMonth.split("-");
    const targetYear = Number(targetYearText);
    const targetMonth = Number(targetMonthTextValue);

    const formatLabel = format === "XLSX" ? "Excel" : "CSV";

    setIsExporting(true);
    setPageMessage({
      variant: "info",
      text: `月次勤怠集計${formatLabel}を出力しています。`,
    });

    try {
      await downloadMonthlyAttendanceSummaryExport({
        targetYear,
        targetMonth,
        targetUserIds: [],
        departmentId: null,
        keyword: exportForm.keyword.trim(),
        includeNotApproved: exportForm.includeNotApproved,
        format,
      });

      setPageMessage({
        variant: "success",
        text: `月次勤怠集計${formatLabel}を出力しました。`,
      });
    } catch (error) {
      setPageMessage({
        variant: "error",
        text: error instanceof Error ? error.message : `月次勤怠集計${formatLabel}の出力に失敗しました。`,
      });
    } finally {
      setIsExporting(false);
    }
  }

  function handleReset() {
    setExportForm(initialExportForm);
    setPageMessage({
      variant: "info",
      text: "出力条件を初期状態に戻しました。",
    });
  }

  return (
    <PageContainer>
      <AdminSideMenu />

      <div className={styles.pageWrap}>
        <section className={styles.pageCard}>
          <div className={styles.headerArea}>
            <PageTitle
              title="月次勤怠集計出力"
              description="承認済みの月次勤怠を集計し、CSVまたは提出用Excelとして出力します。"
            />

            <MessageBox variant={pageMessage.variant}>{pageMessage.text}</MessageBox>
          </div>

          <div className={styles.contentGrid}>
            <section className={styles.formCard}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>出力条件</h2>
                  <p className={styles.sectionDescription}>
                    基本は対象月だけ選択すれば出力できます。必要に応じて従業員名またはメールアドレスで絞り込みます。
                  </p>
                </div>
              </div>

              <form className={styles.exportForm} onSubmit={handleSubmit}>
                <label className={styles.fieldLabel}>
                  対象月

                  <span className={styles.monthPicker}>
                    <span className={styles.monthPickerValue}>
                      {formatMonthPickerLabel(exportForm.targetMonth)}
                    </span>

                    <span className={styles.monthPickerIcon} aria-hidden="true">
                      ▾
                    </span>

                    <input
                      className={styles.monthPickerInput}
                      type="month"
                      value={exportForm.targetMonth}
                      onChange={(event) =>
                        setExportForm((current) => ({
                          ...current,
                          targetMonth: event.target.value,
                        }))
                      }
                      aria-label="対象月を選択"
                    />
                  </span>
                </label>

                <label className={styles.fieldLabel}>
                  従業員キーワード
                  <input
                    className={styles.input}
                    type="text"
                    value={exportForm.keyword}
                    onChange={(event) =>
                      setExportForm((current) => ({
                        ...current,
                        keyword: event.target.value,
                      }))
                    }
                    placeholder="名前またはメールアドレス。未入力なら全員"
                  />
                </label>

                <label className={styles.checkboxLabel}>
                  <input
                    type="checkbox"
                    checked={exportForm.includeNotApproved}
                    onChange={(event) =>
                      setExportForm((current) => ({
                        ...current,
                        includeNotApproved: event.target.checked,
                      }))
                    }
                  />
                  <span>未承認・未申請の従業員もステータスのみ出力に含める</span>
                </label>

                <div className={styles.formActions}>
                  <Button type="submit" variant="primary" disabled={isExporting}>
                    {isExporting ? "出力中..." : "Excel出力"}
                  </Button>

                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => void handleExport("CSV")}
                    disabled={isExporting}
                  >
                    CSV出力
                  </Button>

                  <Button type="button" variant="secondary" onClick={handleReset} disabled={isExporting}>
                    条件をクリア
                  </Button>
                </div>
              </form>
            </section>

            <section className={styles.summaryCard}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>出力内容</h2>
                  <p className={styles.sectionDescription}>
                    Excelは提出用の見やすい表形式、CSVは加工・連携用のデータ形式として出力します。
                  </p>
                </div>
              </div>

              <div className={styles.summaryList}>
                <div className={styles.summaryItem}>
                  <span className={styles.summaryLabel}>対象月</span>
                  <span className={styles.summaryValue}>{targetMonthText}</span>
                </div>

                <div className={styles.summaryItem}>
                  <span className={styles.summaryLabel}>従業員条件</span>
                  <span className={styles.summaryValue}>
                    {exportForm.keyword.trim() ? exportForm.keyword.trim() : "全員"}
                  </span>
                </div>

                <div className={styles.summaryItem}>
                  <span className={styles.summaryLabel}>未承認者</span>
                  <span className={styles.summaryValue}>
                    {exportForm.includeNotApproved ? "ステータスのみ含める" : "含めない"}
                  </span>
                </div>
              </div>

              <div className={styles.noticeBox}>
                <h3 className={styles.noticeTitle}>集計ルール</h3>
                <ul className={styles.noticeList}>
                  <li>承認済み以外は勤怠・給与・交通費・有給・経費の集計値を出力しません。</li>
                  <li>残業は日別超過と週超過を重複しないように分けて集計します。</li>
                  <li>深夜労働は22:00〜翌5:00を休憩除外で集計します。</li>
                  <li>休日出勤は残業とは別枠で集計します。</li>
                  <li>Excel出力では、警告行や合計行を見やすく装飾します。</li>
                </ul>
              </div>
            </section>
          </div>
        </section>
      </div>
    </PageContainer>
  );
}

function validateExportForm(form: ExportFormState) {
  if (!form.targetMonth) {
    return "対象月を選択してください。";
  }

  const [yearText, monthText] = form.targetMonth.split("-");
  const year = Number(yearText);
  const month = Number(monthText);

  if (!year || !month || month < 1 || month > 12) {
    return "対象月の形式が正しくありません。";
  }

  return null;
}

function getCurrentMonthText() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");

  return `${year}-${month}`;
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
