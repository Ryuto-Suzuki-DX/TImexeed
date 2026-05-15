package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/storage"
	"timexeed/backend/internal/utils"
)

/*
 * 管理者用経費Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type ExpenseService interface {
	SearchExpenses(req types.SearchExpensesRequest) results.Result
	GetExpenseDetail(req types.ExpenseDetailRequest) results.Result
	CreateExpense(ctx context.Context, req types.CreateExpenseRequest) results.Result
	UpdateExpense(ctx context.Context, req types.UpdateExpenseRequest) results.Result
	DeleteExpense(req types.DeleteExpenseRequest) results.Result
	DownloadExpenseReceipt(ctx context.Context, req types.ViewExpenseReceiptRequest) (types.ExpenseReceiptFileResponse, results.Result)
}

/*
 * 管理者用経費Service
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや更新用Modelを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB処理を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type expenseService struct {
	expenseBuilder                builders.ExpenseBuilder
	expenseRepository             repositories.ExpenseRepository
	externalStorageLinkRepository repositories.ExternalStorageLinkRepository
	googleDriveService            storage.GoogleDriveService
}

/*
 * ExpenseService生成
 */
func NewExpenseService(
	expenseBuilder builders.ExpenseBuilder,
	expenseRepository repositories.ExpenseRepository,
	externalStorageLinkRepository repositories.ExternalStorageLinkRepository,
	googleDriveService storage.GoogleDriveService,
) ExpenseService {
	return &expenseService{
		expenseBuilder:                expenseBuilder,
		expenseRepository:             expenseRepository,
		externalStorageLinkRepository: externalStorageLinkRepository,
		googleDriveService:            googleDriveService,
	}
}

/*
 * models.Expenseをフロント返却用ExpenseResponseへ変換する
 */
func toExpenseResponse(expense models.Expense) types.ExpenseResponse {
	return types.ExpenseResponse{
		ID:       expense.ID,
		UserID:   expense.UserID,
		UserName: expense.User.Name,
		Email:    expense.User.Email,

		TargetMonth: expense.TargetMonth.Format("2006-01"),
		ExpenseDate: expense.ExpenseDate.Format("2006-01-02"),

		Amount:      expense.Amount,
		Description: expense.Description,
		Memo:        expense.Memo,

		HasReceiptFile:   expense.DriveFileID != nil && *expense.DriveFileID != "",
		OriginalFileName: expense.OriginalFileName,
		StoredFileName:   expense.StoredFileName,
		MimeType:         expense.MimeType,
		SizeBytes:        expense.SizeBytes,

		CreatedAt: expense.CreatedAt,
		UpdatedAt: expense.UpdatedAt,
		DeletedAt: expense.DeletedAt,
	}
}

/*
 * models.Expenseをフロント返却用ExpenseListItemResponseへ変換する
 */
func toExpenseListItemResponse(expense models.Expense) types.ExpenseListItemResponse {
	return types.ExpenseListItemResponse{
		ID:       expense.ID,
		UserID:   expense.UserID,
		UserName: expense.User.Name,
		Email:    expense.User.Email,

		TargetMonth: expense.TargetMonth.Format("2006-01"),
		ExpenseDate: expense.ExpenseDate.Format("2006-01-02"),

		Amount:      expense.Amount,
		Description: expense.Description,
		Memo:        expense.Memo,

		HasReceiptFile:   expense.DriveFileID != nil && *expense.DriveFileID != "",
		OriginalFileName: expense.OriginalFileName,

		CreatedAt: expense.CreatedAt,
		UpdatedAt: expense.UpdatedAt,
	}
}

/*
 * 検索
 *
 * ページング方針：
 * ・初回は offset=0, limit=50
 * ・さらに表示するときは、フロントで現在表示済みの件数を offset として送る
 * ・limit が未指定、0以下の場合は 50件にする
 * ・limit が 50件を超える場合も 50件に丸める
 *
 * hasMore：
 * ・総件数 total が offset + 今回取得件数 より多ければ true
 * ・それ以下なら false
 */
