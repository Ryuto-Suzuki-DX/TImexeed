"use client";

import { useCallback, useEffect, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import {
  createAllowanceType,
  deleteAllowanceType,
  getAllowanceTypeDetail,
  searchAllowanceTypes,
  updateAllowanceType,
} from "@/api/admin/allowanceType";
import type { AllowanceTypeResponse } from "@/types/admin/allowanceType";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

type AllowanceTypeForm = {
  allowanceTypeId: number | null;
  name: string;
  displayOrder: string;
  isActive: boolean;
};

const initialForm: AllowanceTypeForm = {
  allowanceTypeId: null,
  name: "",
  displayOrder: "0",
  isActive: true,
};

export default function AdminAllowanceTypesPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [keyword, setKeyword] = useState("");
  const [searchedKeyword, setSearchedKeyword] = useState("");
  const [includeInactive, setIncludeInactive] = useState(true);
  const [includeDeleted, setIncludeDeleted] = useState(false);
  const [allowanceTypes, setAllowanceTypes] = useState<AllowanceTypeResponse[]>([]);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [form, setForm] = useState<AllowanceTypeForm>(initialForm);
  const [pageMessage, setPageMessage] = useState("手当種別を検索、追加、編集できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");
  const [isPageLoading, setIsPageLoading] = useState(false);

  const resetForm = () => {
    setForm(initialForm);
  };

  const loadAllowanceTypes = useCallback(
    async (nextOffset: number, append: boolean, keywordValue = searchedKeyword) => {
      setIsPageLoading(true);
      setPageMessage("手当種別を取得しています。");
      setPageMessageVariant("info");

      const result = await searchAllowanceTypes({
        keyword: keywordValue,
        includeInactive,
        includeDeleted,
        offset: nextOffset,
        limit: 50,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "手当種別の取得に失敗しました。");
        setPageMessageVariant("error");
        setIsPageLoading(false);
        return;
      }

      setAllowanceTypes((current) =>
        append ? [...current, ...result.data!.allowanceTypes] : result.data!.allowanceTypes,
      );
      setOffset(nextOffset + result.data.allowanceTypes.length);
      setHasMore(result.data.hasMore);
      setPageMessage(
        result.data.allowanceTypes.length === 0 ? "手当種別が見つかりませんでした。" : "手当種別を取得しました。",
      );
      setPageMessageVariant(result.data.allowanceTypes.length === 0 ? "warning" : "success");
      setIsPageLoading(false);
    },
    [includeDeleted, includeInactive, searchedKeyword],
  );

  const handleSearch = async () => {
    setSearchedKeyword(keyword);
    resetForm();
    await loadAllowanceTypes(0, false, keyword);
  };

  const handleStartEdit = async (allowanceType: AllowanceTypeResponse) => {
    if (allowanceType.isDeleted) {
      setPageMessage("削除済みの手当種別は編集できません。");
      setPageMessageVariant("warning");
      return;
    }

    setIsPageLoading(true);
    const result = await getAllowanceTypeDetail({ allowanceTypeId: allowanceType.id });

    if (result.error || !result.data) {
      setPageMessage(result.message || "手当種別の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    const detail = result.data.allowanceType;
    setForm({
      allowanceTypeId: detail.id,
      name: detail.name,
      displayOrder: String(detail.displayOrder),
      isActive: detail.isActive,
    });
    setPageMessage("手当種別を編集できます。");
    setPageMessageVariant("info");
    setIsPageLoading(false);
  };

  const validateForm = () => {
    if (!form.name.trim()) {
      setPageMessage("手当種別名を入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    const displayOrder = Number(form.displayOrder);
    if (!Number.isInteger(displayOrder) || displayOrder < 0) {
      setPageMessage("表示順は0以上の整数で入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    return true;
  };

  const handleSave = async () => {
    if (!validateForm()) {
      return;
    }

    setIsPageLoading(true);
    const request = {
      name: form.name.trim(),
      displayOrder: Number(form.displayOrder),
      isActive: form.isActive,
    };

    const result =
      form.allowanceTypeId === null
        ? await createAllowanceType(request)
        : await updateAllowanceType({
            allowanceTypeId: form.allowanceTypeId,
            ...request,
          });

    if (result.error || !result.data) {
      setPageMessage(result.message || "手当種別の保存に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetForm();
    await loadAllowanceTypes(0, false);
    setPageMessage(result.message || "手当種別を保存しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  const handleDelete = async (allowanceType: AllowanceTypeResponse) => {
    if (allowanceType.isDeleted) {
      return;
    }

    const confirmed = window.confirm(`「${allowanceType.name}」を削除しますか？`);
    if (!confirmed) {
      return;
    }

    setIsPageLoading(true);
    const result = await deleteAllowanceType({ allowanceTypeId: allowanceType.id });

    if (result.error) {
      setPageMessage(result.message || "手当種別の削除に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    if (form.allowanceTypeId === allowanceType.id) {
      resetForm();
    }

    await loadAllowanceTypes(0, false);
    setPageMessage(result.message || "手当種別を削除しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void loadAllowanceTypes(0, false, "");
    }, 0);

    return () => window.clearTimeout(timerId);
    // 初回のみ実行する
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, user]);

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />
        <section className={styles.loadingCard}>
          <PageTitle title="手当種別管理" description="ログイン情報を確認しています。" />
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
              title="手当種別管理"
              description="月次手当入力で選択できる手当種別と表示順を管理します。"
            />
            <MessageBox variant={pageMessageVariant}>{isPageLoading ? "処理中..." : pageMessage}</MessageBox>
          </div>

          <div className={styles.contentGrid}>
            <section className={styles.listPanel}>
              <div className={styles.searchForm}>
                <input
                  className={styles.input}
                  value={keyword}
                  placeholder="手当種別名で検索"
                  onChange={(event) => setKeyword(event.target.value)}
                />
                <Button type="button" variant="primary" onClick={handleSearch} disabled={isPageLoading}>
                  検索
                </Button>
              </div>

              <div className={styles.filterRows}>
                <label className={styles.checkboxRow}>
                  <input
                    type="checkbox"
                    checked={includeInactive}
                    onChange={(event) => setIncludeInactive(event.target.checked)}
                  />
                  利用停止中も含める
                </label>
                <label className={styles.checkboxRow}>
                  <input
                    type="checkbox"
                    checked={includeDeleted}
                    onChange={(event) => setIncludeDeleted(event.target.checked)}
                  />
                  削除済みも含める
                </label>
              </div>

              <div className={styles.tableScroll}>
                <table className={styles.table}>
                  <thead>
                    <tr>
                      <th>表示順</th>
                      <th>手当種別名</th>
                      <th>状態</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {allowanceTypes.map((allowanceType) => (
                      <tr key={allowanceType.id}>
                        <td>{allowanceType.displayOrder}</td>
                        <td>{allowanceType.name}</td>
                        <td>
                          <span
                            className={
                              allowanceType.isDeleted
                                ? styles.deletedBadge
                                : allowanceType.isActive
                                  ? styles.activeBadge
                                  : styles.inactiveBadge
                            }
                          >
                            {allowanceType.isDeleted ? "削除済み" : allowanceType.isActive ? "利用中" : "利用停止"}
                          </span>
                        </td>
                        <td>
                          <div className={styles.rowActions}>
                            <Button
                              type="button"
                              variant="secondary"
                              onClick={() => handleStartEdit(allowanceType)}
                              disabled={allowanceType.isDeleted}
                            >
                              編集
                            </Button>
                            <Button
                              type="button"
                              variant="danger"
                              onClick={() => handleDelete(allowanceType)}
                              disabled={allowanceType.isDeleted}
                            >
                              削除
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                    {allowanceTypes.length === 0 && (
                      <tr>
                        <td colSpan={4} className={styles.emptyCell}>
                          手当種別が登録されていません。
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>

              {hasMore && (
                <div className={styles.moreButtonArea}>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => loadAllowanceTypes(offset, true)}
                    disabled={isPageLoading}
                  >
                    さらに表示
                  </Button>
                </div>
              )}
            </section>

            <section className={styles.formPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>
                    {form.allowanceTypeId === null ? "手当種別新規作成" : "手当種別編集"}
                  </h2>
                  <p className={styles.sectionDescription}>
                    表示順は、月次手当のドロップダウンと集計CSVの列順に使用します。
                  </p>
                </div>
                {form.allowanceTypeId !== null && (
                  <Button type="button" variant="secondary" onClick={resetForm}>
                    編集取消
                  </Button>
                )}
              </div>

              <div className={styles.formGrid}>
                <label className={styles.field}>
                  <span className={styles.fieldLabel}>手当種別名</span>
                  <input
                    className={styles.input}
                    value={form.name}
                    placeholder="例：固定残業手当"
                    onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))}
                  />
                </label>

                <label className={styles.field}>
                  <span className={styles.fieldLabel}>表示順</span>
                  <input
                    className={styles.input}
                    type="number"
                    min="0"
                    step="1"
                    value={form.displayOrder}
                    onChange={(event) =>
                      setForm((current) => ({ ...current, displayOrder: event.target.value }))
                    }
                  />
                </label>

                <label className={styles.checkboxRow}>
                  <input
                    type="checkbox"
                    checked={form.isActive}
                    onChange={(event) =>
                      setForm((current) => ({ ...current, isActive: event.target.checked }))
                    }
                  />
                  新規登録時に選択できる状態にする
                </label>
              </div>

              <div className={styles.formActions}>
                <Button type="button" variant="primary" onClick={handleSave} disabled={isPageLoading}>
                  {form.allowanceTypeId === null ? "作成" : "更新"}
                </Button>
              </div>
            </section>
          </div>
        </section>
      </div>
    </PageContainer>
  );
}
