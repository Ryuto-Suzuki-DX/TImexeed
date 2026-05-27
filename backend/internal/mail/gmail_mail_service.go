package mail

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	netmail "net/mail"
	"os"
	"strings"

	"timexeed/backend/internal/results"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

/*
 * 〇 Gmail APIメール送信Service
 *
 * Google Workspaceのアプリ用アカウントからメール送信するためのService。
 *
 * 想定方式：
 * ・サービスアカウントJSONを使用する
 * ・Google Workspace側でDomain-wide delegationを設定する
 * ・TIMEXEED_GMAIL_FROM に指定した実ユーザーを代理実行して送信する
 *
 * 必要な環境変数：
 * ・TIMEXEED_GMAIL_CREDENTIALS_FILE
 *   Gmail API送信用サービスアカウントJSONのパス。
 *   未指定の場合は GOOGLE_APPLICATION_CREDENTIALS も見る。
 *
 * ・TIMEXEED_GMAIL_FROM
 *   送信元のGoogle Workspace実ユーザー。
 *   例：app@example.com
 *
 * 任意の環境変数：
 * ・TIMEXEED_GMAIL_SENDER_NAME
 *   送信者表示名。
 *   例：Timexeed
 *
 * ・TIMEXEED_GMAIL_SUBJECT_PREFIX
 *   件名接頭辞。
 *   例：[Timexeed]
 *
 * 注意：
 * ・サービスアカウント自身から送るのではなく、FROMの実ユーザーを代理実行する
 * ・環境変数未設定時はアプリ起動を止めず、メール送信無効Serviceを返す
 */
type gmailMailService struct {
	gmailService  *gmail.Service
	from          string
	senderName    string
	subjectPrefix string
}

/*
 * Gmail APIメール送信Service生成
 *
 * 環境変数が不足している場合は、起動を止めないためDisabledMailServiceを返す。
 */
func NewGmailMailServiceFromEnv(ctx context.Context) (MailService, results.Result) {
	credentialsFile := strings.TrimSpace(os.Getenv("TIMEXEED_GMAIL_CREDENTIALS_FILE"))
	if credentialsFile == "" {
		credentialsFile = strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	}

	from := strings.TrimSpace(os.Getenv("TIMEXEED_GMAIL_FROM"))
	senderName := strings.TrimSpace(os.Getenv("TIMEXEED_GMAIL_SENDER_NAME"))
	subjectPrefix := strings.TrimSpace(os.Getenv("TIMEXEED_GMAIL_SUBJECT_PREFIX"))

	if credentialsFile == "" {
		return NewDisabledMailService("TIMEXEED_GMAIL_CREDENTIALS_FILE or GOOGLE_APPLICATION_CREDENTIALS is not set"), results.OK(
			nil,
			"NEW_GMAIL_MAIL_SERVICE_SKIPPED_CREDENTIALS_FILE_EMPTY",
			"メール送信設定が未設定のため、メール送信を無効化しました",
			nil,
		)
	}

	if from == "" {
		return NewDisabledMailService("TIMEXEED_GMAIL_FROM is not set"), results.OK(
			nil,
			"NEW_GMAIL_MAIL_SERVICE_SKIPPED_FROM_EMPTY",
			"メール送信元が未設定のため、メール送信を無効化しました",
			nil,
		)
	}

	if _, err := netmail.ParseAddress(from); err != nil {
		return NewDisabledMailService("TIMEXEED_GMAIL_FROM is invalid"), results.OK(
			nil,
			"NEW_GMAIL_MAIL_SERVICE_SKIPPED_INVALID_FROM",
			"メール送信元が不正なため、メール送信を無効化しました",
			err.Error(),
		)
	}

	credentialsJSON, readErr := os.ReadFile(credentialsFile)
	if readErr != nil {
		return NewDisabledMailService("failed to read gmail credentials file"), results.OK(
			nil,
			"NEW_GMAIL_MAIL_SERVICE_SKIPPED_READ_CREDENTIALS_FAILED",
			"メール送信用認証ファイルを読み込めないため、メール送信を無効化しました",
			readErr.Error(),
		)
	}

	jwtConfig, configErr := google.JWTConfigFromJSON(credentialsJSON, gmail.GmailSendScope)
	if configErr != nil {
		return NewDisabledMailService("failed to parse gmail credentials file"), results.OK(
			nil,
			"NEW_GMAIL_MAIL_SERVICE_SKIPPED_PARSE_CREDENTIALS_FAILED",
			"メール送信用認証ファイルを解析できないため、メール送信を無効化しました",
			configErr.Error(),
		)
	}

	// Domain-wide delegationで、実在するGoogle Workspaceユーザーを代理実行する。
	jwtConfig.Subject = from

	httpClient := jwtConfig.Client(ctx)

	gmailService, serviceErr := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if serviceErr != nil {
		return NewDisabledMailService("failed to create gmail service"), results.OK(
			nil,
			"NEW_GMAIL_MAIL_SERVICE_SKIPPED_CREATE_SERVICE_FAILED",
			"Gmail API Serviceを生成できないため、メール送信を無効化しました",
			serviceErr.Error(),
		)
	}

	return &gmailMailService{
			gmailService:  gmailService,
			from:          from,
			senderName:    senderName,
			subjectPrefix: subjectPrefix,
		}, results.OK(
			nil,
			"NEW_GMAIL_MAIL_SERVICE_SUCCESS",
			"メール送信Serviceを生成しました",
			nil,
		)
}

