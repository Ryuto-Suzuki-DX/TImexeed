"use client";

import { useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import { importHolidayDates, searchHolidayDates } from "@/api/admin/holidayDate";
import type { HolidayDate } from "@/types/admin/holidayDate";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

const OFFICIAL_HOLIDAY_CSV_URL = "https://www8.cao.go.jp/chosei/shukujitsu/syukujitsu.csv";

function toDateInputValue(value: string | null | undefined) {
  if (!value) {
    return "";
  }

  return value.split("T")[0] ?? "";
}

function formatDate(value: string | null | undefined) {
  const dateValue = toDateInputValue(value);

  if (!dateValue) {
    return "-";
  }

  return dateValue.replaceAll("-", "/");
}

function getCurrentYear() {
  return new Date().getFullYear();
}

function getCurrentMonth() {
  return new Date().getMonth() + 1;
}

function getCsvPreview(csvText: string) {
  if (!csvText) {
    return "CSVファイルを選択すると、ここに内容の一部が表示されます。";
  }

  const lines = csvText.split(/\r?\n/).slice(0, 8);
  return lines.join("\n");
}

export default function AdminHolidayDatesPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [csvFileName, setCsvFileName] = useState("");
  const [csvText, setCsvText] = useState("");

  const [targetYear, setTargetYear] = useState(String(getCurrentYear()));
  const [targetMonth, setTargetMonth] = useState(String(getCurrentMonth()));

  const [holidays, setHolidays] = useState<HolidayDate[]>([]);

  const [pageMessage, setPageMessage] = useState(
    "内閣府の国民の祝日CSVを取り込み、登録済み祝日を対象年月ごとに確認できます。",
  );
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");
  const [isPageLoading, setIsPageLoading] = useState(false);

  const csvPreview = useMemo(() => getCsvPreview(csvText), [csvText]);

  const handleReadCsvFile = async (file: File | undefined) => {
    if (!file) {
      setCsvFileName("");
      setCsvText("");
      return;
    }

    setCsvFileName(file.name);

    try {
      const text = await file.text();

      setCsvText(text);
      setPageMessage("CSVファイルを読み込みました。内容を確認してから取り込みボタンを押してください。");
      setPageMessageVariant("success");
    } catch {
      setCsvFileName("");
      setCsvText("");
      setPageMessage("CSVファイルの読み込みに失敗しました。");
      setPageMessageVariant("error");
    }
  };

  const validateTargetYearMonth = () => {
    const year = Number(targetYear);
    const month = Number(targetMonth);

    if (!Number.isInteger(year) || year <= 0) {
      setPageMessage("対象年が正しくありません。");
      setPageMessageVariant("error");
      return null;
    }

    if (!Number.isInteger(month) || month < 1 || month > 12) {
      setPageMessage("対象月は1〜12で入力してください。");
      setPageMessageVariant("error");
      return null;
    }

    return {
      year,
      month,
    };
  };

  const handleSearchHolidayDates = async () => {
    const target = validateTargetYearMonth();

    if (!target) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("祝日一覧を取得しています。");
    setPageMessageVariant("info");

    const result = await searchHolidayDates({
      targetYear: target.year,
      targetMonth: target.month,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "祝日一覧の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    setHolidays(result.data.holidays);
    setPageMessage(result.data.holidays.length === 0 ? "対象年月の祝日は登録されていません。" : "祝日一覧を取得しました。");
    setPageMessageVariant(result.data.holidays.length === 0 ? "warning" : "success");
    setIsPageLoading(false);
  };

  const handleImportHolidayDates = async () => {
    if (!csvText.trim()) {
      setPageMessage("CSVファイルを選択してください。");
      setPageMessageVariant("warning");
      return;
    }

    const confirmed = window.confirm(
      "祝日CSVを取り込みます。既存の祝日データは削除され、CSVの内容で全件登録されます。よろしいですか？",
    );

    if (!confirmed) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("祝日CSVを取り込んでいます。");
    setPageMessageVariant("info");

    const result = await importHolidayDates({
      csvText,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "祝日CSVの取り込みに失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    const importData = result.data;

    setPageMessage(
      `祝日CSVを取り込みました。登録：${importData.importedCount}件 / スキップ：${importData.skippedCount}件 / 削除：${importData.deletedCount}件`,
    );
    setPageMessageVariant("success");
    setIsPageLoading(false);

    await handleSearchHolidayDates();
  };

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void handleSearchHolidayDates();
    }, 0);

    return () => {
      window.clearTimeout(timerId);
    };

    // 初回だけ実行したいので依存配列は固定する
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, user]);

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="祝日CSV管理" description="ログイン情報を確認しています。" />
          <MessageBox variant="info">{message}</MessageBox>
        </section>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <AdminSideMenu />

      <div className={styles.pageWrap}>
        <section className={styles.pageCard}>
          <div className={styles.headerArea}>
            <PageTitle title="祝日CSV管理" description="内閣府の国民の祝日CSVを取り込み、登録済み祝日を確認します。" />

            <MessageBox variant={pageMessageVariant}>{isPageLoading ? "処理中..." : pageMessage}</MessageBox>
          </div>

          <div className={styles.contentGrid}>
            <section className={styles.importPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>CSV取り込み</h2>
                  <p className={styles.sectionDescription}>
                    CSVファイルの内容をフロント側で文字列として読み取り、バックエンドへ送信します。
                  </p>
                </div>
              </div>

              <div className={styles.noticeBox}>
                <p className={styles.noticeTitle}>取り込み仕様</p>
                <p className={styles.noticeText}>
                  既存の祝日データを削除し、CSVの内容を全件登録します。差分更新は行いません。
                </p>
              </div>

              <div className={styles.linkCard}>
                <p className={styles.linkTitle}>公式CSV取得元</p>

                <p className={styles.linkDescription}>
                  内閣府が公開している「国民の祝日.csv」をダウンロードして、この画面から取り込んでください。
                </p>

                <a className={styles.officialLink} href={OFFICIAL_HOLIDAY_CSV_URL} target="_blank" rel="noreferrer">
                  内閣府 国民の祝日.csv を開く
                </a>
              </div>

              <div className={styles.guideCard}>
                <p className={styles.guideTitle}>操作手順</p>

                <ol className={styles.guideList}>
                  <li>「内閣府 国民の祝日.csv を開く」をクリックします。</li>
                  <li>表示されたCSVをファイルとして保存します。</li>
                  <li>この画面の「CSVファイル」から保存したCSVを選択します。</li>
                  <li>CSVプレビューで内容を確認します。</li>
                  <li>「取り込み」ボタンを押します。</li>
                  <li>右側の対象年月検索で、登録済み祝日を確認します。</li>
                </ol>

                <p className={styles.guideNote}>
                  注意：取り込み時は既存の祝日データを削除し、CSVの内容で全件登録し直します。
                </p>
              </div>

              <label className={styles.field}>
                <span className={styles.fieldLabel}>CSVファイル</span>
                <input
                  className={styles.fileInput}
                  type="file"
                  accept=".csv,text/csv"
                  onChange={(event) => {
                    void handleReadCsvFile(event.target.files?.[0]);
                  }}
                  disabled={isPageLoading}
                />
              </label>

              <div className={styles.fileInfo}>
                <span className={styles.fileInfoLabel}>選択中ファイル</span>
                <span className={styles.fileInfoValue}>{csvFileName || "未選択"}</span>
              </div>

              <div className={styles.previewCard}>
                <div className={styles.previewHeader}>
                  <p className={styles.previewTitle}>CSVプレビュー</p>
                  <p className={styles.previewDescription}>先頭8行まで表示します。</p>
                </div>

                <pre className={styles.csvPreview}>{csvPreview}</pre>
              </div>

              <div className={styles.actionArea}>
                <Button type="button" variant="primary" onClick={handleImportHolidayDates} disabled={isPageLoading || !csvText.trim()}>
                  取り込み
                </Button>
              </div>
            </section>

            <section className={styles.searchPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>登録済み祝日検索</h2>
                  <p className={styles.sectionDescription}>対象年月を指定して、現在登録されている祝日を確認します。</p>
                </div>
              </div>

              <div className={styles.searchHelpBox}>
                <p className={styles.searchHelpTitle}>検索条件</p>
                <p className={styles.searchHelpText}>
                  現在のAPIでは、祝日名や1日単位ではなく「対象年」と「対象月」で検索します。
                </p>
              </div>

              <div className={styles.searchForm}>
                <label className={styles.field}>
                  <span className={styles.fieldLabel}>対象年</span>
                  <input
                    className={styles.input}
                    type="number"
                    value={targetYear}
                    min={1}
                    onChange={(event) => setTargetYear(event.target.value)}
                    disabled={isPageLoading}
                  />
                </label>

                <label className={styles.field}>
                  <span className={styles.fieldLabel}>対象月</span>
                  <select
                    className={styles.input}
                    value={targetMonth}
                    onChange={(event) => setTargetMonth(event.target.value)}
                    disabled={isPageLoading}
                  >
                    {Array.from({ length: 12 }, (_, index) => {
                      const month = index + 1;

                      return (
                        <option key={month} value={String(month)}>
                          {month}月
                        </option>
                      );
                    })}
                  </select>
                </label>

                <div className={styles.searchButtonArea}>
                  <Button type="button" variant="primary" onClick={handleSearchHolidayDates} disabled={isPageLoading}>
                    検索
                  </Button>
                </div>
              </div>

              <div className={styles.tableArea}>
                <div className={styles.tableHeader}>
                  <div>
                    <h3 className={styles.tableTitle}>祝日一覧</h3>
                    <p className={styles.tableDescription}>
                      {targetYear}年{targetMonth}月に登録されている祝日です。
                    </p>
                  </div>
                </div>

                <div className={styles.tableScroll}>
                  <table className={styles.table}>
                    <thead>
                      <tr>
                        <th>日付</th>
                        <th>祝日名</th>
                        <th>登録日時</th>
                        <th>更新日時</th>
                      </tr>
                    </thead>

                    <tbody>
                      {holidays.map((holiday) => (
                        <tr key={holiday.id}>
                          <td>{formatDate(holiday.holidayDate)}</td>
                          <td>
                            <span className={styles.holidayName}>{holiday.holidayName}</span>
                          </td>
                          <td>{formatDate(holiday.createdAt)}</td>
                          <td>{formatDate(holiday.updatedAt)}</td>
                        </tr>
                      ))}

                      {holidays.length === 0 && (
                        <tr>
                          <td colSpan={4} className={styles.emptyCell}>
                            対象年月の祝日は登録されていません。
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </section>
          </div>
        </section>
      </div>
    </PageContainer>
  );
}