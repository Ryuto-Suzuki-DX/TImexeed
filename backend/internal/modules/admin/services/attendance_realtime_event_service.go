package services

import (
	"strings"
	"time"

	"timexeed/backend/internal/constants"
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
)

/*
 * 管理者用 勤怠リアルタイムイベントService interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type AttendanceRealtimeEventService interface {
	SearchAttendanceRealtimeEvents(req types.SearchAttendanceRealtimeEventsRequest) results.Result
}

/*
 * 管理者用 勤怠リアルタイムイベントService
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリを作成する
 * ・RepositoryでDB処理を実行する
 * ・成功時はResponse型に変換してControllerへ返す
 */
type attendanceRealtimeEventService struct {
	attendanceRealtimeEventBuilder    builders.AttendanceRealtimeEventBuilder
	attendanceRealtimeEventRepository repositories.AttendanceRealtimeEventRepository
}

/*
 * AttendanceRealtimeEventService生成
 */
func NewAttendanceRealtimeEventService(
	attendanceRealtimeEventBuilder builders.AttendanceRealtimeEventBuilder,
	attendanceRealtimeEventRepository repositories.AttendanceRealtimeEventRepository,
) *attendanceRealtimeEventService {
	return &attendanceRealtimeEventService{
		attendanceRealtimeEventBuilder:    attendanceRealtimeEventBuilder,
		attendanceRealtimeEventRepository: attendanceRealtimeEventRepository,
	}
}

/*
 * models.AttendanceRealtimeEventを管理者返却用AttendanceRealtimeEventResponseへ変換する
 */
func toAttendanceRealtimeEventResponse(
	event models.AttendanceRealtimeEvent,
) types.AttendanceRealtimeEventResponse {
	return types.AttendanceRealtimeEventResponse{
		ID:        event.ID,
		UserID:    event.UserID,
		UserName:  event.User.Name,
		UserEmail: event.User.Email,
		EventDate: event.EventDate,
		EventType: event.EventType,
		EventAt:   event.EventAt,
		Note:      event.Note,
		ClientIP:  event.ClientIP,
		UserAgent: event.UserAgent,
		CreatedAt: event.CreatedAt,
	}
}

/*
 * 勤怠リアルタイムイベント検索
 */
func (service *attendanceRealtimeEventService) SearchAttendanceRealtimeEvents(
	req types.SearchAttendanceRealtimeEventsRequest,
) results.Result {
	if req.Limit <= 0 {
		req.Limit = 50
	}

	if req.Offset < 0 {
		return results.BadRequest(
			"SEARCH_ATTENDANCE_REALTIME_EVENTS_INVALID_OFFSET",
			"勤怠リアルタイムイベント検索の開始位置が正しくありません",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	if len(req.EventTypes) > 0 {
		for _, eventType := range req.EventTypes {
			normalizedEventType := strings.ToUpper(strings.TrimSpace(eventType))
			if !isValidAttendanceRealtimeEventTypeForAdmin(normalizedEventType) {
				return results.BadRequest(
					"SEARCH_ATTENDANCE_REALTIME_EVENTS_INVALID_EVENT_TYPE",
					"勤怠リアルタイムイベント種別が正しくありません",
					map[string]any{
						"eventType": eventType,
					},
				)
			}
		}
	}

	eventDate, parseResult := parseAttendanceRealtimeEventSearchDate(req.TargetDate)
	if parseResult.Error {
		return parseResult
	}

	countQuery, buildCountResult := service.attendanceRealtimeEventBuilder.BuildCountSearchAttendanceRealtimeEventsQuery(eventDate, req)
	if buildCountResult.Error {
		return buildCountResult
	}

	total, countResult := service.attendanceRealtimeEventRepository.CountAttendanceRealtimeEvents(countQuery)
	if countResult.Error {
		return countResult
	}

	query, buildSearchResult := service.attendanceRealtimeEventBuilder.BuildSearchAttendanceRealtimeEventsQuery(eventDate, req)
	if buildSearchResult.Error {
		return buildSearchResult
	}

	events, findResult := service.attendanceRealtimeEventRepository.FindAttendanceRealtimeEvents(query)
	if findResult.Error {
		return findResult
	}

	hasMore := int64(req.Offset+len(events)) < total

	eventResponses := make([]types.AttendanceRealtimeEventResponse, 0, len(events))
	for _, event := range events {
		eventResponses = append(eventResponses, toAttendanceRealtimeEventResponse(event))
	}

	return results.OK(
		types.SearchAttendanceRealtimeEventsResponse{
			Events:  eventResponses,
			Total:   total,
			Offset:  req.Offset,
			Limit:   req.Limit,
			HasMore: hasMore,
		},
		"SEARCH_ATTENDANCE_REALTIME_EVENTS_SUCCESS",
		"勤怠リアルタイムイベント一覧を取得しました",
		nil,
	)
}

func parseAttendanceRealtimeEventSearchDate(targetDate string) (time.Time, results.Result) {
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		location = time.FixedZone("Asia/Tokyo", 9*60*60)
	}

	targetDate = strings.TrimSpace(targetDate)
	if targetDate == "" {
		now := time.Now().In(location)
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location), results.OK(
			nil,
			"PARSE_ATTENDANCE_REALTIME_EVENT_SEARCH_DATE_TODAY_SUCCESS",
			"",
			nil,
		)
	}

	parsedDate, err := time.ParseInLocation("2006-01-02", targetDate, location)
	if err != nil {
		return time.Time{}, results.BadRequest(
			"SEARCH_ATTENDANCE_REALTIME_EVENTS_INVALID_TARGET_DATE",
			"対象日はYYYY-MM-DD形式で入力してください",
			map[string]any{
				"targetDate": targetDate,
			},
		)
	}

	return parsedDate, results.OK(
		nil,
		"PARSE_ATTENDANCE_REALTIME_EVENT_SEARCH_DATE_SUCCESS",
		"",
		nil,
	)
}

func isValidAttendanceRealtimeEventTypeForAdmin(eventType string) bool {
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
