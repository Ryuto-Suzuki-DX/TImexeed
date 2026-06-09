package jobs

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"timexeed/backend/internal/storage"
	"timexeed/backend/internal/systemlogs"

	"gorm.io/gorm"
)

const (
	defaultApiOperationLogUploadTime     = "02:00"
	defaultApiOperationLogRetentionMonth = 6
)

/*
 * API操作ログ日次アップロードJobを開始する。
 *
 * 環境変数：
 * ・TIMEXEED_API_LOG_UPLOAD_TIME
 *   例：02:00
 *   未設定の場合は 02:00 に実行する。
 *
 * 処理内容：
 * ・毎日設定時刻に前日分の api_operation_logs をCSV化する
 * ・SYSTEM_LOG_DRIVE_ROOT のGoogle Driveフォルダへアップロードする
 * ・同名CSVがある場合は更新する
 * ・Google Drive上のCSVは半年分だけ残す
 * ・アップロード成功後、DB側は当日より前のログを削除する
 */
func StartApiOperationLogDailyUploadJob(ctx context.Context, db *gorm.DB, driveService storage.GoogleDriveService) {
	go func() {
		location, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			location = time.FixedZone("JST", 9*60*60)
		}

		for {
			nextRun := calculateNextApiOperationLogUploadTime(time.Now().In(location), location)
			waitDuration := time.Until(nextRun)

			select {
			case <-time.After(waitDuration):
				runApiOperationLogDailyUpload(ctx, db, driveService, location)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func runApiOperationLogDailyUpload(ctx context.Context, db *gorm.DB, driveService storage.GoogleDriveService, location *time.Location) {
	if db == nil {
		log.Println("api operation log daily upload skipped: db is nil")
		return
	}

	if driveService == nil {
		log.Println("api operation log daily upload skipped: google drive service is nil")
		return
	}

	now := time.Now().In(location)
	targetDate := now.AddDate(0, 0, -1)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)

	exportService := systemlogs.NewApiOperationLogExportService(db, driveService)

	if err := exportService.UploadDailyCsv(ctx, targetDate); err != nil {
		log.Printf("failed to upload api operation log csv: %v\n", err)
		return
	}

	if err := exportService.DeleteOldDriveCsvFiles(ctx, now, defaultApiOperationLogRetentionMonth); err != nil {
		log.Printf("failed to delete old api operation log csv files: %v\n", err)
		return
	}

	if err := exportService.DeleteUploadedDbLogsBefore(ctx, todayStart); err != nil {
		log.Printf("failed to delete uploaded api operation logs from db: %v\n", err)
		return
	}

	log.Printf("api operation log daily upload completed: targetDate=%s\n", targetDate.Format("2006-01-02"))
}

func calculateNextApiOperationLogUploadTime(now time.Time, location *time.Location) time.Time {
	hour, minute := loadApiOperationLogUploadHourAndMinute()

	nextRun := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, location)
	if !nextRun.After(now) {
		nextRun = nextRun.AddDate(0, 0, 1)
	}

	return nextRun
}

func loadApiOperationLogUploadHourAndMinute() (int, int) {
	value := strings.TrimSpace(os.Getenv("TIMEXEED_API_LOG_UPLOAD_TIME"))
	if value == "" {
		value = defaultApiOperationLogUploadTime
	}

	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 2, 0
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 2, 0
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 2, 0
	}

	return hour, minute
}
