"use client";

import { useState } from "react";
import Button from "@/components/atoms/Button";
import Input from "@/components/atoms/Input";
import type { AttendanceType } from "@/types/user/attendanceType";
import type {
  AttendanceBreakViewRow,
  AttendanceTransportExpenseViewRow,
  AttendanceViewRow,
} from "@/types/user/attendanceView";
import AttendanceLockedText from "@/components/attendance/lockedText/AttendanceLockedText";
import styles from "./AttendanceRowItem.module.css";

const ACTUAL_WORK_STATUS_NORMAL = "NORMAL";

const actualWorkStatusOptions = [
  { value: "NORMAL", label: "通常" },
  { value: "ABSENCE", label: "欠勤" },
  { value: "SICK_LEAVE", label: "病欠" },
  { value: "LATE", label: "遅刻" },
  { value: "EARLY_LEAVE", label: "早退" },
];

type AttendanceRowItemProps = {
  row: AttendanceViewRow;
  attendanceTypes: AttendanceType[];
  locked: boolean;
  copiedSourceWorkDate: string | null;
  pasteDisabled: boolean;
  onCopyRow: (row: AttendanceViewRow) => void;
  onPasteRow: (row: AttendanceViewRow) => void;
  onChangeRow: <K extends keyof AttendanceViewRow>(
    workDate: string,
    key: K,
    value: AttendanceViewRow[K],
  ) => void;
  onDeleteRow: (row: AttendanceViewRow) => void;
  onAddBreak: (workDate: string) => void;
  onChangeBreak: <K extends keyof AttendanceBreakViewRow>(
    workDate: string,
    breakIndex: number,
    key: K,
    value: AttendanceBreakViewRow[K],
  ) => void;
  onDeleteBreak: (row: AttendanceViewRow, breakIndex: number) => void;
  onAddTransportExpense: (workDate: string) => void;
  onChangeTransportExpense: <K extends keyof AttendanceTransportExpenseViewRow>(
    workDate: string,
    transportExpenseIndex: number,
    key: K,
    value: AttendanceTransportExpenseViewRow[K],
  ) => void;
  onDeleteTransportExpense: (
    row: AttendanceViewRow,
    transportExpenseIndex: number,
  ) => void;
};

function getCalendarRowClass(row: AttendanceViewRow) {
  if (row.isHoliday) {
    return styles.holidayRow;
  }

  if (row.weekday === "土") {
    return styles.saturdayRow;
  }

  if (row.weekday === "日") {
    return styles.sundayRow;
  }

  return "";
}

function hasRowInput(row: AttendanceViewRow) {
  return (
    row.attendanceDayId !== null ||
    row.planAttendanceTypeId !== 0 ||
    row.commonStartTime !== "" ||
    row.commonEndTime !== "" ||
    row.planStartTime !== "" ||
    row.planEndTime !== "" ||
    row.actualStartTime !== "" ||
    row.actualEndTime !== "" ||
    row.scheduledWorkMinutes !== "" ||
    row.remoteWorkAllowanceFlag ||
    row.transportExpenses.length > 0 ||
    row.breaks.length > 0
  );
}

/*
 * 保存しないシステムメッセージを画面側で作る
 *
 * 注意：
 * ・DBには保存しない
 * ・表示専用
 * ・遅刻/早退/欠勤/病欠のフラグは使わない
 */
function buildRowSystemMessage(
  row: AttendanceViewRow,
  selectedPlanType: AttendanceType | undefined,
) {
  if (!selectedPlanType) {
    return "勤務区分未選択";
  }

  if (selectedPlanType.code === "HOLIDAY") {
    return "休日";
  }

  if (selectedPlanType.requiresRequest) {
    return "申請対象";
  }

  if (row.scheduledWorkMinutes !== "") {
    return `所定 ${row.scheduledWorkMinutes}分`;
  }

  return "通常";
}

function getStatusBadgeClass(requiresRequest: boolean) {
  if (requiresRequest) {
    return `${styles.statusBadge} ${styles.statusRequiresRequest}`;
  }

  return `${styles.statusBadge} ${styles.statusNone}`;
}

function getStatusBadgeText(requiresRequest: boolean) {
  if (requiresRequest) {
    return "申請対象";
  }

  return "通常";
}

