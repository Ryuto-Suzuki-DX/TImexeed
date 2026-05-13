import type { AttendanceType } from "@/types/user/attendanceType";
import type { AttendanceDay } from "@/types/user/attendanceDay";
import type { AttendanceBreak } from "@/types/user/attendanceBreak";
import type { MonthlyCommuterPass } from "@/types/user/monthlyCommuterPass";
import type {
  AttendanceBreakViewRow,
  AttendanceViewRow,
  CommuterPassViewForm,
} from "@/types/user/attendanceView";
import type {
  UpdateMonthlyAttendanceSaveBreakRequest,
  UpdateMonthlyAttendanceSaveCommuterPassRequest,
  UpdateMonthlyAttendanceSaveDayRequest,
  UpdateMonthlyAttendanceSaveRequest,
} from "@/types/user/monthlyAttendanceSave";
import {
  buildDayLabel,
  buildWeekdayLabel,
  getDaysInMonth,
  shouldUseNextDay,
  toDateOnly,
  toNullableNumber,
  toNullableString,
  toRfc3339,
  toTimeText,
} from "@/utils/attendance/attendanceDate";

/*
 * 従業員勤怠 Mapper
 *
 * API型と画面用型の変換をここに集約する。
 */

/*
 * 対象月の日数分、空の画面Rowを作る
 */
export function buildBlankAttendanceViewRows(targetYear: number, targetMonth: number): AttendanceViewRow[] {
  const daysInMonth = getDaysInMonth(targetYear, targetMonth);
  const rows: AttendanceViewRow[] = [];

  for (let day = 1; day <= daysInMonth; day += 1) {
    const monthText = String(targetMonth).padStart(2, "0");
    const dayText = String(day).padStart(2, "0");
    const workDate = `${targetYear}-${monthText}-${dayText}`;

    rows.push({
      workDate,
      dayLabel: buildDayLabel(targetMonth, day),
      weekday: buildWeekdayLabel(targetYear, targetMonth, day),

      attendanceDayId: null,

      planAttendanceTypeId: 0,
      actualAttendanceTypeId: null,

      commonStartTime: "",
      commonEndTime: "",

      planStartTime: "",
      planEndTime: "",

      actualStartTime: "",
      actualEndTime: "",

      lateFlag: false,
      earlyLeaveFlag: false,
      absenceFlag: false,
      sickLeaveFlag: false,

      remoteWorkAllowanceFlag: false,

      requestMemo: "",

      transportFrom: "",
      transportTo: "",
      transportMethod: "",
      transportAmount: "",

      breaks: [],

      isDirty: false,
    });
  }

  return rows;
}

/*
 * 日別勤怠Rowを初期値に戻す
 *
 * 注意：
 * ・API削除は呼ばない
 * ・画面stateだけを初期化する
 * ・このあと月次勤怠全体保存APIでDBへ反映する
 */
export function resetAttendanceViewRow(row: AttendanceViewRow): AttendanceViewRow {
  return {
    ...row,

    attendanceDayId: null,

    planAttendanceTypeId: 0,
    actualAttendanceTypeId: null,

    commonStartTime: "",
    commonEndTime: "",

    planStartTime: "",
    planEndTime: "",

    actualStartTime: "",
    actualEndTime: "",

    lateFlag: false,
    earlyLeaveFlag: false,
    absenceFlag: false,
    sickLeaveFlag: false,

    remoteWorkAllowanceFlag: false,

    requestMemo: "",

    transportFrom: "",
    transportTo: "",
    transportMethod: "",
    transportAmount: "",

    breaks: [],

    isDirty: true,
  };
}

/*
 * APIの AttendanceDay を画面Rowへ反映する
 */
