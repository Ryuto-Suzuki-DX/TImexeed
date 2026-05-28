package mail

import (
	"fmt"
	"mime"
	netmail "net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"

	"timexeed/backend/internal/results"
)

/*
 * 〇 SMTPメール送信Service
 *
 * Gmail API / Google Workspace / Domain-wide delegation を使わず、
 * SMTPサーバー経由でメール送信するためのService。
 *
 * 想定用途：
 * ・Resend / SendGrid / Brevo / Mailgun などのSMTP
 * ・その他SMTPリレーサービス
 *
 * 必要な環境変数：
 * ・TIMEXEED_MAIL_HOST
 * ・TIMEXEED_MAIL_PORT
 * ・TIMEXEED_MAIL_USERNAME
 * ・TIMEXEED_MAIL_PASSWORD
 * ・TIMEXEED_MAIL_FROM
 *
 * 任意の環境変数：
 * ・TIMEXEED_MAIL_SENDER_NAME
 * ・TIMEXEED_MAIL_SUBJECT_PREFIX
 *
 * 注意：
 * ・基本は587番ポートのSTARTTLS対応SMTPを想定
 * ・465番のimplicit TLS SMTPにはこの実装では未対応
 * ・環境変数未設定時はアプリ起動を止めず、メール送信無効Serviceを返す
 */
type smtpMailService struct {
	host          string
	port          int
	username      string
	password      string
	from          string
	senderName    string
	subjectPrefix string
}

/*
 * SMTPメール送信Service生成
 *
 * 環境変数が不足している場合は、起動を止めないためDisabledMailServiceを返す。
 */
func NewSMTPMailServiceFromEnv() (MailService, results.Result) {
	host := strings.TrimSpace(os.Getenv("TIMEXEED_MAIL_HOST"))
	portText := strings.TrimSpace(os.Getenv("TIMEXEED_MAIL_PORT"))
	username := strings.TrimSpace(os.Getenv("TIMEXEED_MAIL_USERNAME"))
	password := strings.TrimSpace(os.Getenv("TIMEXEED_MAIL_PASSWORD"))
	from := strings.TrimSpace(os.Getenv("TIMEXEED_MAIL_FROM"))
	senderName := strings.TrimSpace(os.Getenv("TIMEXEED_MAIL_SENDER_NAME"))
	subjectPrefix := strings.TrimSpace(os.Getenv("TIMEXEED_MAIL_SUBJECT_PREFIX"))

	if host == "" {
		return NewDisabledMailService("TIMEXEED_MAIL_HOST is not set"), results.OK(
			nil,
			"NEW_SMTP_MAIL_SERVICE_SKIPPED_HOST_EMPTY",
			"SMTPホストが未設定のため、メール送信を無効化しました",
			nil,
		)
	}

	if portText == "" {
		return NewDisabledMailService("TIMEXEED_MAIL_PORT is not set"), results.OK(
			nil,
			"NEW_SMTP_MAIL_SERVICE_SKIPPED_PORT_EMPTY",
			"SMTPポートが未設定のため、メール送信を無効化しました",
			nil,
		)
	}

	port, parseErr := strconv.Atoi(portText)
	if parseErr != nil || port <= 0 {
		return NewDisabledMailService("TIMEXEED_MAIL_PORT is invalid"), results.OK(
			nil,
			"NEW_SMTP_MAIL_SERVICE_SKIPPED_PORT_INVALID",
			"SMTPポートが不正なため、メール送信を無効化しました",
			map[string]any{
				"port": portText,
			},
		)
	}

	if username == "" {
		return NewDisabledMailService("TIMEXEED_MAIL_USERNAME is not set"), results.OK(
			nil,
			"NEW_SMTP_MAIL_SERVICE_SKIPPED_USERNAME_EMPTY",
			"SMTPユーザー名が未設定のため、メール送信を無効化しました",
			nil,
		)
	}

	if password == "" {
		return NewDisabledMailService("TIMEXEED_MAIL_PASSWORD is not set"), results.OK(
			nil,
			"NEW_SMTP_MAIL_SERVICE_SKIPPED_PASSWORD_EMPTY",
			"SMTPパスワードが未設定のため、メール送信を無効化しました",
			nil,
		)
	}

	if from == "" {
		return NewDisabledMailService("TIMEXEED_MAIL_FROM is not set"), results.OK(
			nil,
			"NEW_SMTP_MAIL_SERVICE_SKIPPED_FROM_EMPTY",
			"メール送信元が未設定のため、メール送信を無効化しました",
			nil,
		)
	}

	if _, err := netmail.ParseAddress(from); err != nil {
		return NewDisabledMailService("TIMEXEED_MAIL_FROM is invalid"), results.OK(
			nil,
			"NEW_SMTP_MAIL_SERVICE_SKIPPED_INVALID_FROM",
			"メール送信元が不正なため、メール送信を無効化しました",
			err.Error(),
		)
	}

	return &smtpMailService{
			host:          host,
			port:          port,
			username:      username,
			password:      password,
			from:          from,
			senderName:    senderName,
			subjectPrefix: subjectPrefix,
		}, results.OK(
			nil,
			"NEW_SMTP_MAIL_SERVICE_SUCCESS",
			"SMTPメール送信Serviceを生成しました",
			nil,
		)
}