export default function AttendanceRowItem({
  row,
  attendanceTypes,
  locked,
  copiedSourceWorkDate,
  pasteDisabled,
  onCopyRow,
  onPasteRow,
  onChangeRow,
  onDeleteRow,
  onAddBreak,
  onChangeBreak,
  onDeleteBreak,
  onAddTransportExpense,
  onChangeTransportExpense,
  onDeleteTransportExpense,
}: AttendanceRowItemProps) {
  const [isTransportOpen, setIsTransportOpen] = useState(false);
  const selectedPlanType = attendanceTypes.find(
    (attendanceType) => attendanceType.id === row.planAttendanceTypeId,
  );

  /*
   * 勤務区分マスタの設定をもとに画面制御する。
   *
   * 重要：
   * ・休日だけは予定・実績・所定労働時間を保存しない。
   * ・有給・休職など syncPlanActual=true の区分は、開始/終了ではなく所定労働時間で扱う。
   * ・通常勤務などは予定/実績の開始終了を任意入力できる。
   * ・遅刻/早退/欠勤/病欠のフラグUIは使わない。
   */
  const isHolidayAttendanceType = selectedPlanType?.code === "HOLIDAY";
  const syncPlanActual = selectedPlanType?.syncPlanActual === true;
  const requiresRequest = selectedPlanType?.requiresRequest === true;
  const allowBreakInput = selectedPlanType
    ? selectedPlanType.allowBreakInput === true
    : true;
  const allowTransportInput = selectedPlanType
    ? selectedPlanType.allowTransportInput === true
    : true;
  const allowActualTimeInput = selectedPlanType
    ? selectedPlanType.allowActualTimeInput === true
    : true;
  const rowSystemMessage = buildRowSystemMessage(row, selectedPlanType);
  const resetDisabled = locked || !hasRowInput(row);
  const isCopiedSource = copiedSourceWorkDate === row.workDate;

  /*
   * syncPlanActual=true かつ休日ではない区分は、
   * 開始/終了を入力しない。
   *
   * 例：
   * ・有給
   * ・特別休暇
   * ・休職
   * ・介護休業
   * ・育児休業
   */
  const showScheduledWorkMinutesInput = !isHolidayAttendanceType;
  const showActualWorkInput = !syncPlanActual && !isHolidayAttendanceType;

  /*
   * 予定区分変更時の制御
   *
   * 実績状態は通常勤務系だけ選択できる。
   * 休日・有給・休職などでは NORMAL に戻す。
   *
   * 休日の場合：
   * ・予定/実績/共通の時刻を全部クリア
   * ・派遣先所定労働時間をクリア
   * ・交通費をクリア
   * ・在宅勤務補助ありをクリア
   */
  const handlePlanAttendanceTypeChange = (attendanceTypeId: number) => {
    const nextType = attendanceTypes.find(
      (attendanceType) => attendanceType.id === attendanceTypeId,
    );

    onChangeRow(row.workDate, "planAttendanceTypeId", attendanceTypeId);

    if (!nextType) {
      onChangeRow(
        row.workDate,
        "actualWorkStatus",
        ACTUAL_WORK_STATUS_NORMAL,
      );
      return;
    }

    if (nextType.code === "HOLIDAY") {
      onChangeRow(row.workDate, "commonStartTime", "");
      onChangeRow(row.workDate, "commonEndTime", "");
      onChangeRow(row.workDate, "planStartTime", "");
      onChangeRow(row.workDate, "planEndTime", "");
      onChangeRow(row.workDate, "actualStartTime", "");
      onChangeRow(row.workDate, "actualEndTime", "");
      onChangeRow(row.workDate, "scheduledWorkMinutes", "");
      onChangeRow(
        row.workDate,
        "actualWorkStatus",
        ACTUAL_WORK_STATUS_NORMAL,
      );

      onChangeRow(row.workDate, "lateFlag", false);
      onChangeRow(row.workDate, "earlyLeaveFlag", false);
      onChangeRow(row.workDate, "absenceFlag", false);
      onChangeRow(row.workDate, "sickLeaveFlag", false);

      onChangeRow(row.workDate, "remoteWorkAllowanceFlag", false);

      onChangeRow(row.workDate, "transportExpenses", []);
      return;
    }

    if (nextType.syncPlanActual) {
      onChangeRow(row.workDate, "commonStartTime", "");
      onChangeRow(row.workDate, "commonEndTime", "");
      onChangeRow(row.workDate, "planStartTime", "");
      onChangeRow(row.workDate, "planEndTime", "");
      onChangeRow(row.workDate, "actualStartTime", "");
      onChangeRow(row.workDate, "actualEndTime", "");
      onChangeRow(
        row.workDate,
        "actualWorkStatus",
        ACTUAL_WORK_STATUS_NORMAL,
      );
    }

    onChangeRow(row.workDate, "lateFlag", false);
    onChangeRow(row.workDate, "earlyLeaveFlag", false);
    onChangeRow(row.workDate, "absenceFlag", false);
    onChangeRow(row.workDate, "sickLeaveFlag", false);

    if (!nextType.allowTransportInput) {
      onChangeRow(row.workDate, "transportExpenses", []);
    }
  };

  const dailyTransportTotal = row.transportExpenses.reduce(
    (total, transportExpense) => {
      const amount = Number(transportExpense.transportAmount);

      if (!Number.isFinite(amount)) {
        return total;
      }

      return total + amount;
    },
    0,
  );

  return (
    <tr
      className={`${styles.row} ${getCalendarRowClass(row)} ${
        requiresRequest ? styles.rowRequestRequired : ""
      } ${locked ? styles.rowLocked : ""}`}
    >
      <td className={`${styles.td} ${styles.copyCell}`}>
        <div className={styles.copyActionList}>
          <button
            type="button"
            className={`${styles.copyButton} ${
              isCopiedSource ? styles.copyButtonActive : ""
            }`}
            onClick={() => onCopyRow(row)}
            disabled={locked}
          >
            {isCopiedSource ? "コピー中" : "コピー"}
          </button>

          <button
            type="button"
            className={styles.pasteButton}
            onClick={() => onPasteRow(row)}
            disabled={locked || pasteDisabled || isCopiedSource}
          >
            ペースト
          </button>
        </div>
      </td>

      <td className={`${styles.td} ${styles.dateCell}`}>
        <p className={styles.dayLabel}>{row.dayLabel}</p>
        <p className={styles.weekday}>{row.weekday}</p>
        {row.holidayName && (
          <p className={styles.holidayName}>{row.holidayName}</p>
        )}
        {row.isDirty && <p className={styles.unsavedText}>未保存</p>}
      </td>

      <td className={styles.td}>
        <div className={styles.horizontalBlock}>
          <select
            aria-label={`${row.dayLabel}の予定区分`}
            value={row.planAttendanceTypeId}
            onChange={(event) =>
              handlePlanAttendanceTypeChange(Number(event.target.value))
            }
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

          {isHolidayAttendanceType ? (
            <p className={styles.syncText}>時間入力なし</p>
          ) : syncPlanActual ? (
            <p className={styles.syncText}>開始/終了なし</p>
          ) : (
            <>
              <Input
                type="time"
                value={row.planStartTime}
                onChange={(event) =>
                  onChangeRow(
                    row.workDate,
                    "planStartTime",
                    event.target.value,
                  )
                }
                disabled={locked}
              />
              <Input
                type="time"
                value={row.planEndTime}
                onChange={(event) =>
                  onChangeRow(
                    row.workDate,
                    "planEndTime",
                    event.target.value,
                  )
                }
                disabled={locked}
              />
            </>
          )}
        </div>

        {isHolidayAttendanceType && (
          <p className={styles.subText}>
            休日は時間なしで予定・実績へ反映します。
          </p>
        )}
        {syncPlanActual && !isHolidayAttendanceType && (
          <p className={styles.subText}>
            開始/終了ではなく所定労働時間で扱います。
          </p>
        )}
      </td>

      <td className={styles.td}>
        <div className={styles.horizontalBlock}>
          {isHolidayAttendanceType ? (
            <p className={styles.syncText}>
              実績状態：通常 / 時間入力なし
            </p>
          ) : syncPlanActual ? (
            <p className={styles.syncText}>
              実績状態：通常 / {selectedPlanType?.name ?? "未選択"}
            </p>
          ) : showActualWorkInput ? (
            <>
              <select
                aria-label={`${row.dayLabel}の実績状態`}
                value={
                  row.actualWorkStatus || ACTUAL_WORK_STATUS_NORMAL
                }
                onChange={(event) =>
                  onChangeRow(
                    row.workDate,
                    "actualWorkStatus",
                    event.target.value,
                  )
                }
                className={`${styles.select} ${styles.actualStatusSelect}`}
                disabled={locked}
              >
                {actualWorkStatusOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>

              <Input
                type="time"
                value={row.actualStartTime}
                onChange={(event) =>
                  onChangeRow(
                    row.workDate,
                    "actualStartTime",
                    event.target.value,
                  )
                }
                disabled={locked || !allowActualTimeInput}
              />

              <Input
                type="time"
                value={row.actualEndTime}
                onChange={(event) =>
                  onChangeRow(
                    row.workDate,
                    "actualEndTime",
                    event.target.value,
                  )
                }
                disabled={locked || !allowActualTimeInput}
              />
            </>
          ) : (
            <p className={styles.syncText}>実績状態：通常</p>
          )}
        </div>
      </td>

      <td className={styles.td}>
        {showScheduledWorkMinutesInput ? (
          <label className={styles.scheduledField}>
            <span className={styles.miniLabel}>分</span>
            <Input
              type="number"
              placeholder="例：480"
              value={row.scheduledWorkMinutes}
              onChange={(event) =>
                onChangeRow(
                  row.workDate,
                  "scheduledWorkMinutes",
                  event.target.value,
                )
              }
              disabled={locked}
            />
            <span className={styles.scheduledHelp}>8時間=480</span>
          </label>
        ) : (
          <p className={styles.noBreakText}>対象外</p>
        )}
      </td>

      <td className={styles.td}>
        {allowBreakInput ? (
          <div className={styles.breakEditArea}>
            <Button
              type="button"
              variant="secondary"
              onClick={() => onAddBreak(row.workDate)}
              disabled={locked}
            >
              休憩追加
            </Button>

            {row.breaks.length === 0 && (
              <p className={styles.noBreakText}>なし</p>
            )}

            {row.breaks.map((breakRow, breakIndex) => (
              <div
                key={`${breakRow.id ?? "new"}-${breakIndex}`}
                className={styles.breakEditRow}
              >
                <Input
                  type="time"
                  value={breakRow.breakStartTime}
                  onChange={(event) =>
                    onChangeBreak(
                      row.workDate,
                      breakIndex,
                      "breakStartTime",
                      event.target.value,
                    )
                  }
                  disabled={locked}
                />

                <Input
                  type="time"
                  value={breakRow.breakEndTime}
                  onChange={(event) =>
                    onChangeBreak(
                      row.workDate,
                      breakIndex,
                      "breakEndTime",
                      event.target.value,
                    )
                  }
                  disabled={locked}
                />

                <Input
                  placeholder="メモ"
                  value={breakRow.breakMemo}
                  onChange={(event) =>
                    onChangeBreak(
                      row.workDate,
                      breakIndex,
                      "breakMemo",
                      event.target.value,
                    )
                  }
                  disabled={locked}
                />

                <Button
                  type="button"
                  variant="danger"
                  onClick={() => onDeleteBreak(row, breakIndex)}
                  disabled={locked}
                >
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
        <div className={styles.inputBlock}>
          {allowTransportInput ? (
            <div className={styles.transportArea}>
              <button
                type="button"
                className={styles.transportSummaryButton}
                onClick={() =>
                  setIsTransportOpen((current) => !current)
                }
              >
                <span>{row.transportExpenses.length}件</span>
                <span>
                  合計 ¥{dailyTransportTotal.toLocaleString()}
                </span>
                <span>{isTransportOpen ? "閉じる" : "開く"}</span>
              </button>

              {isTransportOpen && (
                <div className={styles.transportDetails}>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() =>
                      onAddTransportExpense(row.workDate)
                    }
                    disabled={locked}
                  >
                    交通費追加
                  </Button>

                  {row.transportExpenses.length === 0 && (
                    <p className={styles.noBreakText}>なし</p>
                  )}

                  {row.transportExpenses.map(
                    (transportExpense, transportExpenseIndex) => (
                      <div
                        key={`${
                          transportExpense.id ?? "new"
                        }-${transportExpenseIndex}`}
                        className={styles.transportExpenseRow}
                      >
                        <label className={styles.miniField}>
                          <span className={styles.miniLabel}>
                            出発地
                          </span>
                          <Input
                            placeholder="例：新宿"
                            value={transportExpense.transportFrom}
                            onChange={(event) =>
                              onChangeTransportExpense(
                                row.workDate,
                                transportExpenseIndex,
                                "transportFrom",
                                event.target.value,
                              )
                            }
                            disabled={locked}
                          />
                        </label>

                        <label className={styles.miniField}>
                          <span className={styles.miniLabel}>
                            目的地
                          </span>
                          <Input
                            placeholder="例：品川"
                            value={transportExpense.transportTo}
                            onChange={(event) =>
                              onChangeTransportExpense(
                                row.workDate,
                                transportExpenseIndex,
                                "transportTo",
                                event.target.value,
                              )
                            }
                            disabled={locked}
                          />
                        </label>

                        <label className={styles.miniField}>
                          <span className={styles.miniLabel}>
                            手段
                          </span>
                          <select
                            aria-label={`${row.dayLabel}の交通手段${
                              transportExpenseIndex + 1
                            }`}
                            value={
                              transportExpense.transportMethod
                            }
                            onChange={(event) =>
                              onChangeTransportExpense(
                                row.workDate,
                                transportExpenseIndex,
                                "transportMethod",
                                event.target.value,
                              )
                            }
                            className={styles.select}
                            disabled={locked}
                          >
                            <option value="">選択</option>
                            <option value="電車">電車</option>
                            <option value="バス">バス</option>
                            <option value="車">車</option>
                            <option value="徒歩">徒歩</option>
                            <option value="その他">その他</option>
                          </select>
                        </label>

                        <label className={styles.miniField}>
                          <span className={styles.miniLabel}>
                            金額
                          </span>
                          <Input
                            placeholder="例：320"
                            type="number"
                            value={
                              transportExpense.transportAmount
                            }
                            onChange={(event) =>
                              onChangeTransportExpense(
                                row.workDate,
                                transportExpenseIndex,
                                "transportAmount",
                                event.target.value,
                              )
                            }
                            disabled={locked}
                          />
                        </label>

                        <label
                          className={`${styles.miniField} ${styles.transportMemoField}`}
                        >
                          <span className={styles.miniLabel}>
                            備考
                          </span>
                          <Input
                            placeholder="任意"
                            value={
                              transportExpense.transportMemo
                            }
                            onChange={(event) =>
                              onChangeTransportExpense(
                                row.workDate,
                                transportExpenseIndex,
                                "transportMemo",
                                event.target.value,
                              )
                            }
                            disabled={locked}
                          />
                        </label>

                        <Button
                          type="button"
                          variant="danger"
                          onClick={() =>
                            onDeleteTransportExpense(
                              row,
                              transportExpenseIndex,
                            )
                          }
                          disabled={locked}
                        >
                          削除
                        </Button>
                      </div>
                    ),
                  )}
                </div>
              )}
            </div>
          ) : (
            <p className={styles.noBreakText}>対象外</p>
          )}

          {!isHolidayAttendanceType && (
            <label
              className={`${styles.checkLabel} ${
                locked ? styles.checkLabelDisabled : ""
              }`}
            >
              <input
                type="checkbox"
                checked={row.remoteWorkAllowanceFlag}
                onChange={(event) =>
                  onChangeRow(
                    row.workDate,
                    "remoteWorkAllowanceFlag",
                    event.target.checked,
                  )
                }
                disabled={locked}
              />
              在宅勤務補助あり
            </label>
          )}
        </div>
      </td>

      <td className={styles.td}>
        <span className={getStatusBadgeClass(requiresRequest)}>
          {getStatusBadgeText(requiresRequest)}
        </span>

        <p className={styles.rowMessage}>{rowSystemMessage}</p>

        {requiresRequest && (
          <p className={styles.warningText}>
            月次申請前に確認してください
          </p>
        )}
      </td>

      <td className={styles.td}>
        <div className={styles.actionList}>
          <Button
            type="button"
            variant="danger"
            onClick={() => onDeleteRow(row)}
            disabled={resetDisabled}
          >
            リセット
          </Button>

          {locked && <AttendanceLockedText />}
        </div>
      </td>
    </tr>
  );
}