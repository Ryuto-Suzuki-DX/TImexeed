"use client";

import { useMemo, useState } from "react";
import ConfirmModal from "@/components/atoms/ConfirmModal";
import UserSideMenu from "@/components/user/navigation/UserSideMenu";

/*
 * 勤怠ページ
 *
 * 初期実装方針
 * - API連携なし
 * - DB保存なし
 * - モックデータで画面と操作感を先に作る
 * - 予定と実績を左右表示する
 * - 休憩は予定/実績ともに複数登録できる
 * - 夜勤など日付をまたぐ勤怠はバーで分割表示する
 * - 空白日は休日/未入力として表示する
 * - 勤怠バークリックで編集モーダル
 * - 空白日クリックで新規モーダル
 * - 削除時は ConfirmModal を使う
 */

type AttendanceType =
  | "WORK"
  | "PAID_LEAVE"
  | "ABSENCE"
  | "SICK_LEAVE"
  | "LATE"
  | "EARLY_LEAVE"
  | "LATE_AND_EARLY_LEAVE"
  | "HOLIDAY_WORK";

type AttendanceStatus =
  | "DRAFT"
  | "REQUESTED"
  | "APPROVED"
  | "REJECTED"
  | "CANCELLED";

type BreakType = "SCHEDULED" | "ACTUAL";

type AttendanceBreak = {
  id: number;
  breakType: BreakType;
  breakStartAt: string;
  breakEndAt: string;
};

type Attendance = {
  id: number;
  workDate: string;
  workSiteName: string;
  shiftPatternName: string;
  attendanceType: AttendanceType;

  scheduledStartAt: string;
  scheduledEndAt: string;

  actualStartAt: string | null;
  actualEndAt: string | null;

  breaks: AttendanceBreak[];

  memo: string;
  status: AttendanceStatus;
};

type AttendanceForm = {
  id: number | null;
  workDate: string;
  workSiteName: string;
  shiftPatternName: string;
  attendanceType: AttendanceType;

  scheduledStartAt: string;
  scheduledEndAt: string;

  actualStartAt: string;
  actualEndAt: string;

  breaks: AttendanceBreak[];

  memo: string;
  status: AttendanceStatus;
};

type CalendarBarKind = "SCHEDULED" | "ACTUAL";

type CalendarBar = {
  attendanceId: number;
  displayDate: string;
  kind: CalendarBarKind;
  label: string;
  startAt: string;
  endAt: string;
  continuesFromPreviousDay: boolean;
  continuesToNextDay: boolean;
};

const ATTENDANCE_TYPE_LABEL: Record<AttendanceType, string> = {
  WORK: "出勤",
  PAID_LEAVE: "有給",
  ABSENCE: "欠勤",
  SICK_LEAVE: "病欠",
  LATE: "遅刻",
  EARLY_LEAVE: "早退",
  LATE_AND_EARLY_LEAVE: "遅刻早退",
  HOLIDAY_WORK: "休日出勤",
};

const STATUS_LABEL: Record<AttendanceStatus, string> = {
  DRAFT: "下書き",
  REQUESTED: "申請中",
  APPROVED: "承認済",
  REJECTED: "否認",
  CANCELLED: "取消",
};

const SHIFT_PRESETS = [
  {
    name: "日勤",
    scheduledStartTime: "09:00",
    scheduledEndTime: "18:00",
    scheduledBreakStartTime: "12:00",
    scheduledBreakEndTime: "13:00",
    crossesDay: false,
  },
  {
    name: "早番",
    scheduledStartTime: "07:00",
    scheduledEndTime: "16:00",
    scheduledBreakStartTime: "11:00",
    scheduledBreakEndTime: "12:00",
    crossesDay: false,
  },
  {
    name: "遅番",
    scheduledStartTime: "13:00",
    scheduledEndTime: "22:00",
    scheduledBreakStartTime: "17:00",
    scheduledBreakEndTime: "18:00",
    crossesDay: false,
  },
  {
    name: "夜勤",
    scheduledStartTime: "22:00",
    scheduledEndTime: "07:00",
    scheduledBreakStartTime: "02:00",
    scheduledBreakEndTime: "03:00",
    crossesDay: true,
    breakCrossesDay: true,
  },
];

