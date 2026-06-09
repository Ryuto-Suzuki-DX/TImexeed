package builders

import (
	"strings"
	"time"

	"timexeed/backend/internal/constants"
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用 勤怠リアルタイムイベントBuilder interface
 *
 * ServiceがBuilderに求める処理だけを定義する。
 */
type AttendanceRealtimeEventBuilder interface {
	BuildSearchAttendanceRealtimeEventsQuery(eventDate time.Time, req types.SearchAttendanceRealtimeEventsRequest) (*gorm.DB, results.Result)
	BuildCountSearchAttendanceRealtimeEventsQuery(eventDate time.Time, req types.SearchAttendanceRealtimeEventsRequest) (*gorm.DB, results.Result)
}

/*
 * 管理者用 勤怠リアルタイムイベントBuilder
 *
 * 役割：
 * ・Serviceから受け取った値をもとにGORMクエリを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DBアクセスはしない
 * ・query.Find / query.Count はRepositoryで行う
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
 * 勤怠リアルタイムイベント検索用Query作成
 */
func (builder *attendanceRealtimeEventBuilder) BuildSearchAttendanceRealtimeEventsQuery(
	eventDate time.Time,
	req types.SearchAttendanceRealtimeEventsRequest,
) (*gorm.DB, results.Result) {
	query, result := builder.buildBaseSearchAttendanceRealtimeEventsQuery(eventDate, req)
	if result.Error {
		return nil, result
	}

	query = query.
		Preload("User").
		Order("attendance_realtime_events.event_at DESC").
		Order("attendance_realtime_events.id DESC").
		Offset(req.Offset).
		Limit(req.Limit)

	return query, results.OK(
		nil,
		"BUILD_SEARCH_ATTENDANCE_REALTIME_EVENTS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠リアルタイムイベント検索件数取得用Query作成
 */
func (builder *attendanceRealtimeEventBuilder) BuildCountSearchAttendanceRealtimeEventsQuery(
	eventDate time.Time,
	req types.SearchAttendanceRealtimeEventsRequest,
) (*gorm.DB, results.Result) {
	query, result := builder.buildBaseSearchAttendanceRealtimeEventsQuery(eventDate, req)
	if result.Error {
		return nil, result
	}

	return query, results.OK(
		nil,
		"BUILD_COUNT_SEARCH_ATTENDANCE_REALTIME_EVENTS_QUERY_SUCCESS",
		"",
		nil,
	)
}

func (builder *attendanceRealtimeEventBuilder) buildBaseSearchAttendanceRealtimeEventsQuery(
	eventDate time.Time,
	req types.SearchAttendanceRealtimeEventsRequest,
) (*gorm.DB, results.Result) {
	if builder.db == nil {
		return nil, results.InternalServerError(
			"BUILD_ATTENDANCE_REALTIME_EVENTS_QUERY_DB_IS_NIL",
			"勤怠リアルタイムイベント検索条件の作成に失敗しました",
			nil,
		)
	}

	query := builder.db.
		Model(&models.AttendanceRealtimeEvent{}).
		Joins("JOIN users ON users.id = attendance_realtime_events.user_id").
		Where("attendance_realtime_events.event_date = ?", eventDate).
		Where("users.is_deleted = ?", false)

	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.Where(
			"(users.name ILIKE ? OR users.email ILIKE ?)",
			likeKeyword,
			likeKeyword,
		)
	}

	eventTypes := normalizeAdminAttendanceRealtimeEventTypes(req.EventTypes)
	if len(eventTypes) > 0 {
		query = query.Where("attendance_realtime_events.event_type IN ?", eventTypes)
	}

	return query, results.OK(
		nil,
		"BUILD_ATTENDANCE_REALTIME_EVENTS_QUERY_SUCCESS",
		"",
		nil,
	)
}

func normalizeAdminAttendanceRealtimeEventTypes(eventTypes []string) []string {
	normalized := make([]string, 0, len(eventTypes))
	exists := make(map[string]bool)

	for _, eventType := range eventTypes {
		eventType = strings.ToUpper(strings.TrimSpace(eventType))
		if eventType == "" {
			continue
		}

		if !isValidAdminAttendanceRealtimeEventType(eventType) {
			continue
		}

		if exists[eventType] {
			continue
		}

		exists[eventType] = true
		normalized = append(normalized, eventType)
	}

	return normalized
}

func isValidAdminAttendanceRealtimeEventType(eventType string) bool {
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
