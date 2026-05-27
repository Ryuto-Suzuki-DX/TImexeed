package mail

/*
 * 〇 メール送信メッセージ
 *
 * internal/mail 共通メール送信Serviceで使用する。
 *
 * 注意：
 * ・通知メール送信用の最小構成にしている
 * ・HTMLメールではなく、まずはプレーンテキストメールとして送信する
 * ・To は基本的に1件ずつ送る想定
 */
type MailMessage struct {
	To      string
	Subject string
	Body    string
}