func (service *expenseService) SearchExpenses(req types.SearchExpensesRequest) results.Result {
	normalizedCondition, normalizeResult := utils.NormalizePageSearchCondition(
		utils.PageSearchCondition{
			Keyword: req.Keyword,
			Offset:  req.Offset,
			Limit:   req.Limit,
		},
		"SEARCH_EXPENSES_INVALID_OFFSET",
		"検索開始位置が正しくありません",
	)
	if normalizeResult.Error {
		return normalizeResult
	}

	req.Keyword = normalizedCondition.Keyword
	req.Offset = normalizedCondition.Offset
	req.Limit = normalizedCondition.Limit

	searchQuery, countQuery, buildResult := service.expenseBuilder.BuildSearchExpensesQuery(req)
	if buildResult.Error {
		return buildResult
	}

	expenses, findResult := service.expenseRepository.FindExpenses(searchQuery)
	if findResult.Error {
		return findResult
	}

	total, countResult := service.expenseRepository.CountExpenses(countQuery)
	if countResult.Error {
		return countResult
	}

	expenseResponses := make([]types.ExpenseListItemResponse, 0, len(expenses))
	for _, expense := range expenses {
		expenseResponses = append(expenseResponses, toExpenseListItemResponse(expense))
	}

	hasMore := utils.HasMore(total, req.Offset, len(expenses))

	return results.OK(
		types.SearchExpensesResponse{
			Expenses: expenseResponses,
			Total:    total,
			Offset:   req.Offset,
			Limit:    req.Limit,
			HasMore:  hasMore,
		},
		"SEARCH_EXPENSES_SUCCESS",
		"経費一覧を取得しました",
		nil,
	)
}

/*
 * 詳細
 */
func (service *expenseService) GetExpenseDetail(req types.ExpenseDetailRequest) results.Result {
	findQuery, buildFindResult := service.expenseBuilder.BuildFindExpenseByIDQuery(req.ExpenseID)
	if buildFindResult.Error {
		return buildFindResult
	}

	expense, findResult := service.expenseRepository.FindExpense(findQuery)
	if findResult.Error {
		return findResult
	}

	return results.OK(
		types.ExpenseDetailResponse{
			Expense: toExpenseResponse(expense),
		},
		"GET_EXPENSE_DETAIL_SUCCESS",
		"経費詳細を取得しました",
		nil,
	)
}

/*
 * 新規作成
 *
 * 領収書ファイルがある場合：
 * 1. 経費本体をDBへ作成
 * 2. Google Driveへ領収書をアップロード
 * 3. Drive情報をExpenseへ反映して保存
 * 4. 詳細再取得して返却
 */
func (service *expenseService) CreateExpense(ctx context.Context, req types.CreateExpenseRequest) results.Result {
	expense, buildExpenseResult := service.expenseBuilder.BuildCreateExpenseModel(req)
	if buildExpenseResult.Error {
		return buildExpenseResult
	}

	createdExpense, createResult := service.expenseRepository.CreateExpense(expense)
	if createResult.Error {
		return createResult
	}

	if req.ReceiptFile != nil {
		expenseWithReceipt, uploadResult := service.uploadReceiptFileAndApplyToExpense(ctx, createdExpense, req.ReceiptFile)
		if uploadResult.Error {
			return uploadResult
		}

		savedExpense, saveResult := service.expenseRepository.SaveExpense(expenseWithReceipt)
		if saveResult.Error {
			return saveResult
		}

		createdExpense = savedExpense
	}

	foundExpense, findResult := service.findExpenseByID(createdExpense.ID)
	if findResult.Error {
		return findResult
	}

	return results.Created(
		types.CreateExpenseResponse{
			Expense: toExpenseResponse(foundExpense),
		},
		"CREATE_EXPENSE_SUCCESS",
		"経費を作成しました",
		nil,
	)
}

/*
 * 更新
 *
 * ReceiptFile が nil の場合：
 * ・経費本体のみ更新
 *
 * ReceiptFile がある場合：
 * ・Google Driveへ新規アップロード
 * ・Expenseの領収書情報を差し替え
 *
 * 注意：
 * ・既存Driveファイルの削除はここでは行わない
 * ・履歴・監査の観点で、古いファイルをDrive上に残す運用もあり得るため
 */
func (service *expenseService) UpdateExpense(ctx context.Context, req types.UpdateExpenseRequest) results.Result {
	findQuery, buildFindResult := service.expenseBuilder.BuildFindExpenseByIDQuery(req.ExpenseID)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentExpense, findResult := service.expenseRepository.FindExpense(findQuery)
	if findResult.Error {
		return findResult
	}

	updatedExpense, buildUpdateResult := service.expenseBuilder.BuildUpdateExpenseModel(
		currentExpense,
		req,
	)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	if req.ReceiptFile != nil {
		expenseWithReceipt, uploadResult := service.uploadReceiptFileAndApplyToExpense(ctx, updatedExpense, req.ReceiptFile)
		if uploadResult.Error {
			return uploadResult
		}

		updatedExpense = expenseWithReceipt
	}

	savedExpense, saveResult := service.expenseRepository.SaveExpense(updatedExpense)
	if saveResult.Error {
		return saveResult
	}

	foundExpense, findSavedResult := service.findExpenseByID(savedExpense.ID)
	if findSavedResult.Error {
		return findSavedResult
	}

	return results.OK(
		types.UpdateExpenseResponse{
			Expense: toExpenseResponse(foundExpense),
		},
		"UPDATE_EXPENSE_SUCCESS",
		"経費を更新しました",
		nil,
	)
}

