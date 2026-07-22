"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { downloadMonthlyAttendanceSummaryExport } from "@/api/admin/monthlyAttendanceSummaryExport";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import { useRequireRole } from "@/hooks/useRequireRole";
import styles from "./page.module.css";

type PageMessage = {
  variant: "info" | "success" | "warning" | "error";
  text: string;
};

type ExportFormat = "CSV" | "XLSX";
type ExportTargetType = "USER" | "DEPARTMENT";

type BusinessTargetUser = {
  id: number;
  name: string;
  email: string;
  departmentId?: number | null;
};

type Department = {
  id: number;
  name: string;
};


type SearchBusinessTargetUsersResponse = {
  users: BusinessTargetUser[];
  total?: number;
  offset?: number;
  limit?: number;
  hasMore?: boolean;
};

type SearchDepartmentsResponse = {
  departments: Department[];
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

type ExportFormState = {
  targetMonth: string;
  targetType: ExportTargetType;

  userKeyword: string;
  selectedUserId: number | null;

  selectedDepartmentIds: number[];
  includeUnassignedDepartment: boolean;

  includeNotApproved: boolean;
};

const initialExportForm: ExportFormState = {
  targetMonth: getCurrentMonthText(),
  targetType: "USER",

  userKeyword: "",
  selectedUserId: null,

  selectedDepartmentIds: [],
  includeUnassignedDepartment: false,

  includeNotApproved: true,
};

export default function AdminMonthlyAttendanceSummaryExportsPage() {
  const { user, isLoading, message: authMessage } = useRequireRole("ADMIN");

  const [exportForm, setExportForm] =
    useState<ExportFormState>(initialExportForm);

  const [businessTargetUsers, setBusinessTargetUsers] = useState<
    BusinessTargetUser[]
  >([]);

  const [departments, setDepartments] = useState<Department[]>([]);

  const [isSearchingUsers, setIsSearchingUsers] = useState(false);
  const [isLoadingDepartments, setIsLoadingDepartments] = useState(false);
  const [isExporting, setIsExporting] = useState(false);

  const [pageMessage, setPageMessage] = useState<PageMessage>({
    variant: "info",
    text: "対象月と出力対象を選択して、月次勤怠集計をCSVまたはExcelで出力できます。",
  });

  useEffect(() => {
    if (!user) {
      return;
    }

    void loadDepartments();
  }, [user]);

  const targetMonthText = useMemo(() => {
    if (!exportForm.targetMonth) {
      return "未選択";
    }

    const [year, month] = exportForm.targetMonth.split("-");
    return `${year}年${Number(month)}月`;
  }, [exportForm.targetMonth]);

  const selectedUser = useMemo(
    () =>
      businessTargetUsers.find(
        (targetUser) => targetUser.id === exportForm.selectedUserId,
      ) ?? null,
    [businessTargetUsers, exportForm.selectedUserId],
  );

  const selectedDepartmentNames = useMemo(() => {
    const selectedNames = departments
      .filter((department) =>
        exportForm.selectedDepartmentIds.includes(department.id),
      )
      .map((department) => department.name);

    if (exportForm.includeUnassignedDepartment) {
      selectedNames.push("所属なし");
    }

    return selectedNames;
  }, [
    departments,
    exportForm.selectedDepartmentIds,
    exportForm.includeUnassignedDepartment,
  ]);

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle
            title="月次勤怠集計出力"
            description="ログイン情報を確認しています。"
          />
          <MessageBox variant="info">{authMessage}</MessageBox>
        </section>
      </PageContainer>
    );
  }

  async function loadDepartments() {
    setIsLoadingDepartments(true);

    try {
      /*
       * 既存の共通API関数がある場合は、このfetch部分を
       * searchDepartments(...) の呼び出しに置き換えてよい。
       */
      const response = await fetch(
        buildApiUrl("/admin/departments/search"),
        {
          method: "POST",
          headers: {
            ...buildAuthHeaders(),
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            keyword: "",
            includeDeleted: false,
            offset: 0,
            limit: 50,
          }),
        },
      );

      const payload =
        (await response.json()) as ApiResponse<SearchDepartmentsResponse>;

      if (!response.ok || payload.error) {
        throw new Error(
          payload.message || "所属一覧の取得に失敗しました。",
        );
      }

      if (!payload.data) {
        throw new Error("所属一覧の取得結果が空です。");
      }

      setDepartments(payload.data.departments);
    } catch (error) {
      setPageMessage({
        variant: "error",
        text:
          error instanceof Error
            ? error.message
            : "所属一覧の取得に失敗しました。",
      });
    } finally {
      setIsLoadingDepartments(false);
    }
  }

  async function handleSearchUsers() {
    const keyword = exportForm.userKeyword.trim();

    setIsSearchingUsers(true);
    setPageMessage({
      variant: "info",
      text: "ユーザーを検索しています。",
    });

    try {
      /*
       * 既存の共通API関数がある場合は、このfetch部分を
       * searchBusinessTargetUsers(...) の呼び出しに置き換えてよい。
       */
      const response = await fetch(
        buildApiUrl("/admin/users/search-business-targets"),
        {
          method: "POST",
          headers: {
            ...buildAuthHeaders(),
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            keyword,
            offset: 0,
            limit: 50,
          }),
        },
      );

      const payload =
        (await response.json()) as ApiResponse<SearchBusinessTargetUsersResponse>;

      if (!response.ok || payload.error) {
        throw new Error(
          payload.message || "ユーザー検索に失敗しました。",
        );
      }

      if (!payload.data) {
        throw new Error("ユーザー検索の取得結果が空です。");
      }

      const users = payload.data.users;

      setBusinessTargetUsers(users);
      setExportForm((current) => ({
        ...current,
        selectedUserId: null,
      }));

      setPageMessage({
        variant: "success",
        text:
          users.length > 0
            ? `${users.length}件のユーザーが見つかりました。`
            : "条件に一致するユーザーは見つかりませんでした。",
      });
    } catch (error) {
      setPageMessage({
        variant: "error",
        text:
          error instanceof Error
            ? error.message
            : "ユーザー検索に失敗しました。",
      });
    } finally {
      setIsSearchingUsers(false);
    }
  }

  async function handleSubmit(
    event: FormEvent<HTMLFormElement>,
  ) {
    event.preventDefault();
    await handleExport("XLSX");
  }

  async function handleExport(format: ExportFormat) {
    const validationMessage = validateExportForm(exportForm);

    if (validationMessage) {
      setPageMessage({
        variant: "warning",
        text: validationMessage,
      });
      return;
    }

    const [targetYearText, targetMonthTextValue] =
      exportForm.targetMonth.split("-");

    const targetYear = Number(targetYearText);
    const targetMonth = Number(targetMonthTextValue);
    const formatLabel = format === "XLSX" ? "Excel" : "CSV";

    setIsExporting(true);
    setPageMessage({
      variant: "info",
      text: `月次勤怠集計${formatLabel}を出力しています。`,
    });

    try {
      await downloadMonthlyAttendanceSummaryExport({
        targetYear,
        targetMonth,

        targetType: exportForm.targetType,

        targetUserId:
          exportForm.targetType === "USER"
            ? exportForm.selectedUserId
            : null,

        departmentIds:
          exportForm.targetType === "DEPARTMENT"
            ? exportForm.selectedDepartmentIds
            : [],

        includeUnassignedDepartment:
          exportForm.targetType === "DEPARTMENT"
            ? exportForm.includeUnassignedDepartment
            : false,

        includeNotApproved:
          exportForm.includeNotApproved,

        format,
      });

      setPageMessage({
        variant: "success",
        text: `月次勤怠集計${formatLabel}を出力しました。`,
      });
    } catch (error) {
      setPageMessage({
        variant: "error",
        text:
          error instanceof Error
            ? error.message
            : `月次勤怠集計${formatLabel}の出力に失敗しました。`,
      });
    } finally {
      setIsExporting(false);
    }
  }

  function handleTargetTypeChange(
    targetType: ExportTargetType,
  ) {
    setExportForm((current) => ({
      ...current,
      targetType,

      selectedUserId:
        targetType === "USER"
          ? current.selectedUserId
          : null,

      selectedDepartmentIds:
        targetType === "DEPARTMENT"
          ? current.selectedDepartmentIds
          : [],

      includeUnassignedDepartment:
        targetType === "DEPARTMENT"
          ? current.includeUnassignedDepartment
          : false,
    }));
  }

  function handleDepartmentToggle(
    departmentId: number,
  ) {
    setExportForm((current) => {
      const selected =
        current.selectedDepartmentIds.includes(departmentId);

      return {
        ...current,
        selectedDepartmentIds: selected
          ? current.selectedDepartmentIds.filter(
              (id) => id !== departmentId,
            )
          : [
              ...current.selectedDepartmentIds,
              departmentId,
            ],
      };
    });
  }

  function handleReset() {
    setExportForm(initialExportForm);
    setBusinessTargetUsers([]);

    setPageMessage({
      variant: "info",
      text: "出力条件を初期状態に戻しました。",
    });
  }

  return (
    <PageContainer>
      <AdminSideMenu />

      <div className={styles.pageWrap}>
        <section className={styles.pageCard}>
          <div className={styles.headerArea}>
            <PageTitle
              title="月次勤怠集計出力"
              description="ユーザー単体、または複数所属単位で月次勤怠集計を出力します。"
            />

            <MessageBox variant={pageMessage.variant}>
              {pageMessage.text}
            </MessageBox>
          </div>

          <div className={styles.contentGrid}>
            <section className={styles.formCard}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>
                    出力条件
                  </h2>

                  <p className={styles.sectionDescription}>
                    ユーザー単体、または複数所属を選択して出力できます。
                  </p>
                </div>
              </div>

              <form
                className={styles.exportForm}
                onSubmit={handleSubmit}
              >
                <label className={styles.fieldLabel}>
                  対象月

                  <span className={styles.monthPicker}>
                    <span className={styles.monthPickerValue}>
                      {formatMonthPickerLabel(
                        exportForm.targetMonth,
                      )}
                    </span>

                    <span
                      className={styles.monthPickerIcon}
                      aria-hidden="true"
                    >
                      ▾
                    </span>

                    <input
                      className={styles.monthPickerInput}
                      type="month"
                      value={exportForm.targetMonth}
                      onChange={(event) =>
                        setExportForm((current) => ({
                          ...current,
                          targetMonth:
                            event.target.value,
                        }))
                      }
                      aria-label="対象月を選択"
                    />
                  </span>
                </label>

                <div className={styles.targetTypeArea}>
                  <span className={styles.fieldCaption}>
                    出力対象
                  </span>

                  <div className={styles.targetTypeButtons}>
                    <button
                      type="button"
                      className={
                        exportForm.targetType === "USER"
                          ? styles.targetTypeButtonActive
                          : styles.targetTypeButton
                      }
                      onClick={() =>
                        handleTargetTypeChange("USER")
                      }
                    >
                      ユーザー単体
                    </button>

                    <button
                      type="button"
                      className={
                        exportForm.targetType ===
                        "DEPARTMENT"
                          ? styles.targetTypeButtonActive
                          : styles.targetTypeButton
                      }
                      onClick={() =>
                        handleTargetTypeChange(
                          "DEPARTMENT",
                        )
                      }
                    >
                      所属単位
                    </button>
                  </div>
                </div>

                {exportForm.targetType === "USER" && (
                  <div className={styles.selectionCard}>
                    <label className={styles.fieldLabel}>
                      ユーザー検索

                      <div className={styles.searchRow}>
                        <input
                          className={styles.input}
                          type="text"
                          value={
                            exportForm.userKeyword
                          }
                          onChange={(event) =>
                            setExportForm(
                              (current) => ({
                                ...current,
                                userKeyword:
                                  event.target.value,
                              }),
                            )
                          }
                          placeholder="名前またはメールアドレス"
                        />

                        <Button
                          type="button"
                          variant="secondary"
                          disabled={isSearchingUsers}
                          onClick={() =>
                            void handleSearchUsers()
                          }
                        >
                          {isSearchingUsers
                            ? "検索中..."
                            : "検索"}
                        </Button>
                      </div>
                    </label>

                    <div className={styles.selectionList}>
                      {businessTargetUsers.map(
                        (targetUser) => (
                          <label
                            key={targetUser.id}
                            className={
                              exportForm.selectedUserId ===
                              targetUser.id
                                ? styles.selectionItemActive
                                : styles.selectionItem
                            }
                          >
                            <input
                              type="radio"
                              name="selectedUser"
                              checked={
                                exportForm.selectedUserId ===
                                targetUser.id
                              }
                              onChange={() =>
                                setExportForm(
                                  (current) => ({
                                    ...current,
                                    selectedUserId:
                                      targetUser.id,
                                  }),
                                )
                              }
                            />

                            <span>
                              <strong>
                                {targetUser.name}
                              </strong>
                              <small>
                                {targetUser.email}
                              </small>
                            </span>
                          </label>
                        ),
                      )}
                    </div>
                  </div>
                )}

                {exportForm.targetType ===
                  "DEPARTMENT" && (
                  <div className={styles.selectionCard}>
                    <span className={styles.fieldCaption}>
                      所属を複数選択
                    </span>

                    {isLoadingDepartments ? (
                      <p className={styles.emptyText}>
                        所属一覧を読み込んでいます。
                      </p>
                    ) : (
                      <div
                        className={styles.selectionList}
                      >
                        {departments.map(
                          (department) => (
                            <label
                              key={department.id}
                              className={
                                exportForm.selectedDepartmentIds.includes(
                                  department.id,
                                )
                                  ? styles.selectionItemActive
                                  : styles.selectionItem
                              }
                            >
                              <input
                                type="checkbox"
                                checked={exportForm.selectedDepartmentIds.includes(
                                  department.id,
                                )}
                                onChange={() =>
                                  handleDepartmentToggle(
                                    department.id,
                                  )
                                }
                              />

                              <span>
                                <strong>
                                  {department.name}
                                </strong>
                              </span>
                            </label>
                          ),
                        )}

                        <label
                          className={
                            exportForm.includeUnassignedDepartment
                              ? styles.selectionItemActive
                              : styles.selectionItem
                          }
                        >
                          <input
                            type="checkbox"
                            checked={
                              exportForm.includeUnassignedDepartment
                            }
                            onChange={(event) =>
                              setExportForm(
                                (current) => ({
                                  ...current,
                                  includeUnassignedDepartment:
                                    event.target
                                      .checked,
                                }),
                              )
                            }
                          />

                          <span>
                            <strong>所属なし</strong>
                          </span>
                        </label>
                      </div>
                    )}
                  </div>
                )}

                <label className={styles.checkboxLabel}>
                  <input
                    type="checkbox"
                    checked={
                      exportForm.includeNotApproved
                    }
                    onChange={(event) =>
                      setExportForm((current) => ({
                        ...current,
                        includeNotApproved:
                          event.target.checked,
                      }))
                    }
                  />

                  <span>
                    未承認・未申請の従業員もステータスのみ出力に含める
                  </span>
                </label>

                <div className={styles.formActions}>
                  <Button
                    type="submit"
                    variant="primary"
                    disabled={isExporting}
                  >
                    {isExporting
                      ? "出力中..."
                      : "Excel出力"}
                  </Button>

                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() =>
                      void handleExport("CSV")
                    }
                    disabled={isExporting}
                  >
                    CSV出力
                  </Button>

                  <Button
                    type="button"
                    variant="secondary"
                    onClick={handleReset}
                    disabled={isExporting}
                  >
                    条件をクリア
                  </Button>
                </div>
              </form>
            </section>

            <section className={styles.summaryCard}>
              <div className={styles.sectionHeader}>
                <div>
                  <h2 className={styles.sectionTitle}>
                    出力内容
                  </h2>

                  <p className={styles.sectionDescription}>
                    現在選択している出力条件です。
                  </p>
                </div>
              </div>

              <div className={styles.summaryList}>
                <div className={styles.summaryItem}>
                  <span className={styles.summaryLabel}>
                    対象月
                  </span>
                  <span className={styles.summaryValue}>
                    {targetMonthText}
                  </span>
                </div>

                <div className={styles.summaryItem}>
                  <span className={styles.summaryLabel}>
                    出力単位
                  </span>
                  <span className={styles.summaryValue}>
                    {exportForm.targetType === "USER"
                      ? "ユーザー単体"
                      : "所属単位"}
                  </span>
                </div>

                <div className={styles.summaryItem}>
                  <span className={styles.summaryLabel}>
                    出力対象
                  </span>
                  <span className={styles.summaryValue}>
                    {exportForm.targetType === "USER"
                      ? selectedUser
                        ? `${selectedUser.name}（${selectedUser.email}）`
                        : "未選択"
                      : selectedDepartmentNames.length >
                          0
                        ? selectedDepartmentNames.join(
                            "、",
                          )
                        : "未選択"}
                  </span>
                </div>

                <div className={styles.summaryItem}>
                  <span className={styles.summaryLabel}>
                    未承認者
                  </span>
                  <span className={styles.summaryValue}>
                    {exportForm.includeNotApproved
                      ? "ステータスのみ含める"
                      : "含めない"}
                  </span>
                </div>
              </div>

              <div className={styles.noticeBox}>
                <h3 className={styles.noticeTitle}>
                  集計ルール
                </h3>

                <ul className={styles.noticeList}>
                  <li>
                    ユーザー単体では1人だけ選択して出力します。
                  </li>
                  <li>
                    所属単位では複数所属と所属なしを組み合わせて出力できます。
                  </li>
                  <li>
                    承認済み以外は集計値を出力しません。
                  </li>
                  <li>
                    残業は日別超過と週超過を重複しないように集計します。
                  </li>
                  <li>
                    深夜労働は22:00〜翌5:00を休憩除外で集計します。
                  </li>
                </ul>
              </div>
            </section>
          </div>
        </section>
      </div>
    </PageContainer>
  );
}

