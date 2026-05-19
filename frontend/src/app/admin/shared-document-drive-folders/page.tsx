"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
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
  updateSharedDocumentDriveFolderUsers,
} from "@/api/admin/sharedDocumentDriveFolder";
import { searchUsers } from "@/api/admin/user";
import type {
  SharedDocumentDriveFolder,
  SharedDocumentDriveFolderSearchRow,
  SharedDocumentDriveFolderUser,
} from "@/types/admin/sharedDocumentDriveFolder";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

type UserSearchResult = {
  id: number;
  name: string;
  email: string;
  role: string;
};

type SelectedSharedUser = {
  userId: number;
  userName: string;
  userEmail: string;
  userRole: string;
};

const PAGE_LIMIT = 10;
const USER_SEARCH_LIMIT = 20;

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

function toSelectedSharedUser(sharedUser: SharedDocumentDriveFolderUser): SelectedSharedUser {
  return {
    userId: sharedUser.userId,
    userName: sharedUser.userName,
    userEmail: sharedUser.userEmail,
    userRole: sharedUser.userRole,
  };
}

function toSelectedSharedUserFromSearchUser(user: UserSearchResult): SelectedSharedUser {
  return {
    userId: user.id,
    userName: user.name,
    userEmail: user.email,
    userRole: user.role,
  };
}

export default function AdminSharedDocumentDriveFoldersPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [folders, setFolders] = useState<SharedDocumentDriveFolderSearchRow[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [keyword, setKeyword] = useState("");

  const [selectedFolder, setSelectedFolder] = useState<SharedDocumentDriveFolder | null>(null);
  const [selectedUsers, setSelectedUsers] = useState<SelectedSharedUser[]>([]);

  const [folderName, setFolderName] = useState("");
  const [description, setDescription] = useState("");
  const [driveFolderUrlOrId, setDriveFolderUrlOrId] = useState("");

  const [userKeyword, setUserKeyword] = useState("");
  const [userSearchResults, setUserSearchResults] = useState<UserSearchResult[]>([]);

  const [pageMessage, setPageMessage] = useState("共有資料Driveフォルダを確認・作成できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const [isPageLoading, setIsPageLoading] = useState(false);
  const [isMoreLoading, setIsMoreLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isDetailLoading, setIsDetailLoading] = useState(false);
  const [isUserSearching, setIsUserSearching] = useState(false);
  const [isUserApplying, setIsUserApplying] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);
  const [processingFolderId, setProcessingFolderId] = useState<number | null>(null);

  const selectedUserIdSet = useMemo(() => {
    return new Set(selectedUsers.map((selectedUser) => selectedUser.userId));
  }, [selectedUsers]);

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

      const result = await searchSharedDocumentDriveFolders({
        keyword: keyword.trim(),
        offset: nextOffset,
        limit: PAGE_LIMIT,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "共有資料Driveフォルダ一覧の取得に失敗しました。");
        setPageMessageVariant("error");
        setIsPageLoading(false);
        setIsMoreLoading(false);
        return;
      }

      const data = result.data;

      setFolders((currentFolders) =>
        append
          ? [...currentFolders, ...data.sharedDocumentDriveFolders]
          : data.sharedDocumentDriveFolders,
      );
      setTotal(data.total);
      setHasMore(data.hasMore);
      setOffset(data.offset + data.sharedDocumentDriveFolders.length);

      if (data.sharedDocumentDriveFolders.length === 0 && !append) {
        setPageMessage("条件に一致する共有資料Driveフォルダはありません。");
        setPageMessageVariant("info");
      } else {
        setPageMessage("共有資料Driveフォルダを取得しました。");
        setPageMessageVariant("success");
      }

      setIsPageLoading(false);
      setIsMoreLoading(false);
    },
    [keyword, user],
  );

  const loadFolderDetail = useCallback(async (targetSharedDocumentDriveFolderId: number) => {
    setIsDetailLoading(true);
    setProcessingFolderId(targetSharedDocumentDriveFolderId);
    setPageMessage("共有資料Driveフォルダの詳細を取得しています。");
    setPageMessageVariant("info");

    const result = await getSharedDocumentDriveFolderDetail({
      targetSharedDocumentDriveFolderId,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "共有資料Driveフォルダ詳細の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsDetailLoading(false);
      setProcessingFolderId(null);
      return;
    }

    const data = result.data;

    setSelectedFolder(data.sharedDocumentDriveFolder);
    setSelectedUsers(data.sharedUsers.map(toSelectedSharedUser));

    setFolderName(data.sharedDocumentDriveFolder.folderName);
    setDescription(data.sharedDocumentDriveFolder.description ?? "");
    setDriveFolderUrlOrId(data.sharedDocumentDriveFolder.folderUrl);

    setPageMessage("共有資料Driveフォルダを選択しました。");
    setPageMessageVariant("success");
    setIsDetailLoading(false);
    setProcessingFolderId(null);
  }, []);

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
    setSelectedUsers([]);
    setFolderName("");
    setDescription("");
    setDriveFolderUrlOrId("");
    setUserKeyword("");
    setUserSearchResults([]);
    setPageMessage("新規作成モードに切り替えました。");
    setPageMessageVariant("info");
  };

  const handleSaveFolder = async () => {
    const trimmedFolderName = folderName.trim();
    const trimmedDescription = description.trim();
    const trimmedDriveFolderUrlOrId = driveFolderUrlOrId.trim();

    if (!trimmedDriveFolderUrlOrId) {
      setPageMessage("Google DriveフォルダURLまたはIDを入力してください。");
      setPageMessageVariant("warning");
      return;
    }

    setIsSaving(true);
    setPageMessage(selectedFolder ? "共有資料Driveフォルダを更新しています。" : "共有資料Driveフォルダを作成しています。");
    setPageMessageVariant("info");

    if (selectedFolder) {
      const result = await updateSharedDocumentDriveFolder({
        targetSharedDocumentDriveFolderId: selectedFolder.id,
        folderName: trimmedFolderName,
        description: trimmedDescription || null,
        driveFolderUrlOrId: trimmedDriveFolderUrlOrId,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "共有資料Driveフォルダの更新に失敗しました。");
        setPageMessageVariant("error");
        setIsSaving(false);
        return;
      }

      const updatedFolder = result.data.sharedDocumentDriveFolder;

      setSelectedFolder(updatedFolder);
      setFolderName(updatedFolder.folderName);
      setDescription(updatedFolder.description ?? "");
      setDriveFolderUrlOrId(updatedFolder.folderUrl);

      setPageMessage("共有資料Driveフォルダを更新しました。");
      setPageMessageVariant("success");
      setIsSaving(false);

      void loadFolders(0, false);
      return;
    }

    const result = await createSharedDocumentDriveFolder({
      folderName: trimmedFolderName,
      description: trimmedDescription || null,
      driveFolderUrlOrId: trimmedDriveFolderUrlOrId,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "共有資料Driveフォルダの作成に失敗しました。");
      setPageMessageVariant("error");
      setIsSaving(false);
      return;
    }

    const createdFolder = result.data.sharedDocumentDriveFolder;

    setSelectedFolder(createdFolder);
    setFolderName(createdFolder.folderName);
    setDescription(createdFolder.description ?? "");
    setDriveFolderUrlOrId(createdFolder.folderUrl);
    setSelectedUsers([]);

    setPageMessage("共有資料Driveフォルダを作成しました。続けて共有対象ユーザーを設定できます。");
    setPageMessageVariant("success");
    setIsSaving(false);

    void loadFolders(0, false);
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

    const result = await deleteSharedDocumentDriveFolder({
      targetSharedDocumentDriveFolderId: selectedFolder.id,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "共有資料Driveフォルダの削除に失敗しました。");
      setPageMessageVariant("error");
      setProcessingFolderId(null);
      return;
    }

    handleResetForm();

    setPageMessage("共有資料Driveフォルダを削除しました。");
    setPageMessageVariant("success");
    setProcessingFolderId(null);

    void loadFolders(0, false);
  };

  const handleOpenDriveFolder = () => {
    if (!selectedFolder?.folderUrl) {
      setPageMessage("開くDriveフォルダURLがありません。");
      setPageMessageVariant("warning");
      return;
    }

    window.open(selectedFolder.folderUrl, "_blank", "noopener,noreferrer");
  };

  const handleSearchUsers = async () => {
    setIsUserSearching(true);
    setPageMessage("ユーザーを検索しています。");
    setPageMessageVariant("info");

    const result = await searchUsers({
      keyword: userKeyword.trim(),
      includeDeleted: false,
      offset: 0,
      limit: USER_SEARCH_LIMIT,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "ユーザー検索に失敗しました。");
      setPageMessageVariant("error");
      setIsUserSearching(false);
      return;
    }

    const users = (result.data.users ?? []) as UserSearchResult[];
    const targetUsers = users.filter((searchedUser) => searchedUser.role === "USER");

    setUserSearchResults(targetUsers);

    if (targetUsers.length === 0) {
      setPageMessage("条件に一致する従業員ユーザーはありません。");
      setPageMessageVariant("info");
    } else {
      setPageMessage("ユーザーを検索しました。");
      setPageMessageVariant("success");
    }

    setIsUserSearching(false);
  };

  const handleAddSharedUser = (searchedUser: UserSearchResult) => {
    if (searchedUser.role !== "USER") {
      setPageMessage("共有対象に追加できるのはUSERのみです。");
      setPageMessageVariant("warning");
      return;
    }

    if (selectedUserIdSet.has(searchedUser.id)) {
      setPageMessage("すでに共有対象に追加されています。");
      setPageMessageVariant("warning");
      return;
    }

    setSelectedUsers((currentUsers) => [...currentUsers, toSelectedSharedUserFromSearchUser(searchedUser)]);
    setPageMessage("共有対象ユーザーを追加しました。まだDrive権限には反映されていません。");
    setPageMessageVariant("success");
  };

  const handleRemoveSharedUser = (targetUserId: number) => {
    setSelectedUsers((currentUsers) =>
      currentUsers.filter((currentUser) => currentUser.userId !== targetUserId),
    );
    setPageMessage("共有対象ユーザーを一覧から削除しました。まだDrive権限には反映されていません。");
    setPageMessageVariant("info");
  };

  const handleApplySharedUsers = async () => {
    if (!selectedFolder) {
      setPageMessage("先に共有資料Driveフォルダを選択してください。");
      setPageMessageVariant("warning");
      return;
    }

    setIsUserApplying(true);
    setPageMessage("共有対象ユーザーを適用しています。");
    setPageMessageVariant("info");

    const result = await updateSharedDocumentDriveFolderUsers({
      targetSharedDocumentDriveFolderId: selectedFolder.id,
      targetUserIds: selectedUsers.map((selectedUser) => selectedUser.userId),
      shareAllUsers: false,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "共有対象ユーザーの適用に失敗しました。");
      setPageMessageVariant("error");
      setIsUserApplying(false);
      return;
    }

    setSelectedUsers(result.data.sharedUsers.map(toSelectedSharedUser));
    setPageMessage("共有対象ユーザーを適用しました。Drive権限へ反映するには権限同期を実行してください。");
    setPageMessageVariant("success");
    setIsUserApplying(false);

    void loadFolders(0, false);
  };

  const handleAddAllUsers = async () => {
    if (!selectedFolder) {
      setPageMessage("先に共有資料Driveフォルダを選択してください。");
      setPageMessageVariant("warning");
      return;
    }

    const confirmed = window.confirm("有効なUSER全員を共有対象に追加します。よろしいですか？");

    if (!confirmed) {
      return;
    }

    setIsUserApplying(true);
    setPageMessage("全USERを共有対象に追加しています。");
    setPageMessageVariant("info");

    const result = await updateSharedDocumentDriveFolderUsers({
      targetSharedDocumentDriveFolderId: selectedFolder.id,
      targetUserIds: [],
      shareAllUsers: true,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "全員追加に失敗しました。");
      setPageMessageVariant("error");
      setIsUserApplying(false);
      return;
    }

    setSelectedUsers(result.data.sharedUsers.map(toSelectedSharedUser));
    setPageMessage("全USERを共有対象に追加しました。Drive権限へ反映するには権限同期を実行してください。");
    setPageMessageVariant("success");
    setIsUserApplying(false);

    void loadFolders(0, false);
  };

  const handleClearAllUsers = async () => {
    if (!selectedFolder) {
      setPageMessage("先に共有資料Driveフォルダを選択してください。");
      setPageMessageVariant("warning");
      return;
    }

    const confirmed = window.confirm("共有対象ユーザーを全員削除します。よろしいですか？");

    if (!confirmed) {
      return;
    }

    setIsUserApplying(true);
    setPageMessage("共有対象ユーザーを全員削除しています。");
    setPageMessageVariant("info");

    const result = await updateSharedDocumentDriveFolderUsers({
      targetSharedDocumentDriveFolderId: selectedFolder.id,
      targetUserIds: [],
      shareAllUsers: false,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "共有対象ユーザーの全削除に失敗しました。");
      setPageMessageVariant("error");
      setIsUserApplying(false);
      return;
    }

    setSelectedUsers([]);
    setPageMessage("共有対象ユーザーを全員削除しました。Drive権限へ反映するには権限同期を実行してください。");
    setPageMessageVariant("success");
    setIsUserApplying(false);

    void loadFolders(0, false);
  };

  const handleSyncPermissions = async () => {
    if (!selectedFolder) {
      setPageMessage("先に共有資料Driveフォルダを選択してください。");
      setPageMessageVariant("warning");
      return;
    }

    const confirmed = window.confirm(
      "Google Driveの権限を現在の共有対象ユーザーに同期します。共有対象から外れたユーザーのDrive権限も削除されます。よろしいですか？",
    );

    if (!confirmed) {
      return;
    }

    setIsSyncing(true);
    setPageMessage("Google Drive権限を同期しています。");
    setPageMessageVariant("info");

    const result = await syncSharedDocumentDriveFolder({
      targetSharedDocumentDriveFolderId: selectedFolder.id,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "Google Drive権限同期に失敗しました。");
      setPageMessageVariant("error");
      setIsSyncing(false);
      return;
    }

    setSelectedFolder(result.data.sharedDocumentDriveFolder);
    setSelectedUsers(result.data.sharedUsers.map(toSelectedSharedUser));
    setFolderName(result.data.sharedDocumentDriveFolder.folderName);
    setDescription(result.data.sharedDocumentDriveFolder.description ?? "");
    setDriveFolderUrlOrId(result.data.sharedDocumentDriveFolder.folderUrl);

    setPageMessage("Google Drive権限を同期しました。");
    setPageMessageVariant("success");
    setIsSyncing(false);

    void loadFolders(0, false);
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
              description="Google Driveフォルダを登録し、共有対象ユーザーへの権限を同期できます。"
            />

            <div className={styles.summaryArea}>
              <div className={styles.summaryBox}>
                <p className={styles.summaryLabel}>検索結果</p>
                <p className={styles.summaryValue}>{total}件</p>
              </div>

              <div className={styles.summaryBox}>
                <p className={styles.summaryLabel}>選択中の共有対象</p>
                <p className={styles.summaryValue}>{selectedUsers.length}人</p>
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
                  フォルダ名・説明・DriveフォルダIDで検索できます。
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
                  placeholder="フォルダ名・説明・DriveフォルダID"
                />
              </label>

              <div className={styles.searchActionArea}>
                <Button type="button" variant="primary" onClick={handleSearch} disabled={isPageLoading}>
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
                    選択すると右側で編集・共有対象ユーザー管理ができます。
                  </p>
                </div>
              </div>

              <div className={styles.folderList}>
                {folders.length === 0 && !isPageLoading ? (
                  <div className={styles.emptyBox}>
                    <p className={styles.emptyTitle}>共有資料フォルダはありません</p>
                    <p className={styles.emptyText}>条件に一致するフォルダがあると、ここに表示されます。</p>
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
                          <p className={styles.folderDescription}>{folder.description || "説明なし"}</p>
                        </div>

                        <span className={styles.countBadge}>{folder.sharedUserCount}人</span>
                      </div>

                      <div className={styles.folderMetaGrid}>
                        <div>
                          <p className={styles.metaLabel}>同期日時</p>
                          <p className={styles.metaValue}>{formatDateTime(folder.syncedAt)}</p>
                        </div>

                        <div>
                          <p className={styles.metaLabel}>更新日時</p>
                          <p className={styles.metaValue}>{formatDateTime(folder.updatedAt)}</p>
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
                  <Button type="button" variant="secondary" onClick={handleLoadMore} disabled={isMoreLoading}>
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
                    Google Driveで作成済みのフォルダURLまたはIDを登録してください。
                  </p>
                </div>

                {selectedFolder && (
                  <span className={styles.selectedBadge}>ID: {selectedFolder.id}</span>
                )}
              </div>

              <div className={styles.formGrid}>
                <label className={styles.formLabel}>
                  <span className={styles.labelText}>フォルダ名</span>
                  <input
                    type="text"
                    value={folderName}
                    onChange={(event) => setFolderName(event.target.value)}
                    className={styles.textInput}
                    placeholder="例：全体教育資料"
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

                <label className={styles.formLabel}>
                  <span className={styles.labelText}>Google DriveフォルダURL または ID</span>
                  <input
                    type="text"
                    value={driveFolderUrlOrId}
                    onChange={(event) => setDriveFolderUrlOrId(event.target.value)}
                    className={styles.textInput}
                    placeholder="https://drive.google.com/drive/folders/..."
                    disabled={isSaving}
                  />
                </label>
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
                    onClick={() => void handleDeleteFolder()}
                    disabled={processingFolderId === selectedFolder.id}
                  >
                    {processingFolderId === selectedFolder.id ? "処理中..." : "削除"}
                  </Button>
                )}

                <Button type="button" variant="primary" onClick={() => void handleSaveFolder()} disabled={isSaving}>
                  {isSaving ? "保存中..." : selectedFolder ? "更新" : "作成"}
                </Button>
              </div>

              {selectedFolder && (
                <section className={styles.shareSection}>
                  <div className={styles.sectionHeader}>
                    <div>
                      <h2 className={styles.sectionTitle}>共有対象ユーザー</h2>
                      <p className={styles.sectionDescription}>
                        追加・削除後、「共有対象を適用」してから「Drive権限同期」を実行してください。
                      </p>
                    </div>
                  </div>

                  <div className={styles.userSearchBox}>
                    <div className={styles.userSearchGrid}>
                      <label className={styles.formLabel}>
                        <span className={styles.labelText}>ユーザー検索</span>
                        <input
                          type="text"
                          value={userKeyword}
                          onChange={(event) => setUserKeyword(event.target.value)}
                          className={styles.textInput}
                          placeholder="名前・メールアドレス"
                        />
                      </label>

                      <div className={styles.searchActionArea}>
                        <Button
                          type="button"
                          variant="secondary"
                          onClick={() => void handleSearchUsers()}
                          disabled={isUserSearching}
                        >
                          {isUserSearching ? "検索中..." : "ユーザー検索"}
                        </Button>
                      </div>
                    </div>

                    {userSearchResults.length > 0 && (
                      <div className={styles.userResultList}>
                        {userSearchResults.map((searchedUser) => (
                          <div key={searchedUser.id} className={styles.userResultRow}>
                            <div>
                              <p className={styles.userName}>{searchedUser.name}</p>
                              <p className={styles.userEmail}>{searchedUser.email}</p>
                            </div>

                            <Button
                              type="button"
                              variant="secondary"
                              onClick={() => handleAddSharedUser(searchedUser)}
                              disabled={selectedUserIdSet.has(searchedUser.id)}
                            >
                              {selectedUserIdSet.has(searchedUser.id) ? "追加済み" : "追加"}
                            </Button>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>

                  <div className={styles.bulkActionArea}>
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => void handleAddAllUsers()}
                      disabled={isUserApplying}
                    >
                      {isUserApplying ? "処理中..." : "全員追加"}
                    </Button>

                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => void handleClearAllUsers()}
                      disabled={isUserApplying}
                    >
                      {isUserApplying ? "処理中..." : "全員削除"}
                    </Button>

                    <Button
                      type="button"
                      variant="primary"
                      onClick={() => void handleApplySharedUsers()}
                      disabled={isUserApplying}
                    >
                      {isUserApplying ? "適用中..." : "共有対象を適用"}
                    </Button>

                    <Button
                      type="button"
                      variant="primary"
                      onClick={() => void handleSyncPermissions()}
                      disabled={isSyncing || isUserApplying}
                    >
                      {isSyncing ? "同期中..." : "Drive権限同期"}
                    </Button>
                  </div>

                  <div className={styles.sharedUserList}>
                    {selectedUsers.length === 0 ? (
                      <div className={styles.emptyBox}>
                        <p className={styles.emptyTitle}>共有対象ユーザーはいません</p>
                        <p className={styles.emptyText}>
                          ユーザー検索から追加するか、全員追加を実行してください。
                        </p>
                      </div>
                    ) : (
                      selectedUsers.map((selectedUser) => (
                        <div key={selectedUser.userId} className={styles.sharedUserRow}>
                          <div>
                            <p className={styles.userName}>{selectedUser.userName}</p>
                            <p className={styles.userEmail}>{selectedUser.userEmail}</p>
                          </div>

                          <Button
                            type="button"
                            variant="secondary"
                            onClick={() => handleRemoveSharedUser(selectedUser.userId)}
                            disabled={isUserApplying || isSyncing}
                          >
                            削除
                          </Button>
                        </div>
                      ))
                    )}
                  </div>

                  <div className={styles.syncInfoBox}>
                    <p className={styles.syncInfoTitle}>同期状態</p>
                    <p className={styles.syncInfoText}>
                      最終同期日時：{formatDateTime(selectedFolder.syncedAt)}
                    </p>
                    <p className={styles.syncInfoText}>
                      Drive権限同期では、管理者全員と共有対象ユーザーに writer 権限を付与します。
                    </p>
                  </div>
                </section>
              )}
            </div>
          </section>
        </section>
      </div>
    </PageContainer>
  );
}