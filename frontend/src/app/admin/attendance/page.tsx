"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import AdminSideMenu from "@/components/sideMenu/AdminSideMenu";
import AttendanceMonthHeader from "@/components/attendance/monthHeader/AttendanceMonthHeader";
import AttendanceTable from "@/components/attendance/table/AttendanceTable";
import MonthlyCommuterPassForm from "@/components/attendance/monthlyCommuterPassForm/MonthlyCommuterPassForm";
import AdminUserSearch from "@/components/attendance/adminUserSearch/AdminUserSearch";
import AdminApprovalActions from "@/components/attendance/adminApprovalActions/AdminApprovalActions";
import { useRequireRole } from "@/hooks/useRequireRole";
import { searchAttendanceTypes } from "@/api/admin/attendanceType";
import { searchAttendanceDays } from "@/api/admin/attendanceDay";
import { searchAttendanceBreaks } from "@/api/admin/attendanceBreak";
import { searchHolidayDates } from "@/api/admin/holidayDate";
import { searchMonthlyCommuterPass } from "@/api/admin/monthlyCommuterPass";
import { updateMonthlyAttendanceSave } from "@/api/admin/monthlyAttendanceSave";
import {
  approveMonthlyAttendanceRequest,
  rejectMonthlyAttendanceRequest,
  searchMonthlyAttendanceRequest,
  submitMonthlyAttendanceRequest,
  withdrawMonthlyAttendanceRequest,
} from "@/api/admin/monthlyAttendanceRequest";
import { getPaidLeaveBalance } from "@/api/admin/paidLeaveUsage";
import type { UserResponse } from "@/types/admin/user";
import type { AttendanceType } from "@/types/admin/attendanceType";
import type { AttendanceBreak } from "@/types/admin/attendanceBreak";
import type { MonthlyAttendanceRequest } from "@/types/admin/monthlyAttendanceRequest";
import type { PaidLeaveBalanceResponse } from "@/types/admin/paidLeaveUsage";
import type {
  AttendanceBreakViewRow,
  AttendanceViewRow,
  CommuterPassViewForm,
  PageMessageVariant,
} from "@/types/admin/attendanceView";
import {
  buildTargetMonth,
  getCurrentMonth,
  parseTargetMonth,
} from "@/utils/attendance/attendanceDate";
import { getStatusLabel } from "@/utils/attendance/attendanceStatus";
import {
  attachBreaksToAttendanceViewRows,
  buildAttendanceViewRows,
  buildCommuterPassViewForm,
  buildNewAttendanceBreakViewRow,
  buildUpdateMonthlyAttendanceSaveRequest,
  resetAttendanceViewRow,
  resetCommuterPassViewForm,
} from "@/utils/attendance/adminAttendance/adminAttendanceMapper";
import { getUserDetail } from "@/api/admin/user";
import type { AdminAttendanceInitialSearch } from "@/types/admin/adminAttendanceInitialSearch";
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

