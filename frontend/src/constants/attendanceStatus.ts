export const ACTUAL_WORK_STATUS = {
  NORMAL: "",
  ABSENCE: "absence",
  SICK_LEAVE: "sickLeave",
  LATE: "late",
  EARLY_LEAVE: "earlyLeave",
} as const;

export const ACTUAL_WORK_STATUS_OPTIONS = [
  { value: ACTUAL_WORK_STATUS.NORMAL, label: "通常" },
  { value: ACTUAL_WORK_STATUS.ABSENCE, label: "欠勤" },
  { value: ACTUAL_WORK_STATUS.SICK_LEAVE, label: "病欠" },
  { value: ACTUAL_WORK_STATUS.LATE, label: "遅刻" },
  { value: ACTUAL_WORK_STATUS.EARLY_LEAVE, label: "早退" },
] as const;

export type ActualWorkStatus =
  (typeof ACTUAL_WORK_STATUS_OPTIONS)[number]["value"];