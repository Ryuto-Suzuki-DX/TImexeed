"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Button from "@/components/atoms/Button";
import MessageBox from "@/components/atoms/MessageBox";
import PageContainer from "@/components/atoms/PageContainer";
import PageTitle from "@/components/atoms/PageTitle";
import UserSideMenu from "@/components/sideMenu/UserSideMenu";
import AttendanceMonthHeader from "@/components/attendance/AttendanceMonthHeader";
import AttendanceTable from "@/components/attendance/AttendanceTable";
import MonthlyCommuterPassForm from "@/components/attendance/MonthlyCommuterPassForm";
import { useRequireRole } from "@/hooks/useRequireRole";
import { searchAttendanceTypes } from "@/api/user/attendanceType";
import { deleteAttendanceDay, searchAttendanceDays } from "@/api/user/attendanceDay";
import { searchAttendanceBreaks } from "@/api/user/attendanceBreak";
import { searchMonthlyCommuterPass } from "@/api/user/monthlyCommuterPass";
import { updateMonthlyAttendance } from "@/api/user/monthlyAttendance";
import type { AttendanceType } from "@/types/user/attendanceType";
import type { AttendanceBreak } from "@/types/user/attendanceBreak";
import type { AttendanceBreakViewRow, AttendanceViewRow, CommuterPassViewForm, PageMessageVariant } from "@/types/user/attendanceView";
import { buildTargetMonth, getCurrentMonth, parseTargetMonth } from "@/utils/attendance/attendanceDate";
import { getStatusLabel } from "@/utils/attendance/attendanceStatus";
import {
  attachBreaksToAttendanceViewRows,
  buildAttendanceViewRows,
  buildCommuterPassViewForm,
  buildNewAttendanceBreakViewRow,
  buildUpdateMonthlyAttendanceRequest,
} from "@/utils/attendance/userAttendance/userAttendanceMapper";
import {
  isUserAttendanceRowLocked,
  isUserMonthlyCommuterPassLocked,
  isUserMonthlySubmitDisabled,
} from "@/utils/attendance/userAttendance/userAttendancePermission";
import styles from "./page.module.css";

