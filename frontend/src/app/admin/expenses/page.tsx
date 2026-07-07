"use client";

import { ChangeEvent, FormEvent, useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import {
  createExpense,
  deleteExpense,
  openExpenseReceiptInNewTab,
  searchExpenses,
  updateExpense,
} from "@/api/admin/expense";
import type {
  CreateExpenseRequest,
  ExpenseListItemResponse,
  SearchExpensesResponse,
  UpdateExpenseRequest,
} from "@/types/admin/expense";
import { useRequireRole } from "@/hooks/useRequireRole";
import styles from "./page.module.css";

type PageMessage = {
  variant: "info" | "success" | "warning" | "error";
  text: string;
};

type ExpenseFormState = {
  expenseId: number | null;
  targetMonth: string;
  expenseDate: string;
  amount: string;
  description: string;
  memo: string;
  receiptFile: File | null;
};

type SearchFormState = {
  keyword: string;
  targetMonthFrom: string;
  targetMonthTo: string;
};

type TargetUserCandidate = {
  id: number;
  name: string;
  email: string;
  role?: string;
  isDeleted?: boolean;
};

type SearchUsersResponse = {
  users: TargetUserCandidate[];
  total?: number;
  offset?: number;
  limit?: number;
  hasMore?: boolean;
};

type ApiResponse<TData> = {
  data: TData | null;
  error: boolean;
  code: string;
  message: string;
  details?: unknown;
};

const PAGE_LIMIT = 50;
const TARGET_USER_SEARCH_LIMIT = 10;

const initialExpenseForm: ExpenseFormState = {
  expenseId: null,
  targetMonth: getCurrentMonthText(),
  expenseDate: getTodayText(),
  amount: "",
  description: "",
  memo: "",
  receiptFile: null,
};

const initialSearchForm: SearchFormState = {
  keyword: "",
  targetMonthFrom: getCurrentMonthText(),
  targetMonthTo: getCurrentMonthText(),
};

export default function AdminExpensesPage() {
  const { user, isLoading, message: authMessage } = useRequireRole("ADMIN");

  const [searchForm, setSearchForm] = useState<SearchFormState>(initialSearchForm);
  const [expenseForm, setExpenseForm] = useState<ExpenseFormState>(initialExpenseForm);

  const [targetUserKeyword, setTargetUserKeyword] = useState("");
  const [targetUserCandidates, setTargetUserCandidates] = useState<TargetUserCandidate[]>([]);
  const [selectedTargetUser, setSelectedTargetUser] = useState<TargetUserCandidate | null>(null);

  const [expenses, setExpenses] = useState<ExpenseListItemResponse[]>([]);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [isSearching, setIsSearching] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isDeletingId, setIsDeletingId] = useState<number | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ExpenseListItemResponse | null>(null);
  const [isViewingReceiptId, setIsViewingReceiptId] = useState<number | null>(null);
  const [isTargetUserSearching, setIsTargetUserSearching] = useState(false);
  const [pageMessage, setPageMessage] = useState<PageMessage>({
    variant: "info",
    text: "対象月の期間と従業員キーワードで経費を検索できます。",
  });

  const isEditMode = expenseForm.expenseId !== null;

  const shownCountText = useMemo(() => {
    if (expenses.length === 0) {
      return "表示 0件";
    }

    return `表示 ${expenses.length}件 / 全${total}件`;
  }, [expenses.length, total]);

  useEffect(() => {
    if (!user) {
      return;
    }

    void handleSearch(0, false);
    // 初回検索だけなので依存はuserのみ
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="経費管理" description="ログイン情報を確認しています。" />
          <MessageBox variant="info">{authMessage}</MessageBox>
        </section>
      </PageContainer>
    );
  }

  async function handleSearch(offset: number, append: boolean) {
    setIsSearching(true);
    setPageMessage({
      variant: "info",
      text: append ? "追加の経費を取得しています。" : "経費一覧を取得しています。",
    });

    try {
      const response = await searchExpenses({
        keyword: searchForm.keyword.trim(),
        targetMonthFrom: searchForm.targetMonthFrom,
        targetMonthTo: searchForm.targetMonthTo,
        offset,
        limit: PAGE_LIMIT,
      });

      const data = response.data;
      if (!data) {
        throw new Error("経費一覧の取得結果が空です。");
      }

      setSearchResult(data, append);
      setPageMessage({
        variant: "success",
        text: "経費一覧を取得しました。",
      });
    } catch (error) {
      setPageMessage({
        variant: "error",
        text: error instanceof Error ? error.message : "経費一覧の取得に失敗しました。",
      });
    } finally {
      setIsSearching(false);
    }
  }

  async function handleSearchTargetUsers() {
    const keyword = targetUserKeyword.trim();
    if (!keyword) {
      setPageMessage({
        variant: "warning",
        text: "対象従業員を検索するキーワードを入力してください。",
      });
      return;
    }

    setIsTargetUserSearching(true);
    setTargetUserCandidates([]);
    setPageMessage({
      variant: "info",
      text: "対象従業員を検索しています。",
    });

    try {
      const response = await searchTargetUsers({
        keyword,
        includeDeleted: false,
        offset: 0,
        limit: TARGET_USER_SEARCH_LIMIT,
      });

      const candidates = response.users.filter((candidate) => !candidate.isDeleted);
      setTargetUserCandidates(candidates);

      if (candidates.length === 0) {
        setPageMessage({
          variant: "warning",
          text: "該当する従業員が見つかりませんでした。",
        });
        return;
      }

      setPageMessage({
        variant: "success",
        text: "対象従業員の候補を取得しました。登録対象を選択してください。",
      });
    } catch (error) {
      setPageMessage({
        variant: "error",
        text: error instanceof Error ? error.message : "対象従業員の検索に失敗しました。",
      });
    } finally {
      setIsTargetUserSearching(false);
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const validationMessage = validateExpenseForm(expenseForm, selectedTargetUser);
    if (validationMessage) {
      setPageMessage({
        variant: "warning",
        text: validationMessage,
      });
      return;
    }

    if (!selectedTargetUser) {
      return;
    }

    setIsSaving(true);
    setPageMessage({
      variant: "info",
      text: isEditMode ? "経費を更新しています。" : "経費を登録しています。",
    });

    try {
      if (isEditMode) {
        const request: UpdateExpenseRequest = {
          expenseId: expenseForm.expenseId as number,
          targetUserId: selectedTargetUser.id,
          targetMonth: expenseForm.targetMonth,
          expenseDate: expenseForm.expenseDate,
          amount: Number(expenseForm.amount),
          description: expenseForm.description.trim(),
          memo: normalizeNullableText(expenseForm.memo),
          receiptFile: expenseForm.receiptFile,
        };

        const response = await updateExpense(request);
        if (!response.data) {
          throw new Error("経費更新の結果が空です。");
        }
      } else {
        const request: CreateExpenseRequest = {
          targetUserId: selectedTargetUser.id,
          targetMonth: expenseForm.targetMonth,
          expenseDate: expenseForm.expenseDate,
          amount: Number(expenseForm.amount),
          description: expenseForm.description.trim(),
          memo: normalizeNullableText(expenseForm.memo),
          receiptFile: expenseForm.receiptFile,
        };

        const response = await createExpense(request);
        if (!response.data) {
          throw new Error("経費登録の結果が空です。");
        }
      }

      resetExpenseInputState();
      setPageMessage({
        variant: "success",
        text: isEditMode ? "経費を更新しました。" : "経費を登録しました。",
      });

      await handleSearch(0, false);
    } catch (error) {
      setPageMessage({
        variant: "error",
        text: error instanceof Error ? error.message : "経費の保存に失敗しました。",
      });
    } finally {
      setIsSaving(false);
    }
  }

  function handleOpenDeleteModal(expense: ExpenseListItemResponse) {
    setDeleteTarget(expense);
  }

  function handleCloseDeleteModal() {
    if (isDeletingId !== null) {
      return;
    }

    setDeleteTarget(null);
  }

  async function handleDelete() {
    if (!deleteTarget) {
      return;
    }

    const expense = deleteTarget;

    setIsDeletingId(expense.id);
    setPageMessage({
      variant: "info",
      text: "経費を削除しています。",
    });

    try {
      const response = await deleteExpense({ expenseId: expense.id });
      if (!response.data) {
        throw new Error("経費削除の結果が空です。");
      }

      if (expenseForm.expenseId === expense.id) {
        resetExpenseInputState();
      }

      setDeleteTarget(null);
      setPageMessage({
        variant: "success",
        text: "経費を削除しました。",
      });

      await handleSearch(0, false);
    } catch (error) {
      setPageMessage({
        variant: "error",
        text: error instanceof Error ? error.message : "経費の削除に失敗しました。",
      });
    } finally {
      setIsDeletingId(null);
    }
  }

  async function handleViewReceipt(expense: ExpenseListItemResponse) {
    if (!expense.hasReceiptFile) {
      setPageMessage({
        variant: "warning",
        text: "この経費には領収書ファイルが登録されていません。",
      });
      return;
    }

    setIsViewingReceiptId(expense.id);

    try {
      await openExpenseReceiptInNewTab({ expenseId: expense.id });
      setPageMessage({
        variant: "success",
        text: "領収書ファイルを開きました。",
      });
    } catch (error) {
      setPageMessage({
        variant: "error",
        text: error instanceof Error ? error.message : "領収書ファイルの表示に失敗しました。",
      });
    } finally {
      setIsViewingReceiptId(null);
    }
  }

  function handleSelectTargetUser(candidate: TargetUserCandidate) {
    setSelectedTargetUser(candidate);
    setTargetUserKeyword(`${candidate.name} ${candidate.email}`);
    setTargetUserCandidates([]);
    setPageMessage({
      variant: "info",
      text: "登録対象の従業員を選択しました。",
    });
  }

  function handleEdit(expense: ExpenseListItemResponse) {
    const targetUser = {
      id: expense.userId,
      name: expense.userName,
      email: expense.email,
    };

    setExpenseForm({
      expenseId: expense.id,
      targetMonth: expense.targetMonth,
      expenseDate: expense.expenseDate,
      amount: String(expense.amount),
      description: expense.description,
      memo: expense.memo ?? "",
      receiptFile: null,
    });
    setSelectedTargetUser(targetUser);
    setTargetUserKeyword(`${targetUser.name} ${targetUser.email}`);
    setTargetUserCandidates([]);

    setPageMessage({
      variant: "info",
      text: "選択した経費を編集フォームに読み込みました。領収書を差し替える場合だけファイルを選択してください。",
    });

    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function handleResetForm() {
    resetExpenseInputState();
    setPageMessage({
      variant: "info",
      text: "入力フォームを新規登録状態に戻しました。",
    });
  }

  function resetExpenseInputState() {
    setExpenseForm(initialExpenseForm);
    setTargetUserKeyword("");
    setTargetUserCandidates([]);
    setSelectedTargetUser(null);
  }

  function setSearchResult(data: SearchExpensesResponse, append: boolean) {
    setExpenses((current) => (append ? [...current, ...data.expenses] : data.expenses));
    setTotal(data.total);
    setHasMore(data.hasMore);
  }

  return (
    <PageContainer>
      <AdminSideMenu />

      <div className={styles.pageWrap}>
        <section className={styles.pageCard}>
          <div className={styles.headerArea}>
            <PageTitle
              title="経費管理"
              description="上長確認済みの経費を登録し、従業員キーワードと対象月の期間で検索します。"
            />

            <MessageBox variant={pageMessage.variant}>{pageMessage.text}</MessageBox>
          </div>

          <div className={styles.contentGrid}>
            <section className={styles.formCard}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>{isEditMode ? "経費更新" : "経費登録"}</h2>
                  <p className={styles.sectionDescription}>
                    従業員を検索して登録対象に選択します。領収書画像またはPDFを添付できます。
                  </p>
                </div>

                {isEditMode && <span className={styles.editBadge}>編集中</span>}
              </div>

              <form className={styles.expenseForm} onSubmit={handleSubmit}>
                <div className={styles.userSearchArea}>
                  <label className={styles.fieldLabel}>
                    対象従業員検索
                    <input
                      className={styles.input}
                      type="text"
                      value={targetUserKeyword}
                      onChange={(event) => {
                        setTargetUserKeyword(event.target.value);
                        setSelectedTargetUser(null);
                      }}
                      placeholder="名前またはメールアドレスで検索"
                    />
                  </label>

                  <Button type="button" variant="secondary" onClick={handleSearchTargetUsers} disabled={isTargetUserSearching}>
                    {isTargetUserSearching ? "検索中..." : "従業員検索"}
                  </Button>
                </div>

                {selectedTargetUser && (
                  <div className={styles.selectedUserBox}>
                    <p className={styles.selectedUserTitle}>選択中の従業員</p>
                    <p className={styles.selectedUserName}>{selectedTargetUser.name}</p>
                    <p className={styles.selectedUserMeta}>
                      ID：{selectedTargetUser.id} / {selectedTargetUser.email}
                    </p>
                  </div>
                )}

                {targetUserCandidates.length > 0 && (
                  <div className={styles.userCandidateList}>
                    {targetUserCandidates.map((candidate) => (
                      <button
                        key={candidate.id}
                        type="button"
                        className={styles.userCandidateRow}
                        onClick={() => handleSelectTargetUser(candidate)}
                      >
                        <span className={styles.userCandidateName}>{candidate.name}</span>
                        <span className={styles.userCandidateEmail}>{candidate.email}</span>
                        <span className={styles.userCandidateMeta}>ID：{candidate.id}</span>
                      </button>
                    ))}
                  </div>
                )}

                <div className={styles.twoColumn}>
                  <label className={styles.fieldLabel}>
                    対象月
                    <input
                      className={styles.input}
                      type="month"
                      value={expenseForm.targetMonth}
                      onChange={(event) =>
                        setExpenseForm((current) => ({
                          ...current,
                          targetMonth: event.target.value,
                        }))
                      }
                    />
                  </label>

                  <label className={styles.fieldLabel}>
                    経費発生日
                    <input
                      className={styles.input}
                      type="date"
                      value={expenseForm.expenseDate}
                      onChange={(event) =>
                        setExpenseForm((current) => ({
                          ...current,
                          expenseDate: event.target.value,
                        }))
                      }
                    />
                  </label>
                </div>

                <label className={styles.fieldLabel}>
                  金額
                  <input
                    className={styles.input}
                    type="number"
                    min="1"
                    value={expenseForm.amount}
                    onChange={(event) =>
                      setExpenseForm((current) => ({
                        ...current,
                        amount: event.target.value,
                      }))
                    }
                    placeholder="例：1200"
                  />
                </label>

                <label className={styles.fieldLabel}>
                  内容
                  <input
                    className={styles.input}
                    type="text"
                    value={expenseForm.description}
                    onChange={(event) =>
                      setExpenseForm((current) => ({
                        ...current,
                        description: event.target.value,
                      }))
                    }
                    placeholder="例：取引先訪問時の交通費"
                  />
                </label>

                <label className={styles.fieldLabel}>
                  メモ
                  <textarea
                    className={styles.textarea}
                    value={expenseForm.memo}
                    onChange={(event) =>
                      setExpenseForm((current) => ({
                        ...current,
                        memo: event.target.value,
                      }))
                    }
                    placeholder="任意"
                  />
                </label>

                <label className={styles.fieldLabel}>
                  領収書ファイル
                  <input
                    className={styles.fileInput}
                    type="file"
                    accept="image/*,.pdf"
                    onChange={(event: ChangeEvent<HTMLInputElement>) => {
                      const file = event.target.files?.[0] ?? null;
                      setExpenseForm((current) => ({
                        ...current,
                        receiptFile: file,
                      }));
                    }}
                  />
                </label>

                {expenseForm.receiptFile && (
                  <p className={styles.fileNameText}>選択中：{expenseForm.receiptFile.name}</p>
                )}

                <div className={styles.formActions}>
                  <Button type="submit" variant="primary" disabled={isSaving}>
                    {isSaving ? "保存中..." : isEditMode ? "更新する" : "登録する"}
                  </Button>

                  <Button type="button" variant="secondary" onClick={handleResetForm} disabled={isSaving}>
                    入力をクリア
                  </Button>
                </div>
              </form>
            </section>

            <section className={styles.searchCard}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>経費検索</h2>
                  <p className={styles.sectionDescription}>
                    従業員名・メールアドレスのフリーワードと、対象月の期間で検索します。
                  </p>
                </div>
              </div>

              <form
                className={styles.searchForm}
                onSubmit={(event) => {
                  event.preventDefault();
                  void handleSearch(0, false);
                }}
              >
                <label className={styles.fieldLabel}>
                  従業員キーワード
                  <input
                    className={styles.input}
                    type="text"
                    value={searchForm.keyword}
                    onChange={(event) =>
                      setSearchForm((current) => ({
                        ...current,
                        keyword: event.target.value,
                      }))
                    }
                    placeholder="名前またはメールアドレス"
                  />
                </label>

                <div className={styles.twoColumn}>
                  <label className={styles.fieldLabel}>
                    対象月From
                    <input
                      className={styles.input}
                      type="month"
                      value={searchForm.targetMonthFrom}
                      onChange={(event) =>
                        setSearchForm((current) => ({
                          ...current,
                          targetMonthFrom: event.target.value,
                        }))
                      }
                    />
                  </label>

                  <label className={styles.fieldLabel}>
                    対象月To
                    <input
                      className={styles.input}
                      type="month"
                      value={searchForm.targetMonthTo}
                      onChange={(event) =>
                        setSearchForm((current) => ({
                          ...current,
                          targetMonthTo: event.target.value,
                        }))
                      }
                    />
                  </label>
                </div>

                <div className={styles.formActions}>
                  <Button type="submit" variant="primary" disabled={isSearching}>
                    {isSearching ? "検索中..." : "検索"}
                  </Button>
                </div>
              </form>

              <div className={styles.resultHeader}>
                <p className={styles.resultCount}>{shownCountText}</p>

                {hasMore && (
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => void handleSearch(expenses.length, true)}
                    disabled={isSearching}
                  >
                    さらに表示
                  </Button>
                )}
              </div>

              <div className={styles.tableWrap}>
                <table className={styles.table}>
                  <thead>
                    <tr>
                      <th>対象月</th>
                      <th>発生日</th>
                      <th>従業員</th>
                      <th>金額</th>
                      <th>内容</th>
                      <th>領収書</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {expenses.length === 0 ? (
                      <tr>
                        <td className={styles.emptyCell} colSpan={7}>
                          経費がありません。
                        </td>
                      </tr>
                    ) : (
                      expenses.map((expense) => (
                        <tr key={expense.id}>
                          <td>{expense.targetMonth}</td>
                          <td>{expense.expenseDate}</td>
                          <td>
                            <div className={styles.userCell}>
                              <span className={styles.userName}>{expense.userName}</span>
                              <span className={styles.userEmail}>{expense.email}</span>
                            </div>
                          </td>
                          <td className={styles.amountCell}>{formatYen(expense.amount)}</td>
                          <td>
                            <div className={styles.descriptionCell}>
                              <span>{expense.description}</span>
                              {expense.memo && <span className={styles.memoText}>{expense.memo}</span>}
                            </div>
                          </td>
                          <td>
                            {expense.hasReceiptFile ? (
                              <Button
                                type="button"
                                variant="secondary"
                                onClick={() => void handleViewReceipt(expense)}
                                disabled={isViewingReceiptId === expense.id}
                              >
                                {isViewingReceiptId === expense.id ? "取得中..." : "表示"}
                              </Button>
                            ) : (
                              <span className={styles.noReceiptText}>なし</span>
                            )}
                          </td>
                          <td>
                            <div className={styles.rowActions}>
                              <Button type="button" variant="secondary" onClick={() => handleEdit(expense)}>
                                編集
                              </Button>
                              <Button
                                type="button"
                                variant="danger"
                                onClick={() => handleOpenDeleteModal(expense)}
                                disabled={isDeletingId === expense.id}
                              >
                                {isDeletingId === expense.id ? "削除中..." : "削除"}
                              </Button>
                            </div>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </section>
          </div>
        </section>
      </div>
      {deleteTarget && (
        <div
          className={styles.modalOverlay}
          role="presentation"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) {
              handleCloseDeleteModal();
            }
          }}
        >
          <section
            className={styles.deleteModal}
            role="dialog"
            aria-modal="true"
            aria-labelledby="delete-expense-title"
          >
            <div className={styles.deleteModalIcon} aria-hidden="true">
              !
            </div>

            <h2 id="delete-expense-title" className={styles.deleteModalTitle}>
              経費を削除しますか？
            </h2>

            <p className={styles.deleteModalDescription}>
              この操作は取り消せません。内容を確認してから削除してください。
            </p>

            <dl className={styles.deleteModalDetails}>
              <div className={styles.deleteModalDetailRow}>
                <dt>対象者</dt>
                <dd>{deleteTarget.userName}</dd>
              </div>
              <div className={styles.deleteModalDetailRow}>
                <dt>発生日</dt>
                <dd>{deleteTarget.expenseDate}</dd>
              </div>
              <div className={styles.deleteModalDetailRow}>
                <dt>金額</dt>
                <dd>{formatYen(deleteTarget.amount)}</dd>
              </div>
              <div className={styles.deleteModalDetailRow}>
                <dt>内容</dt>
                <dd>{deleteTarget.description}</dd>
              </div>
            </dl>

            <div className={styles.deleteModalActions}>
              <Button
                type="button"
                variant="secondary"
                onClick={handleCloseDeleteModal}
                disabled={isDeletingId !== null}
              >
                キャンセル
              </Button>

              <Button
                type="button"
                variant="danger"
                onClick={() => void handleDelete()}
                disabled={isDeletingId !== null}
              >
                {isDeletingId !== null ? "削除中..." : "削除する"}
              </Button>
            </div>
          </section>
        </div>
      )}
    </PageContainer>
  );
}

function validateExpenseForm(form: ExpenseFormState, selectedTargetUser: TargetUserCandidate | null) {
  if (!selectedTargetUser) {
    return "対象従業員を検索して選択してください。";
  }

  if (!form.targetMonth) {
    return "対象月を入力してください。";
  }

  if (!form.expenseDate) {
    return "経費発生日を入力してください。";
  }

  if (!form.amount || Number(form.amount) <= 0) {
    return "金額は1円以上で入力してください。";
  }

  if (form.description.trim() === "") {
    return "内容を入力してください。";
  }

  return null;
}

function normalizeNullableText(value: string) {
  const trimmedValue = value.trim();

  if (trimmedValue === "") {
    return null;
  }

  return trimmedValue;
}

async function searchTargetUsers(request: {
  keyword: string;
  includeDeleted: boolean;
  offset: number;
  limit: number;
}) {
  const response = await fetch(buildApiUrl("/admin/users/search"), {
    method: "POST",
    headers: {
      ...buildAuthHeaders(),
      "Content-Type": "application/json",
    },
    body: JSON.stringify(request),
  });

  const payload = (await response.json()) as ApiResponse<SearchUsersResponse>;

  if (!response.ok || payload.error) {
    throw new Error(payload.message || "従業員検索に失敗しました。");
  }

  if (!payload.data) {
    throw new Error("従業員検索の取得結果が空です。");
  }

  return payload.data;
}

function formatYen(value: number) {
  return new Intl.NumberFormat("ja-JP", {
    style: "currency",
    currency: "JPY",
    maximumFractionDigits: 0,
  }).format(value);
}

function getTodayText() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const date = String(now.getDate()).padStart(2, "0");

  return `${year}-${month}-${date}`;
}

function getCurrentMonthText() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");

  return `${year}-${month}`;
}

function buildApiUrl(path: string) {
  const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
  const normalizedBaseUrl = baseUrl.endsWith("/") ? baseUrl.slice(0, -1) : baseUrl;
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  return `${normalizedBaseUrl}${normalizedPath}`;
}

function buildAuthHeaders(): HeadersInit {
  const token = getAccessToken();

  if (!token) {
    return {};
  }

  return {
    Authorization: `Bearer ${token}`,
  };
}

function getAccessToken() {
  if (typeof window === "undefined") {
    return null;
  }

  return window.localStorage.getItem("accessToken");
}
