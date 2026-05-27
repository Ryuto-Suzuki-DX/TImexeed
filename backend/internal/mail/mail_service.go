package mail

import (
	"strings"

	"timexeed/backend/internal/results"
)

/*
 * 〇 共通メール送信Service interface
 *
 * admin/user module から直接Gmail APIを触らないための共通口。
 *
 * 使い方：
 * ・NotificationService がDB通知作成後に呼ぶ
 * ・月次申請Serviceなどの業務ServiceからGmail APIを直接呼ばない
 *
 * 注意：
 * ・メール送信は通知DB作成の副処理
 * ・メール送信に失敗しても、通知DB作成や月次申請処理は原則成功扱いにする
 */
type MailService interface {
	SendMail(message MailMessage) results.Result
	SendNotificationMail(to string, subject string, body string) results.Result
}

/*
 * メール送信無効Service
 *
 * 環境変数未設定などでメール送信を無効化したい場合に使う。
 * アプリ起動自体を止めず、送信処理だけスキップする。
 */
type disabledMailService struct {
	reason string
}

/*
 * メール送信無効Service生成
 */
func NewDisabledMailService(reason string) MailService {
	return &disabledMailService{
		reason: strings.TrimSpace(reason),
	}
}

/*
 * メール送信スキップ
 */
func (service *disabledMailService) SendMail(message MailMessage) results.Result {
	reason := service.reason
	if reason == "" {
		reason = "mail service is disabled"
	}

	return results.OK(
		nil,
		"SEND_MAIL_SKIPPED",
		"メール送信は無効化されています",
		map[string]any{
			"reason":  reason,
			"to":      message.To,
			"subject": message.Subject,
		},
	)
}

/*
 * お知らせメール送信スキップ
 */
func (service *disabledMailService) SendNotificationMail(
	to string,
	subject string,
	body string,
) results.Result {
	return service.SendMail(MailMessage{
		To:      to,
		Subject: subject,
		Body:    body,
	})
}
