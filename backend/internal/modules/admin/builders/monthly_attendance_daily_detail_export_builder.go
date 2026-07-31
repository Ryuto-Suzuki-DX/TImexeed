package builders

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
)

type dailyDetailSheetDefinition struct {
	Name string
	Rows [][]dailyDetailCell
}

type dailyDetailCell struct {
	Value   string
	StyleID int
	Number  bool
	Formula string
}

func (builder *monthlyAttendanceSummaryExportBuilder) BuildDailyDetailExcelFileName(
	targetYear int,
	targetMonth int,
) string {
	return fmt.Sprintf(
		"%04d年%02d月_全従業員_日別勤怠明細_%s.xlsx",
		targetYear,
		targetMonth,
		time.Now().Format("20060102_150405"),
	)
}

func (builder *monthlyAttendanceSummaryExportBuilder) BuildDailyDetailExcel(
	userSheets []types.MonthlyAttendanceDailyDetailUserSheet,
	targetYear int,
	targetMonth int,
) ([]byte, results.Result) {
	if len(userSheets) == 0 {
		return nil, results.BadRequest(
			"BUILD_DAILY_ATTENDANCE_DETAIL_EXCEL_EMPTY",
			"日別勤怠明細Excelの出力対象がありません",
			nil,
		)
	}

	sheets := make([]dailyDetailSheetDefinition, 0, len(userSheets)+1)
	sheets = append(sheets, builder.buildDailyDetailSummarySheet(userSheets, targetYear, targetMonth))

	usedSheetNames := map[string]bool{"全体一覧": true}
	for _, userSheet := range userSheets {
		sheetName := uniqueDailyDetailSheetName(userSheet.UserName, userSheet.UserID, usedSheetNames)
		usedSheetNames[sheetName] = true
		sheets = append(sheets, builder.buildDailyDetailUserSheet(sheetName, userSheet, targetYear, targetMonth))
	}

	buffer := &bytes.Buffer{}
	zipWriter := zip.NewWriter(buffer)

	files := map[string]string{
		"[Content_Types].xml":        buildDailyDetailContentTypesXML(len(sheets)),
		"_rels/.rels":                buildDailyDetailRootRelsXML(),
		"xl/workbook.xml":            buildDailyDetailWorkbookXML(sheets),
		"xl/_rels/workbook.xml.rels": buildDailyDetailWorkbookRelsXML(len(sheets)),
		"xl/styles.xml":              buildDailyDetailStylesXML(),
		"docProps/app.xml":           buildDailyDetailAppXML(sheets),
		"docProps/core.xml":          buildDailyDetailCoreXML(),
	}

	for index, sheet := range sheets {
		files[fmt.Sprintf("xl/worksheets/sheet%d.xml", index+1)] =
			buildDailyDetailWorksheetXML(sheet)
	}

	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		writer, err := zipWriter.Create(path)
		if err != nil {
			_ = zipWriter.Close()
			return nil, results.BadRequest(
				"BUILD_DAILY_ATTENDANCE_DETAIL_EXCEL_ZIP_ENTRY_FAILED",
				"日別勤怠明細Excelの生成に失敗しました",
				map[string]any{"path": path, "error": err.Error()},
			)
		}
		if _, err := writer.Write([]byte(files[path])); err != nil {
			_ = zipWriter.Close()
			return nil, results.BadRequest(
				"BUILD_DAILY_ATTENDANCE_DETAIL_EXCEL_WRITE_FAILED",
				"日別勤怠明細Excelの書き込みに失敗しました",
				map[string]any{"path": path, "error": err.Error()},
			)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, results.BadRequest(
			"BUILD_DAILY_ATTENDANCE_DETAIL_EXCEL_CLOSE_FAILED",
			"日別勤怠明細Excelの終了処理に失敗しました",
			map[string]any{"error": err.Error()},
		)
	}

	return buffer.Bytes(), results.OK(nil, "BUILD_DAILY_ATTENDANCE_DETAIL_EXCEL_SUCCESS", "", nil)
}