function validateExportForm(
  form: ExportFormState,
) {
  if (!form.targetMonth) {
    return "対象月を選択してください。";
  }

  const [yearText, monthText] =
    form.targetMonth.split("-");

  const year = Number(yearText);
  const month = Number(monthText);

  if (!year || !month || month < 1 || month > 12) {
    return "対象月の形式が正しくありません。";
  }

  if (
    form.targetType === "USER" &&
    !form.selectedUserId
  ) {
    return "出力対象のユーザーを選択してください。";
  }

  if (
    form.targetType === "DEPARTMENT" &&
    form.selectedDepartmentIds.length === 0 &&
    !form.includeUnassignedDepartment
  ) {
    return "出力対象の所属を1つ以上選択してください。";
  }

  return null;
}

function getCurrentMonthText() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(
    now.getMonth() + 1,
  ).padStart(2, "0");

  return `${year}-${month}`;
}

function formatMonthPickerLabel(
  value: string,
) {
  const [yearText, monthText] =
    value.split("-");

  const year = Number(yearText);
  const month = Number(monthText);

  if (!year || !month) {
    return "月を選択";
  }

  return `${year}年${month}月`;
}

function buildApiUrl(path: string) {
  const baseUrl =
    process.env.NEXT_PUBLIC_API_BASE_URL ??
    "http://localhost:8080";

  const normalizedBaseUrl = baseUrl.endsWith("/")
    ? baseUrl.slice(0, -1)
    : baseUrl;

  const normalizedPath = path.startsWith("/")
    ? path
    : `/${path}`;

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

