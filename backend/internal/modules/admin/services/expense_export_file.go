package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/storage"
)

const expenseExportContentTypeZIP = "application/zip"

// expenseExportReceipt は、Excelへ埋め込む領収書データを保持する。
// JPEG / PNG / GIF はDriveから取得した原本画像をそのまま埋め込む。
// PDF は経費出力処理の中だけで先頭ページをPNGへ変換して埋め込む。
// 変換・埋め込みに失敗した場合だけ、原本をZIP内の「領収書_その他」へ残す。
type expenseExportReceipt struct {
	ExpenseID      uint
	FileName       string
	OriginalBody   []byte
	EmbeddedBody   []byte
	MimeType       string
	EmbeddedImage  bool
	MediaExtension string
	ImageWidth     int
	ImageHeight    int
}

func hasExpenseReceipt(expenses []models.Expense) bool {
	for _, expense := range expenses {
		if expense.DriveFileID != nil && strings.TrimSpace(*expense.DriveFileID) != "" {
			return true
		}
	}
	return false
}

func buildExpenseExportZip(
	ctx context.Context,
	expenses []models.Expense,
	googleDriveService storage.GoogleDriveService,
	exportedAt time.Time,
) (types.ExpenseExportFileResponse, error) {
	receipts, err := downloadExpenseExportReceipts(ctx, expenses, googleDriveService)
	if err != nil {
		return types.ExpenseExportFileResponse{}, err
	}

	xlsxBody, err := buildExpenseExportXLSX(expenses, receipts, exportedAt)
	if err != nil {
		return types.ExpenseExportFileResponse{}, err
	}

	baseName := buildExpenseExportBaseName(expenses, exportedAt)
	excelFileName := baseName + "_経費集計.xlsx"
	zipFileName := baseName + "_経費一式.zip"

	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)

	if err := writeZIPFile(zipWriter, excelFileName, xlsxBody); err != nil {
		return types.ExpenseExportFileResponse{}, fmt.Errorf("failed to add expense xlsx to zip: %w", err)
	}

	// 画像領収書はExcel内部へ埋め込むため、ZIPへ重複して出力しない。
	// PDFも先頭ページをPNG化してExcelへ埋め込む。
	// 画像化・埋め込みに失敗した領収書だけは、原本欠落を防ぐため同梱する。
	for _, receipt := range receipts {
		if receipt.EmbeddedImage {
			continue
		}
		if err := writeZIPFile(
			zipWriter,
			path.Join("領収書_その他", receipt.FileName),
			receipt.OriginalBody,
		); err != nil {
			return types.ExpenseExportFileResponse{}, fmt.Errorf("failed to add non-image receipt to zip: %w", err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return types.ExpenseExportFileResponse{}, fmt.Errorf("failed to close expense zip: %w", err)
	}

	return types.ExpenseExportFileResponse{
		Body:        zipBuffer.Bytes(),
		FileName:    zipFileName,
		ContentType: expenseExportContentTypeZIP,
	}, nil
}

func downloadExpenseExportReceipts(
	ctx context.Context,
	expenses []models.Expense,
	googleDriveService storage.GoogleDriveService,
) ([]expenseExportReceipt, error) {
	receipts := make([]expenseExportReceipt, 0)

	for _, expense := range expenses {
		if expense.DriveFileID == nil || strings.TrimSpace(*expense.DriveFileID) == "" {
			continue
		}

		if googleDriveService == nil {
			return nil, fmt.Errorf("google drive service is nil")
		}

		downloadedFile, err := googleDriveService.DownloadFile(ctx, *expense.DriveFileID)
		if err != nil {
			return nil, fmt.Errorf("failed to download receipt for expense %d: %w", expense.ID, err)
		}

		body, readErr := io.ReadAll(downloadedFile.Body)
		closeErr := downloadedFile.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("failed to read receipt for expense %d: %w", expense.ID, readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("failed to close receipt for expense %d: %w", expense.ID, closeErr)
		}

		originalFileName := downloadedFile.FileName
		if expense.OriginalFileName != nil && strings.TrimSpace(*expense.OriginalFileName) != "" {
			originalFileName = *expense.OriginalFileName
		}

		receiptFileName := buildExpenseReceiptExportFileName(expense, originalFileName)
		mimeType := strings.TrimSpace(downloadedFile.MimeType)
		if mimeType == "" && expense.MimeType != nil {
			mimeType = strings.TrimSpace(*expense.MimeType)
		}

		embeddedBody := body
		mediaExtension, width, height, embeddedImage := detectExpenseEmbeddedImage(body, receiptFileName, mimeType)

		// PDF はGoogle Driveのサムネイル等には依存せず、
		// backendコンテナ内の pdftoppm で先頭ページだけPNGへ変換する。
		if !embeddedImage && isExpensePDF(body, receiptFileName, mimeType) {
			pngBody, convertErr := renderExpensePDFToPNG(ctx, body)
			if convertErr == nil {
				pngExtension, pngWidth, pngHeight, pngEmbedded := detectExpenseEmbeddedImage(
					pngBody,
					receiptFileName+".png",
					"image/png",
				)
				if pngEmbedded {
					embeddedBody = pngBody
					mediaExtension = pngExtension
					width = pngWidth
					height = pngHeight
					embeddedImage = true
				}
			}
		}

		receipts = append(receipts, expenseExportReceipt{
			ExpenseID:      expense.ID,
			FileName:       receiptFileName,
			OriginalBody:   body,
			EmbeddedBody:   embeddedBody,
			MimeType:       mimeType,
			EmbeddedImage:  embeddedImage,
			MediaExtension: mediaExtension,
			ImageWidth:     width,
			ImageHeight:    height,
		})
	}

	return receipts, nil
}

func detectExpenseEmbeddedImage(body []byte, fileName string, mimeType string) (string, int, int, bool) {
	if len(body) == 0 {
		return "", 0, 0, false
	}

	detectedMimeType := strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(body), ";")[0]))
	providedMimeType := strings.ToLower(strings.TrimSpace(strings.Split(mimeType, ";")[0]))
	extension := strings.ToLower(path.Ext(fileName))

	mediaExtension := ""
	switch {
	case detectedMimeType == "image/jpeg" || providedMimeType == "image/jpeg" || extension == ".jpg" || extension == ".jpeg":
		mediaExtension = "jpeg"
	case detectedMimeType == "image/png" || providedMimeType == "image/png" || extension == ".png":
		mediaExtension = "png"
	case detectedMimeType == "image/gif" || providedMimeType == "image/gif" || extension == ".gif":
		mediaExtension = "gif"
	default:
		return "", 0, 0, false
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil || config.Width <= 0 || config.Height <= 0 {
		return "", 0, 0, false
	}

	return mediaExtension, config.Width, config.Height, true
}

