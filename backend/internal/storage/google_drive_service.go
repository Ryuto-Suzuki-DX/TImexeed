package storage

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

/*
 * Google Drive Service
 *
 * 役割：
 * ・Google Driveへファイルをアップロードする
 * ・Google Driveからファイルを取得する
 * ・Google DriveのフォルダURLから folderId を取り出す
 *
 * 注意：
 * ・Timexeed本体にはファイルを永続保存しない
 * ・Controllerで受け取ったmultipart.Fileを、このServiceへ渡してDriveへ送る
 * ・DBへ保存するのは DriveFileID / FileURL / OriginalFileName / StoredFileName / MimeType / SizeBytes
 *
 * 環境変数：
 * 1. GOOGLE_SERVICE_ACCOUNT_JSON
 *    - サービスアカウントJSONの中身をそのまま入れる
 *    - または base64 エンコードしたJSONを入れる
 *
 * 2. GOOGLE_APPLICATION_CREDENTIALS
 *    - サービスアカウントJSONファイルのパスを入れる
 *
 * どちらかが必要。
 */
type GoogleDriveService interface {
	UploadFile(ctx context.Context, folderID string, storedFileName string, mimeType string, reader io.Reader) (GoogleDriveUploadedFile, error)
	DownloadFile(ctx context.Context, fileID string) (GoogleDriveDownloadedFile, error)
	GetFileMetadata(ctx context.Context, fileID string) (GoogleDriveFileMetadata, error)
	DeleteFile(ctx context.Context, fileID string) error
	ParseFolderID(folderURLOrID string) (string, error)
}

type googleDriveService struct {
	driveService *drive.Service
}

/*
 * Google Driveアップロード結果
 */
type GoogleDriveUploadedFile struct {
	DriveFileID string
	FileName    string
	FileURL     string
	MimeType    string
	SizeBytes   int64
}

/*
 * Google Driveダウンロード結果
 */
type GoogleDriveDownloadedFile struct {
	Body      io.ReadCloser
	FileName  string
	MimeType  string
	SizeBytes int64
}

/*
 * Google Driveファイルメタ情報
 */
type GoogleDriveFileMetadata struct {
	DriveFileID string
	FileName    string
	FileURL     string
	MimeType    string
	SizeBytes   int64
}

/*
 * 環境変数からGoogle Drive Serviceを生成する。
 */
func NewGoogleDriveServiceFromEnv(ctx context.Context) (GoogleDriveService, error) {
	credentialJSON, err := loadGoogleCredentialJSONFromEnv()
	if err != nil {
		return nil, err
	}

	return NewGoogleDriveService(ctx, credentialJSON)
}

/*
 * サービスアカウントJSONからGoogle Drive Serviceを生成する。
 */
func NewGoogleDriveService(ctx context.Context, credentialJSON []byte) (GoogleDriveService, error) {
	if len(credentialJSON) == 0 {
		return nil, errors.New("google credential json is empty")
	}

	driveService, err := drive.NewService(
		ctx,
		option.WithCredentialsJSON(credentialJSON),
		option.WithScopes(drive.DriveScope),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create google drive service: %w", err)
	}

	return &googleDriveService{
		driveService: driveService,
	}, nil
}

/*
 * ファイルアップロード
 *
 * folderID：
 * ・Google DriveのフォルダID
 *
 * storedFileName：
 * ・Driveに保存する実ファイル名
 * ・ユーザーがアップロードした元ファイル名ではなく、Timexeed側で採番した名前を渡す
 */
func (service *googleDriveService) UploadFile(
	ctx context.Context,
	folderID string,
	storedFileName string,
	mimeType string,
	reader io.Reader,
) (GoogleDriveUploadedFile, error) {
	folderID = strings.TrimSpace(folderID)
	storedFileName = strings.TrimSpace(storedFileName)
	mimeType = strings.TrimSpace(mimeType)

	if folderID == "" {
		return GoogleDriveUploadedFile{}, errors.New("google drive folder id is empty")
	}

	if storedFileName == "" {
		return GoogleDriveUploadedFile{}, errors.New("stored file name is empty")
	}

	if reader == nil {
		return GoogleDriveUploadedFile{}, errors.New("file reader is nil")
	}

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	file := &drive.File{
		Name:     storedFileName,
		MimeType: mimeType,
		Parents:  []string{folderID},
	}

	createdFile, err := service.driveService.Files.
		Create(file).
		Media(reader).
		Fields("id", "name", "mimeType", "size", "webViewLink").
		Context(ctx).
		Do()
	if err != nil {
		return GoogleDriveUploadedFile{}, fmt.Errorf("failed to upload file to google drive: %w", err)
	}

	return GoogleDriveUploadedFile{
		DriveFileID: createdFile.Id,
		FileName:    createdFile.Name,
		FileURL:     createdFile.WebViewLink,
		MimeType:    createdFile.MimeType,
		SizeBytes:   createdFile.Size,
	}, nil
}

/*
 * ファイルダウンロード
 *
 * Controller側では、このBodyを必ずCloseすること。
 */
func (service *googleDriveService) DownloadFile(ctx context.Context, fileID string) (GoogleDriveDownloadedFile, error) {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return GoogleDriveDownloadedFile{}, errors.New("google drive file id is empty")
	}

	metadata, err := service.GetFileMetadata(ctx, fileID)
	if err != nil {
		return GoogleDriveDownloadedFile{}, err
	}

	response, err := service.driveService.Files.
		Get(fileID).
		Context(ctx).
		Download()
	if err != nil {
		return GoogleDriveDownloadedFile{}, fmt.Errorf("failed to download file from google drive: %w", err)
	}

	return GoogleDriveDownloadedFile{
		Body:      response.Body,
		FileName:  metadata.FileName,
		MimeType:  metadata.MimeType,
		SizeBytes: metadata.SizeBytes,
	}, nil
}

