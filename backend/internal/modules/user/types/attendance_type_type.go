package types

/*
 * 〇 勤務区分マスタ検索リクエスト
 *
 * ユーザー側では勤怠入力画面のプルダウン用に、
 * 有効な勤務区分を全件取得する。
 *
 * 現時点では検索条件なし。
 */
type SearchAttendanceTypesRequest struct {
}

/*
 * 〇 勤務区分マスタレスポンス
 *
 * フロントはこの情報を見て入力欄を切り替える。
 */
type AttendanceTypeResponse struct {
	ID       uint   `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Category string `json:"category"`

	SyncPlanActual bool `json:"syncPlanActual"`

	AllowActualTimeInput bool `json:"allowActualTimeInput"`
	AllowBreakInput      bool `json:"allowBreakInput"`
	AllowTransportInput  bool `json:"allowTransportInput"`

	AllowLateFlag       bool `json:"allowLateFlag"`
	AllowEarlyLeaveFlag bool `json:"allowEarlyLeaveFlag"`
	AllowAbsenceFlag    bool `json:"allowAbsenceFlag"`
	AllowSickLeaveFlag  bool `json:"allowSickLeaveFlag"`

	RequiresRequest bool `json:"requiresRequest"`

	DisplayOrder int `json:"displayOrder"`
}

/*
 * 〇 勤務区分マスタ検索結果
 */
type SearchAttendanceTypesResponse struct {
	AttendanceTypes []AttendanceTypeResponse `json:"attendanceTypes"`
}
