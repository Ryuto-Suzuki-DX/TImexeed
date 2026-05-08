"use client";

import type { AttendanceType } from "@/types/user/attendanceType";
import type { AttendanceBreakViewRow, AttendanceViewRow } from "@/types/user/attendanceView";
import AttendanceRowItem from "@/components/attendance/AttendanceRowItem";
import styles from "@/app/user/attendance/page.module.css";

type AttendanceTableProps = {
  rows: AttendanceViewRow[];
  attendanceTypes: AttendanceType[];
  getRowLocked: (row: AttendanceViewRow) => boolean;
  onChangeRow: <K extends keyof AttendanceViewRow>(workDate: string, key: K, value: AttendanceViewRow[K]) => void;
  onDeleteRow: (row: AttendanceViewRow) => void;
  onAddBreak: (workDate: string) => void;
  onChangeBreak: <K extends keyof AttendanceBreakViewRow>(workDate: string, breakIndex: number, key: K, value: AttendanceBreakViewRow[K]) => void;
  onDeleteBreak: (row: AttendanceViewRow, breakIndex: number) => void;
};

export default function AttendanceTable({
  rows,
  attendanceTypes,
  getRowLocked,
  onChangeRow,
  onDeleteRow,
  onAddBreak,
  onChangeBreak,
  onDeleteBreak,
}: AttendanceTableProps) {
  return (
    <section>
      <div className={styles.sectionHeader}>
        <div>
          <h2 className={styles.sectionTitle}>日別勤怠</h2>
          <p className={styles.sectionDescription}>勤務区分マスタの設定に従って、入力欄を切り替えます。</p>
        </div>
      </div>

      <div className={styles.tableWrap}>
        <table className={styles.table}>
          <thead>
            <tr className={styles.tableHeadRow}>
              <th className={`${styles.th} ${styles.dateColumn}`}>日付</th>
              <th className={`${styles.th} ${styles.planColumn}`}>予定</th>
              <th className={`${styles.th} ${styles.actualColumn}`}>実績</th>
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
                onChangeRow={onChangeRow}
                onDeleteRow={onDeleteRow}
                onAddBreak={onAddBreak}
                onChangeBreak={onChangeBreak}
                onDeleteBreak={onDeleteBreak}
              />
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}