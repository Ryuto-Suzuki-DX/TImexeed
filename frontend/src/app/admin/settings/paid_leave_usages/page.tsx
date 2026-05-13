"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import { searchUsers } from "@/api/admin/user";
import {
  createPaidLeaveUsage,
  deletePaidLeaveUsage,
  getPaidLeaveBalance,
  searchPaidLeaveUsages,
  updatePaidLeaveUsage,
} from "@/api/admin/paidLeaveUsage";
import type { UserResponse } from "@/types/admin/user";
import type { PaidLeaveBalanceResponse, PaidLeaveUsageResponse } from "@/types/admin/paidLeaveUsage";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

type PaidLeaveUsageForm = {
  usageDate: string;
  usageDays: string;
  memo: string;
};

const initialForm: PaidLeaveUsageForm = {
  usageDate: "",
  usageDays: "1",
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

function formatNumber(value: number | null | undefined) {
  if (value === null || value === undefined) {
    return "-";
  }

  return value.toFixed(1).replace(".0", "");
}

export default function AdminPaidLeaveUsagesPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [keyword, setKeyword] = useState("");
  const [searchedKeyword, setSearchedKeyword] = useState("");
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [userOffset, setUserOffset] = useState(0);
  const [userHasMore, setUserHasMore] = useState(false);
  const [isUserSearching, setIsUserSearching] = useState(false);

  const [selectedUser, setSelectedUser] = useState<UserResponse | null>(null);
  const [paidLeaveUsages, setPaidLeaveUsages] = useState<PaidLeaveUsageResponse[]>([]);
  const [usageOffset, setUsageOffset] = useState(0);
  const [usageHasMore, setUsageHasMore] = useState(false);
  const [balance, setBalance] = useState<PaidLeaveBalanceResponse | null>(null);

  const [form, setForm] = useState<PaidLeaveUsageForm>(initialForm);
  const [editingUsageId, setEditingUsageId] = useState<number | null>(null);

  const [pageMessage, setPageMessage] = useState("ユーザーを検索して、有給過去使用分を管理してください。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");
  const [isPageLoading, setIsPageLoading] = useState(false);

  const selectedUserLabel = useMemo(() => {
    if (!selectedUser) {
      return "未選択";
    }

    return `${selectedUser.name}（${selectedUser.email}）`;
  }, [selectedUser]);

  const resetForm = () => {
    setForm(initialForm);
    setEditingUsageId(null);
  };

  const loadPaidLeaveData = useCallback(async (targetUserId: number, nextOffset: number, append: boolean) => {
    setIsPageLoading(true);
    setPageMessage("有給情報を取得しています。");
    setPageMessageVariant("info");

    const [usagesResult, balanceResult] = await Promise.all([
      searchPaidLeaveUsages({
        targetUserId,
        includeDeleted: false,
        offset: nextOffset,
        limit: 50,
      }),
      getPaidLeaveBalance({
        targetUserId,
      }),
    ]);

    if (usagesResult.error || !usagesResult.data) {
      setPageMessage(usagesResult.message || "有給使用日一覧の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    if (balanceResult.error || !balanceResult.data) {
      setPageMessage(balanceResult.message || "有給残数の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    const usagesData = usagesResult.data;
    const balanceData = balanceResult.data;

    setPaidLeaveUsages((current) =>
      append ? [...current, ...usagesData.paidLeaveUsages] : usagesData.paidLeaveUsages,
    );
    setUsageOffset(nextOffset + usagesData.paidLeaveUsages.length);
    setUsageHasMore(usagesData.hasMore);
    setBalance(balanceData);

    setPageMessage("有給情報を取得しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  }, []);

  const handleSearchUsers = async (nextOffset: number, append: boolean) => {
    setIsUserSearching(true);
    setPageMessage("ユーザーを検索しています。");
    setPageMessageVariant("info");

    const searchKeyword = append ? searchedKeyword : keyword;

    const result = await searchUsers({
      keyword: searchKeyword,
      includeDeleted: false,
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

    if (!append) {
      setSelectedUser(null);
      setPaidLeaveUsages([]);
      setUsageOffset(0);
      setUsageHasMore(false);
      setBalance(null);
      resetForm();
    }

    setPageMessage(searchData.users.length === 0 ? "該当するユーザーが見つかりませんでした。" : "ユーザー検索が完了しました。");
    setPageMessageVariant(searchData.users.length === 0 ? "warning" : "success");
    setIsUserSearching(false);
  };

  const handleSelectUser = async (targetUser: UserResponse) => {
    setSelectedUser(targetUser);
    setPaidLeaveUsages([]);
    setUsageOffset(0);
    setUsageHasMore(false);
    setBalance(null);
    resetForm();

    await loadPaidLeaveData(targetUser.id, 0, false);
  };

  const handleLoadMoreUsers = async () => {
    await handleSearchUsers(userOffset, true);
  };

  const handleLoadMoreUsages = async () => {
    if (!selectedUser) {
      return;
    }

    await loadPaidLeaveData(selectedUser.id, usageOffset, true);
  };

  const handleCreatePaidLeaveUsage = async () => {
    if (!selectedUser) {
      setPageMessage("先に対象ユーザーを選択してください。");
      setPageMessageVariant("warning");
      return;
    }

    if (!form.usageDate) {
      setPageMessage("有給使用日を入力してください。");
      setPageMessageVariant("error");
      return;
    }

    const usageDays = Number(form.usageDays);

    if (usageDays !== 1 && usageDays !== 0.5) {
      setPageMessage("使用日数は 1 または 0.5 を選択してください。");
      setPageMessageVariant("error");
      return;
    }

    setIsPageLoading(true);
    setPageMessage("有給使用日を追加しています。");
    setPageMessageVariant("info");

    const result = await createPaidLeaveUsage({
      targetUserId: selectedUser.id,
      usageDate: form.usageDate,
      usageDays,
      memo: form.memo,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "有給使用日の追加に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetForm();
    await loadPaidLeaveData(selectedUser.id, 0, false);

    setPageMessage(result.message || "有給使用日を追加しました。");
    setPageMessageVariant("success");
  };

  const handleStartEdit = (usage: PaidLeaveUsageResponse) => {
    if (!usage.isManual) {
      setPageMessage("手動追加ではない有給使用日は、この画面から編集できません。");
      setPageMessageVariant("warning");
      return;
    }

    setEditingUsageId(usage.id);
    setForm({
      usageDate: toDateInputValue(usage.usageDate),
      usageDays: String(usage.usageDays),
      memo: usage.memo,
    });

    setPageMessage("編集内容を入力して、更新ボタンを押してください。");
    setPageMessageVariant("info");
  };

  const handleUpdatePaidLeaveUsage = async () => {
    if (!selectedUser || editingUsageId === null) {
      setPageMessage("編集対象の有給使用日を選択してください。");
      setPageMessageVariant("warning");
      return;
    }

    if (!form.usageDate) {
      setPageMessage("有給使用日を入力してください。");
      setPageMessageVariant("error");
      return;
    }

    const usageDays = Number(form.usageDays);

    if (usageDays !== 1 && usageDays !== 0.5) {
      setPageMessage("使用日数は 1 または 0.5 を選択してください。");
      setPageMessageVariant("error");
      return;
    }

    setIsPageLoading(true);
    setPageMessage("有給使用日を更新しています。");
    setPageMessageVariant("info");

    const result = await updatePaidLeaveUsage({
      targetUserId: selectedUser.id,
      targetPaidLeaveUsageId: editingUsageId,
      usageDate: form.usageDate,
      usageDays,
      memo: form.memo,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "有給使用日の更新に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetForm();
    await loadPaidLeaveData(selectedUser.id, 0, false);

    setPageMessage(result.message || "有給使用日を更新しました。");
    setPageMessageVariant("success");
  };

  const handleDeletePaidLeaveUsage = async (usage: PaidLeaveUsageResponse) => {
    if (!selectedUser) {
      return;
    }

    if (!usage.isManual) {
      setPageMessage("手動追加ではない有給使用日は、この画面から削除できません。");
      setPageMessageVariant("warning");
      return;
    }

    const confirmed = window.confirm(`${formatDate(usage.usageDate)} の有給使用日を削除しますか？`);

    if (!confirmed) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("有給使用日を削除しています。");
    setPageMessageVariant("info");

    const result = await deletePaidLeaveUsage({
      targetUserId: selectedUser.id,
      targetPaidLeaveUsageId: usage.id,
    });

    if (result.error) {
      setPageMessage(result.message || "有給使用日の削除に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    if (editingUsageId === usage.id) {
      resetForm();
    }

    await loadPaidLeaveData(selectedUser.id, 0, false);

    setPageMessage(result.message || "有給使用日を削除しました。");
    setPageMessageVariant("success");
  };

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
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
          <PageTitle title="有給過去使用分管理" description="ログイン情報を確認しています。" />
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
            <PageTitle title="有給過去使用分管理" description="システム導入前の有給使用日を管理します。" />

            <MessageBox variant={pageMessageVariant}>{isPageLoading ? "処理中..." : pageMessage}</MessageBox>
          </div>

          <div className={styles.contentGrid}>
            <section className={styles.searchPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>ユーザー検索</h2>
                  <p className={styles.sectionDescription}>名前・メールアドレスなどで対象ユーザーを検索します。</p>
                </div>
              </div>

              <div className={styles.searchForm}>
                <input
                  className={styles.searchInput}
                  value={keyword}
                  placeholder="名前・メールアドレスで検索"
                  onChange={(event) => setKeyword(event.target.value)}
                />

                <Button type="button" variant="primary" onClick={() => handleSearchUsers(0, false)} disabled={isUserSearching}>
                  検索
                </Button>
              </div>

              <div className={styles.userList}>
                {users.map((targetUser) => (
                  <button
                    key={targetUser.id}
                    type="button"
                    className={`${styles.userRow} ${selectedUser?.id === targetUser.id ? styles.userRowSelected : ""}`}
                    onClick={() => handleSelectUser(targetUser)}
                  >
                    <span className={styles.userName}>{targetUser.name}</span>
                    <span className={styles.userEmail}>{targetUser.email}</span>
                    <span className={styles.userMeta}>入社日：{formatDate(targetUser.hireDate)}</span>
                  </button>
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

            <section className={styles.detailPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>対象ユーザー</h2>
                  <p className={styles.sectionDescription}>{selectedUserLabel}</p>
                </div>
              </div>

              {balance && (
                <div className={styles.balanceGrid}>
                  <div className={styles.balanceCard}>
                    <p className={styles.balanceLabel}>付与合計</p>
                    <p className={styles.balanceValue}>{formatNumber(balance.totalGrantedDays)}日</p>
                  </div>

                  <div className={styles.balanceCard}>
                    <p className={styles.balanceLabel}>使用済み</p>
                    <p className={styles.balanceValue}>{formatNumber(balance.usedDays)}日</p>
                  </div>

                  <div className={styles.balanceCardStrong}>
                    <p className={styles.balanceLabel}>残数</p>
                    <p className={styles.balanceValue}>{formatNumber(balance.remainingDays)}日</p>
                  </div>

                  <div className={styles.balanceCard}>
                    <p className={styles.balanceLabel}>次回付与</p>
                    <p className={styles.balanceSubValue}>
                      {balance.nextGrantDate ? `${formatDate(balance.nextGrantDate)} に ${formatNumber(balance.nextGrantDays)}日` : "-"}
                    </p>
                  </div>
                </div>
              )}

              <div className={styles.formCard}>
                <div className={styles.formHeader}>
                  <div>
                    <h3 className={styles.formTitle}>{editingUsageId === null ? "過去有給使用日追加" : "過去有給使用日編集"}</h3>
                    <p className={styles.formDescription}>
                      {editingUsageId === null
                        ? "管理者が追加するため、バックエンド側で手動追加として保存されます。"
                        : "手動追加された有給使用日のみ編集できます。"}
                    </p>
                  </div>

                  {editingUsageId !== null && (
                    <Button type="button" variant="secondary" onClick={resetForm}>
                      編集取消
                    </Button>
                  )}
                </div>

                <div className={styles.formGrid}>
                  <label className={styles.field}>
                    <span className={styles.fieldLabel}>使用日</span>
                    <input
                      className={styles.input}
                      type="date"
                      value={form.usageDate}
                      onChange={(event) => setForm((current) => ({ ...current, usageDate: event.target.value }))}
                      disabled={!selectedUser}
                    />
                  </label>

                  <label className={styles.field}>
                    <span className={styles.fieldLabel}>使用日数</span>
                    <select
                      className={styles.input}
                      value={form.usageDays}
                      onChange={(event) => setForm((current) => ({ ...current, usageDays: event.target.value }))}
                      disabled={!selectedUser}
                    >
                      <option value="1">1日</option>
                      <option value="0.5">0.5日</option>
                    </select>
                  </label>

                  <label className={styles.fieldWide}>
                    <span className={styles.fieldLabel}>メモ</span>
                    <input
                      className={styles.input}
                      value={form.memo}
                      placeholder="例：システム導入前使用分"
                      onChange={(event) => setForm((current) => ({ ...current, memo: event.target.value }))}
                      disabled={!selectedUser}
                    />
                  </label>
                </div>

                <div className={styles.formActions}>
                  {editingUsageId === null ? (
                    <Button type="button" variant="primary" onClick={handleCreatePaidLeaveUsage} disabled={!selectedUser || isPageLoading}>
                      追加
                    </Button>
                  ) : (
                    <Button type="button" variant="primary" onClick={handleUpdatePaidLeaveUsage} disabled={!selectedUser || isPageLoading}>
                      更新
                    </Button>
                  )}
                </div>
              </div>

              <div className={styles.tableArea}>
                <div className={styles.tableHeader}>
                  <div>
                    <h3 className={styles.tableTitle}>有給使用日一覧</h3>
                    <p className={styles.tableDescription}>手動追加分のみ、この画面から編集・削除できます。</p>
                  </div>
                </div>

                <div className={styles.tableScroll}>
                  <table className={styles.table}>
                    <thead>
                      <tr>
                        <th>使用日</th>
                        <th>使用日数</th>
                        <th>区分</th>
                        <th>メモ</th>
                        <th>操作</th>
                      </tr>
                    </thead>

                    <tbody>
                      {paidLeaveUsages.map((usage) => (
                        <tr key={usage.id}>
                          <td>{formatDate(usage.usageDate)}</td>
                          <td>{formatNumber(usage.usageDays)}日</td>
                          <td>
                            <span className={usage.isManual ? styles.manualBadge : styles.systemBadge}>
                              {usage.isManual ? "手動追加" : "システム連携"}
                            </span>
                          </td>
                          <td>{usage.memo || "-"}</td>
                          <td>
                            <div className={styles.rowActions}>
                              <Button type="button" variant="secondary" onClick={() => handleStartEdit(usage)} disabled={!usage.isManual}>
                                編集
                              </Button>

                              <Button type="button" variant="danger" onClick={() => handleDeletePaidLeaveUsage(usage)} disabled={!usage.isManual}>
                                削除
                              </Button>
                            </div>
                          </td>
                        </tr>
                      ))}

                      {paidLeaveUsages.length === 0 && (
                        <tr>
                          <td colSpan={5} className={styles.emptyCell}>
                            {selectedUser ? "有給使用日はまだ登録されていません。" : "ユーザーを選択してください。"}
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>

                {usageHasMore && (
                  <div className={styles.moreButtonArea}>
                    <Button type="button" variant="secondary" onClick={handleLoadMoreUsages} disabled={isPageLoading}>
                      有給使用日をさらに表示
                    </Button>
                  </div>
                )}
              </div>
            </section>
          </div>
        </section>
      </div>
    </PageContainer>
  );
}