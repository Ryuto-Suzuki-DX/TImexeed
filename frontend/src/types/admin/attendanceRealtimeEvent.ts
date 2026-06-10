/*
 * 管理者 勤怠リアルタイムイベント Type
 *
 * 管理者側で、従業員がmypageで押した
 * ・出勤
 * ・退勤
 * ・その他
 * の時刻を確認するために使用する。
 *
 * 注意：
 * ・管理者はイベントを作成しない
 * ・検索と一覧表示のみ
 * ・月次勤怠には反映しない
 */

/*
 * 勤怠リアルタイムイベント種別
 */
export type AttendanceRealtimeEventType = "CLOCK_IN" | "CLOCK_OUT" | "OTHER";

/*
 * =========================================================
 * Request
 * =========================================================
 */

/*
 * 勤怠リアルタイムイベント検索リクエスト
 *
 * targetDate:
 * ・YYYY-MM-DD
 * ・空の場合、バックエンド側でJSTの本日扱い
 */
export type SearchAttendanceRealtimeEventsRequest = {
  targetDate: string;
  keyword: string;
  eventTypes: AttendanceRealtimeEventType[];
  limit: number;
  offset: number;
};

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
  userId: number;
  userName: string;
  userEmail: string;
  eventDate: string;
  eventType: AttendanceRealtimeEventType;
  eventAt: string;
  note: string | null;
  clientIp: string | null;
  userAgent: string | null;
  createdAt: string;
};

/*
 * 勤怠リアルタイムイベント検索レスポンス
 */
export type SearchAttendanceRealtimeEventsResponse = {
  events: AttendanceRealtimeEventResponse[];
  total: number;
  offset: number;
  limit: number;
  hasMore: boolean;
};