func (builder *monthlyAttendanceSummaryExportBuilder) buildDailyDetailSummarySheet(
	userSheets []types.MonthlyAttendanceDailyDetailUserSheet,
	targetYear int,
	targetMonth int,
) dailyDetailSheetDefinition {
	rows := [][]dailyDetailCell{
		{{Value: fmt.Sprintf("%04d年%02d月 全従業員 日別勤怠明細", targetYear, targetMonth), StyleID: 1}},
		{{Value: "氏名", StyleID: 2}, {Value: "所属", StyleID: 2}, {Value: "承認状態", StyleID: 2},
			{Value: "勤務日数", StyleID: 2}, {Value: "実働時間", StyleID: 2}, {Value: "休憩時間", StyleID: 2},
			{Value: "残業時間", StyleID: 2}, {Value: "深夜時間", StyleID: 2}, {Value: "休日労働", StyleID: 2},
			{Value: "遅刻", StyleID: 2}, {Value: "早退", StyleID: 2}, {Value: "交通費", StyleID: 2}},
	}

	for _, userSheet := range userSheets {
		workDays := 0
		actual := 0
		breakMinutes := 0
		overtime := 0
		lateNight := 0
		holiday := 0
		late := 0
		early := 0
		transport := 0
		for _, row := range userSheet.Rows {
			if row.ActualWorkMinutes > 0 {
				workDays++
			}
			actual += row.ActualWorkMinutes
			breakMinutes += row.BreakMinutes
			overtime += row.OvertimeMinutes
			lateNight += row.LateNightWorkMinutes
			holiday += row.HolidayWorkMinutes
			late += row.LateMinutes
			early += row.EarlyLeaveMinutes
			transport += row.TransportAmount
		}
		rows = append(rows, []dailyDetailCell{
			{Value: userSheet.UserName, StyleID: 3},
			{Value: userSheet.DepartmentName, StyleID: 3},
			{Value: monthlyStatusLabel(userSheet.MonthlyStatus), StyleID: dailyDetailStatusStyle(userSheet.MonthlyStatus)},
			{Value: strconv.Itoa(workDays), StyleID: 4, Number: true},
			{Value: excelMinutes(actual), StyleID: 5},
			{Value: excelMinutes(breakMinutes), StyleID: 5},
			{Value: excelMinutes(overtime), StyleID: 5},
			{Value: excelMinutes(lateNight), StyleID: 5},
			{Value: excelMinutes(holiday), StyleID: 5},
			{Value: excelMinutes(late), StyleID: 5},
			{Value: excelMinutes(early), StyleID: 5},
			{Value: strconv.Itoa(transport), StyleID: 6, Number: true},
		})
	}

	return dailyDetailSheetDefinition{Name: "全体一覧", Rows: rows}
}

