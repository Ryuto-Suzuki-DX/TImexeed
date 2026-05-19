"use client";

import { useCallback, useEffect, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import {
  createNotificationReminder,
  deleteNotificationReminder,
  searchNotificationReminders,
  toggleNotificationReminderEnabled,
  updateNotificationReminder,
} from "@/api/admin/notificationReminder";
import type { NotificationReminder } from "@/types/admin/notificationReminder";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

const PAGE_LIMIT = 10;

type ReminderForm = {
  title: string;
  message: string;
  dayOffsetFromMonthEnd: string;
  sendHour: string;
  sendMinute: string;
};

const initialForm: ReminderForm = {
  title: "",
  message: "",
  dayOffsetFromMonthEnd: "3",
  sendHour: "9",
  sendMinute: "0",
};

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

function formatSendTime(hour: number, minute: number) {
  return `${String(hour).padStart(2, "0")}:${String(minute).padStart(2, "0")}`;
}

function toNumber(value: string) {
  const numberValue = Number(value);

  if (!Number.isFinite(numberValue)) {
    return null;
  }

  return numberValue;
}

export default function AdminNotificationRemindersPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [reminders, setReminders] = useState<NotificationReminder[]>([]);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);

  const [keyword, setKeyword] = useState("");
  const [includeDisabled, setIncludeDisabled] = useState(true);
  const [includeDeleted, setIncludeDeleted] = useState(false);

  const [form, setForm] = useState<ReminderForm>(initialForm);
  const [editingReminderId, setEditingReminderId] = useState<number | null>(null);
  const [editingIsEnabled, setEditingIsEnabled] = useState(true);

  const [pageMessage, setPageMessage] = useState("自動リマインド設定を管理できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const [isPageLoading, setIsPageLoading] = useState(false);
  const [isMoreLoading, setIsMoreLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [processingReminderId, setProcessingReminderId] = useState<number | null>(null);

  const loadReminders = useCallback(
    async (nextOffset: number, append: boolean) => {
      if (!user) {
        return;
      }

      if (append) {
        setIsMoreLoading(true);
      } else {
        setIsPageLoading(true);
        setPageMessage("自動リマインド設定を取得しています。");
        setPageMessageVariant("info");
      }

      const result = await searchNotificationReminders({
        keyword: keyword.trim(),
        includeDisabled,
        includeDeleted,
        limit: PAGE_LIMIT,
        offset: nextOffset,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "自動リマインド設定の取得に失敗しました。");
        setPageMessageVariant("error");
        setIsPageLoading(false);
        setIsMoreLoading(false);
        return;
      }

      const data = result.data;

      setReminders((currentReminders) => (append ? [...currentReminders, ...data.reminders] : data.reminders));
      setHasMore(data.hasMore);
      setOffset(nextOffset + data.reminders.length);

      if (data.reminders.length === 0 && !append) {
        setPageMessage("条件に一致する自動リマインド設定はありません。");
        setPageMessageVariant("info");
      } else {
        setPageMessage("自動リマインド設定を取得しました。");
        setPageMessageVariant("success");
      }

      setIsPageLoading(false);
      setIsMoreLoading(false);
    },
    [includeDeleted, includeDisabled, keyword, user],
  );

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void loadReminders(0, false);
    }, 0);

    return () => {
      window.clearTimeout(timerId);
    };
  }, [isLoading, loadReminders, user]);

  const handleChangeForm = (key: keyof ReminderForm, value: string) => {
    setForm((currentForm) => ({
      ...currentForm,
      [key]: value,
    }));
  };

  const resetForm = () => {
    setForm(initialForm);
    setEditingReminderId(null);
    setEditingIsEnabled(true);
  };

  const validateForm = () => {
    const title = form.title.trim();
    const messageValue = form.message.trim();
    const dayOffsetFromMonthEnd = toNumber(form.dayOffsetFromMonthEnd);
    const sendHour = toNumber(form.sendHour);
    const sendMinute = toNumber(form.sendMinute);

    if (!title) {
      setPageMessage("タイトルを入力してください。");
      setPageMessageVariant("warning");
      return null;
    }

    if (!messageValue) {
      setPageMessage("本文を入力してください。");
      setPageMessageVariant("warning");
      return null;
    }

    if (dayOffsetFromMonthEnd === null || dayOffsetFromMonthEnd < 0) {
      setPageMessage("月末からの日数は0以上の数値で入力してください。");
      setPageMessageVariant("warning");
      return null;
    }

    if (sendHour === null || sendHour < 0 || sendHour > 23) {
      setPageMessage("送信時は0〜23で入力してください。");
      setPageMessageVariant("warning");
      return null;
    }

    if (sendMinute === null || sendMinute < 0 || sendMinute > 59) {
      setPageMessage("送信分は0〜59で入力してください。");
      setPageMessageVariant("warning");
      return null;
    }

    return {
      title,
      message: messageValue,
      dayOffsetFromMonthEnd,
      sendHour,
      sendMinute,
    };
  };

  const handleSaveReminder = async () => {
    const validatedForm = validateForm();

    if (!validatedForm) {
      return;
    }

    setIsSaving(true);

    if (editingReminderId === null) {
      setPageMessage("自動リマインド設定を作成しています。");
      setPageMessageVariant("info");

      const result = await createNotificationReminder(validatedForm);

      if (result.error || !result.data) {
        setPageMessage(result.message || "自動リマインド設定の作成に失敗しました。");
        setPageMessageVariant("error");
        setIsSaving(false);
        return;
      }

      resetForm();
      setPageMessage("自動リマインド設定を作成しました。");
      setPageMessageVariant("success");
      setIsSaving(false);

      void loadReminders(0, false);
      return;
    }

    setPageMessage("自動リマインド設定を更新しています。");
    setPageMessageVariant("info");

    const result = await updateNotificationReminder({
      reminderId: editingReminderId,
      ...validatedForm,
      isEnabled: editingIsEnabled,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "自動リマインド設定の更新に失敗しました。");
      setPageMessageVariant("error");
      setIsSaving(false);
      return;
    }

    const data = result.data;

    setReminders((currentReminders) =>
      currentReminders.map((currentReminder) =>
        currentReminder.id === data.reminder.id ? data.reminder : currentReminder,
      ),
    );

    resetForm();
    setPageMessage("自動リマインド設定を更新しました。");
    setPageMessageVariant("success");
    setIsSaving(false);
  };

  const handleEditReminder = (reminder: NotificationReminder) => {
    setEditingReminderId(reminder.id);
    setEditingIsEnabled(reminder.isEnabled);
    setForm({
      title: reminder.title,
      message: reminder.message,
      dayOffsetFromMonthEnd: String(reminder.dayOffsetFromMonthEnd),
      sendHour: String(reminder.sendHour),
      sendMinute: String(reminder.sendMinute),
    });

    setPageMessage("編集内容をフォームに反映しました。");
    setPageMessageVariant("info");
  };

  const handleToggleEnabled = async (reminder: NotificationReminder) => {
    setProcessingReminderId(reminder.id);

    const result = await toggleNotificationReminderEnabled({
      reminderId: reminder.id,
      isEnabled: !reminder.isEnabled,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "有効/無効の切替に失敗しました。");
      setPageMessageVariant("error");
      setProcessingReminderId(null);
      return;
    }

    const data = result.data;

    setReminders((currentReminders) =>
      currentReminders.map((currentReminder) =>
        currentReminder.id === data.reminder.id ? data.reminder : currentReminder,
      ),
    );

    if (editingReminderId === data.reminder.id) {
      setEditingIsEnabled(data.reminder.isEnabled);
    }

    setPageMessage(data.reminder.isEnabled ? "自動リマインド設定を有効にしました。" : "自動リマインド設定を無効にしました。");
    setPageMessageVariant("success");
    setProcessingReminderId(null);
  };

  const handleDeleteReminder = async (reminder: NotificationReminder) => {
    const confirmed = window.confirm("この自動リマインド設定を削除します。よろしいですか？");

    if (!confirmed) {
      return;
    }

    setProcessingReminderId(reminder.id);

    const result = await deleteNotificationReminder({
      reminderId: reminder.id,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "自動リマインド設定の削除に失敗しました。");
      setPageMessageVariant("error");
      setProcessingReminderId(null);
      return;
    }

    if (includeDeleted) {
      const data = result.data;

      setReminders((currentReminders) =>
        currentReminders.map((currentReminder) =>
          currentReminder.id === data.reminder.id ? data.reminder : currentReminder,
        ),
      );
    } else {
      setReminders((currentReminders) =>
        currentReminders.filter((currentReminder) => currentReminder.id !== reminder.id),
      );
    }

    if (editingReminderId === reminder.id) {
      resetForm();
    }

    setPageMessage("自動リマインド設定を削除しました。");
    setPageMessageVariant("success");
    setProcessingReminderId(null);
  };

  const handleSearch = () => {
    void loadReminders(0, false);
  };

  const handleLoadMore = () => {
    void loadReminders(offset, true);
  };

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="自動リマインド設定" description="ログイン情報を確認しています。" />
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
              title="自動リマインド設定"
              description="月末から指定した日数・時刻で、全員宛のお知らせを自動作成する設定を管理できます。"
            />

            <div className={styles.summaryBox}>
              <p className={styles.summaryLabel}>表示中</p>
              <p className={styles.summaryValue}>{reminders.length}件</p>
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

              <label className={styles.checkLabel}>
                <input
                  type="checkbox"
                  checked={includeDisabled}
                  onChange={(event) => setIncludeDisabled(event.target.checked)}
                />
                <span>無効も含める</span>
              </label>

              <label className={styles.checkLabel}>
                <input
                  type="checkbox"
                  checked={includeDeleted}
                  onChange={(event) => setIncludeDeleted(event.target.checked)}
                />
                <span>削除済みも含める</span>
              </label>

              <div className={styles.searchActionArea}>
                <Button type="button" variant="primary" onClick={handleSearch} disabled={isPageLoading}>
                  検索
                </Button>
              </div>
            </div>
          </section>

          <section className={styles.formCard}>
            <div className={styles.sectionHeader}>
              <div>
                <h2 className={styles.sectionTitle}>
                  {editingReminderId === null ? "自動リマインド新規作成" : "自動リマインド編集"}
                </h2>
                <p className={styles.sectionDescription}>
                  例：月末3日前の09:00に、月次申請の締切リマインドを全員に作成します。
                </p>
              </div>

              {editingReminderId !== null && (
                <Button type="button" variant="secondary" onClick={resetForm} disabled={isSaving}>
                  新規作成に戻す
                </Button>
              )}
            </div>

            <div className={styles.formGrid}>
              <label className={styles.formLabel}>
                <span className={styles.labelText}>タイトル</span>
                <input
                  type="text"
                  value={form.title}
                  onChange={(event) => handleChangeForm("title", event.target.value)}
                  className={styles.textInput}
                  placeholder="例：月次申請締切リマインド"
                  disabled={isSaving}
                />
              </label>

              <label className={styles.formLabel}>
                <span className={styles.labelText}>本文</span>
                <textarea
                  value={form.message}
                  onChange={(event) => handleChangeForm("message", event.target.value)}
                  className={styles.textArea}
                  placeholder="例：月次勤怠の申請期限が近づいています。"
                  disabled={isSaving}
                />
              </label>

              <div className={styles.timeGrid}>
                <label className={styles.formLabel}>
                  <span className={styles.labelText}>月末から何日前</span>
                  <input
                    type="number"
                    min="0"
                    value={form.dayOffsetFromMonthEnd}
                    onChange={(event) => handleChangeForm("dayOffsetFromMonthEnd", event.target.value)}
                    className={styles.textInput}
                    disabled={isSaving}
                  />
                </label>

                <label className={styles.formLabel}>
                  <span className={styles.labelText}>送信時</span>
                  <input
                    type="number"
                    min="0"
                    max="23"
                    value={form.sendHour}
                    onChange={(event) => handleChangeForm("sendHour", event.target.value)}
                    className={styles.textInput}
                    disabled={isSaving}
                  />
                </label>

                <label className={styles.formLabel}>
                  <span className={styles.labelText}>送信分</span>
                  <input
                    type="number"
                    min="0"
                    max="59"
                    value={form.sendMinute}
                    onChange={(event) => handleChangeForm("sendMinute", event.target.value)}
                    className={styles.textInput}
                    disabled={isSaving}
                  />
                </label>
              </div>

              {editingReminderId !== null && (
                <label className={styles.checkLabel}>
                  <input
                    type="checkbox"
                    checked={editingIsEnabled}
                    onChange={(event) => setEditingIsEnabled(event.target.checked)}
                    disabled={isSaving}
                  />
                  <span>有効にする</span>
                </label>
              )}
            </div>

            <div className={styles.formActionArea}>
              <Button type="button" variant="primary" onClick={handleSaveReminder} disabled={isSaving}>
                {isSaving ? "保存中..." : editingReminderId === null ? "作成" : "更新"}
              </Button>
            </div>
          </section>

          <section className={styles.listSection}>
            <div className={styles.sectionHeader}>
              <div>
                <h2 className={styles.sectionTitle}>自動リマインド一覧</h2>
                <p className={styles.sectionDescription}>作成済みの自動リマインド設定を確認・編集できます。</p>
              </div>
            </div>

            <div className={styles.reminderList}>
              {reminders.length === 0 && !isPageLoading ? (
                <div className={styles.emptyBox}>
                  <p className={styles.emptyTitle}>自動リマインド設定はありません</p>
                  <p className={styles.emptyText}>新規作成すると、ここに表示されます。</p>
                </div>
              ) : (
                reminders.map((reminder) => (
                  <article
                    key={reminder.id}
                    className={`${styles.reminderCard} ${
                      reminder.isDeleted ? styles.deletedCard : reminder.isEnabled ? styles.enabledCard : styles.disabledCard
                    }`}
                  >
                    <div className={styles.reminderHeader}>
                      <div className={styles.reminderTitleArea}>
                        <span
                          className={`${styles.statusBadge} ${
                            reminder.isDeleted
                              ? styles.deletedBadge
                              : reminder.isEnabled
                                ? styles.enabledBadge
                                : styles.disabledBadge
                          }`}
                        >
                          {reminder.isDeleted ? "削除済み" : reminder.isEnabled ? "有効" : "無効"}
                        </span>

                        <h2 className={styles.reminderTitle}>{reminder.title}</h2>
                      </div>

                      <p className={styles.updatedAt}>更新：{formatDateTime(reminder.updatedAt)}</p>
                    </div>

                    <div className={styles.reminderMeta}>
                      <span>月末{reminder.dayOffsetFromMonthEnd}日前</span>
                      <span>{formatSendTime(reminder.sendHour, reminder.sendMinute)}</span>
                    </div>

                    <p className={styles.reminderMessage}>{reminder.message}</p>

                    <div className={styles.reminderFooter}>
                      <p className={styles.createdAt}>作成：{formatDateTime(reminder.createdAt)}</p>

                      <div className={styles.actionArea}>
                        {!reminder.isDeleted && (
                          <>
                            <Button type="button" variant="secondary" onClick={() => handleEditReminder(reminder)}>
                              編集
                            </Button>

                            <Button
                              type="button"
                              variant="secondary"
                              onClick={() => void handleToggleEnabled(reminder)}
                              disabled={processingReminderId === reminder.id}
                            >
                              {reminder.isEnabled ? "無効にする" : "有効にする"}
                            </Button>

                            <Button
                              type="button"
                              variant="secondary"
                              onClick={() => void handleDeleteReminder(reminder)}
                              disabled={processingReminderId === reminder.id}
                            >
                              {processingReminderId === reminder.id ? "処理中..." : "削除"}
                            </Button>
                          </>
                        )}
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