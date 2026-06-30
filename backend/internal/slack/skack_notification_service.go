package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"timexeed/backend/internal/constants"
)

/*
 * Slack通知Service
 *
 * 役割：
 * ・Slack Incoming Webhookへ通知を送信する
 * ・リアルタイム勤怠イベントの通知で使用する
 * ・月次勤怠申請の通知で使用する
 *
 * 注意：
 * ・Webhook URLが未設定の場合は送信をスキップする
 * ・Slack通知失敗時の扱いは呼び出し元で決める
 * ・出勤/退勤のDB保存自体をSlack失敗で失敗扱いにしない
 * ・月次勤怠申請のDB保存自体をSlack失敗で失敗扱いにしない
 */
type SlackNotificationService interface {
	SendAttendanceRealtimeEventNotification(
		req AttendanceRealtimeEventSlackNotificationRequest,
	) error

	SendMonthlyAttendanceRequestNotification(
		req MonthlyAttendanceRequestSlackNotificationRequest,
	) error
}

/*
 * 勤怠リアルタイムイベントSlack通知Request
 *
 * メールアドレスはSlackへ送信しない。
 */
type AttendanceRealtimeEventSlackNotificationRequest struct {
	EventType string
	UserName  string
	EventAt   time.Time
	Note      *string
}

/*
 * 月次勤怠申請Slack通知Request
 *
 * メールアドレスはSlackへ送信しない。
 */
type MonthlyAttendanceRequestSlackNotificationRequest struct {
	Action         string
	UserName       string
	TargetYear     int
	TargetMonth    int
	RequestMemo    *string
	CanceledReason *string
}

type slackNotificationService struct {
	attendanceWebhookURL        string
	monthlyAttendanceWebhookURL string
	httpClient                  *http.Client
}

/*
 * 環境変数からSlack通知Serviceを生成する
 *
 * 環境変数：
 * TIMEXEED_SLACK_ATTENDANCE_WEBHOOK_URL
 * TIMEXEED_SLACK_MONTHLY_ATTENDANCE_WEBHOOK_URL
 */
