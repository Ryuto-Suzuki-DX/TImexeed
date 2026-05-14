package types

/*
 * 〇 管理者 勤務区分マスタ検索リクエスト
 *
 * 管理者側では対象ユーザーの勤怠入力・編集画面のプルダウン用に、
 * 有効な勤務区分を全件取得する。
 *
 * 現時点では検索条件なし。
 *
 * 注意：
 * ・勤務区分マスタは全ユーザー共通
 * ・targetUserId は不要
 * ・管理者用APIとして /admin/attendance-types/search から取得する
 */
type SearchAttendanceTypesRequest struct {
}

/*
 * 〇 管理者 勤務区分マスタレスポンス
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
 * 〇 管理者 勤務区分マスタ検索結果
 */
type SearchAttendanceTypesResponse struct {
	AttendanceTypes []AttendanceTypeResponse `json:"attendanceTypes"`
}