export default function UserAttendancePage() {
  const { user, isLoading, message } = useRequireRole("USER");

  const [targetMonth, setTargetMonth] = useState(getCurrentMonth());
  const [pendingTargetMonth, setPendingTargetMonth] = useState<string | null>(null);
  const [attendanceTypes, setAttendanceTypes] = useState<AttendanceType[]>([]);
  const [attendanceRows, setAttendanceRows] = useState<AttendanceViewRow[]>([]);
  const [commuterPass, setCommuterPass] = useState<CommuterPassViewForm>({
    commuterFrom: "",
    commuterTo: "",
    commuterMethod: "",
    commuterAmount: "",
    monthlyStatus: "DRAFT",
  });
  const [isCommuterPassDirty, setIsCommuterPassDirty] = useState(false);
  const [pageMessage, setPageMessage] = useState("対象月の勤怠を入力できます。");
  const [pageMessageVariant, setPageMessageVariant] = useState<PageMessageVariant>("info");
  const [isPageLoading, setIsPageLoading] = useState(false);

  const { targetYear, targetMonthValue } = useMemo(() => parseTargetMonth(targetMonth), [targetMonth]);

  const hasUnsavedChanges = useMemo(() => {
    return isCommuterPassDirty || attendanceRows.some((row) => row.isDirty || row.breaks.some((breakRow) => breakRow.isDirty));
  }, [attendanceRows, isCommuterPassDirty]);

  const loadPageData = useCallback(async () => {
    if (!user) {
      return;
    }

    setIsPageLoading(true);
    setPageMessage("勤怠情報を取得しています。");
    setPageMessageVariant("info");

    const attendanceTypesResult = await searchAttendanceTypes({});

    if (attendanceTypesResult.error || !attendanceTypesResult.data) {
      setPageMessage(attendanceTypesResult.message || "勤務区分マスタの取得に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
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
      setIsPageLoading(false);
      return;
    }

    const commuterPassResult = await searchMonthlyCommuterPass({
      targetYear,
      targetMonth: targetMonthValue,
    });

    if (commuterPassResult.error || !commuterPassResult.data) {
      setPageMessage(commuterPassResult.message || "月次通勤定期の取得に失敗しました。");
      setPageMessageVariant("error");
      setIsPageLoading(false);
      return;
    }

    const rows = buildAttendanceViewRows(targetYear, targetMonthValue, attendanceDaysResult.data.attendanceDays);
    const registeredRows = rows.filter((row) => row.attendanceDayId !== null);

    const breakResults = await Promise.all(
      registeredRows.map(async (row) => {
        const result = await searchAttendanceBreaks({
          workDate: row.workDate,
        });

        if (result.error || !result.data) {
          return {
            workDate: row.workDate,
            breaks: [] as AttendanceBreak[],
          };
        }

        return {
          workDate: row.workDate,
          breaks: result.data.attendanceBreaks,
        };
      }),
    );

    const breakMap = new Map<string, AttendanceBreak[]>();

    breakResults.forEach((result) => {
      breakMap.set(result.workDate, result.breaks);
    });

    setAttendanceTypes(nextAttendanceTypes);
    setAttendanceRows(attachBreaksToAttendanceViewRows(rows, breakMap));
    setCommuterPass(buildCommuterPassViewForm(commuterPassResult.data.monthlyCommuterPass));
    setIsCommuterPassDirty(false);

    setPageMessage("対象月の勤怠を入力できます。");
    setPageMessageVariant("info");
    setIsPageLoading(false);
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

  const monthlyStatus = useMemo(() => {
    if (commuterPass.monthlyStatus !== "DRAFT") {
      return commuterPass.monthlyStatus;
    }

    const statusFromAttendance = attendanceRows.find((row) => row.monthlyStatus && row.monthlyStatus !== "DRAFT")?.monthlyStatus;

    return statusFromAttendance ?? "DRAFT";
  }, [attendanceRows, commuterPass.monthlyStatus]);

  const hasUnsubmittedRequest = useMemo(() => {
    return attendanceRows.some((row) => {
      const selectedPlanType = attendanceTypes.find((attendanceType) => attendanceType.id === row.planAttendanceTypeId);

      return Boolean(selectedPlanType?.requiresRequest) && row.requestStatus !== "PENDING" && row.requestStatus !== "APPROVED";
    });
  }, [attendanceRows, attendanceTypes]);

  const updateRow = <K extends keyof AttendanceViewRow>(workDate: string, key: K, value: AttendanceViewRow[K]) => {
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

  const updateCommuterPassForm = <K extends keyof CommuterPassViewForm>(key: K, value: CommuterPassViewForm[K]) => {
    setCommuterPass((current) => ({ ...current, [key]: value }));
    setIsCommuterPassDirty(true);
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

  /*
   * 月次勤怠全体保存
   *
   * 保存対象：
   * ・月次通勤定期
   * ・変更された日別勤怠
   * ・変更された日別休憩
   *
   * 休憩は個別保存しない。
   * 画面に残っている休憩だけを送信し、バックエンド側で既存休憩を入れ替える。
   */
  const handleSaveAllAttendanceDays = async () => {
    const saveTargetRows = attendanceRows.filter((row) => row.isDirty || row.breaks.some((breakRow) => breakRow.isDirty));

    if (!isCommuterPassDirty && saveTargetRows.length === 0) {
      setPageMessage("保存対象の勤怠はありません。");
      setPageMessageVariant("info");
      return true;
    }

    for (const row of saveTargetRows) {
      const selectedPlanType = attendanceTypes.find((attendanceType) => attendanceType.id === row.planAttendanceTypeId);

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

    setPageMessage("月次勤怠を全体保存しています。");
    setPageMessageVariant("info");

    let request;

    try {
      request = buildUpdateMonthlyAttendanceRequest(targetYear, targetMonthValue, commuterPass, saveTargetRows, attendanceTypes);
    } catch (error) {
      setPageMessage(error instanceof Error ? error.message : "月次勤怠全体保存のリクエスト作成に失敗しました。");
      setPageMessageVariant("error");
      return false;
    }

    const result = await updateMonthlyAttendance(request);

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

  const handleDeleteAttendanceDay = async (row: AttendanceViewRow) => {
    const result = await deleteAttendanceDay({
      workDate: row.workDate,
    });

    if (result.error) {
      setPageMessage(result.message || "勤怠の削除に失敗しました。");
      setPageMessageVariant("error");
      return;
    }

    setPageMessage(result.message || "勤怠を削除しました。");
    setPageMessageVariant("success");

    await loadPageData();
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

  const handleChangeBreak = <K extends keyof AttendanceBreakViewRow>(workDate: string, breakIndex: number, key: K, value: AttendanceBreakViewRow[K]) => {
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

  /*
   * 休憩削除
   *
   * ここではAPIを呼ばない。
   * 画面stateから削除し、月次勤怠全体保存時にDBへ反映する。
   */
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

  const handleMonthlySubmit = () => {
    setPageMessage("月次申請APIはまだこの画面には接続していません。");
    setPageMessageVariant("warning");
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
            monthlySubmitDisabled={isUserMonthlySubmitDisabled(hasUnsubmittedRequest, monthlyStatus) || hasUnsavedChanges}
            saveDisabled={!hasUnsavedChanges}
            onChangeMonth={handleChangeMonth}
            onPreviousMonth={handlePreviousMonth}
            onNextMonth={handleNextMonth}
            onSaveAll={handleSaveAllAttendanceDays}
            onMonthlySubmit={handleMonthlySubmit}
          />

          {pendingTargetMonth && (
            <div className={styles.unsavedBar}>
              <p className={styles.unsavedBarText}>未保存の変更があります。移動先：{pendingTargetMonth}</p>

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
            <MessageBox variant={pageMessageVariant}>{isPageLoading ? "読み込み中..." : pageMessage}</MessageBox>

            <div className={styles.monthlyStatusBox}>
              <p className={styles.monthlyStatusLabel}>月次申請状態</p>
              <p className={styles.monthlyStatusValue}>{getStatusLabel(monthlyStatus)}</p>
            </div>
          </div>

          <MonthlyCommuterPassForm
            commuterPass={commuterPass}
            disabled={isUserMonthlyCommuterPassLocked(monthlyStatus)}
            onChange={updateCommuterPassForm}
          />

          <AttendanceTable
            rows={attendanceRows}
            attendanceTypes={attendanceTypes}
            getRowLocked={isUserAttendanceRowLocked}
            onChangeRow={updateRow}
            onDeleteRow={handleDeleteAttendanceDay}
            onAddBreak={handleAddBreak}
            onChangeBreak={handleChangeBreak}
            onDeleteBreak={handleDeleteBreak}
          />
        </section>
      </div>
    </PageContainer>
  );
}