const initialAttendances: Attendance[] = [
  {
    id: 1,
    workDate: "2026-05-01",
    workSiteName: "A病院",
    shiftPatternName: "日勤",
    attendanceType: "WORK",
    scheduledStartAt: "2026-05-01T09:00",
    scheduledEndAt: "2026-05-01T18:00",
    actualStartAt: "2026-05-01T09:00",
    actualEndAt: "2026-05-01T18:00",
    breaks: [
      {
        id: 1,
        breakType: "SCHEDULED",
        breakStartAt: "2026-05-01T12:00",
        breakEndAt: "2026-05-01T13:00",
      },
      {
        id: 2,
        breakType: "ACTUAL",
        breakStartAt: "2026-05-01T12:00",
        breakEndAt: "2026-05-01T13:00",
      },
    ],
    memo: "",
    status: "DRAFT",
  },
  {
    id: 2,
    workDate: "2026-05-02",
    workSiteName: "A病院",
    shiftPatternName: "夜勤",
    attendanceType: "WORK",
    scheduledStartAt: "2026-05-02T22:00",
    scheduledEndAt: "2026-05-03T07:00",
    actualStartAt: "2026-05-02T22:00",
    actualEndAt: "2026-05-03T07:30",
    breaks: [
      {
        id: 1,
        breakType: "SCHEDULED",
        breakStartAt: "2026-05-03T02:00",
        breakEndAt: "2026-05-03T03:00",
      },
      {
        id: 2,
        breakType: "ACTUAL",
        breakStartAt: "2026-05-03T02:00",
        breakEndAt: "2026-05-03T03:00",
      },
    ],
    memo: "夜勤。30分残業",
    status: "DRAFT",
  },
  {
    id: 3,
    workDate: "2026-05-04",
    workSiteName: "A病院",
    shiftPatternName: "日勤",
    attendanceType: "PAID_LEAVE",
    scheduledStartAt: "2026-05-04T09:00",
    scheduledEndAt: "2026-05-04T18:00",
    actualStartAt: null,
    actualEndAt: null,
    breaks: [
      {
        id: 1,
        breakType: "SCHEDULED",
        breakStartAt: "2026-05-04T12:00",
        breakEndAt: "2026-05-04T13:00",
      },
    ],
    memo: "有給申請",
    status: "REQUESTED",
  },
];

