"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import {
  createNotificationForAllUsers,
  deleteNotification,
  readNotification,
  searchNotifications,
} from "@/api/admin/notification";
import type { Notification } from "@/types/admin/notification";
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

export default function AdminNotificationsPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);

  const [keyword, setKeyword] = useState("");

  const [title, setTitle] = useState("");
  const [notificationMessage, setNotificationMessage] = useState("");

  const [pageMessage, setPageMessage] = useState("お知らせを確認・作成できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const [isPageLoading, setIsPageLoading] = useState(false);
  const [isMoreLoading, setIsMoreLoading] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [processingNotificationId, setProcessingNotificationId] = useState<number | null>(null);

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
        keyword: keyword.trim(),
        offset: nextOffset,
        limit: PAGE_LIMIT,
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

      setTotal(data.total);
      setHasMore(data.hasMore);
      setOffset(data.offset + data.notifications.length);

      if (data.notifications.length === 0 && !append) {
        setPageMessage("条件に一致するお知らせはありません。");
        setPageMessageVariant("info");
      } else {
        setPageMessage("お知らせを取得しました。");
        setPageMessageVariant("success");
      }

      setIsPageLoading(false);
      setIsMoreLoading(false);
    },
    [keyword, user],
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

  const handleSearch = () => {
    void loadNotifications(0, false);
  };

  const handleCreateNotification = async () => {
    const trimmedTitle = title.trim();
    const trimmedMessage = notificationMessage.trim();

    if (!trimmedTitle) {
      setPageMessage("タイトルを入力してください。");
      setPageMessageVariant("warning");
      return;
    }

    if (!trimmedMessage) {
      setPageMessage("本文を入力してください。");
      setPageMessageVariant("warning");
      return;
    }

    setIsCreating(true);
    setPageMessage("全員宛のお知らせを作成しています。");
    setPageMessageVariant("info");

    const result = await createNotificationForAllUsers({
      title: trimmedTitle,
      message: trimmedMessage,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "全員宛お知らせの作成に失敗しました。");
      setPageMessageVariant("error");
      setIsCreating(false);
      return;
    }

    setTitle("");
    setNotificationMessage("");

    setPageMessage(`全員宛のお知らせを作成しました。作成件数：${result.data.createdCount}件`);
    setPageMessageVariant("success");
    setIsCreating(false);

    void loadNotifications(0, false);
  };

  const handleReadNotification = async (notification: Notification) => {
    if (notification.isRead) {
      return;
    }

    setProcessingNotificationId(notification.id);

    const result = await readNotification({
      notificationId: notification.id,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "お知らせの既読更新に失敗しました。");
      setPageMessageVariant("error");
      setProcessingNotificationId(null);
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
    setProcessingNotificationId(null);
  };

  const handleDeleteNotification = async (notification: Notification) => {
    const confirmed = window.confirm("このお知らせを削除します。よろしいですか？");

    if (!confirmed) {
      return;
    }

    setProcessingNotificationId(notification.id);

    const result = await deleteNotification({
      notificationId: notification.id,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "お知らせの削除に失敗しました。");
      setPageMessageVariant("error");
      setProcessingNotificationId(null);
      return;
    }

    setNotifications((currentNotifications) =>
      currentNotifications.filter((currentNotification) => currentNotification.id !== notification.id),
    );

    setTotal((currentTotal) => Math.max(currentTotal - 1, 0));
    setPageMessage("お知らせを削除しました。");
    setPageMessageVariant("success");
    setProcessingNotificationId(null);
  };

  const handleLoadMore = () => {
    void loadNotifications(offset, true);
  };

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="お知らせ管理" description="ログイン情報を確認しています。" />
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
          <div className={styles.header}>
            <PageTitle
              title="お知らせ管理"
              description="全員宛のお知らせ作成と、管理者宛のお知らせ確認ができます。"
            />

            <div className={styles.summaryArea}>
              <div className={styles.summaryBox}>
                <p className={styles.summaryLabel}>検索結果</p>
                <p className={styles.summaryValue}>{total}件</p>
              </div>

              <div className={styles.summaryBox}>
                <p className={styles.summaryLabel}>表示中の未読</p>
                <p className={styles.summaryValue}>{unreadCount}件</p>
              </div>
            </div>
          </div>

          <div className={styles.messageArea}>
            <MessageBox variant={pageMessageVariant}>{isPageLoading ? "読み込み中..." : pageMessage}</MessageBox>
          </div>

          <section className={styles.searchCard}>
            <div className={styles.sectionHeader}>
              <div>
                <h2 className={styles.sectionTitle}>検索条件</h2>
                <p className={styles.sectionDescription}>タイトルや本文をキーワードで検索できます。</p>
              </div>
            </div>

            <div className={styles.searchGrid}>
              <label className={styles.formLabel}>
                <span className={styles.labelText}>キーワード</span>
                <input
                  type="text"
                  value={keyword}
                  onChange={(event) => setKeyword(event.target.value)}
                  className={styles.textInput}
                  placeholder="タイトル・本文"
                />
              </label>

              <div className={styles.searchActionArea}>
                <Button type="button" variant="primary" onClick={handleSearch} disabled={isPageLoading}>
                  検索
                </Button>
              </div>
            </div>
          </section>

          <section className={styles.createCard}>
            <div className={styles.sectionHeader}>
              <div>
                <h2 className={styles.sectionTitle}>全員宛お知らせ作成</h2>
                <p className={styles.sectionDescription}>ADMIN / USER 両方に同じお知らせを作成します。</p>
              </div>
            </div>

            <div className={styles.formGrid}>
              <label className={styles.formLabel}>
                <span className={styles.labelText}>タイトル</span>
                <input
                  type="text"
                  value={title}
                  onChange={(event) => setTitle(event.target.value)}
                  className={styles.textInput}
                  placeholder="例：月次申請の締切について"
                  disabled={isCreating}
                />
              </label>

              <label className={styles.formLabel}>
                <span className={styles.labelText}>本文</span>
                <textarea
                  value={notificationMessage}
                  onChange={(event) => setNotificationMessage(event.target.value)}
                  className={styles.textArea}
                  placeholder="お知らせ本文を入力してください。"
                  disabled={isCreating}
                />
              </label>
            </div>

            <div className={styles.formActionArea}>
              <Button type="button" variant="primary" onClick={handleCreateNotification} disabled={isCreating}>
                {isCreating ? "作成中..." : "全員宛に作成"}
              </Button>
            </div>
          </section>

          <section className={styles.listSection}>
            <div className={styles.sectionHeader}>
              <div>
                <h2 className={styles.sectionTitle}>管理者宛お知らせ一覧</h2>
                <p className={styles.sectionDescription}>管理者本人宛に作成されたお知らせを確認できます。</p>
              </div>
            </div>

            <div className={styles.notificationList}>
              {notifications.length === 0 && !isPageLoading ? (
                <div className={styles.emptyBox}>
                  <p className={styles.emptyTitle}>お知らせはありません</p>
                  <p className={styles.emptyText}>条件に一致するお知らせがあると、ここに表示されます。</p>
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

                      <div className={styles.actionArea}>
                        {!notification.isRead && (
                          <Button
                            type="button"
                            variant="secondary"
                            onClick={() => void handleReadNotification(notification)}
                            disabled={processingNotificationId === notification.id}
                          >
                            {processingNotificationId === notification.id ? "処理中..." : "既読にする"}
                          </Button>
                        )}

                        <Button
                          type="button"
                          variant="secondary"
                          onClick={() => void handleDeleteNotification(notification)}
                          disabled={processingNotificationId === notification.id}
                        >
                          {processingNotificationId === notification.id ? "処理中..." : "削除"}
                        </Button>
                      </div>
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
        </section>
      </div>
    </PageContainer>
  );
}
