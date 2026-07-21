"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import UserSideMenu from "@/components/sideMenu/UserSideMenu";
import AttendanceMonthHeader from "@/components/attendance/monthHeader/AttendanceMonthHeader";
import AttendanceTable from "@/components/attendance/table/AttendanceTable";
import MonthlyCommuterPassForm from "@/components/attendance/monthlyCommuterPassForm/MonthlyCommuterPassForm";
import { useRequireRole } from "@/hooks/useRequireRole";
import { searchAttendanceTypes } from "@/api/user/attendanceType";
import { searchAttendanceDays } from "@/api/user/attendanceDay";
import { searchAttendanceTransportExpenses } from "@/api/user/attendanceTransportExpense";
import { searchAttendanceBreaks } from "@/api/user/attendanceBreak";
import { searchHolidayDates } from "@/api/user/holidayDate";
import { searchMonthlyCommuterPass } from "@/api/user/monthlyCommuterPass";
import { updateMonthlyAttendanceSave } from "@/api/user/monthlyAttendanceSave";
import {
  searchMonthlyAttendanceRequest,
  submitMonthlyAttendanceRequest,
  withdrawMonthlyAttendanceRequest,
} from "@/api/user/monthlyAttendanceRequest";
import { getPaidLeaveBalance } from "@/api/user/paidLeave";
import type { AttendanceType } from "@/types/user/attendanceType";
import type { AttendanceBreak } from "@/types/user/attendanceBreak";
import type { MonthlyAttendanceRequest } from "@/types/user/monthlyAttendanceRequest";
import type { PaidLeaveBalanceResponse } from "@/types/user/paidLeave";
import type {
  AttendanceBreakViewRow,
  AttendanceTransportExpenseViewRow,
  AttendanceViewRow,
  CommuterPassViewForm,
  PageMessageVariant,
} from "@/types/user/attendanceView";
import {
  buildTargetMonth,
  getCurrentMonth,
  parseTargetMonth,
} from "@/utils/attendance/attendanceDate";
import { getStatusLabel } from "@/utils/attendance/attendanceStatus";
import {
  attachBreaksToAttendanceViewRows,
  attachTransportExpensesToAttendanceViewRows,
  buildAttendanceViewRows,
  buildCommuterPassViewForms,
  buildNewAttendanceBreakViewRow,
  buildNewAttendanceTransportExpenseViewRow,
  buildUpdateMonthlyAttendanceSaveRequest,
  resetAttendanceViewRow,
  buildNewCommuterPassViewForm,
  resetCommuterPassViewForms,
} from "@/utils/attendance/userAttendance/userAttendanceMapper";
import {
  isUserAttendanceRowLocked,
  isUserMonthlyCommuterPassLocked,
  isUserMonthlySubmitDisabled,
  isUserMonthlyWithdrawDisabled,
} from "@/utils/attendance/userAttendance/userAttendancePermission";
import styles from "./page.module.css";

function formatPaidLeaveDays(value: number | null | undefined) {
  if (value === null || value === undefined) {
    return "-";
  }

  return value.toFixed(1).replace(".0", "");
}

function isPaidLeaveAttendanceType(attendanceType: AttendanceType | undefined) {
  return attendanceType?.code === "PAID_LEAVE" || attendanceType?.name === "有給";
}

