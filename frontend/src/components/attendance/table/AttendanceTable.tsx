"use client";

import { useEffect, useRef, useState } from "react";
import type { AttendanceType } from "@/types/user/attendanceType";
import type {
  AttendanceBreakViewRow,
  AttendanceTransportExpenseViewRow,
  AttendanceViewRow as UserAttendanceViewRow,
} from "@/types/user/attendanceView";
import AttendanceRowItem from "@/components/attendance/rowItem/AttendanceRowItem";
import styles from "./AttendanceTable.module.css";

type AttendanceTableProps<TRow extends UserAttendanceViewRow> = {
  rows: TRow[];
  attendanceTypes: AttendanceType[];
  getRowLocked: (row: TRow) => boolean;
  onChangeRow: <K extends keyof TRow>(
    workDate: string,
    key: K,
    value: TRow[K],
  ) => void;
  onDeleteRow: (row: TRow) => void;
  onAddBreak: (workDate: string) => void;
  onChangeBreak: <K extends keyof AttendanceBreakViewRow>(
    workDate: string,
    breakIndex: number,
    key: K,
    value: AttendanceBreakViewRow[K],
  ) => void;
  onDeleteBreak: (row: TRow, breakIndex: number) => void;
  onAddTransportExpense: (workDate: string) => void;
  onChangeTransportExpense: <K extends keyof AttendanceTransportExpenseViewRow>(
    workDate: string,
    transportExpenseIndex: number,
    key: K,
    value: AttendanceTransportExpenseViewRow[K],
  ) => void;
  onDeleteTransportExpense: (
    row: TRow,
    transportExpenseIndex: number,
  ) => void;
};

type CopiedAttendanceRow = {
  sourceWorkDate: string;
  sourceDayLabel: string;
  planAttendanceTypeId: UserAttendanceViewRow["planAttendanceTypeId"];
  commonStartTime: UserAttendanceViewRow["commonStartTime"];
  commonEndTime: UserAttendanceViewRow["commonEndTime"];
  planStartTime: UserAttendanceViewRow["planStartTime"];
  planEndTime: UserAttendanceViewRow["planEndTime"];
  actualStartTime: UserAttendanceViewRow["actualStartTime"];
  actualEndTime: UserAttendanceViewRow["actualEndTime"];
  scheduledWorkMinutes: UserAttendanceViewRow["scheduledWorkMinutes"];
  actualWorkStatus: UserAttendanceViewRow["actualWorkStatus"];
  lateFlag: UserAttendanceViewRow["lateFlag"];
  earlyLeaveFlag: UserAttendanceViewRow["earlyLeaveFlag"];
  absenceFlag: UserAttendanceViewRow["absenceFlag"];
  sickLeaveFlag: UserAttendanceViewRow["sickLeaveFlag"];
  remoteWorkAllowanceFlag: UserAttendanceViewRow["remoteWorkAllowanceFlag"];
  breaks: AttendanceBreakViewRow[];
  transportExpenses: AttendanceTransportExpenseViewRow[];
};

function buildCopiedAttendanceRow(
  row: UserAttendanceViewRow,
): CopiedAttendanceRow {
  return {
    sourceWorkDate: row.workDate,
    sourceDayLabel: row.dayLabel,
    planAttendanceTypeId: row.planAttendanceTypeId,
    commonStartTime: row.commonStartTime,
    commonEndTime: row.commonEndTime,
    planStartTime: row.planStartTime,
    planEndTime: row.planEndTime,
    actualStartTime: row.actualStartTime,
    actualEndTime: row.actualEndTime,
    scheduledWorkMinutes: row.scheduledWorkMinutes,
    actualWorkStatus: row.actualWorkStatus,
    lateFlag: row.lateFlag,
    earlyLeaveFlag: row.earlyLeaveFlag,
    absenceFlag: row.absenceFlag,
    sickLeaveFlag: row.sickLeaveFlag,
    remoteWorkAllowanceFlag: row.remoteWorkAllowanceFlag,
    breaks: row.breaks.map((breakRow) => ({
      ...breakRow,
      id: null,
      isDirty: true,
    })),
    transportExpenses: row.transportExpenses.map(
      (transportExpense, index) => ({
        ...transportExpense,
        id: null,
        sortOrder: index + 1,
        isDirty: true,
      }),
    ),
  };
}

