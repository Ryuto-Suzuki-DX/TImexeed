"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { removeAccessToken } from "@/api/auth";
import { countUnreadNotifications } from "@/api/user/notification";
import { searchMonthlyAttendanceRequest } from "@/api/user/monthlyAttendanceRequest";
import Button from "@/components/atoms/Button";
import { useRequireRole } from "@/hooks/useRequireRole";
import UserSideMenu from "@/components/sideMenu/UserSideMenu";
import type {
  MonthlyAttendanceRequestStatus,
  SearchMonthlyAttendanceRequestResponse,
} from "@/types/user/monthlyAttendanceRequest";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

type TargetMonth = {
  year: number;
  month: number;
  label: string;
};

type MonthlyStatusSummary = {
  year: number;
  month: number;
  label: string;
  status: MonthlyAttendanceRequestStatus | "NONE";
  adminMessage: string | null;
};

function getTargetMonths(): TargetMonth[] {
  const today = new Date();

  const currentYear = today.getFullYear();
  const currentMonth = today.getMonth() + 1;

  const previousDate = new Date(currentYear, currentMonth - 2, 1);
  const currentDate = new Date(currentYear, currentMonth - 1, 1);

  const previousYear = previousDate.getFullYear();
  const previousMonth = previousDate.getMonth() + 1;

  return [
    {
      year: previousYear,
      month: previousMonth,
      label: "前月",
    },
    {
      year: currentDate.getFullYear(),
      month: currentDate.getMonth() + 1,
      label: "当月",
    },
  ];
}

function toMonthlyStatusSummary(
  targetMonth: TargetMonth,
  response: SearchMonthlyAttendanceRequestResponse,
): MonthlyStatusSummary {
  return {
    year: targetMonth.year,
    month: targetMonth.month,
    label: targetMonth.label,
    status: response.monthlyAttendanceRequest?.status ?? "NONE",
    adminMessage: response.monthlyAttendanceRequest?.adminMessage ?? null,
  };
}

function formatYearMonth(year: number, month: number) {
  return `${year}年${month}月`;
}

function formatMonthlyStatus(status: MonthlyStatusSummary["status"]) {
  switch (status) {
    case "DRAFT":
      return "未申請";
    case "PENDING":
      return "申請中";
    case "APPROVED":
      return "承認済み";
    case "REJECTED":
      return "否認";
    case "NONE":
      return "未申請";
    default:
      return "未申請";
  }
}

function getMonthlyStatusVariant(status: MonthlyStatusSummary["status"]) {
  switch (status) {
    case "APPROVED":
      return "success";
    case "PENDING":
      return "info";
    case "REJECTED":
      return "error";
    case "DRAFT":
    case "NONE":
    default:
      return "warning";
  }
}

