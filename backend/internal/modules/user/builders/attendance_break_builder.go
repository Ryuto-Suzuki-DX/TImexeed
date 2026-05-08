package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type AttendanceBreakBuilder interface {
	BuildSearchAttendanceBreaksQuery(attendanceDayID uint) (*gorm.DB, results.Result)
	BuildFindAttendanceBreakByIDAndAttendanceDayIDQuery(attendanceBreakID uint, attendanceDayID uint) (*gorm.DB, results.Result)
	BuildCreateAttendanceBreakModel(
		attendanceDayID uint,
		req types.CreateAttendanceBreakRequest,
		breakStartAt time.Time,
		breakEndAt time.Time,
	) (models.AttendanceBreak, results.Result)
	BuildUpdateAttendanceBreakModel(
		currentAttendanceBreak models.AttendanceBreak,
		req types.UpdateAttendanceBreakRequest,
		breakStartAt time.Time,
		breakEndAt time.Time,
	) (models.AttendanceBreak, results.Result)
	BuildDeleteAttendanceBreakModel(currentAttendanceBreak models.AttendanceBreak) (models.AttendanceBreak, results.Result)
}

/*
 * 従業員用休憩Builder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Serviceから受け取ったRequestをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Create / Save はRepositoryに任せる
 * ・日付文字列、日時文字列の変換はServiceで行う
 * ・Builderでは変換済みの time.Time を受け取る
 */
type attendanceBreakBuilder struct {
	db *gorm.DB
}

/*
 * AttendanceBreakBuilder生成
 */
func NewAttendanceBreakBuilder(db *gorm.DB) AttendanceBreakBuilder {
	return &attendanceBreakBuilder{db: db}
}

/*
 * 休憩検索用クエリ作成
 *
 * 勤怠日IDに紐づく休憩一覧を取得する。
 */
