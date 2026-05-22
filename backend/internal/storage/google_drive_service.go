package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
 * ・Google Drive上にフォルダを作成する
 * ・Google Driveのフォルダ/ファイル権限を最新状態へ同期する
 *
 * 注意：
 * ・Timexeed本体にはファイルを永続保存しない
 * ・DBへ保存するのは DriveFileID / FileURL / OriginalFileName / StoredFileName / MimeType / SizeBytes などのメタ情報
 * ・親フォルダから継承された権限は、子フォルダ側のAPIでは削除できない
 * ・「対象ユーザー本人＋管理者全員だけ」にしたい場合、親フォルダ自体に余計な共有権限を付けない運用にする
 *
 * 認証方式：
 * ・OAuth 2.0 refresh token 方式
 * ・サービスアカウント方式は使わない
 *
 * 環境変数：
 * ・GOOGLE_OAUTH_CLIENT_ID
 * ・GOOGLE_OAUTH_CLIENT_SECRET
 * ・GOOGLE_OAUTH_REFRESH_TOKEN
 */
type GoogleDriveService interface {
	UploadFile(ctx context.Context, folderID string, storedFileName string, mimeType string, reader io.Reader) (GoogleDriveUploadedFile, error)
	DownloadFile(ctx context.Context, fileID string) (GoogleDriveDownloadedFile, error)
	GetFileMetadata(ctx context.Context, fileID string) (GoogleDriveFileMetadata, error)
	DeleteFile(ctx context.Context, fileID string) error
	ParseFolderID(folderURLOrID string) (string, error)

	CreateFolder(ctx context.Context, parentFolderID string, folderName string) (GoogleDriveFolderMetadata, error)
	GetFolderMetadata(ctx context.Context, folderID string) (GoogleDriveFolderMetadata, error)
	SyncPermissions(ctx context.Context, fileOrFolderID string, permissions []GoogleDrivePermissionSetting, removeUnexpectedDirectPermissions bool) error
}

type googleDriveService struct {
	driveService *drive.Service
}

/*
 * OAuth認証情報
 */
type googleDriveOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
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
 * Google Driveフォルダメタ情報
 */
type GoogleDriveFolderMetadata struct {
	DriveFolderID string
	FolderName    string
	FolderURL     string
	MimeType      string
}

/*
 * Google Drive権限設定
 *
 * EmailAddress：
 * ・権限を付与するGoogleアカウントのメールアドレス
 *
 * Role：
 * ・reader / writer など
 */
type GoogleDrivePermissionSetting struct {
	EmailAddress string
	Role         string
}

/*
 * 環境変数からGoogle Drive Serviceを生成する。
 */
func NewGoogleDriveServiceFromEnv(ctx context.Context) (GoogleDriveService, error) {
	oauthConfig, err := loadGoogleDriveOAuthConfigFromEnv()
	if err != nil {
		return nil, err
	}

	return NewGoogleDriveServiceWithOAuth(ctx, oauthConfig)
}

/*
 * OAuth refresh tokenからGoogle Drive Serviceを生成する。
 */
func NewGoogleDriveServiceWithOAuth(ctx context.Context, oauthConfig googleDriveOAuthConfig) (GoogleDriveService, error) {
	clientID := strings.TrimSpace(oauthConfig.ClientID)
	clientSecret := strings.TrimSpace(oauthConfig.ClientSecret)
	refreshToken := strings.TrimSpace(oauthConfig.RefreshToken)

	if clientID == "" {
		return nil, errors.New("GOOGLE_OAUTH_CLIENT_ID is empty")
	}

	if clientSecret == "" {
		return nil, errors.New("GOOGLE_OAUTH_CLIENT_SECRET is empty")
	}

	if refreshToken == "" {
		return nil, errors.New("GOOGLE_OAUTH_REFRESH_TOKEN is empty")
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			drive.DriveScope,
		},
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	httpClient := config.Client(ctx, token)

	driveService, err := drive.NewService(
		ctx,
		option.WithHTTPClient(httpClient),
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
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		return GoogleDriveUploadedFile{}, fmt.Errorf("failed to upload file to google drive: %w", err)
	}

	return GoogleDriveUploadedFile{
		DriveFileID: createdFile.Id,
		FileName:    createdFile.Name,
		FileURL:     buildGoogleDriveWebURL(createdFile.Id, createdFile.WebViewLink),
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
		SupportsAllDrives(true).
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
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		return GoogleDriveFileMetadata{}, fmt.Errorf("failed to get google drive file metadata: %w", err)
	}

	return GoogleDriveFileMetadata{
		DriveFileID: file.Id,
		FileName:    file.Name,
		FileURL:     buildGoogleDriveWebURL(file.Id, file.WebViewLink),
		MimeType:    file.MimeType,
		SizeBytes:   file.Size,
	}, nil
}

/*
 * ファイル削除
 */
