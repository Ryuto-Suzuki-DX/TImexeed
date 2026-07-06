package builders

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 月次勤怠集計CSV出力 Builder interface
 *
 * 注意：
 * ・CSVヘッダーとCSVレコードの生成を担当する
 * ・勤怠計算そのものはServiceで行う
 */
type MonthlyAttendanceSummaryExportBuilder interface {
	BuildCSV(rows []types.MonthlyAttendanceSummaryCsvRow) ([]byte, results.Result)
	BuildExcel(rows []types.MonthlyAttendanceSummaryCsvRow, targetYear int, targetMonth int) ([]byte, results.Result)
	BuildFileName(targetYear int, targetMonth int) string
	BuildExcelFileName(targetYear int, targetMonth int) string
}

/*
 * 月次勤怠集計CSV出力 Builder
 */
type monthlyAttendanceSummaryExportBuilder struct {
	db *gorm.DB
}

/*
 * MonthlyAttendanceSummaryExportBuilder生成
 */
func NewMonthlyAttendanceSummaryExportBuilder(db *gorm.DB) MonthlyAttendanceSummaryExportBuilder {
	return &monthlyAttendanceSummaryExportBuilder{
		db: db,
	}
}

/*
 * CSVファイル名生成
 */
func (builder *monthlyAttendanceSummaryExportBuilder) BuildFileName(targetYear int, targetMonth int) string {
	return fmt.Sprintf("monthly_attendance_summary_%04d_%02d.csv", targetYear, targetMonth)
}

/*
 * Excelファイル名生成
 */
func (builder *monthlyAttendanceSummaryExportBuilder) BuildExcelFileName(targetYear int, targetMonth int) string {
	return fmt.Sprintf("monthly_attendance_summary_%04d_%02d.xlsx", targetYear, targetMonth)
}

/*
 * CSV生成
 *
 * Excelでも文字化けしにくいようにUTF-8 BOMを付与する。
 */
