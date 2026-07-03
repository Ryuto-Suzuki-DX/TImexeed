package schedulers

import (
	"log"
	"time"

	"timexeed/backend/internal/modules/admin/services"
)

/*
 * お知らせ自動リマインドScheduler
 *
 * バックエンド起動後、次の分の開始時刻まで待機し、
 * 以降1分ごとに自動リマインド実行Serviceを呼び出す。
 *
 * 起動直後には実行しない。
 * 同じ1分内の再起動による重複通知を避けるため。
 */
func StartNotificationReminderScheduler(
	notificationReminderExecutionService services.NotificationReminderExecutionService,
) {
	go func() {
		waitUntilNextMinute()

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		executeNotificationReminders(notificationReminderExecutionService)

		for range ticker.C {
			executeNotificationReminders(notificationReminderExecutionService)
		}
	}()
}

/*
 * 次の分の開始時刻まで待機する
 */
func waitUntilNextMinute() {
	now := time.Now()
	nextMinute := now.Truncate(time.Minute).Add(time.Minute)
	time.Sleep(time.Until(nextMinute))
}

/*
 * 自動リマインド実行
 *
 * Scheduler内のエラーではバックエンドを停止させず、
 * エラーログだけを出力する。
 */
func executeNotificationReminders(
	notificationReminderExecutionService services.NotificationReminderExecutionService,
) {
	if err := notificationReminderExecutionService.ExecuteNotificationReminders(); err != nil {
		log.Printf(
			"[ERROR] notification reminder scheduler failed: %v",
			err,
		)
	}
}
