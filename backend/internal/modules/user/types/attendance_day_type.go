package types

import "time"

/*
 * 〇 勤怠検索リクエスト
 *
 * 従業員本人の対象年月の勤怠一覧を取得する。
 *
 * 注意：
 * ・userId / targetUserId は受け取らない
 * ・ログイン中ユーザーIDはControllerでJWTから取得してServiceへ渡す
 */
type SearchAttendanceDaysRequest struct {
	// 対象年
	TargetYear int `json:"targetYear" binding:"required"`

	// 対象月
	TargetMonth int `json:"targetMonth" binding:"required"`
}

/*
 * 〇 勤怠更新リクエスト
 *
 * 月次一覧画面の1行を直接更新するためのリクエスト。
 *
 * 仕様：
 * ・workDateで対象日を指定する
 * ・未登録の日付なら新規作成する
 * ・登録済みの日付なら更新する
 *
 * commonStartAt / commonEndAt：
 * ・DBには保存しない
 * ・syncPlanActual = true の勤務区分で使う
 * ・Serviceで plan / actual の両方へ反映する
 *
 * 注意：
 * ・userId / targetUserId は受け取らない
 * ・attendanceDayId も基本使わない
 * ・userID + workDate で対象勤怠を特定する
 */
type UpdateAttendanceDayRequest struct {
	// 対象日
	// 例：2026-05-05
	WorkDate string `json:"workDate" binding:"required"`

	// 予定区分ID
	PlanAttendanceTypeID uint `json:"planAttendanceTypeId" binding:"required"`

	// 実績区分ID
	// 通常勤務など、予定と実績を分ける区分で使う
	// syncPlanActual = true の場合はService側で予定区分IDを実績にも反映する
	ActualAttendanceTypeID *uint `json:"actualAttendanceTypeId"`

	// 共通開始日時
	// 有給・欠勤・病欠・休職・介護休業など、
	// syncPlanActual = true の区分で使う
	CommonStartAt *string `json:"commonStartAt"`

	// 共通終了日時
	// 有給・欠勤・病欠・休職・介護休業など、
	// syncPlanActual = true の区分で使う
	CommonEndAt *string `json:"commonEndAt"`

	// 予定開始日時
	// 通常勤務など、予定と実績を分ける区分で使う
	PlanStartAt *string `json:"planStartAt"`

	// 予定終了日時
	// 通常勤務など、予定と実績を分ける区分で使う
	PlanEndAt *string `json:"planEndAt"`

	// 実績開始日時
	// 通常勤務など、予定と実績を分ける区分で使う
	ActualStartAt *string `json:"actualStartAt"`

	// 実績終了日時
	// 通常勤務など、予定と実績を分ける区分で使う
	ActualEndAt *string `json:"actualEndAt"`

	// 遅刻フラグ
	LateFlag bool `json:"lateFlag"`

	// 早退フラグ
	EarlyLeaveFlag bool `json:"earlyLeaveFlag"`

	// 欠勤フラグ
	AbsenceFlag bool `json:"absenceFlag"`

	// 病欠フラグ
	SickLeaveFlag bool `json:"sickLeaveFlag"`

	// 申請メモ
	RequestMemo *string `json:"requestMemo"`

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
 * 〇 勤怠削除リクエスト
 *
 * 従業員本人の指定日の勤怠を論理削除する。
 *
 * 注意：
 * ・userId / targetUserId は受け取らない
 * ・attendanceDayId ではなく workDate で対象日を指定する
 * ・userID + workDate で削除対象を特定する
 */
type DeleteAttendanceDayRequest struct {
	// 対象日
	// 例：2026-05-05
	WorkDate string `json:"workDate" binding:"required"`
}

/*
 * 〇 勤怠レスポンス
 *
 * フロントの月次一覧画面に返す1日分の勤怠データ。
 *
 * 注意：
 * ・日付や日時は time.Time / *time.Time のまま返す
 * ・表示形式の整形はフロント側で行う
 * ・勤務区分名は勤務区分マスタのレスポンス側で持つ想定
 */
type AttendanceDayResponse struct {
	ID uint `json:"id"`

	// 対象日
	WorkDate time.Time `json:"workDate"`

	// 予定区分ID
	PlanAttendanceTypeID uint `json:"planAttendanceTypeId"`

	// 実績区分ID
	ActualAttendanceTypeID uint `json:"actualAttendanceTypeId"`

	// 予定開始日時
	PlanStartAt *time.Time `json:"planStartAt"`

	// 予定終了日時
	PlanEndAt *time.Time `json:"planEndAt"`

	// 実績開始日時
	ActualStartAt *time.Time `json:"actualStartAt"`

	// 実績終了日時
	ActualEndAt *time.Time `json:"actualEndAt"`

	// 申請状態
	RequestStatus string `json:"requestStatus"`

	// 申請メモ
	RequestMemo *string `json:"requestMemo"`

	// 承認者ID
	ApprovedBy *uint `json:"approvedBy"`

	// 承認日時
	ApprovedAt *time.Time `json:"approvedAt"`

	// 否認理由
	RejectedReason *string `json:"rejectedReason"`

	// 遅刻フラグ
	LateFlag bool `json:"lateFlag"`

	// 早退フラグ
	EarlyLeaveFlag bool `json:"earlyLeaveFlag"`

	// 欠勤フラグ
	AbsenceFlag bool `json:"absenceFlag"`

	// 病欠フラグ
	SickLeaveFlag bool `json:"sickLeaveFlag"`

	// 画面表示用メッセージ
	SystemMessage *string `json:"systemMessage"`

	// 日別交通費：出発地
	TransportFrom *string `json:"transportFrom"`

	// 日別交通費：目的地
	TransportTo *string `json:"transportTo"`

	// 日別交通費：手段
	TransportMethod *string `json:"transportMethod"`

	// 日別交通費：金額
	TransportAmount *int `json:"transportAmount"`

	// 月次申請状態
	MonthlyStatus string `json:"monthlyStatus"`

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
 * 〇 勤怠検索レスポンス
 *
 * 対象年月の勤怠一覧を返す。
 */
type SearchAttendanceDaysResponse struct {
	TargetYear  int `json:"targetYear"`
	TargetMonth int `json:"targetMonth"`

	AttendanceDays []AttendanceDayResponse `json:"attendanceDays"`
}

/*
 * 〇 勤怠更新レスポンス
 *
 * 更新後、または新規作成後の勤怠データを返す。
 */
type UpdateAttendanceDayResponse struct {
	AttendanceDay AttendanceDayResponse `json:"attendanceDay"`
}

/*
 * 〇 勤怠削除レスポンス
 *
 * 論理削除した対象日を返す。
 */
type DeleteAttendanceDayResponse struct {
	WorkDate string `json:"workDate"`
}