func (builder *monthlyAttendanceSummaryExportBuilder) BuildCSV(
	rows []types.MonthlyAttendanceSummaryCsvRow,
) ([]byte, results.Result) {
	buffer := &bytes.Buffer{}

	// UTF-8 BOM
	buffer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buffer)

	if err := writer.Write(builder.buildHeader()); err != nil {
		return nil, results.BadRequest(
			"BUILD_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_CSV_HEADER_FAILED",
			"月次勤怠集計CSVのヘッダー生成に失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	for _, row := range rows {
		if err := writer.Write(builder.buildRecord(row)); err != nil {
			return nil, results.BadRequest(
				"BUILD_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_CSV_RECORD_FAILED",
				"月次勤怠集計CSVの行生成に失敗しました",
				map[string]any{
					"error": err.Error(),
				},
			)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, results.BadRequest(
			"BUILD_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_CSV_FLUSH_FAILED",
			"月次勤怠集計CSVの書き込みに失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	return buffer.Bytes(), results.OK(
		nil,
		"BUILD_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_CSV_SUCCESS",
		"",
		nil,
	)
}

/*
 * CSVヘッダー生成
 *
 * 経理提出・給与確認に必要な集計結果だけを中心に出力する。
 * 内部ID、権限、給与設定の詳細、計算途中の基準値、boolean警告フラグは出力しない。
 */
func (builder *monthlyAttendanceSummaryExportBuilder) buildHeader() []string {
	return []string{
		"対象年",
		"対象月",
		"出力日時",
		"集計状態",

		"従業員ID",
		"従業員名",
		"部署名",
		"入社日",
		"退職日",
		"対象月退職済み",

		"月次申請状態",
		"申請日時",
		"承認日時",

		"暦日数",
		"勤怠登録日数",
		"勤怠未登録日数",
		"予定出勤日数",
		"実出勤日数",
		"日勤出勤日数",
		"夜勤出勤日数",
		"公休数",
		"有給日数",
		"半日有給回数",
		"欠勤日数",
		"病欠日数",
		"休日出勤日数",
		"遅刻回数",
		"早退回数",
		"予定あり実績なし日数",
		"実績あり予定なし日数",
		"予定労働時間未設定日数",
		"平日数",

		"予定労働時間_分",
		"総労働時間_分",
		"日中労働時間_分",
		"夜勤労働時間_分",
		"休憩時間_分",
		"所定内労働時間_分",
		"控除対象不足時間_分",
		"総残業時間_分",
		"日中残業時間_分",
		"夜勤残業時間_分",
		"深夜労働時間_分",
		"休日労働時間_分",
		"有給換算時間_分",
		"欠勤控除時間_分",
		"病欠控除時間_分",
		"遅刻控除時間_分",
		"早退控除時間_分",

		"実労働稼働率_％",

		"日別交通費合計",
		"月次定期代",
		"交通費合計",
		"日別交通費登録回数",

		"有給使用日数",
		"有給使用換算時間_分",

		"経費合計",
		"経費件数",
		"交通費系経費",
		"備品系経費",
		"通信費系経費",
		"その他経費",

		"警告件数",
		"警告内容",
		"休憩不整合件数",
		"時刻不整合件数",
	}
}

/*
 * CSVレコード生成
 */
func (builder *monthlyAttendanceSummaryExportBuilder) buildRecord(row types.MonthlyAttendanceSummaryCsvRow) []string {
	calculated := row.CalculationStatus == types.MonthlyAttendanceSummaryCalculationStatusCalculated

	return []string{
		strconv.Itoa(row.ExportTargetYear),
		strconv.Itoa(row.ExportTargetMonth),
		row.ExportedAt,
		calculationStatusLabel(row.CalculationStatus),

		uintToString(row.UserID),
		row.UserName,
		row.DepartmentName,
		row.HireDate,
		row.RetirementDate,
		boolToJapanese(row.IsRetiredInTargetMonth),

		monthlyStatusLabel(row.MonthlyStatus),
		row.RequestedAt,
		row.ApprovedAt,

		calcIntToString(calculated, row.CalendarDays),
		calcIntToString(calculated, row.RegisteredAttendanceDays),
		calcIntToString(calculated, row.MissingAttendanceDays),
		calcIntToString(calculated, row.ScheduledWorkDays),
		calcIntToString(calculated, row.ActualWorkDays),
		calcIntToString(calculated, row.DayShiftWorkDays),
		calcIntToString(calculated, row.NightShiftWorkDays),
		calcIntToString(calculated, row.PlannedHolidayDays),
		calcIntToString(calculated, row.PaidLeaveDays),
		calcIntToString(calculated, row.HalfPaidLeaveDays),
		calcIntToString(calculated, row.AbsenceDays),
		calcIntToString(calculated, row.SickLeaveDays),
		calcIntToString(calculated, row.HolidayWorkDays),
		calcIntToString(calculated, row.LateDays),
		calcIntToString(calculated, row.EarlyLeaveDays),
		calcIntToString(calculated, row.ScheduledButNoActualDays),
		calcIntToString(calculated, row.ActualButNoScheduledDays),
		calcIntToString(calculated, row.MissingScheduledWorkDays),
		calcIntToString(calculated, row.WorkingDayCount),

		calcIntToString(calculated, row.ScheduledWorkMinutes),
		calcIntToString(calculated, row.ActualWorkMinutes),
		calcIntToString(calculated, row.DayWorkMinutes),
		calcIntToString(calculated, row.NightWorkMinutes),
		calcIntToString(calculated, row.BreakMinutes),
		calcIntToString(calculated, row.RegularWorkMinutes),
		calcIntToString(calculated, row.WorkShortageMinutes),
		calcIntToString(calculated, row.OvertimeMinutes),
		calcIntToString(calculated, row.DayOvertimeMinutes),
		calcIntToString(calculated, row.NightOvertimeMinutes),
		calcIntToString(calculated, row.LateNightWorkMinutes),
		calcIntToString(calculated, row.HolidayWorkMinutes),
		calcIntToString(calculated, row.PaidLeaveMinutes),
		calcIntToString(calculated, row.AbsenceMinutes),
		calcIntToString(calculated, row.SickLeaveMinutes),
		calcIntToString(calculated, row.LateMinutes),
		calcIntToString(calculated, row.EarlyLeaveMinutes),

		calcFloatToString(calculated, row.ActualOperationRate),

		calcIntToString(calculated, row.DailyTransportationAmount),
		calcIntToString(calculated, row.CommuterPassAmount),
		calcIntToString(calculated, row.TotalTransportationAmount),
		calcIntToString(calculated, row.DailyTransportationCount),

		calcFloatToString(calculated, row.PaidLeaveUsedDays),
		calcIntToString(calculated, row.PaidLeaveUsedMinutes),

		calcIntToString(calculated, row.ExpenseTotalAmount),
		calcIntToString(calculated, row.ExpenseCount),
		calcIntToString(calculated, row.TransportationExpenseAmount),
		calcIntToString(calculated, row.SuppliesExpenseAmount),
		calcIntToString(calculated, row.CommunicationExpenseAmount),
		calcIntToString(calculated, row.OtherExpenseAmount),

		intToString(row.WarningCount),
		row.Warnings,
		intToString(row.InvalidBreakCount),
		intToString(row.InvalidTimeCount),
	}
}

/*
 * Excel生成
 *
 * CSVと同じ集計行を使い、提出用に見やすい表として出力する。
 * 外部ライブラリを増やさないため、xlsxの最小構成を標準ライブラリで生成する。
 */
func (builder *monthlyAttendanceSummaryExportBuilder) BuildExcel(
	rows []types.MonthlyAttendanceSummaryCsvRow,
	targetYear int,
	targetMonth int,
) ([]byte, results.Result) {
	header := builder.buildHeader()
	records := make([][]string, 0, len(rows))
	for _, row := range rows {
		records = append(records, builder.buildRecord(row))
	}

	buffer := &bytes.Buffer{}
	zipWriter := zip.NewWriter(buffer)

	exportedAt := ""
	if len(rows) > 0 {
		exportedAt = rows[0].ExportedAt
	}

	files := map[string]string{
		"[Content_Types].xml":        buildExcelContentTypesXML(),
		"_rels/.rels":                buildExcelRootRelsXML(),
		"docProps/app.xml":           buildExcelAppXML(),
		"docProps/core.xml":          buildExcelCoreXML(),
		"xl/workbook.xml":            buildExcelWorkbookXML(),
		"xl/_rels/workbook.xml.rels": buildExcelWorkbookRelsXML(),
		"xl/styles.xml":              buildExcelStylesXML(),
		"xl/worksheets/sheet1.xml":   buildExcelWorksheetXML(header, records, targetYear, targetMonth, exportedAt),
	}

	fileNames := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"docProps/app.xml",
		"docProps/core.xml",
		"xl/workbook.xml",
		"xl/_rels/workbook.xml.rels",
		"xl/styles.xml",
		"xl/worksheets/sheet1.xml",
	}

	for _, fileName := range fileNames {
		if err := writeExcelZipFile(zipWriter, fileName, files[fileName]); err != nil {
			return nil, results.BadRequest(
				"BUILD_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_EXCEL_FILE_FAILED",
				"月次勤怠集計Excelのファイル生成に失敗しました",
				map[string]any{
					"fileName": fileName,
					"error":    err.Error(),
				},
			)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, results.BadRequest(
			"BUILD_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_EXCEL_CLOSE_FAILED",
			"月次勤怠集計Excelの書き込み完了に失敗しました",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	return buffer.Bytes(), results.OK(
		nil,
		"BUILD_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_EXCEL_SUCCESS",
		"",
		nil,
	)
}

func writeExcelZipFile(zipWriter *zip.Writer, fileName string, content string) error {
	writer, err := zipWriter.Create(fileName)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(content))
	return err
}

func buildExcelWorksheetXML(
	header []string,
	records [][]string,
	targetYear int,
	targetMonth int,
	exportedAt string,
) string {
	lastColumn := excelColumnName(len(header))
	headerRowNumber := 4
	dataStartRowNumber := 5
	dataEndRowNumber := dataStartRowNumber + len(records) - 1
	totalRowNumber := dataStartRowNumber + len(records)
	if len(records) == 0 {
		dataEndRowNumber = headerRowNumber
		totalRowNumber = headerRowNumber + 1
	}

	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	builder.WriteString(`<sheetViews><sheetView workbookViewId="0"><pane ySplit="4" topLeftCell="A5" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>`)
	builder.WriteString(buildExcelColumnsXML(header, records))
	builder.WriteString(`<sheetData>`)

	builder.WriteString(`<row r="1" ht="24" customHeight="1">`)
	builder.WriteString(buildExcelStringCell("A1", fmt.Sprintf("%04d年%02d月 月次勤怠集計表", targetYear, targetMonth), 1))
	builder.WriteString(`</row>`)

	builder.WriteString(`<row r="2">`)
	builder.WriteString(buildExcelStringCell("A2", "対象年月", 2))
	builder.WriteString(buildExcelStringCell("B2", fmt.Sprintf("%04d年%02d月", targetYear, targetMonth), 2))
	builder.WriteString(buildExcelStringCell("C2", "出力日時", 2))
	builder.WriteString(buildExcelStringCell("D2", exportedAt, 2))
	builder.WriteString(`</row>`)

	builder.WriteString(`<row r="4">`)
	for index, headerName := range header {
		ref := fmt.Sprintf("%s%d", excelColumnName(index+1), headerRowNumber)
		builder.WriteString(buildExcelStringCell(ref, headerName, 3))
	}
	builder.WriteString(`</row>`)

	for rowIndex, record := range records {
		rowNumber := dataStartRowNumber + rowIndex
		warningRow := isExcelWarningRecord(header, record)
		alternateRow := rowIndex%2 == 1

		builder.WriteString(fmt.Sprintf(`<row r="%d">`, rowNumber))
		for columnIndex, value := range record {
			ref := fmt.Sprintf("%s%d", excelColumnName(columnIndex+1), rowNumber)
			headerName := header[columnIndex]
			styleID := excelBodyStyleID(headerName, warningRow, alternateRow)
			builder.WriteString(buildExcelCell(ref, value, styleID, isExcelNumericHeader(headerName)))
		}
		builder.WriteString(`</row>`)
	}

	totalRecord := buildExcelTotalRecord(header, records)
	builder.WriteString(fmt.Sprintf(`<row r="%d">`, totalRowNumber))
	for columnIndex, value := range totalRecord {
		ref := fmt.Sprintf("%s%d", excelColumnName(columnIndex+1), totalRowNumber)
		headerName := header[columnIndex]
		styleID := 10
		if isExcelNumericHeader(headerName) {
			styleID = 11
		}
		builder.WriteString(buildExcelCell(ref, value, styleID, isExcelNumericHeader(headerName)))
	}
	builder.WriteString(`</row>`)

	builder.WriteString(`</sheetData>`)
	builder.WriteString(fmt.Sprintf(`<autoFilter ref="A4:%s%d"/>`, lastColumn, maxIntForExcel(dataEndRowNumber, headerRowNumber)))
	builder.WriteString(fmt.Sprintf(`<mergeCells count="1"><mergeCell ref="A1:%s1"/></mergeCells>`, lastColumn))
	builder.WriteString(`<pageMargins left="0.7" right="0.7" top="0.75" bottom="0.75" header="0.3" footer="0.3"/>`)
	builder.WriteString(`</worksheet>`)

	return builder.String()
}

func buildExcelColumnsXML(header []string, records [][]string) string {
	var builder strings.Builder
	builder.WriteString(`<cols>`)

	for index, headerName := range header {
		maxWidth := excelTextWidth(headerName) + 2
		for _, record := range records {
			if index >= len(record) {
				continue
			}

			width := excelTextWidth(record[index]) + 2
			if width > maxWidth {
				maxWidth = width
			}
		}

		if maxWidth < 10 {
			maxWidth = 10
		}
		if maxWidth > 36 {
			maxWidth = 36
		}

		columnNumber := index + 1
		builder.WriteString(fmt.Sprintf(`<col min="%d" max="%d" width="%d" customWidth="1"/>`, columnNumber, columnNumber, maxWidth))
	}

	builder.WriteString(`</cols>`)
	return builder.String()
}

func buildExcelCell(ref string, value string, styleID int, numeric bool) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return fmt.Sprintf(`<c r="%s" s="%d"/>`, ref, styleID)
	}

	if numeric {
		if _, err := strconv.ParseFloat(trimmedValue, 64); err == nil {
			return fmt.Sprintf(`<c r="%s" s="%d"><v>%s</v></c>`, ref, styleID, trimmedValue)
		}
	}

	return buildExcelStringCell(ref, value, styleID)
}

func buildExcelStringCell(ref string, value string, styleID int) string {
	return fmt.Sprintf(`<c r="%s" s="%d" t="inlineStr"><is><t>%s</t></is></c>`, ref, styleID, escapeExcelXML(value))
}

func buildExcelTotalRecord(header []string, records [][]string) []string {
	totalRecord := make([]string, len(header))
	if len(header) == 0 {
		return totalRecord
	}

	for columnIndex, headerName := range header {
		if columnIndex == 5 {
			totalRecord[columnIndex] = "合計"
			continue
		}

		if !isExcelSummableHeader(headerName) {
			continue
		}

		sum := 0.0
		hasValue := false
		for _, record := range records {
			if columnIndex >= len(record) {
				continue
			}

			value := strings.TrimSpace(record[columnIndex])
			if value == "" {
				continue
			}

			parsedValue, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}

			sum += parsedValue
			hasValue = true
		}

		if !hasValue {
			continue
		}

		if sum == float64(int64(sum)) {
			totalRecord[columnIndex] = strconv.FormatInt(int64(sum), 10)
		} else {
			totalRecord[columnIndex] = strconv.FormatFloat(sum, 'f', 1, 64)
		}
	}

	return totalRecord
}

