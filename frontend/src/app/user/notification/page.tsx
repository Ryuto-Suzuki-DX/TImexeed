"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import UserSideMenu from "@/components/sideMenu/UserSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import { readNotification, searchNotifications } from "@/api/user/notification";
import type { Notification } from "@/types/user/notification";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

const PAGE_LIMIT = 10;

function formatDateTime(value: string | null | undefined) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "-";
  }

  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const hour = String(date.getHours()).padStart(2, "0");
  const minute = String(date.getMinutes()).padStart(2, "0");

  return `${year}/${month}/${day} ${hour}:${minute}`;
}

export default function UserNotificationsPage() {
  const { user, isLoading, message } = useRequireRole("USER");

  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [pageMessage, setPageMessage] = useState("お知らせを確認できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");
  const [isPageLoading, setIsPageLoading] = useState(false);
  const [isMoreLoading, setIsMoreLoading] = useState(false);

  const unreadCount = useMemo(() => {
    return notifications.filter((notification) => !notification.isRead).length;
  }, [notifications]);

  const loadNotifications = useCallback(
    async (nextOffset: number, append: boolean) => {
      if (!user) {
        return;
      }

      if (append) {
        setIsMoreLoading(true);
      } else {
        setIsPageLoading(true);
        setPageMessage("お知らせを取得しています。");
        setPageMessageVariant("info");
      }

      const result = await searchNotifications({
        limit: PAGE_LIMIT,
        offset: nextOffset,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "お知らせ一覧の取得に失敗しました。");
        setPageMessageVariant("error");
        setIsPageLoading(false);
        setIsMoreLoading(false);
        return;
      }

      const data = result.data;

      setNotifications((currentNotifications) =>
        append ? [...currentNotifications, ...data.notifications] : data.notifications,
      );

      setHasMore(data.hasMore);
      setOffset(nextOffset + data.notifications.length);

      if (data.notifications.length === 0 && !append) {
        setPageMessage("現在、お知らせはありません。");
        setPageMessageVariant("info");
      } else {
        setPageMessage("お知らせを取得しました。");
        setPageMessageVariant("success");
      }

      setIsPageLoading(false);
      setIsMoreLoading(false);
    },
    [user],
  );

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void loadNotifications(0, false);
    }, 0);

    return () => {
      window.clearTimeout(timerId);
    };
  }, [isLoading, loadNotifications, user]);

  const handleReadNotification = async (notification: Notification) => {
    if (notification.isRead) {
      return;
    }

    const result = await readNotification({
      notificationId: notification.id,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "お知らせの既読更新に失敗しました。");
      setPageMessageVariant("error");
      return;
    }

    const data = result.data;

    setNotifications((currentNotifications) =>
      currentNotifications.map((currentNotification) =>
        currentNotification.id === notification.id ? data.notification : currentNotification,
      ),
    );

    setPageMessage("お知らせを既読にしました。");
    setPageMessageVariant("success");
  };

  const handleLoadMore = () => {
    void loadNotifications(offset, true);
  };

  if (isLoading || !user) {
    return (
      <PageContainer>
        <UserSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="お知らせ" description="ログイン情報を確認しています。" />
          <MessageBox variant="info">{message}</MessageBox>
        </section>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <UserSideMenu />

      <div className={styles.pageWrap}>
        <section className={styles.pageCard}>
          <div className={styles.header}>
            <PageTitle title="お知らせ" description="月次申請や承認結果などのお知らせを確認できます。" />

            <div className={styles.summaryBox}>
              <p className={styles.summaryLabel}>未読</p>
              <p className={styles.summaryValue}>{unreadCount}件</p>
            </div>
          </div>

          <div className={styles.messageArea}>
            <MessageBox variant={pageMessageVariant}>{isPageLoading ? "読み込み中..." : pageMessage}</MessageBox>
          </div>

          <div className={styles.notificationList}>
            {notifications.length === 0 && !isPageLoading ? (
              <div className={styles.emptyBox}>
                <p className={styles.emptyTitle}>お知らせはありません</p>
                <p className={styles.emptyText}>月次申請や承認結果があると、ここに表示されます。</p>
              </div>
            ) : (
              notifications.map((notification) => (
                <article
                  key={notification.id}
                  className={`${styles.notificationCard} ${
                    notification.isRead ? styles.readCard : styles.unreadCard
                  }`}
                >
                  <div className={styles.notificationHeader}>
                    <div className={styles.notificationTitleArea}>
                      {!notification.isRead && <span className={styles.newBadge}>NEW</span>}
                      <h2 className={styles.notificationTitle}>{notification.title}</h2>
                    </div>

                    <p className={styles.createdAt}>{formatDateTime(notification.createdAt)}</p>
                  </div>

                  <p className={styles.notificationMessage}>{notification.message}</p>

                  <div className={styles.notificationFooter}>
                    <p className={styles.readStatus}>
                      {notification.isRead ? `既読：${formatDateTime(notification.readAt)}` : "未読"}
                    </p>

                    {!notification.isRead && (
                      <Button type="button" variant="secondary" onClick={() => void handleReadNotification(notification)}>
                        既読にする
                      </Button>
                    )}
                  </div>
                </article>
              ))
            )}
          </div>

          {hasMore && (
            <div className={styles.moreArea}>
              <Button type="button" variant="secondary" onClick={handleLoadMore} disabled={isMoreLoading}>
                {isMoreLoading ? "読み込み中..." : "もっと見る"}
              </Button>
            </div>
          )}
        </section>
      </div>
    </PageContainer>
  );
}