package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用勤務区分マスタController
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * 注意：
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・Requestを別の型へ詰め直さない
 * ・c.JSONは直接使わず responses.JSON を使う
 *
 * エラー方針：
 * ・Controllerで発生したエラーはControllerでcode/messageを決める
 * ・Serviceで発生したエラーはServiceでcode/messageを決める
 * ・Builderで発生したエラーはBuilderでcode/messageを決める
 * ・Repositoryで発生したエラーはRepositoryでcode/messageを決める
 * ・Controllerは最終的に responses.JSON で返す
 *
 * このControllerで扱うもの：
 * ・勤務区分マスタの検索
 *
 * ユーザー側では勤務区分マスタの作成・更新・削除はしない。
 * 勤怠入力画面で使用する選択肢を取得するだけ。
 */
type AttendanceTypeController struct {
	attendanceTypeService services.AttendanceTypeService
}

/*
 * AttendanceTypeController生成
 */
func NewAttendanceTypeController(attendanceTypeService services.AttendanceTypeService) *AttendanceTypeController {
	return &AttendanceTypeController{
		attendanceTypeService: attendanceTypeService,
	}
}

/*
 * 勤務区分マスタ検索
 *
 * POST /user/attendance-types/search
 *
 * 用途：
 * ・勤怠入力画面で勤務区分の選択肢を取得する
 * ・フロント側で入力欄の表示制御に使う
 *
 * フロントが見る主な項目：
 * ・syncPlanActual
 * ・allowActualTimeInput
 * ・allowBreakInput
 * ・allowTransportInput
 * ・allowLateFlag
 * ・allowEarlyLeaveFlag
 * ・allowAbsenceFlag
 * ・allowSickLeaveFlag
 * ・requiresRequest
 */
func (controller *AttendanceTypeController) SearchAttendanceTypes(c *gin.Context) {
	var req types.SearchAttendanceTypesRequest

	// リクエストJSONをSearchAttendanceTypesRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		// Controllerで発生したbindエラーなので、Controller用のcode/messageを詰めて返す
		responses.JSON(c, results.BadRequest(
			"SEARCH_ATTENDANCE_TYPES_INVALID_REQUEST",
			"勤務区分マスタ検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.attendanceTypeService.SearchAttendanceTypes(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
