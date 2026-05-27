"use client";

import { useCallback, useEffect, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import {
  createSharedDocumentDriveFolder,
  deleteSharedDocumentDriveFolder,
  getSharedDocumentDriveFolderDetail,
  searchSharedDocumentDriveFolders,
  syncSharedDocumentDriveFolder,
  updateSharedDocumentDriveFolder,
} from "@/api/admin/sharedDocumentDriveFolder";
import type {
  SharedDocumentDriveFolder,
  SharedDocumentDriveFolderSearchRow,
} from "@/types/admin/sharedDocumentDriveFolder";
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

export default function AdminSharedDocumentDriveFoldersPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [folders, setFolders] = useState<SharedDocumentDriveFolderSearchRow[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [keyword, setKeyword] = useState("");

  const [selectedFolder, setSelectedFolder] = useState<SharedDocumentDriveFolder | null>(null);

  const [folderName, setFolderName] = useState("");
  const [description, setDescription] = useState("");

  const [pageMessage, setPageMessage] = useState(
    "共有資料Driveフォルダを確認・作成できます。",
  );
  const [pageMessageVariant, setPageMessageVariant] =
    useState<PageMessageVariant>("info");

  const [isPageLoading, setIsPageLoading] = useState(false);
  const [isMoreLoading, setIsMoreLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isDetailLoading, setIsDetailLoading] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);
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
          setPageMessage("条件に一致する共有資料Driveフォルダはありません。");
          setPageMessageVariant("info");
        } else {
          setPageMessage("共有資料Driveフォルダを取得しました。");
          setPageMessageVariant("success");
        }
      } catch (error) {
        console.error(error);

        if (!append) {
          setFolders([]);
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

        const folder = result.data.sharedDocumentDriveFolder;

        setSelectedFolder(folder);
        setFolderName(folder.folderName);
        setDescription(folder.description ?? "");

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

  const handleResetForm = () => {
    setSelectedFolder(null);
    setFolderName("");
    setDescription("");
    setPageMessage("新規作成モードに切り替えました。");
    setPageMessageVariant("info");
  };

  const handleSaveFolder = async () => {
    const trimmedFolderName = folderName.trim();
    const trimmedDescription = description.trim();

    if (!trimmedFolderName) {
      setPageMessage("フォルダ名を入力してください。");
      setPageMessageVariant("warning");
      return;
    }

    setIsSaving(true);
    setPageMessage(
      selectedFolder
        ? "共有資料Driveフォルダを更新しています。"
        : "共有資料Driveフォルダを作成しています。",
    );
    setPageMessageVariant("info");

    try {
      if (selectedFolder) {
        const result = await updateSharedDocumentDriveFolder({
          targetSharedDocumentDriveFolderId: selectedFolder.id,
          folderName: trimmedFolderName,
          description: trimmedDescription || null,
        });

        if (result.error || !result.data) {
          setPageMessage(
            result.message || "共有資料Driveフォルダの更新に失敗しました。",
          );
          setPageMessageVariant("error");
          return;
        }

        const updatedFolder = result.data.sharedDocumentDriveFolder;

        setSelectedFolder(updatedFolder);
        setFolderName(updatedFolder.folderName);
        setDescription(updatedFolder.description ?? "");

        setPageMessage("共有資料Driveフォルダを更新しました。");
        setPageMessageVariant("success");

        void loadFolders(0, false);
        return;
      }

      const result = await createSharedDocumentDriveFolder({
        folderName: trimmedFolderName,
        description: trimmedDescription || null,
      });

      if (result.error || !result.data) {
        setPageMessage(
          result.message || "共有資料Driveフォルダの作成に失敗しました。",
        );
        setPageMessageVariant("error");
        return;
      }

      const createdFolder = result.data.sharedDocumentDriveFolder;

      setSelectedFolder(createdFolder);
      setFolderName(createdFolder.folderName);
      setDescription(createdFolder.description ?? "");

      setPageMessage(
        "共有資料Driveフォルダを作成しました。必要に応じてDrive権限同期を実行してください。",
      );
      setPageMessageVariant("success");

      void loadFolders(0, false);
    } catch (error) {
      console.error(error);
      setPageMessage(
        "共有資料Driveフォルダの保存中に予期しないエラーが発生しました。",
      );
      setPageMessageVariant("error");
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeleteFolder = async () => {
    if (!selectedFolder) {
      setPageMessage("削除する共有資料Driveフォルダを選択してください。");
      setPageMessageVariant("warning");
      return;
    }

    const confirmed = window.confirm(
      "この共有資料DriveフォルダをTimexeed上から削除します。Google Drive上のフォルダ自体は削除されません。よろしいですか？",
    );

    if (!confirmed) {
      return;
    }

    setProcessingFolderId(selectedFolder.id);
    setPageMessage("共有資料Driveフォルダを削除しています。");
    setPageMessageVariant("info");

    try {
      const result = await deleteSharedDocumentDriveFolder({
        targetSharedDocumentDriveFolderId: selectedFolder.id,
      });

      if (result.error || !result.data) {
        setPageMessage(
          result.message || "共有資料Driveフォルダの削除に失敗しました。",
        );
        setPageMessageVariant("error");
        return;
      }

      handleResetForm();

      setPageMessage("共有資料Driveフォルダを削除しました。");
      setPageMessageVariant("success");

      void loadFolders(0, false);
    } catch (error) {
      console.error(error);
      setPageMessage(
        "共有資料Driveフォルダの削除中に予期しないエラーが発生しました。",
      );
      setPageMessageVariant("error");
    } finally {
      setProcessingFolderId(null);
    }
  };

  const handleOpenDriveFolder = () => {
    if (!selectedFolder?.folderUrl) {
      setPageMessage("開くDriveフォルダURLがありません。");
      setPageMessageVariant("warning");
      return;
    }

    window.open(selectedFolder.folderUrl, "_blank", "noopener,noreferrer");
  };

  const handleSyncPermissions = async (targetSharedDocumentDriveFolderId: number) => {
    const isAllFolders = targetSharedDocumentDriveFolderId === 0;

    const confirmed = window.confirm(
      isAllFolders
        ? "有効な共有資料Driveフォルダ全件について、有効な管理者・一般ユーザーへGoogle Drive権限を同期します。よろしいですか？"
        : "選択中の共有資料Driveフォルダについて、有効な管理者・一般ユーザーへGoogle Drive権限を同期します。よろしいですか？",
    );

    if (!confirmed) {
      return;
    }

    setIsSyncing(true);
    setProcessingFolderId(isAllFolders ? null : targetSharedDocumentDriveFolderId);
    setPageMessage("Google Drive権限を同期しています。");
    setPageMessageVariant("info");

    try {
      const result = await syncSharedDocumentDriveFolder({
        targetSharedDocumentDriveFolderId,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "Google Drive権限同期に失敗しました。");
        setPageMessageVariant("error");
        return;
      }

      const syncedFolders = result.data.sharedDocumentDriveFolders ?? [];

      if (selectedFolder) {
        const syncedSelectedFolder = syncedFolders.find(
          (folder) => folder.id === selectedFolder.id,
        );

        if (syncedSelectedFolder) {
          setSelectedFolder(syncedSelectedFolder);
          setFolderName(syncedSelectedFolder.folderName);
          setDescription(syncedSelectedFolder.description ?? "");
        } else if (!isAllFolders && syncedFolders.length > 0) {
          const firstSyncedFolder = syncedFolders[0];
          setSelectedFolder(firstSyncedFolder);
          setFolderName(firstSyncedFolder.folderName);
          setDescription(firstSyncedFolder.description ?? "");
        }
      }

      setPageMessage(
        `Google Drive権限を同期しました。同期フォルダ数：${result.data.syncedFolderCount}件 / 管理者：${result.data.targetAdminCount}人 / 一般ユーザー：${result.data.targetUserCount}人`,
      );
      setPageMessageVariant("success");

      void loadFolders(0, false);
    } catch (error) {
      console.error(error);
      setPageMessage("Google Drive権限同期中に予期しないエラーが発生しました。");
      setPageMessageVariant("error");
    } finally {
      setIsSyncing(false);
      setProcessingFolderId(null);
    }
  };

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="共有資料管理" description="ログイン情報を確認しています。" />
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
              title="共有資料管理"
              description="external_storage_links の共有資料Drive親フォルダ配下に、全ユーザー向け資料フォルダを作成・管理します。"
            />

            <div className={styles.headerActionArea}>
              <Button
                type="button"
                variant="primary"
                onClick={() => void handleSyncPermissions(0)}
                disabled={isSyncing || total === 0}
              >
                {isSyncing ? "同期中..." : "全フォルダ権限同期"}
              </Button>
            </div>
          </div>

          <div className={styles.summaryArea}>
            <div className={styles.summaryBox}>
              <p className={styles.summaryLabel}>検索結果</p>
              <p className={styles.summaryValue}>{total}件</p>
            </div>

            <div className={styles.summaryBox}>
              <p className={styles.summaryLabel}>共有対象</p>
              <p className={styles.summaryValue}>全USER</p>
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
                  フォルダ名・説明で検索できます。
                </p>
              </div>

              <Button type="button" variant="secondary" onClick={handleResetForm}>
                新規作成に戻る
              </Button>
            </div>

            <div className={styles.searchGrid}>
              <label className={styles.formLabel}>
                <span className={styles.labelText}>キーワード</span>
                <input
                  type="text"
                  value={keyword}
                  onChange={(event) => setKeyword(event.target.value)}
                  className={styles.textInput}
                  placeholder="フォルダ名・説明"
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

          <section className={styles.editorGrid}>
            <div className={styles.folderListCard}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>共有資料フォルダ一覧</h2>
                  <p className={styles.sectionDescription}>
                    選択すると右側で表示名・説明を編集できます。
                  </p>
                </div>
              </div>

              <div className={styles.folderList}>
                {folders.length === 0 && !isPageLoading ? (
                  <div className={styles.emptyBox}>
                    <p className={styles.emptyTitle}>共有資料フォルダはありません</p>
                    <p className={styles.emptyText}>
                      条件に一致するフォルダがあると、ここに表示されます。
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

                        <span className={styles.scopeBadge}>全USER</span>
                      </div>

                      <div className={styles.folderMetaGrid}>
                        <div>
                          <p className={styles.metaLabel}>同期日時</p>
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
                          {processingFolderId === folder.id ? "取得中..." : "選択"}
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
                  <h2 className={styles.sectionTitle}>
                    {selectedFolder ? "共有資料フォルダ編集" : "共有資料フォルダ作成"}
                  </h2>
                  <p className={styles.sectionDescription}>
                    作成時は、共有資料Drive親フォルダ配下に指定名のDriveフォルダを自動作成します。
                  </p>
                </div>

                {selectedFolder && <span className={styles.selectedBadge}>ID: {selectedFolder.id}</span>}
              </div>

              <div className={styles.formGrid}>
                <label className={styles.formLabel}>
                  <span className={styles.labelText}>フォルダ名</span>
                  <input
                    type="text"
                    value={folderName}
                    onChange={(event) => setFolderName(event.target.value)}
                    className={styles.textInput}
                    placeholder="例：入社後書類 / FAQ / 勤怠マニュアル"
                    disabled={isSaving}
                  />
                </label>

                <label className={styles.formLabel}>
                  <span className={styles.labelText}>説明</span>
                  <textarea
                    value={description}
                    onChange={(event) => setDescription(event.target.value)}
                    className={styles.textArea}
                    placeholder="共有資料の説明を入力してください。"
                    disabled={isSaving}
                  />
                </label>

                {selectedFolder && (
                  <div className={styles.readOnlyInfoBox}>
                    <p className={styles.readOnlyInfoTitle}>Driveフォルダ</p>
                    <p className={styles.readOnlyInfoText}>
                      DriveフォルダID：{selectedFolder.driveFolderId}
                    </p>
                    <p className={styles.readOnlyInfoText}>
                      DriveフォルダURL：{selectedFolder.folderUrl}
                    </p>
                    <p className={styles.readOnlyInfoText}>
                      最終同期日時：{formatDateTime(selectedFolder.syncedAt)}
                    </p>
                  </div>
                )}
              </div>

              <div className={styles.formActionArea}>
                {selectedFolder && (
                  <Button type="button" variant="secondary" onClick={handleOpenDriveFolder}>
                    Driveを開く
                  </Button>
                )}

                {selectedFolder && (
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => void handleSyncPermissions(selectedFolder.id)}
                    disabled={isSyncing}
                  >
                    {isSyncing ? "同期中..." : "選択フォルダ権限同期"}
                  </Button>
                )}

                {selectedFolder && (
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => void handleDeleteFolder()}
                    disabled={processingFolderId === selectedFolder.id}
                  >
                    {processingFolderId === selectedFolder.id ? "処理中..." : "削除"}
                  </Button>
                )}

                <Button
                  type="button"
                  variant="primary"
                  onClick={() => void handleSaveFolder()}
                  disabled={isSaving}
                >
                  {isSaving ? "保存中..." : selectedFolder ? "更新" : "作成"}
                </Button>
              </div>

              <section className={styles.policySection}>
                <h2 className={styles.sectionTitle}>現在の共有仕様</h2>
                <p className={styles.policyText}>
                  この画面で作成した共有資料Driveフォルダは、全ユーザー向け資料として扱います。
                  個別ユーザーの選択・共有対象リスト管理は行いません。
                </p>
                <p className={styles.policyText}>
                  Drive権限同期では、有効な管理者と有効な一般ユーザーをバックエンド側で取得し、
                  対象フォルダへ権限を付与します。
                </p>
              </section>
            </div>
          </section>
        </section>
      </div>
    </PageContainer>
  );
}