/*
 * 論理削除
 */
func (service *expenseService) DeleteExpense(req types.DeleteExpenseRequest) results.Result {
	findQuery, buildFindResult := service.expenseBuilder.BuildFindExpenseByIDQuery(req.ExpenseID)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentExpense, findResult := service.expenseRepository.FindExpense(findQuery)
	if findResult.Error {
		return findResult
	}

	deletedExpense, buildDeleteResult := service.expenseBuilder.BuildDeleteExpenseModel(currentExpense)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	_, saveResult := service.expenseRepository.SaveExpense(deletedExpense)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteExpenseResponse{
			ExpenseID: req.ExpenseID,
		},
		"DELETE_EXPENSE_SUCCESS",
		"経費を削除しました",
		nil,
	)
}

/*
 * 領収書ファイル取得
 *
 * Controller側で DataFromReader する。
 */
func (service *expenseService) DownloadExpenseReceipt(ctx context.Context, req types.ViewExpenseReceiptRequest) (types.ExpenseReceiptFileResponse, results.Result) {
	if service.googleDriveService == nil {
		return types.ExpenseReceiptFileResponse{}, results.InternalServerError(
			"GOOGLE_DRIVE_SERVICE_NOT_INITIALIZED",
			"Google Drive連携が初期化されていません",
			nil,
		)
	}

	findQuery, buildFindResult := service.expenseBuilder.BuildFindExpenseByIDQuery(req.ExpenseID)
	if buildFindResult.Error {
		return types.ExpenseReceiptFileResponse{}, buildFindResult
	}

	expense, findResult := service.expenseRepository.FindExpense(findQuery)
	if findResult.Error {
		return types.ExpenseReceiptFileResponse{}, findResult
	}

	if expense.DriveFileID == nil || strings.TrimSpace(*expense.DriveFileID) == "" {
		return types.ExpenseReceiptFileResponse{}, results.NotFound(
			"EXPENSE_RECEIPT_NOT_FOUND",
			"経費領収書が見つかりません",
			map[string]any{
				"expenseId": req.ExpenseID,
			},
		)
	}

	downloadedFile, err := service.googleDriveService.DownloadFile(ctx, *expense.DriveFileID)
	if err != nil {
		return types.ExpenseReceiptFileResponse{}, results.InternalServerError(
			"DOWNLOAD_EXPENSE_RECEIPT_FAILED",
			"経費領収書の取得に失敗しました",
			err.Error(),
		)
	}

	fileName := downloadedFile.FileName
	if expense.OriginalFileName != nil && strings.TrimSpace(*expense.OriginalFileName) != "" {
		fileName = *expense.OriginalFileName
	}

	return types.ExpenseReceiptFileResponse{
			Body:      downloadedFile.Body,
			FileName:  fileName,
			MimeType:  downloadedFile.MimeType,
			SizeBytes: downloadedFile.SizeBytes,
		}, results.OK(
			nil,
			"DOWNLOAD_EXPENSE_RECEIPT_SUCCESS",
			"",
			nil,
		)
}

/*
 * ID指定で経費を再取得する。
 */
func (service *expenseService) findExpenseByID(expenseID uint) (models.Expense, results.Result) {
	findQuery, buildFindResult := service.expenseBuilder.BuildFindExpenseByIDQuery(expenseID)
	if buildFindResult.Error {
		return models.Expense{}, buildFindResult
	}

	foundExpense, findResult := service.expenseRepository.FindExpense(findQuery)
	if findResult.Error {
		return models.Expense{}, findResult
	}

	return foundExpense, results.OK(
		nil,
		"FIND_EXPENSE_BY_ID_SUCCESS",
		"",
		nil,
	)
}

/*
 * 領収書ファイルをGoogle Driveへアップロードし、Expenseへ反映する。
 */