func isExpensePDF(body []byte, fileName string, mimeType string) bool {
	if len(body) == 0 {
		return false
	}

	detectedMimeType := strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(body), ";")[0]))
	providedMimeType := strings.ToLower(strings.TrimSpace(strings.Split(mimeType, ";")[0]))
	extension := strings.ToLower(path.Ext(fileName))

	return detectedMimeType == "application/pdf" || providedMimeType == "application/pdf" || extension == ".pdf"
}

// renderExpensePDFToPNG は経費出力専用。
// PDFの先頭ページだけをPNGへ変換し、Excel埋め込み用のbytesとして返す。
// 外部APIは使わず、backendコンテナ内の pdftoppm(poppler-utils) だけを利用する。
func renderExpensePDFToPNG(ctx context.Context, pdfBody []byte) ([]byte, error) {
	if len(pdfBody) == 0 {
		return nil, fmt.Errorf("pdf body is empty")
	}

	tempDir, err := os.MkdirTemp("", "timexeed-expense-pdf-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	pdfPath := filepath.Join(tempDir, "receipt.pdf")
	outputBase := filepath.Join(tempDir, "receipt")
	outputPNGPath := outputBase + ".png"

	if err := os.WriteFile(pdfPath, pdfBody, 0o600); err != nil {
		return nil, fmt.Errorf("failed to write temporary pdf: %w", err)
	}

	// 144dpiなら領収書の文字を確認しやすく、Excelが極端に巨大化しにくい。
	command := exec.CommandContext(
		ctx,
		"pdftoppm",
		"-png",
		"-f", "1",
		"-singlefile",
		"-r", "144",
		pdfPath,
		outputBase,
	)
	if output, err := command.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to convert pdf to png: %w: %s", err, strings.TrimSpace(string(output)))
	}

	pngBody, err := os.ReadFile(outputPNGPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read converted png: %w", err)
	}
	if len(pngBody) == 0 {
		return nil, fmt.Errorf("converted png is empty")
	}

	return pngBody, nil
}