func isExcelWarningRecord(header []string, record []string) bool {
	for index, headerName := range header {
		if headerName != "警告件数" || index >= len(record) {
			continue
		}

		warningCount, err := strconv.Atoi(strings.TrimSpace(record[index]))
		return err == nil && warningCount > 0
	}

	return false
}

func excelBodyStyleID(headerName string, warningRow bool, alternateRow bool) int {
	if warningRow {
		if isExcelNumericHeader(headerName) {
			return 9
		}
		return 6
	}

	if isExcelNumericHeader(headerName) {
		if alternateRow {
			return 8
		}
		return 7
	}

	if alternateRow {
		return 5
	}

	return 4
}

func isExcelNumericHeader(headerName string) bool {
	textHeaders := map[string]bool{
		"出力日時":    true,
		"集計状態":    true,
		"従業員名":    true,
		"部署名":     true,
		"入社日":     true,
		"退職日":     true,
		"対象月退職済み": true,
		"月次申請状態":  true,
		"申請日時":    true,
		"承認日時":    true,
		"警告内容":    true,
	}

	return !textHeaders[headerName]
}

func isExcelSummableHeader(headerName string) bool {
	if strings.Contains(headerName, "率") || strings.Contains(headerName, "判定") {
		return false
	}

	nonSummableHeaders := map[string]bool{
		"対象年":   true,
		"対象月":   true,
		"従業員ID": true,
		"暦日数":   true,
		"平日数":   true,
	}
	if nonSummableHeaders[headerName] {
		return false
	}

	summableWords := []string{"日数", "回数", "_分", "合計", "金額", "件数", "控除"}
	for _, word := range summableWords {
		if strings.Contains(headerName, word) {
			return true
		}
	}

	return false
}

