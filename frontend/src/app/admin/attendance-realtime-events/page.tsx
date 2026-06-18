"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { removeAccessToken } from "@/api/auth";
import { searchAttendanceRealtimeEvents } from "@/api/admin/attendanceRealtimeEvent";
import Button from "@/components/atoms/Button";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import type { AttendanceRealtimeEventResponse } from "@/types/admin/attendanceRealtimeEvent";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

type EventType = "CLOCK_IN" | "CLOCK_OUT" | "OTHER";

type MatrixUser = {
  userId: number;
  userName: string;
  userEmail: string;
};

type EventsByDateAndUser = Record<string, Record<number, AttendanceRealtimeEventResponse[]>>;

const EVENT_TYPE_ORDER: Record<string, number> = {
  CLOCK_IN: 1,
  CLOCK_OUT: 2,
  OTHER: 3,
};

function getCurrentYearMonth() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");

  return `${year}-${month}`;
}

function parseYearMonth(value: string) {
  const [yearText, monthText] = value.split("-");
  const year = Number(yearText);
  const month = Number(monthText);

  if (!year || !month || month < 1 || month > 12) {
    return null;
  }

  return { year, month };
}

function buildMonthDates(targetYearMonth: string) {
  const parsed = parseYearMonth(targetYearMonth);

  if (!parsed) {
    return [];
  }

  const { year, month } = parsed;
  const lastDay = new Date(year, month, 0).getDate();

  return Array.from({ length: lastDay }, (_, index) => {
    const day = index + 1;
    return `${year}-${String(month).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
  });
}

function formatDateLabel(dateText: string) {
  const date = new Date(`${dateText}T00:00:00`);

  if (Number.isNaN(date.getTime())) {
    return dateText;
  }

  const dayOfWeek = new Intl.DateTimeFormat("ja-JP", {
    weekday: "short",
  }).format(date);

  return `${date.getMonth() + 1}/${date.getDate()}(${dayOfWeek})`;
}

function formatTime(value: string | null | undefined) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "-";
  }

  return new Intl.DateTimeFormat("ja-JP", {
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function formatDateTime(value: string | null | undefined) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "-";
  }

  return new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function formatEventType(eventType: string) {
  switch (eventType) {
    case "CLOCK_IN":
      return "出勤";
    case "CLOCK_OUT":
      return "退勤";
    case "OTHER":
      return "その他";
    default:
      return eventType || "-";
  }
}

function getEventBadgeClass(eventType: string) {
  switch (eventType) {
    case "CLOCK_IN":
      return styles.eventBadgeClockIn;
    case "CLOCK_OUT":
      return styles.eventBadgeClockOut;
    case "OTHER":
      return styles.eventBadgeOther;
    default:
      return styles.eventBadgeOther;
  }
}

function getDateTextFromEventDate(value: string | null | undefined) {
  if (!value) {
    return "";
  }

  return value.split("T")[0] || "";
}

function sortEvents(events: AttendanceRealtimeEventResponse[]) {
  return [...events].sort((a, b) => {
    const orderA = EVENT_TYPE_ORDER[a.eventType] ?? 99;
    const orderB = EVENT_TYPE_ORDER[b.eventType] ?? 99;

    if (orderA !== orderB) {
      return orderA - orderB;
    }

    return new Date(a.eventAt).getTime() - new Date(b.eventAt).getTime();
  });
}

function buildUsers(events: AttendanceRealtimeEventResponse[]) {
  const userMap = new Map<number, MatrixUser>();

  for (const event of events) {
    if (!userMap.has(event.userId)) {
      userMap.set(event.userId, {
        userId: event.userId,
        userName: event.userName,
        userEmail: event.userEmail,
      });
    }
  }

  return Array.from(userMap.values()).sort((a, b) => {
    return a.userName.localeCompare(b.userName, "ja");
  });
}

function buildEventsByDateAndUser(events: AttendanceRealtimeEventResponse[]) {
  const grouped: EventsByDateAndUser = {};

  for (const event of events) {
    const dateText = getDateTextFromEventDate(event.eventDate);

    if (!dateText) {
      continue;
    }

    if (!grouped[dateText]) {
      grouped[dateText] = {};
    }

    if (!grouped[dateText][event.userId]) {
      grouped[dateText][event.userId] = [];
    }

    grouped[dateText][event.userId].push(event);
  }

  for (const dateText of Object.keys(grouped)) {
    for (const userIdText of Object.keys(grouped[dateText])) {
      const userId = Number(userIdText);
      grouped[dateText][userId] = sortEvents(grouped[dateText][userId]);
    }
  }

  return grouped;
}

export default function AdminAttendanceRealtimeEventsPage() {
  const router = useRouter();
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [targetYearMonth, setTargetYearMonth] = useState(getCurrentYearMonth());
  const [keyword, setKeyword] = useState("");
  const [selectedEventTypes, setSelectedEventTypes] = useState<EventType[]>([]);
  const [events, setEvents] = useState<AttendanceRealtimeEventResponse[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [pageMessage, setPageMessage] = useState("対象月を指定して、出退勤リアルタイム記録を確認できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const monthDates = useMemo(() => buildMonthDates(targetYearMonth), [targetYearMonth]);
  const users = useMemo(() => buildUsers(events), [events]);
  const eventsByDateAndUser = useMemo(() => buildEventsByDateAndUser(events), [events]);

  const loadMonthlyEvents = useCallback(async () => {
    if (!user) {
      return;
    }

    const dates = buildMonthDates(targetYearMonth);

    if (dates.length === 0) {
      setPageMessage("対象月の指定が正しくありません。");
      setPageMessageVariant("error");
      return;
    }

    setIsSearching(true);
    setPageMessage("出退勤リアルタイム記録を取得しています。");
    setPageMessageVariant("info");

    const allEvents: AttendanceRealtimeEventResponse[] = [];
    const trimmedKeyword = keyword.trim();

    for (const targetDate of dates) {
      const result = await searchAttendanceRealtimeEvents({
        targetDate,
        keyword: trimmedKeyword,
        eventTypes: selectedEventTypes,
        limit: 500,
        offset: 0,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || `${targetDate} の出退勤リアルタイム記録の取得に失敗しました。`);
        setPageMessageVariant("error");
        setIsSearching(false);
        return;
      }

      allEvents.push(...result.data.events);
    }

    setEvents(allEvents);

    if (allEvents.length === 0) {
      setPageMessage("対象月の出退勤リアルタイム記録はありません。");
      setPageMessageVariant("warning");
    } else {
      setPageMessage(`${allEvents.length}件の出退勤リアルタイム記録を取得しました。`);
      setPageMessageVariant("success");
    }

    setIsSearching(false);
  }, [keyword, selectedEventTypes, targetYearMonth, user]);

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void loadMonthlyEvents();
    }, 0);

    return () => {
      window.clearTimeout(timerId);
    };
  }, [isLoading, loadMonthlyEvents, user]);

  const handleLogout = () => {
    removeAccessToken();
    router.push("/login");
  };

  const handleToggleEventType = (eventType: EventType) => {
    setSelectedEventTypes((current) => {
      if (current.includes(eventType)) {
        return current.filter((value) => value !== eventType);
      }

      return [...current, eventType];
    });
  };

  return (
    <main className={styles.page}>
      <AdminSideMenu />

      <section className={styles.card}>
        <div className={styles.header}>
          <div>
            <h1 className={styles.title}>出退勤リアルタイム一覧</h1>
            <p className={styles.description}>
              従業員が押した出勤・退勤・その他ボタンの履歴を、月単位で確認します。
            </p>
          </div>

          <div className={styles.headerActionArea}>
            <Button type="button" variant="secondary" onClick={() => void loadMonthlyEvents()} disabled={isLoading || isSearching || !user}>
              {isSearching ? "検索中..." : "再検索"}
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
              {isSearching ? "検索中..." : pageMessage}
            </div>

            <section className={styles.searchSection}>
              <div className={styles.searchGrid}>
                <label className={styles.formField}>
                  <span className={styles.formLabel}>対象月</span>
                  <input
                    className={styles.input}
                    type="month"
                    value={targetYearMonth}
                    onChange={(event) => setTargetYearMonth(event.target.value)}
                  />
                </label>

                <label className={styles.formField}>
                  <span className={styles.formLabel}>ユーザー検索</span>
                  <input
                    className={styles.input}
                    type="text"
                    value={keyword}
                    onChange={(event) => setKeyword(event.target.value)}
                    placeholder="氏名・メールアドレス"
                  />
                </label>

                <div className={styles.formField}>
                  <span className={styles.formLabel}>種別</span>
                  <div className={styles.checkboxRow}>
                    <label className={styles.checkboxLabel}>
                      <input
                        type="checkbox"
                        checked={selectedEventTypes.includes("CLOCK_IN")}
                        onChange={() => handleToggleEventType("CLOCK_IN")}
                      />
                      出勤
                    </label>

                    <label className={styles.checkboxLabel}>
                      <input
                        type="checkbox"
                        checked={selectedEventTypes.includes("CLOCK_OUT")}
                        onChange={() => handleToggleEventType("CLOCK_OUT")}
                      />
                      退勤
                    </label>

                    <label className={styles.checkboxLabel}>
                      <input
                        type="checkbox"
                        checked={selectedEventTypes.includes("OTHER")}
                        onChange={() => handleToggleEventType("OTHER")}
                      />
                      その他
                    </label>
                  </div>
                </div>
              </div>

              <div className={styles.searchActionArea}>
                <Button type="button" variant="primary" onClick={() => void loadMonthlyEvents()} disabled={isSearching}>
                  {isSearching ? "検索中..." : "この条件で検索"}
                </Button>
              </div>
            </section>

            <section className={styles.summarySection}>
              <div className={styles.summaryBox}>
                <p className={styles.summaryLabel}>対象月</p>
                <p className={styles.summaryValue}>{targetYearMonth.replace("-", "年")}月</p>
              </div>

              <div className={styles.summaryBox}>
                <p className={styles.summaryLabel}>表示ユーザー</p>
                <p className={styles.summaryValue}>{users.length}人</p>
              </div>

              <div className={styles.summaryBox}>
                <p className={styles.summaryLabel}>記録件数</p>
                <p className={styles.summaryValue}>{events.length}件</p>
              </div>
            </section>

            <section className={styles.matrixSection}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>月次一覧</h2>
                  <p className={styles.sectionDescription}>
                    横にユーザー、縦に日付を並べ、各セルに出勤・退勤・その他とコメントを表示します。
                  </p>
                </div>
              </div>

              {events.length === 0 ? (
                <div className={styles.emptyBox}>
                  <p className={styles.emptyTitle}>表示できる記録がありません</p>
                  <p className={styles.emptyText}>対象月・ユーザー検索・種別条件を変更して再検索してください。</p>
                </div>
              ) : (
                <div className={styles.tableScroll}>
                  <table className={styles.matrixTable}>
                    <thead>
                      <tr>
                        <th className={styles.dateHeader}>日付</th>
                        {users.map((matrixUser) => (
                          <th key={matrixUser.userId} className={styles.userHeader}>
                            <span className={styles.userName}>{matrixUser.userName}</span>
                            <span className={styles.userEmail}>{matrixUser.userEmail}</span>
                          </th>
                        ))}
                      </tr>
                    </thead>

                    <tbody>
                      {monthDates.map((dateText) => (
                        <tr key={dateText}>
                          <th className={styles.dateCell}>{formatDateLabel(dateText)}</th>

                          {users.map((matrixUser) => {
                            const cellEvents = eventsByDateAndUser[dateText]?.[matrixUser.userId] || [];

                            return (
                              <td key={`${dateText}-${matrixUser.userId}`} className={styles.eventCell}>
                                {cellEvents.length === 0 ? (
                                  <p className={styles.noEventText}>未押下</p>
                                ) : (
                                  <div className={styles.cellEventList}>
                                    {cellEvents.map((event) => (
                                      <div key={event.id} className={styles.cellEventItem}>
                                        <div className={styles.cellEventHeader}>
                                          <span className={`${styles.eventBadge} ${getEventBadgeClass(event.eventType)}`}>
                                            {formatEventType(event.eventType)}
                                          </span>
                                          <span className={styles.eventTime}>{formatTime(event.eventAt)}</span>
                                        </div>

                                        {event.note && (
                                          <p className={styles.eventNote}>{event.note}</p>
                                        )}

                                        <p className={styles.createdAtText}>記録：{formatDateTime(event.createdAt)}</p>
                                      </div>
                                    ))}
                                  </div>
                                )}
                              </td>
                            );
                          })}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </section>
          </>
        )}
      </section>
    </main>
  );
}
