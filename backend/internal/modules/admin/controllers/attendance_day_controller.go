package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用勤怠Controller
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * このControllerで扱うもの：
 * ・管理者が指定した対象ユーザーの勤怠日別データの検索
 *
 * このControllerで扱わないもの：
 * ・勤怠日別データの単体更新
 * ・勤怠日別データの単体削除
 * ・月次申請状態の判定
 * ・月次承認状態の判定
 * ・編集可能かどうかの判定
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
 * 状態管理方針：
 * ・AttendanceDay は日別勤怠データだけを管理する
 * ・月次申請状態は MonthlyAttendanceRequest で管理する
 * ・このControllerでは状態を直接判定しない
 * ・月次申請状態は Service 側で response に組み立てる
 *
 * 保存方針：
 * ・勤怠日別データの保存は monthly_attendances/update の全体保存から行う
 * ・そのため、このControllerには単体更新APIを用意しない
 *
 * 管理者編集方針：
 * ・管理者は月次申請状態に関係なく編集できる
 * ・編集ロックはかけない
 *
 * エラー方針：
 * ・Controllerで発生したエラーはControllerでcode/messageを決める
 * ・Serviceで発生したエラーはServiceでcode/messageを決める
 * ・Builderで発生したエラーはBuilderでcode/messageを決める
 * ・Repositoryで発生したエラーはRepositoryでcode/messageを決める
 * ・Controllerは最終的に responses.JSON で返す
 */
type AttendanceDayController struct {
	attendanceDayService services.AttendanceDayService
}

/*
 * AttendanceDayController生成
 */
func NewAttendanceDayController(attendanceDayService services.AttendanceDayService) *AttendanceDayController {
	return &AttendanceDayController{
		attendanceDayService: attendanceDayService,
	}
}

/*
 * 勤怠検索
 *
 * POST /admin/attendance-days/search
 *
 * 用途：
 * ・管理者が指定した対象ユーザーの対象月の勤怠日別データを取得する
 * ・管理者用の月次勤怠編集画面に表示する
 * ・月次申請状態もService側で組み立てて返す
 *
 * 仕様：
 * ・対象ユーザーIDは request body の targetUserId で受け取る
 * ・管理者本人のIDは対象データ検索には使わない
 * ・月次申請状態に関係なく編集可能とする
 *
 * 状態管理：
 * ・AttendanceDay 自体は申請状態を持たない
 * ・月次申請状態は MonthlyAttendanceRequest を見て判断する
 * ・対象月の MonthlyAttendanceRequest がなければ未申請扱いにする
 * ・月次申請状態は表示用として response に組み立てる
 */
func (controller *AttendanceDayController) SearchAttendanceDays(c *gin.Context) {
	var req types.SearchAttendanceDaysRequest

	// リクエストJSONをSearchAttendanceDaysRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_ATTENDANCE_DAYS_INVALID_REQUEST",
			"勤怠検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.attendanceDayService.SearchAttendanceDays(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