func writeZIPFile(zipWriter *zip.Writer, filePath string, body []byte) error {
	header := &zip.FileHeader{
		Name:   filePath,
		Method: zip.Deflate,
	}
	header.SetModTime(time.Now())
	header.SetMode(0o644)

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = writer.Write(body)
	return err
}

func buildExpenseExportBaseName(expenses []models.Expense, exportedAt time.Time) string {
	userNames := uniqueExpenseExportUserNames(expenses)
	monthNames := uniqueExpenseExportMonths(expenses)

	userPart := "経費検索結果"
	if len(userNames) == 1 {
		userPart = sanitizeExpenseExportFileNamePart(userNames[0])
	} else if len(userNames) > 1 {
		userPart = fmt.Sprintf("経費検索結果_%d名", len(userNames))
	}

	monthPart := "対象月不明"
	if len(monthNames) == 1 {
		monthPart = strings.ReplaceAll(monthNames[0], "-", "年") + "月"
	} else if len(monthNames) > 1 {
		monthPart = strings.ReplaceAll(monthNames[0], "-", "年") + "月-" + strings.ReplaceAll(monthNames[len(monthNames)-1], "-", "年") + "月"
	}

	return fmt.Sprintf(
		"%s_%s_%s",
		userPart,
		monthPart,
		exportedAt.Format("20060102_150405"),
	)
}

func uniqueExpenseExportUserNames(expenses []models.Expense) []string {
	set := make(map[string]struct{})
	for _, expense := range expenses {
		name := strings.TrimSpace(expense.User.Name)
		if name == "" {
			name = fmt.Sprintf("user_%d", expense.UserID)
		}
		set[name] = struct{}{}
	}

	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func uniqueExpenseExportMonths(expenses []models.Expense) []string {
	set := make(map[string]struct{})
	for _, expense := range expenses {
		set[expense.TargetMonth.Format("2006-01")] = struct{}{}
	}

	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func expenseCategoryDisplayName(category string) string {
	switch category {
	case models.ExpenseCategoryTransportation:
		return "交通費系経費"
	case models.ExpenseCategorySupplies:
		return "備品系経費"
	case models.ExpenseCategoryCommunication:
		return "通信系経費"
	default:
		return "その他経費"
	}
}

func buildExpenseReceiptExportFileName(expense models.Expense, originalFileName string) string {
	extension := path.Ext(strings.TrimSpace(originalFileName))
	if len(extension) > 20 {
		extension = ""
	}

	description := sanitizeExpenseExportFileNamePart(expense.Description)
	if utf8.RuneCountInString(description) > 30 {
		description = string([]rune(description)[:30])
	}

	return fmt.Sprintf(
		"%s_%s_%d円_expense_%d%s",
		expense.ExpenseDate.Format("20060102"),
		description,
		expense.Amount,
		expense.ID,
		extension,
	)
}

func sanitizeExpenseExportFileNamePart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "未設定"
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
		"\r", "_",
		"\n", "_",
		"\t", "_",
	)

	return replacer.Replace(value)
}