/*
 * お知らせメール送信
 */
func (service *gmailMailService) SendNotificationMail(
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
func (service *gmailMailService) SendMail(message MailMessage) results.Result {
	if service == nil || service.gmailService == nil {
		return results.InternalServerError(
			"SEND_MAIL_SERVICE_IS_NIL",
			"メール送信に失敗しました",
			nil,
		)
	}

	to := strings.TrimSpace(message.To)
	subject := strings.TrimSpace(message.Subject)
	body := strings.TrimSpace(message.Body)

	if to == "" {
		return results.BadRequest(
			"SEND_MAIL_TO_EMPTY",
			"メール送信先が空です",
			nil,
		)
	}

	if _, err := netmail.ParseAddress(to); err != nil {
		return results.BadRequest(
			"SEND_MAIL_TO_INVALID",
			"メール送信先が正しくありません",
			err.Error(),
		)
	}

	if subject == "" {
		return results.BadRequest(
			"SEND_MAIL_SUBJECT_EMPTY",
			"メール件名が空です",
			nil,
		)
	}

	if body == "" {
		return results.BadRequest(
			"SEND_MAIL_BODY_EMPTY",
			"メール本文が空です",
			nil,
		)
	}

	finalSubject := service.buildSubject(subject)
	rawMessage := service.buildRawMessage(to, finalSubject, body)

	gmailMessage := &gmail.Message{
		Raw: rawMessage,
	}

	if _, err := service.gmailService.Users.Messages.Send("me", gmailMessage).Do(); err != nil {
		return results.InternalServerError(
			"SEND_MAIL_FAILED",
			"メール送信に失敗しました",
			err.Error(),
		)
	}

	return results.OK(
		nil,
		"SEND_MAIL_SUCCESS",
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
func (service *gmailMailService) buildSubject(subject string) string {
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
 * Gmail API送信用Raw Message作成
 *
 * Gmail APIのmessages.sendでは、RFC 2822形式のメール本文をbase64urlエンコードして渡す。
 */
func (service *gmailMailService) buildRawMessage(to string, subject string, body string) string {
	fromHeader := service.from
	if service.senderName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("UTF-8", service.senderName), service.from)
	}

	encodedSubject := mime.QEncoding.Encode("UTF-8", sanitizeMailHeader(subject))
	encodedBody := base64.StdEncoding.EncodeToString([]byte(body))

	message := strings.Join([]string{
		fmt.Sprintf("From: %s", fromHeader),
		fmt.Sprintf("To: %s", sanitizeMailHeader(to)),
		fmt.Sprintf("Subject: %s", encodedSubject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: base64",
		"",
		encodedBody,
	}, "\r\n")

	return base64.URLEncoding.EncodeToString([]byte(message))
}

/*
 * メールヘッダーインジェクション対策
 */
func sanitizeMailHeader(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return strings.TrimSpace(value)
}