func (builder *monthlyAttendanceSummaryExportBuilder) buildDailyDetailUserSheet(
	sheetName string,
	userSheet types.MonthlyAttendanceDailyDetailUserSheet,
	targetYear int,
	targetMonth int,
) dailyDetailSheetDefinition {
	rows := [][]dailyDetailCell{
		{{Value: fmt.Sprintf("%s　%04d年%02d月 日別勤怠明細", userSheet.UserName, targetYear, targetMonth), StyleID: 1}},
		{{Value: "所属", StyleID: 7}, {Value: userSheet.DepartmentName, StyleID: 3},
			{Value: "メール", StyleID: 7}, {Value: userSheet.UserEmail, StyleID: 3},
			{Value: "承認状態", StyleID: 7}, {Value: monthlyStatusLabel(userSheet.MonthlyStatus), StyleID: dailyDetailStatusStyle(userSheet.MonthlyStatus)}},
		{},
		{
			{Value: "日付", StyleID: 2}, {Value: "曜", StyleID: 2}, {Value: "予定区分", StyleID: 2},
			{Value: "実績状態", StyleID: 2}, {Value: "予定開始", StyleID: 2}, {Value: "予定終了", StyleID: 2},
			{Value: "実績開始", StyleID: 2}, {Value: "実績終了", StyleID: 2}, {Value: "予定時間", StyleID: 2},
			{Value: "拘束時間", StyleID: 2}, {Value: "休憩合計", StyleID: 2}, {Value: "休憩詳細", StyleID: 2},
			{Value: "実働時間", StyleID: 2}, {Value: "所定内", StyleID: 2}, {Value: "残業", StyleID: 2},
			{Value: "深夜", StyleID: 2}, {Value: "休日労働", StyleID: 2}, {Value: "遅刻", StyleID: 2},
			{Value: "早退", StyleID: 2}, {Value: "交通費", StyleID: 2}, {Value: "警告・備考", StyleID: 2},
		},
	}

	totalScheduled := 0
	totalGross := 0
	totalBreak := 0
	totalActual := 0
	totalRegular := 0
	totalOvertime := 0
	totalLateNight := 0
	totalHoliday := 0
	totalLate := 0
	totalEarly := 0
	totalTransport := 0

	for _, row := range userSheet.Rows {
		bodyStyle := 3
		if row.IsSunday {
			bodyStyle = 9
		} else if row.IsSaturday {
			bodyStyle = 8
		}
		rows = append(rows, []dailyDetailCell{
			{Value: row.WorkDate, StyleID: bodyStyle},
			{Value: row.Weekday, StyleID: bodyStyle},
			{Value: row.PlanAttendanceType, StyleID: bodyStyle},
			{Value: row.ActualWorkStatus, StyleID: bodyStyle},
			{Value: row.PlanStart, StyleID: bodyStyle},
			{Value: row.PlanEnd, StyleID: bodyStyle},
			{Value: row.ActualStart, StyleID: bodyStyle},
			{Value: row.ActualEnd, StyleID: bodyStyle},
			{Value: excelMinutes(row.ScheduledWorkMinutes), StyleID: bodyStyle},
			{Value: excelMinutes(row.GrossWorkMinutes), StyleID: bodyStyle},
			{Value: excelMinutes(row.BreakMinutes), StyleID: bodyStyle},
			{Value: row.BreakDetails, StyleID: bodyStyle},
			{Value: excelMinutes(row.ActualWorkMinutes), StyleID: bodyStyle},
			{Value: excelMinutes(row.RegularWorkMinutes), StyleID: bodyStyle},
			{Value: excelMinutes(row.OvertimeMinutes), StyleID: bodyStyle},
			{Value: excelMinutes(row.LateNightWorkMinutes), StyleID: bodyStyle},
			{Value: excelMinutes(row.HolidayWorkMinutes), StyleID: bodyStyle},
			{Value: excelMinutes(row.LateMinutes), StyleID: bodyStyle},
			{Value: excelMinutes(row.EarlyLeaveMinutes), StyleID: bodyStyle},
			{Value: strconv.Itoa(row.TransportAmount), StyleID: 6, Number: true},
			{Value: row.Warnings, StyleID: bodyStyle},
		})
		totalScheduled += row.ScheduledWorkMinutes
		totalGross += row.GrossWorkMinutes
		totalBreak += row.BreakMinutes
		totalActual += row.ActualWorkMinutes
		totalRegular += row.RegularWorkMinutes
		totalOvertime += row.OvertimeMinutes
		totalLateNight += row.LateNightWorkMinutes
		totalHoliday += row.HolidayWorkMinutes
		totalLate += row.LateMinutes
		totalEarly += row.EarlyLeaveMinutes
		totalTransport += row.TransportAmount
	}

	rows = append(rows,
		[]dailyDetailCell{},
		[]dailyDetailCell{
			{Value: "月合計", StyleID: 10}, {}, {}, {}, {}, {}, {}, {},
			{Value: excelMinutes(totalScheduled), StyleID: 10},
			{Value: excelMinutes(totalGross), StyleID: 10},
			{Value: excelMinutes(totalBreak), StyleID: 10},
			{},
			{Value: excelMinutes(totalActual), StyleID: 10},
			{Value: excelMinutes(totalRegular), StyleID: 10},
			{Value: excelMinutes(totalOvertime), StyleID: 10},
			{Value: excelMinutes(totalLateNight), StyleID: 10},
			{Value: excelMinutes(totalHoliday), StyleID: 10},
			{Value: excelMinutes(totalLate), StyleID: 10},
			{Value: excelMinutes(totalEarly), StyleID: 10},
			{Value: strconv.Itoa(totalTransport), StyleID: 11, Number: true},
			{},
		},
	)

	return dailyDetailSheetDefinition{Name: sheetName, Rows: rows}
}

func excelMinutes(minutes int) string {
	if minutes <= 0 {
		return ""
	}
	return fmt.Sprintf("%d:%02d", minutes/60, minutes%60)
}

func dailyDetailStatusStyle(status string) int {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "APPROVED":
		return 12
	case "REJECTED", "CANCELED":
		return 14
	case "PENDING":
		return 13
	default:
		return 15
	}
}

func uniqueDailyDetailSheetName(name string, userID uint, used map[string]bool) string {
	replacer := strings.NewReplacer("/", "／", "\\", "￥", "?", "？", "*", "＊", "[", "（", "]", "）", ":", "：")
	base := strings.TrimSpace(replacer.Replace(name))
	if base == "" {
		base = fmt.Sprintf("ユーザー%d", userID)
	}
	if len([]rune(base)) > 31 {
		base = string([]rune(base)[:31])
	}
	if !used[base] {
		return base
	}

	suffix := fmt.Sprintf("_%d", userID)
	maxBaseLength := 31 - len([]rune(suffix))
	runes := []rune(base)
	if len(runes) > maxBaseLength {
		runes = runes[:maxBaseLength]
	}
	return string(runes) + suffix
}

