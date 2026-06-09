package systemlogs

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/storage"

	"gorm.io/gorm"
)

const (
	SystemLogDriveRootLinkType = "SYSTEM_LOG_DRIVE_ROOT"
	ApiOperationLogCsvMimeType = "text/csv"
	ApiOperationLogCsvPrefix   = "timexeed-api-operation-log"
)

/*
 * API操作ログCSV出力Service
 *
 * 役割：
 * ・api_operation_logs から対象日1日分のログを取得する
 * ・CSVを生成する
 * ・external_storage_links の SYSTEM_LOG_DRIVE_ROOT を参照する
 * ・Google DriveへCSVをアップロードする
 * ・同名CSVが既にある場合は更新する
 * ・Google Drive上の半年超過ログCSVを削除する
 */
type ApiOperationLogExportService interface {
	UploadDailyCsv(ctx context.Context, targetDate time.Time) error
	DeleteOldDriveCsvFiles(ctx context.Context, now time.Time, retentionMonths int) error
	DeleteUploadedDbLogsBefore(ctx context.Context, cutoff time.Time) error
}

type apiOperationLogExportService struct {
	db           *gorm.DB
	driveService storage.GoogleDriveService
	location     *time.Location
}

func NewApiOperationLogExportService(db *gorm.DB, driveService storage.GoogleDriveService) ApiOperationLogExportService {
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		location = time.FixedZone("JST", 9*60*60)
	}

	return &apiOperationLogExportService{
		db:           db,
		driveService: driveService,
		location:     location,
	}
}

/*
 * 対象日1日分のAPI操作ログをCSV化してGoogle Driveへアップロードする。
 */
func (service *apiOperationLogExportService) UploadDailyCsv(ctx context.Context, targetDate time.Time) error {
	if service.db == nil {
		return fmt.Errorf("db is nil")
	}

	if service.driveService == nil {
		return fmt.Errorf("google drive service is nil")
	}

	dayStart, nextDayStart := service.buildDateRange(targetDate)

	logs := make([]models.ApiOperationLog, 0)
	if err := service.db.WithContext(ctx).
		Where("started_at >= ? AND started_at < ?", dayStart, nextDayStart).
		Order("started_at ASC, id ASC").
		Find(&logs).Error; err != nil {
		return fmt.Errorf("failed to find api operation logs: %w", err)
	}

	csvBytes, err := service.buildCsv(logs)
	if err != nil {
		return err
	}

	folderID, err := service.findSystemLogDriveFolderID(ctx)
	if err != nil {
		return err
	}

	fileName := service.buildCsvFileName(dayStart)

	existingFile, exists, err := service.driveService.FindFileByNameInFolder(ctx, folderID, fileName)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(csvBytes)
	if exists && existingFile != nil {
		_, err = service.driveService.UpdateFile(ctx, existingFile.DriveFileID, ApiOperationLogCsvMimeType, reader)
		if err != nil {
			return err
		}

		return nil
	}

	_, err = service.driveService.UploadFile(ctx, folderID, fileName, ApiOperationLogCsvMimeType, reader)
	if err != nil {
		return err
	}

	return nil
}

/*
 * Google Drive上の保持期間超過CSVを削除する。
 */
func (service *apiOperationLogExportService) DeleteOldDriveCsvFiles(ctx context.Context, now time.Time, retentionMonths int) error {
	if service.driveService == nil {
		return fmt.Errorf("google drive service is nil")
	}

	if retentionMonths <= 0 {
		retentionMonths = 6
	}

	folderID, err := service.findSystemLogDriveFolderID(ctx)
	if err != nil {
		return err
	}

	files, err := service.driveService.ListFilesInFolder(ctx, folderID)
	if err != nil {
		return err
	}

	cutoffDate := now.In(service.location).AddDate(0, -retentionMonths, 0)

	for _, file := range files {
		logDate, ok := parseApiOperationLogCsvDate(file.FileName, service.location)
		if !ok {
			continue
		}

		if logDate.Before(beginningOfDay(cutoffDate, service.location)) {
			if err := service.driveService.DeleteFile(ctx, file.DriveFileID); err != nil {
				return err
			}
		}
	}

	return nil
}

