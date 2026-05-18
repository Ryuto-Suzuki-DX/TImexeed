"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import {
  searchPersonalInformationDriveFolders,
  syncPersonalInformationDriveFolder,
  viewPersonalInformationDriveFolder,
} from "@/api/admin/personalInformationDriveFolder";
import type { PersonalInformationDriveFolderSearchRow } from "@/types/admin/personalInformationDriveFolder";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

const PAGE_LIMIT = 50;

function formatDateTime(value: string | null) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function getRoleLabel(role: string) {
  if (role === "ADMIN") {
    return "管理者";
  }

  if (role === "USER") {
    return "従業員";
  }

  return role || "-";
}

export default function AdminPersonalInformationDriveFoldersPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [keyword, setKeyword] = useState("");
  const [folders, setFolders] = useState<PersonalInformationDriveFolderSearchRow[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);

  const [isPageLoading, setIsPageLoading] = useState(false);
  const [processingUserId, setProcessingUserId] = useState<number | null>(null);
  const [pageMessage, setPageMessage] = useState("個人情報Driveフォルダを検索できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const displayedCount = useMemo(() => folders.length, [folders.length]);

  const loadFolders = useCallback(
    async (nextOffset: number, append: boolean) => {
      setIsPageLoading(true);
      setPageMessage("個人情報Driveフォルダ一覧を取得しています。");
      setPageMessageVariant("info");

      try {
        const result = await searchPersonalInformationDriveFolders({
          keyword,
          offset: nextOffset,
          limit: PAGE_LIMIT,
        });

        if (result.error || !result.data) {
          setPageMessage(result.message || "個人情報Driveフォルダ一覧の取得に失敗しました。");
          setPageMessageVariant("error");
          return;
        }

        const data = result.data;

        setFolders((currentFolders) =>
          append
            ? [...currentFolders, ...data.personalInformationDriveFolders]
            : data.personalInformationDriveFolders,
        );
        setTotal(data.total);
        setOffset(data.offset + data.personalInformationDriveFolders.length);
        setHasMore(data.hasMore);
        setPageMessage("個人情報Driveフォルダ一覧を取得しました。");
        setPageMessageVariant("success");
      } catch (error) {
        setPageMessage(
          error instanceof Error
            ? error.message
            : "個人情報Driveフォルダ一覧の取得中に予期しないエラーが発生しました。",
        );
        setPageMessageVariant("error");
      } finally {
        setIsPageLoading(false);
      }
    },
    [keyword],
  );

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void loadFolders(0, false);
    }, 0);

    return () => {
      window.clearTimeout(timerId);
    };
  }, [isLoading, user, loadFolders]);

  const handleSearch = () => {
    void loadFolders(0, false);
  };

  const handleLoadMore = () => {
    void loadFolders(offset, true);
  };

  const handleSync = async (row: PersonalInformationDriveFolderSearchRow) => {
    setProcessingUserId(row.userId);
    setPageMessage(`${row.userName} さんの個人情報Driveフォルダを最新状態に更新しています。`);
    setPageMessageVariant("info");

    try {
      const result = await syncPersonalInformationDriveFolder({
        targetUserId: row.userId,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "個人情報Driveフォルダの更新に失敗しました。");
        setPageMessageVariant("error");
        return;
      }

      setPageMessage(result.message || "個人情報Driveフォルダを最新状態に更新しました。");
      setPageMessageVariant("success");

      await loadFolders(0, false);
    } catch (error) {
      setPageMessage(
        error instanceof Error
          ? error.message
          : "個人情報Driveフォルダの更新中に予期しないエラーが発生しました。",
      );
      setPageMessageVariant("error");
    } finally {
      setProcessingUserId(null);
    }
  };

  const handleOpenFolder = async (row: PersonalInformationDriveFolderSearchRow) => {
    if (!row.folderRegistered) {
      setPageMessage("先に個人情報Driveフォルダを作成/権限同期してください。");
      setPageMessageVariant("warning");
      return;
    }

    setProcessingUserId(row.userId);
    setPageMessage(`${row.userName} さんの個人情報DriveフォルダURLを取得しています。`);
    setPageMessageVariant("info");

    try {
      const result = await viewPersonalInformationDriveFolder({
        targetUserId: row.userId,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "個人情報DriveフォルダURLの取得に失敗しました。");
        setPageMessageVariant("error");
        return;
      }

      window.open(result.data.personalInformationDriveFolder.folderUrl, "_blank", "noopener,noreferrer");
      setPageMessage("個人情報Driveフォルダを開きました。");
      setPageMessageVariant("success");
    } catch (error) {
      setPageMessage(
        error instanceof Error
          ? error.message
          : "個人情報DriveフォルダURLの取得中に予期しないエラーが発生しました。",
      );
      setPageMessageVariant("error");
    } finally {
      setProcessingUserId(null);
    }
  };

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="個人情報Driveフォルダ" description="ログイン情報を確認しています。" />
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
              title="個人情報Driveフォルダ"
              description="ユーザーごとの個人情報Driveフォルダを作成し、管理者全員と対象ユーザー本人の権限を最新状態に更新できます。"
            />
          </div>

          <div className={styles.searchCard}>
            <div className={styles.searchFields}>
              <label className={styles.fieldLabel}>
                フリーワード
                <input
                  className={styles.searchInput}
                  type="text"
                  value={keyword}
                  onChange={(event) => setKeyword(event.target.value)}
                  placeholder="氏名・メールアドレス・権限で検索"
                />
              </label>

              <div className={styles.searchActions}>
                <Button type="button" onClick={handleSearch} disabled={isPageLoading}>
                  検索
                </Button>
              </div>
            </div>
          </div>

          <div className={styles.messageArea}>
            <MessageBox variant={pageMessageVariant}>
              {isPageLoading ? "読み込み中..." : pageMessage}
            </MessageBox>

            <div className={styles.summaryBox}>
              <p className={styles.summaryLabel}>表示件数</p>
              <p className={styles.summaryValue}>
                {displayedCount} / {total}
              </p>
            </div>

            <div className={styles.summaryBox}>
              <p className={styles.summaryLabel}>操作</p>
              <p className={styles.summarySmallText}>作成/権限同期 → 開く</p>
            </div>
          </div>

          <div className={styles.tableWrap}>
            <table className={styles.folderTable}>
              <thead>
                <tr>
                  <th>ユーザーID</th>
                  <th>氏名</th>
                  <th>メールアドレス</th>
                  <th>権限</th>
                  <th>フォルダ状態</th>
                  <th>フォルダ名</th>
                  <th>最終同期</th>
                  <th>操作</th>
                </tr>
              </thead>

              <tbody>
                {folders.length === 0 ? (
                  <tr>
                    <td className={styles.emptyCell} colSpan={8}>
                      表示できる個人情報Driveフォルダ情報がありません。
                    </td>
                  </tr>
                ) : (
                  folders.map((row) => {
                    const isProcessing = processingUserId === row.userId;

                    return (
                      <tr key={row.userId}>
                        <td>{row.userId}</td>
                        <td className={styles.strongCell}>{row.userName}</td>
                        <td>{row.userEmail}</td>
                        <td>{getRoleLabel(row.userRole)}</td>
                        <td>
                          <span
                            className={
                              row.folderRegistered ? styles.statusCreated : styles.statusMissing
                            }
                          >
                            {row.folderRegistered ? "作成済み" : "未作成"}
                          </span>
                        </td>
                        <td>{row.folderName ?? "-"}</td>
                        <td>{formatDateTime(row.syncedAt)}</td>
                        <td>
                          <div className={styles.rowActions}>
                            <Button
                              type="button"
                              variant="secondary"
                              onClick={() => void handleSync(row)}
                              disabled={isPageLoading || processingUserId !== null}
                            >
                              {isProcessing ? "更新中" : "作成/権限同期"}
                            </Button>

                            <Button
                              type="button"
                              onClick={() => void handleOpenFolder(row)}
                              disabled={
                                isPageLoading ||
                                processingUserId !== null ||
                                !row.folderRegistered
                              }
                            >
                              開く
                            </Button>
                          </div>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>

          {hasMore && (
            <div className={styles.loadMoreArea}>
              <Button type="button" variant="secondary" onClick={handleLoadMore} disabled={isPageLoading}>
                さらに表示
              </Button>
            </div>
          )}
        </section>
      </div>
    </PageContainer>
  );
}
