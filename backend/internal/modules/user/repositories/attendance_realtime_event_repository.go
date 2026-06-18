package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 従業員用 勤怠リアルタイムイベントRepository interface
 *
 * ServiceがRepositoryに求めるDB処理だけを定義する。
 */
type AttendanceRealtimeEventRepository interface {
	FindAttendanceRealtimeEvents(query *gorm.DB) ([]models.AttendanceRealtimeEvent, results.Result)
	FindAttendanceRealtimeEventByIDWithUser(eventID uint) (models.AttendanceRealtimeEvent, results.Result)
	CountAttendanceRealtimeEvents(query *gorm.DB) (int64, results.Result)
	CreateAttendanceRealtimeEvent(event models.AttendanceRealtimeEvent) (models.AttendanceRealtimeEvent, results.Result)
}

/*
 * 従業員用 勤怠リアルタイムイベントRepository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・DBへのCreateを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 * ・二重登録可否などはServiceに任せる
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
 * 勤怠リアルタイムイベントをUser情報付きで1件取得する。
 *
 * 用途：
 * ・Slack通知で氏名/メールアドレスを表示するため
 */
func (repository *attendanceRealtimeEventRepository) FindAttendanceRealtimeEventByIDWithUser(
	eventID uint,
) (models.AttendanceRealtimeEvent, results.Result) {
	if repository.db == nil {
		return models.AttendanceRealtimeEvent{}, results.InternalServerError(
			"FIND_ATTENDANCE_REALTIME_EVENT_BY_ID_WITH_USER_DB_IS_NIL",
			"勤怠リアルタイムイベントの取得に失敗しました",
			nil,
		)
	}

	if eventID == 0 {
		return models.AttendanceRealtimeEvent{}, results.InternalServerError(
			"FIND_ATTENDANCE_REALTIME_EVENT_BY_ID_WITH_USER_EMPTY_EVENT_ID",
			"勤怠リアルタイムイベントの取得に失敗しました",
			nil,
		)
	}

	var event models.AttendanceRealtimeEvent

	if err := repository.db.
		Preload("User").
		First(&event, eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AttendanceRealtimeEvent{}, results.NotFound(
				"FIND_ATTENDANCE_REALTIME_EVENT_BY_ID_WITH_USER_NOT_FOUND",
				"勤怠リアルタイムイベントが見つかりません",
				map[string]any{
					"eventId": eventID,
				},
			)
		}

		return models.AttendanceRealtimeEvent{}, results.InternalServerError(
			"FIND_ATTENDANCE_REALTIME_EVENT_BY_ID_WITH_USER_FAILED",
			"勤怠リアルタイムイベントの取得に失敗しました",
			err.Error(),
		)
	}

	return event, results.OK(
		nil,
		"FIND_ATTENDANCE_REALTIME_EVENT_BY_ID_WITH_USER_SUCCESS",
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

/*
 * 勤怠リアルタイムイベント作成
 */
func (repository *attendanceRealtimeEventRepository) CreateAttendanceRealtimeEvent(
	event models.AttendanceRealtimeEvent,
) (models.AttendanceRealtimeEvent, results.Result) {
	if event.UserID == 0 {
		return models.AttendanceRealtimeEvent{}, results.InternalServerError(
			"CREATE_ATTENDANCE_REALTIME_EVENT_EMPTY_USER_ID",
			"勤怠リアルタイムイベントの作成に失敗しました",
			nil,
		)
	}

	if event.EventType == "" {
		return models.AttendanceRealtimeEvent{}, results.InternalServerError(
			"CREATE_ATTENDANCE_REALTIME_EVENT_EMPTY_EVENT_TYPE",
			"勤怠リアルタイムイベントの作成に失敗しました",
			nil,
		)
	}

	if err := repository.db.Create(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return models.AttendanceRealtimeEvent{}, results.BadRequest(
				"ATTENDANCE_REALTIME_EVENT_ALREADY_RECORDED",
				"本日のこの操作はすでに記録済みです",
				nil,
			)
		}

		return models.AttendanceRealtimeEvent{}, results.InternalServerError(
			"CREATE_ATTENDANCE_REALTIME_EVENT_FAILED",
			"勤怠リアルタイムイベントの作成に失敗しました",
			err.Error(),
		)
	}

	return event, results.Created(
		nil,
		"CREATE_ATTENDANCE_REALTIME_EVENT_SUCCESS",
		"",
		nil,
	)
}