/*
 * ファイルメタ情報取得
 */
func (service *googleDriveService) GetFileMetadata(ctx context.Context, fileID string) (GoogleDriveFileMetadata, error) {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return GoogleDriveFileMetadata{}, errors.New("google drive file id is empty")
	}

	file, err := service.driveService.Files.
		Get(fileID).
		Fields("id", "name", "mimeType", "size", "webViewLink").
		Context(ctx).
		Do()
	if err != nil {
		return GoogleDriveFileMetadata{}, fmt.Errorf("failed to get google drive file metadata: %w", err)
	}

	return GoogleDriveFileMetadata{
		DriveFileID: file.Id,
		FileName:    file.Name,
		FileURL:     file.WebViewLink,
		MimeType:    file.MimeType,
		SizeBytes:   file.Size,
	}, nil
}

/*
 * ファイル削除
 *
 * 経費更新で領収書を差し替える場合などに使える。
 * ただし履歴を残したい運用なら、すぐ削除せずDrive上に残す判断も可能。
 */
func (service *googleDriveService) DeleteFile(ctx context.Context, fileID string) error {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return errors.New("google drive file id is empty")
	}

	if err := service.driveService.Files.Delete(fileID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("failed to delete google drive file: %w", err)
	}

	return nil
}

/*
 * Google DriveフォルダURL、またはフォルダIDから folderId を取り出す。
 *
 * 対応例：
 * ・https://drive.google.com/drive/folders/{folderId}
 * ・https://drive.google.com/open?id={folderId}
 * ・{folderId}
 */
func (service *googleDriveService) ParseFolderID(folderURLOrID string) (string, error) {
	return ParseGoogleDriveFolderID(folderURLOrID)
}

/*
 * package外からも使えるように関数としても定義しておく。
 */
func ParseGoogleDriveFolderID(folderURLOrID string) (string, error) {
	value := strings.TrimSpace(folderURLOrID)
	if value == "" {
		return "", errors.New("google drive folder url or id is empty")
	}

	foldersPattern := regexp.MustCompile(`/folders/([a-zA-Z0-9_-]+)`)
	if matches := foldersPattern.FindStringSubmatch(value); len(matches) == 2 {
		return matches[1], nil
	}

	idQueryPattern := regexp.MustCompile(`[?&]id=([a-zA-Z0-9_-]+)`)
	if matches := idQueryPattern.FindStringSubmatch(value); len(matches) == 2 {
		return matches[1], nil
	}

	if regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(value) {
		return value, nil
	}

	return "", fmt.Errorf("invalid google drive folder url or id: %s", value)
}

/*
 * Google Driveへ保存するファイル名を作る。
 *
 * 例：
 * expense_user_12_20260516_153012_receipt.jpg
 */
func BuildGoogleDriveStoredFileName(prefix string, userID uint, timestamp string, originalFileName string) string {
	prefix = sanitizeFileNamePart(prefix)
	timestamp = sanitizeFileNamePart(timestamp)

	extension := extractFileExtension(originalFileName)
	if extension != "" {
		return fmt.Sprintf("%s_user_%d_%s%s", prefix, userID, timestamp, extension)
	}

	return fmt.Sprintf("%s_user_%d_%s", prefix, userID, timestamp)
}

/*
 * multipart.FileHeader から MIMEタイプとサイズを取り出す補助。
 */
func GetMultipartFileInfo(fileHeader interface {
	Get(string) string
}) string {
	if fileHeader == nil {
		return ""
	}

	return strings.TrimSpace(fileHeader.Get("Content-Type"))
}

/*
 * ファイル名に使いにくい文字を置換する。
 */
func sanitizeFileNamePart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "file"
	}

	replacer := strings.NewReplacer(
		" ", "_",
		"　", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)

	return replacer.Replace(value)
}

/*
 * 拡張子だけ取り出す。
 */
func extractFileExtension(fileName string) string {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return ""
	}

	lastDotIndex := strings.LastIndex(fileName, ".")
	if lastDotIndex < 0 || lastDotIndex == len(fileName)-1 {
		return ""
	}

	extension := fileName[lastDotIndex:]
	if len(extension) > 20 {
		return ""
	}

	return sanitizeFileNamePart(extension)
}

/*
 * 環境変数からサービスアカウントJSONを読み込む。
 */
func loadGoogleCredentialJSONFromEnv() ([]byte, error) {
	rawJSON := strings.TrimSpace(os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON"))
	if rawJSON != "" {
		if strings.HasPrefix(rawJSON, "{") {
			return []byte(rawJSON), nil
		}

		decodedJSON, err := base64.StdEncoding.DecodeString(rawJSON)
		if err == nil && len(decodedJSON) > 0 {
			return decodedJSON, nil
		}

		return nil, errors.New("GOOGLE_SERVICE_ACCOUNT_JSON is not valid json or base64 json")
	}

	credentialPath := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if credentialPath != "" {
		credentialJSON, err := os.ReadFile(credentialPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read GOOGLE_APPLICATION_CREDENTIALS file: %w", err)
		}

		return credentialJSON, nil
	}

	return nil, errors.New("GOOGLE_SERVICE_ACCOUNT_JSON or GOOGLE_APPLICATION_CREDENTIALS is required")
}
