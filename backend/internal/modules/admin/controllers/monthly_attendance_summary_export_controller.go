package controllers

import (
	"fmt"
	"net/http"

	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 月次勤怠集計ファイル出力 Controller
 *
 * 管理者専用。
 *
 * 注意：
 * ・給与計算そのものは行わない
 * ・APPROVED の月だけ勤怠/給与/交通費/有給/経費の集計値を出力する
 * ・APPROVED 以外はステータスと警告のみ出力する
 * ・format が XLSX の場合は見た目付きExcelを返す
 * ・format 未指定または CSV の場合は従来通りCSVを返す
 * ・ファイルを返すため、正常時は responses.JSON ではなく c.Data で返す
 */
type MonthlyAttendanceSummaryExportController struct {
	monthlyAttendanceSummaryExportService services.MonthlyAttendanceSummaryExportService
}

/*
 * MonthlyAttendanceSummaryExportController生成
 */
func NewMonthlyAttendanceSummaryExportController(
	monthlyAttendanceSummaryExportService services.MonthlyAttendanceSummaryExportService,
) *MonthlyAttendanceSummaryExportController {
	return &MonthlyAttendanceSummaryExportController{
		monthlyAttendanceSummaryExportService: monthlyAttendanceSummaryExportService,
	}
}

/*
 * 月次勤怠集計ファイル出力
 *
 * POST /admin/monthly-attendance-summary-exports/export
 */
func (controller *MonthlyAttendanceSummaryExportController) ExportMonthlyAttendanceSummaryCsv(c *gin.Context) {
	var request types.ExportMonthlyAttendanceSummaryCsvRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		responses.JSON(c, results.BadRequest(
			"INVALID_MONTHLY_ATTENDANCE_SUMMARY_EXPORT_REQUEST",
			"月次勤怠集計出力リクエストの形式が正しくありません",
			map[string]any{
				"error": err.Error(),
			},
		))
		return
	}

	fileBytes, fileName, contentType, result := controller.monthlyAttendanceSummaryExportService.ExportMonthlyAttendanceSummaryFile(request)
	if result.Error {
		responses.JSON(c, result)
		return
	}

	if fileName == "" {
		fileName = fmt.Sprintf("monthly_attendance_summary_%04d_%02d.csv", request.TargetYear, request.TargetMonth)
	}

	if contentType == "" {
		contentType = "text/csv; charset=utf-8"
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, fileBytes)
}
