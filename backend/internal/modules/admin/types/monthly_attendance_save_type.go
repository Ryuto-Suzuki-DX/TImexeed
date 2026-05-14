package types

/*
 * 〇 管理者 月次勤怠全体保存リクエスト
 *
 * 管理者用月次勤怠画面の「全体保存」用。
 *
 * 保存対象：
 * ・月次通勤定期
 * ・日別勤怠
 * ・日別休憩
 *
 * 重要：
 * ・管理者APIでは対象ユーザーIDを targetUserId としてRequestで受け取る
 * ・ControllerではJWTのuserIdを対象ユーザーIDとして使わない
 * ・管理者側では月次申請状態による編集ロックを行わない
 *
 * 保存方針：
 * ・このRequestをControllerでbindする
 * ・Service側で既存のadmin用Serviceへ処理を振り分ける
 * ・日別勤怠は AttendanceDayService.UpdateAttendanceDay を使う
 * ・休憩は AttendanceBreakService.UpdateAttendanceBreaksByWorkDate を使う
 * ・月次通勤定期は MonthlyCommuterPassService.UpdateMonthlyCommuterPass を使う
 *
 * 注意：
 * ・このtypeは画面から一括保存されるデータ構造
 * ・DB保存用Modelではない
 * ・Repository / Builder は基本的に作らない
 */
type UpdateMonthlyAttendanceRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`

	// 月次通勤定期
	CommuterPass *UpdateMonthlyAttendanceCommuterPassRequest `json:"commuterPass"`

	// 日別勤怠一覧
	AttendanceDays []UpdateMonthlyAttendanceDayRequest `json:"attendanceDays"`
}

/*
 * 〇 管理者 月次勤怠全体保存：月次通勤定期
 *
 * 注意：
 * ・targetUserId / targetYear / targetMonth は親Requestから引き継ぐ
 * ・Service側で UpdateMonthlyCommuterPassRequest へ詰め替える
 */
type UpdateMonthlyAttendanceCommuterPassRequest struct {
	// 定期：出発地
	CommuterFrom *string `json:"commuterFrom"`

	// 定期：目的地
	CommuterTo *string `json:"commuterTo"`

	// 定期：手段
	CommuterMethod *string `json:"commuterMethod"`

	// 定期：金額
	CommuterAmount *int `json:"commuterAmount"`
}

/*
 * 〇 管理者 月次勤怠全体保存：日別勤怠
 *
 * 注意：
 * ・targetUserId は親Requestから引き継ぐ
 * ・Service側で UpdateAttendanceDayRequest へ詰め替える
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
type UpdateMonthlyAttendanceDayRequest struct {
	// 対象日
	// 例：2026-05-05
	WorkDate string `json:"workDate" binding:"required"`

	// 予定区分ID
	PlanAttendanceTypeID uint `json:"planAttendanceTypeId" binding:"required"`

	// 実績区分ID
	ActualAttendanceTypeID *uint `json:"actualAttendanceTypeId"`

	// 共通開始日時
	CommonStartAt *string `json:"commonStartAt"`

	// 共通終了日時
	CommonEndAt *string `json:"commonEndAt"`

	// 予定開始日時
	PlanStartAt *string `json:"planStartAt"`

	// 予定終了日時
	PlanEndAt *string `json:"planEndAt"`

	// 実績開始日時
	ActualStartAt *string `json:"actualStartAt"`

	// 実績終了日時
	ActualEndAt *string `json:"actualEndAt"`

	// 遅刻フラグ
	LateFlag bool `json:"lateFlag"`

	// 早退フラグ
	EarlyLeaveFlag bool `json:"earlyLeaveFlag"`

	// 欠勤フラグ
	AbsenceFlag bool `json:"absenceFlag"`

	// 病欠フラグ
	SickLeaveFlag bool `json:"sickLeaveFlag"`

	// 在宅勤務補助対象フラグ
	RemoteWorkAllowanceFlag bool `json:"remoteWorkAllowanceFlag"`

	// 申請メモ
	//
	// 注意：
	// ・現時点では AttendanceDay には申請メモを保存しない
	// ・既存user側typeに合わせて残している
	// ・Service側で使わない場合は無視する
	RequestMemo *string `json:"requestMemo"`

	// 日別交通費：出発地
	TransportFrom *string `json:"transportFrom"`

	// 日別交通費：目的地
	TransportTo *string `json:"transportTo"`

	// 日別交通費：手段
	TransportMethod *string `json:"transportMethod"`

	// 日別交通費：金額
	TransportAmount *int `json:"transportAmount"`

	// 休憩一覧
	Breaks []UpdateMonthlyAttendanceBreakRequest `json:"breaks"`
}

/*
 * 〇 管理者 月次勤怠全体保存：休憩
 *
 * 方針：
 * ・画面に残っている休憩だけ送る
 * ・保存時は差分保存する
 * ・attendanceBreakId がある休憩は更新する
 * ・attendanceBreakId がない休憩は新規作成する
 * ・DBに存在するがリクエストから消えた休憩は論理削除する
 *
 * 注意：
 * ・targetUserId / workDate は親の日別勤怠Requestから引き継ぐ
 * ・Service側で UpdateAttendanceBreaksByWorkDateRequest へ詰め替える
 */
type UpdateMonthlyAttendanceBreakRequest struct {
	// 休憩ID
	// 新規作成の場合は nil
	AttendanceBreakID *uint `json:"attendanceBreakId"`

	// 休憩開始日時
	BreakStartAt string `json:"breakStartAt" binding:"required"`

	// 休憩終了日時
	BreakEndAt string `json:"breakEndAt" binding:"required"`

	// 休憩メモ
	BreakMemo *string `json:"breakMemo"`
}

/*
 * 〇 管理者 月次勤怠全体保存レスポンス
 */
type UpdateMonthlyAttendanceResponse struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId"`

	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	SavedMonthlyCommuterPass  bool `json:"savedMonthlyCommuterPass"`
	SavedAttendanceDayCount   int  `json:"savedAttendanceDayCount"`
	SavedAttendanceBreakCount int  `json:"savedAttendanceBreakCount"`
}
