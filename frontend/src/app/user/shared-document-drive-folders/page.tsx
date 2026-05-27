"use client";

import { useCallback, useEffect, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import UserSideMenu from "@/components/sideMenu/UserSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import {
  getSharedDocumentDriveFolderDetail,
  searchSharedDocumentDriveFolders,
} from "@/api/user/sharedDocumentDriveFolder";
import type {
  SharedDocumentDriveFolder,
  SharedDocumentDriveFolderSearchRow,
} from "@/types/user/sharedDocumentDriveFolder";
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

export default function UserSharedDocumentDriveFoldersPage() {
  const { user, isLoading, message } = useRequireRole("USER");

  const [folders, setFolders] = useState<SharedDocumentDriveFolderSearchRow[]>([]);
  const [selectedFolder, setSelectedFolder] =
    useState<SharedDocumentDriveFolder | null>(null);

  const [keyword, setKeyword] = useState("");
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);

  const [pageMessage, setPageMessage] = useState(
    "共有資料Driveフォルダを確認できます。",
  );
  const [pageMessageVariant, setPageMessageVariant] =
    useState<PageMessageVariant>("info");

  const [isPageLoading, setIsPageLoading] = useState(false);
  const [isMoreLoading, setIsMoreLoading] = useState(false);
  const [isDetailLoading, setIsDetailLoading] = useState(false);
  const [processingFolderId, setProcessingFolderId] = useState<number | null>(null);

  const selectedFolderId = selectedFolder?.id ?? null;

  const loadFolders = useCallback(
    async (nextOffset: number, append: boolean) => {
      if (!user) {
        return;
      }

      if (append) {
        setIsMoreLoading(true);
      } else {
        setIsPageLoading(true);
        setPageMessage("共有資料Driveフォルダを取得しています。");
        setPageMessageVariant("info");
      }

      try {
        const result = await searchSharedDocumentDriveFolders({
          keyword: keyword.trim(),
          offset: nextOffset,
          limit: PAGE_LIMIT,
        });

        if (result.error || !result.data) {
          if (!append) {
            setFolders([]);
            setSelectedFolder(null);
            setTotal(0);
            setOffset(0);
            setHasMore(false);
          }

          setPageMessage(
            result.message || "共有資料Driveフォルダ一覧の取得に失敗しました。",
          );
          setPageMessageVariant("error");
          return;
        }

        const data = result.data;
        const nextFolders = data.sharedDocumentDriveFolders ?? [];

        setFolders((currentFolders) =>
          append ? [...currentFolders, ...nextFolders] : nextFolders,
        );
        setTotal(data.total ?? 0);
        setHasMore(data.hasMore ?? false);
        setOffset((data.offset ?? nextOffset) + nextFolders.length);

        if (nextFolders.length === 0 && !append) {
          setSelectedFolder(null);
          setPageMessage("現在表示できる共有資料Driveフォルダはありません。");
          setPageMessageVariant("info");
        } else {
          setPageMessage("共有資料Driveフォルダを取得しました。");
          setPageMessageVariant("success");
        }
      } catch (error) {
        console.error(error);

        if (!append) {
          setFolders([]);
          setSelectedFolder(null);
          setTotal(0);
          setOffset(0);
          setHasMore(false);
        }

        setPageMessage(
          "共有資料Driveフォルダ一覧の取得中に予期しないエラーが発生しました。",
        );
        setPageMessageVariant("error");
      } finally {
        setIsPageLoading(false);
        setIsMoreLoading(false);
      }
    },
    [keyword, user],
  );

  const loadFolderDetail = useCallback(
    async (targetSharedDocumentDriveFolderId: number) => {
      setIsDetailLoading(true);
      setProcessingFolderId(targetSharedDocumentDriveFolderId);
      setPageMessage("共有資料Driveフォルダの詳細を取得しています。");
      setPageMessageVariant("info");

      try {
        const result = await getSharedDocumentDriveFolderDetail({
          targetSharedDocumentDriveFolderId,
        });

        if (result.error || !result.data) {
          setPageMessage(
            result.message || "共有資料Driveフォルダ詳細の取得に失敗しました。",
          );
          setPageMessageVariant("error");
          return;
        }

        setSelectedFolder(result.data.sharedDocumentDriveFolder);
        setPageMessage("共有資料Driveフォルダを選択しました。");
        setPageMessageVariant("success");
      } catch (error) {
        console.error(error);
        setPageMessage(
          "共有資料Driveフォルダ詳細の取得中に予期しないエラーが発生しました。",
        );
        setPageMessageVariant("error");
      } finally {
        setIsDetailLoading(false);
        setProcessingFolderId(null);
      }
    },
    [],
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
  }, [isLoading, loadFolders, user]);

  const handleSearch = () => {
    void loadFolders(0, false);
  };

  const handleLoadMore = () => {
    void loadFolders(offset, true);
  };

  const handleOpenDriveFolder = (folderUrl: string | null | undefined) => {
    if (!folderUrl) {
      setPageMessage("開くDriveフォルダURLがありません。");
      setPageMessageVariant("warning");
      return;
    }

    window.open(folderUrl, "_blank", "noopener,noreferrer");
  };

  if (isLoading || !user) {
    return (
      <PageContainer>
        <UserSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="共有資料" description="ログイン情報を確認しています。" />
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
            <PageTitle
              title="共有資料"
              description="会社から共有されている資料・FAQ・マニュアルを確認できます。"
            />

            <div className={styles.summaryArea}>
              <div className={styles.summaryBox}>
                <p className={styles.summaryLabel}>共有資料</p>
                <p className={styles.summaryValue}>{total}件</p>
              </div>

              <div className={styles.summaryBox}>
                <p className={styles.summaryLabel}>選択中</p>
                <p className={styles.summaryValue}>{selectedFolder ? "1件" : "0件"}</p>
              </div>
            </div>
          </div>

          <div className={styles.messageArea}>
            <MessageBox variant={pageMessageVariant}>
              {isPageLoading ? "読み込み中..." : pageMessage}
            </MessageBox>
          </div>

          <section className={styles.searchCard}>
            <div className={styles.sectionHeader}>
              <div>
                <h2 className={styles.sectionTitle}>検索条件</h2>
                <p className={styles.sectionDescription}>
                  フォルダ名・説明で共有資料を検索できます。
                </p>
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
                  placeholder="例：FAQ / 入社後書類 / 勤怠マニュアル"
                />
              </label>

              <div className={styles.searchActionArea}>
                <Button
                  type="button"
                  variant="primary"
                  onClick={handleSearch}
                  disabled={isPageLoading}
                >
                  検索
                </Button>
              </div>
            </div>
          </section>

          <section className={styles.contentGrid}>
            <div className={styles.folderListCard}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>共有資料一覧</h2>
                  <p className={styles.sectionDescription}>
                    全ユーザー向けに公開されている共有資料フォルダが表示されます。
                  </p>
                </div>
              </div>

              <div className={styles.folderList}>
                {folders.length === 0 && !isPageLoading ? (
                  <div className={styles.emptyBox}>
                    <p className={styles.emptyTitle}>共有資料はありません</p>
                    <p className={styles.emptyText}>
                      表示できる共有資料があると、ここに表示されます。
                    </p>
                  </div>
                ) : (
                  folders.map((folder) => (
                    <article
                      key={folder.id}
                      className={`${styles.folderCard} ${
                        selectedFolderId === folder.id ? styles.selectedFolderCard : ""
                      }`}
                    >
                      <div className={styles.folderCardHeader}>
                        <div className={styles.folderTitleArea}>
                          <h2 className={styles.folderTitle}>{folder.folderName}</h2>
                          <p className={styles.folderDescription}>
                            {folder.description || "説明なし"}
                          </p>
                        </div>

                        <span className={styles.scopeBadge}>共有</span>
                      </div>

                      <div className={styles.folderMetaGrid}>
                        <div>
                          <p className={styles.metaLabel}>最終同期日時</p>
                          <p className={styles.metaValue}>
                            {formatDateTime(folder.syncedAt)}
                          </p>
                        </div>

                        <div>
                          <p className={styles.metaLabel}>更新日時</p>
                          <p className={styles.metaValue}>
                            {formatDateTime(folder.updatedAt)}
                          </p>
                        </div>
                      </div>

                      <div className={styles.folderActionArea}>
                        <Button
                          type="button"
                          variant="secondary"
                          onClick={() => void loadFolderDetail(folder.id)}
                          disabled={processingFolderId === folder.id || isDetailLoading}
                        >
                          {processingFolderId === folder.id ? "取得中..." : "詳細"}
                        </Button>

                        <Button
                          type="button"
                          variant="primary"
                          onClick={() => handleOpenDriveFolder(folder.folderUrl)}
                        >
                          Driveを開く
                        </Button>
                      </div>
                    </article>
                  ))
                )}
              </div>

              {hasMore && (
                <div className={styles.moreArea}>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={handleLoadMore}
                    disabled={isMoreLoading}
                  >
                    {isMoreLoading ? "読み込み中..." : "もっと見る"}
                  </Button>
                </div>
              )}
            </div>

            <div className={styles.detailCard}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>詳細</h2>
                  <p className={styles.sectionDescription}>
                    選択した共有資料フォルダの情報を表示します。
                  </p>
                </div>

                {selectedFolder && (
                  <span className={styles.selectedBadge}>ID: {selectedFolder.id}</span>
                )}
              </div>

              {!selectedFolder ? (
                <div className={styles.emptyBox}>
                  <p className={styles.emptyTitle}>フォルダが選択されていません</p>
                  <p className={styles.emptyText}>
                    左側の一覧から詳細を確認するフォルダを選択してください。
                  </p>
                </div>
              ) : (
                <div className={styles.detailBody}>
                  <div className={styles.detailItem}>
                    <p className={styles.detailLabel}>フォルダ名</p>
                    <p className={styles.detailValue}>{selectedFolder.folderName}</p>
                  </div>

                  <div className={styles.detailItem}>
                    <p className={styles.detailLabel}>説明</p>
                    <p className={styles.detailValue}>
                      {selectedFolder.description || "説明なし"}
                    </p>
                  </div>

                  <div className={styles.detailItem}>
                    <p className={styles.detailLabel}>作成日時</p>
                    <p className={styles.detailValue}>
                      {formatDateTime(selectedFolder.createdAt)}
                    </p>
                  </div>

                  <div className={styles.detailItem}>
                    <p className={styles.detailLabel}>最終同期日時</p>
                    <p className={styles.detailValue}>
                      {formatDateTime(selectedFolder.syncedAt)}
                    </p>
                  </div>

                  <div className={styles.detailItem}>
                    <p className={styles.detailLabel}>更新日時</p>
                    <p className={styles.detailValue}>
                      {formatDateTime(selectedFolder.updatedAt)}
                    </p>
                  </div>

                  <div className={styles.formActionArea}>
                    <Button
                      type="button"
                      variant="primary"
                      onClick={() => handleOpenDriveFolder(selectedFolder.folderUrl)}
                    >
                      Driveを開く
                    </Button>
                  </div>

                  <div className={styles.noticeBox}>
                    <p className={styles.noticeTitle}>注意</p>
                    <p className={styles.noticeText}>
                      Google Drive側で権限がまだ反映されていない場合、リンクを開いてもアクセスできないことがあります。
                      その場合は管理者に確認してください。
                    </p>
                  </div>
                </div>
              )}
            </div>
          </section>
        </section>
      </div>
    </PageContainer>
  );
}