// buildExpenseExportXLSX は外部ライブラリを追加せず、必要最小限のXLSXを生成する。
// 画像領収書はxl/media配下へ格納し、領収書シートのDrawingとしてExcel内部へ埋め込む。
func buildExpenseExportXLSX(
	expenses []models.Expense,
	receipts []expenseExportReceipt,
	exportedAt time.Time,
) ([]byte, error) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)

	receiptByExpenseID := make(map[uint]expenseExportReceipt, len(receipts))
	for _, receipt := range receipts {
		receiptByExpenseID[receipt.ExpenseID] = receipt
	}

	summarySheetXML := buildExpenseSummarySheetXML(expenses, receiptByExpenseID, exportedAt)
	receiptSheetXML, drawingXML, drawingRelationshipsXML, receiptSheetRelationshipsXML, embeddedReceipts := buildExpenseReceiptSheetXML(expenses, receiptByExpenseID)

	files := map[string][]byte{
		"[Content_Types].xml":        []byte(expenseXLSXContentTypesXML),
		"_rels/.rels":                []byte(expenseXLSXRootRelationshipsXML),
		"xl/workbook.xml":            []byte(expenseXLSXWorkbookXML),
		"xl/_rels/workbook.xml.rels": []byte(expenseXLSXWorkbookRelationshipsXML),
		"xl/styles.xml":              []byte(expenseXLSXStylesXML),
		"xl/worksheets/sheet1.xml":   []byte(summarySheetXML),
		"xl/worksheets/sheet2.xml":   []byte(receiptSheetXML),
	}

	if len(embeddedReceipts) > 0 {
		files["xl/worksheets/_rels/sheet2.xml.rels"] = []byte(receiptSheetRelationshipsXML)
		files["xl/drawings/drawing1.xml"] = []byte(drawingXML)
		files["xl/drawings/_rels/drawing1.xml.rels"] = []byte(drawingRelationshipsXML)
	}

	for index, receipt := range embeddedReceipts {
		mediaFileName := fmt.Sprintf("xl/media/receipt_%d.%s", index+1, receipt.MediaExtension)
		files[mediaFileName] = receipt.EmbeddedBody
	}

	fileNames := make([]string, 0, len(files))
	for fileName := range files {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	for _, fileName := range fileNames {
		fileWriter, err := writer.Create(fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to create xlsx entry %s: %w", fileName, err)
		}
		if _, err := fileWriter.Write(files[fileName]); err != nil {
			return nil, fmt.Errorf("failed to write xlsx entry %s: %w", fileName, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close xlsx: %w", err)
	}

	return buffer.Bytes(), nil
}

func buildExpenseSummarySheetXML(
	expenses []models.Expense,
	receiptByExpenseID map[uint]expenseExportReceipt,
	exportedAt time.Time,
) string {
	var rows strings.Builder

	rows.WriteString(`<row r="1" ht="28" customHeight="1">`)
	rows.WriteString(inlineStringCell("A1", "経費集計", 1))
	rows.WriteString(`</row>`)

	targetSummary := buildExpenseExportTargetSummary(expenses)
	rows.WriteString(`<row r="2">`)
	rows.WriteString(inlineStringCell("A2", "出力対象", 2))
	rows.WriteString(inlineStringCell("B2", targetSummary, 3))
	rows.WriteString(inlineStringCell("G2", "出力日時", 2))
	rows.WriteString(inlineStringCell("H2", exportedAt.Format("2006/01/02 15:04:05"), 3))
	rows.WriteString(`</row>`)

	rows.WriteString(`<row r="3">`)
	rows.WriteString(inlineStringCell("A3", "件数", 2))
	rows.WriteString(numberCell("B3", len(expenses), 3))
	rows.WriteString(inlineStringCell("G3", "合計金額", 2))
	rows.WriteString(numberCell("H3", sumExpenseAmounts(expenses), 10))
	rows.WriteString(`</row>`)

	headers := []string{"No.", "対象月", "経費発生日", "従業員", "メールアドレス", "カテゴリ", "内容", "メモ", "金額", "領収書"}
	rows.WriteString(`<row r="5" ht="24" customHeight="1">`)
	for index, header := range headers {
		rows.WriteString(inlineStringCell(cellReference(index+1, 5), header, 4))
	}
	rows.WriteString(`</row>`)

	receiptNumber := 0
	for index, expense := range expenses {
		row := index + 6
		rows.WriteString(`<row r="` + strconv.Itoa(row) + `" ht="22" customHeight="1">`)
		rows.WriteString(numberCell(cellReference(1, row), index+1, 5))
		rows.WriteString(inlineStringCell(cellReference(2, row), expense.TargetMonth.Format("2006年01月"), 5))
		rows.WriteString(inlineStringCell(cellReference(3, row), expense.ExpenseDate.Format("2006/01/02"), 6))
		rows.WriteString(inlineStringCell(cellReference(4, row), expense.User.Name, 5))
		rows.WriteString(inlineStringCell(cellReference(5, row), expense.User.Email, 5))
		rows.WriteString(inlineStringCell(cellReference(6, row), expenseCategoryDisplayName(expense.Category), 5))
		rows.WriteString(inlineStringCell(cellReference(7, row), expense.Description, 5))
		rows.WriteString(inlineStringCell(cellReference(8, row), stringPointerValue(expense.Memo), 5))
		rows.WriteString(numberCell(cellReference(9, row), expense.Amount, 7))

		receipt, hasReceipt := receiptByExpenseID[expense.ID]
		if hasReceipt {
			receiptNumber++
			if receipt.EmbeddedImage {
				rows.WriteString(inlineStringCell(cellReference(10, row), fmt.Sprintf("領収書シート No.%d", receiptNumber), 5))
			} else {
				rows.WriteString(inlineStringCell(cellReference(10, row), fmt.Sprintf("領収書シート No.%d（別ファイル同梱）", receiptNumber), 5))
			}
		} else {
			rows.WriteString(inlineStringCell(cellReference(10, row), "なし", 5))
		}
		rows.WriteString(`</row>`)
	}

	totalRow := len(expenses) + 6
	rows.WriteString(`<row r="` + strconv.Itoa(totalRow) + `" ht="24" customHeight="1">`)
	rows.WriteString(inlineStringCell(cellReference(8, totalRow), "合計", 9))
	if len(expenses) > 0 {
		rows.WriteString(
			formulaCell(
				cellReference(9, totalRow),
				fmt.Sprintf("SUM(I6:I%d)", totalRow-1),
				sumExpenseAmounts(expenses),
				10,
			),
		)
	} else {
		rows.WriteString(numberCell(cellReference(9, totalRow), 0, 10))
	}
	rows.WriteString(`</row>`)

	dimension := fmt.Sprintf("A1:J%d", totalRow)
	autoFilter := fmt.Sprintf("A5:J%d", totalRow-1)

	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
		`<dimension ref="` + dimension + `"/>` +
		`<sheetViews><sheetView workbookViewId="0"><pane ySplit="5" topLeftCell="A6" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>` +
		`<sheetFormatPr defaultRowHeight="18"/>` +
		`<cols>` +
		`<col min="1" max="1" width="7" customWidth="1"/>` +
		`<col min="2" max="3" width="15" customWidth="1"/>` +
		`<col min="4" max="4" width="18" customWidth="1"/>` +
		`<col min="5" max="5" width="30" customWidth="1"/>` +
		`<col min="6" max="6" width="18" customWidth="1"/>` +
		`<col min="7" max="7" width="34" customWidth="1"/>` +
		`<col min="8" max="8" width="30" customWidth="1"/>` +
		`<col min="9" max="9" width="14" customWidth="1"/>` +
		`<col min="10" max="10" width="30" customWidth="1"/>` +
		`</cols>` +
		`<sheetData>` + rows.String() + `</sheetData>` +
		`<autoFilter ref="` + autoFilter + `"/>` +
		`<mergeCells count="1"><mergeCell ref="A1:J1"/></mergeCells>` +
		`<pageMargins left="0.3" right="0.3" top="0.5" bottom="0.5" header="0.2" footer="0.2"/>` +
		`<pageSetup orientation="landscape" fitToWidth="1" fitToHeight="0"/>` +
		`</worksheet>`
}

func buildExpenseReceiptSheetXML(
	expenses []models.Expense,
	receiptByExpenseID map[uint]expenseExportReceipt,
) (string, string, string, string, []expenseExportReceipt) {
	const receiptBlockRows = 40

	var rows strings.Builder
	var drawingAnchors strings.Builder
	var drawingRelationships strings.Builder

	embeddedReceipts := make([]expenseExportReceipt, 0)
	receiptNumber := 0
	maxRow := 2

	for _, expense := range expenses {
		receipt, hasReceipt := receiptByExpenseID[expense.ID]
		if !hasReceipt {
			continue
		}

		receiptNumber++
		startRow := 1 + (receiptNumber-1)*receiptBlockRows
		maxRow = startRow + receiptBlockRows - 1

		rows.WriteString(`<row r="` + strconv.Itoa(startRow) + `" ht="24" customHeight="1">`)
		rows.WriteString(inlineStringCell(cellReference(1, startRow), fmt.Sprintf("領収書 No.%d", receiptNumber), 4))
		rows.WriteString(inlineStringCell(cellReference(2, startRow), expense.User.Name, 4))
		rows.WriteString(inlineStringCell(cellReference(3, startRow), expense.ExpenseDate.Format("2006/01/02"), 4))
		rows.WriteString(inlineStringCell(cellReference(4, startRow), expense.Description, 4))
		rows.WriteString(numberCell(cellReference(8, startRow), expense.Amount, 10))
		rows.WriteString(`</row>`)

		rows.WriteString(`<row r="` + strconv.Itoa(startRow+1) + `">`)
		rows.WriteString(inlineStringCell(cellReference(1, startRow+1), "元ファイル名", 2))
		rows.WriteString(inlineStringCell(cellReference(2, startRow+1), receipt.FileName, 3))
		rows.WriteString(`</row>`)

		if receipt.EmbeddedImage {
			embeddedReceipts = append(embeddedReceipts, receipt)
			imageIndex := len(embeddedReceipts)
			widthEMU, heightEMU := fitExpenseReceiptImageEMU(receipt.ImageWidth, receipt.ImageHeight)
			anchorRow := startRow + 2
			drawingAnchors.WriteString(buildExpenseReceiptDrawingAnchor(imageIndex, anchorRow, widthEMU, heightEMU))
			drawingRelationships.WriteString(
				`<Relationship Id="rId` + strconv.Itoa(imageIndex) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../media/receipt_` + strconv.Itoa(imageIndex) + `.` + receipt.MediaExtension + `"/>`,
			)
		} else {
			rows.WriteString(`<row r="` + strconv.Itoa(startRow+3) + `" ht="30" customHeight="1">`)
			rows.WriteString(inlineStringCell(cellReference(1, startRow+3), "この領収書は画像形式ではないためExcel内へ画像表示できません。ZIP内の「領収書_その他」に原本を同梱しています。", 3))
			rows.WriteString(`</row>`)
		}
	}

	if receiptNumber == 0 {
		rows.WriteString(`<row r="1"><c r="A1" s="3" t="inlineStr"><is><t>領収書はありません</t></is></c></row>`)
		maxRow = 1
	}

	sheet := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
		`<dimension ref="A1:I` + strconv.Itoa(maxRow) + `"/>` +
		`<sheetViews><sheetView workbookViewId="0"/></sheetViews>` +
		`<sheetFormatPr defaultRowHeight="18"/>` +
		`<cols>` +
		`<col min="1" max="1" width="18" customWidth="1"/>` +
		`<col min="2" max="2" width="28" customWidth="1"/>` +
		`<col min="3" max="3" width="16" customWidth="1"/>` +
		`<col min="4" max="7" width="24" customWidth="1"/>` +
		`<col min="8" max="8" width="16" customWidth="1"/>` +
		`</cols>` +
		`<sheetData>` + rows.String() + `</sheetData>`

	drawingXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<xdr:wsDr xmlns:xdr="http://schemas.openxmlformats.org/drawingml/2006/spreadsheetDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">` +
		drawingAnchors.String() +
		`</xdr:wsDr>`
	drawingRelationshipsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		drawingRelationships.String() +
		`</Relationships>`
	receiptSheetRelationshipsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/drawing" Target="../drawings/drawing1.xml"/>` +
		`</Relationships>`

	sheet += `<pageMargins left="0.3" right="0.3" top="0.5" bottom="0.5" header="0.2" footer="0.2"/>`
	if len(embeddedReceipts) > 0 {
		sheet += `<drawing r:id="rId1"/>`
	}
	sheet += `</worksheet>`

	return sheet, drawingXML, drawingRelationshipsXML, receiptSheetRelationshipsXML, embeddedReceipts
}

func fitExpenseReceiptImageEMU(width int, height int) (int64, int64) {
	const (
		maxWidthPX  = 960.0
		maxHeightPX = 720.0
		emuPerPixel = 9525.0
	)

	if width <= 0 || height <= 0 {
		return int64(640 * emuPerPixel), int64(480 * emuPerPixel)
	}

	scale := 1.0
	if float64(width) > maxWidthPX {
		scale = maxWidthPX / float64(width)
	}
	if float64(height)*scale > maxHeightPX {
		scale = maxHeightPX / float64(height)
	}

	return int64(float64(width) * scale * emuPerPixel), int64(float64(height) * scale * emuPerPixel)
}

func buildExpenseReceiptDrawingAnchor(imageIndex int, startRow int, widthEMU int64, heightEMU int64) string {
	zeroBasedRow := startRow - 1
	relID := "rId" + strconv.Itoa(imageIndex)
	name := "Receipt " + strconv.Itoa(imageIndex)

	return `<xdr:oneCellAnchor>` +
		`<xdr:from><xdr:col>0</xdr:col><xdr:colOff>0</xdr:colOff><xdr:row>` + strconv.Itoa(zeroBasedRow) + `</xdr:row><xdr:rowOff>0</xdr:rowOff></xdr:from>` +
		`<xdr:ext cx="` + strconv.FormatInt(widthEMU, 10) + `" cy="` + strconv.FormatInt(heightEMU, 10) + `"/>` +
		`<xdr:pic>` +
		`<xdr:nvPicPr><xdr:cNvPr id="` + strconv.Itoa(imageIndex) + `" name="` + xmlEscape(name) + `"/><xdr:cNvPicPr><a:picLocks noChangeAspect="1"/></xdr:cNvPicPr></xdr:nvPicPr>` +
		`<xdr:blipFill><a:blip xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" r:embed="` + relID + `"/><a:stretch><a:fillRect/></a:stretch></xdr:blipFill>` +
		`<xdr:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="` + strconv.FormatInt(widthEMU, 10) + `" cy="` + strconv.FormatInt(heightEMU, 10) + `"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></xdr:spPr>` +
		`</xdr:pic>` +
		`<xdr:clientData/>` +
		`</xdr:oneCellAnchor>`
}

func inlineStringCell(reference string, value string, style int) string {
	return `<c r="` + reference + `" s="` + strconv.Itoa(style) + `" t="inlineStr"><is><t xml:space="preserve">` + xmlEscape(value) + `</t></is></c>`
}

func numberCell(reference string, value int, style int) string {
	return `<c r="` + reference + `" s="` + strconv.Itoa(style) + `"><v>` + strconv.Itoa(value) + `</v></c>`
}

func formulaCell(reference string, formula string, value int, style int) string {
	return `<c r="` + reference + `" s="` + strconv.Itoa(style) + `"><f>` + xmlEscape(formula) + `</f><v>` + strconv.Itoa(value) + `</v></c>`
}

func xmlEscape(value string) string {
	value = removeInvalidXMLCharacters(value)

	var buffer bytes.Buffer
	_ = xml.EscapeText(&buffer, []byte(value))
	return buffer.String()
}

/*
 * XML 1.0で使用できない制御文字を除去する。
 *
 * メモや経費内容にコピー＆ペースト由来の制御文字が含まれていても、
 * sheet1.xml全体が壊れないようにする。
 */
func removeInvalidXMLCharacters(value string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == 0x09:
			return r
		case r == 0x0A:
			return r
		case r == 0x0D:
			return r
		case r >= 0x20 && r <= 0xD7FF:
			return r
		case r >= 0xE000 && r <= 0xFFFD:
			return r
		case r >= 0x10000 && r <= 0x10FFFF:
			return r
		default:
			return -1
		}
	}, value)
}

