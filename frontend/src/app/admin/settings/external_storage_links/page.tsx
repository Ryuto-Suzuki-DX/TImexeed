"use client";

import { useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import {
  createExternalStorageLink,
  deleteExternalStorageLink,
  getExternalStorageLinkDetail,
  searchExternalStorageLinks,
  updateExternalStorageLink,
} from "@/api/admin/externalStorageLink";
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
  const [includeDeleted, setIncludeDeleted] = useState(false);

  const [externalStorageLinks, setExternalStorageLinks] = useState<ExternalStorageLinkResponse[]>([]);
  const [externalStorageLinkOffset, setExternalStorageLinkOffset] = useState(0);
  const [externalStorageLinkHasMore, setExternalStorageLinkHasMore] = useState(false);

  const [form, setForm] = useState<ExternalStorageLinkForm>(initialForm);
  const [isEditing, setIsEditing] = useState(false);

  const [pageMessage, setPageMessage] = useState("外部ストレージリンクを検索・作成・編集できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const [isPageLoading, setIsPageLoading] = useState(false);
  const [isExternalStorageLinkSearching, setIsExternalStorageLinkSearching] = useState(false);

  const formTitle = useMemo(() => {
    return isEditing ? "外部ストレージリンク編集" : "外部ストレージリンク新規作成";
  }, [isEditing]);

  const formDescription = useMemo(() => {
    return isEditing
      ? "選択した外部ストレージリンクを更新します。"
      : "Google DriveなどのフォルダURL・ファイルURLを登録します。";
  }, [isEditing]);

  const resetForm = () => {
    setForm(initialForm);
    setIsEditing(false);
  };

  const handleSearchExternalStorageLinks = async (
    nextOffset: number,
    append: boolean,
    includeDeletedValue = includeDeleted
  ) => {
    setIsExternalStorageLinkSearching(true);
    setPageMessage("外部ストレージリンクを検索しています。");
    setPageMessageVariant("info");

    const searchKeyword = append ? searchedKeyword : keyword;
    const searchLinkType = append ? searchedLinkType : linkType;

    const result = await searchExternalStorageLinks({
      keyword: searchKeyword,
      linkType: searchLinkType,
      includeDeleted: includeDeletedValue,
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

  const handleToggleIncludeDeleted = async () => {
    const nextIncludeDeleted = !includeDeleted;

    setIncludeDeleted(nextIncludeDeleted);
    resetForm();
    await handleSearchExternalStorageLinks(0, false, nextIncludeDeleted);
  };

  const handleLoadMoreExternalStorageLinks = async () => {
    await handleSearchExternalStorageLinks(externalStorageLinkOffset, true);
  };

  const handleStartCreate = () => {
    resetForm();
    setPageMessage("外部ストレージリンク情報を入力してください。");
    setPageMessageVariant("info");
  };

  const handleStartEdit = async (externalStorageLink: ExternalStorageLinkResponse) => {
    if (externalStorageLink.isDeleted) {
      setPageMessage("削除済みの外部ストレージリンクは編集できません。");
      setPageMessageVariant("warning");
      return;
    }

    setIsPageLoading(true);
    setPageMessage("外部ストレージリンク詳細を取得しています。");
    setPageMessageVariant("info");

    const result = await getExternalStorageLinkDetail({
      externalStorageLinkId: externalStorageLink.id,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "外部ストレージリンク詳細の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    const detailData = result.data;
    const externalStorageLinkDetail = detailData.externalStorageLink;

    setForm({
      externalStorageLinkId: externalStorageLinkDetail.id,
      linkType: externalStorageLinkDetail.linkType,
      linkName: externalStorageLinkDetail.linkName,
      url: externalStorageLinkDetail.url,
      description: externalStorageLinkDetail.description ?? "",
      memo: externalStorageLinkDetail.memo ?? "",
    });

    setIsEditing(true);
    setPageMessage("外部ストレージリンク情報を編集できます。");
    setPageMessageVariant("info");
    setIsPageLoading(false);
  };

  const validateForm = () => {
    if (!form.linkType.trim()) {
      setPageMessage("リンク種別を入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    if (!form.linkName.trim()) {
      setPageMessage("リンク名を入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    if (!form.url.trim()) {
      setPageMessage("URLを入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    return true;
  };

  const handleCreateExternalStorageLink = async () => {
    if (!validateForm()) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("外部ストレージリンクを作成しています。");
    setPageMessageVariant("info");

    const result = await createExternalStorageLink({
      linkType: form.linkType,
      linkName: form.linkName,
      url: form.url,
      description: toNullableText(form.description),
      memo: toNullableText(form.memo),
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "外部ストレージリンクの作成に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetForm();
    await handleSearchExternalStorageLinks(0, false);

    setPageMessage(result.message || "外部ストレージリンクを作成しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  const handleUpdateExternalStorageLink = async () => {
    if (!validateForm()) {
      return;
    }

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
      linkType: form.linkType,
      linkName: form.linkName,
      url: form.url,
      description: toNullableText(form.description),
      memo: toNullableText(form.memo),
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "外部ストレージリンクの更新に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetForm();
    await handleSearchExternalStorageLinks(0, false);

    setPageMessage(result.message || "外部ストレージリンクを更新しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  const handleDeleteExternalStorageLink = async (externalStorageLink: ExternalStorageLinkResponse) => {
    if (externalStorageLink.isDeleted) {
      setPageMessage("この外部ストレージリンクはすでに削除済みです。");
      setPageMessageVariant("warning");
      return;
    }

    const confirmed = window.confirm(`${externalStorageLink.linkName} を削除しますか？`);

    if (!confirmed) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("外部ストレージリンクを削除しています。");
    setPageMessageVariant("info");

    const result = await deleteExternalStorageLink({
      externalStorageLinkId: externalStorageLink.id,
    });

    if (result.error) {
      setPageMessage(result.message || "外部ストレージリンクの削除に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    if (form.externalStorageLinkId === externalStorageLink.id) {
      resetForm();
    }

    await handleSearchExternalStorageLinks(0, false);

    setPageMessage(result.message || "外部ストレージリンクを削除しました。");
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
              description="Google DriveなどのフォルダURL・ファイルURLを管理します。"
            />

            <MessageBox variant={pageMessageVariant}>{isPageLoading ? "処理中..." : pageMessage}</MessageBox>
          </div>

          <div className={styles.contentGrid}>
            <section className={styles.searchPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>外部ストレージリンク検索</h2>
                  <p className={styles.sectionDescription}>リンク名、リンク種別、URLなどで検索できます。</p>
                </div>

                <Button type="button" variant="secondary" onClick={handleStartCreate}>
                  新規作成
                </Button>
              </div>

              <div className={styles.searchForm}>
                <input
                  className={styles.searchInput}
                  value={keyword}
                  placeholder="リンク名・URLなどで検索"
                  onChange={(event) => setKeyword(event.target.value)}
                />

                <input
                  className={styles.searchInput}
                  value={linkType}
                  placeholder="リンク種別で絞り込み"
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

              <button type="button" className={styles.switchRow} onClick={handleToggleIncludeDeleted}>
                <span className={`${styles.switch} ${includeDeleted ? styles.switchOn : ""}`}>
                  <span className={styles.switchThumb} />
                </span>

                <span className={styles.switchText}>削除済みリンクも含める</span>
              </button>

              <div className={styles.departmentList}>
                {externalStorageLinks.map((externalStorageLink) => (
                  <article
                    key={externalStorageLink.id}
                    className={`${styles.departmentRow} ${externalStorageLink.isDeleted ? styles.departmentRowDeleted : ""}`}
                  >
                    <div className={styles.departmentRowMain}>
                      <div>
                        <div className={styles.departmentNameLine}>
                          <p className={styles.departmentName}>{externalStorageLink.linkName}</p>

                          {externalStorageLink.isDeleted && <span className={styles.deletedBadge}>削除済み</span>}
                        </div>

                        <p className={styles.departmentMeta}>リンク種別：{externalStorageLink.linkType}</p>

                        <p className={styles.departmentMeta}>URL：{externalStorageLink.url}</p>

                        {externalStorageLink.description && (
                          <p className={styles.departmentMeta}>説明：{externalStorageLink.description}</p>
                        )}

                        <p className={styles.departmentMeta}>
                          ID：{externalStorageLink.id} / 作成日：{formatDate(externalStorageLink.createdAt)} / 更新日：
                          {formatDate(externalStorageLink.updatedAt)}
                        </p>

                        {externalStorageLink.deletedAt && (
                          <p className={styles.departmentMeta}>削除日：{formatDate(externalStorageLink.deletedAt)}</p>
                        )}
                      </div>

                      <div className={styles.rowActions}>
                        <Button
                          type="button"
                          variant="secondary"
                          onClick={() => handleStartEdit(externalStorageLink)}
                          disabled={externalStorageLink.isDeleted}
                        >
                          編集
                        </Button>

                        <Button
                          type="button"
                          variant="danger"
                          onClick={() => handleDeleteExternalStorageLink(externalStorageLink)}
                          disabled={externalStorageLink.isDeleted}
                        >
                          削除
                        </Button>
                      </div>
                    </div>
                  </article>
                ))}

                {externalStorageLinks.length === 0 && <p className={styles.emptyText}>外部ストレージリンクが見つかりません。</p>}
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

              <div className={styles.formGrid}>
                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>リンク種別</span>
                  <input
                    className={styles.input}
                    value={form.linkType}
                    placeholder="例：EXPENSE_RECEIPT_BOX"
                    onChange={(event) => setForm((current) => ({ ...current, linkType: event.target.value }))}
                  />
                </label>

                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>リンク名</span>
                  <input
                    className={styles.input}
                    value={form.linkName}
                    placeholder="例：経費レシート格納先"
                    onChange={(event) => setForm((current) => ({ ...current, linkName: event.target.value }))}
                  />
                </label>

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
                    className={styles.input}
                    value={form.description}
                    placeholder="このリンクの用途を入力"
                    onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))}
                  />
                </label>

                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>管理メモ</span>
                  <textarea
                    className={styles.input}
                    value={form.memo}
                    placeholder="管理者用メモ"
                    onChange={(event) => setForm((current) => ({ ...current, memo: event.target.value }))}
                  />
                </label>
              </div>

              <div className={styles.formActions}>
                {isEditing ? (
                  <Button type="button" variant="primary" onClick={handleUpdateExternalStorageLink} disabled={isPageLoading}>
                    更新
                  </Button>
                ) : (
                  <Button type="button" variant="primary" onClick={handleCreateExternalStorageLink} disabled={isPageLoading}>
                    作成
                  </Button>
                )}
              </div>
            </section>
          </div>
        </section>
      </div>
    </PageContainer>
  );
}