func (builder *attendanceBreakBuilder) BuildSearchAttendanceBreaksQuery(
	attendanceDayID uint,
) (*gorm.DB, results.Result) {
	if attendanceDayID == 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_ATTENDANCE_BREAKS_QUERY_INVALID_ATTENDANCE_DAY_ID",
			"休憩検索条件の作成に失敗しました",
			map[string]any{
				"attendanceDayId": attendanceDayID,
			},
		)
	}

	query := builder.db.
		Model(&models.AttendanceBreak{}).
		Where("attendance_day_id = ?", attendanceDayID).
		Where("is_deleted = ?", false).
		Order("break_start_at ASC").
		Order("id ASC")

	return query, results.OK(
		nil,
		"BUILD_SEARCH_ATTENDANCE_BREAKS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 休憩ID + 勤怠日IDで休憩1件取得用クエリ作成
 *
 * 更新・削除時に使う。
 */
func (builder *attendanceBreakBuilder) BuildFindAttendanceBreakByIDAndAttendanceDayIDQuery(
	attendanceBreakID uint,
	attendanceDayID uint,
) (*gorm.DB, results.Result) {
	if attendanceBreakID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ATTENDANCE_BREAK_QUERY_INVALID_ATTENDANCE_BREAK_ID",
			"休憩取得条件の作成に失敗しました",
			map[string]any{
				"attendanceBreakId": attendanceBreakID,
			},
		)
	}

	if attendanceDayID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ATTENDANCE_BREAK_QUERY_INVALID_ATTENDANCE_DAY_ID",
			"休憩取得条件の作成に失敗しました",
			map[string]any{
				"attendanceDayId": attendanceDayID,
			},
		)
	}

	query := builder.db.
		Model(&models.AttendanceBreak{}).
		Where("id = ?", attendanceBreakID).
		Where("attendance_day_id = ?", attendanceDayID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_ATTENDANCE_BREAK_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 休憩作成用Model作成
 */
func (builder *attendanceBreakBuilder) BuildCreateAttendanceBreakModel(
	attendanceDayID uint,
	req types.CreateAttendanceBreakRequest,
	breakStartAt time.Time,
	breakEndAt time.Time,
) (models.AttendanceBreak, results.Result) {
	if attendanceDayID == 0 {
		return models.AttendanceBreak{}, results.BadRequest(
			"BUILD_CREATE_ATTENDANCE_BREAK_MODEL_INVALID_ATTENDANCE_DAY_ID",
			"休憩作成データの作成に失敗しました",
			map[string]any{
				"attendanceDayId": attendanceDayID,
			},
		)
	}

	if breakStartAt.IsZero() {
		return models.AttendanceBreak{}, results.BadRequest(
			"BUILD_CREATE_ATTENDANCE_BREAK_MODEL_EMPTY_BREAK_START_AT",
			"休憩作成データの作成に失敗しました",
			nil,
		)
	}

	if breakEndAt.IsZero() {
		return models.AttendanceBreak{}, results.BadRequest(
			"BUILD_CREATE_ATTENDANCE_BREAK_MODEL_EMPTY_BREAK_END_AT",
			"休憩作成データの作成に失敗しました",
			nil,
		)
	}

	attendanceBreak := models.AttendanceBreak{
		AttendanceDayID: attendanceDayID,
		BreakStartAt:    breakStartAt,
		BreakEndAt:      breakEndAt,
		BreakMemo:       req.BreakMemo,
		IsDeleted:       false,
	}

	return attendanceBreak, results.OK(
		nil,
		"BUILD_CREATE_ATTENDANCE_BREAK_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 休憩更新用Model作成
 */
func (builder *attendanceBreakBuilder) BuildUpdateAttendanceBreakModel(
	currentAttendanceBreak models.AttendanceBreak,
	req types.UpdateAttendanceBreakRequest,
	breakStartAt time.Time,
	breakEndAt time.Time,
) (models.AttendanceBreak, results.Result) {
	if currentAttendanceBreak.ID == 0 {
		return models.AttendanceBreak{}, results.BadRequest(
			"BUILD_UPDATE_ATTENDANCE_BREAK_MODEL_EMPTY_CURRENT_ATTENDANCE_BREAK",
			"休憩更新データの作成に失敗しました",
			nil,
		)
	}

	if breakStartAt.IsZero() {
		return models.AttendanceBreak{}, results.BadRequest(
			"BUILD_UPDATE_ATTENDANCE_BREAK_MODEL_EMPTY_BREAK_START_AT",
			"休憩更新データの作成に失敗しました",
			nil,
		)
	}

	if breakEndAt.IsZero() {
		return models.AttendanceBreak{}, results.BadRequest(
			"BUILD_UPDATE_ATTENDANCE_BREAK_MODEL_EMPTY_BREAK_END_AT",
			"休憩更新データの作成に失敗しました",
			nil,
		)
	}

	currentAttendanceBreak.BreakStartAt = breakStartAt
	currentAttendanceBreak.BreakEndAt = breakEndAt
	currentAttendanceBreak.BreakMemo = req.BreakMemo

	return currentAttendanceBreak, results.OK(
		nil,
		"BUILD_UPDATE_ATTENDANCE_BREAK_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 休憩論理削除用Model作成
 */
func (builder *attendanceBreakBuilder) BuildDeleteAttendanceBreakModel(
	currentAttendanceBreak models.AttendanceBreak,
) (models.AttendanceBreak, results.Result) {
	if currentAttendanceBreak.ID == 0 {
		return models.AttendanceBreak{}, results.BadRequest(
			"BUILD_DELETE_ATTENDANCE_BREAK_MODEL_EMPTY_CURRENT_ATTENDANCE_BREAK",
			"休憩削除データの作成に失敗しました",
			nil,
		)
	}

	now := time.Now()

	currentAttendanceBreak.IsDeleted = true
	currentAttendanceBreak.DeletedAt = &now

	return currentAttendanceBreak, results.OK(
		nil,
		"BUILD_DELETE_ATTENDANCE_BREAK_MODEL_SUCCESS",
		"",
		nil,
	)
}