func (service *expenseService) uploadReceiptFileAndApplyToExpense(
	ctx context.Context,
	expense models.Expense,
	receiptFileHeader *multipart.FileHeader,
) (models.Expense, results.Result) {
	if service.googleDriveService == nil {
		return models.Expense{}, results.InternalServerError(
			"GOOGLE_DRIVE_SERVICE_NOT_INITIALIZED",
			"Google Drive連携が初期化されていません",
			nil,
		)
	}

	if receiptFileHeader == nil {
		return expense, results.OK(
			nil,
			"UPLOAD_RECEIPT_FILE_SKIPPED",
			"",
			nil,
		)
	}

	storageLink, storageLinkResult := service.findExpenseReceiptStorageLink()
	if storageLinkResult.Error {
		return models.Expense{}, storageLinkResult
	}

	folderID, err := service.googleDriveService.ParseFolderID(storageLink.URL)
	if err != nil {
		return models.Expense{}, results.BadRequest(
			"EXPENSE_RECEIPT_STORAGE_LINK_INVALID_URL",
			"経費レシート格納先のGoogle DriveフォルダURLが正しくありません",
			err.Error(),
		)
	}

	receiptFile, err := receiptFileHeader.Open()
	if err != nil {
		return models.Expense{}, results.BadRequest(
			"OPEN_EXPENSE_RECEIPT_FILE_FAILED",
			"領収書ファイルを開けませんでした",
			err.Error(),
		)
	}
	defer receiptFile.Close()

	originalFileName := strings.TrimSpace(receiptFileHeader.Filename)
	storedFileName := storage.BuildGoogleDriveStoredFileName(
		"expense_receipt",
		expense.UserID,
		time.Now().Format("20060102_150405"),
		originalFileName,
	)

	mimeType := strings.TrimSpace(receiptFileHeader.Header.Get("Content-Type"))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	uploadedFile, err := service.googleDriveService.UploadFile(
		ctx,
		folderID,
		storedFileName,
		mimeType,
		receiptFile,
	)
	if err != nil {
		return models.Expense{}, results.InternalServerError(
			"UPLOAD_EXPENSE_RECEIPT_FAILED",
			"領収書ファイルのアップロードに失敗しました",
			err.Error(),
		)
	}

	sizeBytes := receiptFileHeader.Size
	if uploadedFile.SizeBytes > 0 {
		sizeBytes = uploadedFile.SizeBytes
	}

	expenseWithReceipt, buildApplyResult := service.expenseBuilder.BuildApplyReceiptFileModel(
		expense,
		builders.ExpenseReceiptFileModel{
			OriginalFileName:      originalFileName,
			StoredFileName:        uploadedFile.FileName,
			FileURL:               uploadedFile.FileURL,
			DriveFileID:           uploadedFile.DriveFileID,
			ExternalStorageLinkID: storageLink.ID,
			MimeType:              mimeType,
			SizeBytes:             sizeBytes,
		},
	)
	if buildApplyResult.Error {
		return models.Expense{}, buildApplyResult
	}

	return expenseWithReceipt, results.OK(
		nil,
		"UPLOAD_RECEIPT_FILE_AND_APPLY_TO_EXPENSE_SUCCESS",
		"",
		nil,
	)
}

/*
 * 経費レシート保存先を external_storage_links から取得する。
 */
func (service *expenseService) findExpenseReceiptStorageLink() (models.ExternalStorageLink, results.Result) {
	storageLinkQuery, buildStorageLinkResult := service.expenseBuilder.BuildFindExpenseReceiptStorageLinkQuery()
	if buildStorageLinkResult.Error {
		return models.ExternalStorageLink{}, buildStorageLinkResult
	}

	storageLink, findStorageLinkResult := service.externalStorageLinkRepository.FindExternalStorageLink(storageLinkQuery)
	if findStorageLinkResult.Error {
		return models.ExternalStorageLink{}, results.BadRequest(
			"EXPENSE_RECEIPT_STORAGE_LINK_NOT_FOUND",
			"経費レシート格納先が設定されていません",
			map[string]any{
				"linkType": builders.ExpenseReceiptExternalStorageLinkType,
				"reason":   findStorageLinkResult.Message,
			},
		)
	}

	if strings.TrimSpace(storageLink.URL) == "" {
		return models.ExternalStorageLink{}, results.BadRequest(
			"EXPENSE_RECEIPT_STORAGE_LINK_URL_EMPTY",
			"経費レシート格納先URLが設定されていません",
			map[string]any{
				"externalStorageLinkId": storageLink.ID,
			},
		)
	}

	return storageLink, results.OK(
		nil,
		"FIND_EXPENSE_RECEIPT_STORAGE_LINK_SUCCESS",
		"",
		nil,
	)
}

/*
 * デバッグ用の安全なファイル説明。
 */
func describeReceiptFile(fileHeader *multipart.FileHeader) string {
	if fileHeader == nil {
		return ""
	}

	return fmt.Sprintf(
		"name=%s,size=%d,mime=%s",
		fileHeader.Filename,
		fileHeader.Size,
		fileHeader.Header.Get("Content-Type"),
	)
}