func cellReference(column int, row int) string {
	columnName := ""
	for column > 0 {
		column--
		columnName = string(rune('A'+column%26)) + columnName
		column /= 26
	}
	return columnName + strconv.Itoa(row)
}

func stringPointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func sumExpenseAmounts(expenses []models.Expense) int {
	total := 0
	for _, expense := range expenses {
		total += expense.Amount
	}
	return total
}

func buildExpenseExportTargetSummary(expenses []models.Expense) string {
	userNames := uniqueExpenseExportUserNames(expenses)
	months := uniqueExpenseExportMonths(expenses)

	userSummary := strings.Join(userNames, "、")
	if userSummary == "" {
		userSummary = "該当従業員"
	}

	monthSummary := "対象月不明"
	if len(months) == 1 {
		monthSummary = strings.ReplaceAll(months[0], "-", "年") + "月"
	} else if len(months) > 1 {
		monthSummary = strings.ReplaceAll(months[0], "-", "年") + "月 ～ " + strings.ReplaceAll(months[len(months)-1], "-", "年") + "月"
	}

	return userSummary + " / " + monthSummary
}

const expenseXLSXContentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Default Extension="jpg" ContentType="image/jpeg"/>
  <Default Extension="jpeg" ContentType="image/jpeg"/>
  <Default Extension="png" ContentType="image/png"/>
  <Default Extension="gif" ContentType="image/gif"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/worksheets/sheet2.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/drawings/drawing1.xml" ContentType="application/vnd.openxmlformats-officedocument.drawing+xml"/>
  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`

const expenseXLSXRootRelationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`

const expenseXLSXWorkbookXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <bookViews><workbookView xWindow="0" yWindow="0" windowWidth="24000" windowHeight="12000"/></bookViews>
  <sheets>
    <sheet name="経費集計" sheetId="1" r:id="rId1"/>
    <sheet name="領収書" sheetId="2" r:id="rId2"/>
  </sheets>
  <calcPr calcId="191029" fullCalcOnLoad="1"/>
</workbook>`

const expenseXLSXWorkbookRelationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

const expenseXLSXStylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <numFmts count="1"><numFmt numFmtId="164" formatCode="¥#,##0"/></numFmts>
  <fonts count="5">
    <font><sz val="11"/><name val="Yu Gothic"/><family val="2"/></font>
    <font><b/><sz val="18"/><color rgb="FFFFFFFF"/><name val="Yu Gothic"/><family val="2"/></font>
    <font><b/><sz val="11"/><color rgb="FFFFFFFF"/><name val="Yu Gothic"/><family val="2"/></font>
    <font><b/><sz val="11"/><color rgb="FF1F2937"/><name val="Yu Gothic"/><family val="2"/></font>
    <font><u/><sz val="11"/><color rgb="FF0563C1"/><name val="Yu Gothic"/><family val="2"/></font>
  </fonts>
  <fills count="6">
    <fill><patternFill patternType="none"/></fill>
    <fill><patternFill patternType="gray125"/></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFF97316"/><bgColor indexed="64"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FF1F2937"/><bgColor indexed="64"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFFFF7ED"/><bgColor indexed="64"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFF3F4F6"/><bgColor indexed="64"/></patternFill></fill>
  </fills>
  <borders count="2">
    <border><left/><right/><top/><bottom/><diagonal/></border>
    <border><left style="thin"><color rgb="FFD1D5DB"/></left><right style="thin"><color rgb="FFD1D5DB"/></right><top style="thin"><color rgb="FFD1D5DB"/></top><bottom style="thin"><color rgb="FFD1D5DB"/></bottom><diagonal/></border>
  </borders>
  <cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
  <cellXfs count="11">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
    <xf numFmtId="0" fontId="1" fillId="2" borderId="0" xfId="0" applyAlignment="1"><alignment horizontal="center" vertical="center"/></xf>
    <xf numFmtId="0" fontId="2" fillId="3" borderId="1" xfId="0" applyAlignment="1"><alignment horizontal="center" vertical="center"/></xf>
    <xf numFmtId="0" fontId="0" fillId="4" borderId="1" xfId="0" applyAlignment="1"><alignment vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="2" fillId="2" borderId="1" xfId="0" applyAlignment="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyAlignment="1"><alignment vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyAlignment="1"><alignment horizontal="center" vertical="center"/></xf>
    <xf numFmtId="164" fontId="0" fillId="0" borderId="1" xfId="0" applyNumberFormat="1" applyAlignment="1"><alignment horizontal="right" vertical="center"/></xf>
    <xf numFmtId="0" fontId="4" fillId="0" borderId="1" xfId="0" applyAlignment="1"><alignment horizontal="center" vertical="center"/></xf>
    <xf numFmtId="0" fontId="3" fillId="5" borderId="1" xfId="0" applyAlignment="1"><alignment horizontal="right" vertical="center"/></xf>
    <xf numFmtId="164" fontId="3" fillId="5" borderId="1" xfId="0" applyNumberFormat="1" applyAlignment="1"><alignment horizontal="right" vertical="center"/></xf>
  </cellXfs>
  <cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`