export default function UserMyPage() {
  const router = useRouter();

  const { user, isLoading, message } = useRequireRole("USER");

  const [unreadNotificationCount, setUnreadNotificationCount] = useState(0);
  const [monthlyStatuses, setMonthlyStatuses] = useState<MonthlyStatusSummary[]>([]);

  const [isDashboardLoading, setIsDashboardLoading] = useState(false);
  const [pageMessage, setPageMessage] = useState("従業員ホーム情報を確認できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const targetMonths = useMemo(() => getTargetMonths(), []);

  const loadDashboard = useCallback(async () => {
    if (!user) {
      return;
    }

    setIsDashboardLoading(true);
    setPageMessage("ホーム情報を取得しています。");
    setPageMessageVariant("info");

    const [notificationResult, previousMonthResult, currentMonthResult] = await Promise.all([
      countUnreadNotifications({}),
      searchMonthlyAttendanceRequest({
        targetYear: targetMonths[0].year,
        targetMonth: targetMonths[0].month,
      }),
      searchMonthlyAttendanceRequest({
        targetYear: targetMonths[1].year,
        targetMonth: targetMonths[1].month,
      }),
    ]);

    if (notificationResult.error || !notificationResult.data) {
      setPageMessage(notificationResult.message || "未読お知らせ件数の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsDashboardLoading(false);
      return;
    }

    if (previousMonthResult.error || !previousMonthResult.data) {
      setPageMessage(previousMonthResult.message || "前月の月次勤怠申請状態の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsDashboardLoading(false);
      return;
    }

    if (currentMonthResult.error || !currentMonthResult.data) {
      setPageMessage(currentMonthResult.message || "当月の月次勤怠申請状態の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsDashboardLoading(false);
      return;
    }

    const nextMonthlyStatuses = [
      toMonthlyStatusSummary(targetMonths[0], previousMonthResult.data),
      toMonthlyStatusSummary(targetMonths[1], currentMonthResult.data),
    ];

    setUnreadNotificationCount(notificationResult.data.unreadCount);
    setMonthlyStatuses(nextMonthlyStatuses);

    const hasUnreadNotifications = notificationResult.data.unreadCount > 0;
    const hasAttentionMonthlyStatus = nextMonthlyStatuses.some((monthlyStatus) => {
      return monthlyStatus.status === "DRAFT" || monthlyStatus.status === "NONE" || monthlyStatus.status === "REJECTED";
    });

    if (hasUnreadNotifications || hasAttentionMonthlyStatus) {
      setPageMessage("確認が必要な項目があります。");
      setPageMessageVariant("warning");
    } else {
      setPageMessage("現在、確認が必要な項目はありません。");
      setPageMessageVariant("success");
    }

    setIsDashboardLoading(false);
  }, [targetMonths, user]);

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void loadDashboard();
    }, 0);

    return () => {
      window.clearTimeout(timerId);
    };
  }, [isLoading, loadDashboard, user]);

  const handleLogout = () => {
    removeAccessToken();
    router.push("/login");
  };

  const handleReload = () => {
    void loadDashboard();
  };

  return (
    <main className={styles.page}>
      <UserSideMenu />

      <section className={styles.card}>
        <div className={styles.header}>
          <div>
            <h1 className={styles.title}>従業員マイページ</h1>
            <p className={styles.description}>ログイン中の従業員情報と、確認が必要な項目を表示しています。</p>
          </div>

          <div className={styles.headerActionArea}>
            <Button type="button" variant="secondary" onClick={handleReload} disabled={isLoading || isDashboardLoading || !user}>
              {isDashboardLoading ? "更新中..." : "再読み込み"}
            </Button>

            <Button type="button" variant="secondary" onClick={handleLogout}>
              ログアウト
            </Button>
          </div>
        </div>

        {isLoading && <p className={styles.loadingText}>{message}</p>}

        {!isLoading && user && (
          <>
            <div className={`${styles.pageMessage} ${styles[`pageMessage_${pageMessageVariant}`]}`}>
              {isDashboardLoading ? "読み込み中..." : pageMessage}
            </div>

            <div className={styles.infoList}>
              <div className={styles.infoBox}>
                <p className={styles.infoLabel}>名前</p>
                <p className={styles.infoValue}>{user.name}</p>
              </div>

              <div className={styles.infoBox}>
                <p className={styles.infoLabel}>ロール</p>
                <p className={styles.infoValue}>{user.role}</p>
              </div>

              <div className={styles.infoBox}>
                <p className={styles.infoLabel}>メールアドレス</p>
                <p className={styles.infoValue}>{user.email}</p>
              </div>
            </div>

            <div className={styles.dashboardGrid}>
              <section className={styles.dashboardCard}>
                <div className={styles.dashboardCardHeader}>
                  <div>
                    <h2 className={styles.dashboardTitle}>お知らせ</h2>
                    <p className={styles.dashboardDescription}>あなた宛の未読お知らせを確認します。</p>
                  </div>

                  {unreadNotificationCount > 0 ? (
                    <span className={styles.warningBadge}>要確認</span>
                  ) : (
                    <span className={styles.successBadge}>OK</span>
                  )}
                </div>

                <p className={styles.dashboardValue}>{unreadNotificationCount}件</p>
                <p className={styles.dashboardText}>
                  {unreadNotificationCount > 0
                    ? "未読のお知らせがあります。内容を確認してください。"
                    : "未読のお知らせはありません。"}
                </p>

                <div className={styles.cardActionArea}>
                  <Button type="button" variant="secondary" onClick={() => router.push("/user/notifications")}>
                    お知らせを見る
                  </Button>
                </div>
              </section>

              <section className={styles.dashboardCard}>
                <div className={styles.dashboardCardHeader}>
                  <div>
                    <h2 className={styles.dashboardTitle}>勤怠申請</h2>
                    <p className={styles.dashboardDescription}>前月と当月の月次勤怠申請状態を確認します。</p>
                  </div>

                  {monthlyStatuses.some((monthlyStatus) => monthlyStatus.status === "DRAFT" || monthlyStatus.status === "NONE" || monthlyStatus.status === "REJECTED") ? (
                    <span className={styles.warningBadge}>要確認</span>
                  ) : (
                    <span className={styles.successBadge}>OK</span>
                  )}
                </div>

                <div className={styles.monthlyStatusList}>
                  {monthlyStatuses.map((monthlyStatus) => {
                    const variant = getMonthlyStatusVariant(monthlyStatus.status);

                    return (
                      <div key={`${monthlyStatus.year}-${monthlyStatus.month}`} className={styles.monthlyStatusItem}>
                        <div>
                          <p className={styles.monthlyStatusLabel}>{monthlyStatus.label}</p>
                          <p className={styles.monthlyStatusMonth}>{formatYearMonth(monthlyStatus.year, monthlyStatus.month)}</p>
                        </div>

                        <span className={`${styles.monthlyStatusBadge} ${styles[`monthlyStatusBadge_${variant}`]}`}>
                          {formatMonthlyStatus(monthlyStatus.status)}
                        </span>
                      </div>
                    );
                  })}
                </div>

                <p className={styles.dashboardText}>
                  未申請または否認の月がある場合は、勤怠画面から内容を確認してください。
                </p>
              </section>
            </div>

            <section className={styles.statusSection}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>月次勤怠申請状況</h2>
                  <p className={styles.sectionDescription}>前月・当月の申請状態をまとめて表示しています。</p>
                </div>
              </div>

              <div className={styles.statusList}>
                {monthlyStatuses.map((monthlyStatus) => {
                  const variant = getMonthlyStatusVariant(monthlyStatus.status);

                  return (
                    <article key={`${monthlyStatus.year}-${monthlyStatus.month}-detail`} className={styles.statusItem}>
                      <div className={styles.statusItemHeader}>
                        <div>
                          <h3 className={styles.statusTitle}>
                            {monthlyStatus.label}：{formatYearMonth(monthlyStatus.year, monthlyStatus.month)}
                          </h3>
                          <p className={styles.statusDescription}>
                            {monthlyStatus.adminMessage || "管理者からのメッセージはありません。"}
                          </p>
                        </div>

                        <span className={`${styles.monthlyStatusBadge} ${styles[`monthlyStatusBadge_${variant}`]}`}>
                          {formatMonthlyStatus(monthlyStatus.status)}
                        </span>
                      </div>
                    </article>
                  );
                })}

                {monthlyStatuses.length === 0 && !isDashboardLoading && (
                  <div className={styles.emptyBox}>
                    <p className={styles.emptyTitle}>申請状況を取得できませんでした</p>
                    <p className={styles.emptyText}>再読み込みを行うか、時間をおいて確認してください。</p>
                  </div>
                )}
              </div>
            </section>
          </>
        )}
      </section>
    </main>
  );
}