export default function AttendanceTable<TRow extends UserAttendanceViewRow>({
  rows,
  attendanceTypes,
  getRowLocked,
  onChangeRow,
  onDeleteRow,
  onAddBreak,
  onChangeBreak,
  onDeleteBreak,
  onAddTransportExpense,
  onChangeTransportExpense,
  onDeleteTransportExpense,
}: AttendanceTableProps<TRow>) {
  const topScrollRef = useRef<HTMLDivElement | null>(null);
  const tableWrapRef = useRef<HTMLDivElement | null>(null);
  const tableRef = useRef<HTMLTableElement | null>(null);
  const isSyncingScrollRef = useRef(false);

  const [copiedRow, setCopiedRow] = useState<CopiedAttendanceRow | null>(
    null,
  );
  const [showTopScrollbar, setShowTopScrollbar] = useState(false);
  const [scrollContentWidth, setScrollContentWidth] = useState(0);

  const handleChangeRow = <K extends keyof UserAttendanceViewRow>(
    workDate: string,
    key: K,
    value: UserAttendanceViewRow[K],
  ) => {
    onChangeRow(
      workDate,
      key as unknown as keyof TRow,
      value as unknown as TRow[keyof TRow],
    );
  };

  useEffect(() => {
    const updateScrollbarState = () => {
      const tableWrap = tableWrapRef.current;
      const table = tableRef.current;

      if (!tableWrap || !table) {
        return;
      }

      const nextScrollWidth = Math.max(
        table.scrollWidth,
        tableWrap.scrollWidth,
      );

      setScrollContentWidth(nextScrollWidth);
      setShowTopScrollbar(nextScrollWidth > tableWrap.clientWidth + 1);

      if (topScrollRef.current) {
        topScrollRef.current.scrollLeft = tableWrap.scrollLeft;
      }
    };

    updateScrollbarState();

    const resizeObserver = new ResizeObserver(updateScrollbarState);

    if (tableWrapRef.current) {
      resizeObserver.observe(tableWrapRef.current);
    }

    if (tableRef.current) {
      resizeObserver.observe(tableRef.current);
    }

    window.addEventListener("resize", updateScrollbarState);

    return () => {
      resizeObserver.disconnect();
      window.removeEventListener("resize", updateScrollbarState);
    };
  }, [rows]);

  const syncScroll = (
    source: HTMLDivElement,
    target: HTMLDivElement | null,
  ) => {
    if (!target || isSyncingScrollRef.current) {
      return;
    }

    isSyncingScrollRef.current = true;
    target.scrollLeft = source.scrollLeft;

    window.requestAnimationFrame(() => {
      isSyncingScrollRef.current = false;
    });
  };

  const handleTopScroll = () => {
    if (!topScrollRef.current) {
      return;
    }

    syncScroll(topScrollRef.current, tableWrapRef.current);
  };

  const handleTableScroll = () => {
    if (!tableWrapRef.current) {
      return;
    }

    syncScroll(tableWrapRef.current, topScrollRef.current);
  };

  const handleCopyRow = (row: UserAttendanceViewRow) => {
    setCopiedRow(buildCopiedAttendanceRow(row));
  };

  const handlePasteRow = (targetRow: UserAttendanceViewRow) => {
    if (!copiedRow) {
      return;
    }

    handleChangeRow(
      targetRow.workDate,
      "planAttendanceTypeId",
      copiedRow.planAttendanceTypeId,
    );
    handleChangeRow(
      targetRow.workDate,
      "commonStartTime",
      copiedRow.commonStartTime,
    );
    handleChangeRow(
      targetRow.workDate,
      "commonEndTime",
      copiedRow.commonEndTime,
    );
    handleChangeRow(
      targetRow.workDate,
      "planStartTime",
      copiedRow.planStartTime,
    );
    handleChangeRow(
      targetRow.workDate,
      "planEndTime",
      copiedRow.planEndTime,
    );
    handleChangeRow(
      targetRow.workDate,
      "actualStartTime",
      copiedRow.actualStartTime,
    );
    handleChangeRow(
      targetRow.workDate,
      "actualEndTime",
      copiedRow.actualEndTime,
    );
    handleChangeRow(
      targetRow.workDate,
      "scheduledWorkMinutes",
      copiedRow.scheduledWorkMinutes,
    );
    handleChangeRow(
      targetRow.workDate,
      "actualWorkStatus",
      copiedRow.actualWorkStatus,
    );
    handleChangeRow(targetRow.workDate, "lateFlag", copiedRow.lateFlag);
    handleChangeRow(
      targetRow.workDate,
      "earlyLeaveFlag",
      copiedRow.earlyLeaveFlag,
    );
    handleChangeRow(
      targetRow.workDate,
      "absenceFlag",
      copiedRow.absenceFlag,
    );
    handleChangeRow(
      targetRow.workDate,
      "sickLeaveFlag",
      copiedRow.sickLeaveFlag,
    );
    handleChangeRow(
      targetRow.workDate,
      "remoteWorkAllowanceFlag",
      copiedRow.remoteWorkAllowanceFlag,
    );
    handleChangeRow(
      targetRow.workDate,
      "breaks",
      copiedRow.breaks.map((breakRow) => ({
        ...breakRow,
        id: null,
        isDirty: true,
      })),
    );
    handleChangeRow(
      targetRow.workDate,
      "transportExpenses",
      copiedRow.transportExpenses.map((transportExpense, index) => ({
        ...transportExpense,
        id: null,
        sortOrder: index + 1,
        isDirty: true,
      })),
    );
  };

  return (
    <section>
      <div className={styles.sectionHeader}>
        <div>
          <h2 className={styles.sectionTitle}>日別勤怠</h2>
          <p className={styles.sectionDescription}>
            勤務区分マスタの設定に従って、入力欄を切り替えます。
          </p>
          {copiedRow && (
            <p className={styles.copyStatus}>
              {copiedRow.sourceDayLabel} の勤怠をコピー中
            </p>
          )}
        </div>
      </div>

      {showTopScrollbar && (
        <div className={styles.topScrollbarArea}>
          <span className={styles.topScrollbarLabel}>横スクロール</span>
          <div
            ref={topScrollRef}
            className={styles.topScrollbar}
            onScroll={handleTopScroll}
            aria-label="勤怠表の横スクロール"
          >
            <div
              className={styles.topScrollbarContent}
              style={{ width: `${scrollContentWidth}px` }}
            />
          </div>
        </div>
      )}

      <div
        ref={tableWrapRef}
        className={styles.tableWrap}
        onScroll={handleTableScroll}
      >
        <table ref={tableRef} className={styles.table}>
          <thead>
            <tr className={styles.tableHeadRow}>
              <th className={`${styles.th} ${styles.copyColumn}`}>
                コピー操作
              </th>
              <th className={`${styles.th} ${styles.dateColumn}`}>日付</th>
              <th className={`${styles.th} ${styles.planColumn}`}>予定</th>
              <th className={`${styles.th} ${styles.actualColumn}`}>実績</th>
              <th className={`${styles.th} ${styles.scheduledColumn}`}>所定</th>
              <th className={`${styles.th} ${styles.breakColumn}`}>休憩</th>
              <th className={`${styles.th} ${styles.transportColumn}`}>交通費</th>
              <th className={`${styles.th} ${styles.statusColumn}`}>状態</th>
              <th className={`${styles.th} ${styles.actionColumn}`}>操作</th>
            </tr>
          </thead>

          <tbody>
            {rows.map((row) => (
              <AttendanceRowItem
                key={row.workDate}
                row={row}
                attendanceTypes={attendanceTypes}
                locked={getRowLocked(row)}
                copiedSourceWorkDate={copiedRow?.sourceWorkDate ?? null}
                pasteDisabled={!copiedRow}
                onCopyRow={handleCopyRow}
                onPasteRow={handlePasteRow}
                onChangeRow={handleChangeRow}
                onDeleteRow={() => onDeleteRow(row)}
                onAddBreak={onAddBreak}
                onChangeBreak={onChangeBreak}
                onDeleteBreak={(_, breakIndex) =>
                  onDeleteBreak(row, breakIndex)
                }
                onAddTransportExpense={onAddTransportExpense}
                onChangeTransportExpense={onChangeTransportExpense}
                onDeleteTransportExpense={(_, transportExpenseIndex) =>
                  onDeleteTransportExpense(row, transportExpenseIndex)
                }
              />
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}