func excelColumnName(columnNumber int) string {
	name := ""
	for columnNumber > 0 {
		columnNumber--
		name = string(rune('A'+columnNumber%26)) + name
		columnNumber /= 26
	}

	return name
}

func excelTextWidth(value string) int {
	width := 0
	for _, r := range value {
		if r <= 127 {
			width++
		} else {
			width += 2
		}
	}

	return width
}

func escapeExcelXML(value string) string {
	buffer := &bytes.Buffer{}
	_ = xml.EscapeText(buffer, []byte(value))
	return buffer.String()
}

func maxIntForExcel(a int, b int) int {
	if a > b {
		return a
	}

	return b
}

func buildExcelContentTypesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>
<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`
}

func buildExcelRootRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`
}

func buildExcelWorkbookXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<sheets><sheet name="月次勤怠集計" sheetId="1" r:id="rId1"/></sheets>
</workbook>`
}

func buildExcelWorkbookRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`
}

func buildExcelAppXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">
<Application>Timexeed</Application>
</Properties>`
}

func buildExcelCoreXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
<dc:creator>Timexeed</dc:creator>
<cp:lastModifiedBy>Timexeed</cp:lastModifiedBy>
</cp:coreProperties>`
}

func buildExcelStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<fonts count="4">
<font><sz val="11"/><color rgb="FF000000"/><name val="Yu Gothic"/></font>
<font><b/><sz val="16"/><color rgb="FF1F2937"/><name val="Yu Gothic"/></font>
<font><b/><sz val="11"/><color rgb="FFFFFFFF"/><name val="Yu Gothic"/></font>
<font><b/><sz val="11"/><color rgb="FF000000"/><name val="Yu Gothic"/></font>
</fonts>
<fills count="7">
<fill><patternFill patternType="none"/></fill>
<fill><patternFill patternType="gray125"/></fill>
<fill><patternFill patternType="solid"><fgColor rgb="FF1F4E78"/><bgColor indexed="64"/></patternFill></fill>
<fill><patternFill patternType="solid"><fgColor rgb="FFF7F7F7"/><bgColor indexed="64"/></patternFill></fill>
<fill><patternFill patternType="solid"><fgColor rgb="FFEAF2F8"/><bgColor indexed="64"/></patternFill></fill>
<fill><patternFill patternType="solid"><fgColor rgb="FFFFE5E5"/><bgColor indexed="64"/></patternFill></fill>
<fill><patternFill patternType="solid"><fgColor rgb="FFE2F0D9"/><bgColor indexed="64"/></patternFill></fill>
</fills>
<borders count="2">
<border><left/><right/><top/><bottom/><diagonal/></border>
<border><left style="thin"><color rgb="FFD9D9D9"/></left><right style="thin"><color rgb="FFD9D9D9"/></right><top style="thin"><color rgb="FFD9D9D9"/></top><bottom style="thin"><color rgb="FFD9D9D9"/></bottom><diagonal/></border>
</borders>
<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
<cellXfs count="12">
<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
<xf numFmtId="0" fontId="1" fillId="0" borderId="0" xfId="0" applyFont="1"/>
<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
<xf numFmtId="0" fontId="2" fillId="2" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
<xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1"><alignment vertical="center" wrapText="1"/></xf>
<xf numFmtId="0" fontId="0" fillId="3" borderId="1" xfId="0" applyFill="1" applyBorder="1"><alignment vertical="center" wrapText="1"/></xf>
<xf numFmtId="0" fontId="0" fillId="5" borderId="1" xfId="0" applyFill="1" applyBorder="1"><alignment vertical="center" wrapText="1"/></xf>
<xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1"><alignment horizontal="right" vertical="center"/></xf>
<xf numFmtId="0" fontId="0" fillId="3" borderId="1" xfId="0" applyFill="1" applyBorder="1"><alignment horizontal="right" vertical="center"/></xf>
<xf numFmtId="0" fontId="0" fillId="5" borderId="1" xfId="0" applyFill="1" applyBorder="1"><alignment horizontal="right" vertical="center"/></xf>
<xf numFmtId="0" fontId="3" fillId="6" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"><alignment vertical="center" wrapText="1"/></xf>
<xf numFmtId="0" fontId="3" fillId="6" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"><alignment horizontal="right" vertical="center"/></xf>
</cellXfs>
<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`
}

func calculationStatusLabel(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "CALCULATED":
		return "集計済み"
	case "SKIPPED_NOT_APPROVED":
		return "未承認のため集計対象外"
	case "SKIPPED":
		return "集計対象外"
	case "ERROR":
		return "集計エラー"
	default:
		return status
	}
}

func monthlyStatusLabel(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "NONE", "NOT_SUBMITTED":
		return "未申請"
	case "DRAFT":
		return "下書き"
	case "PENDING":
		return "申請中"
	case "APPROVED":
		return "承認済み"
	case "REJECTED":
		return "差し戻し"
	case "CANCELED", "CANCELLED", "WITHDRAWN":
		return "取り下げ"
	default:
		return status
	}
}

func boolToJapanese(value bool) string {
	if value {
		return "はい"
	}

	return "いいえ"
}

func calcString(calculated bool, value string) string {
	if !calculated {
		return ""
	}

	return value
}

func calcIntToString(calculated bool, value int) string {
	if !calculated {
		return ""
	}

	return intToString(value)
}

func calcFloatToString(calculated bool, value float64) string {
	if !calculated {
		return ""
	}

	return floatToString(value)
}

func calcBoolToString(calculated bool, value bool) string {
	if !calculated {
		return ""
	}

	return boolToString(value)
}

func uintToString(value uint) string {
	if value == 0 {
		return ""
	}

	return strconv.FormatUint(uint64(value), 10)
}

func intToString(value int) string {
	return strconv.Itoa(value)
}

func floatToString(value float64) string {
	return strconv.FormatFloat(value, 'f', 1, 64)
}

func boolToString(value bool) string {
	if value {
		return "true"
	}

	return "false"
}
