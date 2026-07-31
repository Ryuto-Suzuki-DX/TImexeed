"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import { searchBusinessTargetUsers } from "@/api/admin/user";
import {
  createUserSalaryDetail,
  deleteUserSalaryDetail,
  getUserSalaryDetail,
  searchUserSalaryDetails,
  updateUserSalaryDetail,
} from "@/api/admin/userSalaryDetail";
import { searchAllowanceTypes } from "@/api/admin/allowanceType";
import {
  createMonthlyAllowance,
  deleteMonthlyAllowance,
  getMonthlyAllowanceDetail,
  searchMonthlyAllowances,
  updateMonthlyAllowance,
} from "@/api/admin/monthlyAllowance";
import type { UserResponse } from "@/types/admin/user";
import type { AllowanceTypeResponse } from "@/types/admin/allowanceType";
import type { MonthlyAllowanceResponse } from "@/types/admin/monthlyAllowance";
import type { SalaryType, UserSalaryDetailResponse } from "@/types/admin/userSalaryDetail";
import styles from "./page.module.css";

type PageMessageVariant = "info" | "success" | "warning" | "error";

type UserSalaryDetailForm = {
  userSalaryDetailId: number | null;
  salaryType: SalaryType;
  baseAmount: string;
  isPayrollTarget: boolean;
  effectiveFrom: string;
  effectiveTo: string;
  memo: string;
};

type MonthlyAllowanceForm = {
  monthlyAllowanceId: number | null;
  targetYear: string;
  targetMonth: string;
  allowanceTypeId: string;
  amount: string;
  memo: string;
};

const now = new Date();

const initialSalaryForm: UserSalaryDetailForm = {
  userSalaryDetailId: null,
  salaryType: "MONTHLY",
  baseAmount: "0",
  isPayrollTarget: true,
  effectiveFrom: "",
  effectiveTo: "",
  memo: "",
};

const initialAllowanceForm: MonthlyAllowanceForm = {
  monthlyAllowanceId: null,
  targetYear: String(now.getFullYear()),
  targetMonth: String(now.getMonth() + 1),
  allowanceTypeId: "",
  amount: "0",
  memo: "",
};

function toDateInputValue(value: string | null | undefined) {
  return value ? (value.split("T")[0] ?? "") : "";
}

function formatDate(value: string | null | undefined) {
  const dateValue = toDateInputValue(value);
  return dateValue ? dateValue.replaceAll("-", "/") : "-";
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
  }
}

function toNumberValue(value: string) {
  const normalizedValue = value.trim();
  return normalizedValue ? Number(normalizedValue) : 0;
}

function toSalaryForm(detail: UserSalaryDetailResponse): UserSalaryDetailForm {
  return {
    userSalaryDetailId: detail.id,
    salaryType: detail.salaryType,
    baseAmount: String(detail.baseAmount),
    isPayrollTarget: detail.isPayrollTarget,
    effectiveFrom: toDateInputValue(detail.effectiveFrom),
    effectiveTo: toDateInputValue(detail.effectiveTo),
    memo: detail.memo,
  };
}

