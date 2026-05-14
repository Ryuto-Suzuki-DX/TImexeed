package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用月次勤怠全体保存Controller
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * このControllerで扱うもの：
 * ・管理者が指定した対象ユーザーの月次勤怠全体保存
 *
 * 保存対象：
 * ・月次通勤定期
 * ・日別勤怠
 * ・日別休憩
 *
 * このControllerで扱わないもの：
 * ・月次勤怠申請
 * ・月次勤怠承認
 * ・月次勤怠否認
 * ・DB処理
 * ・業務ルール
 *
 * 注意：
 * ・DB処理はしない
 * ・業務ルールは書かない
 * ・Requestを別の型へ詰め直さない
 * ・c.JSONは直接使わず responses.JSON を使う
 * ・管理者APIでは対象ユーザーIDを request body の targetUserId で受け取る
 *
 * 管理者編集方針：
 * ・管理者は月次申請状態に関係なく編集できる
 * ・編集ロックはかけない
 * ・月次申請状態による保存制限はService側にも入れない
 *
 * 名前方針：
 * ・これは月次勤怠データそのもののControllerではない
 * ・月次勤怠画面の「全体保存」用Controller
 * ・そのため monthly_attendance_save_controller.go とする
 *
 * エラー方針：
 * ・Controllerで発生したエラーはControllerでcode/messageを決める
 * ・Serviceで発生したエラーはServiceでcode/messageを決める
 * ・Builderで発生したエラーはBuilderでcode/messageを決める
 * ・Repositoryで発生したエラーはRepositoryでcode/messageを決める
 * ・Controllerは最終的に responses.JSON で返す
 */
type MonthlyAttendanceSaveController struct {
	monthlyAttendanceSaveService services.MonthlyAttendanceSaveService
}

/*
 * MonthlyAttendanceSaveController生成
 */
func NewMonthlyAttendanceSaveController(
	monthlyAttendanceSaveService services.MonthlyAttendanceSaveService,
) *MonthlyAttendanceSaveController {
	return &MonthlyAttendanceSaveController{
		monthlyAttendanceSaveService: monthlyAttendanceSaveService,
	}
}

/*
 * 月次勤怠全体保存
 *
 * POST /admin/monthly-attendance-saves/update
 *
 * 用途：
 * ・管理者が指定した対象ユーザーの対象月の勤怠をまとめて保存する
 * ・月次通勤定期、日別勤怠、日別休憩を一括保存する
 *
 * 仕様：
 * ・対象ユーザーIDは request body の targetUserId で受け取る
 * ・管理者本人のIDは対象データ保存には使わない
 * ・月次申請状態に関係なく保存可能とする
 */
func (controller *MonthlyAttendanceSaveController) UpdateMonthlyAttendance(c *gin.Context) {
	var req types.UpdateMonthlyAttendanceRequest

	// リクエストJSONをUpdateMonthlyAttendanceRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"UPDATE_MONTHLY_ATTENDANCE_INVALID_REQUEST",
			"月次勤怠全体保存のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.monthlyAttendanceSaveService.UpdateMonthlyAttendance(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