func NewSlackNotificationServiceFromEnv() SlackNotificationService {
	return &slackNotificationService{
		attendanceWebhookURL: strings.TrimSpace(
			os.Getenv("TIMEXEED_SLACK_ATTENDANCE_WEBHOOK_URL"),
		),
		monthlyAttendanceWebhookURL: strings.TrimSpace(
			os.Getenv("TIMEXEED_SLACK_MONTHLY_ATTENDANCE_WEBHOOK_URL"),
		),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

/*
 * 勤怠リアルタイムイベント通知をSlackへ送信する。
 */
func (service *slackNotificationService) SendAttendanceRealtimeEventNotification(
	req AttendanceRealtimeEventSlackNotificationRequest,
) error {
	if service == nil {
		return fmt.Errorf("slack notification service is nil")
	}

	if service.attendanceWebhookURL == "" {
		/*
		 * Slack未設定でもアプリ本体は動かす。
		 * ローカル環境や検証環境でWebhook未設定でもエラーにしない。
		 */
		return nil
	}

	message := buildAttendanceRealtimeEventSlackMessage(req)

	return service.sendSlackWebhook(
		service.attendanceWebhookURL,
		message,
	)
}

/*
 * 月次勤怠申請通知をSlackへ送信する。
 */
func (service *slackNotificationService) SendMonthlyAttendanceRequestNotification(
	req MonthlyAttendanceRequestSlackNotificationRequest,
) error {
	if service == nil {
		return fmt.Errorf("slack notification service is nil")
	}

	if service.monthlyAttendanceWebhookURL == "" {
		/*
		 * Slack未設定でもアプリ本体は動かす。
		 * ローカル環境や検証環境でWebhook未設定でもエラーにしない。
		 */
		return nil
	}

	message := buildMonthlyAttendanceRequestSlackMessage(req)

	return service.sendSlackWebhook(
		service.monthlyAttendanceWebhookURL,
		message,
	)
}

/*
 * Slack Incoming Webhookへ本文を送信する。
 */
func (service *slackNotificationService) sendSlackWebhook(
	webhookURL string,
	message string,
) error {
	payload := map[string]string{
		"text": message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		webhookURL,
		bytes.NewReader(payloadBytes),
	)
	if err != nil {
		return fmt.Errorf("failed to create slack request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpClient := service.httpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send slack request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf(
			"slack returned non-2xx status: %d",
			resp.StatusCode,
		)
	}

	return nil
}

/*
 * Slackへ送信する本文を作成する。
 */
func buildAttendanceRealtimeEventSlackMessage(
	req AttendanceRealtimeEventSlackNotificationRequest,
) string {
	eventTypeLabel := attendanceRealtimeEventTypeLabel(req.EventType)
	eventAtText := formatSlackEventAt(req.EventAt)

	noteText := "なし"
	if req.Note != nil {
		trimmedNote := strings.TrimSpace(*req.Note)
		if trimmedNote != "" {
			noteText = trimmedNote
		}
	}

	userName := strings.TrimSpace(req.UserName)
	if userName == "" {
		userName = "不明"
	}

	return fmt.Sprintf(
		"【Timexeed 勤怠リアルタイム通知】\n\n種別：%s\n氏名：%s\n押下日時：%s\nコメント：%s",
		eventTypeLabel,
		userName,
		eventAtText,
		noteText,
	)
}

/*
 * 月次勤怠申請通知のSlack本文を作成する。
 */
func buildMonthlyAttendanceRequestSlackMessage(
	req MonthlyAttendanceRequestSlackNotificationRequest,
) string {
	actionLabel := monthlyAttendanceRequestActionLabel(req.Action)

	userName := strings.TrimSpace(req.UserName)
	if userName == "" {
		userName = "不明"
	}

	detailText := "なし"

	switch strings.ToUpper(strings.TrimSpace(req.Action)) {
	case "CANCELED":
		if req.CanceledReason != nil {
			trimmedCanceledReason := strings.TrimSpace(*req.CanceledReason)
			if trimmedCanceledReason != "" {
				detailText = trimmedCanceledReason
			}
		}

	default:
		if req.RequestMemo != nil {
			trimmedRequestMemo := strings.TrimSpace(*req.RequestMemo)
			if trimmedRequestMemo != "" {
				detailText = trimmedRequestMemo
			}
		}
	}

	return fmt.Sprintf(
		"【Timexeed 月次勤怠申請通知】\n\n種別：%s\n氏名：%s\n対象月：%04d年%d月\nメモ・理由：%s",
		actionLabel,
		userName,
		req.TargetYear,
		req.TargetMonth,
		detailText,
	)
}

/*
 * イベント種別をSlack表示用の日本語へ変換する。
 */
func attendanceRealtimeEventTypeLabel(eventType string) string {
	switch strings.ToUpper(strings.TrimSpace(eventType)) {
	case constants.AttendanceRealtimeEventTypeClockIn:
		return "出勤"

	case constants.AttendanceRealtimeEventTypeClockOut:
		return "退勤"

	default:
		return eventType
	}
}

/*
 * 月次勤怠申請操作をSlack表示用の日本語へ変換する。
 */
func monthlyAttendanceRequestActionLabel(action string) string {
	switch strings.ToUpper(strings.TrimSpace(action)) {
	case "SUBMITTED":
		return "申請"

	case "RESUBMITTED":
		return "再申請"

	case "CANCELED":
		return "取り下げ"

	case "APPROVED":
		return "承認"

	case "REJECTED":
		return "否認"

	default:
		return action
	}
}

/*
 * イベント日時をJSTで表示する。
 */
func formatSlackEventAt(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		location = time.FixedZone("JST", 9*60*60)
	}

	return value.In(location).Format("2006-01-02 15:04:05")
}
