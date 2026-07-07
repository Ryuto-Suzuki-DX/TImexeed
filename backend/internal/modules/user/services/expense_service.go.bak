package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/storage"
	"timexeed/backend/internal/utils"
)

type ExpenseService interface {
	SearchExpenses(userID uint, req types.SearchExpensesRequest) results.Result
	GetExpenseDetail(userID uint, req types.ExpenseDetailRequest) results.Result
	CreateExpense(ctx context.Context, userID uint, req types.CreateExpenseRequest) results.Result
	UpdateExpense(ctx context.Context, userID uint, req types.UpdateExpenseRequest) results.Result
	DeleteExpense(userID uint, req types.DeleteExpenseRequest) results.Result
	DownloadExpenseReceipt(ctx context.Context, userID uint, req types.ViewExpenseReceiptRequest) (types.ExpenseReceiptFileResponse, results.Result)
}

type expenseService struct {
	expenseBuilder     builders.ExpenseBuilder
	expenseRepository  repositories.ExpenseRepository
	googleDriveService storage.GoogleDriveService
}

func NewExpenseService(
	expenseBuilder builders.ExpenseBuilder,
	expenseRepository repositories.ExpenseRepository,
	googleDriveService storage.GoogleDriveService,
) ExpenseService {
	return &expenseService{
		expenseBuilder:     expenseBuilder,
		expenseRepository:  expenseRepository,
		googleDriveService: googleDriveService,
	}
}

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

func (service *expenseService) SearchExpenses(userID uint, req types.SearchExpensesRequest) results.Result {
	if userID == 0 {
		return results.Unauthorized("SEARCH_EXPENSES_INVALID_USER_ID", "認証情報のユーザーIDが正しくありません", nil)
	}

	normalizedCondition, normalizeResult := utils.NormalizePageSearchCondition(
		utils.PageSearchCondition{
			Offset: req.Offset,
			Limit:  req.Limit,
		},
		"SEARCH_EXPENSES_INVALID_OFFSET",
		"検索開始位置が正しくありません",
	)
	if normalizeResult.Error {
		return normalizeResult
	}

	req.Offset = normalizedCondition.Offset
	req.Limit = normalizedCondition.Limit

	searchQuery, countQuery, buildResult := service.expenseBuilder.BuildSearchExpensesQuery(userID, req)
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

	return results.OK(
		types.SearchExpensesResponse{
			Expenses: expenseResponses,
			Total:    total,
			Offset:   req.Offset,
			Limit:    req.Limit,
			HasMore:  utils.HasMore(total, req.Offset, len(expenses)),
		},
		"SEARCH_EXPENSES_SUCCESS",
		"経費一覧を取得しました",
		nil,
	)
}

func (service *expenseService) GetExpenseDetail(userID uint, req types.ExpenseDetailRequest) results.Result {
	if userID == 0 {
		return results.Unauthorized("GET_EXPENSE_DETAIL_INVALID_USER_ID", "認証情報のユーザーIDが正しくありません", nil)
	}

	expense, findResult := service.findExpenseByID(userID, req.ExpenseID)
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

func (service *expenseService) CreateExpense(ctx context.Context, userID uint, req types.CreateExpenseRequest) results.Result {
	if userID == 0 {
		return results.Unauthorized("CREATE_EXPENSE_INVALID_USER_ID", "認証情報のユーザーIDが正しくありません", nil)
	}

	expense, buildExpenseResult := service.expenseBuilder.BuildCreateExpenseModel(userID, req)
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

	foundExpense, findResult := service.findExpenseByID(userID, createdExpense.ID)
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

func (service *expenseService) UpdateExpense(ctx context.Context, userID uint, req types.UpdateExpenseRequest) results.Result {
	if userID == 0 {
		return results.Unauthorized("UPDATE_EXPENSE_INVALID_USER_ID", "認証情報のユーザーIDが正しくありません", nil)
	}

	currentExpense, findResult := service.findExpenseByID(userID, req.ExpenseID)
	if findResult.Error {
		return findResult
	}

	updatedExpense, buildUpdateResult := service.expenseBuilder.BuildUpdateExpenseModel(currentExpense, req)
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

	foundExpense, findSavedResult := service.findExpenseByID(userID, savedExpense.ID)
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

func (service *expenseService) DeleteExpense(userID uint, req types.DeleteExpenseRequest) results.Result {
	if userID == 0 {
		return results.Unauthorized("DELETE_EXPENSE_INVALID_USER_ID", "認証情報のユーザーIDが正しくありません", nil)
	}

	currentExpense, findResult := service.findExpenseByID(userID, req.ExpenseID)
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

func (service *expenseService) DownloadExpenseReceipt(ctx context.Context, userID uint, req types.ViewExpenseReceiptRequest) (types.ExpenseReceiptFileResponse, results.Result) {
	if userID == 0 {
		return types.ExpenseReceiptFileResponse{}, results.Unauthorized("DOWNLOAD_EXPENSE_RECEIPT_INVALID_USER_ID", "認証情報のユーザーIDが正しくありません", nil)
	}

	if service.googleDriveService == nil {
		return types.ExpenseReceiptFileResponse{}, results.InternalServerError(
			"GOOGLE_DRIVE_SERVICE_NOT_INITIALIZED",
			"Google Drive連携が初期化されていません",
			nil,
		)
	}

	expense, findResult := service.findExpenseByID(userID, req.ExpenseID)
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

func (service *expenseService) findExpenseByID(userID uint, expenseID uint) (models.Expense, results.Result) {
	findQuery, buildFindResult := service.expenseBuilder.BuildFindExpenseByIDQuery(userID, expenseID)
	if buildFindResult.Error {
		return models.Expense{}, buildFindResult
	}

	foundExpense, findResult := service.expenseRepository.FindExpense(findQuery)
	if findResult.Error {
		return models.Expense{}, findResult
	}

	return foundExpense, results.OK(nil, "FIND_EXPENSE_BY_ID_SUCCESS", "", nil)
}

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
		return expense, results.OK(nil, "UPLOAD_RECEIPT_FILE_SKIPPED", "", nil)
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

	uploadedFile, err := service.googleDriveService.UploadFile(ctx, folderID, storedFileName, mimeType, receiptFile)
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

	return expenseWithReceipt, results.OK(nil, "UPLOAD_RECEIPT_FILE_AND_APPLY_TO_EXPENSE_SUCCESS", "", nil)
}

func (service *expenseService) findExpenseReceiptStorageLink() (models.ExternalStorageLink, results.Result) {
	storageLinkQuery, buildStorageLinkResult := service.expenseBuilder.BuildFindExpenseReceiptStorageLinkQuery()
	if buildStorageLinkResult.Error {
		return models.ExternalStorageLink{}, buildStorageLinkResult
	}

	storageLink, findStorageLinkResult := service.expenseRepository.FindExternalStorageLink(storageLinkQuery)
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

	return storageLink, results.OK(nil, "FIND_EXPENSE_RECEIPT_STORAGE_LINK_SUCCESS", "", nil)
}

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