/*
 * アップロード済みのDBログを削除する。
 *
 * 毎日アップロード成功後、当日より前のログを削除する想定。
 * これによりアプリ内部には基本的に当日分だけ残す。
 */
func (service *apiOperationLogExportService) DeleteUploadedDbLogsBefore(ctx context.Context, cutoff time.Time) error {
	if service.db == nil {
		return fmt.Errorf("db is nil")
	}

	if cutoff.IsZero() {
		return fmt.Errorf("cutoff is zero")
	}

	if err := service.db.WithContext(ctx).
		Where("started_at < ?", cutoff).
		Delete(&models.ApiOperationLog{}).Error; err != nil {
		return fmt.Errorf("failed to delete uploaded api operation logs: %w", err)
	}

	return nil
}

func (service *apiOperationLogExportService) findSystemLogDriveFolderID(ctx context.Context) (string, error) {
	var externalStorageLink models.ExternalStorageLink
	if err := service.db.WithContext(ctx).
		Where("link_type = ? AND is_deleted = ?", SystemLogDriveRootLinkType, false).
		First(&externalStorageLink).Error; err != nil {
		return "", fmt.Errorf("failed to find system log drive root external storage link: %w", err)
	}

	folderID, err := service.driveService.ParseFolderID(externalStorageLink.URL)
	if err != nil {
		return "", err
	}

	return folderID, nil
}

func (service *apiOperationLogExportService) buildDateRange(targetDate time.Time) (time.Time, time.Time) {
	dayStart := beginningOfDay(targetDate, service.location)
	return dayStart, dayStart.AddDate(0, 0, 1)
}

func (service *apiOperationLogExportService) buildCsvFileName(targetDate time.Time) string {
	return fmt.Sprintf("%s-%s.csv", ApiOperationLogCsvPrefix, targetDate.In(service.location).Format("2006-01-02"))
}

func (service *apiOperationLogExportService) buildCsv(logs []models.ApiOperationLog) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)

	/*
	 * Excelで文字化けしにくいようにUTF-8 BOMを付ける。
	 */
	buffer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buffer)

	headers := []string{
		"ID",
		"ユーザーID",
		"メールアドレス",
		"権限",
		"HTTPメソッド",
		"APIパス",
		"ステータスコード",
		"IPアドレス",
		"ユーザーエージェント",
		"処理時間ms",
		"エラー",
		"開始日時",
		"終了日時",
		"作成日時",
	}

	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("failed to write api operation log csv header: %w", err)
	}

	for _, log := range logs {
		record := []string{
			strconv.FormatUint(uint64(log.ID), 10),
			uintPointerToString(log.UserID),
			stringPointerToString(log.Email),
			stringPointerToString(log.Role),
			log.Method,
			log.Path,
			strconv.Itoa(log.StatusCode),
			log.ClientIP,
			log.UserAgent,
			strconv.FormatInt(log.DurationMs, 10),
			stringPointerToString(log.ErrorMessage),
			formatTime(log.StartedAt, service.location),
			formatTime(log.FinishedAt, service.location),
			formatTime(log.CreatedAt, service.location),
		}

		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write api operation log csv record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush api operation log csv: %w", err)
	}

	return buffer.Bytes(), nil
}

func parseApiOperationLogCsvDate(fileName string, location *time.Location) (time.Time, bool) {
	fileName = strings.TrimSpace(fileName)
	pattern := regexp.MustCompile(`^timexeed-api-operation-log-(\d{4}-\d{2}-\d{2})\.csv$`)
	matches := pattern.FindStringSubmatch(fileName)
	if len(matches) != 2 {
		return time.Time{}, false
	}

	parsedDate, err := time.ParseInLocation("2006-01-02", matches[1], location)
	if err != nil {
		return time.Time{}, false
	}

	return parsedDate, true
}

func beginningOfDay(value time.Time, location *time.Location) time.Time {
	localTime := value.In(location)
	return time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, location)
}

func uintPointerToString(value *uint) string {
	if value == nil {
		return ""
	}

	return strconv.FormatUint(uint64(*value), 10)
}

func stringPointerToString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func formatTime(value time.Time, location *time.Location) string {
	if value.IsZero() {
		return ""
	}

	return value.In(location).Format("2006-01-02 15:04:05")
}