function toAllowanceForm(detail: MonthlyAllowanceResponse): MonthlyAllowanceForm {
  return {
    monthlyAllowanceId: detail.id,
    targetYear: String(detail.targetYear),
    targetMonth: String(detail.targetMonth),
    allowanceTypeId: String(detail.allowanceTypeId),
    amount: String(detail.amount),
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

  const [salaryDetails, setSalaryDetails] = useState<UserSalaryDetailResponse[]>([]);
  const [salaryDetailOffset, setSalaryDetailOffset] = useState(0);
  const [salaryDetailHasMore, setSalaryDetailHasMore] = useState(false);
  const [includeDeletedSalaryDetails, setIncludeDeletedSalaryDetails] = useState(false);
  const [salaryForm, setSalaryForm] = useState<UserSalaryDetailForm>(initialSalaryForm);

  const [allowanceTypes, setAllowanceTypes] = useState<AllowanceTypeResponse[]>([]);
  const [monthlyAllowances, setMonthlyAllowances] = useState<MonthlyAllowanceResponse[]>([]);
  const [monthlyAllowanceOffset, setMonthlyAllowanceOffset] = useState(0);
  const [monthlyAllowanceHasMore, setMonthlyAllowanceHasMore] = useState(false);
  const [includeDeletedMonthlyAllowances, setIncludeDeletedMonthlyAllowances] = useState(false);
  const [allowanceForm, setAllowanceForm] = useState<MonthlyAllowanceForm>(initialAllowanceForm);

  const [pageMessage, setPageMessage] = useState("給与・手当を管理するユーザーを検索してください。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");
  const [isPageLoading, setIsPageLoading] = useState(false);
  const [isUserSearching, setIsUserSearching] = useState(false);

  const salaryFormTitle = useMemo(
    () => (salaryForm.userSalaryDetailId === null ? "給与詳細新規作成" : "給与詳細編集"),
    [salaryForm.userSalaryDetailId],
  );

  const allowanceFormTitle = useMemo(
    () => (allowanceForm.monthlyAllowanceId === null ? "月次手当追加" : "月次手当編集"),
    [allowanceForm.monthlyAllowanceId],
  );

  const resetSalaryForm = () => setSalaryForm(initialSalaryForm);
  const resetAllowanceForm = () => setAllowanceForm(initialAllowanceForm);

  const loadAllowanceTypes = useCallback(async () => {
    const result = await searchAllowanceTypes({
      keyword: "",
      includeInactive: false,
      includeDeleted: false,
      offset: 0,
      limit: 50,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "手当種別の取得に失敗しました。");
      setPageMessageVariant("error");
      return false;
    }

    setAllowanceTypes(result.data.allowanceTypes);
    return true;
  }, []);

  const loadSalaryDetails = useCallback(
    async (targetUserId: number, nextOffset: number, append: boolean, includeDeleted = includeDeletedSalaryDetails) => {
      const result = await searchUserSalaryDetails({
        targetUserId,
        includeDeleted,
        offset: nextOffset,
        limit: 50,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "給与詳細の取得に失敗しました。");
        setPageMessageVariant("error");
        return false;
      }

      setSalaryDetails((current) =>
        append ? [...current, ...result.data!.userSalaryDetails] : result.data!.userSalaryDetails,
      );
      setSalaryDetailOffset(nextOffset + result.data.userSalaryDetails.length);
      setSalaryDetailHasMore(result.data.hasMore);
      return true;
    },
    [includeDeletedSalaryDetails],
  );

  const loadMonthlyAllowances = useCallback(
    async (
      targetUserId: number,
      nextOffset: number,
      append: boolean,
      includeDeleted = includeDeletedMonthlyAllowances,
    ) => {
      const result = await searchMonthlyAllowances({
        targetUserId,
        targetYear: null,
        targetMonth: null,
        allowanceTypeId: null,
        includeDeleted,
        offset: nextOffset,
        limit: 50,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "月次手当の取得に失敗しました。");
        setPageMessageVariant("error");
        return false;
      }

      setMonthlyAllowances((current) =>
        append ? [...current, ...result.data!.monthlyAllowances] : result.data!.monthlyAllowances,
      );
      setMonthlyAllowanceOffset(nextOffset + result.data.monthlyAllowances.length);
      setMonthlyAllowanceHasMore(result.data.hasMore);
      return true;
    },
    [includeDeletedMonthlyAllowances],
  );

  const handleSearchUsers = async (nextOffset: number, append: boolean) => {
    setIsUserSearching(true);
    const searchKeyword = append ? searchedKeyword : keyword;

    const result = await searchBusinessTargetUsers({
      keyword: searchKeyword,
      offset: nextOffset,
      limit: 50,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "ユーザー検索に失敗しました。");
      setPageMessageVariant("error");
      setIsUserSearching(false);
      return;
    }

    setSearchedKeyword(searchKeyword);
    setUsers((current) => (append ? [...current, ...result.data!.users] : result.data!.users));
    setUserOffset(nextOffset + result.data.users.length);
    setUserHasMore(result.data.hasMore);

    if (!append) {
      setSelectedUser(null);
      setSalaryDetails([]);
      setMonthlyAllowances([]);
      resetSalaryForm();
      resetAllowanceForm();
    }

    setPageMessage(result.data.users.length === 0 ? "該当するユーザーが見つかりません。" : "ユーザー検索が完了しました。");
    setPageMessageVariant(result.data.users.length === 0 ? "warning" : "success");
    setIsUserSearching(false);
  };

  const handleSelectUser = async (targetUser: UserResponse) => {
    setSelectedUser(targetUser);
    resetSalaryForm();
    resetAllowanceForm();
    setIsPageLoading(true);
    setPageMessage(`${targetUser.name} さんの給与・手当情報を取得しています。`);
    setPageMessageVariant("info");

    const [salarySuccess, allowanceSuccess] = await Promise.all([
      loadSalaryDetails(targetUser.id, 0, false),
      loadMonthlyAllowances(targetUser.id, 0, false),
    ]);

    if (salarySuccess && allowanceSuccess) {
      setPageMessage(`${targetUser.name} さんの給与・手当情報を取得しました。`);
      setPageMessageVariant("success");
    }
    setIsPageLoading(false);
  };

  const validateSalaryForm = () => {
    if (!selectedUser) {
      setPageMessage("対象ユーザーを選択してください。");
      setPageMessageVariant("error");
      return false;
    }

    if (!salaryForm.effectiveFrom) {
      setPageMessage("給与詳細の適用開始日を入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    if (salaryForm.effectiveTo && salaryForm.effectiveFrom > salaryForm.effectiveTo) {
      setPageMessage("適用終了日は適用開始日以降にしてください。");
      setPageMessageVariant("error");
      return false;
    }

    const baseAmount = toNumberValue(salaryForm.baseAmount);
    if (Number.isNaN(baseAmount) || baseAmount < 0) {
      setPageMessage("基本金額は0以上の数値で入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    return true;
  };

  const handleSaveSalaryDetail = async () => {
    if (!validateSalaryForm() || !selectedUser) {
      return;
    }

    setIsPageLoading(true);
    const commonRequest = {
      salaryType: salaryForm.salaryType,
      baseAmount: toNumberValue(salaryForm.baseAmount),
      isPayrollTarget: salaryForm.isPayrollTarget,
      effectiveFrom: salaryForm.effectiveFrom,
      effectiveTo: salaryForm.effectiveTo || null,
      memo: salaryForm.memo,
    };

    const result =
      salaryForm.userSalaryDetailId === null
        ? await createUserSalaryDetail({ targetUserId: selectedUser.id, ...commonRequest })
        : await updateUserSalaryDetail({
            userSalaryDetailId: salaryForm.userSalaryDetailId,
            ...commonRequest,
          });

    if (result.error || !result.data) {
      setPageMessage(result.message || "給与詳細の保存に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetSalaryForm();
    await loadSalaryDetails(selectedUser.id, 0, false);
    setPageMessage(result.message || "給与詳細を保存しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  const handleEditSalaryDetail = async (detail: UserSalaryDetailResponse) => {
    if (detail.isDeleted) return;

    setIsPageLoading(true);
    const result = await getUserSalaryDetail({ userSalaryDetailId: detail.id });

    if (result.error || !result.data) {
      setPageMessage(result.message || "給与詳細の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    setSalaryForm(toSalaryForm(result.data.userSalaryDetail));
    setPageMessage("給与詳細を編集できます。");
    setPageMessageVariant("info");
    setIsPageLoading(false);
  };

  const handleDeleteSalaryDetail = async (detail: UserSalaryDetailResponse) => {
    if (!selectedUser || detail.isDeleted) return;
    if (!window.confirm(`${getSalaryTypeLabel(detail.salaryType)} ${formatAmount(detail.baseAmount)} の給与詳細を削除しますか？`)) return;

    setIsPageLoading(true);
    const result = await deleteUserSalaryDetail({ userSalaryDetailId: detail.id });

    if (result.error) {
      setPageMessage(result.message || "給与詳細の削除に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    if (salaryForm.userSalaryDetailId === detail.id) resetSalaryForm();
    await loadSalaryDetails(selectedUser.id, 0, false);
    setPageMessage(result.message || "給与詳細を削除しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  const validateAllowanceForm = () => {
    if (!selectedUser) {
      setPageMessage("対象ユーザーを選択してください。");
      setPageMessageVariant("error");
      return false;
    }

    const year = Number(allowanceForm.targetYear);
    const month = Number(allowanceForm.targetMonth);
    const allowanceTypeId = Number(allowanceForm.allowanceTypeId);
    const amount = toNumberValue(allowanceForm.amount);

    if (!Number.isInteger(year) || year < 2000 || year > 2100) {
      setPageMessage("対象年を正しく入力してください。");
      setPageMessageVariant("error");
      return false;
    }
    if (!Number.isInteger(month) || month < 1 || month > 12) {
      setPageMessage("対象月を1～12で入力してください。");
      setPageMessageVariant("error");
      return false;
    }
    if (!allowanceTypeId) {
      setPageMessage("手当種別を選択してください。");
      setPageMessageVariant("error");
      return false;
    }
    if (Number.isNaN(amount) || amount < 0) {
      setPageMessage("手当金額は0以上の数値で入力してください。");
      setPageMessageVariant("error");
      return false;
    }

    return true;
  };

  const handleSaveMonthlyAllowance = async () => {
    if (!validateAllowanceForm() || !selectedUser) return;

    setIsPageLoading(true);
    const commonRequest = {
      targetUserId: selectedUser.id,
      targetYear: Number(allowanceForm.targetYear),
      targetMonth: Number(allowanceForm.targetMonth),
      allowanceTypeId: Number(allowanceForm.allowanceTypeId),
      amount: toNumberValue(allowanceForm.amount),
      memo: allowanceForm.memo,
    };

    const result =
      allowanceForm.monthlyAllowanceId === null
        ? await createMonthlyAllowance(commonRequest)
        : await updateMonthlyAllowance({
            monthlyAllowanceId: allowanceForm.monthlyAllowanceId,
            ...commonRequest,
          });

    if (result.error || !result.data) {
      setPageMessage(result.message || "月次手当の保存に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    resetAllowanceForm();
    await loadMonthlyAllowances(selectedUser.id, 0, false);
    setPageMessage(result.message || "月次手当を保存しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  const handleEditMonthlyAllowance = async (detail: MonthlyAllowanceResponse) => {
    if (detail.isDeleted) return;

    setIsPageLoading(true);
    const result = await getMonthlyAllowanceDetail({ monthlyAllowanceId: detail.id });

    if (result.error || !result.data) {
      setPageMessage(result.message || "月次手当の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    setAllowanceForm(toAllowanceForm(result.data.monthlyAllowance));
    setPageMessage("月次手当を編集できます。");
    setPageMessageVariant("info");
    setIsPageLoading(false);
  };

  const handleDeleteMonthlyAllowance = async (detail: MonthlyAllowanceResponse) => {
    if (!selectedUser || detail.isDeleted) return;
    if (!window.confirm(`${detail.targetYear}年${detail.targetMonth}月の「${detail.allowanceTypeName}」を削除しますか？`)) return;

    setIsPageLoading(true);
    const result = await deleteMonthlyAllowance({ monthlyAllowanceId: detail.id });

    if (result.error) {
      setPageMessage(result.message || "月次手当の削除に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    if (allowanceForm.monthlyAllowanceId === detail.id) resetAllowanceForm();
    await loadMonthlyAllowances(selectedUser.id, 0, false);
    setPageMessage(result.message || "月次手当を削除しました。");
    setPageMessageVariant("success");
    setIsPageLoading(false);
  };

  useEffect(() => {
    if (isLoading || !user) return;

    const timerId = window.setTimeout(() => {
      void Promise.all([handleSearchUsers(0, false), loadAllowanceTypes()]);
    }, 0);

    return () => window.clearTimeout(timerId);
    // 初回だけ実行する
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, user]);

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />
        <section className={styles.loadingCard}>
          <PageTitle title="ユーザー給与・月次手当管理" description="ログイン情報を確認しています。" />
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
              title="ユーザー給与・月次手当管理"
              description="対象ユーザーを選択し、給与詳細と月ごとの手当を管理します。"
            />
            <MessageBox variant={pageMessageVariant}>{isPageLoading ? "処理中..." : pageMessage}</MessageBox>
          </div>

          <section className={styles.userSearchPanel}>
            <div className={styles.sectionHeader}>
              <div>
                <h2 className={styles.sectionTitle}>ユーザー検索</h2>
                <p className={styles.sectionDescription}>給与・手当を設定する一般ユーザーを選択します。</p>
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
                <button
                  key={targetUser.id}
                  type="button"
                  className={`${styles.userRow} ${selectedUser?.id === targetUser.id ? styles.userRowSelected : ""}`}
                  onClick={() => handleSelectUser(targetUser)}
                >
                  <span className={styles.userName}>{targetUser.name}</span>
                  <span className={styles.userMeta}>{targetUser.email}</span>
                </button>
              ))}
            </div>

            {userHasMore && (
              <div className={styles.moreButtonArea}>
                <Button type="button" variant="secondary" onClick={() => handleSearchUsers(userOffset, true)}>
                  さらに表示
                </Button>
              </div>
            )}
          </section>

          {selectedUser && (
            <div className={styles.selectedUserBox}>
              <p className={styles.selectedUserLabel}>対象ユーザー</p>
              <p className={styles.selectedUserName}>{selectedUser.name}</p>
              <p className={styles.selectedUserMeta}>{selectedUser.email}</p>
            </div>
          )}

          <div className={styles.managementGrid}>
            <section className={styles.managementPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>給与詳細</h2>
                  <p className={styles.sectionDescription}>給与区分、基本金額、適用期間を履歴管理します。</p>
                </div>
              </div>

              <div className={styles.formCard}>
                <div className={styles.formHeader}>
                  <h3 className={styles.formTitle}>{salaryFormTitle}</h3>
                  {salaryForm.userSalaryDetailId !== null && (
                    <Button type="button" variant="secondary" onClick={resetSalaryForm}>編集取消</Button>
                  )}
                </div>

                <div className={styles.formGrid}>
                  <label className={styles.field}>
                    <span className={styles.fieldLabel}>給与区分</span>
                    <select
                      className={styles.input}
                      value={salaryForm.salaryType}
                      onChange={(event) =>
                        setSalaryForm((current) => ({ ...current, salaryType: event.target.value as SalaryType }))
                      }
                      disabled={!selectedUser}
                    >
                      <option value="MONTHLY">月給</option>
                      <option value="HOURLY">時給</option>
                      <option value="DAILY">日給</option>
                    </select>
                  </label>

                  <label className={styles.field}>
                    <span className={styles.fieldLabel}>基本金額</span>
                    <input
                      className={styles.input}
                      type="number"
                      min="0"
                      value={salaryForm.baseAmount}
                      onChange={(event) =>
                        setSalaryForm((current) => ({ ...current, baseAmount: event.target.value }))
                      }
                      disabled={!selectedUser}
                    />
                  </label>

                  <label className={styles.field}>
                    <span className={styles.fieldLabel}>適用開始日</span>
                    <input
                      className={styles.input}
                      type="date"
                      value={salaryForm.effectiveFrom}
                      onChange={(event) =>
                        setSalaryForm((current) => ({ ...current, effectiveFrom: event.target.value }))
                      }
                      disabled={!selectedUser}
                    />
                  </label>

                  <label className={styles.field}>
                    <span className={styles.fieldLabel}>適用終了日</span>
                    <input
                      className={styles.input}
                      type="date"
                      value={salaryForm.effectiveTo}
                      onChange={(event) =>
                        setSalaryForm((current) => ({ ...current, effectiveTo: event.target.value }))
                      }
                      disabled={!selectedUser}
                    />
                  </label>

                  <label className={styles.fieldWide}>
                    <span className={styles.fieldLabel}>メモ</span>
                    <textarea
                      className={styles.textarea}
                      value={salaryForm.memo}
                      onChange={(event) =>
                        setSalaryForm((current) => ({ ...current, memo: event.target.value }))
                      }
                      disabled={!selectedUser}
                    />
                  </label>

                  <label className={styles.checkboxRow}>
                    <input
                      type="checkbox"
                      checked={salaryForm.isPayrollTarget}
                      onChange={(event) =>
                        setSalaryForm((current) => ({ ...current, isPayrollTarget: event.target.checked }))
                      }
                      disabled={!selectedUser}
                    />
                    給与計算対象にする
                  </label>
                </div>

                <div className={styles.formActions}>
                  <Button type="button" variant="primary" onClick={handleSaveSalaryDetail} disabled={!selectedUser || isPageLoading}>
                    {salaryForm.userSalaryDetailId === null ? "作成" : "更新"}
                  </Button>
                </div>
              </div>

              <button
                type="button"
                className={styles.switchRow}
                onClick={async () => {
                  if (!selectedUser) return;
                  const next = !includeDeletedSalaryDetails;
                  setIncludeDeletedSalaryDetails(next);
                  await loadSalaryDetails(selectedUser.id, 0, false, next);
                }}
                disabled={!selectedUser}
              >
                <span className={`${styles.switch} ${includeDeletedSalaryDetails ? styles.switchOn : ""}`}>
                  <span className={styles.switchThumb} />
                </span>
                削除済みも含める
              </button>

              <div className={styles.recordList}>
                {salaryDetails.map((detail) => (
                  <article key={detail.id} className={`${styles.recordRow} ${detail.isDeleted ? styles.deletedRow : ""}`}>
                    <div>
                      <p className={styles.recordTitle}>
                        {getSalaryTypeLabel(detail.salaryType)} / {formatAmount(detail.baseAmount)}
                      </p>
                      <p className={styles.recordMeta}>
                        適用期間：{formatDate(detail.effectiveFrom)} ～ {formatDate(detail.effectiveTo)}
                      </p>
                    </div>
                    <div className={styles.rowActions}>
                      <Button type="button" variant="secondary" onClick={() => handleEditSalaryDetail(detail)} disabled={detail.isDeleted}>
                        編集
                      </Button>
                      <Button type="button" variant="danger" onClick={() => handleDeleteSalaryDetail(detail)} disabled={detail.isDeleted}>
                        削除
                      </Button>
                    </div>
                  </article>
                ))}
              </div>

              {salaryDetailHasMore && (
                <div className={styles.moreButtonArea}>
                  <Button type="button" variant="secondary" onClick={() => selectedUser && loadSalaryDetails(selectedUser.id, salaryDetailOffset, true)}>
                    給与詳細をさらに表示
                  </Button>
                </div>
              )}
            </section>

            <section className={styles.managementPanel}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>月次手当</h2>
                  <p className={styles.sectionDescription}>対象年月、手当種別、金額、メモを明細単位で登録します。</p>
                </div>
              </div>

              <div className={styles.formCard}>
                <div className={styles.formHeader}>
                  <h3 className={styles.formTitle}>{allowanceFormTitle}</h3>
                  {allowanceForm.monthlyAllowanceId !== null && (
                    <Button type="button" variant="secondary" onClick={resetAllowanceForm}>編集取消</Button>
                  )}
                </div>

                <div className={styles.formGrid}>
                  <label className={styles.field}>
                    <span className={styles.fieldLabel}>対象年</span>
                    <input
                      className={styles.input}
                      type="number"
                      min="2000"
                      max="2100"
                      value={allowanceForm.targetYear}
                      onChange={(event) =>
                        setAllowanceForm((current) => ({ ...current, targetYear: event.target.value }))
                      }
                      disabled={!selectedUser}
                    />
                  </label>

                  <label className={styles.field}>
                    <span className={styles.fieldLabel}>対象月</span>
                    <select
                      className={styles.input}
                      value={allowanceForm.targetMonth}
                      onChange={(event) =>
                        setAllowanceForm((current) => ({ ...current, targetMonth: event.target.value }))
                      }
                      disabled={!selectedUser}
                    >
                      {Array.from({ length: 12 }, (_, index) => index + 1).map((month) => (
                        <option key={month} value={month}>{month}月</option>
                      ))}
                    </select>
                  </label>

                  <label className={styles.fieldWide}>
                    <span className={styles.fieldLabel}>手当種別</span>
                    <select
                      className={styles.input}
                      value={allowanceForm.allowanceTypeId}
                      onChange={(event) =>
                        setAllowanceForm((current) => ({ ...current, allowanceTypeId: event.target.value }))
                      }
                      disabled={!selectedUser}
                    >
                      <option value="">選択してください</option>
                      {allowanceTypes.map((allowanceType) => (
                        <option key={allowanceType.id} value={allowanceType.id}>
                          {allowanceType.name}
                        </option>
                      ))}
                    </select>
                  </label>

                  <label className={styles.field}>
                    <span className={styles.fieldLabel}>金額</span>
                    <input
                      className={styles.input}
                      type="number"
                      min="0"
                      value={allowanceForm.amount}
                      onChange={(event) =>
                        setAllowanceForm((current) => ({ ...current, amount: event.target.value }))
                      }
                      disabled={!selectedUser}
                    />
                  </label>

                  <label className={styles.fieldWide}>
                    <span className={styles.fieldLabel}>メモ</span>
                    <textarea
                      className={styles.textarea}
                      value={allowanceForm.memo}
                      onChange={(event) =>
                        setAllowanceForm((current) => ({ ...current, memo: event.target.value }))
                      }
                      disabled={!selectedUser}
                    />
                  </label>
                </div>

                <div className={styles.formActions}>
                  <Button type="button" variant="primary" onClick={handleSaveMonthlyAllowance} disabled={!selectedUser || isPageLoading}>
                    {allowanceForm.monthlyAllowanceId === null ? "追加" : "更新"}
                  </Button>
                </div>
              </div>

              <button
                type="button"
                className={styles.switchRow}
                onClick={async () => {
                  if (!selectedUser) return;
                  const next = !includeDeletedMonthlyAllowances;
                  setIncludeDeletedMonthlyAllowances(next);
                  await loadMonthlyAllowances(selectedUser.id, 0, false, next);
                }}
                disabled={!selectedUser}
              >
                <span className={`${styles.switch} ${includeDeletedMonthlyAllowances ? styles.switchOn : ""}`}>
                  <span className={styles.switchThumb} />
                </span>
                削除済みも含める
              </button>

              <div className={styles.recordList}>
                {monthlyAllowances.map((detail) => (
                  <article key={detail.id} className={`${styles.recordRow} ${detail.isDeleted ? styles.deletedRow : ""}`}>
                    <div>
                      <p className={styles.recordTitle}>
                        {detail.targetYear}年{detail.targetMonth}月 / {detail.allowanceTypeName}
                      </p>
                      <p className={styles.recordMeta}>
                        {formatAmount(detail.amount)} / メモ：{detail.memo || "-"}
                      </p>
                    </div>
                    <div className={styles.rowActions}>
                      <Button type="button" variant="secondary" onClick={() => handleEditMonthlyAllowance(detail)} disabled={detail.isDeleted}>
                        編集
                      </Button>
                      <Button type="button" variant="danger" onClick={() => handleDeleteMonthlyAllowance(detail)} disabled={detail.isDeleted}>
                        削除
                      </Button>
                    </div>
                  </article>
                ))}
              </div>

              {monthlyAllowanceHasMore && (
                <div className={styles.moreButtonArea}>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => selectedUser && loadMonthlyAllowances(selectedUser.id, monthlyAllowanceOffset, true)}
                  >
                    月次手当をさらに表示
                  </Button>
                </div>
              )}
            </section>
          </div>
        </section>
      </div>
    </PageContainer>
  );
}
