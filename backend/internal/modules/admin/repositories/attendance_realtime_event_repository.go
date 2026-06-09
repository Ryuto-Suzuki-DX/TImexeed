package repositories

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用 勤怠リアルタイムイベントRepository interface
 *
 * ServiceがRepositoryに求めるDB処理だけを定義する。
 */
type AttendanceRealtimeEventRepository interface {
	FindAttendanceRealtimeEvents(query *gorm.DB) ([]models.AttendanceRealtimeEvent, results.Result)
	CountAttendanceRealtimeEvents(query *gorm.DB) (int64, results.Result)
}

/*
 * 管理者用 勤怠リアルタイムイベントRepository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 */
type attendanceRealtimeEventRepository struct {
	db *gorm.DB
}

/*
 * AttendanceRealtimeEventRepository生成
 */
func NewAttendanceRealtimeEventRepository(db *gorm.DB) AttendanceRealtimeEventRepository {
	return &attendanceRealtimeEventRepository{db: db}
}

/*
 * 勤怠リアルタイムイベント一覧取得
 */
func (repository *attendanceRealtimeEventRepository) FindAttendanceRealtimeEvents(
	query *gorm.DB,
) ([]models.AttendanceRealtimeEvent, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_ATTENDANCE_REALTIME_EVENTS_QUERY_IS_NIL",
			"勤怠リアルタイムイベント一覧の取得に失敗しました",
			nil,
		)
	}

	var events []models.AttendanceRealtimeEvent

	if err := query.Find(&events).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_ATTENDANCE_REALTIME_EVENTS_FAILED",
			"勤怠リアルタイムイベント一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return events, results.OK(
		nil,
		"FIND_ATTENDANCE_REALTIME_EVENTS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠リアルタイムイベント件数取得
 */
func (repository *attendanceRealtimeEventRepository) CountAttendanceRealtimeEvents(query *gorm.DB) (int64, results.Result) {
	if query == nil {
		return 0, results.InternalServerError(
			"COUNT_ATTENDANCE_REALTIME_EVENTS_QUERY_IS_NIL",
			"勤怠リアルタイムイベント件数の取得に失敗しました",
			nil,
		)
	}

	var count int64

	if err := query.Count(&count).Error; err != nil {
		return 0, results.InternalServerError(
			"COUNT_ATTENDANCE_REALTIME_EVENTS_FAILED",
			"勤怠リアルタイムイベント件数の取得に失敗しました",
			err.Error(),
		)
	}

	return count, results.OK(
		nil,
		"COUNT_ATTENDANCE_REALTIME_EVENTS_SUCCESS",
		"",
		nil,
	)
}
