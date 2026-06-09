package builders

import (
	"strings"
	"time"

	"timexeed/backend/internal/constants"
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 従業員用 勤怠リアルタイムイベントBuilder interface
 *
 * ServiceがBuilderに求める処理だけを定義する。
 */
type AttendanceRealtimeEventBuilder interface {
	BuildFindTodayEventsByUserIDQuery(userID uint, eventDate time.Time) (*gorm.DB, results.Result)
	BuildCountEventByUserIDDateAndTypeQuery(userID uint, eventDate time.Time, eventType string) (*gorm.DB, results.Result)
	BuildCreateAttendanceRealtimeEventModel(userID uint, eventDate time.Time, eventType string, note string, clientIP string, userAgent string, eventAt time.Time) (models.AttendanceRealtimeEvent, results.Result)
}

/*
 * 従業員用 勤怠リアルタイムイベントBuilder
 *
 * 役割：
 * ・Serviceから受け取った値をもとにGORMクエリを作成する
 * ・Serviceから受け取った値をもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DBアクセスはしない
 * ・query.Find / query.First / db.Create はRepositoryで行う
 * ・業務処理の流れはServiceに任せる
 */
type attendanceRealtimeEventBuilder struct {
	db *gorm.DB
}

/*
 * AttendanceRealtimeEventBuilder生成
 */
func NewAttendanceRealtimeEventBuilder(db *gorm.DB) AttendanceRealtimeEventBuilder {
	return &attendanceRealtimeEventBuilder{db: db}
}

/*
 * 本日の勤怠リアルタイムイベント取得用Query作成
 */
func (builder *attendanceRealtimeEventBuilder) BuildFindTodayEventsByUserIDQuery(
	userID uint,
	eventDate time.Time,
) (*gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, results.InternalServerError(
			"BUILD_FIND_TODAY_ATTENDANCE_REALTIME_EVENTS_QUERY_DB_IS_NIL",
			"本日の勤怠リアルタイムイベント取得条件の作成に失敗しました",
			nil,
		)
	}

	if userID == 0 {
		return nil, results.Unauthorized(
			"BUILD_FIND_TODAY_ATTENDANCE_REALTIME_EVENTS_QUERY_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	query := builder.db.
		Model(&models.AttendanceRealtimeEvent{}).
		Where("user_id = ?", userID).
		Where("event_date = ?", eventDate).
		Order("event_at ASC").
		Order("id ASC")

	return query, results.OK(
		nil,
		"BUILD_FIND_TODAY_ATTENDANCE_REALTIME_EVENTS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 同日同種別イベント件数取得用Query作成
 */
func (builder *attendanceRealtimeEventBuilder) BuildCountEventByUserIDDateAndTypeQuery(
	userID uint,
	eventDate time.Time,
	eventType string,
) (*gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, results.InternalServerError(
			"BUILD_COUNT_ATTENDANCE_REALTIME_EVENT_QUERY_DB_IS_NIL",
			"勤怠リアルタイムイベント件数条件の作成に失敗しました",
			nil,
		)
	}

	if userID == 0 {
		return nil, results.Unauthorized(
			"BUILD_COUNT_ATTENDANCE_REALTIME_EVENT_QUERY_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	eventType = strings.TrimSpace(eventType)
	if !isValidAttendanceRealtimeEventType(eventType) {
		return nil, results.BadRequest(
			"BUILD_COUNT_ATTENDANCE_REALTIME_EVENT_QUERY_INVALID_EVENT_TYPE",
			"勤怠リアルタイムイベント種別が正しくありません",
			map[string]any{
				"eventType": eventType,
			},
		)
	}

	query := builder.db.
		Model(&models.AttendanceRealtimeEvent{}).
		Where("user_id = ?", userID).
		Where("event_date = ?", eventDate).
		Where("event_type = ?", eventType)

	return query, results.OK(
		nil,
		"BUILD_COUNT_ATTENDANCE_REALTIME_EVENT_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠リアルタイムイベント作成用Model作成
 */
func (builder *attendanceRealtimeEventBuilder) BuildCreateAttendanceRealtimeEventModel(
	userID uint,
	eventDate time.Time,
	eventType string,
	note string,
	clientIP string,
	userAgent string,
	eventAt time.Time,
) (models.AttendanceRealtimeEvent, results.Result) {
	eventType = strings.TrimSpace(eventType)
	note = strings.TrimSpace(note)
	clientIP = strings.TrimSpace(clientIP)
	userAgent = strings.TrimSpace(userAgent)

	if userID == 0 {
		return models.AttendanceRealtimeEvent{}, results.Unauthorized(
			"BUILD_CREATE_ATTENDANCE_REALTIME_EVENT_MODEL_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	if !isValidAttendanceRealtimeEventType(eventType) {
		return models.AttendanceRealtimeEvent{}, results.BadRequest(
			"BUILD_CREATE_ATTENDANCE_REALTIME_EVENT_MODEL_INVALID_EVENT_TYPE",
			"勤怠リアルタイムイベント種別が正しくありません",
			map[string]any{
				"eventType": eventType,
			},
		)
	}

	var notePointer *string
	if note != "" {
		notePointer = &note
	}

	var clientIPPointer *string
	if clientIP != "" {
		clientIPPointer = &clientIP
	}

	var userAgentPointer *string
	if userAgent != "" {
		userAgentPointer = &userAgent
	}

	return models.AttendanceRealtimeEvent{
			UserID:    userID,
			EventDate: eventDate,
			EventType: eventType,
			EventAt:   eventAt,
			Note:      notePointer,
			ClientIP:  clientIPPointer,
			UserAgent: userAgentPointer,
		}, results.OK(
			nil,
			"BUILD_CREATE_ATTENDANCE_REALTIME_EVENT_MODEL_SUCCESS",
			"",
			nil,
		)
}

func isValidAttendanceRealtimeEventType(eventType string) bool {
	switch eventType {
	case constants.AttendanceRealtimeEventTypeClockIn:
		return true
	case constants.AttendanceRealtimeEventTypeClockOut:
		return true
	case constants.AttendanceRealtimeEventTypeOther:
		return true
	default:
		return false
	}
}