export function applyAttendanceDayToViewRow(row: AttendanceViewRow, attendanceDay: AttendanceDay): AttendanceViewRow {
  return {
    ...row,

    attendanceDayId: attendanceDay.id,

    planAttendanceTypeId: attendanceDay.planAttendanceTypeId,
    actualAttendanceTypeId: attendanceDay.actualAttendanceTypeId || null,

    commonStartTime: toTimeText(attendanceDay.planStartAt),
    commonEndTime: toTimeText(attendanceDay.planEndAt),

    planStartTime: toTimeText(attendanceDay.planStartAt),
    planEndTime: toTimeText(attendanceDay.planEndAt),

    actualStartTime: toTimeText(attendanceDay.actualStartAt),
    actualEndTime: toTimeText(attendanceDay.actualEndAt),

    lateFlag: attendanceDay.lateFlag,
    earlyLeaveFlag: attendanceDay.earlyLeaveFlag,
    absenceFlag: attendanceDay.absenceFlag,
    sickLeaveFlag: attendanceDay.sickLeaveFlag,

    remoteWorkAllowanceFlag: attendanceDay.remoteWorkAllowanceFlag,

    requestMemo: attendanceDay.requestMemo ?? "",

    transportFrom: attendanceDay.transportFrom ?? "",
    transportTo: attendanceDay.transportTo ?? "",
    transportMethod: attendanceDay.transportMethod ?? "",
    transportAmount: attendanceDay.transportAmount === null ? "" : String(attendanceDay.transportAmount),

    isDirty: false,
  };
}

/*
 * 対象月の空Rowに、APIから取得した勤怠一覧を反映する
 */
export function buildAttendanceViewRows(targetYear: number, targetMonth: number, attendanceDays: AttendanceDay[]): AttendanceViewRow[] {
  const blankRows = buildBlankAttendanceViewRows(targetYear, targetMonth);
  const attendanceDayMap = new Map<string, AttendanceDay>();

  attendanceDays.forEach((attendanceDay) => {
    attendanceDayMap.set(toDateOnly(attendanceDay.workDate), attendanceDay);
  });

  return blankRows.map((row) => {
    const attendanceDay = attendanceDayMap.get(row.workDate);

    if (!attendanceDay) {
      return row;
    }

    return applyAttendanceDayToViewRow(row, attendanceDay);
  });
}

/*
 * APIの AttendanceBreak を画面用休憩Rowへ変換する
 */
export function toAttendanceBreakViewRow(attendanceBreak: AttendanceBreak): AttendanceBreakViewRow {
  return {
    id: attendanceBreak.id,
    breakStartTime: toTimeText(attendanceBreak.breakStartAt),
    breakEndTime: toTimeText(attendanceBreak.breakEndAt),
    breakMemo: attendanceBreak.breakMemo ?? "",
    isNew: false,
    isDirty: false,
  };
}

/*
 * 新規休憩Rowを作る
 */
export function buildNewAttendanceBreakViewRow(): AttendanceBreakViewRow {
  return {
    id: null,
    breakStartTime: "",
    breakEndTime: "",
    breakMemo: "",
    isNew: true,
    isDirty: true,
  };
}

/*
 * Row一覧に休憩一覧を反映する
 */
export function attachBreaksToAttendanceViewRows(rows: AttendanceViewRow[], breakMap: Map<string, AttendanceBreak[]>): AttendanceViewRow[] {
  return rows.map((row) => ({
    ...row,
    breaks: (breakMap.get(row.workDate) ?? []).map(toAttendanceBreakViewRow),
  }));
}

/*
 * 月次通勤定期APIレスポンスを画面Formへ変換する
 */
export function buildCommuterPassViewForm(monthlyCommuterPass: MonthlyCommuterPass | null): CommuterPassViewForm {
  return {
    commuterFrom: monthlyCommuterPass?.commuterFrom ?? "",
    commuterTo: monthlyCommuterPass?.commuterTo ?? "",
    commuterMethod: monthlyCommuterPass?.commuterMethod ?? "",
    commuterAmount:
      monthlyCommuterPass?.commuterAmount === null || monthlyCommuterPass?.commuterAmount === undefined
        ? ""
        : String(monthlyCommuterPass.commuterAmount),
  };
}

