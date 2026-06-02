"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { removeAccessToken } from "@/api/auth";
import { countUnreadNotifications } from "@/api/admin/notification";
import { searchPaidLeaveRequiredUseWarnings } from "@/api/admin/paidLeaveUsage";
import Button from "@/components/atoms/Button";
import { useRequireRole } from "@/hooks/useRequireRole";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import type { PaidLeaveRequiredUseWarningResponse } from "@/types/admin/paidLeaveUsage";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

function formatDate(value: string | null | undefined) {
  if (!value) {
    return "-";
  }

  const dateValue = value.split("T")[0];

  if (!dateValue) {
    return "-";
  }

  return dateValue.replaceAll("-", "/");
}

function formatNumber(value: number | null | undefined) {
  if (value === null || value === undefined) {
    return "-";
  }

  return value.toFixed(1).replace(".0", "");
}

export default function AdminMyPage() {
  const router = useRouter();

  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [unreadNotificationCount, setUnreadNotificationCount] = useState(0);
  const [paidLeaveWarnings, setPaidLeaveWarnings] = useState<PaidLeaveRequiredUseWarningResponse[]>([]);

  const [isDashboardLoading, setIsDashboardLoading] = useState(false);
  const [pageMessage, setPageMessage] = useState("管理者ホーム情報を確認できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const loadDashboard = useCallback(async () => {
    if (!user) {
      return;
    }

    setIsDashboardLoading(true);
    setPageMessage("ホーム情報を取得しています。");
    setPageMessageVariant("info");

    const [notificationResult, paidLeaveResult] = await Promise.all([
      countUnreadNotifications({}),
      searchPaidLeaveRequiredUseWarnings({
        deadlineWithinDays: 90,
      }),
    ]);

    if (notificationResult.error || !notificationResult.data) {
      setPageMessage(notificationResult.message || "未読お知らせ件数の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsDashboardLoading(false);
      return;
    }

    if (paidLeaveResult.error || !paidLeaveResult.data) {
      setPageMessage(paidLeaveResult.message || "有給取得警告の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsDashboardLoading(false);
      return;
    }

    setUnreadNotificationCount(notificationResult.data.unreadCount);
    setPaidLeaveWarnings(paidLeaveResult.data.warnings);

    if (notificationResult.data.unreadCount > 0 || paidLeaveResult.data.warnings.length > 0) {
      setPageMessage("確認が必要な項目があります。");
      setPageMessageVariant("warning");
    } else {
      setPageMessage("現在、確認が必要な項目はありません。");
      setPageMessageVariant("success");
    }

    setIsDashboardLoading(false);
  }, [user]);

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
      <AdminSideMenu />

      <section className={styles.card}>
        <div className={styles.header}>
          <div>
            <h1 className={styles.title}>管理者マイページ</h1>
            <p className={styles.description}>ログイン中の管理者情報と、確認が必要な項目を表示しています。</p>
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
                    <p className={styles.dashboardDescription}>管理者宛の未読お知らせを確認します。</p>
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
                    ? "未読のお知らせがあります。お知らせ管理画面で内容を確認してください。"
                    : "未読のお知らせはありません。"}
                </p>

                <div className={styles.cardActionArea}>
                  <Button type="button" variant="secondary" onClick={() => router.push("/admin/notifications")}>
                    お知らせを見る
                  </Button>
                </div>
              </section>

              <section className={styles.dashboardCard}>
                <div className={styles.dashboardCardHeader}>
                  <div>
                    <h2 className={styles.dashboardTitle}>有給取得義務</h2>
                    <p className={styles.dashboardDescription}>期限90日以内で、年5日取得が未達のユーザーを確認します。</p>
                  </div>

                  {paidLeaveWarnings.length > 0 ? (
                    <span className={styles.warningBadge}>要対応</span>
                  ) : (
                    <span className={styles.successBadge}>OK</span>
                  )}
                </div>

                <p className={styles.dashboardValue}>{paidLeaveWarnings.length}人</p>
                <p className={styles.dashboardText}>
                  {paidLeaveWarnings.length > 0
                    ? "有給取得を促したいユーザーがいます。期限と不足日数を確認してください。"
                    : "期限が近い有給取得義務の未達ユーザーはいません。"}
                </p>

                <div className={styles.cardActionArea}>
                  <Button type="button" variant="secondary" onClick={() => router.push("/admin/paid-leave-check")}>
                    有給確認へ
                  </Button>
                </div>
              </section>
            </div>

            <section className={styles.warningSection}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>有給取得を促したいユーザー</h2>
                  <p className={styles.sectionDescription}>期限まで90日以内で、年5日取得義務を満たしていないユーザーです。</p>
                </div>
              </div>

              {paidLeaveWarnings.length === 0 ? (
                <div className={styles.emptyBox}>
                  <p className={styles.emptyTitle}>現在、対象ユーザーはいません</p>
                  <p className={styles.emptyText}>有給取得義務の期限が近い未達ユーザーがいる場合、ここに表示されます。</p>
                </div>
              ) : (
                <div className={styles.warningList}>
                  {paidLeaveWarnings.map((warning) => (
                    <article key={warning.userId} className={styles.warningItem}>
                      <div className={styles.warningItemMain}>
                        <div>
                          <h3 className={styles.warningUserName}>{warning.userName}</h3>
                          <p className={styles.warningUserMeta}>
                            {warning.departmentName || "所属未設定"} / {warning.userEmail}
                          </p>
                        </div>

                        <span className={styles.deadlineBadge}>期限まで {warning.deadlineRemainingDays}日</span>
                      </div>

                      <div className={styles.warningDetailGrid}>
                        <div className={styles.warningDetailBox}>
                          <p className={styles.warningDetailLabel}>対象期間</p>
                          <p className={styles.warningDetailValue}>
                            {formatDate(warning.requiredUseStartDate)} 〜 {formatDate(warning.requiredUseDeadline)}
                          </p>
                        </div>

                        <div className={styles.warningDetailBox}>
                          <p className={styles.warningDetailLabel}>取得済み</p>
                          <p className={styles.warningDetailValue}>{formatNumber(warning.usedDaysInPeriod)}日</p>
                        </div>

                        <div className={styles.warningDetailBoxStrong}>
                          <p className={styles.warningDetailLabel}>残り必要</p>
                          <p className={styles.warningDetailValue}>{formatNumber(warning.requiredUseRemainingDays)}日</p>
                        </div>
                      </div>
                    </article>
                  ))}
                </div>
              )}
            </section>
          </>
        )}
      </section>
    </main>
  );
}