/*
 * 従業員 勤怠リアルタイムイベント Type
 *
 * ユーザー側mypageの
 * ・出勤
 * ・退勤
 * ボタンで使用する。
 *
 * 注意：
 * ・ユーザーIDはリクエストで送らない
 * ・JWTからバックエンド側で取得する
 * ・月次勤怠には反映しない
 * ・同じユーザーが同じ日に同じイベント種別を登録できるのは1回だけ
 * ・登録後の取消・編集はしない
 */

/*
 * 勤怠リアルタイムイベント種別
 */
export type AttendanceRealtimeEventType = "CLOCK_IN" | "CLOCK_OUT";

/*
 * =========================================================
 * Request
 * =========================================================
 */

/*
 * 勤怠リアルタイムイベント作成リクエスト
 */
export type CreateAttendanceRealtimeEventRequest = {
  eventType: AttendanceRealtimeEventType;
  note: string;
};

/*
 * 本日の勤怠リアルタイムイベント状態取得リクエスト
 *
 * ユーザーIDは送らない。
 */
export type GetTodayAttendanceRealtimeEventsRequest = Record<string, never>;

/*
 * =========================================================
 * Response
 * =========================================================
 */

/*
 * 勤怠リアルタイムイベントレスポンス
 */
export type AttendanceRealtimeEventResponse = {
  id: number;
  eventDate: string;
  eventType: AttendanceRealtimeEventType;
  eventAt: string;
  note: string | null;
  createdAt: string;
};

/*
 * 勤怠リアルタイムイベント作成レスポンス
 */
export type CreateAttendanceRealtimeEventResponse = {
  event: AttendanceRealtimeEventResponse;
};

/*
 * 本日の勤怠リアルタイムイベント状態取得レスポンス
 *
 * mypage側では、
 * clockInRecorded / clockOutRecorded を見て
 * 出勤・退勤ボタンをdisabledにする。
 *
 * 登録済みの場合は、
 * 押下時刻とコメントを表示する。
 */
export type GetTodayAttendanceRealtimeEventsResponse = {
  clockInRecorded: boolean;
  clockOutRecorded: boolean;
  clockInAt: string | null;
  clockOutAt: string | null;
  clockInNote: string | null;
  clockOutNote: string | null;
  events: AttendanceRealtimeEventResponse[];
};
