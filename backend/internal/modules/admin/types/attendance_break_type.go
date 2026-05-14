package types

import "time"

/*
 * 〇 管理者 休憩 Type
 *
 * 管理者が対象ユーザーの指定日の休憩データを扱う型。
 *
 * 重要：
 * ・管理者APIでは対象ユーザーIDを targetUserId としてRequestで受け取る
 * ・attendanceDayId はRequestで受け取らない
 * ・Serviceで targetUserId + workDate から勤怠日を特定する
 * ・attendanceBreakId がその勤怠日に紐づくか確認する
 * ・管理者側では月次申請状態による編集ロックを行わない
 *
 * user側との差分：
 * ・user側はJWTからログイン中ユーザーIDを取得する
 * ・admin側はrequest bodyのtargetUserIdで対象ユーザーを指定する
 * ・user側は月次申請状態により編集不可になる場合がある
 * ・admin側は月次申請状態に関係なく編集できる
 */

/*
 * 〇 休憩検索リクエスト
 *
 * 管理者が対象ユーザーの指定日の休憩一覧を取得する。
 */
type SearchAttendanceBreaksRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 対象日
	// 例：2026-05-05
	WorkDate string `json:"workDate" binding:"required"`
}

/*
 * 〇 休憩作成リクエスト
 *
 * 管理者が対象ユーザーの指定日の休憩を作成する。
 *
 * 注意：
 * ・attendanceDayId は受け取らない
 * ・Serviceで targetUserId + workDate から勤怠日を特定する
 */
type CreateAttendanceBreakRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 対象日
	// 例：2026-05-05
	WorkDate string `json:"workDate" binding:"required"`

	// 休憩開始日時
	// 例：2026-05-05T12:00:00+09:00
	BreakStartAt string `json:"breakStartAt" binding:"required"`

	// 休憩終了日時
	// 例：2026-05-05T13:00:00+09:00
	BreakEndAt string `json:"breakEndAt" binding:"required"`

	// 休憩メモ
	BreakMemo *string `json:"breakMemo"`
}

/*
 * 〇 休憩更新リクエスト
 *
 * 管理者が対象ユーザーの指定日の休憩を更新する。
 *
 * 注意：
 * ・attendanceBreakId は受け取る
 * ・Serviceで targetUserId + workDate から勤怠日を特定する
 * ・attendanceBreakId がその勤怠日に紐づくか確認する
 */
type UpdateAttendanceBreakRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 対象日
	// 例：2026-05-05
	WorkDate string `json:"workDate" binding:"required"`

	// 休憩ID
	AttendanceBreakID uint `json:"attendanceBreakId" binding:"required"`

	// 休憩開始日時
	// 例：2026-05-05T12:00:00+09:00
	BreakStartAt string `json:"breakStartAt" binding:"required"`

	// 休憩終了日時
	// 例：2026-05-05T13:00:00+09:00
	BreakEndAt string `json:"breakEndAt" binding:"required"`

	// 休憩メモ
	BreakMemo *string `json:"breakMemo"`
}

/*
 * 〇 休憩削除リクエスト
 *
 * 管理者が対象ユーザーの指定日の休憩を論理削除する。
 *
 * 注意：
 * ・Serviceで targetUserId + workDate から勤怠日を特定する
 * ・attendanceBreakId がその勤怠日に紐づくか確認する
 */
type DeleteAttendanceBreakRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 対象日
	// 例：2026-05-05
	WorkDate string `json:"workDate" binding:"required"`

	// 休憩ID
	AttendanceBreakID uint `json:"attendanceBreakId" binding:"required"`
}

/*
 * 〇 対象日の休憩差分保存リクエスト
 *
 * monthly_attendances/update の月次全体保存から内部的に使う。
 *
 * 保存方針：
 * ・attendanceBreakId がある休憩は更新する
 * ・attendanceBreakId がない休憩は新規作成する
 * ・DBに存在するがリクエストから消えた休憩は論理削除する
 *
 * 注意：
 * ・管理者側では targetUserId を受け取る
 * ・attendanceDayId は受け取らない
 * ・Serviceで targetUserId + workDate から勤怠日を特定する
 * ・月次申請状態による編集ロックは行わない
 */
type UpdateAttendanceBreaksByWorkDateRequest struct {
	// 対象ユーザーID
	TargetUserID uint `json:"targetUserId" binding:"required"`

	// 対象日
	// 例：2026-05-05
	WorkDate string `json:"workDate" binding:"required"`

	// 対象日の休憩一覧
	Breaks []UpdateAttendanceBreaksByWorkDateBreakRequest `json:"breaks"`
}

/*
 * 〇 対象日の休憩差分保存リクエスト 1件分
 *
 * attendanceBreakId:
 * ・nil または 0 の場合は新規作成
 * ・1以上の場合は既存休憩更新
 */
type UpdateAttendanceBreaksByWorkDateBreakRequest struct {
	// 休憩ID
	// 新規作成の場合は nil
	AttendanceBreakID *uint `json:"attendanceBreakId"`

	// 休憩開始日時
	// 例：2026-05-05T12:00:00+09:00
	BreakStartAt string `json:"breakStartAt" binding:"required"`

	// 休憩終了日時
	// 例：2026-05-05T13:00:00+09:00
	BreakEndAt string `json:"breakEndAt" binding:"required"`

	// 休憩メモ
	BreakMemo *string `json:"breakMemo"`
}

/*
 * 〇 休憩レスポンス
 *
 * フロントへ返す1件分の休憩データ。
 */
type AttendanceBreakResponse struct {
	ID uint `json:"id"`

	// 紐づく勤怠日ID
	AttendanceDayID uint `json:"attendanceDayId"`

	// 休憩開始日時
	BreakStartAt time.Time `json:"breakStartAt"`

	// 休憩終了日時
	BreakEndAt time.Time `json:"breakEndAt"`

	// 休憩メモ
	BreakMemo *string `json:"breakMemo"`

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
 * 〇 休憩検索レスポンス
 */
type SearchAttendanceBreaksResponse struct {
	TargetUserID uint `json:"targetUserId"`

	WorkDate string `json:"workDate"`

	AttendanceBreaks []AttendanceBreakResponse `json:"attendanceBreaks"`
}

/*
 * 〇 休憩作成レスポンス
 */
type CreateAttendanceBreakResponse struct {
	AttendanceBreak AttendanceBreakResponse `json:"attendanceBreak"`
}

/*
 * 〇 休憩更新レスポンス
 */
type UpdateAttendanceBreakResponse struct {
	AttendanceBreak AttendanceBreakResponse `json:"attendanceBreak"`
}

/*
 * 〇 休憩削除レスポンス
 */
type DeleteAttendanceBreakResponse struct {
	TargetUserID uint `json:"targetUserId"`

	WorkDate string `json:"workDate"`

	AttendanceBreakID uint `json:"attendanceBreakId"`
}

/*
 * 〇 対象日の休憩差分保存レスポンス
 */
type UpdateAttendanceBreaksByWorkDateResponse struct {
	TargetUserID uint `json:"targetUserId"`

	WorkDate string `json:"workDate"`

	// 作成・更新された休憩数
	// 論理削除は保存件数には含めない
	SavedAttendanceBreakCount int `json:"savedAttendanceBreakCount"`

	// 新規作成した休憩数
	CreatedAttendanceBreakCount int `json:"createdAttendanceBreakCount"`

	// 更新した休憩数
	UpdatedAttendanceBreakCount int `json:"updatedAttendanceBreakCount"`

	// 論理削除した休憩数
	DeletedAttendanceBreakCount int `json:"deletedAttendanceBreakCount"`
}
