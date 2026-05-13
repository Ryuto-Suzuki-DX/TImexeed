"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import { searchUsers } from "@/api/admin/user";
import { getPaidLeaveBalance, searchPaidLeaveUsages } from "@/api/admin/paidLeaveUsage";
import type { UserResponse } from "@/types/admin/user";
import type { PaidLeaveBalanceResponse, PaidLeaveUsageResponse } from "@/types/admin/paidLeaveUsage";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

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

function calculateDaysUntil(value: string | null | undefined) {
  const dateValue = toDateInputValue(value);

  if (!dateValue) {
    return null;
  }

  const today = new Date();
  const deadline = new Date(`${dateValue}T00:00:00`);

  today.setHours(0, 0, 0, 0);
  deadline.setHours(0, 0, 0, 0);

  const diffMs = deadline.getTime() - today.getTime();

  return Math.ceil(diffMs / (1000 * 60 * 60 * 24));
}

function calculateRequiredUseCount(remainingRequiredDays: number) {
  if (remainingRequiredDays <= 0) {
    return 0;
  }

  return Math.ceil(remainingRequiredDays);
}

export default function AdminPaidLeaveCheckPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [keyword, setKeyword] = useState("");
  const [searchedKeyword, setSearchedKeyword] = useState("");
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [userOffset, setUserOffset] = useState(0);
  const [userHasMore, setUserHasMore] = useState(false);
  const [isUserSearching, setIsUserSearching] = useState(false);

  const [selectedUser, setSelectedUser] = useState<UserResponse | null>(null);
  const [paidLeaveUsages, setPaidLeaveUsages] = useState<PaidLeaveUsageResponse[]>([]);
  const [balance, setBalance] = useState<PaidLeaveBalanceResponse | null>(null);

  const [isUsageModalOpen, setIsUsageModalOpen] = useState(false);

  const [pageMessage, setPageMessage] = useState("ユーザーを検索して、有給状況を確認してください。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");
  const [isPageLoading, setIsPageLoading] = useState(false);

  const selectedUserLabel = useMemo(() => {
    if (!selectedUser) {
      return "未選択";
    }

    return `${selectedUser.name}（${selectedUser.email}）`;
  }, [selectedUser]);

  const legalCheck = useMemo(() => {
    if (!balance) {
      return {
        remainingUseDays: 0,
        remainingUseCount: 0,
        daysUntilDeadline: null as number | null,
        message: "対象ユーザーを選択してください。",
        variant: "info" as PageMessageVariant,
      };
    }

    const remainingUseDays = balance.requiredUseRemainingDays;
    const remainingUseCount = calculateRequiredUseCount(remainingUseDays);
    const daysUntilDeadline = calculateDaysUntil(balance.requiredUseDeadline);

    if (!balance.requiredUseDeadline) {
      return {
        remainingUseDays,
        remainingUseCount,
        daysUntilDeadline,
        message: "現時点では年5日取得義務の対象期間がありません。",
        variant: "info" as PageMessageVariant,
      };
    }

    if (remainingUseDays <= 0) {
      return {
        remainingUseDays,
        remainingUseCount,
        daysUntilDeadline,
        message: "年5日取得義務は満たしています。",
        variant: "success" as PageMessageVariant,
      };
    }

    if (daysUntilDeadline !== null && daysUntilDeadline < 0) {
      return {
        remainingUseDays,
        remainingUseCount,
        daysUntilDeadline,
        message: `期限を過ぎています。あと${formatNumber(remainingUseDays)}日分の取得が不足しています。`,
        variant: "error" as PageMessageVariant,
      };
    }

    if (daysUntilDeadline !== null && daysUntilDeadline <= 30) {
      return {
        remainingUseDays,
        remainingUseCount,
        daysUntilDeadline,
        message: `期限が近いです。あと${remainingUseCount}回、合計${formatNumber(remainingUseDays)}日分の取得が必要です。`,
        variant: "warning" as PageMessageVariant,
      };
    }

    return {
      remainingUseDays,
      remainingUseCount,
      daysUntilDeadline,
      message: `あと${remainingUseCount}回、合計${formatNumber(remainingUseDays)}日分の取得が必要です。`,
      variant: "warning" as PageMessageVariant,
    };
  }, [balance]);

  const loadPaidLeaveData = useCallback(async (targetUserId: number) => {
    setIsPageLoading(true);
    setPageMessage("有給情報を取得しています。");
    setPageMessageVariant("info");

    const [usagesResult, balanceResult] = await Promise.all([
      searchPaidLeaveUsages({
        targetUserId,
        includeDeleted: false,
        offset: 0,
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

    setPaidLeaveUsages(usagesData.paidLeaveUsages);
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
      setBalance(null);
      setIsUsageModalOpen(false);
    }

    setPageMessage(searchData.users.length === 0 ? "該当するユーザーが見つかりませんでした。" : "ユーザー検索が完了しました。");
    setPageMessageVariant(searchData.users.length === 0 ? "warning" : "success");
    setIsUserSearching(false);
  };

  const handleSelectUser = async (targetUser: UserResponse) => {
    setSelectedUser(targetUser);
    setPaidLeaveUsages([]);
    setBalance(null);
    setIsUsageModalOpen(false);

    await loadPaidLeaveData(targetUser.id);
  };

  const handleLoadMoreUsers = async () => {
    await handleSearchUsers(userOffset, true);
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
          <PageTitle title="有給確認" description="ログイン情報を確認しています。" />
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
            <PageTitle title="有給確認" description="ユーザーごとの有給残数、使用履歴、年5日取得義務の状況を確認します。" />

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
                  <h2 className={styles.sectionTitle}>有給状況</h2>
                  <p className={styles.sectionDescription}>{selectedUserLabel}</p>
                </div>

                <Button type="button" variant="secondary" onClick={() => setIsUsageModalOpen(true)} disabled={!selectedUser}>
                  使用履歴を見る
                </Button>
              </div>

              {balance ? (
                <>
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

                  <div className={styles.legalCard}>
                    <div className={styles.legalCardHeader}>
                      <div>
                        <h3 className={styles.legalTitle}>年5日取得義務チェック</h3>
                        <p className={styles.legalDescription}>現時点の有給使用状況から、取得義務の残りを確認します。</p>
                      </div>

                      <span className={`${styles.legalBadge} ${styles[`legalBadge_${legalCheck.variant}`]}`}>{legalCheck.variant}</span>
                    </div>

                    <MessageBox variant={legalCheck.variant}>{legalCheck.message}</MessageBox>

                    <div className={styles.legalGrid}>
                      <div className={styles.legalItem}>
                        <p className={styles.legalLabel}>期限</p>
                        <p className={styles.legalValue}>{formatDate(balance.requiredUseDeadline)}</p>
                      </div>

                      <div className={styles.legalItem}>
                        <p className={styles.legalLabel}>期限まで</p>
                        <p className={styles.legalValue}>
                          {legalCheck.daysUntilDeadline === null ? "-" : `${legalCheck.daysUntilDeadline}日`}
                        </p>
                      </div>

                      <div className={styles.legalItem}>
                        <p className={styles.legalLabel}>残り必要日数</p>
                        <p className={styles.legalValue}>{formatNumber(legalCheck.remainingUseDays)}日</p>
                      </div>

                      <div className={styles.legalItemStrong}>
                        <p className={styles.legalLabel}>あと何回</p>
                        <p className={styles.legalValue}>{legalCheck.remainingUseCount}回</p>
                      </div>
                    </div>
                  </div>
                </>
              ) : (
                <div className={styles.emptyDetail}>
                  <p>左側のユーザー検索から対象ユーザーを選択してください。</p>
                </div>
              )}
            </section>
          </div>
        </section>
      </div>

      {isUsageModalOpen && (
        <div className={styles.modalOverlay} role="presentation" onClick={() => setIsUsageModalOpen(false)}>
          <section className={styles.modal} role="dialog" aria-modal="true" aria-labelledby="paid-leave-usage-modal-title" onClick={(event) => event.stopPropagation()}>
            <div className={styles.modalHeader}>
              <div>
                <h2 id="paid-leave-usage-modal-title" className={styles.modalTitle}>
                  有給使用履歴
                </h2>
                <p className={styles.modalDescription}>{selectedUserLabel}</p>
              </div>

              <Button type="button" variant="secondary" onClick={() => setIsUsageModalOpen(false)}>
                閉じる
              </Button>
            </div>

            <div className={styles.tableScroll}>
              <table className={styles.table}>
                <thead>
                  <tr>
                    <th>使用日</th>
                    <th>使用日数</th>
                    <th>区分</th>
                    <th>メモ</th>
                    <th>登録日</th>
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
                      <td>{formatDate(usage.createdAt)}</td>
                    </tr>
                  ))}

                  {paidLeaveUsages.length === 0 && (
                    <tr>
                      <td colSpan={5} className={styles.emptyCell}>
                        有給使用履歴はまだありません。
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </section>
        </div>
      )}
    </PageContainer>
  );
}