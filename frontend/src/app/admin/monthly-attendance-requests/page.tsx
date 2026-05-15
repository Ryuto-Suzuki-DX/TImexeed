"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import {
  approveMonthlyAttendanceRequest,
  rejectMonthlyAttendanceRequest,
  searchMonthlyAttendanceRequests,
} from "@/api/admin/monthlyAttendanceRequest";
import type {
  MonthlyAttendanceRequestListRow,
  MonthlyAttendanceRequestStatus,
} from "@/types/admin/monthlyAttendanceRequest";
import {
  buildTargetMonth,
  getCurrentMonth,
  parseTargetMonth,
} from "@/utils/attendance/attendanceDate";
import { getStatusLabel } from "@/utils/attendance/attendanceStatus";
import type { AdminAttendanceInitialSearch } from "@/types/admin/adminAttendanceInitialSearch";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

type StatusOption = {
  status: MonthlyAttendanceRequestStatus;
  label: string;
};

const STATUS_OPTIONS: StatusOption[] = [
  { status: "NOT_SUBMITTED", label: "未申請" },
  { status: "PENDING", label: "申請中" },
  { status: "REJECTED", label: "否認済み" },
  { status: "APPROVED", label: "承認済み" },
  { status: "CANCELED", label: "取り下げ済み" },
];

const SEARCH_LIMIT = 20;

function formatDateTime(value: string | null | undefined) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value.replace("T", " ").slice(0, 16);
  }

  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  const hour = `${date.getHours()}`.padStart(2, "0");
  const minute = `${date.getMinutes()}`.padStart(2, "0");

  return `${year}/${month}/${day} ${hour}:${minute}`;
}

function formatTargetMonth(targetYear: number, targetMonth: number) {
  return `${targetYear}年${targetMonth}月`;
}

function getStatusClassName(status: MonthlyAttendanceRequestStatus) {
  switch (status) {
    case "NOT_SUBMITTED":
      return styles.statusBadgeNotSubmitted;
    case "PENDING":
      return styles.statusBadgePending;
    case "APPROVED":
      return styles.statusBadgeApproved;
    case "REJECTED":
      return styles.statusBadgeRejected;
    case "CANCELED":
      return styles.statusBadgeCanceled;
    default:
      return styles.statusBadge;
  }
}

export default function AdminMonthlyAttendanceRequestsPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [targetMonth, setTargetMonth] = useState(getCurrentMonth());
  const [keyword, setKeyword] = useState("");
  const [searchedKeyword, setSearchedKeyword] = useState("");

  const [selectedStatuses, setSelectedStatuses] = useState<MonthlyAttendanceRequestStatus[]>([
    "PENDING",
  ]);

  const [rows, setRows] = useState<MonthlyAttendanceRequestListRow[]>([]);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);

  const [pageMessage, setPageMessage] =
    useState("対象月と申請状態を指定して、月次申請を検索してください。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");
  const [isPageLoading, setIsPageLoading] = useState(false);

  const { targetYear, targetMonthValue } = useMemo(
    () => parseTargetMonth(targetMonth),
    [targetMonth],
  );

  const targetYearOptions = useMemo(() => {
    const currentYear = new Date().getFullYear();
    const years: number[] = [];

    for (let year = currentYear - 5; year <= currentYear + 1; year += 1) {
      years.push(year);
    }

    if (!years.includes(targetYear)) {
      years.push(targetYear);
      years.sort((a, b) => a - b);
    }

    return years;
  }, [targetYear]);

  const selectedStatusLabel = useMemo(() => {
    if (selectedStatuses.length === 0) {
      return "未選択";
    }

    return STATUS_OPTIONS.filter((option) => selectedStatuses.includes(option.status))
      .map((option) => option.label)
      .join(" / ");
  }, [selectedStatuses]);

  const searchMonthlyRequests = useCallback(
    async (nextOffset: number, append: boolean) => {
      if (selectedStatuses.length === 0) {
        setPageMessage("申請状態を1つ以上選択してください。");
        setPageMessageVariant("error");
        return;
      }

      setIsPageLoading(true);
      setPageMessage("月次申請一覧を検索しています。");
      setPageMessageVariant("info");

      const searchKeyword = append ? searchedKeyword : keyword.trim();

      try {
        const result = await searchMonthlyAttendanceRequests({
          keyword: searchKeyword,
          targetYear,
          targetMonth: targetMonthValue,
          statuses: selectedStatuses,
          includeDeletedUsers: false,
          offset: nextOffset,
          limit: SEARCH_LIMIT,
        });

        if (result.error || !result.data) {
          setPageMessage(result.message || "月次申請一覧の検索に失敗しました。");
          setPageMessageVariant("error");
          return;
        }

        const data = result.data;

        setSearchedKeyword(searchKeyword);
        setRows((currentRows) =>
          append
            ? [...currentRows, ...data.monthlyAttendanceRequests]
            : data.monthlyAttendanceRequests,
        );
        setOffset(nextOffset + data.monthlyAttendanceRequests.length);
        setHasMore(data.hasMore);

        if (data.monthlyAttendanceRequests.length === 0 && !append) {
          setPageMessage("該当する月次申請はありません。");
          setPageMessageVariant("warning");
          return;
        }

        setPageMessage("月次申請一覧を取得しました。");
        setPageMessageVariant("success");
      } catch (error) {
        setPageMessage(
          error instanceof Error
            ? error.message
            : "月次申請一覧の検索中に予期しないエラーが発生しました。",
        );
        setPageMessageVariant("error");
      } finally {
        setIsPageLoading(false);
      }
    },
    [keyword, searchedKeyword, selectedStatuses, targetMonthValue, targetYear],
  );

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void searchMonthlyRequests(0, false);
    }, 0);

    return () => {
      window.clearTimeout(timerId);
    };

    // 初回だけ自動検索したいので依存配列は固定する
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, user]);

  const handleSearch = () => {
    void searchMonthlyRequests(0, false);
  };

  const handleLoadMore = () => {
    void searchMonthlyRequests(offset, true);
  };

  const handleToggleStatus = (status: MonthlyAttendanceRequestStatus) => {
    setSelectedStatuses((currentStatuses) =>
      currentStatuses.includes(status)
        ? currentStatuses.filter((currentStatus) => currentStatus !== status)
        : [...currentStatuses, status],
    );
  };

  const handlePreviousMonth = () => {
    const previousMonthDate = new Date(targetYear, targetMonthValue - 2, 1);
    const year = previousMonthDate.getFullYear();
    const month = previousMonthDate.getMonth() + 1;

    setTargetMonth(buildTargetMonth(year, month));
  };

  const handleNextMonth = () => {
    const nextMonthDate = new Date(targetYear, targetMonthValue, 1);
    const year = nextMonthDate.getFullYear();
    const month = nextMonthDate.getMonth() + 1;

    setTargetMonth(buildTargetMonth(year, month));
  };

  const handleChangeTargetYear = (value: string) => {
    const nextYear = Number(value);

    if (Number.isNaN(nextYear)) {
      return;
    }

    setTargetMonth(buildTargetMonth(nextYear, targetMonthValue));
  };

  const handleChangeTargetMonth = (value: string) => {
    const nextMonth = Number(value);

    if (Number.isNaN(nextMonth)) {
      return;
    }

    setTargetMonth(buildTargetMonth(targetYear, nextMonth));
  };

  const handleOpenAttendancePage = (row: MonthlyAttendanceRequestListRow) => {
    const initialSearch: AdminAttendanceInitialSearch = {
      targetUserId: row.targetUserId,
      targetUserName: row.userName,
      targetYear: row.targetYear,
      targetMonth: row.targetMonth,
      monthlyRequestId: row.monthlyAttendanceRequest.id,
    };

    const initialKey = crypto.randomUUID();
    const storageKey = `adminAttendanceInitialSearch:${initialKey}`;

    localStorage.setItem(storageKey, JSON.stringify(initialSearch));

    window.open(`/admin/attendance?initialKey=${initialKey}`, "_blank", "noopener,noreferrer");
  };

  const handleApprove = async (row: MonthlyAttendanceRequestListRow) => {
    const targetRequestId = row.monthlyAttendanceRequest.id;

    if (!targetRequestId) {
      setPageMessage("承認対象の月次申請がありません。");
      setPageMessageVariant("error");
      return;
    }

    setIsPageLoading(true);
    setPageMessage("月次申請を承認しています。");
    setPageMessageVariant("info");

    try {
      const result = await approveMonthlyAttendanceRequest({
        targetRequestId,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "月次申請の承認に失敗しました。");
        setPageMessageVariant("error");
        return;
      }

      setPageMessage(result.message || "月次申請を承認しました。");
      setPageMessageVariant("success");

      await searchMonthlyRequests(0, false);
    } catch (error) {
      setPageMessage(
        error instanceof Error
          ? error.message
          : "月次申請の承認中に予期しないエラーが発生しました。",
      );
      setPageMessageVariant("error");
    } finally {
      setIsPageLoading(false);
    }
  };

  const handleReject = async (row: MonthlyAttendanceRequestListRow) => {
    const targetRequestId = row.monthlyAttendanceRequest.id;

    if (!targetRequestId) {
      setPageMessage("否認対象の月次申請がありません。");
      setPageMessageVariant("error");
      return;
    }

    const rejectedReason = window.prompt("否認理由を入力してください。");

    if (rejectedReason === null) {
      return;
    }

    if (rejectedReason.trim() === "") {
      setPageMessage("否認理由を入力してください。");
      setPageMessageVariant("error");
      return;
    }

    setIsPageLoading(true);
    setPageMessage("月次申請を否認しています。");
    setPageMessageVariant("info");

    try {
      const result = await rejectMonthlyAttendanceRequest({
        targetRequestId,
        rejectedReason: rejectedReason.trim(),
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "月次申請の否認に失敗しました。");
        setPageMessageVariant("error");
        return;
      }

      setPageMessage(result.message || "月次申請を否認しました。");
      setPageMessageVariant("success");

      await searchMonthlyRequests(0, false);
    } catch (error) {
      setPageMessage(
        error instanceof Error
          ? error.message
          : "月次申請の否認中に予期しないエラーが発生しました。",
      );
      setPageMessageVariant("error");
    } finally {
      setIsPageLoading(false);
    }
  };

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="月次申請確認" description="ログイン情報を確認しています。" />
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
            <PageTitle
              title="月次申請確認"
              description="対象月、ユーザー、申請状態で月次勤怠申請を確認し、申請中のものを承認・否認します。"
            />

            <MessageBox variant={pageMessageVariant}>
              {isPageLoading ? "処理中..." : pageMessage}
            </MessageBox>
          </div>

          <section className={styles.searchCard}>
            <div className={styles.searchGrid}>
              <div className={styles.monthControl}>
                <p className={styles.searchLabel}>対象月</p>

                <div className={styles.monthInputRow}>
                  <Button type="button" variant="secondary" onClick={handlePreviousMonth}>
                    前月
                  </Button>

                  <select
                    className={styles.monthSelect}
                    value={targetYear}
                    title="対象年"
                    onChange={(event) => handleChangeTargetYear(event.target.value)}
                  >
                    {targetYearOptions.map((year) => (
                      <option key={year} value={year}>
                        {year}年
                      </option>
                    ))}
                  </select>

                  <select
                    className={styles.monthSelect}
                    value={targetMonthValue}
                    title="対象月"
                    onChange={(event) => handleChangeTargetMonth(event.target.value)}
                  >
                    {Array.from({ length: 12 }, (_, index) => index + 1).map((month) => (
                      <option key={month} value={month}>
                        {month}月
                      </option>
                    ))}
                  </select>

                  <Button type="button" variant="secondary" onClick={handleNextMonth}>
                    翌月
                  </Button>
                </div>
              </div>

              <label className={styles.keywordControl}>
                <span className={styles.searchLabel}>ユーザー検索</span>

                <input
                  className={styles.keywordInput}
                  value={keyword}
                  placeholder="名前・メール・所属名で検索"
                  onChange={(event) => setKeyword(event.target.value)}
                />
              </label>

              <div className={styles.searchActionArea}>
                <Button type="button" variant="primary" onClick={handleSearch} disabled={isPageLoading}>
                  検索
                </Button>
              </div>
            </div>

            <div className={styles.statusFilterArea}>
              <p className={styles.searchLabel}>申請状態</p>

              <div className={styles.statusButtonList}>
                {STATUS_OPTIONS.map((option) => {
                  const isSelected = selectedStatuses.includes(option.status);

                  return (
                    <button
                      key={option.status}
                      type="button"
                      className={`${styles.statusFilterButton} ${
                        isSelected ? styles.statusFilterButtonSelected : ""
                      }`}
                      onClick={() => handleToggleStatus(option.status)}
                    >
                      {option.label}
                    </button>
                  );
                })}
              </div>

              <p className={styles.selectedStatusText}>選択中：{selectedStatusLabel}</p>
            </div>
          </section>

          <section className={styles.summaryCard}>
            <div>
              <p className={styles.summaryLabel}>対象月</p>
              <p className={styles.summaryValue}>{formatTargetMonth(targetYear, targetMonthValue)}</p>
            </div>

            <div>
              <p className={styles.summaryLabel}>検索ワード</p>
              <p className={styles.summaryValue}>{searchedKeyword || "指定なし"}</p>
            </div>

            <div>
              <p className={styles.summaryLabel}>表示件数</p>
              <p className={styles.summaryValue}>{rows.length}件</p>
            </div>
          </section>

          <section className={styles.tableCard}>
            <div className={styles.tableScroll}>
              <table className={styles.table}>
                <thead>
                  <tr>
                    <th>ユーザー</th>
                    <th>所属</th>
                    <th>対象月</th>
                    <th>状態</th>
                    <th>申請日時</th>
                    <th>承認日時</th>
                    <th>否認理由</th>
                    <th>操作</th>
                  </tr>
                </thead>

                <tbody>
                  {rows.map((row) => {
                    const request = row.monthlyAttendanceRequest;
                    const status = request.status;
                    const canApprove = request.canApprove && request.id !== null;
                    const canReject = request.canReject && request.id !== null;

                    return (
                      <tr key={`${row.targetUserId}-${row.targetYear}-${row.targetMonth}`}>
                        <td>
                          <div className={styles.userCell}>
                            <span className={styles.userName}>{row.userName}</span>
                            <span className={styles.userEmail}>{row.email}</span>
                          </div>
                        </td>

                        <td>{row.departmentName || "-"}</td>

                        <td>{formatTargetMonth(row.targetYear, row.targetMonth)}</td>

                        <td>
                          <span className={`${styles.statusBadge} ${getStatusClassName(status)}`}>
                            {getStatusLabel(status)}
                          </span>
                        </td>

                        <td>{formatDateTime(request.requestedAt)}</td>

                        <td>{formatDateTime(request.approvedAt)}</td>

                        <td className={styles.reasonCell}>{request.rejectedReason || "-"}</td>

                        <td>
                          <div className={styles.actionButtonList}>
                            <Button
                              type="button"
                              variant="secondary"
                              onClick={() => handleOpenAttendancePage(row)}
                            >
                              勤怠確認
                            </Button>

                            <Button
                              type="button"
                              variant="primary"
                              disabled={isPageLoading || !canApprove}
                              onClick={() => handleApprove(row)}
                            >
                              承認
                            </Button>

                            <Button
                              type="button"
                              variant="danger"
                              disabled={isPageLoading || !canReject}
                              onClick={() => handleReject(row)}
                            >
                              否認
                            </Button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}

                  {rows.length === 0 && (
                    <tr>
                      <td colSpan={8} className={styles.emptyCell}>
                        月次申請が見つかりません。
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {hasMore && (
              <div className={styles.moreButtonArea}>
                <Button type="button" variant="secondary" onClick={handleLoadMore} disabled={isPageLoading}>
                  さらに表示
                </Button>
              </div>
            )}
          </section>
        </section>
      </div>
    </PageContainer>
  );
}