export default function UserAttendancePage() {
  const { user, isLoading, message } = useRequireRole("USER");

  const [targetMonth, setTargetMonth] = useState(getCurrentMonth());
  const [pendingTargetMonth, setPendingTargetMonth] = useState<string | null>(null);
  const [attendanceTypes, setAttendanceTypes] = useState<AttendanceType[]>([]);
  const [attendanceRows, setAttendanceRows] = useState<AttendanceViewRow[]>([]);
  const [commuterPasses, setCommuterPasses] = useState<CommuterPassViewForm[]>([]);
  const [monthlyAttendanceRequest, setMonthlyAttendanceRequest] =
    useState<MonthlyAttendanceRequest | null>(null);
  const [paidLeaveBalance, setPaidLeaveBalance] = useState<PaidLeaveBalanceResponse | null>(null);
  const [pageMessage, setPageMessage] = useState("対象月の勤怠を入力できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");
  const [isPageLoading, setIsPageLoading] = useState(false);

  const { targetYear, targetMonthValue } = useMemo(
    () => parseTargetMonth(targetMonth),
    [targetMonth],
  );

  const monthlyStatus = monthlyAttendanceRequest?.status ?? "DRAFT";

  const hasUnsavedChanges = useMemo(() => {
    return (
      commuterPasses.some((commuterPass) => commuterPass.isDirty) ||
      attendanceRows.some(
        (row) =>
          row.isDirty ||
          row.breaks.some((breakRow) => breakRow.isDirty) ||
          row.transportExpenses.some((transportExpense) => transportExpense.isDirty),
      )
    );
  }, [attendanceRows, commuterPasses]);

  const hasNoPaidLeaveBalance = paidLeaveBalance === null || paidLeaveBalance.remainingDays <= 0;

  const loadPageData = useCallback(async () => {
    if (!user) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("勤怠情報を取得しています。");
    setPageMessageVariant("info");

    try {
      const paidLeaveBalanceResult = await getPaidLeaveBalance();

      if (paidLeaveBalanceResult.error || !paidLeaveBalanceResult.data) {
        setPaidLeaveBalance(null);
        setPageMessage(paidLeaveBalanceResult.message || "有給残数の取得に失敗しました。");
        setPageMessageVariant("error");
        return;
      }

      const nextPaidLeaveBalance = paidLeaveBalanceResult.data;

      const attendanceTypesResult = await searchAttendanceTypes({});

      if (attendanceTypesResult.error || !attendanceTypesResult.data) {
        setPageMessage(attendanceTypesResult.message || "勤務区分マスタの取得に失敗しました。");
        setPageMessageVariant("error");
        return;
      }

      const nextAttendanceTypes = attendanceTypesResult.data.attendanceTypes;

      const attendanceDaysResult = await searchAttendanceDays({
        targetYear,
        targetMonth: targetMonthValue,
      });

      if (attendanceDaysResult.error || !attendanceDaysResult.data) {
        setPageMessage(attendanceDaysResult.message || "勤怠一覧の取得に失敗しました。");
        setPageMessageVariant("error");
        return;
      }

      const nextAttendanceDays = attendanceDaysResult.data.attendanceDays;

      const attendanceTransportExpensesResult =
        await searchAttendanceTransportExpenses({
          targetYear,
          targetMonth: targetMonthValue,
        });

      if (
        attendanceTransportExpensesResult.error ||
        !attendanceTransportExpensesResult.data
      ) {
        setPageMessage(
          attendanceTransportExpensesResult.message ||
            "日別交通費一覧の取得に失敗しました。",
        );
        setPageMessageVariant("error");
        return;
      }

      const nextAttendanceTransportExpenses =
        attendanceTransportExpensesResult.data.attendanceTransportExpenses;

      const holidayDatesResult = await searchHolidayDates({
        targetYear,
        targetMonth: targetMonthValue,
      });

      if (holidayDatesResult.error || !holidayDatesResult.data) {
        setPageMessage(holidayDatesResult.message || "祝日一覧の取得に失敗しました。");
        setPageMessageVariant("error");
        return;
      }

      const nextHolidayDates = holidayDatesResult.data.holidays;

      const commuterPassResult = await searchMonthlyCommuterPass({
        targetYear,
        targetMonth: targetMonthValue,
      });

      if (commuterPassResult.error || !commuterPassResult.data) {
        setPageMessage(commuterPassResult.message || "月次通勤定期の取得に失敗しました。");
        setPageMessageVariant("error");
        return;
      }

      const nextMonthlyCommuterPasses =
        commuterPassResult.data.monthlyCommuterPasses;

      const monthlyAttendanceRequestResult = await searchMonthlyAttendanceRequest({
        targetYear,
        targetMonth: targetMonthValue,
      });

      if (monthlyAttendanceRequestResult.error || !monthlyAttendanceRequestResult.data) {
        setMonthlyAttendanceRequest(null);
        setPageMessage(
          monthlyAttendanceRequestResult.message || "月次申請状態の取得に失敗しました。",
        );
        setPageMessageVariant("error");
        return;
      }

      const nextMonthlyAttendanceRequest =
        monthlyAttendanceRequestResult.data.monthlyAttendanceRequest;

      const rows = buildAttendanceViewRows(
        targetYear,
        targetMonthValue,
        nextAttendanceDays,
        nextHolidayDates,
      );
      const breakMap = new Map<string, AttendanceBreak[]>();

      /*
       * 休憩検索APIは1日単位。
       * ただし、1日分の休憩取得で失敗してもページ全体を止めない。
       * 失敗した日は休憩なしとして表示し、画面上に警告を出す。
       */
      let hasBreakLoadError = false;

      await Promise.all(
        rows.map(async (row) => {
          if (row.attendanceDayId === null) {
            breakMap.set(row.workDate, []);
            return;
          }

          const result = await searchAttendanceBreaks({
            workDate: row.workDate,
          });

          if (result.error || !result.data) {
            hasBreakLoadError = true;
            breakMap.set(row.workDate, []);
            return;
          }

          breakMap.set(row.workDate, result.data.attendanceBreaks);
        }),
      );

      setPaidLeaveBalance(nextPaidLeaveBalance);
      setAttendanceTypes(nextAttendanceTypes);
      const rowsWithTransportExpenses =
        attachTransportExpensesToAttendanceViewRows(
          rows,
          nextAttendanceTransportExpenses,
        );

      setAttendanceRows(
        attachBreaksToAttendanceViewRows(rowsWithTransportExpenses, breakMap),
      );
      setCommuterPasses(
        buildCommuterPassViewForms(nextMonthlyCommuterPasses),
      );
      setMonthlyAttendanceRequest(nextMonthlyAttendanceRequest);

      if (hasBreakLoadError) {
        setPageMessage("勤怠情報を取得しました。一部の日付の休憩取得に失敗しました。");
        setPageMessageVariant("warning");
        return;
      }

      setPageMessage("対象月の勤怠を入力できます。");
      setPageMessageVariant("info");
    } catch (error) {
      setPageMessage(
        error instanceof Error
          ? error.message
          : "勤怠情報の取得中に予期しないエラーが発生しました。",
      );
      setPageMessageVariant("error");
    } finally {
      setIsPageLoading(false);
    }
  }, [targetMonthValue, targetYear, user]);

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void loadPageData();
    }, 0);

    return () => {
      window.clearTimeout(timerId);
    };
  }, [isLoading, loadPageData, user]);

  useEffect(() => {
    const handleBeforeUnload = (event: BeforeUnloadEvent) => {
      if (!hasUnsavedChanges) {
        return;
      }

      event.preventDefault();
      event.returnValue = "";
    };

    window.addEventListener("beforeunload", handleBeforeUnload);

    return () => {
      window.removeEventListener("beforeunload", handleBeforeUnload);
    };
  }, [hasUnsavedChanges]);

  const updateRow = <K extends keyof AttendanceViewRow>(
    workDate: string,
    key: K,
    value: AttendanceViewRow[K],
  ) => {
    if (key === "planAttendanceTypeId" && typeof value === "number") {
      const nextAttendanceType = attendanceTypes.find(
        (attendanceType) => attendanceType.id === value,
      );

      if (isPaidLeaveAttendanceType(nextAttendanceType) && hasNoPaidLeaveBalance) {
        setPageMessage(
          "有給残数が取得できていない、または残数が0日のため、有給は選択できません。",
        );
        setPageMessageVariant("error");
        return;
      }
    }

    setAttendanceRows((currentRows) =>
      currentRows.map((row) =>
        row.workDate === workDate
          ? {
              ...row,
              [key]: value,
              isDirty: true,
            }
          : row,
      ),
    );
  };

  const updateCommuterPassForm = <K extends keyof CommuterPassViewForm>(
    index: number,
    key: K,
    value: CommuterPassViewForm[K],
  ) => {
    setCommuterPasses((current) =>
      current.map((commuterPass, currentIndex) =>
        currentIndex === index
          ? {
              ...commuterPass,
              [key]: value,
              isDirty: true,
            }
          : commuterPass,
      ),
    );
  };

  const handleAddCommuterPass = () => {
    setCommuterPasses((current) => [
      ...current,
      buildNewCommuterPassViewForm(),
    ]);
    setPageMessage("月次通勤定期の入力欄を追加しました。");
    setPageMessageVariant("info");
  };

  const handleRemoveCommuterPass = (index: number) => {
    setCommuterPasses((current) =>
      current
        .filter((_, currentIndex) => currentIndex !== index)
        .map((commuterPass) => ({
          ...commuterPass,
          isDirty: true,
        })),
    );
    setPageMessage("月次通勤定期を削除対象にしました。全体保存で反映されます。");
    setPageMessageVariant("info");
  };

  const handleResetCommuterPasses = () => {
    setCommuterPasses(resetCommuterPassViewForms());
    setPageMessage("月次通勤定期をすべて初期値に戻しました。全体保存で反映されます。");
    setPageMessageVariant("info");
  };

  const requestMonthChange = (nextTargetMonth: string) => {
    if (hasUnsavedChanges) {
      setPendingTargetMonth(nextTargetMonth);
      setPageMessage("未保存の変更があります。保存して移動するか、保存せず移動してください。");
      setPageMessageVariant("warning");
      return;
    }

    setTargetMonth(nextTargetMonth);
  };

  const handlePreviousMonth = () => {
    const previousMonthDate = new Date(targetYear, targetMonthValue - 2, 1);
    const year = previousMonthDate.getFullYear();
    const month = previousMonthDate.getMonth() + 1;

    requestMonthChange(buildTargetMonth(year, month));
  };

  const handleNextMonth = () => {
    const nextMonthDate = new Date(targetYear, targetMonthValue, 1);
    const year = nextMonthDate.getFullYear();
    const month = nextMonthDate.getMonth() + 1;

    requestMonthChange(buildTargetMonth(year, month));
  };

  const handleChangeMonth = (value: string) => {
    requestMonthChange(value);
  };

  const handleSaveAllAttendanceDays = async () => {
    const saveTargetRows = attendanceRows.filter(
      (row) =>
        row.isDirty ||
        row.breaks.some((breakRow) => breakRow.isDirty) ||
        row.transportExpenses.some((transportExpense) => transportExpense.isDirty),
    );

    const hasDirtyCommuterPass = commuterPasses.some(
      (commuterPass) => commuterPass.isDirty,
    );

    if (!hasDirtyCommuterPass && saveTargetRows.length === 0) {
      setPageMessage("保存対象の勤怠はありません。");
      setPageMessageVariant("info");
      return true;
    }

    const hasPaidLeaveSaveTarget = saveTargetRows.some((row) => {
      const selectedPlanType = attendanceTypes.find(
        (attendanceType) => attendanceType.id === row.planAttendanceTypeId,
      );

      return isPaidLeaveAttendanceType(selectedPlanType);
    });

    if (hasPaidLeaveSaveTarget && hasNoPaidLeaveBalance) {
      setPageMessage("有給残数が取得できていない、または残数が0日のため、有給を保存できません。");
      setPageMessageVariant("error");
      return false;
    }

    for (const commuterPass of commuterPasses) {
      if (
        !commuterPass.commuterFrom.trim() ||
        !commuterPass.commuterTo.trim() ||
        !commuterPass.commuterMethod.trim()
      ) {
        setPageMessage(
          "月次通勤定期の出発地・目的地・手段を入力してください。",
        );
        setPageMessageVariant("error");
        return false;
      }

      const commuterAmount = Number(commuterPass.commuterAmount);
      if (
        commuterPass.commuterAmount.trim() === "" ||
        !Number.isFinite(commuterAmount) ||
        commuterAmount < 0
      ) {
        setPageMessage("月次通勤定期の金額を正しく入力してください。");
        setPageMessageVariant("error");
        return false;
      }
    }

    for (const row of saveTargetRows) {
      if (row.planAttendanceTypeId === 0) {
        const hasInput =
          row.planStartTime !== "" ||
          row.planEndTime !== "" ||
          row.actualStartTime !== "" ||
          row.actualEndTime !== "" ||
          row.scheduledWorkMinutes !== "" ||
          row.remoteWorkAllowanceFlag ||
          row.breaks.length > 0 ||
          row.transportExpenses.length > 0;

        if (hasInput) {
          setPageMessage(
            `${row.dayLabel} は入力があります。先に予定区分を選択してください。`,
          );
          setPageMessageVariant("error");
          return false;
        }

        continue;
      }

      const selectedPlanType = attendanceTypes.find(
        (attendanceType) => attendanceType.id === row.planAttendanceTypeId,
      );

      if (!selectedPlanType) {
        setPageMessage(`${row.dayLabel} の予定区分を選択してください。`);
        setPageMessageVariant("error");
        return false;
      }

      for (const breakRow of row.breaks) {
        if (!breakRow.breakStartTime || !breakRow.breakEndTime) {
          setPageMessage(`${row.dayLabel} の休憩開始時刻と終了時刻を入力してください。`);
          setPageMessageVariant("error");
          return false;
        }
      }

      for (const transportExpense of row.transportExpenses) {
        if (
          !transportExpense.transportFrom.trim() ||
          !transportExpense.transportTo.trim() ||
          !transportExpense.transportMethod.trim()
        ) {
          setPageMessage(
            `${row.dayLabel} の交通費の出発地・目的地・手段を入力してください。`,
          );
          setPageMessageVariant("error");
          return false;
        }

        const transportAmount = Number(transportExpense.transportAmount);
        if (
          transportExpense.transportAmount.trim() === "" ||
          !Number.isFinite(transportAmount) ||
          transportAmount < 0
        ) {
          setPageMessage(`${row.dayLabel} の交通費金額を正しく入力してください。`);
          setPageMessageVariant("error");
          return false;
        }
      }
    }

    setPageMessage("月次勤怠を全体保存しています。");
    setPageMessageVariant("info");

    let request;

    try {
      request = buildUpdateMonthlyAttendanceSaveRequest(
        targetYear,
        targetMonthValue,
        commuterPasses,
        saveTargetRows,
        attendanceTypes,
      );
    } catch (error) {
      setPageMessage(
        error instanceof Error
          ? error.message
          : "月次勤怠全体保存のリクエスト作成に失敗しました。",
      );
      setPageMessageVariant("error");
      return false;
    }

    const result = await updateMonthlyAttendanceSave(request);

    if (result.error || !result.data) {
      setPageMessage(result.message || "月次勤怠の全体保存に失敗しました。");
      setPageMessageVariant("error");
      return false;
    }

    setPageMessage(result.message || "月次勤怠を全体保存しました。");
    setPageMessageVariant("success");

    await loadPageData();

    return true;
  };

  const handleSaveAllAndMove = async () => {
    if (!pendingTargetMonth) {
      return;
    }

    const isSaved = await handleSaveAllAttendanceDays();

    if (!isSaved) {
      return;
    }

    setPendingTargetMonth(null);
    setTargetMonth(pendingTargetMonth);
  };

  const handleDiscardAndMove = () => {
    if (!pendingTargetMonth) {
      return;
    }

    setPendingTargetMonth(null);
    setTargetMonth(pendingTargetMonth);
  };

  const handleResetAttendanceDay = (row: AttendanceViewRow) => {
    setAttendanceRows((currentRows) =>
      currentRows.map((currentRow) =>
        currentRow.workDate === row.workDate ? resetAttendanceViewRow(currentRow) : currentRow,
      ),
    );

    setPageMessage(`${row.dayLabel} を初期値に戻しました。全体保存で反映されます。`);
    setPageMessageVariant("info");
  };

  const handleAddBreak = (workDate: string) => {
    setAttendanceRows((currentRows) =>
      currentRows.map((row) =>
        row.workDate === workDate
          ? {
              ...row,
              breaks: [...row.breaks, buildNewAttendanceBreakViewRow()],
              isDirty: true,
            }
          : row,
      ),
    );
  };

  const handleChangeBreak = <K extends keyof AttendanceBreakViewRow>(
    workDate: string,
    breakIndex: number,
    key: K,
    value: AttendanceBreakViewRow[K],
  ) => {
    setAttendanceRows((currentRows) =>
      currentRows.map((row) => {
        if (row.workDate !== workDate) {
          return row;
        }

        return {
          ...row,
          isDirty: true,
          breaks: row.breaks.map((breakRow, currentIndex) =>
            currentIndex === breakIndex
              ? {
                  ...breakRow,
                  [key]: value,
                  isDirty: true,
                }
              : breakRow,
          ),
        };
      }),
    );
  };

  const handleDeleteBreak = (row: AttendanceViewRow, breakIndex: number) => {
    setAttendanceRows((currentRows) =>
      currentRows.map((currentRow) =>
        currentRow.workDate === row.workDate
          ? {
              ...currentRow,
              breaks: currentRow.breaks.filter((_, currentIndex) => currentIndex !== breakIndex),
              isDirty: true,
            }
          : currentRow,
      ),
    );
  };


  const handleAddTransportExpense = (workDate: string) => {
    setAttendanceRows((currentRows) =>
      currentRows.map((row) =>
        row.workDate === workDate
          ? {
              ...row,
              transportExpenses: [
                ...row.transportExpenses,
                buildNewAttendanceTransportExpenseViewRow(
                  row.transportExpenses.length + 1,
                ),
              ],
              isDirty: true,
            }
          : row,
      ),
    );
  };

  const handleChangeTransportExpense = <
    K extends keyof AttendanceTransportExpenseViewRow,
  >(
    workDate: string,
    transportExpenseIndex: number,
    key: K,
    value: AttendanceTransportExpenseViewRow[K],
  ) => {
    setAttendanceRows((currentRows) =>
      currentRows.map((row) => {
        if (row.workDate !== workDate) {
          return row;
        }

        return {
          ...row,
          isDirty: true,
          transportExpenses: row.transportExpenses.map(
            (transportExpense, currentIndex) =>
              currentIndex === transportExpenseIndex
                ? {
                    ...transportExpense,
                    [key]: value,
                    isDirty: true,
                  }
                : transportExpense,
          ),
        };
      }),
    );
  };

  const handleDeleteTransportExpense = (
    row: AttendanceViewRow,
    transportExpenseIndex: number,
  ) => {
    setAttendanceRows((currentRows) =>
      currentRows.map((currentRow) =>
        currentRow.workDate === row.workDate
          ? {
              ...currentRow,
              transportExpenses: currentRow.transportExpenses
                .filter((_, currentIndex) => currentIndex !== transportExpenseIndex)
                .map((transportExpense, currentIndex) => ({
                  ...transportExpense,
                  sortOrder: currentIndex + 1,
                  isDirty: true,
                })),
              isDirty: true,
            }
          : currentRow,
      ),
    );
  };

  const handleMonthlySubmit = async () => {
    if (hasUnsavedChanges) {
      setPageMessage("未保存の変更があります。先に全体保存してください。");
      setPageMessageVariant("warning");
      return;
    }

    setPageMessage("月次申請しています。");
    setPageMessageVariant("info");

    const result = await submitMonthlyAttendanceRequest({
      targetYear,
      targetMonth: targetMonthValue,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "月次申請に失敗しました。");
      setPageMessageVariant("error");
      return;
    }

    setMonthlyAttendanceRequest(result.data.monthlyAttendanceRequest);
    setPageMessage(result.message || "月次申請しました。");
    setPageMessageVariant("success");

    await loadPageData();
  };

  const handleMonthlyWithdraw = async () => {
    if (hasUnsavedChanges) {
      setPageMessage("未保存の変更があります。先に全体保存するか、変更を破棄してください。");
      setPageMessageVariant("warning");
      return;
    }

    setPageMessage("月次申請を取り下げています。");
    setPageMessageVariant("info");

    const result = await withdrawMonthlyAttendanceRequest({
      targetYear,
      targetMonth: targetMonthValue,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "月次申請の取り下げに失敗しました。");
      setPageMessageVariant("error");
      return;
    }

    setMonthlyAttendanceRequest(result.data.monthlyAttendanceRequest);
    setPageMessage(result.message || "月次申請を取り下げました。");
    setPageMessageVariant("success");

    await loadPageData();
  };

  if (isLoading || !user) {
    return (
      <PageContainer>
        <UserSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="勤怠入力" description="ログイン情報を確認しています。" />
          <MessageBox variant="info">{message}</MessageBox>
        </section>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <UserSideMenu />

      <div className={styles.pageWrap}>
        <section className={styles.pageCard}>
          <AttendanceMonthHeader
            targetMonth={targetMonth}
            monthlyStatus={monthlyStatus}
            monthlySubmitDisabled={isUserMonthlySubmitDisabled(
              monthlyStatus,
              hasUnsavedChanges,
            )}
            monthlyWithdrawDisabled={isUserMonthlyWithdrawDisabled(
              monthlyStatus,
              hasUnsavedChanges,
            )}
            saveDisabled={!hasUnsavedChanges}
            onChangeMonth={handleChangeMonth}
            onPreviousMonth={handlePreviousMonth}
            onNextMonth={handleNextMonth}
            onSaveAll={handleSaveAllAttendanceDays}
            onMonthlySubmit={handleMonthlySubmit}
            onMonthlyWithdraw={handleMonthlyWithdraw}
          />

          {pendingTargetMonth && (
            <div className={styles.unsavedBar}>
              <p className={styles.unsavedBarText}>
                未保存の変更があります。移動先：{pendingTargetMonth}
              </p>

              <div className={styles.unsavedBarActions}>
                <Button type="button" variant="secondary" onClick={handleSaveAllAndMove}>
                  保存して移動
                </Button>

                <Button type="button" variant="danger" onClick={handleDiscardAndMove}>
                  保存せず移動
                </Button>
              </div>
            </div>
          )}

          <div className={styles.messageArea}>
            <MessageBox variant={pageMessageVariant}>
              {isPageLoading ? "読み込み中..." : pageMessage}
            </MessageBox>

            <div className={styles.paidLeaveBalanceBox}>
              <div className={styles.paidLeaveBalanceHeader}>
                <p className={styles.paidLeaveBalanceLabel}>有給残数</p>
                <span
                  className={
                    paidLeaveBalance && paidLeaveBalance.remainingDays > 0
                      ? styles.paidLeaveStatusGood
                      : styles.paidLeaveStatusAlert
                  }
                >
                  {paidLeaveBalance && paidLeaveBalance.remainingDays > 0 ? "取得可能" : "要確認"}
                </span>
              </div>

              <div className={styles.paidLeaveBalanceMain}>
                <span className={styles.paidLeaveBalanceValue}>
                  {paidLeaveBalance ? formatPaidLeaveDays(paidLeaveBalance.remainingDays) : "-"}
                </span>
                <span className={styles.paidLeaveBalanceUnit}>日</span>
              </div>

              <div className={styles.paidLeaveBalanceSubGrid}>
                <p className={styles.paidLeaveBalanceSubText}>
                  付与{" "}
                  {paidLeaveBalance
                    ? formatPaidLeaveDays(paidLeaveBalance.totalGrantedDays)
                    : "-"}
                  日
                </p>
                <p className={styles.paidLeaveBalanceSubText}>
                  使用 {paidLeaveBalance ? formatPaidLeaveDays(paidLeaveBalance.usedDays) : "-"}日
                </p>
              </div>
            </div>

            <div className={styles.monthlyStatusBox}>
              <p className={styles.monthlyStatusLabel}>月次申請状態</p>
              <p className={styles.monthlyStatusValue}>{getStatusLabel(monthlyStatus)}</p>
            </div>
          </div>

          <MonthlyCommuterPassForm
            commuterPasses={commuterPasses}
            disabled={isUserMonthlyCommuterPassLocked(monthlyStatus)}
            onChange={updateCommuterPassForm}
            onAdd={handleAddCommuterPass}
            onRemove={handleRemoveCommuterPass}
            onReset={handleResetCommuterPasses}
          />

          <AttendanceTable
            rows={attendanceRows}
            attendanceTypes={attendanceTypes}
            getRowLocked={() => isUserAttendanceRowLocked(monthlyStatus)}
            onChangeRow={updateRow}
            onDeleteRow={handleResetAttendanceDay}
            onAddBreak={handleAddBreak}
            onChangeBreak={handleChangeBreak}
            onDeleteBreak={handleDeleteBreak}
            onAddTransportExpense={handleAddTransportExpense}
            onChangeTransportExpense={handleChangeTransportExpense}
            onDeleteTransportExpense={handleDeleteTransportExpense}
          />
        </section>
      </div>
    </PageContainer>
  );
}