function formatDateKey(date: Date) {
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function formatDateTimeLocal(date: Date) {
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  const hour = `${date.getHours()}`.padStart(2, "0");
  const minute = `${date.getMinutes()}`.padStart(2, "0");
  return `${year}-${month}-${day}T${hour}:${minute}`;
}

function getDaysInMonth(year: number, month: number) {
  const lastDate = new Date(year, month, 0).getDate();

  return Array.from({ length: lastDate }, (_, index) => {
    const date = new Date(year, month - 1, index + 1);
    return {
      date,
      dateKey: formatDateKey(date),
    };
  });
}

function getDayLabel(date: Date) {
  const labels = ["日", "月", "火", "水", "木", "金", "土"];
  return labels[date.getDay()];
}

function addDays(dateKey: string, days: number) {
  const date = new Date(`${dateKey}T00:00`);
  date.setDate(date.getDate() + days);
  return formatDateKey(date);
}

function startOfDate(dateKey: string) {
  return new Date(`${dateKey}T00:00`);
}

function endOfDate(dateKey: string) {
  return new Date(`${dateKey}T23:59:59`);
}

function toDateTimeLocalValue(workDate: string, time: string, addDay: boolean) {
  const date = new Date(`${workDate}T00:00`);

  if (addDay) {
    date.setDate(date.getDate() + 1);
  }

  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");

  return `${year}-${month}-${day}T${time}`;
}

function toMinutesFromStartOfDay(dateTime: string) {
  const date = new Date(dateTime);
  return date.getHours() * 60 + date.getMinutes();
}

function formatTime(dateTime: string | null) {
  if (!dateTime) {
    return "";
  }

  const date = new Date(dateTime);
  return `${`${date.getHours()}`.padStart(2, "0")}:${`${date.getMinutes()}`.padStart(2, "0")}`;
}

function formatTimeRange(startAt: string | null, endAt: string | null) {
  if (!startAt || !endAt) {
    return "";
  }

  const start = new Date(startAt);
  const end = new Date(endAt);
  const isNextDay = start.toDateString() !== end.toDateString();

  return `${formatTime(startAt)}〜${isNextDay ? "翌" : ""}${formatTime(endAt)}`;
}

function isActualTimeType(attendanceType: AttendanceType) {
  return (
    attendanceType === "WORK" ||
    attendanceType === "LATE" ||
    attendanceType === "EARLY_LEAVE" ||
    attendanceType === "LATE_AND_EARLY_LEAVE" ||
    attendanceType === "HOLIDAY_WORK"
  );
}

function getBreaksByType(breaks: AttendanceBreak[], breakType: BreakType) {
  return breaks.filter((breakTime) => breakTime.breakType === breakType);
}

function createPresetBreaks(workDate: string, shiftPatternName: string): AttendanceBreak[] {
  const preset = SHIFT_PRESETS.find((shift) => shift.name === shiftPatternName) ?? SHIFT_PRESETS[0];

  const scheduledBreakStartAt = toDateTimeLocalValue(
    workDate,
    preset.scheduledBreakStartTime,
    Boolean(preset.breakCrossesDay),
  );

  const scheduledBreakEndAt = toDateTimeLocalValue(
    workDate,
    preset.scheduledBreakEndTime,
    Boolean(preset.breakCrossesDay),
  );

  return [
    {
      id: 1,
      breakType: "SCHEDULED",
      breakStartAt: scheduledBreakStartAt,
      breakEndAt: scheduledBreakEndAt,
    },
    {
      id: 2,
      breakType: "ACTUAL",
      breakStartAt: scheduledBreakStartAt,
      breakEndAt: scheduledBreakEndAt,
    },
  ];
}

function createEmptyForm(workDate: string): AttendanceForm {
  return {
    id: null,
    workDate,
    workSiteName: "A病院",
    shiftPatternName: "日勤",
    attendanceType: "WORK",
    scheduledStartAt: `${workDate}T09:00`,
    scheduledEndAt: `${workDate}T18:00`,
    actualStartAt: `${workDate}T09:00`,
    actualEndAt: `${workDate}T18:00`,
    breaks: createPresetBreaks(workDate, "日勤"),
    memo: "",
    status: "DRAFT",
  };
}

function attendanceToForm(attendance: Attendance): AttendanceForm {
  return {
    id: attendance.id,
    workDate: attendance.workDate,
    workSiteName: attendance.workSiteName,
    shiftPatternName: attendance.shiftPatternName,
    attendanceType: attendance.attendanceType,
    scheduledStartAt: attendance.scheduledStartAt,
    scheduledEndAt: attendance.scheduledEndAt,
    actualStartAt: attendance.actualStartAt ?? "",
    actualEndAt: attendance.actualEndAt ?? "",
    breaks: attendance.breaks,
    memo: attendance.memo,
    status: attendance.status,
  };
}

function formToAttendance(form: AttendanceForm, nextId: number): Attendance {
  const shouldUseActual = isActualTimeType(form.attendanceType);

  return {
    id: form.id ?? nextId,
    workDate: form.workDate,
    workSiteName: form.workSiteName,
    shiftPatternName: form.shiftPatternName,
    attendanceType: form.attendanceType,
    scheduledStartAt: form.scheduledStartAt,
    scheduledEndAt: form.scheduledEndAt,
    actualStartAt: shouldUseActual ? form.actualStartAt : null,
    actualEndAt: shouldUseActual ? form.actualEndAt : null,
    breaks: shouldUseActual
      ? form.breaks
      : form.breaks.filter((breakTime) => breakTime.breakType === "SCHEDULED"),
    memo: form.memo,
    status: form.status,
  };
}

function createCalendarBars(attendances: Attendance[]): CalendarBar[] {
  const bars: CalendarBar[] = [];

  attendances.forEach((attendance) => {
    bars.push(
      ...createBarsForRange({
        attendance,
        kind: "SCHEDULED",
        startAt: attendance.scheduledStartAt,
        endAt: attendance.scheduledEndAt,
        label: `${attendance.shiftPatternName} ${formatTimeRange(
          attendance.scheduledStartAt,
          attendance.scheduledEndAt,
        )}`,
      }),
    );

    const actualDisplayStartAt = isActualTimeType(attendance.attendanceType)
      ? attendance.actualStartAt
      : attendance.scheduledStartAt;

    const actualDisplayEndAt = isActualTimeType(attendance.attendanceType)
      ? attendance.actualEndAt
      : attendance.scheduledEndAt;

    if (actualDisplayStartAt && actualDisplayEndAt) {
      bars.push(
        ...createBarsForRange({
          attendance,
          kind: "ACTUAL",
          startAt: actualDisplayStartAt,
          endAt: actualDisplayEndAt,
          label: `${ATTENDANCE_TYPE_LABEL[attendance.attendanceType]} ${formatTimeRange(
            actualDisplayStartAt,
            actualDisplayEndAt,
          )}`,
        }),
      );
    }
  });

  return bars;
}

function createBarsForRange({
  attendance,
  kind,
  startAt,
  endAt,
  label,
}: {
  attendance: Attendance;
  kind: CalendarBarKind;
  startAt: string;
  endAt: string;
  label: string;
}): CalendarBar[] {
  const result: CalendarBar[] = [];

  const start = new Date(startAt);
  const end = new Date(endAt);

  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime()) || start >= end) {
    return result;
  }

  let currentDateKey = formatDateKey(start);

  while (new Date(`${currentDateKey}T00:00`) <= end) {
    const displayDateStart = startOfDate(currentDateKey);
    const displayDateEnd = endOfDate(currentDateKey);

    const segmentStart = start > displayDateStart ? start : displayDateStart;
    const segmentEnd = end < displayDateEnd ? end : displayDateEnd;

    if (segmentStart < segmentEnd) {
      result.push({
        attendanceId: attendance.id,
        displayDate: currentDateKey,
        kind,
        label,
        startAt: formatDateTimeLocal(segmentStart),
        endAt: formatDateTimeLocal(segmentEnd),
        continuesFromPreviousDay: start < displayDateStart,
        continuesToNextDay: end > displayDateEnd,
      });
    }

    currentDateKey = addDays(currentDateKey, 1);
  }

  return result;
}