func (service *googleDriveService) DeleteFile(ctx context.Context, fileID string) error {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return errors.New("google drive file id is empty")
	}

	if err := service.driveService.Files.Delete(fileID).SupportsAllDrives(true).Context(ctx).Do(); err != nil {
		return fmt.Errorf("failed to delete google drive file: %w", err)
	}

	return nil
}

/*
 * Google Driveフォルダ作成
 */
func (service *googleDriveService) CreateFolder(ctx context.Context, parentFolderID string, folderName string) (GoogleDriveFolderMetadata, error) {
	parentFolderID = strings.TrimSpace(parentFolderID)
	folderName = strings.TrimSpace(folderName)

	if parentFolderID == "" {
		return GoogleDriveFolderMetadata{}, errors.New("parent google drive folder id is empty")
	}

	if folderName == "" {
		return GoogleDriveFolderMetadata{}, errors.New("google drive folder name is empty")
	}

	folder := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentFolderID},
	}

	createdFolder, err := service.driveService.Files.
		Create(folder).
		Fields("id", "name", "mimeType", "webViewLink").
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		return GoogleDriveFolderMetadata{}, fmt.Errorf("failed to create google drive folder: %w", err)
	}

	return GoogleDriveFolderMetadata{
		DriveFolderID: createdFolder.Id,
		FolderName:    createdFolder.Name,
		FolderURL:     buildGoogleDriveFolderURL(createdFolder.Id, createdFolder.WebViewLink),
		MimeType:      createdFolder.MimeType,
	}, nil
}

/*
 * Google Driveフォルダメタ情報取得
 */
func (service *googleDriveService) GetFolderMetadata(ctx context.Context, folderID string) (GoogleDriveFolderMetadata, error) {
	folderID = strings.TrimSpace(folderID)
	if folderID == "" {
		return GoogleDriveFolderMetadata{}, errors.New("google drive folder id is empty")
	}

	folder, err := service.driveService.Files.
		Get(folderID).
		Fields("id", "name", "mimeType", "webViewLink").
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		return GoogleDriveFolderMetadata{}, fmt.Errorf("failed to get google drive folder metadata: %w", err)
	}

	if folder.MimeType != "application/vnd.google-apps.folder" {
		return GoogleDriveFolderMetadata{}, fmt.Errorf("google drive item is not folder: %s", folderID)
	}

	return GoogleDriveFolderMetadata{
		DriveFolderID: folder.Id,
		FolderName:    folder.Name,
		FolderURL:     buildGoogleDriveFolderURL(folder.Id, folder.WebViewLink),
		MimeType:      folder.MimeType,
	}, nil
}

/*
 * Google Drive権限同期
 *
 * 処理内容：
 * ・指定メールアドレスに必要なroleを付与する
 * ・既に権限がある場合、roleが違えば更新する
 * ・removeUnexpectedDirectPermissions=true の場合、許可リスト外の直接共有権限を削除する
 *
 * 注意：
 * ・owner権限は削除/変更しない
 * ・inherited=true の権限は削除/変更しない
 * ・親フォルダから継承された権限はここでは消せない
 */
func (service *googleDriveService) SyncPermissions(
	ctx context.Context,
	fileOrFolderID string,
	permissions []GoogleDrivePermissionSetting,
	removeUnexpectedDirectPermissions bool,
) error {
	fileOrFolderID = strings.TrimSpace(fileOrFolderID)
	if fileOrFolderID == "" {
		return errors.New("google drive file or folder id is empty")
	}

	normalizedPermissions := normalizePermissionSettings(permissions)
	if len(normalizedPermissions) == 0 {
		return errors.New("google drive permissions are empty")
	}

	existingPermissions, err := service.driveService.Permissions.
		List(fileOrFolderID).
		Fields("permissions(id,type,emailAddress,role,deleted,permissionDetails)").
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed to list google drive permissions: %w", err)
	}

	existingUserPermissionsByEmail := make(map[string]*drive.Permission)
	allowedEmails := make(map[string]bool)

	for _, permission := range normalizedPermissions {
		allowedEmails[permission.EmailAddress] = true
	}

	for _, existingPermission := range existingPermissions.Permissions {
		email := strings.ToLower(strings.TrimSpace(existingPermission.EmailAddress))
		if email != "" && existingPermission.Type == "user" && !existingPermission.Deleted {
			existingUserPermissionsByEmail[email] = existingPermission
		}
	}

	for _, permission := range normalizedPermissions {
		existingPermission, exists := existingUserPermissionsByEmail[permission.EmailAddress]
		if !exists {
			createPermission := &drive.Permission{
				Type:         "user",
				Role:         permission.Role,
				EmailAddress: permission.EmailAddress,
			}

			if _, err := service.driveService.Permissions.
				Create(fileOrFolderID, createPermission).
				SendNotificationEmail(true).
				SupportsAllDrives(true).
				Context(ctx).
				Do(); err != nil {
				return fmt.Errorf("failed to create google drive permission for %s: %w", permission.EmailAddress, err)
			}

			continue
		}

		if existingPermission.Role != permission.Role && existingPermission.Role != "owner" && !isInheritedGoogleDrivePermission(existingPermission) {
			updatePermission := &drive.Permission{
				Role: permission.Role,
			}

			if _, err := service.driveService.Permissions.
				Update(fileOrFolderID, existingPermission.Id, updatePermission).
				SupportsAllDrives(true).
				Context(ctx).
				Do(); err != nil {
				return fmt.Errorf("failed to update google drive permission for %s: %w", permission.EmailAddress, err)
			}
		}
	}

	if !removeUnexpectedDirectPermissions {
		return nil
	}

	for _, existingPermission := range existingPermissions.Permissions {
		if shouldKeepGoogleDrivePermission(existingPermission, allowedEmails) {
			continue
		}

		if err := service.driveService.Permissions.
			Delete(fileOrFolderID, existingPermission.Id).
			SupportsAllDrives(true).
			Context(ctx).
			Do(); err != nil {
			return fmt.Errorf("failed to delete unexpected google drive permission: %w", err)
		}
	}

	return nil
}

