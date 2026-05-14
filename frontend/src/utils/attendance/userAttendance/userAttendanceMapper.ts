import type { AttendanceType } from "@/types/user/attendanceType";
import type { AttendanceDay } from "@/types/user/attendanceDay";
import type { AttendanceBreak } from "@/types/user/attendanceBreak";
import type { HolidayDate } from "@/types/user/holidayDate";
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
 * 祝日一覧から日付ごとの祝日名Mapを作る
 */
function buildHolidayNameMap(holidayDates: HolidayDate[]): Map<string, string> {
  const holidayNameMap = new Map<string, string>();

  holidayDates.forEach((holidayDate) => {
    holidayNameMap.set(toDateOnly(holidayDate.holidayDate), holidayDate.holidayName);
  });

  return holidayNameMap;
}

/*
 * 対象月の日数分、空の画面Rowを作る
 */
export function buildBlankAttendanceViewRows(
  targetYear: number,
  targetMonth: number,
  holidayDates: HolidayDate[] = [],
): AttendanceViewRow[] {
  const daysInMonth = getDaysInMonth(targetYear, targetMonth);
  const rows: AttendanceViewRow[] = [];
  const holidayNameMap = buildHolidayNameMap(holidayDates);

  for (let day = 1; day <= daysInMonth; day += 1) {
    const monthText = String(targetMonth).padStart(2, "0");
    const dayText = String(day).padStart(2, "0");
    const workDate = `${targetYear}-${monthText}-${dayText}`;
    const holidayName = holidayNameMap.get(workDate) ?? null;

    rows.push({
      workDate,
      dayLabel: buildDayLabel(targetMonth, day),
      weekday: buildWeekdayLabel(targetYear, targetMonth, day),

      isHoliday: holidayName !== null,
      holidayName,

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
 * ・祝日情報は画面表示用なので維持する
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
export function applyAttendanceDayToViewRow(
  row: AttendanceViewRow,
  attendanceDay: AttendanceDay,
): AttendanceViewRow {
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

    transportFrom: attendanceDay.transportFrom ?? "",
    transportTo: attendanceDay.transportTo ?? "",
    transportMethod: attendanceDay.transportMethod ?? "",
    transportAmount: attendanceDay.transportAmount === null ? "" : String(attendanceDay.transportAmount),

    isDirty: false,
  };
}

/*
 * 対象月の空Rowに、APIから取得した勤怠一覧と祝日一覧を反映する
 */
export function buildAttendanceViewRows(
  targetYear: number,
  targetMonth: number,
  attendanceDays: AttendanceDay[],
  holidayDates: HolidayDate[] = [],
): AttendanceViewRow[] {
  const blankRows = buildBlankAttendanceViewRows(targetYear, targetMonth, holidayDates);
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
export function attachBreaksToAttendanceViewRows(
  rows: AttendanceViewRow[],
  breakMap: Map<string, AttendanceBreak[]>,
): AttendanceViewRow[] {
  return rows.map((row) => ({
    ...row,
    breaks: (breakMap.get(row.workDate) ?? []).map(toAttendanceBreakViewRow),
  }));
}

/*
 * 月次通勤定期APIレスポンスを画面Formへ変換する
 */
export function buildCommuterPassViewForm(
  monthlyCommuterPass: MonthlyCommuterPass | null,
): CommuterPassViewForm {
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
   * 実績区分IDは予定区分IDと同じ値を送る。
   */
  if (selectedPlanType.code === "HOLIDAY") {
    return {
      workDate: row.workDate,

      planAttendanceTypeId: row.planAttendanceTypeId,
      actualAttendanceTypeId: row.planAttendanceTypeId,

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

      transportFrom: null,
      transportTo: null,
      transportMethod: null,
      transportAmount: null,

      breaks: [],
    };
  }

  /*
   * 有給・特別休暇・休職など、予定と実績を同期する区分。
   * 実績区分IDは予定区分IDと同じ値を送る。
   */
  if (selectedPlanType.syncPlanActual) {
    const commonEndUsesNextDay = shouldUseNextDay(row.commonStartTime, row.commonEndTime);

    return {
      workDate: row.workDate,

      planAttendanceTypeId: row.planAttendanceTypeId,
      actualAttendanceTypeId: row.planAttendanceTypeId,

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

      transportFrom: toNullableString(row.transportFrom),
      transportTo: toNullableString(row.transportTo),
      transportMethod: toNullableString(row.transportMethod),
      transportAmount: toNullableNumber(row.transportAmount),

      breaks,
    };
  }

  /*
   * 通常勤務など、予定時間と実績時間を分ける区分。
   *
   * 実績区分IDは予定区分IDと同じ値を送る。
   * 欠勤、病欠、遅刻、早退は actualAttendanceTypeId ではなく、
   * lateFlag / earlyLeaveFlag / absenceFlag / sickLeaveFlag で表現する。
   *
   * 夜勤は勤務区分ではない。
   * 深夜時間は actualStartAt / actualEndAt から集計時に計算する。
   */
  const planEndUsesNextDay = shouldUseNextDay(row.planStartTime, row.planEndTime);
  const actualEndUsesNextDay = shouldUseNextDay(row.actualStartTime, row.actualEndTime);

  return {
    workDate: row.workDate,

    planAttendanceTypeId: row.planAttendanceTypeId,
    actualAttendanceTypeId: row.planAttendanceTypeId,

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