export default function AdminAttendancePage() {
  const { user, isLoading, message } = useRequireRole("ADMIN");

  const [targetMonth, setTargetMonth] = useState(getCurrentMonth());
  const [pendingTargetMonth, setPendingTargetMonth] = useState<string | null>(null);

  const [selectedUser, setSelectedUser] = useState<UserResponse | null>(null);
  const [loadedUser, setLoadedUser] = useState<UserResponse | null>(null);

  const [attendanceTypes, setAttendanceTypes] = useState<AttendanceType[]>([]);
  const [attendanceRows, setAttendanceRows] = useState<AttendanceViewRow[]>([]);
  const [commuterPass, setCommuterPass] = useState<CommuterPassViewForm>({
    commuterFrom: "",
    commuterTo: "",
    commuterMethod: "",
    commuterAmount: "",
  });

  const [monthlyAttendanceRequest, setMonthlyAttendanceRequest] =
    useState<MonthlyAttendanceRequest | null>(null);
  const [paidLeaveBalance, setPaidLeaveBalance] = useState<PaidLeaveBalanceResponse | null>(null);

  const [isCommuterPassDirty, setIsCommuterPassDirty] = useState(false);
  const [pageMessage, setPageMessage] =
    useState("対象ユーザーを検索して、勤怠を読み込んでください。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");
  const [isPageLoading, setIsPageLoading] = useState(false);

  const { targetYear, targetMonthValue } = useMemo(
    () => parseTargetMonth(targetMonth),
    [targetMonth],
  );

  const monthlyStatus = monthlyAttendanceRequest?.status ?? "NOT_SUBMITTED";

  const hasUnsavedChanges = useMemo(() => {
    return (
      isCommuterPassDirty ||
      attendanceRows.some(
        (row) => row.isDirty || row.breaks.some((breakRow) => breakRow.isDirty),
      )
    );
  }, [attendanceRows, isCommuterPassDirty]);

  const hasNoPaidLeaveBalance = paidLeaveBalance === null || paidLeaveBalance.remainingDays <= 0;

  const loadPageData = useCallback(
    async (
      targetUser: UserResponse,
      loadTargetYear = targetYear,
      loadTargetMonth = targetMonthValue,
    ) => {
      setIsPageLoading(true);
      setPageMessage("管理者用の勤怠情報を取得しています。");
      setPageMessageVariant("info");

      try {
        const paidLeaveBalanceResult = await getPaidLeaveBalance({
          targetUserId: targetUser.id,
        });

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
          targetUserId: targetUser.id,
          targetYear: loadTargetYear,
          targetMonth: loadTargetMonth,
        });

        if (attendanceDaysResult.error || !attendanceDaysResult.data) {
          setPageMessage(attendanceDaysResult.message || "勤怠一覧の取得に失敗しました。");
          setPageMessageVariant("error");
          return;
        }

        const nextAttendanceDays = attendanceDaysResult.data.attendanceDays;

        const holidayDatesResult = await searchHolidayDates({
          targetYear: loadTargetYear,
          targetMonth: loadTargetMonth,
        });

        if (holidayDatesResult.error || !holidayDatesResult.data) {
          setPageMessage(holidayDatesResult.message || "祝日一覧の取得に失敗しました。");
          setPageMessageVariant("error");
          return;
        }

        const nextHolidayDates = holidayDatesResult.data.holidays;

        const commuterPassResult = await searchMonthlyCommuterPass({
          targetUserId: targetUser.id,
          targetYear: loadTargetYear,
          targetMonth: loadTargetMonth,
        });

        if (commuterPassResult.error || !commuterPassResult.data) {
          setPageMessage(commuterPassResult.message || "月次通勤定期の取得に失敗しました。");
          setPageMessageVariant("error");
          return;
        }

        const nextMonthlyCommuterPass = commuterPassResult.data.monthlyCommuterPass;

        const monthlyAttendanceRequestResult = await searchMonthlyAttendanceRequest({
          targetUserId: targetUser.id,
          targetYear: loadTargetYear,
          targetMonth: loadTargetMonth,
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
          loadTargetYear,
          loadTargetMonth,
          nextAttendanceDays,
          nextHolidayDates,
        );

        const breakMap = new Map<string, AttendanceBreak[]>();
        let hasBreakLoadError = false;

        await Promise.all(
          rows.map(async (row) => {
            if (row.attendanceDayId === null) {
              breakMap.set(row.workDate, []);
              return;
            }

            const result = await searchAttendanceBreaks({
              targetUserId: targetUser.id,
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

        setLoadedUser(targetUser);
        setPaidLeaveBalance(nextPaidLeaveBalance);
        setAttendanceTypes(nextAttendanceTypes);
        setAttendanceRows(attachBreaksToAttendanceViewRows(rows, breakMap));
        setCommuterPass(buildCommuterPassViewForm(nextMonthlyCommuterPass));
        setMonthlyAttendanceRequest(nextMonthlyAttendanceRequest);
        setIsCommuterPassDirty(false);

        if (hasBreakLoadError) {
          setPageMessage("勤怠情報を取得しました。一部の日付の休憩取得に失敗しました。");
          setPageMessageVariant("warning");
          return;
        }

        setPageMessage("管理者として対象ユーザーの勤怠を編集できます。");
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
    },
    [targetMonthValue, targetYear],
  );

  const loadInitialSearchFromStorage = useCallback(async () => {
    const searchParams = new URLSearchParams(window.location.search);
    const initialKey = searchParams.get("initialKey");

    if (!initialKey) {
      return;
    }

    const storageKey = `adminAttendanceInitialSearch:${initialKey}`;
    const savedInitialSearch = localStorage.getItem(storageKey);

    if (!savedInitialSearch) {
      setPageMessage("申請一覧からの遷移情報が見つかりません。");
      setPageMessageVariant("error");
      return;
    }

    localStorage.removeItem(storageKey);

    let initialSearch: AdminAttendanceInitialSearch;

    try {
      initialSearch = JSON.parse(savedInitialSearch) as AdminAttendanceInitialSearch;
    } catch {
      setPageMessage("申請一覧からの遷移情報の読み取りに失敗しました。");
      setPageMessageVariant("error");
      return;
    }

    if (
      !initialSearch.targetUserId ||
      !initialSearch.targetYear ||
      !initialSearch.targetMonth
    ) {
      setPageMessage("申請一覧からの遷移情報が不足しています。");
      setPageMessageVariant("error");
      return;
    }

    setPageMessage("申請一覧で選択したユーザーの勤怠を読み込んでいます。");
    setPageMessageVariant("info");

    try {
      const result = await getUserDetail({
        targetUserId: initialSearch.targetUserId,
      });

      if (result.error || !result.data) {
        setPageMessage(result.message || "対象ユーザー情報の取得に失敗しました。");
        setPageMessageVariant("error");
        return;
      }

      const targetUser = result.data.user;
      const nextTargetMonth = buildTargetMonth(
        initialSearch.targetYear,
        initialSearch.targetMonth,
      );

      setSelectedUser(targetUser);
      setTargetMonth(nextTargetMonth);

      await loadPageData(
        targetUser,
        initialSearch.targetYear,
        initialSearch.targetMonth,
      );
    } catch (error) {
      setPageMessage(
        error instanceof Error
          ? error.message
          : "申請一覧からの勤怠読み込み中に予期しないエラーが発生しました。",
      );
      setPageMessageVariant("error");
    }
  }, [loadPageData]);

  useEffect(() => {
    if (isLoading || !user) {
      return;
    }

    const timerId = window.setTimeout(() => {
      void loadInitialSearchFromStorage();
    }, 0);

    return () => {
      window.clearTimeout(timerId);
    };
  }, [isLoading, user, loadInitialSearchFromStorage]);

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

  const handleSelectUser = (targetUser: UserResponse) => {
    if (hasUnsavedChanges) {
      setPageMessage("未保存の変更があります。先に保存するか、変更を破棄してください。");
      setPageMessageVariant("warning");
      return;
    }

    setSelectedUser(targetUser);
    void loadPageData(targetUser);
  };

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
    key: K,
    value: CommuterPassViewForm[K],
  ) => {
    setCommuterPass((current) => ({ ...current, [key]: value }));
    setIsCommuterPassDirty(true);
  };

  const handleResetCommuterPass = () => {
    if (!loadedUser) {
      setPageMessage("先に対象ユーザーの勤怠を読み込んでください。");
      setPageMessageVariant("error");
      return;
    }

    setCommuterPass(resetCommuterPassViewForm());
    setIsCommuterPassDirty(true);
    setPageMessage("月次通勤定期を初期値に戻しました。全体保存で反映されます。");
    setPageMessageVariant("info");
  };

  const requestMonthChange = (nextTargetMonth: string) => {
    if (hasUnsavedChanges) {
      setPendingTargetMonth(nextTargetMonth);
      setPageMessage("未保存の変更があります。保存して移動するか、保存せず移動してください。");
      setPageMessageVariant("warning");
      return;
    }

    const nextParsedMonth = parseTargetMonth(nextTargetMonth);

    setTargetMonth(nextTargetMonth);

    if (loadedUser) {
      void loadPageData(
        loadedUser,
        nextParsedMonth.targetYear,
        nextParsedMonth.targetMonthValue,
      );
    }
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
    if (!loadedUser) {
      setPageMessage("先に対象ユーザーの勤怠を読み込んでください。");
      setPageMessageVariant("error");
      return false;
    }

    const saveTargetRows = attendanceRows.filter(
      (row) => row.isDirty || row.breaks.some((breakRow) => breakRow.isDirty),
    );

    if (!isCommuterPassDirty && saveTargetRows.length === 0) {
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

    for (const row of saveTargetRows) {
      if (row.planAttendanceTypeId === 0) {
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
    }

    setPageMessage("管理者として月次勤怠を全体保存しています。");
    setPageMessageVariant("info");

    let request;

    try {
      request = buildUpdateMonthlyAttendanceSaveRequest(
        loadedUser.id,
        targetYear,
        targetMonthValue,
        commuterPass,
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

    await loadPageData(loadedUser);

    return true;
  };

  const handleSaveAllAndMove = async () => {
    if (!pendingTargetMonth || !loadedUser) {
      return;
    }

    const isSaved = await handleSaveAllAttendanceDays();

    if (!isSaved) {
      return;
    }

    const nextParsedMonth = parseTargetMonth(pendingTargetMonth);

    setPendingTargetMonth(null);
    setTargetMonth(pendingTargetMonth);

    await loadPageData(
      loadedUser,
      nextParsedMonth.targetYear,
      nextParsedMonth.targetMonthValue,
    );
  };

  const handleDiscardAndMove = async () => {
    if (!pendingTargetMonth) {
      return;
    }

    const nextParsedMonth = parseTargetMonth(pendingTargetMonth);

    setPendingTargetMonth(null);
    setTargetMonth(pendingTargetMonth);

    if (loadedUser) {
      await loadPageData(
        loadedUser,
        nextParsedMonth.targetYear,
        nextParsedMonth.targetMonthValue,
      );
    }
  };

  const handleResetAttendanceDay = (row: AttendanceViewRow) => {
    if (!loadedUser) {
      setPageMessage("先に対象ユーザーの勤怠を読み込んでください。");
      setPageMessageVariant("error");
      return;
    }

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

  const handleMonthlySubmit = async () => {
    if (!loadedUser) {
      setPageMessage("先に対象ユーザーの勤怠を読み込んでください。");
      setPageMessageVariant("error");
      return;
    }

    if (hasUnsavedChanges) {
      setPageMessage("未保存の変更があります。先に全体保存してください。");
      setPageMessageVariant("warning");
      return;
    }

    setPageMessage("管理者として月次申請しています。");
    setPageMessageVariant("info");

    const result = await submitMonthlyAttendanceRequest({
      targetUserId: loadedUser.id,
      targetYear,
      targetMonth: targetMonthValue,
      requestMemo: null,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "月次申請に失敗しました。");
      setPageMessageVariant("error");
      return;
    }

    const data = result.data;

    setMonthlyAttendanceRequest(data.monthlyAttendanceRequest);
    setPageMessage(result.message || "月次申請しました。");
    setPageMessageVariant("success");

    await loadPageData(loadedUser);
  };

  const handleMonthlyWithdraw = async () => {
    if (!loadedUser) {
      setPageMessage("先に対象ユーザーの勤怠を読み込んでください。");
      setPageMessageVariant("error");
      return;
    }

    if (hasUnsavedChanges) {
      setPageMessage("未保存の変更があります。先に全体保存するか、変更を破棄してください。");
      setPageMessageVariant("warning");
      return;
    }

    setPageMessage("管理者として月次申請を取り下げています。");
    setPageMessageVariant("info");

    const result = await withdrawMonthlyAttendanceRequest({
      targetUserId: loadedUser.id,
      targetYear,
      targetMonth: targetMonthValue,
      canceledReason: null,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "月次申請の取り下げに失敗しました。");
      setPageMessageVariant("error");
      return;
    }

    const data = result.data;

    setMonthlyAttendanceRequest(data.monthlyAttendanceRequest);
    setPageMessage(result.message || "月次申請を取り下げました。");
    setPageMessageVariant("success");

    await loadPageData(loadedUser);
  };

  const handleApproveMonthlyAttendanceRequest = async () => {
    if (!loadedUser || !monthlyAttendanceRequest?.id) {
      setPageMessage("承認対象の月次申請がありません。");
      setPageMessageVariant("error");
      return;
    }

    if (hasUnsavedChanges) {
      setPageMessage("未保存の変更があります。先に全体保存してください。");
      setPageMessageVariant("warning");
      return;
    }

    setPageMessage("月次申請を承認しています。");
    setPageMessageVariant("info");

    const result = await approveMonthlyAttendanceRequest({
      targetRequestId: monthlyAttendanceRequest.id,
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "月次申請の承認に失敗しました。");
      setPageMessageVariant("error");
      return;
    }

    const data = result.data;

    setMonthlyAttendanceRequest(data.monthlyAttendanceRequest);
    setPageMessage(result.message || "月次申請を承認しました。");
    setPageMessageVariant("success");

    await loadPageData(loadedUser);
  };

  const handleRejectMonthlyAttendanceRequest = async () => {
    if (!loadedUser || !monthlyAttendanceRequest?.id) {
      setPageMessage("否認対象の月次申請がありません。");
      setPageMessageVariant("error");
      return;
    }

    if (hasUnsavedChanges) {
      setPageMessage("未保存の変更があります。先に全体保存してください。");
      setPageMessageVariant("warning");
      return;
    }

    const rejectedReason = window.prompt("否認理由を入力してください。");

    if (rejectedReason === null) {
      return;
    }

    if (rejectedReason.trim() === "") {
      setPageMessage("否認理由を入力してください。");
      setPageMessageVariant("error");
      return;
    }

    setPageMessage("月次申請を否認しています。");
    setPageMessageVariant("info");

    const result = await rejectMonthlyAttendanceRequest({
      targetRequestId: monthlyAttendanceRequest.id,
      rejectedReason: rejectedReason.trim(),
    });

    if (result.error || !result.data) {
      setPageMessage(result.message || "月次申請の否認に失敗しました。");
      setPageMessageVariant("error");
      return;
    }

    const data = result.data;

    setMonthlyAttendanceRequest(data.monthlyAttendanceRequest);
    setPageMessage(result.message || "月次申請を否認しました。");
    setPageMessageVariant("success");

    await loadPageData(loadedUser);
  };

  if (isLoading || !user) {
    return (
      <PageContainer>
        <AdminSideMenu />

        <section className={styles.loadingCard}>
          <PageTitle title="管理者 勤怠編集" description="ログイン情報を確認しています。" />
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
          <AdminUserSearch
            selectedUser={selectedUser}
            disabled={hasUnsavedChanges || isPageLoading}
            onSelectUser={handleSelectUser}
          />

          <AttendanceMonthHeader
            targetMonth={targetMonth}
            monthlyStatus={monthlyStatus}
            monthlySubmitDisabled={
              !loadedUser ||
              hasUnsavedChanges ||
              isPageLoading ||
              monthlyAttendanceRequest?.canSubmit !== true
            }
            monthlyWithdrawDisabled={
              !loadedUser ||
              hasUnsavedChanges ||
              isPageLoading ||
              monthlyAttendanceRequest?.canCancel !== true
            }
            saveDisabled={!loadedUser || !hasUnsavedChanges || isPageLoading}
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

          <AdminApprovalActions
            monthlyStatus={monthlyStatus}
            monthlyRequestId={monthlyAttendanceRequest?.id ?? null}
            disabled={!loadedUser || hasUnsavedChanges || isPageLoading}
            canApprove={monthlyAttendanceRequest?.canApprove === true}
            canReject={monthlyAttendanceRequest?.canReject === true}
            onApprove={handleApproveMonthlyAttendanceRequest}
            onReject={handleRejectMonthlyAttendanceRequest}
          />

          <MonthlyCommuterPassForm
            commuterPass={commuterPass}
            disabled={!loadedUser || isPageLoading}
            onChange={updateCommuterPassForm}
            onReset={handleResetCommuterPass}
          />

          <AttendanceTable
            rows={attendanceRows}
            attendanceTypes={attendanceTypes}
            getRowLocked={() => !loadedUser || isPageLoading}
            onChangeRow={updateRow}
            onDeleteRow={handleResetAttendanceDay}
            onAddBreak={handleAddBreak}
            onChangeBreak={handleChangeBreak}
            onDeleteBreak={handleDeleteBreak}
          />
        </section>
      </div>
    </PageContainer>
  );
}
