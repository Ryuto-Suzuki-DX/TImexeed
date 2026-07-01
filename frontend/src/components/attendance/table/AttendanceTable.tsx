"use client";

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

  return (
    <section>
      <div className={styles.sectionHeader}>
        <div>
          <h2 className={styles.sectionTitle}>日別勤怠</h2>
          <p className={styles.sectionDescription}>
            勤務区分マスタの設定に従って、入力欄を切り替えます。
          </p>
        </div>
      </div>

      <div className={styles.tableWrap}>
        <table className={styles.table}>
          <thead>
            <tr className={styles.tableHeadRow}>
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
                onChangeRow={handleChangeRow}
                onDeleteRow={() => onDeleteRow(row)}
                onAddBreak={onAddBreak}
                onChangeBreak={onChangeBreak}
                onDeleteBreak={(_, breakIndex) => onDeleteBreak(row, breakIndex)}
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
