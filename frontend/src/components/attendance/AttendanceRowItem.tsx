"use client";

import Button from "@/components/atoms/Button";
import Input from "@/components/atoms/Input";
import type { AttendanceType } from "@/types/user/attendanceType";
import type { AttendanceBreakViewRow, AttendanceViewRow } from "@/types/user/attendanceView";
import AttendanceLockedText from "@/components/attendance/AttendanceLockedText";
import AttendanceStatusBadge from "@/components/attendance/AttendanceStatusBadge";
import styles from "@/app/user/attendance/page.module.css";

type AttendanceRowItemProps = {
  row: AttendanceViewRow;
  attendanceTypes: AttendanceType[];
  locked: boolean;
  onChangeRow: <K extends keyof AttendanceViewRow>(workDate: string, key: K, value: AttendanceViewRow[K]) => void;
  onDeleteRow: (row: AttendanceViewRow) => void;
  onAddBreak: (workDate: string) => void;
  onChangeBreak: <K extends keyof AttendanceBreakViewRow>(workDate: string, breakIndex: number, key: K, value: AttendanceBreakViewRow[K]) => void;
  onDeleteBreak: (row: AttendanceViewRow, breakIndex: number) => void;
};

function getWorkFlagValue(row: AttendanceViewRow) {
  if (row.lateFlag) {
    return "late";
  }

  if (row.earlyLeaveFlag) {
    return "earlyLeave";
  }

  if (row.absenceFlag) {
    return "absence";
  }

  if (row.sickLeaveFlag) {
    return "sickLeave";
  }

  return "";
}

function getWeekdayRowClass(weekday: string) {
  if (weekday === "土") {
    return styles.saturdayRow;
  }

  if (weekday === "日") {
    return styles.sundayRow;
  }

  return "";
}

