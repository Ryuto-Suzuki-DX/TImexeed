"use client";

import { useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import {
  createDepartment,
  deleteDepartment,
  getDepartmentDetail,
  searchDepartments,
  updateDepartment,
} from "@/api/admin/department";
import type { DepartmentResponse } from "@/types/admin/department";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

type DepartmentForm = {
  departmentId: number | null;
  name: string;
};

const initialForm: DepartmentForm = {
  departmentId: null,
  name: "",
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

export default function AdminDepartmentsPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [keyword, setKeyword] = useState("");
  const [searchedKeyword, setSearchedKeyword] = useState("");
  const [includeDeleted, setIncludeDeleted] = useState(false);

  const [departments, setDepartments] = useState<DepartmentResponse[]>([]);
  const [departmentOffset, setDepartmentOffset] = useState(0);
  const [departmentHasMore, setDepartmentHasMore] = useState(false);

  const [form, setForm] = useState<DepartmentForm>(initialForm);
  const [isEditing, setIsEditing] = useState(false);

  const [pageMessage, setPageMessage] = useState("所属を検索・作成・編集できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const [isPageLoading, setIsPageLoading] = useState(false);
  const [isDepartmentSearching, setIsDepartmentSearching] = useState(false);

  const formTitle = useMemo(() => {
    return isEditing ? "所属編集" : "所属新規作成";
  }, [isEditing]);

  const formDescription = useMemo(() => {
    return isEditing ? "選択した所属名を更新します。" : "新しい所属を作成します。";
  }, [isEditing]);

  const resetForm = () => {
    setForm(initialForm);
    setIsEditing(false);
  };

  const handleSearchDepartments = async (nextOffset: number, append: boolean, includeDeletedValue = includeDeleted) => {
    setIsDepartmentSearching(true);
    setPageMessage("所属を検索しています。");
    setPageMessageVariant("info");

    const searchKeyword = append ? searchedKeyword : keyword;

    const result = await searchDepartments({
      keyword: searchKeyword,
      includeDeleted: includeDeletedValue,
      offset: nextOffset,
      limit: 50,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "所属検索に失敗しました。");
      setPageMessageVariant("error");
      setIsDepartmentSearching(false);
      return;
    }

    const searchData = result.data;

    setSearchedKeyword(searchKeyword);
    setDepartments((current) => (append ? [...current, ...searchData.departments] : searchData.departments));
    setDepartmentOffset(nextOffset + searchData.departments.length);
    setDepartmentHasMore(searchData.hasMore);

    setPageMessage(searchData.departments.length === 0 ? "該当する所属が見つかりませんでした。" : "所属検索が完了しました。");
    setPageMessageVariant(searchData.departments.length === 0 ? "warning" : "success");
    setIsDepartmentSearching(false);
  };

  const handleToggleIncludeDeleted = async () => {
    const nextIncludeDeleted = !includeDeleted;

    setIncludeDeleted(nextIncludeDeleted);
    resetForm();
    await handleSearchDepartments(0, false, nextIncludeDeleted);
  };

  const handleLoadMoreDepartments = async () => {
    await handleSearchDepartments(departmentOffset, true);
  };

  const handleStartCreate = () => {
    resetForm();
    setPageMessage("新規所属名を入力してください。");
    setPageMessageVariant("info");
  };

  const handleStartEdit = async (department: DepartmentResponse) => {
    if (department.isDeleted) {
      setPageMessage("削除済み所属は編集できません。");
      setPageMessageVariant("warning");
      return;
    }

    setIsPageLoading(true);
    setPageMessage("所属詳細を取得しています。");
    setPageMessageVariant("info");

    const result = await getDepartmentDetail({
      departmentId: department.id,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "所属詳細の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    const detailData = result.data;
    const departmentDetail = detailData.department;

    setForm({
      departmentId: departmentDetail.id,
      name: departmentDetail.name,
    });

    setIsEditing(true);
    setPageMessage("所属情報を編集できます。");
    setPageMessageVariant("info");
    setIsPageLoading(false);
  };

  const validateForm = () => {
    if (!form.name.trim()) {
      setPageMessage("所属名を入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    return true;
  };

  const handleCreateDepartment = async () => {
    if (!validateForm()) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("所属を作成しています。");
    setPageMessageVariant("info");

    const result = await createDepartment({
      name: form.name,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "所属の作成に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetForm();
    await handleSearchDepartments(0, false);

    setPageMessage(result.message || "所属を作成しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  const handleUpdateDepartment = async () => {
    if (!validateForm()) {
      return;
    }

    if (form.departmentId === null) {
      setPageMessage("更新対象の所属が選択されていません。");
      setPageMessageVariant("error");
      return;
    }

    setIsPageLoading(true);
    setPageMessage("所属を更新しています。");
    setPageMessageVariant("info");

    const result = await updateDepartment({
      departmentId: form.departmentId,
      name: form.name,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "所属の更新に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetForm();
    await handleSearchDepartments(0, false);

    setPageMessage(result.message || "所属を更新しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  const handleDeleteDepartment = async (department: DepartmentResponse) => {
    if (department.isDeleted) {
      setPageMessage("この所属はすでに削除済みです。");
      setPageMessageVariant("warning");
      return;
    }

    const confirmed = window.confirm(`${department.name} を削除しますか？`);

    if (!confirmed) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("所属を削除しています。");
    setPageMessageVariant("info");

    const result = await deleteDepartment({
      departmentId: department.id,
    });

    if (result.error) {
      setPageMessage(result.message || "所属の削除に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    if (form.departmentId === department.id) {
      resetForm();
    }

    await handleSearchDepartments(0, false);

    setPageMessage(result.message || "所属を削除しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void handleSearchDepartments(0, false);
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
          <PageTitle title="所属管理" description="ログイン情報を確認しています。" />
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
            <PageTitle title="所属管理" description="所属の検索、作成、編集、削除を行います。" />

            <MessageBox variant={pageMessageVariant}>{isPageLoading ? "処理中..." : pageMessage}</MessageBox>
          </div>

          <div className={styles.contentGrid}>
            <section className={styles.searchPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>所属検索</h2>
                  <p className={styles.sectionDescription}>所属名で検索できます。</p>
                </div>

                <Button type="button" variant="secondary" onClick={handleStartCreate}>
                  新規作成
                </Button>
              </div>

              <div className={styles.searchForm}>
                <input
                  className={styles.searchInput}
                  value={keyword}
                  placeholder="所属名で検索"
                  onChange={(event) => setKeyword(event.target.value)}
                />

                <Button type="button" variant="primary" onClick={() => handleSearchDepartments(0, false)} disabled={isDepartmentSearching}>
                  検索
                </Button>
              </div>

              <button type="button" className={styles.switchRow} onClick={handleToggleIncludeDeleted}>
                <span className={`${styles.switch} ${includeDeleted ? styles.switchOn : ""}`}>
                  <span className={styles.switchThumb} />
                </span>

                <span className={styles.switchText}>削除済み所属も含める</span>
              </button>

              <div className={styles.departmentList}>
                {departments.map((department) => (
                  <article key={department.id} className={`${styles.departmentRow} ${department.isDeleted ? styles.departmentRowDeleted : ""}`}>
                    <div className={styles.departmentRowMain}>
                      <div>
                        <div className={styles.departmentNameLine}>
                          <p className={styles.departmentName}>{department.name}</p>

                          {department.isDeleted && <span className={styles.deletedBadge}>削除済み</span>}
                        </div>

                        <p className={styles.departmentMeta}>
                          ID：{department.id} / 作成日：{formatDate(department.createdAt)} / 更新日：{formatDate(department.updatedAt)}
                        </p>

                        {department.deletedAt && <p className={styles.departmentMeta}>削除日：{formatDate(department.deletedAt)}</p>}
                      </div>

                      <div className={styles.rowActions}>
                        <Button type="button" variant="secondary" onClick={() => handleStartEdit(department)} disabled={department.isDeleted}>
                          編集
                        </Button>

                        <Button type="button" variant="danger" onClick={() => handleDeleteDepartment(department)} disabled={department.isDeleted}>
                          削除
                        </Button>
                      </div>
                    </div>
                  </article>
                ))}

                {departments.length === 0 && <p className={styles.emptyText}>所属が見つかりません。</p>}
              </div>

              {departmentHasMore && (
                <div className={styles.moreButtonArea}>
                  <Button type="button" variant="secondary" onClick={handleLoadMoreDepartments} disabled={isDepartmentSearching}>
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
                  <span className={styles.fieldLabel}>所属名</span>
                  <input
                    className={styles.input}
                    value={form.name}
                    placeholder="例：開発部"
                    onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))}
                  />
                </label>
              </div>

              <div className={styles.formActions}>
                {isEditing ? (
                  <Button type="button" variant="primary" onClick={handleUpdateDepartment} disabled={isPageLoading}>
                    更新
                  </Button>
                ) : (
                  <Button type="button" variant="primary" onClick={handleCreateDepartment} disabled={isPageLoading}>
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