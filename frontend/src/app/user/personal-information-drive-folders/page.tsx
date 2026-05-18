"use client";

import { useCallback, useEffect, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import UserSideMenu from "@/components/sideMenu/UserSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import { getMyPersonalInformationDriveFolder } from "@/api/user/personalInformationDriveFolder";
import type { MyPersonalInformationDriveFolder } from "@/types/user/personalInformationDriveFolder";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

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

export default function UserPersonalInformationDriveFolderPage() {
  const { user, isLoading, message } = useRequireRole("USER");

  const [folder, setFolder] = useState<MyPersonalInformationDriveFolder | null>(null);
  const [isPageLoading, setIsPageLoading] = useState(false);
  const [pageMessage, setPageMessage] = useState("個人情報Driveフォルダを確認できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const loadFolder = useCallback(async () => {
    setIsPageLoading(true);
    setPageMessage("個人情報Driveフォルダを取得しています。");
    setPageMessageVariant("info");

    try {
      const result = await getMyPersonalInformationDriveFolder({});

      if (result.error || !result.data) {
        setFolder(null);
        setPageMessage(
          result.message ||
            "個人情報Driveフォルダがまだ作成されていません。管理者へ確認してください。",
        );
        setPageMessageVariant("warning");
        return;
      }

      setFolder(result.data.personalInformationDriveFolder);
      setPageMessage("個人情報Driveフォルダを取得しました。");
      setPageMessageVariant("success");
    } catch (error) {
      setFolder(null);
      setPageMessage(
        error instanceof Error
          ? error.message
          : "個人情報Driveフォルダの取得中に予期しないエラーが発生しました。",
      );
      setPageMessageVariant("error");
    } finally {
      setIsPageLoading(false);
    }
  }, []);

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void loadFolder();
    }, 0);

    return () => {
      window.clearTimeout(timerId);
    };
  }, [isLoading, loadFolder, user]);

  const handleOpenFolder = () => {
    if (!folder) {
      setPageMessage("個人情報Driveフォルダが取得できていません。");
      setPageMessageVariant("warning");
      return;
    }

    window.open(folder.folderUrl, "_blank", "noopener,noreferrer");
    setPageMessage("個人情報Driveフォルダを開きました。");
    setPageMessageVariant("success");
  };

  if (isLoading || !user) {
    return (
      <PageContainer>
        <UserSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="個人情報" description="ログイン情報を確認しています。" />
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
          <div className={styles.headerArea}>
            <PageTitle
              title="個人情報"
              description="管理者が用意したあなた専用の個人情報Driveフォルダを開けます。"
            />
          </div>

          <div className={styles.messageArea}>
            <MessageBox variant={pageMessageVariant}>
              {isPageLoading ? "読み込み中..." : pageMessage}
            </MessageBox>

            <div className={styles.statusBox}>
              <p className={styles.statusLabel}>フォルダ状態</p>
              <p className={styles.statusValue}>{folder ? "利用可能" : "未作成"}</p>
            </div>

            <div className={styles.statusBox}>
              <p className={styles.statusLabel}>最終同期</p>
              <p className={styles.statusSmallText}>{formatDateTime(folder?.syncedAt ?? null)}</p>
            </div>
          </div>

          <div className={styles.folderCard}>
            <div className={styles.folderInfoGrid}>
              <div className={styles.infoItem}>
                <p className={styles.infoLabel}>氏名</p>
                <p className={styles.infoValue}>{folder?.userName ?? user.name ?? "-"}</p>
              </div>

              <div className={styles.infoItem}>
                <p className={styles.infoLabel}>メールアドレス</p>
                <p className={styles.infoValue}>{folder?.userEmail ?? user.email ?? "-"}</p>
              </div>

              <div className={styles.infoItem}>
                <p className={styles.infoLabel}>フォルダ名</p>
                <p className={styles.infoValue}>{folder?.folderName ?? "-"}</p>
              </div>

              <div className={styles.infoItem}>
                <p className={styles.infoLabel}>作成日時</p>
                <p className={styles.infoValue}>{formatDateTime(folder?.createdAt ?? null)}</p>
              </div>
            </div>

            <div className={styles.actionArea}>
              <Button type="button" onClick={handleOpenFolder} disabled={!folder || isPageLoading}>
                個人情報Driveフォルダを開く
              </Button>

              <Button type="button" variant="secondary" onClick={() => void loadFolder()} disabled={isPageLoading}>
                再読み込み
              </Button>
            </div>

            <p className={styles.noteText}>
              フォルダが未作成の場合、管理者側で「作成/権限同期」を実行すると利用できるようになります。
            </p>
          </div>
        </section>
      </div>
    </PageContainer>
  );
}
