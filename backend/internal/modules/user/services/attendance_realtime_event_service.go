package services

import (
	"strings"
	"time"

	"timexeed/backend/internal/constants"
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
)

/*
 * 従業員用 勤怠リアルタイムイベントService interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・従業員APIでは userId / targetUserId をRequestで受け取らない
 * ・ControllerでAuthMiddleware由来のuserIdを取得し、Serviceへ渡す
 * ・月次勤怠には反映しない
 * ・登録後の取消・編集はしない
 */
type AttendanceRealtimeEventService interface {
	CreateAttendanceRealtimeEvent(userID uint, req types.CreateAttendanceRealtimeEventRequest, clientIP string, userAgent string) results.Result
	GetTodayAttendanceRealtimeEvents(userID uint, req types.GetTodayAttendanceRealtimeEventsRequest) results.Result
}

/*
 * 従業員用 勤怠リアルタイムイベントService
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや保存用Modelを作成する
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
 * models.AttendanceRealtimeEventをフロント返却用AttendanceRealtimeEventResponseへ変換する
 */
func toAttendanceRealtimeEventResponse(
	event models.AttendanceRealtimeEvent,
) types.AttendanceRealtimeEventResponse {
	return types.AttendanceRealtimeEventResponse{
		ID:        event.ID,
		EventDate: event.EventDate,
		EventType: event.EventType,
		EventAt:   event.EventAt,
		Note:      event.Note,
		CreatedAt: event.CreatedAt,
	}
}

/*
 * 勤怠リアルタイムイベント作成
 */
func (service *attendanceRealtimeEventService) CreateAttendanceRealtimeEvent(
	userID uint,
	req types.CreateAttendanceRealtimeEventRequest,
	clientIP string,
	userAgent string,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"CREATE_ATTENDANCE_REALTIME_EVENT_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	eventType := strings.ToUpper(strings.TrimSpace(req.EventType))
	if !isValidAttendanceRealtimeEventTypeForUser(eventType) {
		return results.BadRequest(
			"CREATE_ATTENDANCE_REALTIME_EVENT_INVALID_EVENT_TYPE",
			"勤怠リアルタイムイベント種別が正しくありません",
			map[string]any{
				"eventType": req.EventType,
			},
		)
	}

	now := time.Now()
	eventDate := buildAttendanceRealtimeEventDate(now)

	countQuery, buildCountResult := service.attendanceRealtimeEventBuilder.BuildCountEventByUserIDDateAndTypeQuery(
		userID,
		eventDate,
		eventType,
	)
	if buildCountResult.Error {
		return buildCountResult
	}

	count, countResult := service.attendanceRealtimeEventRepository.CountAttendanceRealtimeEvents(countQuery)
	if countResult.Error {
		return countResult
	}

	if count > 0 {
		return results.BadRequest(
			"ATTENDANCE_REALTIME_EVENT_ALREADY_RECORDED",
			"本日のこの操作はすでに記録済みです",
			map[string]any{
				"eventType": eventType,
			},
		)
	}

	event, buildCreateResult := service.attendanceRealtimeEventBuilder.BuildCreateAttendanceRealtimeEventModel(
		userID,
		eventDate,
		eventType,
		req.Note,
		clientIP,
		userAgent,
		now,
	)
	if buildCreateResult.Error {
		return buildCreateResult
	}

	createdEvent, createResult := service.attendanceRealtimeEventRepository.CreateAttendanceRealtimeEvent(event)
	if createResult.Error {
		return createResult
	}

	return results.Created(
		types.CreateAttendanceRealtimeEventResponse{
			Event: toAttendanceRealtimeEventResponse(createdEvent),
		},
		"CREATE_ATTENDANCE_REALTIME_EVENT_SUCCESS",
		"勤怠リアルタイムイベントを記録しました",
		nil,
	)
}

/*
 * 本日の勤怠リアルタイムイベント状態取得
 */
func (service *attendanceRealtimeEventService) GetTodayAttendanceRealtimeEvents(
	userID uint,
	req types.GetTodayAttendanceRealtimeEventsRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"GET_TODAY_ATTENDANCE_REALTIME_EVENTS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	now := time.Now()
	eventDate := buildAttendanceRealtimeEventDate(now)

	query, buildResult := service.attendanceRealtimeEventBuilder.BuildFindTodayEventsByUserIDQuery(userID, eventDate)
	if buildResult.Error {
		return buildResult
	}

	events, findResult := service.attendanceRealtimeEventRepository.FindAttendanceRealtimeEvents(query)
	if findResult.Error {
		return findResult
	}

	response := types.GetTodayAttendanceRealtimeEventsResponse{
		Events: make([]types.AttendanceRealtimeEventResponse, 0, len(events)),
	}

	for _, event := range events {
		eventResponse := toAttendanceRealtimeEventResponse(event)
		response.Events = append(response.Events, eventResponse)

		switch event.EventType {
		case constants.AttendanceRealtimeEventTypeClockIn:
			response.ClockInRecorded = true
			eventAt := event.EventAt
			response.ClockInAt = &eventAt
		case constants.AttendanceRealtimeEventTypeClockOut:
			response.ClockOutRecorded = true
			eventAt := event.EventAt
			response.ClockOutAt = &eventAt
		case constants.AttendanceRealtimeEventTypeOther:
			response.OtherRecorded = true
			eventAt := event.EventAt
			response.OtherAt = &eventAt
		}
	}

	return results.OK(
		response,
		"GET_TODAY_ATTENDANCE_REALTIME_EVENTS_SUCCESS",
		"本日の勤怠リアルタイムイベント状態を取得しました",
		nil,
	)
}

func buildAttendanceRealtimeEventDate(value time.Time) time.Time {
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		location = time.FixedZone("Asia/Tokyo", 9*60*60)
	}

	jstValue := value.In(location)
	return time.Date(jstValue.Year(), jstValue.Month(), jstValue.Day(), 0, 0, 0, 0, location)
}

func isValidAttendanceRealtimeEventTypeForUser(eventType string) bool {
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
