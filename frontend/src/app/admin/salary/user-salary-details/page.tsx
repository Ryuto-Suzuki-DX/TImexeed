"use client";

import { useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import { searchUsers } from "@/api/admin/user";
import {
  createUserSalaryDetail,
  deleteUserSalaryDetail,
  getUserSalaryDetail,
  searchUserSalaryDetails,
  updateUserSalaryDetail,
} from "@/api/admin/userSalaryDetail";
import type { UserResponse } from "@/types/admin/user";
import type { SalaryType, UserSalaryDetailResponse } from "@/types/admin/userSalaryDetail";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

type UserSalaryDetailForm = {
  userSalaryDetailId: number | null;
  salaryType: SalaryType;
  baseAmount: string;
  extraAllowanceAmount: string;
  extraAllowanceMemo: string;
  fixedDeductionAmount: string;
  fixedDeductionMemo: string;
  isPayrollTarget: boolean;
  effectiveFrom: string;
  effectiveTo: string;
  memo: string;
};

const initialForm: UserSalaryDetailForm = {
  userSalaryDetailId: null,
  salaryType: "MONTHLY",
  baseAmount: "0",
  extraAllowanceAmount: "0",
  extraAllowanceMemo: "",
  fixedDeductionAmount: "0",
  fixedDeductionMemo: "",
  isPayrollTarget: true,
  effectiveFrom: "",
  effectiveTo: "",
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

function formatAmount(value: number) {
  return `${value.toLocaleString()}円`;
}

function getSalaryTypeLabel(value: SalaryType) {
  switch (value) {
    case "MONTHLY":
      return "月給";
    case "HOURLY":
      return "時給";
    case "DAILY":
      return "日給";
    default:
      return value;
  }
}

function toNumberValue(value: string) {
  const normalizedValue = value.trim();

  if (!normalizedValue) {
    return 0;
  }

  return Number(normalizedValue);
}

function toForm(detail: UserSalaryDetailResponse): UserSalaryDetailForm {
  return {
    userSalaryDetailId: detail.id,
    salaryType: detail.salaryType,
    baseAmount: String(detail.baseAmount),
    extraAllowanceAmount: String(detail.extraAllowanceAmount),
    extraAllowanceMemo: detail.extraAllowanceMemo,
    fixedDeductionAmount: String(detail.fixedDeductionAmount),
    fixedDeductionMemo: detail.fixedDeductionMemo,
    isPayrollTarget: detail.isPayrollTarget,
    effectiveFrom: toDateInputValue(detail.effectiveFrom),
    effectiveTo: toDateInputValue(detail.effectiveTo),
    memo: detail.memo,
  };
}

export default function AdminUserSalaryDetailsPage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [keyword, setKeyword] = useState("");
  const [searchedKeyword, setSearchedKeyword] = useState("");

  const [users, setUsers] = useState<UserResponse[]>([]);
  const [userOffset, setUserOffset] = useState(0);
  const [userHasMore, setUserHasMore] = useState(false);
  const [selectedUser, setSelectedUser] = useState<UserResponse | null>(null);

  const [includeDeletedSalaryDetails, setIncludeDeletedSalaryDetails] = useState(false);
  const [salaryDetails, setSalaryDetails] = useState<UserSalaryDetailResponse[]>([]);
  const [salaryDetailOffset, setSalaryDetailOffset] = useState(0);
  const [salaryDetailHasMore, setSalaryDetailHasMore] = useState(false);

  const [form, setForm] = useState<UserSalaryDetailForm>(initialForm);
  const [isEditing, setIsEditing] = useState(false);

  const [pageMessage, setPageMessage] = useState("給与詳細を管理するユーザーを検索してください。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");

  const [isPageLoading, setIsPageLoading] = useState(false);
  const [isUserSearching, setIsUserSearching] = useState(false);
  const [isSalaryDetailSearching, setIsSalaryDetailSearching] = useState(false);

  const formTitle = useMemo(() => {
    return isEditing ? "ユーザー給与詳細編集" : "ユーザー給与詳細新規作成";
  }, [isEditing]);

  const formDescription = useMemo(() => {
    if (!selectedUser) {
      return "先に左側で対象ユーザーを選択してください。";
    }

    return isEditing ? "選択した給与詳細を更新します。" : `${selectedUser.name} さんの給与詳細を作成します。`;
  }, [isEditing, selectedUser]);

  const resetForm = () => {
    setForm(initialForm);
    setIsEditing(false);
  };

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

    setPageMessage(searchData.users.length === 0 ? "該当するユーザーが見つかりませんでした。" : "ユーザー検索が完了しました。");
    setPageMessageVariant(searchData.users.length === 0 ? "warning" : "success");
    setIsUserSearching(false);
  };

  const handleLoadMoreUsers = async () => {
    await handleSearchUsers(userOffset, true);
  };

  const handleSearchSalaryDetails = async (
    targetUserId: number,
    nextOffset: number,
    append: boolean,
    includeDeletedValue = includeDeletedSalaryDetails
  ) => {
    setIsSalaryDetailSearching(true);
    setPageMessage("給与詳細を検索しています。");
    setPageMessageVariant("info");

    const result = await searchUserSalaryDetails({
      targetUserId,
      includeDeleted: includeDeletedValue,
      offset: nextOffset,
      limit: 50,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "給与詳細検索に失敗しました。");
      setPageMessageVariant("error");
      setIsSalaryDetailSearching(false);
      return;
    }

    const searchData = result.data;

    setSalaryDetails((current) => (append ? [...current, ...searchData.userSalaryDetails] : searchData.userSalaryDetails));
    setSalaryDetailOffset(nextOffset + searchData.userSalaryDetails.length);
    setSalaryDetailHasMore(searchData.hasMore);

    setPageMessage(searchData.userSalaryDetails.length === 0 ? "給与詳細が登録されていません。" : "給与詳細検索が完了しました。");
    setPageMessageVariant(searchData.userSalaryDetails.length === 0 ? "warning" : "success");
    setIsSalaryDetailSearching(false);
  };

  const handleSelectUser = async (targetUser: UserResponse) => {
    if (targetUser.isDeleted) {
      setPageMessage("削除済みユーザーは選択できません。");
      setPageMessageVariant("warning");
      return;
    }

    setSelectedUser(targetUser);
    resetForm();
    setSalaryDetails([]);
    setSalaryDetailOffset(0);
    setSalaryDetailHasMore(false);
    await handleSearchSalaryDetails(targetUser.id, 0, false);
  };

  const handleToggleIncludeDeletedSalaryDetails = async () => {
    if (!selectedUser) {
      setPageMessage("先に対象ユーザーを選択してください。");
      setPageMessageVariant("warning");
      return;
    }

    const nextIncludeDeleted = !includeDeletedSalaryDetails;

    setIncludeDeletedSalaryDetails(nextIncludeDeleted);
    resetForm();
    await handleSearchSalaryDetails(selectedUser.id, 0, false, nextIncludeDeleted);
  };

  const handleLoadMoreSalaryDetails = async () => {
    if (!selectedUser) {
      return;
    }

    await handleSearchSalaryDetails(selectedUser.id, salaryDetailOffset, true);
  };

  const handleStartCreate = () => {
    if (!selectedUser) {
      setPageMessage("先に対象ユーザーを選択してください。");
      setPageMessageVariant("warning");
      return;
    }

    resetForm();
    setPageMessage("新規給与詳細を入力してください。");
    setPageMessageVariant("info");
  };

  const handleStartEdit = async (salaryDetail: UserSalaryDetailResponse) => {
    if (salaryDetail.isDeleted) {
      setPageMessage("削除済み給与詳細は編集できません。");
      setPageMessageVariant("warning");
      return;
    }

    setIsPageLoading(true);
    setPageMessage("給与詳細を取得しています。");
    setPageMessageVariant("info");

    const result = await getUserSalaryDetail({
      userSalaryDetailId: salaryDetail.id,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "給与詳細の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    setForm(toForm(result.data.userSalaryDetail));
    setIsEditing(true);
    setPageMessage("給与詳細を編集できます。");
    setPageMessageVariant("info");
    setIsPageLoading(false);
  };

  const validateForm = () => {
    if (!selectedUser) {
      setPageMessage("対象ユーザーが選択されていません。");
      setPageMessageVariant("error");
      return false;
    }

    if (!form.effectiveFrom) {
      setPageMessage("適用開始日を入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    if (form.effectiveTo && form.effectiveFrom > form.effectiveTo) {
      setPageMessage("適用終了日は適用開始日以降の日付にしてください。");
      setPageMessageVariant("error");
      return false;
    }

    const amountValues = [form.baseAmount, form.extraAllowanceAmount, form.fixedDeductionAmount];
    const hasInvalidAmount = amountValues.some((value) => Number.isNaN(toNumberValue(value)) || toNumberValue(value) < 0);

    if (hasInvalidAmount) {
      setPageMessage("金額には0以上の数値を入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    return true;
  };

  const handleCreateUserSalaryDetail = async () => {
    if (!validateForm() || !selectedUser) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("給与詳細を作成しています。");
    setPageMessageVariant("info");

    const result = await createUserSalaryDetail({
      targetUserId: selectedUser.id,
      salaryType: form.salaryType,
      baseAmount: toNumberValue(form.baseAmount),
      extraAllowanceAmount: toNumberValue(form.extraAllowanceAmount),
      extraAllowanceMemo: form.extraAllowanceMemo,
      fixedDeductionAmount: toNumberValue(form.fixedDeductionAmount),
      fixedDeductionMemo: form.fixedDeductionMemo,
      isPayrollTarget: form.isPayrollTarget,
      effectiveFrom: form.effectiveFrom,
      effectiveTo: form.effectiveTo || null,
      memo: form.memo,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "給与詳細の作成に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetForm();
    await handleSearchSalaryDetails(selectedUser.id, 0, false);

    setPageMessage(result.message || "給与詳細を作成しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  const handleUpdateUserSalaryDetail = async () => {
    if (!validateForm() || !selectedUser) {
      return;
    }

    if (form.userSalaryDetailId === null) {
      setPageMessage("更新対象の給与詳細が選択されていません。");
      setPageMessageVariant("error");
      return;
    }

    setIsPageLoading(true);
    setPageMessage("給与詳細を更新しています。");
    setPageMessageVariant("info");

    const result = await updateUserSalaryDetail({
      userSalaryDetailId: form.userSalaryDetailId,
      salaryType: form.salaryType,
      baseAmount: toNumberValue(form.baseAmount),
      extraAllowanceAmount: toNumberValue(form.extraAllowanceAmount),
      extraAllowanceMemo: form.extraAllowanceMemo,
      fixedDeductionAmount: toNumberValue(form.fixedDeductionAmount),
      fixedDeductionMemo: form.fixedDeductionMemo,
      isPayrollTarget: form.isPayrollTarget,
      effectiveFrom: form.effectiveFrom,
      effectiveTo: form.effectiveTo || null,
      memo: form.memo,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "給与詳細の更新に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetForm();
    await handleSearchSalaryDetails(selectedUser.id, 0, false);

    setPageMessage(result.message || "給与詳細を更新しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  const handleDeleteUserSalaryDetail = async (salaryDetail: UserSalaryDetailResponse) => {
    if (!selectedUser) {
      return;
    }

    if (salaryDetail.isDeleted) {
      setPageMessage("この給与詳細はすでに削除済みです。");
      setPageMessageVariant("warning");
      return;
    }

    const confirmed = window.confirm(`${getSalaryTypeLabel(salaryDetail.salaryType)} ${formatAmount(salaryDetail.baseAmount)} の給与詳細を削除しますか？`);

    if (!confirmed) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("給与詳細を削除しています。");
    setPageMessageVariant("info");

    const result = await deleteUserSalaryDetail({
      userSalaryDetailId: salaryDetail.id,
    });

    if (result.error) {
      setPageMessage(result.message || "給与詳細の削除に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    if (form.userSalaryDetailId === salaryDetail.id) {
      resetForm();
    }

    await handleSearchSalaryDetails(selectedUser.id, 0, false);

    setPageMessage(result.message || "給与詳細を削除しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
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
          <PageTitle title="ユーザー給与詳細" description="ログイン情報を確認しています。" />
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
            <PageTitle title="ユーザー給与詳細" description="ユーザーごとの給与区分、基本金額、固定手当、固定控除を管理します。" />

            <MessageBox variant={pageMessageVariant}>{isPageLoading ? "処理中..." : pageMessage}</MessageBox>
          </div>

          <div className={styles.contentGrid}>
            <section className={styles.searchPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>ユーザー検索</h2>
                  <p className={styles.sectionDescription}>給与詳細を管理するユーザーを選択します。</p>
                </div>
              </div>

              <div className={styles.searchForm}>
                <input
                  className={styles.searchInput}
                  value={keyword}
                  placeholder="ユーザー名・メールで検索"
                  onChange={(event) => setKeyword(event.target.value)}
                />

                <Button type="button" variant="primary" onClick={() => handleSearchUsers(0, false)} disabled={isUserSearching}>
                  検索
                </Button>
              </div>

              <div className={styles.userList}>
                {users.map((targetUser) => (
                  <article
                    key={targetUser.id}
                    className={`${styles.userRow} ${selectedUser?.id === targetUser.id ? styles.userRowSelected : ""}`}
                  >
                    <div className={styles.userRowMain}>
                      <div>
                        <p className={styles.userName}>{targetUser.name}</p>
                        <p className={styles.userMeta}>ID：{targetUser.id} / メール：{targetUser.email}</p>
                      </div>

                      <Button type="button" variant="secondary" onClick={() => handleSelectUser(targetUser)}>
                        選択
                      </Button>
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

              <div className={styles.salarySectionDivider} />

              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>給与詳細履歴</h2>
                  <p className={styles.sectionDescription}>
                    {selectedUser ? `${selectedUser.name} さんの給与詳細履歴です。` : "ユーザーを選択すると給与詳細履歴を表示します。"}
                  </p>
                </div>

                <Button type="button" variant="secondary" onClick={handleStartCreate} disabled={!selectedUser}>
                  新規作成
                </Button>
              </div>

              <button type="button" className={styles.switchRow} onClick={handleToggleIncludeDeletedSalaryDetails} disabled={!selectedUser}>
                <span className={`${styles.switch} ${includeDeletedSalaryDetails ? styles.switchOn : ""}`}>
                  <span className={styles.switchThumb} />
                </span>

                <span className={styles.switchText}>削除済み給与詳細も含める</span>
              </button>

              <div className={styles.salaryDetailList}>
                {salaryDetails.map((salaryDetail) => (
                  <article
                    key={salaryDetail.id}
                    className={`${styles.salaryDetailRow} ${salaryDetail.isDeleted ? styles.salaryDetailRowDeleted : ""}`}
                  >
                    <div className={styles.salaryDetailRowMain}>
                      <div>
                        <div className={styles.salaryDetailNameLine}>
                          <p className={styles.salaryDetailTitle}>
                            {getSalaryTypeLabel(salaryDetail.salaryType)} / {formatAmount(salaryDetail.baseAmount)}
                          </p>

                          {!salaryDetail.isPayrollTarget && <span className={styles.warningBadge}>給与対象外</span>}
                          {salaryDetail.isDeleted && <span className={styles.deletedBadge}>削除済み</span>}
                        </div>

                        <p className={styles.salaryDetailMeta}>
                          適用期間：{formatDate(salaryDetail.effectiveFrom)} ～ {formatDate(salaryDetail.effectiveTo)}
                        </p>
                        <p className={styles.salaryDetailMeta}>
                          固定手当：{formatAmount(salaryDetail.extraAllowanceAmount)} / 固定控除：{formatAmount(salaryDetail.fixedDeductionAmount)}
                        </p>
                      </div>

                      <div className={styles.rowActions}>
                        <Button type="button" variant="secondary" onClick={() => handleStartEdit(salaryDetail)} disabled={salaryDetail.isDeleted}>
                          編集
                        </Button>

                        <Button type="button" variant="danger" onClick={() => handleDeleteUserSalaryDetail(salaryDetail)} disabled={salaryDetail.isDeleted}>
                          削除
                        </Button>
                      </div>
                    </div>
                  </article>
                ))}

                {selectedUser && salaryDetails.length === 0 && <p className={styles.emptyText}>給与詳細が見つかりません。</p>}
                {!selectedUser && <p className={styles.emptyText}>先にユーザーを選択してください。</p>}
              </div>

              {salaryDetailHasMore && (
                <div className={styles.moreButtonArea}>
                  <Button type="button" variant="secondary" onClick={handleLoadMoreSalaryDetails} disabled={isSalaryDetailSearching}>
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

              {selectedUser && (
                <div className={styles.selectedUserBox}>
                  <p className={styles.selectedUserLabel}>対象ユーザー</p>
                  <p className={styles.selectedUserName}>{selectedUser.name}</p>
                  <p className={styles.selectedUserMeta}>{selectedUser.email}</p>
                </div>
              )}

              <div className={styles.formGrid}>
                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>給与区分</span>
                  <select
                    className={styles.input}
                    title="給与区分"
                    value={form.salaryType}
                    onChange={(event) => setForm((current) => ({ ...current, salaryType: event.target.value as SalaryType }))}
                    disabled={!selectedUser}
                  >
                    <option value="MONTHLY">月給</option>
                    <option value="HOURLY">時給</option>
                    <option value="DAILY">日給</option>
                  </select>
                </label>

                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>基本金額</span>
                  <input
                    className={styles.input}
                    type="number"
                    min="0"
                    value={form.baseAmount}
                    placeholder="例：250000"
                    onChange={(event) => setForm((current) => ({ ...current, baseAmount: event.target.value }))}
                    disabled={!selectedUser}
                  />
                </label>

                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>その他固定手当</span>
                  <input
                    className={styles.input}
                    type="number"
                    min="0"
                    value={form.extraAllowanceAmount}
                    placeholder="例：10000"
                    onChange={(event) => setForm((current) => ({ ...current, extraAllowanceAmount: event.target.value }))}
                    disabled={!selectedUser}
                  />
                </label>

                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>その他固定手当メモ</span>
                  <textarea
                    className={styles.textarea}
                    value={form.extraAllowanceMemo}
                    placeholder="例：住宅手当、資格手当など"
                    onChange={(event) => setForm((current) => ({ ...current, extraAllowanceMemo: event.target.value }))}
                    disabled={!selectedUser}
                  />
                </label>

                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>その他固定控除</span>
                  <input
                    className={styles.input}
                    type="number"
                    min="0"
                    value={form.fixedDeductionAmount}
                    placeholder="例：5000"
                    onChange={(event) => setForm((current) => ({ ...current, fixedDeductionAmount: event.target.value }))}
                    disabled={!selectedUser}
                  />
                </label>

                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>その他固定控除メモ</span>
                  <textarea
                    className={styles.textarea}
                    value={form.fixedDeductionMemo}
                    placeholder="例：社宅控除など"
                    onChange={(event) => setForm((current) => ({ ...current, fixedDeductionMemo: event.target.value }))}
                    disabled={!selectedUser}
                  />
                </label>

                <label className={styles.checkboxRow}>
                  <input
                    type="checkbox"
                    checked={form.isPayrollTarget}
                    onChange={(event) => setForm((current) => ({ ...current, isPayrollTarget: event.target.checked }))}
                    disabled={!selectedUser}
                  />
                  <span>給与計算対象にする</span>
                </label>

                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>適用開始日</span>
                  <input
                    className={styles.input}
                    type="date"
                    value={form.effectiveFrom}
                    onChange={(event) => setForm((current) => ({ ...current, effectiveFrom: event.target.value }))}
                    disabled={!selectedUser}
                  />
                </label>

                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>適用終了日</span>
                  <input
                    className={styles.input}
                    type="date"
                    value={form.effectiveTo}
                    onChange={(event) => setForm((current) => ({ ...current, effectiveTo: event.target.value }))}
                    disabled={!selectedUser}
                  />
                </label>

                <label className={styles.fieldWide}>
                  <span className={styles.fieldLabel}>メモ</span>
                  <textarea
                    className={styles.textarea}
                    value={form.memo}
                    placeholder="給与詳細全体のメモ"
                    onChange={(event) => setForm((current) => ({ ...current, memo: event.target.value }))}
                    disabled={!selectedUser}
                  />
                </label>
              </div>

              <div className={styles.formActions}>
                {isEditing ? (
                  <Button type="button" variant="primary" onClick={handleUpdateUserSalaryDetail} disabled={isPageLoading || !selectedUser}>
                    更新
                  </Button>
                ) : (
                  <Button type="button" variant="primary" onClick={handleCreateUserSalaryDetail} disabled={isPageLoading || !selectedUser}>
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