/*
 * お知らせメール送信
 */
func (service *smtpMailService) SendNotificationMail(
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

/*
 * メール送信
 */
func (service *smtpMailService) SendMail(message MailMessage) results.Result {
	if service == nil {
		return results.InternalServerError(
			"SEND_SMTP_MAIL_SERVICE_IS_NIL",
			"メール送信に失敗しました",
			nil,
		)
	}

	to := strings.TrimSpace(message.To)
	subject := strings.TrimSpace(message.Subject)
	body := strings.TrimSpace(message.Body)

	if to == "" {
		return results.BadRequest(
			"SEND_SMTP_MAIL_TO_EMPTY",
			"メール送信先が空です",
			nil,
		)
	}

	if _, err := netmail.ParseAddress(to); err != nil {
		return results.BadRequest(
			"SEND_SMTP_MAIL_TO_INVALID",
			"メール送信先が正しくありません",
			err.Error(),
		)
	}

	if subject == "" {
		return results.BadRequest(
			"SEND_SMTP_MAIL_SUBJECT_EMPTY",
			"メール件名が空です",
			nil,
		)
	}

	if body == "" {
		return results.BadRequest(
			"SEND_SMTP_MAIL_BODY_EMPTY",
			"メール本文が空です",
			nil,
		)
	}

	finalSubject := service.buildSubject(subject)
	rawMessage := service.buildRawMessage(to, finalSubject, body)

	address := fmt.Sprintf("%s:%d", service.host, service.port)
	auth := smtp.PlainAuth("", service.username, service.password, service.host)

	if err := smtp.SendMail(
		address,
		auth,
		service.from,
		[]string{to},
		[]byte(rawMessage),
	); err != nil {
		return results.InternalServerError(
			"SEND_SMTP_MAIL_FAILED",
			"メール送信に失敗しました",
			err.Error(),
		)
	}

	return results.OK(
		nil,
		"SEND_SMTP_MAIL_SUCCESS",
		"メールを送信しました",
		map[string]any{
			"to":      to,
			"subject": finalSubject,
		},
	)
}

/*
 * 件名作成
 */
func (service *smtpMailService) buildSubject(subject string) string {
	subject = strings.TrimSpace(subject)

	if service.subjectPrefix == "" {
		return subject
	}

	if strings.HasPrefix(subject, service.subjectPrefix) {
		return subject
	}

	return fmt.Sprintf("%s %s", service.subjectPrefix, subject)
}

/*
 * SMTP送信用Raw Message作成
 */
func (service *smtpMailService) buildRawMessage(to string, subject string, body string) string {
	fromHeader := service.from
	if service.senderName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("UTF-8", service.senderName), service.from)
	}

	encodedSubject := mime.QEncoding.Encode("UTF-8", sanitizeSMTPMailHeader(subject))

	return strings.Join([]string{
		fmt.Sprintf("From: %s", fromHeader),
		fmt.Sprintf("To: %s", sanitizeSMTPMailHeader(to)),
		fmt.Sprintf("Subject: %s", encodedSubject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
}

/*
 * メールヘッダーインジェクション対策
 */
func sanitizeSMTPMailHeader(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return strings.TrimSpace(value)
}
