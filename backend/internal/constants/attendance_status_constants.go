package constants

/*
 * 〇 実績状態 固定値
 *
 * attendance_types は予定・勤務区分マスタとして使う。
 * 実績状態はDBマスタにせず、コード内の固定値として扱う。
 *
 * 対象：
 * 	・通常
 * 	・欠勤
 * 	・病欠
 * 	・遅刻
 * 	・早退
 *
 * 注意：
 * 	夜勤は実績状態ではない。
 * 	夜勤・深夜時間は actual_start_at / actual_end_at から集計時に計算する。
 */
const (
	ActualWorkStatusNormal     = "NORMAL"
	ActualWorkStatusAbsence    = "ABSENCE"
	ActualWorkStatusSickLeave  = "SICK_LEAVE"
	ActualWorkStatusLate       = "LATE"
	ActualWorkStatusEarlyLeave = "EARLY_LEAVE"
)

/*
 * 〇 実績状態 表示名
 *
 * フロント表示やレスポンス生成時に使う。
 */
var ActualWorkStatusLabels = map[string]string{
	ActualWorkStatusNormal:     "通常",
	ActualWorkStatusAbsence:    "欠勤",
	ActualWorkStatusSickLeave:  "病欠",
	ActualWorkStatusLate:       "遅刻",
	ActualWorkStatusEarlyLeave: "早退",
}
