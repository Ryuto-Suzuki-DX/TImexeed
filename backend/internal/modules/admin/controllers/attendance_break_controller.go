package controllers

import (
	"timexeed/backend/internal/modules/admin/services"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 管理者用休憩Controller
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * このControllerで扱うもの：
 * ・管理者が指定した対象ユーザーの対象日の休憩データ検索
 *
 * このControllerで扱わないもの：
 * ・休憩データの単体作成
 * ・休憩データの単体更新
 * ・休憩データの単体削除
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
 * ・AttendanceBreak は休憩データだけを管理する
 * ・月次申請状態は MonthlyAttendanceRequest で管理する
 * ・このControllerでは状態を直接判定しない
 * ・管理者側では月次申請状態による編集ロックを行わない
 *
 * 保存方針：
 * ・休憩データの保存は monthly_attendances/update の全体保存から行う
 * ・そのため、このControllerには単体作成・更新・削除APIを用意しない
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
type AttendanceBreakController struct {
	attendanceBreakService services.AttendanceBreakService
}

/*
 * AttendanceBreakController生成
 */
func NewAttendanceBreakController(attendanceBreakService services.AttendanceBreakService) *AttendanceBreakController {
	return &AttendanceBreakController{
		attendanceBreakService: attendanceBreakService,
	}
}

/*
 * 休憩検索
 *
 * POST /admin/attendance-breaks/search
 *
 * 用途：
 * ・管理者が指定した対象ユーザーの対象日の休憩データを取得する
 * ・管理者用の月次勤怠編集画面に表示する
 *
 * 仕様：
 * ・対象ユーザーIDは request body の targetUserId で受け取る
 * ・管理者本人のIDは対象データ検索には使わない
 * ・targetUserId + workDate から対象日の勤怠を特定する
 * ・対象日の勤怠に紐づく休憩だけを取得する
 * ・月次申請状態に関係なく編集可能とする
 */
func (controller *AttendanceBreakController) SearchAttendanceBreaks(c *gin.Context) {
	var req types.SearchAttendanceBreaksRequest

	// リクエストJSONをSearchAttendanceBreaksRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_ATTENDANCE_BREAKS_INVALID_REQUEST",
			"休憩検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型をそのままServiceへ渡す
	result := controller.attendanceBreakService.SearchAttendanceBreaks(req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