func buildDailyDetailWorksheetXML(sheet dailyDetailSheetDefinition) string {
	var builder strings.Builder
	builder.WriteString(xml.Header)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	builder.WriteString(`<sheetViews><sheetView workbookViewId="0"><pane ySplit="4" topLeftCell="A5" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>`)
	builder.WriteString(`<cols>`)
	widths := []float64{12, 5, 16, 14, 10, 10, 10, 10, 11, 11, 11, 32, 11, 11, 11, 11, 11, 10, 10, 12, 42}
	for index, width := range widths {
		builder.WriteString(fmt.Sprintf(`<col min="%d" max="%d" width="%.1f" customWidth="1"/>`, index+1, index+1, width))
	}
	builder.WriteString(`</cols><sheetData>`)

	for rowIndex, row := range sheet.Rows {
		excelRow := rowIndex + 1
		builder.WriteString(fmt.Sprintf(`<row r="%d">`, excelRow))
		for columnIndex, cell := range row {
			if cell.Value == "" && cell.Formula == "" && cell.StyleID == 0 {
				continue
			}
			ref := fmt.Sprintf("%s%d", excelColumnName(columnIndex+1), excelRow)
			if cell.Formula != "" {
				builder.WriteString(fmt.Sprintf(`<c r="%s" s="%d"><f>%s</f></c>`, ref, cell.StyleID, escapeExcelXML(cell.Formula)))
			} else if cell.Number {
				builder.WriteString(fmt.Sprintf(`<c r="%s" s="%d"><v>%s</v></c>`, ref, cell.StyleID, escapeExcelXML(cell.Value)))
			} else {
				builder.WriteString(fmt.Sprintf(`<c r="%s" s="%d" t="inlineStr"><is><t xml:space="preserve">%s</t></is></c>`, ref, cell.StyleID, escapeExcelXML(cell.Value)))
			}
		}
		builder.WriteString(`</row>`)
	}

	builder.WriteString(`</sheetData>`)
	if sheet.Name == "全体一覧" {
		builder.WriteString(`<autoFilter ref="A2:L2"/>`)
	} else {
		builder.WriteString(`<autoFilter ref="A4:U4"/>`)
	}
	builder.WriteString(`</worksheet>`)
	return builder.String()
}