/*
 * 月次通勤定期Formを初期値に戻す
 *
 * 注意：
 * ・API削除は呼ばない
 * ・画面stateだけを初期化する
 * ・このあと月次勤怠全体保存APIでDBへ反映する
 */
export function resetCommuterPassViewForm(): CommuterPassViewForm {
  return {
    commuterFrom: "",
    commuterTo: "",
    commuterMethod: "",
    commuterAmount: "",
  };
}

/*
 * 画面用の月次通勤定期Formから月次勤怠全体保存API用Requestを作る
 */
export function buildUpdateMonthlyAttendanceSaveCommuterPassRequest(
  commuterPass: CommuterPassViewForm,
): UpdateMonthlyAttendanceSaveCommuterPassRequest {
  return {
    commuterFrom: toNullableString(commuterPass.commuterFrom),
    commuterTo: toNullableString(commuterPass.commuterTo),
    commuterMethod: toNullableString(commuterPass.commuterMethod),
    commuterAmount: toNullableNumber(commuterPass.commuterAmount),
  };
}

/*
 * 画面用休憩Rowから月次勤怠全体保存API用の休憩Requestを作る
 */
export function buildUpdateMonthlyAttendanceSaveBreakRequest(
  workDate: string,
  breakRow: AttendanceBreakViewRow,
): UpdateMonthlyAttendanceSaveBreakRequest {
  const breakEndUsesNextDay = shouldUseNextDay(breakRow.breakStartTime, breakRow.breakEndTime);

  return {
    breakStartAt: toRfc3339(workDate, breakRow.breakStartTime, false) ?? "",
    breakEndAt: toRfc3339(workDate, breakRow.breakEndTime, breakEndUsesNextDay) ?? "",
    breakMemo: toNullableString(breakRow.breakMemo),
  };
}

/*
 * 画面Rowから月次勤怠全体保存API用の日別勤怠Requestを作る
 */
