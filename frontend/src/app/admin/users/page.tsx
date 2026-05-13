"use client";

import { useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import { createUser, deleteUser, getUserDetail, searchUsers, updateUser } from "@/api/admin/user";
import { searchDepartments } from "@/api/admin/department";
import type { UserResponse } from "@/types/admin/user";
import type { DepartmentResponse } from "@/types/admin/department";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

type UserForm = {
  targetUserId: number | null;
  name: string;
  email: string;
  password: string;
  role: string;
  departmentId: string;
  hireDate: string;
  retirementDate: string;
};

const initialForm: UserForm = {
  targetUserId: null,
  name: "",
  email: "",
  password: "",
  role: "USER",
  departmentId: "",
  hireDate: "",
  retirementDate: "",
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

function parseDepartmentId(value: string) {
  if (!value.trim()) {
    return null;
  }

  const parsed = Number(value);

  if (!Number.isInteger(parsed) || parsed <= 0) {
    return null;
  }

  return parsed;
}

export default function AdminUsersPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [keyword, setKeyword] = useState("");
  const [searchedKeyword, setSearchedKeyword] = useState("");
  const [includeDeleted, setIncludeDeleted] = useState(false);

  const [users, setUsers] = useState<UserResponse[]>([]);
  const [userOffset, setUserOffset] = useState(0);
  const [userHasMore, setUserHasMore] = useState(false);

  const [departments, setDepartments] = useState<DepartmentResponse[]>([]);

  const [form, setForm] = useState<UserForm>(initialForm);
  const [isEditing, setIsEditing] = useState(false);

  const [pageMessage, setPageMessage] = useState("ユーザーを検索・作成・編集できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const [isPageLoading, setIsPageLoading] = useState(false);
  const [isUserSearching, setIsUserSearching] = useState(false);

  const formTitle = useMemo(() => {
    return isEditing ? "ユーザー編集" : "ユーザー新規作成";
  }, [isEditing]);

  const formDescription = useMemo(() => {
    return isEditing
      ? "選択したユーザー情報を更新します。パスワードはこの画面では変更しません。"
      : "新しいユーザーを作成します。初期パスワードは8文字以上で入力してください。";
  }, [isEditing]);

  const currentDepartmentExists = useMemo(() => {
    if (!form.departmentId) {
      return true;
    }

    return departments.some((department) => String(department.id) === form.departmentId);
  }, [departments, form.departmentId]);

  const resetForm = () => {
    setForm(initialForm);
    setIsEditing(false);
  };

  const loadDepartments = async () => {
    const result = await searchDepartments({
      keyword: "",
      includeDeleted: false,
      offset: 0,
      limit: 50,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "所属一覧の取得に失敗しました。");
      setPageMessageVariant("error");
      return;
    }

    const departmentData = result.data;

    setDepartments(departmentData.departments);
  };

  const handleSearchUsers = async (nextOffset: number, append: boolean, includeDeletedValue = includeDeleted) => {
    setIsUserSearching(true);
    setPageMessage("ユーザーを検索しています。");
    setPageMessageVariant("info");

    const searchKeyword = append ? searchedKeyword : keyword;

    const result = await searchUsers({
      keyword: searchKeyword,
      includeDeleted: includeDeletedValue,
      offset: nextOffset,
      limit: 50,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "ユーザー検索に失敗しました。");
      setPageMessageVariant("error");
      setIsUserSearching(false);
      return;
    }

    const searchData = result.data;

    setSearchedKeyword(searchKeyword);
    setUsers((current) => (append ? [...current, ...searchData.users] : searchData.users));
    setUserOffset(nextOffset + searchData.users.length);
    setUserHasMore(searchData.hasMore);

    setPageMessage(searchData.users.length === 0 ? "該当するユーザーが見つかりませんでした。" : "ユーザー検索が完了しました。");
    setPageMessageVariant(searchData.users.length === 0 ? "warning" : "success");
    setIsUserSearching(false);
  };

  const handleToggleIncludeDeleted = async () => {
    const nextIncludeDeleted = !includeDeleted;

    setIncludeDeleted(nextIncludeDeleted);
    resetForm();
    await handleSearchUsers(0, false, nextIncludeDeleted);
  };

  const handleLoadMoreUsers = async () => {
    await handleSearchUsers(userOffset, true);
  };

  const handleStartCreate = () => {
    resetForm();
    setPageMessage("新規ユーザー情報を入力してください。");
    setPageMessageVariant("info");
  };

  const handleStartEdit = async (targetUser: UserResponse) => {
    if (targetUser.isDeleted) {
      setPageMessage("削除済みユーザーは編集できません。");
      setPageMessageVariant("warning");
      return;
    }

    setIsPageLoading(true);
    setPageMessage("ユーザー詳細を取得しています。");
    setPageMessageVariant("info");

    const result = await getUserDetail({
      targetUserId: targetUser.id,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "ユーザー詳細の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    const detailData = result.data;
    const userDetail = detailData.user;

    setForm({
      targetUserId: userDetail.id,
      name: userDetail.name,
      email: userDetail.email,
      password: "",
      role: userDetail.role,
      departmentId: userDetail.departmentId === null ? "" : String(userDetail.departmentId),
      hireDate: toDateInputValue(userDetail.hireDate),
      retirementDate: toDateInputValue(userDetail.retirementDate),
    });

    setIsEditing(true);
    setPageMessage("ユーザー情報を編集できます。");
    setPageMessageVariant("info");
    setIsPageLoading(false);
  };

  const validateForm = () => {
    if (!form.name.trim()) {
      setPageMessage("ユーザー名を入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    if (!form.email.trim()) {
      setPageMessage("メールアドレスを入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    if (!isEditing && form.password.length < 8) {
      setPageMessage("初期パスワードは8文字以上で入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    if (form.role !== "USER" && form.role !== "ADMIN") {
      setPageMessage("権限を選択してください。");
      setPageMessageVariant("error");
      return false;
    }

    if (!form.hireDate) {
      setPageMessage("入社日を入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    return true;
  };

  const handleCreateUser = async () => {
    if (!validateForm()) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("ユーザーを作成しています。");
    setPageMessageVariant("info");

    const result = await createUser({
      name: form.name,
      email: form.email,
      password: form.password,
      role: form.role,
      departmentId: parseDepartmentId(form.departmentId),
      hireDate: form.hireDate,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "ユーザーの作成に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetForm();
    await handleSearchUsers(0, false);

    setPageMessage(result.message || "ユーザーを作成しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  const handleUpdateUser = async () => {
    if (!validateForm()) {
      return;
    }

    if (form.targetUserId === null) {
      setPageMessage("更新対象のユーザーが選択されていません。");
      setPageMessageVariant("error");
      return;
    }

    setIsPageLoading(true);
    setPageMessage("ユーザーを更新しています。");
    setPageMessageVariant("info");

    const result = await updateUser({
      targetUserId: form.targetUserId,
      name: form.name,
      email: form.email,
      role: form.role,
      departmentId: parseDepartmentId(form.departmentId),
      hireDate: form.hireDate,
      retirementDate: form.retirementDate || null,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "ユーザーの更新に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetForm();
    await handleSearchUsers(0, false);

    setPageMessage(result.message || "ユーザーを更新しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  const handleDeleteUser = async (targetUser: UserResponse) => {
    if (targetUser.isDeleted) {
      setPageMessage("このユーザーはすでに削除済みです。");
      setPageMessageVariant("warning");
      return;
    }

    const confirmed = window.confirm(`${targetUser.name} を削除しますか？`);

    if (!confirmed) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("ユーザーを削除しています。");
    setPageMessageVariant("info");

    const result = await deleteUser({
      targetUserId: targetUser.id,
    });

    if (result.error) {
      setPageMessage(result.message || "ユーザーの削除に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    if (form.targetUserId === targetUser.id) {
      resetForm();
    }

    await handleSearchUsers(0, false);

    setPageMessage(result.message || "ユーザーを削除しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void loadDepartments();
      void handleSearchUsers(0, false);
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
          <PageTitle title="ユーザー管理" description="ログイン情報を確認しています。" />
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
            <PageTitle title="ユーザー管理" description="管理者・一般ユーザーの検索、作成、編集、削除を行います。" />

            <MessageBox variant={pageMessageVariant}>{isPageLoading ? "処理中..." : pageMessage}</MessageBox>
          </div>

          <div className={styles.contentGrid}>
            <section className={styles.searchPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>ユーザー検索</h2>
                  <p className={styles.sectionDescription}>名前・メールアドレス・権限で検索できます。</p>
                </div>

                <Button type="button" variant="secondary" onClick={handleStartCreate}>
                  新規作成
                </Button>
              </div>

              <div className={styles.searchForm}>
                <input
                  className={styles.searchInput}
                  value={keyword}
                  placeholder="名前・メールアドレス・権限で検索"
                  onChange={(event) => setKeyword(event.target.value)}
                />

                <Button type="button" variant="primary" onClick={() => handleSearchUsers(0, false)} disabled={isUserSearching}>
                  検索
                </Button>
              </div>

              <button type="button" className={styles.switchRow} onClick={handleToggleIncludeDeleted}>
                <span className={`${styles.switch} ${includeDeleted ? styles.switchOn : ""}`}>
                  <span className={styles.switchThumb} />
                </span>

                <span className={styles.switchText}>削除済みユーザーも含める</span>
              </button>

              <div className={styles.userList}>
                {users.map((targetUser) => (
                  <article key={targetUser.id} className={`${styles.userRow} ${targetUser.isDeleted ? styles.userRowDeleted : ""}`}>
                    <div className={styles.userRowMain}>
                      <div>
                        <div className={styles.userNameLine}>
                          <p className={styles.userName}>{targetUser.name}</p>

                          <span className={targetUser.role === "ADMIN" ? styles.adminBadge : styles.userBadge}>{targetUser.role}</span>

                          {targetUser.isDeleted && <span className={styles.deletedBadge}>削除済み</span>}
                        </div>

                        <p className={styles.userEmail}>{targetUser.email}</p>
                        <p className={styles.userMeta}>
                          所属ID：{targetUser.departmentId ?? "-"} / 入社日：{formatDate(targetUser.hireDate)} / 退職日：
                          {formatDate(targetUser.retirementDate)}
                        </p>
                      </div>

                      <div className={styles.rowActions}>
                        <Button type="button" variant="secondary" onClick={() => handleStartEdit(targetUser)} disabled={targetUser.isDeleted}>
                          編集
                        </Button>

                        <Button type="button" variant="danger" onClick={() => handleDeleteUser(targetUser)} disabled={targetUser.isDeleted}>
                          削除
                        </Button>
                      </div>
                    </div>
                  </article>
                ))}

                {users.length === 0 && <p className={styles.emptyText}>ユーザーが見つかりません。</p>}
              </div>

              {userHasMore && (
                <div className={styles.moreButtonArea}>
                  <Button type="button" variant="secondary" onClick={handleLoadMoreUsers} disabled={isUserSearching}>
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
                  <span className={styles.fieldLabel}>ユーザー名</span>
                  <input
                    className={styles.input}
                    value={form.name}
                    placeholder="山田太郎"
                    onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))}
                  />
                </label>

                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>メールアドレス</span>
                  <input
                    className={styles.input}
                    type="email"
                    value={form.email}
                    placeholder="user@example.com"
                    onChange={(event) => setForm((current) => ({ ...current, email: event.target.value }))}
                  />
                </label>

                {!isEditing && (
                  <label className={styles.fieldWide}>
                    <span className={styles.fieldLabel}>初期パスワード</span>
                    <input
                      className={styles.input}
                      type="password"
                      value={form.password}
                      placeholder="8文字以上"
                      onChange={(event) => setForm((current) => ({ ...current, password: event.target.value }))}
                    />
                  </label>
                )}

                <label className={styles.field}>
                  <span className={styles.fieldLabel}>権限</span>
                  <select
                    className={styles.input}
                    value={form.role}
                    onChange={(event) => setForm((current) => ({ ...current, role: event.target.value }))}
                  >
                    <option value="USER">USER</option>
                    <option value="ADMIN">ADMIN</option>
                  </select>
                </label>

                <label className={styles.field}>
                  <span className={styles.fieldLabel}>所属</span>
                  <select
                    className={styles.input}
                    value={form.departmentId}
                    onChange={(event) => setForm((current) => ({ ...current, departmentId: event.target.value }))}
                  >
                    <option value="">未所属</option>

                    {!currentDepartmentExists && form.departmentId && (
                      <option value={form.departmentId}>現在の所属ID：{form.departmentId}</option>
                    )}

                    {departments.map((department) => (
                      <option key={department.id} value={department.id}>
                        {department.name}
                      </option>
                    ))}
                  </select>
                </label>

                <label className={styles.field}>
                  <span className={styles.fieldLabel}>入社日</span>
                  <input
                    className={styles.input}
                    type="date"
                    value={form.hireDate}
                    onChange={(event) => setForm((current) => ({ ...current, hireDate: event.target.value }))}
                  />
                </label>

                {isEditing && (
                  <label className={styles.field}>
                    <span className={styles.fieldLabel}>退職日</span>
                    <input
                      className={styles.input}
                      type="date"
                      value={form.retirementDate}
                      onChange={(event) => setForm((current) => ({ ...current, retirementDate: event.target.value }))}
                    />
                  </label>
                )}
              </div>

              <div className={styles.formActions}>
                {isEditing ? (
                  <Button type="button" variant="primary" onClick={handleUpdateUser} disabled={isPageLoading}>
                    更新
                  </Button>
                ) : (
                  <Button type="button" variant="primary" onClick={handleCreateUser} disabled={isPageLoading}>
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