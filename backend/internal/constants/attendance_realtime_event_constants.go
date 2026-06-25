package constants

/*
 * 勤怠リアルタイムイベント種別
 */
const (
	/*
	 * 現在ユーザーが新規登録できるイベント種別。
	 */
	AttendanceRealtimeEventTypeClockIn  = "CLOCK_IN"
	AttendanceRealtimeEventTypeClockOut = "CLOCK_OUT"

	/*
	 * 旧仕様との互換性用。
	 *
	 * ユーザーAPIからは新規登録できない。
	 * 既存のOTHERデータや、管理者側・集計処理などから
	 * 参照されている可能性があるため定数のみ残す。
	 */
	AttendanceRealtimeEventTypeOther = "OTHER"
)