export function buildUpdateMonthlyAttendanceSaveDayRequest(
  row: AttendanceViewRow,
  selectedPlanType: AttendanceType | null,
): UpdateMonthlyAttendanceSaveDayRequest {
  const breaks = row.breaks.map((breakRow) => buildUpdateMonthlyAttendanceSaveBreakRequest(row.workDate, breakRow));

  /*
   * リセット行
   *
   * 注意：
   * ・削除APIは呼ばない
   * ・画面stateを初期値に戻した行を、月次勤怠全体保存APIへ送る
   * ・バックエンド側で planAttendanceTypeId=0 を初期値戻しとして扱う想定
   */
  if (row.planAttendanceTypeId === 0) {
    return {
      workDate: row.workDate,

      planAttendanceTypeId: 0,
      actualAttendanceTypeId: null,

      commonStartAt: null,
      commonEndAt: null,

      planStartAt: null,
      planEndAt: null,

      actualStartAt: null,
      actualEndAt: null,

      lateFlag: false,
      earlyLeaveFlag: false,
      absenceFlag: false,
      sickLeaveFlag: false,

      remoteWorkAllowanceFlag: false,

      requestMemo: null,

      transportFrom: null,
      transportTo: null,
      transportMethod: null,
      transportAmount: null,

      breaks: [],
    };
  }

  if (!selectedPlanType) {
    throw new Error(`${row.dayLabel} の予定区分を選択してください。`);
  }

  /*
   * 休日だけは予定にも実績にも時間を保存しない。
   */
  if (selectedPlanType.code === "HOLIDAY") {
    return {
      workDate: row.workDate,

      planAttendanceTypeId: row.planAttendanceTypeId,
      actualAttendanceTypeId: null,

      commonStartAt: null,
      commonEndAt: null,

      planStartAt: null,
      planEndAt: null,

      actualStartAt: null,
      actualEndAt: null,

      lateFlag: false,
      earlyLeaveFlag: false,
      absenceFlag: false,
      sickLeaveFlag: false,

      remoteWorkAllowanceFlag: false,

      requestMemo: toNullableString(row.requestMemo),

      transportFrom: null,
      transportTo: null,
      transportMethod: null,
      transportAmount: null,

      breaks: [],
    };
  }

  /*
   * 有給・特別休暇・休職など、予定と実績を同期する区分。
   */
  if (selectedPlanType.syncPlanActual) {
    const commonEndUsesNextDay = shouldUseNextDay(row.commonStartTime, row.commonEndTime);

    return {
      workDate: row.workDate,

      planAttendanceTypeId: row.planAttendanceTypeId,
      actualAttendanceTypeId: null,

      commonStartAt: toRfc3339(row.workDate, row.commonStartTime, false),
      commonEndAt: toRfc3339(row.workDate, row.commonEndTime, commonEndUsesNextDay),

      planStartAt: null,
      planEndAt: null,

      actualStartAt: null,
      actualEndAt: null,

      lateFlag: false,
      earlyLeaveFlag: false,
      absenceFlag: false,
      sickLeaveFlag: false,

      remoteWorkAllowanceFlag: row.remoteWorkAllowanceFlag,

      requestMemo: toNullableString(row.requestMemo),

      transportFrom: toNullableString(row.transportFrom),
      transportTo: toNullableString(row.transportTo),
      transportMethod: toNullableString(row.transportMethod),
      transportAmount: toNullableNumber(row.transportAmount),

      breaks,
    };
  }

  /*
   * 通常勤務・夜勤など、予定と実績を分ける区分。
   */
  const planEndUsesNextDay = shouldUseNextDay(row.planStartTime, row.planEndTime);
  const actualEndUsesNextDay = shouldUseNextDay(row.actualStartTime, row.actualEndTime);

  return {
    workDate: row.workDate,

    planAttendanceTypeId: row.planAttendanceTypeId,
    actualAttendanceTypeId: row.actualAttendanceTypeId,

    commonStartAt: null,
    commonEndAt: null,

    planStartAt: toRfc3339(row.workDate, row.planStartTime, false),
    planEndAt: toRfc3339(row.workDate, row.planEndTime, planEndUsesNextDay),

    actualStartAt: toRfc3339(row.workDate, row.actualStartTime, false),
    actualEndAt: toRfc3339(row.workDate, row.actualEndTime, actualEndUsesNextDay),

    lateFlag: row.lateFlag,
    earlyLeaveFlag: row.earlyLeaveFlag,
    absenceFlag: row.absenceFlag,
    sickLeaveFlag: row.sickLeaveFlag,

    remoteWorkAllowanceFlag: row.remoteWorkAllowanceFlag,

    requestMemo: toNullableString(row.requestMemo),

    transportFrom: toNullableString(row.transportFrom),
    transportTo: toNullableString(row.transportTo),
    transportMethod: toNullableString(row.transportMethod),
    transportAmount: toNullableNumber(row.transportAmount),

    breaks,
  };
}

/*
 * 月次勤怠画面のstateから月次勤怠全体保存API用Requestを作る
 */
export function buildUpdateMonthlyAttendanceSaveRequest(
  targetYear: number,
  targetMonth: number,
  commuterPass: CommuterPassViewForm,
  attendanceRows: AttendanceViewRow[],
  attendanceTypes: AttendanceType[],
): UpdateMonthlyAttendanceSaveRequest {
  const attendanceDays = attendanceRows.map((row) => {
    const selectedPlanType =
      attendanceTypes.find((attendanceType) => attendanceType.id === row.planAttendanceTypeId) ?? null;

    return buildUpdateMonthlyAttendanceSaveDayRequest(row, selectedPlanType);
  });

  return {
    targetYear,
    targetMonth,
    commuterPass: buildUpdateMonthlyAttendanceSaveCommuterPassRequest(commuterPass),
    attendanceDays,
  };
}