export default function AttendanceRowItem({
  row,
  attendanceTypes,
  locked,
  onChangeRow,
  onDeleteRow,
  onAddBreak,
  onChangeBreak,
  onDeleteBreak,
}: AttendanceRowItemProps) {
  const selectedPlanType = attendanceTypes.find((attendanceType) => attendanceType.id === row.planAttendanceTypeId);
  const selectedActualType = attendanceTypes.find((attendanceType) => attendanceType.id === row.actualAttendanceTypeId);

  /*
   * 勤務区分マスタの設定をもとに画面制御する。
   *
   * 重要：
   * 休日だけは予定・実績の時間入力を出さない。
   * 有給・休職などは syncPlanActual=true でも時間入力が必要なので、時間欄を出す。
   */
  const isHoliday = selectedPlanType?.code === "HOLIDAY";
  const syncPlanActual = selectedPlanType?.syncPlanActual === true;
  const requiresRequest = selectedPlanType?.requiresRequest === true;
  const allowBreakInput = selectedPlanType ? selectedPlanType.allowBreakInput === true : true;
  const allowTransportInput = selectedPlanType ? selectedPlanType.allowTransportInput === true : true;

  /*
   * syncPlanActual=true かつ休日ではない区分は、
   * 共通時間を入力して予定・実績に反映する。
   *
   * 例：
   * ・有給
   * ・特別休暇
   * ・休職
   * ・介護休業
   * ・育児休業
   */
  const showCommonTimeInput = syncPlanActual && !isHoliday;

  const handleWorkFlagChange = (value: string) => {
    onChangeRow(row.workDate, "lateFlag", value === "late");
    onChangeRow(row.workDate, "earlyLeaveFlag", value === "earlyLeave");
    onChangeRow(row.workDate, "absenceFlag", value === "absence");
    onChangeRow(row.workDate, "sickLeaveFlag", value === "sickLeave");
  };

  /*
   * 予定区分変更時の制御
   *
   * syncPlanActual=true の区分は、実績区分も同じ区分にする。
   *
   * 休日の場合：
   * ・予定/実績/共通の時刻を全部クリア
   * ・遅刻/早退/欠勤/病欠をクリア
   * ・交通費をクリア
   */
  const handlePlanAttendanceTypeChange = (attendanceTypeId: number) => {
    const nextType = attendanceTypes.find((attendanceType) => attendanceType.id === attendanceTypeId);

    onChangeRow(row.workDate, "planAttendanceTypeId", attendanceTypeId);

    if (!nextType) {
      onChangeRow(row.workDate, "actualAttendanceTypeId", null);
      return;
    }

    if (nextType.syncPlanActual) {
      onChangeRow(row.workDate, "actualAttendanceTypeId", attendanceTypeId);
    }

    if (nextType.code === "HOLIDAY") {
      onChangeRow(row.workDate, "commonStartTime", "");
      onChangeRow(row.workDate, "commonEndTime", "");
      onChangeRow(row.workDate, "planStartTime", "");
      onChangeRow(row.workDate, "planEndTime", "");
      onChangeRow(row.workDate, "actualStartTime", "");
      onChangeRow(row.workDate, "actualEndTime", "");

      onChangeRow(row.workDate, "lateFlag", false);
      onChangeRow(row.workDate, "earlyLeaveFlag", false);
      onChangeRow(row.workDate, "absenceFlag", false);
      onChangeRow(row.workDate, "sickLeaveFlag", false);

      onChangeRow(row.workDate, "transportFrom", "");
      onChangeRow(row.workDate, "transportTo", "");
      onChangeRow(row.workDate, "transportMethod", "");
      onChangeRow(row.workDate, "transportAmount", "");
      return;
    }

    if (!nextType.allowLateFlag) {
      onChangeRow(row.workDate, "lateFlag", false);
    }

    if (!nextType.allowEarlyLeaveFlag) {
      onChangeRow(row.workDate, "earlyLeaveFlag", false);
    }

    if (!nextType.allowAbsenceFlag) {
      onChangeRow(row.workDate, "absenceFlag", false);
    }

    if (!nextType.allowSickLeaveFlag) {
      onChangeRow(row.workDate, "sickLeaveFlag", false);
    }

    if (!nextType.allowTransportInput) {
      onChangeRow(row.workDate, "transportFrom", "");
      onChangeRow(row.workDate, "transportTo", "");
      onChangeRow(row.workDate, "transportMethod", "");
      onChangeRow(row.workDate, "transportAmount", "");
    }
  };

  return (
    <tr className={`${styles.row} ${getWeekdayRowClass(row.weekday)} ${requiresRequest ? styles.rowRequestRequired : ""} ${locked ? styles.rowLocked : ""}`}>
      <td className={styles.td}>
        <p className={styles.dayLabel}>{row.dayLabel}</p>
        <p className={styles.weekday}>{row.weekday}</p>
        {row.isDirty && <p className={styles.unsavedText}>未保存</p>}
      </td>

      <td className={styles.td}>
        <div className={styles.horizontalBlock}>
          <select
            aria-label={`${row.dayLabel}の予定区分`}
            value={row.planAttendanceTypeId}
            onChange={(event) => handlePlanAttendanceTypeChange(Number(event.target.value))}
            className={styles.select}
            disabled={locked}
          >
            <option value={0}>選択</option>
            {attendanceTypes.map((attendanceType) => (
              <option key={attendanceType.id} value={attendanceType.id}>
                {attendanceType.name}
              </option>
            ))}
          </select>

          {isHoliday ? (
            <p className={styles.syncText}>時間入力なし</p>
          ) : showCommonTimeInput ? (
            <>
              <Input type="time" value={row.commonStartTime} onChange={(event) => onChangeRow(row.workDate, "commonStartTime", event.target.value)} disabled={locked} />
              <Input type="time" value={row.commonEndTime} onChange={(event) => onChangeRow(row.workDate, "commonEndTime", event.target.value)} disabled={locked} />
            </>
          ) : (
            <>
              <Input type="time" value={row.planStartTime} onChange={(event) => onChangeRow(row.workDate, "planStartTime", event.target.value)} disabled={locked} />
              <Input type="time" value={row.planEndTime} onChange={(event) => onChangeRow(row.workDate, "planEndTime", event.target.value)} disabled={locked} />
            </>
          )}
        </div>

        {isHoliday && <p className={styles.subText}>休日は時間なしで予定・実績へ反映します。</p>}
        {showCommonTimeInput && <p className={styles.subText}>予定・実績へ同時反映</p>}
      </td>

      <td className={styles.td}>
        <div className={styles.horizontalBlock}>
          {syncPlanActual ? (
            <p className={styles.syncText}>
              実績：{selectedPlanType?.name ?? "未選択"}
              {isHoliday ? " / 時間入力なし" : ""}
            </p>
          ) : (
            <>
              <select
                aria-label={`${row.dayLabel}の実績区分`}
                value={row.actualAttendanceTypeId ?? 0}
                onChange={(event) => onChangeRow(row.workDate, "actualAttendanceTypeId", Number(event.target.value))}
                className={styles.select}
                disabled={locked}
              >
                <option value={0}>選択</option>
                {attendanceTypes.map((attendanceType) => (
                  <option key={attendanceType.id} value={attendanceType.id}>
                    {attendanceType.name}
                  </option>
                ))}
              </select>

              <Input
                type="time"
                value={row.actualStartTime}
                onChange={(event) => onChangeRow(row.workDate, "actualStartTime", event.target.value)}
                disabled={locked || !(selectedActualType?.allowActualTimeInput ?? true)}
              />

              <Input
                type="time"
                value={row.actualEndTime}
                onChange={(event) => onChangeRow(row.workDate, "actualEndTime", event.target.value)}
                disabled={locked || !(selectedActualType?.allowActualTimeInput ?? true)}
              />
            </>
          )}

          <select
            aria-label={`${row.dayLabel}の勤怠状態`}
            value={getWorkFlagValue(row)}
            onChange={(event) => handleWorkFlagChange(event.target.value)}
            className={styles.select}
            disabled={locked || syncPlanActual || isHoliday}
          >
            <option value="">通常</option>
            {(selectedPlanType?.allowLateFlag ?? true) && <option value="late">遅刻</option>}
            {(selectedPlanType?.allowEarlyLeaveFlag ?? true) && <option value="earlyLeave">早退</option>}
            {(selectedPlanType?.allowAbsenceFlag ?? true) && <option value="absence">欠勤</option>}
            {(selectedPlanType?.allowSickLeaveFlag ?? true) && <option value="sickLeave">病欠</option>}
          </select>
        </div>
      </td>

      <td className={styles.td}>
        {allowBreakInput ? (
          <div className={styles.breakEditArea}>
            <Button type="button" variant="secondary" onClick={() => onAddBreak(row.workDate)} disabled={locked}>
              休憩追加
            </Button>

            {row.breaks.length === 0 && <p className={styles.noBreakText}>なし</p>}

            {row.breaks.map((breakRow, breakIndex) => (
              <div key={`${breakRow.id ?? "new"}-${breakIndex}`} className={styles.breakEditRow}>
                <Input type="time" value={breakRow.breakStartTime} onChange={(event) => onChangeBreak(row.workDate, breakIndex, "breakStartTime", event.target.value)} disabled={locked} />

                <Input type="time" value={breakRow.breakEndTime} onChange={(event) => onChangeBreak(row.workDate, breakIndex, "breakEndTime", event.target.value)} disabled={locked} />

                <Input placeholder="メモ" value={breakRow.breakMemo} onChange={(event) => onChangeBreak(row.workDate, breakIndex, "breakMemo", event.target.value)} disabled={locked} />

                <Button type="button" variant="danger" onClick={() => onDeleteBreak(row, breakIndex)} disabled={locked}>
                  削除
                </Button>
              </div>
            ))}
          </div>
        ) : (
          <p className={styles.noBreakText}>対象外</p>
        )}
      </td>

      <td className={styles.td}>
        {allowTransportInput ? (
          <div className={styles.transportCompactGrid}>
            <label className={styles.miniField}>
              <span className={styles.miniLabel}>出発地</span>
              <Input placeholder="例：新宿" value={row.transportFrom} onChange={(event) => onChangeRow(row.workDate, "transportFrom", event.target.value)} disabled={locked} />
            </label>

            <label className={styles.miniField}>
              <span className={styles.miniLabel}>目的地</span>
              <Input placeholder="例：品川" value={row.transportTo} onChange={(event) => onChangeRow(row.workDate, "transportTo", event.target.value)} disabled={locked} />
            </label>

            <label className={styles.miniField}>
              <span className={styles.miniLabel}>手段</span>
              <select aria-label={`${row.dayLabel}の交通手段`} value={row.transportMethod} onChange={(event) => onChangeRow(row.workDate, "transportMethod", event.target.value)} className={styles.select} disabled={locked}>
                <option value="">選択</option>
                <option value="電車">電車</option>
                <option value="バス">バス</option>
                <option value="車">車</option>
                <option value="徒歩">徒歩</option>
                <option value="その他">その他</option>
              </select>
            </label>

            <label className={styles.miniField}>
              <span className={styles.miniLabel}>金額</span>
              <Input placeholder="例：320" type="number" value={row.transportAmount} onChange={(event) => onChangeRow(row.workDate, "transportAmount", event.target.value)} disabled={locked} />
            </label>
          </div>
        ) : (
          <p className={styles.noBreakText}>対象外</p>
        )}
      </td>

      <td className={styles.td}>
        <AttendanceStatusBadge status={row.requestStatus} requiresRequest={requiresRequest} />

        <p className={styles.rowMessage}>{row.systemMessage || "通常"}</p>

        {requiresRequest && row.requestStatus !== "PENDING" && row.requestStatus !== "APPROVED" && <p className={styles.warningText}>申請必要</p>}

        {row.rejectedReason && <p className={styles.rejectedReason}>否認理由：{row.rejectedReason}</p>}

        <div className={styles.requestMemo}>
          <Input placeholder="申請メモ" value={row.requestMemo} onChange={(event) => onChangeRow(row.workDate, "requestMemo", event.target.value)} disabled={locked} />
        </div>
      </td>

      <td className={styles.td}>
        <div className={styles.actionList}>
          <Button type="button" variant="danger" onClick={() => onDeleteRow(row)} disabled={locked || row.attendanceDayId === null}>
            削除
          </Button>

          {locked && <AttendanceLockedText />}
        </div>
      </td>
    </tr>
  );
}