func buildDailyDetailContentTypesXML(sheetCount int) string {
	var builder strings.Builder
	builder.WriteString(xml.Header)
	builder.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	builder.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	builder.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	builder.WriteString(`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	builder.WriteString(`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	for index := 1; index <= sheetCount; index++ {
		builder.WriteString(fmt.Sprintf(`<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, index))
	}
	builder.WriteString(`<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>`)
	builder.WriteString(`<Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`)
	builder.WriteString(`</Types>`)
	return builder.String()
}

func buildDailyDetailRootRelsXML() string {
	return xml.Header + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>` +
		`<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>` +
		`<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>` +
		`</Relationships>`
}

func buildDailyDetailWorkbookXML(sheets []dailyDetailSheetDefinition) string {
	var builder strings.Builder
	builder.WriteString(xml.Header)
	builder.WriteString(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	for index, sheet := range sheets {
		builder.WriteString(fmt.Sprintf(`<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, escapeExcelXML(sheet.Name), index+1, index+1))
	}
	builder.WriteString(`</sheets></workbook>`)
	return builder.String()
}

func buildDailyDetailWorkbookRelsXML(sheetCount int) string {
	var builder strings.Builder
	builder.WriteString(xml.Header)
	builder.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for index := 1; index <= sheetCount; index++ {
		builder.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, index, index))
	}
	builder.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, sheetCount+1))
	builder.WriteString(`</Relationships>`)
	return builder.String()
}

func buildDailyDetailStylesXML() string {
	return xml.Header + `<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
		`<fonts count="4">` +
		`<font><sz val="10"/><name val="Yu Gothic"/></font>` +
		`<font><b/><sz val="15"/><color rgb="FFFFFFFF"/><name val="Yu Gothic"/></font>` +
		`<font><b/><sz val="10"/><color rgb="FFFFFFFF"/><name val="Yu Gothic"/></font>` +
		`<font><b/><sz val="10"/><name val="Yu Gothic"/></font>` +
		`</fonts>` +
		`<fills count="10">` +
		`<fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill>` +
		`<fill><patternFill patternType="solid"><fgColor rgb="FF1F4E78"/></patternFill></fill>` +
		`<fill><patternFill patternType="solid"><fgColor rgb="FF5B9BD5"/></patternFill></fill>` +
		`<fill><patternFill patternType="solid"><fgColor rgb="FFDDEBF7"/></patternFill></fill>` +
		`<fill><patternFill patternType="solid"><fgColor rgb="FFFCE4D6"/></patternFill></fill>` +
		`<fill><patternFill patternType="solid"><fgColor rgb="FFFFE699"/></patternFill></fill>` +
		`<fill><patternFill patternType="solid"><fgColor rgb="FFE2F0D9"/></patternFill></fill>` +
		`<fill><patternFill patternType="solid"><fgColor rgb="FFFFC7CE"/></patternFill></fill>` +
		`<fill><patternFill patternType="solid"><fgColor rgb="FFEDEDED"/></patternFill></fill>` +
		`</fills>` +
		`<borders count="2"><border/><border><left style="thin"><color rgb="FFB7B7B7"/></left><right style="thin"><color rgb="FFB7B7B7"/></right><top style="thin"><color rgb="FFB7B7B7"/></top><bottom style="thin"><color rgb="FFB7B7B7"/></bottom></border></borders>` +
		`<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>` +
		`<cellXfs count="16">` +
		`<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>` +
		`<xf numFmtId="0" fontId="1" fillId="2" borderId="1" xfId="0" applyAlignment="1"><alignment horizontal="left" vertical="center"/></xf>` +
		`<xf numFmtId="0" fontId="2" fillId="3" borderId="1" xfId="0" applyAlignment="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>` +
		`<xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyAlignment="1"><alignment vertical="center" wrapText="1"/></xf>` +
		`<xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyAlignment="1"><alignment horizontal="right" vertical="center"/></xf>` +
		`<xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyAlignment="1"><alignment horizontal="center" vertical="center"/></xf>` +
		`<xf numFmtId="3" fontId="0" fillId="0" borderId="1" xfId="0" applyNumberFormat="1" applyAlignment="1"><alignment horizontal="right"/></xf>` +
		`<xf numFmtId="0" fontId="3" fillId="4" borderId="1" xfId="0"/>` +
		`<xf numFmtId="0" fontId="0" fillId="4" borderId="1" xfId="0"/>` +
		`<xf numFmtId="0" fontId="0" fillId="5" borderId="1" xfId="0"/>` +
		`<xf numFmtId="0" fontId="3" fillId="6" borderId="1" xfId="0" applyAlignment="1"><alignment horizontal="center"/></xf>` +
		`<xf numFmtId="3" fontId="3" fillId="6" borderId="1" xfId="0" applyNumberFormat="1" applyAlignment="1"><alignment horizontal="right"/></xf>` +
		`<xf numFmtId="0" fontId="3" fillId="7" borderId="1" xfId="0"/>` +
		`<xf numFmtId="0" fontId="3" fillId="6" borderId="1" xfId="0"/>` +
		`<xf numFmtId="0" fontId="3" fillId="8" borderId="1" xfId="0"/>` +
		`<xf numFmtId="0" fontId="3" fillId="9" borderId="1" xfId="0"/>` +
		`</cellXfs><cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>` +
		`</styleSheet>`
}

func buildDailyDetailAppXML(sheets []dailyDetailSheetDefinition) string {
	var titles strings.Builder
	titles.WriteString(fmt.Sprintf(`<vt:vector size="%d" baseType="lpstr">`, len(sheets)))
	for _, sheet := range sheets {
		titles.WriteString(fmt.Sprintf(`<vt:lpstr>%s</vt:lpstr>`, escapeExcelXML(sheet.Name)))
	}
	titles.WriteString(`</vt:vector>`)
	return xml.Header + `<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">` +
		`<Application>Timexeed</Application><TitlesOfParts>` + titles.String() + `</TitlesOfParts></Properties>`
}

func buildDailyDetailCoreXML() string {
	now := time.Now().UTC().Format(time.RFC3339)
	return xml.Header + `<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">` +
		`<dc:creator>Timexeed</dc:creator><cp:lastModifiedBy>Timexeed</cp:lastModifiedBy>` +
		`<dcterms:created xsi:type="dcterms:W3CDTF">` + now + `</dcterms:created>` +
		`<dcterms:modified xsi:type="dcterms:W3CDTF">` + now + `</dcterms:modified></cp:coreProperties>`
}
