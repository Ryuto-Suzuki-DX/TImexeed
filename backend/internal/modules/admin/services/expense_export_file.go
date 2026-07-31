package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"path"
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

// expenseExportReceipt は、ZIP内の領収書ファイルとExcelリンクを対応させる。
type expenseExportReceipt struct {
	ExpenseID uint
	ZIPPath   string
	Body      []byte
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
	receipts, receiptPathByExpenseID, err := downloadExpenseExportReceipts(ctx, expenses, googleDriveService)
	if err != nil {
		return types.ExpenseExportFileResponse{}, err
	}

	xlsxBody, err := buildExpenseExportXLSX(expenses, receiptPathByExpenseID, exportedAt)
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

	for _, receipt := range receipts {
		if err := writeZIPFile(zipWriter, receipt.ZIPPath, receipt.Body); err != nil {
			return types.ExpenseExportFileResponse{}, fmt.Errorf("failed to add receipt to zip: %w", err)
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
) ([]expenseExportReceipt, map[uint]string, error) {
	receipts := make([]expenseExportReceipt, 0)
	receiptPathByExpenseID := make(map[uint]string)

	for _, expense := range expenses {
		if expense.DriveFileID == nil || strings.TrimSpace(*expense.DriveFileID) == "" {
			continue
		}

		if googleDriveService == nil {
			return nil, nil, fmt.Errorf("google drive service is nil")
		}

		downloadedFile, err := googleDriveService.DownloadFile(ctx, *expense.DriveFileID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to download receipt for expense %d: %w", expense.ID, err)
		}

		body, readErr := io.ReadAll(downloadedFile.Body)
		closeErr := downloadedFile.Body.Close()
		if readErr != nil {
			return nil, nil, fmt.Errorf("failed to read receipt for expense %d: %w", expense.ID, readErr)
		}
		if closeErr != nil {
			return nil, nil, fmt.Errorf("failed to close receipt for expense %d: %w", expense.ID, closeErr)
		}

		originalFileName := downloadedFile.FileName
		if expense.OriginalFileName != nil && strings.TrimSpace(*expense.OriginalFileName) != "" {
			originalFileName = *expense.OriginalFileName
		}

		receiptFileName := buildExpenseReceiptExportFileName(expense, originalFileName)
		receiptZIPPath := path.Join("領収書", receiptFileName)

		receipts = append(receipts, expenseExportReceipt{
			ExpenseID: expense.ID,
			ZIPPath:   receiptZIPPath,
			Body:      body,
		})
		receiptPathByExpenseID[expense.ID] = receiptZIPPath
	}

	return receipts, receiptPathByExpenseID, nil
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
func buildExpenseExportXLSX(
	expenses []models.Expense,
	receiptPathByExpenseID map[uint]string,
	exportedAt time.Time,
) ([]byte, error) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)

	sheetXML, relationshipsXML := buildExpenseSheetXML(expenses, receiptPathByExpenseID, exportedAt)

	files := map[string]string{
		"[Content_Types].xml":                 expenseXLSXContentTypesXML,
		"_rels/.rels":                         expenseXLSXRootRelationshipsXML,
		"xl/workbook.xml":                     expenseXLSXWorkbookXML,
		"xl/_rels/workbook.xml.rels":          expenseXLSXWorkbookRelationshipsXML,
		"xl/styles.xml":                       expenseXLSXStylesXML,
		"xl/worksheets/sheet1.xml":            sheetXML,
		"xl/worksheets/_rels/sheet1.xml.rels": relationshipsXML,
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
		if _, err := fileWriter.Write([]byte(files[fileName])); err != nil {
			return nil, fmt.Errorf("failed to write xlsx entry %s: %w", fileName, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close xlsx: %w", err)
	}

	return buffer.Bytes(), nil
}

func buildExpenseSheetXML(
	expenses []models.Expense,
	receiptPathByExpenseID map[uint]string,
	exportedAt time.Time,
) (string, string) {
	var rows strings.Builder
	var hyperlinks strings.Builder
	var relationships strings.Builder

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

	headers := []string{"No.", "対象月", "経費発生日", "従業員", "メールアドレス", "内容", "メモ", "金額", "領収書"}
	rows.WriteString(`<row r="5" ht="24" customHeight="1">`)
	for index, header := range headers {
		rows.WriteString(inlineStringCell(cellReference(index+1, 5), header, 4))
	}
	rows.WriteString(`</row>`)

	relationshipID := 1
	for index, expense := range expenses {
		row := index + 6
		rows.WriteString(`<row r="` + strconv.Itoa(row) + `" ht="22" customHeight="1">`)
		rows.WriteString(numberCell(cellReference(1, row), index+1, 5))
		rows.WriteString(inlineStringCell(cellReference(2, row), expense.TargetMonth.Format("2006年01月"), 5))
		rows.WriteString(inlineStringCell(cellReference(3, row), expense.ExpenseDate.Format("2006/01/02"), 6))
		rows.WriteString(inlineStringCell(cellReference(4, row), expense.User.Name, 5))
		rows.WriteString(inlineStringCell(cellReference(5, row), expense.User.Email, 5))
		rows.WriteString(inlineStringCell(cellReference(6, row), expense.Description, 5))
		rows.WriteString(inlineStringCell(cellReference(7, row), stringPointerValue(expense.Memo), 5))
		rows.WriteString(numberCell(cellReference(8, row), expense.Amount, 7))

		receiptPath, hasReceipt := receiptPathByExpenseID[expense.ID]
		if hasReceipt {
			cellRef := cellReference(9, row)
			rows.WriteString(inlineStringCell(cellRef, "領収書を開く", 8))
			relID := fmt.Sprintf("rId%d", relationshipID)
			hyperlinks.WriteString(`<hyperlink ref="` + cellRef + `" r:id="` + relID + `"/>`)
			relationships.WriteString(`<Relationship Id="` + relID + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink" Target="` + xmlEscape(buildExpenseRelationshipTarget(receiptPath)) + `" TargetMode="External"/>`)
			relationshipID++
		} else {
			rows.WriteString(inlineStringCell(cellReference(9, row), "なし", 5))
		}
		rows.WriteString(`</row>`)
	}

	totalRow := len(expenses) + 6
	rows.WriteString(`<row r="` + strconv.Itoa(totalRow) + `" ht="24" customHeight="1">`)
	rows.WriteString(inlineStringCell(cellReference(7, totalRow), "合計", 9))
	if len(expenses) > 0 {
		rows.WriteString(
			formulaCell(
				cellReference(8, totalRow),
				fmt.Sprintf("SUM(H6:H%d)", totalRow-1),
				sumExpenseAmounts(expenses),
				10,
			),
		)
	} else {
		rows.WriteString(numberCell(cellReference(8, totalRow), 0, 10))
	}
	rows.WriteString(`</row>`)

	dimension := fmt.Sprintf("A1:I%d", totalRow)
	autoFilter := fmt.Sprintf("A5:I%d", totalRow-1)

	sheet := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
		`<dimension ref="` + dimension + `"/>` +
		`<sheetViews><sheetView workbookViewId="0"><pane ySplit="5" topLeftCell="A6" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>` +
		`<sheetFormatPr defaultRowHeight="18"/>` +
		`<cols>` +
		`<col min="1" max="1" width="7" customWidth="1"/>` +
		`<col min="2" max="3" width="15" customWidth="1"/>` +
		`<col min="4" max="4" width="18" customWidth="1"/>` +
		`<col min="5" max="5" width="30" customWidth="1"/>` +
		`<col min="6" max="6" width="34" customWidth="1"/>` +
		`<col min="7" max="7" width="30" customWidth="1"/>` +
		`<col min="8" max="8" width="14" customWidth="1"/>` +
		`<col min="9" max="9" width="18" customWidth="1"/>` +
		`</cols>` +
		`<sheetData>` + rows.String() + `</sheetData>` +
		`<autoFilter ref="` + autoFilter + `"/>` +
		`<mergeCells count="1"><mergeCell ref="A1:I1"/></mergeCells>`

	if hyperlinks.Len() > 0 {
		sheet += `<hyperlinks>` + hyperlinks.String() + `</hyperlinks>`
	}

	sheet += `<pageMargins left="0.3" right="0.3" top="0.5" bottom="0.5" header="0.2" footer="0.2"/>` +
		`<pageSetup orientation="landscape" fitToWidth="1" fitToHeight="0"/>` +
		`</worksheet>`

	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		relationships.String() +
		`</Relationships>`

	return sheet, rels
}

func buildExpenseRelationshipTarget(filePath string) string {
	parts := strings.Split(filePath, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
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
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`

const expenseXLSXRootRelationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`

const expenseXLSXWorkbookXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <bookViews><workbookView xWindow="0" yWindow="0" windowWidth="24000" windowHeight="12000"/></bookViews>
  <sheets><sheet name="経費集計" sheetId="1" r:id="rId1"/></sheets>
  <calcPr calcId="191029" fullCalcOnLoad="1"/>
</workbook>`

const expenseXLSXWorkbookRelationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
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
