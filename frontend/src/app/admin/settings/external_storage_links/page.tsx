"use client";

import { useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import { searchExternalStorageLinks, updateExternalStorageLink } from "@/api/admin/externalStorageLink";
import type { ExternalStorageLinkResponse } from "@/types/admin/externalStorageLink";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

type ExternalStorageLinkForm = {
  externalStorageLinkId: number | null;
  linkType: string;
  linkName: string;
  url: string;
  description: string;
  memo: string;
};

const initialForm: ExternalStorageLinkForm = {
  externalStorageLinkId: null,
  linkType: "",
  linkName: "",
  url: "",
  description: "",
  memo: "",
};

function toDateInputValue(value: string | null | undefined) {
  if (!value) {
    return "";
  }

  return value.split("T")[0] ?? "";
}

function formatDate(value: string | null | undefined) {
  const dateValue = toDateInputValue(value);

  if (!dateValue) {
    return "-";
  }

  return dateValue.replaceAll("-", "/");
}

function toNullableText(value: string) {
  const trimmedValue = value.trim();

  if (!trimmedValue) {
    return null;
  }

  return trimmedValue;
}

export default function AdminExternalStorageLinksPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [keyword, setKeyword] = useState("");
  const [linkType, setLinkType] = useState("");
  const [searchedKeyword, setSearchedKeyword] = useState("");
  const [searchedLinkType, setSearchedLinkType] = useState("");

  const [externalStorageLinks, setExternalStorageLinks] = useState<ExternalStorageLinkResponse[]>([]);
  const [externalStorageLinkOffset, setExternalStorageLinkOffset] = useState(0);
  const [externalStorageLinkHasMore, setExternalStorageLinkHasMore] = useState(false);

  const [form, setForm] = useState<ExternalStorageLinkForm>(initialForm);
  const [isEditing, setIsEditing] = useState(false);

  const [pageMessage, setPageMessage] = useState(
    "固定された外部ストレージリンクのURL、説明、管理メモを編集できます。"
  );
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const [isPageLoading, setIsPageLoading] = useState(false);
  const [isExternalStorageLinkSearching, setIsExternalStorageLinkSearching] = useState(false);

  const formTitle = useMemo(() => {
    return isEditing ? "外部ストレージリンク編集" : "外部ストレージリンク設定";
  }, [isEditing]);

  const formDescription = useMemo(() => {
    return isEditing
      ? "選択した固定リンクのURL、説明、管理メモを更新します。"
      : "左側の一覧から編集するリンクを選択してください。";
  }, [isEditing]);

  const resetForm = () => {
    setForm(initialForm);
    setIsEditing(false);
  };

  const handleSearchExternalStorageLinks = async (
    nextOffset: number,
    append: boolean
  ) => {
    setIsExternalStorageLinkSearching(true);
    setPageMessage("外部ストレージリンクを検索しています。");
    setPageMessageVariant("info");

    const searchKeyword = append ? searchedKeyword : keyword;
    const searchLinkType = append ? searchedLinkType : linkType;

    const result = await searchExternalStorageLinks({
      keyword: searchKeyword,
      linkType: searchLinkType,
      includeDeleted: false,
      offset: nextOffset,
      limit: 50,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "外部ストレージリンク検索に失敗しました。");
      setPageMessageVariant("error");
      setIsExternalStorageLinkSearching(false);
      return;
    }

    const searchData = result.data;

    setSearchedKeyword(searchKeyword);
    setSearchedLinkType(searchLinkType);
    setExternalStorageLinks((current) =>
      append ? [...current, ...searchData.externalStorageLinks] : searchData.externalStorageLinks
    );
    setExternalStorageLinkOffset(nextOffset + searchData.externalStorageLinks.length);
    setExternalStorageLinkHasMore(searchData.hasMore);

    setPageMessage(
      searchData.externalStorageLinks.length === 0
        ? "該当する外部ストレージリンクが見つかりませんでした。"
        : "外部ストレージリンク検索が完了しました。"
    );
    setPageMessageVariant(searchData.externalStorageLinks.length === 0 ? "warning" : "success");
    setIsExternalStorageLinkSearching(false);
  };

  const handleLoadMoreExternalStorageLinks = async () => {
    await handleSearchExternalStorageLinks(externalStorageLinkOffset, true);
  };

  const handleStartEdit = (externalStorageLink: ExternalStorageLinkResponse) => {
    if (externalStorageLink.isDeleted) {
      setPageMessage("削除済みの外部ストレージリンクは編集できません。");
      setPageMessageVariant("warning");
      return;
    }

    setForm({
      externalStorageLinkId: externalStorageLink.id,
      linkType: externalStorageLink.linkType,
      linkName: externalStorageLink.linkName,
      url: externalStorageLink.url,
      description: externalStorageLink.description ?? "",
      memo: externalStorageLink.memo ?? "",
    });

    setIsEditing(true);
    setPageMessage("外部ストレージリンク情報を編集できます。");
    setPageMessageVariant("info");
  };

  const handleOpenLink = (url: string) => {
    const trimmedUrl = url.trim();

    if (!trimmedUrl) {
      setPageMessage("URLが設定されていません。");
      setPageMessageVariant("warning");
      return;
    }

    window.open(trimmedUrl, "_blank", "noopener,noreferrer");
  };

  const handleUpdateExternalStorageLink = async () => {
    if (form.externalStorageLinkId === null) {
      setPageMessage("更新対象の外部ストレージリンクが選択されていません。");
      setPageMessageVariant("error");
      return;
    }

    setIsPageLoading(true);
    setPageMessage("外部ストレージリンクを更新しています。");
    setPageMessageVariant("info");

    const result = await updateExternalStorageLink({
      externalStorageLinkId: form.externalStorageLinkId,
      url: form.url.trim(),
      description: toNullableText(form.description),
      memo: toNullableText(form.memo),
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "外部ストレージリンクの更新に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    const updatedExternalStorageLink = result.data.externalStorageLink;

    setExternalStorageLinks((current) =>
      current.map((externalStorageLink) =>
        externalStorageLink.id === updatedExternalStorageLink.id ? updatedExternalStorageLink : externalStorageLink
      )
    );

    setForm({
      externalStorageLinkId: updatedExternalStorageLink.id,
      linkType: updatedExternalStorageLink.linkType,
      linkName: updatedExternalStorageLink.linkName,
      url: updatedExternalStorageLink.url,
      description: updatedExternalStorageLink.description ?? "",
      memo: updatedExternalStorageLink.memo ?? "",
    });

    setPageMessage(result.message || "外部ストレージリンクを更新しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void handleSearchExternalStorageLinks(0, false);
    }, 0);

    return () => {
      window.clearTimeout(timerId);
    };

    // 初回だけ実行したいので依存配列は固定する
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, user]);

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="外部ストレージリンク管理" description="ログイン情報を確認しています。" />
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
              title="外部ストレージリンク管理"
              description="Google Driveなど、用途ごとに固定された外部ストレージリンクを設定します。"
            />

            <MessageBox variant={pageMessageVariant}>{isPageLoading ? "処理中..." : pageMessage}</MessageBox>
          </div>

          <div className={styles.contentGrid}>
            <section className={styles.searchPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>リンク一覧</h2>
                  <p className={styles.sectionDescription}>
                    固定された用途ごとのリンクを表示します。編集ボタンからURL、説明、管理メモを変更できます。
                  </p>
                </div>
              </div>

              <div className={styles.searchForm}>
                <input
                  className={styles.searchInput}
                  value={keyword}
                  placeholder="リンク名・URLなどで検索"
                  aria-label="リンク名・URLなどで検索"
                  onChange={(event) => setKeyword(event.target.value)}
                />

                <input
                  className={styles.searchInput}
                  value={linkType}
                  placeholder="リンク種別で絞り込み"
                  aria-label="リンク種別で絞り込み"
                  onChange={(event) => setLinkType(event.target.value)}
                />

                <Button
                  type="button"
                  variant="primary"
                  onClick={() => handleSearchExternalStorageLinks(0, false)}
                  disabled={isExternalStorageLinkSearching}
                >
                  検索
                </Button>
              </div>

              <div className={styles.linkList}>
                {externalStorageLinks.map((externalStorageLink) => (
                  <article
                    key={externalStorageLink.id}
                    className={`${styles.linkRow} ${
                      form.externalStorageLinkId === externalStorageLink.id ? styles.linkRowSelected : ""
                    } ${externalStorageLink.isDeleted ? styles.linkRowDeleted : ""}`}
                  >
                    <div className={styles.linkRowMain}>
                      <div className={styles.linkInfo}>
                        <div className={styles.linkNameLine}>
                          <p className={styles.linkName}>{externalStorageLink.linkName}</p>

                          <span className={styles.linkTypeBadge}>{externalStorageLink.linkType}</span>

                          {externalStorageLink.isDeleted && <span className={styles.deletedBadge}>削除済み</span>}
                        </div>

                        <p className={styles.linkUrl}>{externalStorageLink.url || "URL未設定"}</p>

                        {externalStorageLink.description && (
                          <p className={styles.linkMeta}>説明：{externalStorageLink.description}</p>
                        )}

                        <p className={styles.linkMeta}>
                          ID：{externalStorageLink.id} / 作成日：{formatDate(externalStorageLink.createdAt)} / 更新日：
                          {formatDate(externalStorageLink.updatedAt)}
                        </p>
                      </div>

                      <div className={styles.rowActions}>
                        <Button
                          type="button"
                          variant="secondary"
                          onClick={() => handleOpenLink(externalStorageLink.url)}
                          disabled={!externalStorageLink.url.trim()}
                        >
                          開く
                        </Button>

                        <Button
                          type="button"
                          variant="primary"
                          onClick={() => handleStartEdit(externalStorageLink)}
                          disabled={externalStorageLink.isDeleted}
                        >
                          編集
                        </Button>
                      </div>
                    </div>
                  </article>
                ))}

                {externalStorageLinks.length === 0 && (
                  <p className={styles.emptyText}>外部ストレージリンクが見つかりません。</p>
                )}
              </div>

              {externalStorageLinkHasMore && (
                <div className={styles.moreButtonArea}>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={handleLoadMoreExternalStorageLinks}
                    disabled={isExternalStorageLinkSearching}
                  >
                    さらに表示
                  </Button>
                </div>
              )}
            </section>

            <section className={styles.formPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>{formTitle}</h2>
                  <p className={styles.sectionDescription}>{formDescription}</p>
                </div>

                {isEditing && (
                  <Button type="button" variant="secondary" onClick={resetForm}>
                    編集取消
                  </Button>
                )}
              </div>

              {!isEditing && (
                <div className={styles.notSelectedBox}>
                  <p className={styles.notSelectedTitle}>編集対象が選択されていません</p>
                  <p className={styles.notSelectedText}>左側のリンク一覧から「編集」を押してください。</p>
                </div>
              )}

              {isEditing && (
                <>
                  <div className={styles.readOnlyBox}>
                    <div>
                      <p className={styles.readOnlyLabel}>リンク名</p>
                      <p className={styles.readOnlyValue}>{form.linkName}</p>
                    </div>

                    <div>
                      <p className={styles.readOnlyLabel}>リンク種別</p>
                      <p className={styles.readOnlyValue}>{form.linkType}</p>
                    </div>
                  </div>

                  <div className={styles.formGrid}>
                    <label className={styles.fieldWide}>
                      <span className={styles.fieldLabel}>URL</span>
                      <input
                        className={styles.input}
                        value={form.url}
                        placeholder="例：https://drive.google.com/..."
                        onChange={(event) => setForm((current) => ({ ...current, url: event.target.value }))}
                      />
                    </label>

                    <label className={styles.fieldWide}>
                      <span className={styles.fieldLabel}>説明</span>
                      <textarea
                        className={`${styles.input} ${styles.textarea}`}
                        value={form.description}
                        placeholder="このリンクの用途を入力"
                        onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))}
                      />
                    </label>

                    <label className={styles.fieldWide}>
                      <span className={styles.fieldLabel}>管理メモ</span>
                      <textarea
                        className={`${styles.input} ${styles.textarea}`}
                        value={form.memo}
                        placeholder="管理者用メモ"
                        onChange={(event) => setForm((current) => ({ ...current, memo: event.target.value }))}
                      />
                    </label>
                  </div>

                  <div className={styles.formActions}>
                    <Button type="button" variant="primary" onClick={handleUpdateExternalStorageLink} disabled={isPageLoading}>
                      更新
                    </Button>
                  </div>
                </>
              )}
            </section>
          </div>
        </section>
      </div>
    </PageContainer>
  );
}