/*
 * Google DriveフォルダURL、またはフォルダIDから folderId を取り出す。
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
 * 個人情報Driveフォルダ名を作る。
 *
 * 例：
 * user_12_山田太郎
 */
func BuildGoogleDriveUserFolderName(prefix string, userID uint, userName string) string {
	prefix = sanitizeFileNamePart(prefix)
	userName = sanitizeFileNamePart(userName)

	return fmt.Sprintf("%s_%d_%s", prefix, userID, userName)
}

/*
 * multipart.FileHeader から MIMEタイプを取り出す補助。
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
 * 環境変数からOAuth認証情報を読み込む。
 */
func loadGoogleDriveOAuthConfigFromEnv() (googleDriveOAuthConfig, error) {
	clientID := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"))
	refreshToken := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_REFRESH_TOKEN"))

	if clientID == "" {
		return googleDriveOAuthConfig{}, errors.New("GOOGLE_OAUTH_CLIENT_ID is required")
	}

	if clientSecret == "" {
		return googleDriveOAuthConfig{}, errors.New("GOOGLE_OAUTH_CLIENT_SECRET is required")
	}

	if refreshToken == "" {
		return googleDriveOAuthConfig{}, errors.New("GOOGLE_OAUTH_REFRESH_TOKEN is required")
	}

	return googleDriveOAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
	}, nil
}

func normalizePermissionSettings(permissions []GoogleDrivePermissionSetting) []GoogleDrivePermissionSetting {
	permissionByEmail := make(map[string]GoogleDrivePermissionSetting)

	for _, permission := range permissions {
		email := strings.ToLower(strings.TrimSpace(permission.EmailAddress))
		role := strings.TrimSpace(permission.Role)

		if email == "" {
			continue
		}

		if role == "" {
			role = "reader"
		}

		permissionByEmail[email] = GoogleDrivePermissionSetting{
			EmailAddress: email,
			Role:         role,
		}
	}

	emails := make([]string, 0, len(permissionByEmail))
	for email := range permissionByEmail {
		emails = append(emails, email)
	}
	sort.Strings(emails)

	normalizedPermissions := make([]GoogleDrivePermissionSetting, 0, len(emails))
	for _, email := range emails {
		normalizedPermissions = append(normalizedPermissions, permissionByEmail[email])
	}

	return normalizedPermissions
}

func shouldKeepGoogleDrivePermission(permission *drive.Permission, allowedEmails map[string]bool) bool {
	if permission == nil {
		return true
	}

	if permission.Id == "" {
		return true
	}

	if permission.Role == "owner" {
		return true
	}

	if isInheritedGoogleDrivePermission(permission) {
		return true
	}

	if permission.Type == "user" {
		email := strings.ToLower(strings.TrimSpace(permission.EmailAddress))
		return allowedEmails[email]
	}

	return false
}

func isInheritedGoogleDrivePermission(permission *drive.Permission) bool {
	if permission == nil {
		return false
	}

	for _, detail := range permission.PermissionDetails {
		if detail.Inherited {
			return true
		}
	}

	return false
}

func buildGoogleDriveWebURL(fileID string, webViewLink string) string {
	webViewLink = strings.TrimSpace(webViewLink)
	if webViewLink != "" {
		return webViewLink
	}

	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return ""
	}

	return fmt.Sprintf("https://drive.google.com/file/d/%s/view", fileID)
}

func buildGoogleDriveFolderURL(folderID string, webViewLink string) string {
	webViewLink = strings.TrimSpace(webViewLink)
	if webViewLink != "" {
		return webViewLink
	}

	folderID = strings.TrimSpace(folderID)
	if folderID == "" {
		return ""
	}

	return fmt.Sprintf("https://drive.google.com/drive/folders/%s", folderID)
}
