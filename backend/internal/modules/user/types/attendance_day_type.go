package types

import "time"

/*
 * 〇 従業員 勤怠日別 Type
 *
 * 従業員本人の勤怠日別データを扱う型。
 *
 * 重要：
 * ・AttendanceDay は日別勤怠データだけを持つ
 * ・申請状態、承認状態は AttendanceDay では持たない
 * ・月次申請状態は MonthlyAttendanceRequestResponse として返す
 * ・従業員APIでは userId / targetUserId をRequestで受け取らない
 *
 * 勤務区分と実績状態の整理：
 * ・PlanAttendanceTypeID は attendance_types のIDを使う
 * ・ActualAttendanceTypeID も attendance_types のIDを保存する
 * ・ただし、通常勤務ではバックエンド側で PlanAttendanceTypeID と同じ値を保存する
 * ・欠勤、病欠、遅刻、早退は attendance_types ではなく各Flagで表現する
 * ・夜勤は勤務区分ではなく、実績時間から集計時に深夜時間として計算する
 */

/*
 * 勤怠検索 Request
 *
 * POST /user/attendance-days/search
 */
type SearchAttendanceDaysRequest struct {
	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`
}

/*
 * 勤怠更新 Request
 *
 * APIとして直接公開しない。
 * monthly_attendances/update の月次全体保存から内部的に使う。
 */
type UpdateAttendanceDayRequest struct {
	// 対象日
	WorkDate string `json:"workDate" binding:"required"`

	// 予定区分ID
	// attendance_types のIDを指定する。
	PlanAttendanceTypeID uint `json:"planAttendanceTypeId" binding:"required"`

	// 実績区分ID
	// 現在はフロントからの送信必須ではない。
	// 通常勤務、休日、予定・実績同期区分では、Service側で PlanAttendanceTypeID と同じ値を保存する。
	//
	// 注意：
	// 欠勤、病欠、遅刻、早退をここに入れない。
	// それらは AbsenceFlag / SickLeaveFlag / LateFlag / EarlyLeaveFlag で表現する。
	ActualAttendanceTypeID *uint `json:"actualAttendanceTypeId"`

	// 予定開始日時
	PlanStartAt *string `json:"planStartAt"`

	// 予定終了日時
	PlanEndAt *string `json:"planEndAt"`

	// 実績開始日時
	ActualStartAt *string `json:"actualStartAt"`

	// 実績終了日時
	ActualEndAt *string `json:"actualEndAt"`

	// 共通開始日時
	// 有給、特別休暇、休職、介護休業、育児休業など、予定・実績を同期する区分で使う。
	CommonStartAt *string `json:"commonStartAt"`

	// 共通終了日時
	// 有給、特別休暇、休職、介護休業、育児休業など、予定・実績を同期する区分で使う。
	CommonEndAt *string `json:"commonEndAt"`

	// 遅刻フラグ
	// 出勤日にだけ使用する。
	LateFlag bool `json:"lateFlag"`

	// 早退フラグ
	// 出勤日にだけ使用する。
	EarlyLeaveFlag bool `json:"earlyLeaveFlag"`

	// 欠勤フラグ
	// 出勤日にだけ使用する。
	AbsenceFlag bool `json:"absenceFlag"`

	// 病欠フラグ
	// 出勤日にだけ使用する。
	SickLeaveFlag bool `json:"sickLeaveFlag"`

	// 在宅勤務補助対象フラグ
	RemoteWorkAllowanceFlag bool `json:"remoteWorkAllowanceFlag"`

	// 日別交通費：出発地
	TransportFrom *string `json:"transportFrom"`

	// 日別交通費：目的地
	TransportTo *string `json:"transportTo"`

	// 日別交通費：手段
	TransportMethod *string `json:"transportMethod"`

	// 日別交通費：金額
	TransportAmount *int `json:"transportAmount"`
}

/*
 * 勤怠削除 Request
 *
 * 現時点ではAPIとして直接公開しない。
 * 必要になった場合の内部用として残す。
 */
type DeleteAttendanceDayRequest struct {
	// 対象日
	WorkDate string `json:"workDate" binding:"required"`
}

/*
 * 勤怠日別 Response
 *
 * AttendanceDay 自体のデータだけを返す。
 * 月次申請状態はここには入れない。
 */
type AttendanceDayResponse struct {
	// 勤怠ID
	ID uint `json:"id"`

	// 対象日
	WorkDate time.Time `json:"workDate"`

	// 予定区分ID
	PlanAttendanceTypeID uint `json:"planAttendanceTypeId"`

	// 実績区分ID
	// 基本的に予定区分IDと同じ勤務区分IDが入る。
	// 欠勤、病欠、遅刻、早退は各Flagで表現する。
	ActualAttendanceTypeID uint `json:"actualAttendanceTypeId"`

	// 予定開始日時
	PlanStartAt *time.Time `json:"planStartAt"`

	// 予定終了日時
	PlanEndAt *time.Time `json:"planEndAt"`

	// 実績開始日時
	ActualStartAt *time.Time `json:"actualStartAt"`

	// 実績終了日時
	ActualEndAt *time.Time `json:"actualEndAt"`

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

	// 日別交通費：出発地
	TransportFrom *string `json:"transportFrom"`

	// 日別交通費：目的地
	TransportTo *string `json:"transportTo"`

	// 日別交通費：手段
	TransportMethod *string `json:"transportMethod"`

	// 日別交通費：金額
	TransportAmount *int `json:"transportAmount"`

	// 論理削除フラグ
	IsDeleted bool `json:"isDeleted"`

	// 作成日時
	CreatedAt time.Time `json:"createdAt"`

	// 更新日時
	UpdatedAt time.Time `json:"updatedAt"`

	// 論理削除日時
	DeletedAt *time.Time `json:"deletedAt"`
}

/*
 * 勤怠検索 Response
 *
 * 勤怠日別データと、対象月の月次申請状態を一緒に返す。
 */
type SearchAttendanceDaysResponse struct {
	// 対象年
	TargetYear int `json:"targetYear"`

	// 対象月
	TargetMonth int `json:"targetMonth"`

	// 対象月の月次申請状態
	MonthlyAttendanceRequest MonthlyAttendanceRequestResponse `json:"monthlyAttendanceRequest"`

	// 勤怠日別一覧
	AttendanceDays []AttendanceDayResponse `json:"attendanceDays"`
}

/*
 * 勤怠更新 Response
 *
 * monthly_attendances/update の内部処理で使う。
 */
type UpdateAttendanceDayResponse struct {
	AttendanceDay AttendanceDayResponse `json:"attendanceDay"`
}

/*
 * 勤怠削除 Response
 *
 * 必要になった場合の内部用。
 */
type DeleteAttendanceDayResponse struct {
	WorkDate string `json:"workDate"`
}