export default function UserAttendancePage() {
  const today = new Date();

  const [displayYear, setDisplayYear] = useState(today.getFullYear());
  const [displayMonth, setDisplayMonth] = useState(today.getMonth() + 1);
  const [attendances, setAttendances] = useState<Attendance[]>(initialAttendances);

  const [isAttendanceModalOpen, setIsAttendanceModalOpen] = useState(false);
  const [form, setForm] = useState<AttendanceForm>(createEmptyForm(formatDateKey(today)));

  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);

  const days = useMemo(() => {
    return getDaysInMonth(displayYear, displayMonth);
  }, [displayYear, displayMonth]);

  const attendanceMap = useMemo(() => {
    return attendances.reduce<Record<string, Attendance>>((map, attendance) => {
      map[attendance.workDate] = attendance;
      return map;
    }, {});
  }, [attendances]);

  const calendarBarsByDate = useMemo(() => {
    const bars = createCalendarBars(attendances);

    return bars.reduce<Record<string, CalendarBar[]>>((map, bar) => {
      if (!map[bar.displayDate]) {
        map[bar.displayDate] = [];
      }

      map[bar.displayDate].push(bar);
      return map;
    }, {});
  }, [attendances]);

  const modalTitle = form.id ? "勤怠編集" : "勤怠新規作成";
  const isApproved = form.status === "APPROVED";
  const actualFieldsDisabled = !isActualTimeType(form.attendanceType);

  const scheduledBreaks = getBreaksByType(form.breaks, "SCHEDULED");
  const actualBreaks = getBreaksByType(form.breaks, "ACTUAL");

  function handlePrevMonth() {
    if (displayMonth === 1) {
      setDisplayYear(displayYear - 1);
      setDisplayMonth(12);
      return;
    }

    setDisplayMonth(displayMonth - 1);
  }

  function handleNextMonth() {
    if (displayMonth === 12) {
      setDisplayYear(displayYear + 1);
      setDisplayMonth(1);
      return;
    }

    setDisplayMonth(displayMonth + 1);
  }

  function handleOpenNew(workDate: string) {
    setForm(createEmptyForm(workDate));
    setIsAttendanceModalOpen(true);
  }

  function handleOpenNewToday() {
    handleOpenNew(formatDateKey(today));
  }

  function handleOpenEdit(attendance: Attendance) {
    setForm(attendanceToForm(attendance));
    setIsAttendanceModalOpen(true);
  }

  function handleCloseModal() {
    setIsAttendanceModalOpen(false);
    setIsDeleteConfirmOpen(false);
  }

  function handleChangeForm<K extends keyof AttendanceForm>(
    key: K,
    value: AttendanceForm[K],
  ) {
    setForm((prev) => ({
      ...prev,
      [key]: value,
    }));
  }

  function handleChangeAttendanceType(value: AttendanceType) {
    const shouldUseActual = isActualTimeType(value);

    setForm((prev) => ({
      ...prev,
      attendanceType: value,
      actualStartAt: shouldUseActual ? prev.scheduledStartAt : "",
      actualEndAt: shouldUseActual ? prev.scheduledEndAt : "",
      breaks: shouldUseActual
        ? ensureActualBreaks(prev.breaks)
        : prev.breaks.filter((breakTime) => breakTime.breakType === "SCHEDULED"),
    }));
  }

  function ensureActualBreaks(breaks: AttendanceBreak[]) {
    const scheduled = breaks.filter((breakTime) => breakTime.breakType === "SCHEDULED");
    const actual = breaks.filter((breakTime) => breakTime.breakType === "ACTUAL");

    if (actual.length > 0) {
      return breaks;
    }

    const copiedActual = scheduled.map((breakTime, index) => ({
      id: Date.now() + index,
      breakType: "ACTUAL" as const,
      breakStartAt: breakTime.breakStartAt,
      breakEndAt: breakTime.breakEndAt,
    }));

    return [...breaks, ...copiedActual];
  }

  function handleChangeShiftPreset(shiftName: string) {
    const preset = SHIFT_PRESETS.find((shift) => shift.name === shiftName);

    if (!preset) {
      return;
    }

    const scheduledStartAt = toDateTimeLocalValue(
      form.workDate,
      preset.scheduledStartTime,
      false,
    );

    const scheduledEndAt = toDateTimeLocalValue(
      form.workDate,
      preset.scheduledEndTime,
      preset.crossesDay,
    );

    const presetBreaks = createPresetBreaks(form.workDate, preset.name);
    const shouldUseActual = isActualTimeType(form.attendanceType);

    setForm((prev) => ({
      ...prev,
      shiftPatternName: preset.name,
      scheduledStartAt,
      scheduledEndAt,
      actualStartAt: shouldUseActual ? scheduledStartAt : "",
      actualEndAt: shouldUseActual ? scheduledEndAt : "",
      breaks: shouldUseActual
        ? presetBreaks
        : presetBreaks.filter((breakTime) => breakTime.breakType === "SCHEDULED"),
    }));
  }

  function handleChangeBreak(
    breakId: number,
    key: "breakStartAt" | "breakEndAt",
    value: string,
  ) {
    setForm((prev) => ({
      ...prev,
      breaks: prev.breaks.map((breakTime) => {
        if (breakTime.id !== breakId) {
          return breakTime;
        }

        return {
          ...breakTime,
          [key]: value,
        };
      }),
    }));
  }

  function handleAddBreak(breakType: BreakType) {
    setForm((prev) => ({
      ...prev,
      breaks: [
        ...prev.breaks,
        {
          id: Date.now(),
          breakType,
          breakStartAt: `${prev.workDate}T12:00`,
          breakEndAt: `${prev.workDate}T13:00`,
        },
      ],
    }));
  }

  function handleRemoveBreak(breakId: number) {
    setForm((prev) => ({
      ...prev,
      breaks: prev.breaks.filter((breakTime) => breakTime.id !== breakId),
    }));
  }

  function handleSave() {
    if (isApproved) {
      return;
    }

    const nextId =
      attendances.length === 0
        ? 1
        : Math.max(...attendances.map((attendance) => attendance.id)) + 1;

    const savedAttendance = formToAttendance(form, nextId);

    setAttendances((prev) => {
      const exists = prev.some((attendance) => attendance.id === savedAttendance.id);

      if (exists) {
        return prev.map((attendance) =>
          attendance.id === savedAttendance.id ? savedAttendance : attendance,
        );
      }

      return [...prev, savedAttendance];
    });

    handleCloseModal();
  }

  function handleDelete() {
    if (!form.id || isApproved) {
      return;
    }

    setAttendances((prev) =>
      prev.filter((attendance) => attendance.id !== form.id),
    );

    handleCloseModal();
  }

  function renderBreakInputs(breaks: AttendanceBreak[], breakType: BreakType) {
    return (
      <section style={{ border: "1px solid #e5e7eb", borderRadius: "10px", padding: "16px", backgroundColor: "#ffffff" }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "12px" }}>
          <h3 style={{ margin: 0, fontSize: "16px" }}>
            {breakType === "SCHEDULED" ? "予定休憩" : "実績休憩"}
          </h3>

          <button
            type="button"
            disabled={isApproved || (breakType === "ACTUAL" && actualFieldsDisabled)}
            onClick={() => handleAddBreak(breakType)}
            style={{
              border: "1px solid #ddd",
              background: "#fff",
              padding: "8px 12px",
              borderRadius: "8px",
              cursor: isApproved || (breakType === "ACTUAL" && actualFieldsDisabled) ? "not-allowed" : "pointer",
              fontWeight: 700,
            }}
          >
            休憩追加
          </button>
        </div>

        {breaks.length === 0 ? (
          <div style={{ fontSize: "13px", color: "#777" }}>
            休憩なし
          </div>
        ) : (
          <div style={{ display: "flex", flexDirection: "column", gap: "10px" }}>
            {breaks.map((breakTime) => (
              <div key={breakTime.id} style={{ display: "grid", gridTemplateColumns: "1fr 1fr auto", gap: "10px", alignItems: "end" }}>
                <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "14px", fontWeight: 700 }}>
                  休憩開始
                  <input
                    type="datetime-local"
                    value={breakTime.breakStartAt}
                    disabled={isApproved || (breakType === "ACTUAL" && actualFieldsDisabled)}
                    onChange={(event) => handleChangeBreak(breakTime.id, "breakStartAt", event.target.value)}
                    style={{ padding: "10px", border: "1px solid #ddd", borderRadius: "8px" }}
                  />
                </label>

                <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "14px", fontWeight: 700 }}>
                  休憩終了
                  <input
                    type="datetime-local"
                    value={breakTime.breakEndAt}
                    disabled={isApproved || (breakType === "ACTUAL" && actualFieldsDisabled)}
                    onChange={(event) => handleChangeBreak(breakTime.id, "breakEndAt", event.target.value)}
                    style={{ padding: "10px", border: "1px solid #ddd", borderRadius: "8px" }}
                  />
                </label>

                <button
                  type="button"
                  disabled={isApproved || (breakType === "ACTUAL" && actualFieldsDisabled)}
                  onClick={() => handleRemoveBreak(breakTime.id)}
                  style={{
                    border: "none",
                    background: "#ef4444",
                    color: "#fff",
                    padding: "10px 12px",
                    borderRadius: "8px",
                    cursor: isApproved || (breakType === "ACTUAL" && actualFieldsDisabled) ? "not-allowed" : "pointer",
                    fontWeight: 700,
                  }}
                >
                  削除
                </button>
              </div>
            ))}
          </div>
        )}
      </section>
    );
  }

  return (
    <div style={{ minHeight: "100vh", background: "#f5f6f8" }}>
      <UserSideMenu />

      <main style={{ padding: "24px" }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "20px" }}>
          <div>
            <h1 style={{ margin: 0, fontSize: "24px", fontWeight: 700 }}>
              勤怠登録
            </h1>
            <p style={{ margin: "6px 0 0", color: "#666", fontSize: "14px" }}>
              予定と実績を登録します。夜勤は日付をまたいでバー表示します。
            </p>
          </div>

          <button
            type="button"
            onClick={handleOpenNewToday}
            style={{ border: "none", background: "#f97316", color: "#fff", padding: "10px 18px", borderRadius: "8px", cursor: "pointer", fontWeight: 700 }}
          >
            新規作成
          </button>
        </div>

        <section style={{ background: "#fff", borderRadius: "12px", padding: "16px", boxShadow: "0 1px 4px rgba(0,0,0,0.08)" }}>
          <div style={{ display: "flex", justifyContent: "center", alignItems: "center", gap: "16px", marginBottom: "16px" }}>
            <button
              type="button"
              onClick={handlePrevMonth}
              style={{ border: "1px solid #ddd", background: "#fff", padding: "8px 14px", borderRadius: "8px", cursor: "pointer", fontWeight: 700 }}
            >
              前月
            </button>

            <strong style={{ fontSize: "20px" }}>
              {displayYear}年{displayMonth}月
            </strong>

            <button
              type="button"
              onClick={handleNextMonth}
              style={{ border: "1px solid #ddd", background: "#fff", padding: "8px 14px", borderRadius: "8px", cursor: "pointer", fontWeight: 700 }}
            >
              次月
            </button>
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: "10px" }}>
            {days.map(({ date, dateKey }) => {
              const attendance = attendanceMap[dateKey];
              const bars = calendarBarsByDate[dateKey] ?? [];

              const isSunday = date.getDay() === 0;
              const isSaturday = date.getDay() === 6;

              return (
                <div
                  key={dateKey}
                  onClick={() => {
                    if (bars.length === 0 && !attendance) {
                      handleOpenNew(dateKey);
                    }
                  }}
                  style={{
                    border: "1px solid #e5e7eb",
                    borderRadius: "10px",
                    padding: "14px",
                    background: bars.length > 0 ? "#fff" : "#fafafa",
                    cursor: bars.length > 0 ? "default" : "pointer",
                  }}
                >
                  <div style={{ display: "flex", alignItems: "center", marginBottom: "10px", gap: "8px" }}>
                    <strong style={{ fontSize: "16px", color: isSunday ? "#dc2626" : isSaturday ? "#2563eb" : "#111827" }}>
                      {displayMonth}/{date.getDate()}（{getDayLabel(date)}）
                    </strong>

                    {bars.length === 0 && (
                      <span style={{ fontSize: "12px", color: "#777" }}>
                        休日 / 未入力
                      </span>
                    )}

                    {bars.some((bar) => bar.continuesFromPreviousDay) && (
                      <span style={{ fontSize: "12px", color: "#6b7280", background: "#f3f4f6", padding: "2px 8px", borderRadius: "999px" }}>
                        前日から継続
                      </span>
                    )}

                    {bars.some((bar) => bar.continuesToNextDay) && (
                      <span style={{ fontSize: "12px", color: "#6b7280", background: "#f3f4f6", padding: "2px 8px", borderRadius: "999px" }}>
                        翌日へ継続
                      </span>
                    )}
                  </div>

                  {bars.length > 0 ? (
                    <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
                      {bars.map((bar) => {
                        const startMinutes = toMinutesFromStartOfDay(bar.startAt);
                        const endMinutes = toMinutesFromStartOfDay(bar.endAt);

                        const leftPercent = (startMinutes / 1440) * 100;
                        const widthPercent = Math.max(((endMinutes - startMinutes) / 1440) * 100, 2);

                        const targetAttendance = attendances.find(
                          (item) => item.id === bar.attendanceId,
                        );

                        return (
                          <button
                            key={`${bar.attendanceId}-${bar.kind}-${bar.displayDate}-${bar.startAt}`}
                            type="button"
                            onClick={(event) => {
                              event.stopPropagation();

                              if (targetAttendance) {
                                handleOpenEdit(targetAttendance);
                              }
                            }}
                            style={{
                              border: "none",
                              background: "transparent",
                              padding: 0,
                              cursor: "pointer",
                              textAlign: "left",
                            }}
                          >
                            <div style={{ display: "grid", gridTemplateColumns: "64px 1fr", gap: "10px", alignItems: "center" }}>
                              <div style={{ fontSize: "12px", fontWeight: 700, color: bar.kind === "SCHEDULED" ? "#9a3412" : "#1d4ed8" }}>
                                {bar.kind === "SCHEDULED" ? "予定" : "実績"}
                              </div>

                              <div style={{ position: "relative", height: "34px", background: "#f3f4f6", borderRadius: "999px", overflow: "hidden" }}>
                                <div
                                  style={{
                                    position: "absolute",
                                    top: "6px",
                                    left: `${leftPercent}%`,
                                    width: `${widthPercent}%`,
                                    height: "22px",
                                    borderRadius: "999px",
                                    background: bar.kind === "SCHEDULED" ? "#fed7aa" : "#bfdbfe",
                                    borderLeft: bar.continuesFromPreviousDay ? "4px solid #555" : "none",
                                    borderRight: bar.continuesToNextDay ? "4px solid #555" : "none",
                                  }}
                                />

                                <div
                                  style={{
                                    position: "absolute",
                                    top: "8px",
                                    left: "12px",
                                    right: "12px",
                                    fontSize: "12px",
                                    fontWeight: 700,
                                    color: "#111827",
                                    whiteSpace: "nowrap",
                                    overflow: "hidden",
                                    textOverflow: "ellipsis",
                                  }}
                                >
                                  {bar.continuesFromPreviousDay && "← "}
                                  {bar.label}
                                  {bar.continuesToNextDay && " →"}
                                </div>
                              </div>
                            </div>
                          </button>
                        );
                      })}
                    </div>
                  ) : (
                    <div style={{ color: "#999", fontSize: "14px" }}>
                      クリックして勤怠を登録
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </section>
      </main>

      {isAttendanceModalOpen && (
        <div
          style={{
            position: "fixed",
            inset: 0,
            background: "rgba(0,0,0,0.45)",
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            zIndex: 1000,
          }}
          onClick={handleCloseModal}
        >
          <div
            style={{
              width: "min(960px, 92vw)",
              maxHeight: "90vh",
              overflowY: "auto",
              background: "#fff",
              borderRadius: "12px",
              padding: "24px",
              boxShadow: "0 10px 30px rgba(0,0,0,0.2)",
            }}
            onClick={(event) => event.stopPropagation()}
          >
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "18px" }}>
              <h2 style={{ margin: 0, fontSize: "22px" }}>
                {modalTitle}
              </h2>

              <button
                type="button"
                onClick={handleCloseModal}
                style={{ border: "none", background: "transparent", fontSize: "24px", cursor: "pointer" }}
              >
                ×
              </button>
            </div>

            {isApproved && (
              <div style={{ background: "#fef3c7", border: "1px solid #f59e0b", color: "#92400e", padding: "10px 12px", borderRadius: "8px", marginBottom: "16px", fontSize: "14px" }}>
                承認済の勤怠は編集・削除できません。
              </div>
            )}

            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px" }}>
              <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "14px", fontWeight: 700 }}>
                対象日
                <input
                  type="date"
                  value={form.workDate}
                  disabled={isApproved}
                  onChange={(event) => handleChangeForm("workDate", event.target.value)}
                  style={{ padding: "10px", border: "1px solid #ddd", borderRadius: "8px" }}
                />
              </label>

              <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "14px", fontWeight: 700 }}>
                勤務先
                <input
                  type="text"
                  value={form.workSiteName}
                  disabled={isApproved}
                  onChange={(event) => handleChangeForm("workSiteName", event.target.value)}
                  style={{ padding: "10px", border: "1px solid #ddd", borderRadius: "8px" }}
                />
              </label>

              <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "14px", fontWeight: 700 }}>
                勤怠区分
                <select
                  value={form.attendanceType}
                  disabled={isApproved}
                  onChange={(event) => handleChangeAttendanceType(event.target.value as AttendanceType)}
                  style={{ padding: "10px", border: "1px solid #ddd", borderRadius: "8px" }}
                >
                  {Object.entries(ATTENDANCE_TYPE_LABEL).map(([value, label]) => (
                    <option key={value} value={value}>
                      {label}
                    </option>
                  ))}
                </select>
              </label>

              <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "14px", fontWeight: 700 }}>
                シフトパターン
                <select
                  value={form.shiftPatternName}
                  disabled={isApproved}
                  onChange={(event) => handleChangeShiftPreset(event.target.value)}
                  style={{ padding: "10px", border: "1px solid #ddd", borderRadius: "8px" }}
                >
                  {SHIFT_PRESETS.map((shift) => (
                    <option key={shift.name} value={shift.name}>
                      {shift.name}
                    </option>
                  ))}
                </select>
              </label>
            </div>

            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px", marginTop: "20px" }}>
              <section style={{ border: "1px solid #fed7aa", background: "#fff7ed", borderRadius: "10px", padding: "16px" }}>
                <h3 style={{ margin: "0 0 12px", fontSize: "16px", color: "#9a3412" }}>
                  予定
                </h3>

                <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
                  <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "14px", fontWeight: 700 }}>
                    予定開始日時
                    <input
                      type="datetime-local"
                      value={form.scheduledStartAt}
                      disabled={isApproved}
                      onChange={(event) => handleChangeForm("scheduledStartAt", event.target.value)}
                      style={{ padding: "10px", border: "1px solid #ddd", borderRadius: "8px" }}
                    />
                  </label>

                  <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "14px", fontWeight: 700 }}>
                    予定終了日時
                    <input
                      type="datetime-local"
                      value={form.scheduledEndAt}
                      disabled={isApproved}
                      onChange={(event) => handleChangeForm("scheduledEndAt", event.target.value)}
                      style={{ padding: "10px", border: "1px solid #ddd", borderRadius: "8px" }}
                    />
                  </label>
                </div>
              </section>

              <section style={{ border: "1px solid #bfdbfe", background: "#eff6ff", borderRadius: "10px", padding: "16px" }}>
                <h3 style={{ margin: "0 0 12px", fontSize: "16px", color: "#1d4ed8" }}>
                  実績
                </h3>

                {actualFieldsDisabled && (
                  <div style={{ background: "#fff", border: "1px solid #ddd", padding: "10px", borderRadius: "8px", color: "#666", fontSize: "13px", marginBottom: "12px" }}>
                    有給・欠勤・病欠は実働ではないため、実績開始/終了は保存しません。画面表示では予定時間を使います。
                  </div>
                )}

                <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
                  <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "14px", fontWeight: 700 }}>
                    実績開始日時
                    <input
                      type="datetime-local"
                      value={form.actualStartAt}
                      disabled={isApproved || actualFieldsDisabled}
                      onChange={(event) => handleChangeForm("actualStartAt", event.target.value)}
                      style={{ padding: "10px", border: "1px solid #ddd", borderRadius: "8px", background: actualFieldsDisabled ? "#eee" : "#fff" }}
                    />
                  </label>

                  <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "14px", fontWeight: 700 }}>
                    実績終了日時
                    <input
                      type="datetime-local"
                      value={form.actualEndAt}
                      disabled={isApproved || actualFieldsDisabled}
                      onChange={(event) => handleChangeForm("actualEndAt", event.target.value)}
                      style={{ padding: "10px", border: "1px solid #ddd", borderRadius: "8px", background: actualFieldsDisabled ? "#eee" : "#fff" }}
                    />
                  </label>
                </div>
              </section>
            </div>

            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px", marginTop: "20px" }}>
              {renderBreakInputs(scheduledBreaks, "SCHEDULED")}
              {renderBreakInputs(actualBreaks, "ACTUAL")}
            </div>

            <section style={{ marginTop: "20px" }}>
              <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "14px", fontWeight: 700 }}>
                メモ
                <textarea
                  value={form.memo}
                  disabled={isApproved}
                  onChange={(event) => handleChangeForm("memo", event.target.value)}
                  rows={4}
                  style={{ padding: "10px", border: "1px solid #ddd", borderRadius: "8px", resize: "vertical" }}
                />
              </label>
            </section>

            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: "24px" }}>
              <div>
                {form.id && (
                  <button
                    type="button"
                    disabled={isApproved}
                    onClick={() => setIsDeleteConfirmOpen(true)}
                    style={{ border: "none", background: "#dc2626", color: "#fff", padding: "10px 18px", borderRadius: "8px", cursor: isApproved ? "not-allowed" : "pointer", fontWeight: 700 }}
                  >
                    削除
                  </button>
                )}
              </div>

              <div style={{ display: "flex", gap: "10px" }}>
                <button
                  type="button"
                  onClick={handleCloseModal}
                  style={{ border: "1px solid #ddd", background: "#fff", padding: "10px 18px", borderRadius: "8px", cursor: "pointer", fontWeight: 700 }}
                >
                  キャンセル
                </button>

                <button
                  type="button"
                  disabled={isApproved}
                  onClick={handleSave}
                  style={{ border: "none", background: "#f97316", color: "#fff", padding: "10px 18px", borderRadius: "8px", cursor: isApproved ? "not-allowed" : "pointer", fontWeight: 700 }}
                >
                  保存
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {isDeleteConfirmOpen && (
        <ConfirmModal
          title="勤怠を削除しますか？"
          message="この勤怠データを削除します。よろしいですか？"
          confirmText="削除する"
          cancelText="キャンセル"
          onCancel={() => setIsDeleteConfirmOpen(false)}
          onConfirm={handleDelete}
        />
      )}
    </div>
  );
}