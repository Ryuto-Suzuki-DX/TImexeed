package controllers

import (
	"timexeed/backend/internal/modules/user/services"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/responses"
	"timexeed/backend/internal/results"

	"github.com/gin-gonic/gin"
)

/*
 * 従業員用月次通勤定期Controller
 *
 * 役割：
 * ・リクエストJSONをRequest型にbindする
 * ・AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
 * ・bind失敗時はControllerでcode/message/detailsを作って返す
 * ・Serviceを呼び出す
 * ・Service結果を共通レスポンス形式で返す
 *
 * このControllerで扱うもの：
 * ・従業員本人の対象月の月次通勤定期検索
 *
 * このControllerで扱わないもの：
 * ・月次通勤定期の単体更新
 * ・月次通勤定期の単体削除
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
 * ・従業員APIでは userId / targetUserId を request body で受け取らない
 *
 * 状態管理方針：
 * ・MonthlyCommuterPass は月次通勤定期データだけを管理する
 * ・月次申請状態は MonthlyAttendanceRequest で管理する
 * ・このControllerでは状態を直接判定しない
 * ・編集可否は Service 側、または monthly_attendances/update 側で判定する
 *
 * 保存方針：
 * ・月次通勤定期の保存は monthly_attendances/update の全体保存から行う
 * ・そのため、このControllerには単体更新・削除APIを用意しない
 *
 * エラー方針：
 * ・Controllerで発生したエラーはControllerでcode/messageを決める
 * ・Serviceで発生したエラーはServiceでcode/messageを決める
 * ・Builderで発生したエラーはBuilderでcode/messageを決める
 * ・Repositoryで発生したエラーはRepositoryでcode/messageを決める
 * ・Controllerは最終的に responses.JSON で返す
 */
type MonthlyCommuterPassController struct {
	monthlyCommuterPassService services.MonthlyCommuterPassService
}

/*
 * MonthlyCommuterPassController生成
 */
func NewMonthlyCommuterPassController(
	monthlyCommuterPassService services.MonthlyCommuterPassService,
) *MonthlyCommuterPassController {
	return &MonthlyCommuterPassController{
		monthlyCommuterPassService: monthlyCommuterPassService,
	}
}

/*
 * 月次通勤定期検索
 *
 * POST /user/monthly-commuter-passes/search
 *
 * 用途：
 * ・従業員本人の対象月の月次通勤定期を取得する
 * ・月次勤怠画面に表示する
 *
 * 仕様：
 * ・フロントから userId / targetUserId は送らない
 * ・AuthMiddlewareでJWTから取得した userId を使う
 * ・ログイン中ユーザー本人の月次通勤定期だけを取得する
 */
func (controller *MonthlyCommuterPassController) SearchMonthlyCommuterPass(c *gin.Context) {
	var req types.SearchMonthlyCommuterPassRequest

	// AuthMiddlewareでJWTから取得したログイン中ユーザーIDを取得する
	userIDValue, exists := c.Get("userId")
	if !exists {
		responses.JSON(c, results.Unauthorized(
			"SEARCH_MONTHLY_COMMUTER_PASS_USER_ID_NOT_FOUND",
			"認証情報からユーザーIDを取得できません",
			nil,
		))
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		responses.JSON(c, results.Unauthorized(
			"SEARCH_MONTHLY_COMMUTER_PASS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		))
		return
	}

	// リクエストJSONをSearchMonthlyCommuterPassRequest型にbindする
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSON(c, results.BadRequest(
			"SEARCH_MONTHLY_COMMUTER_PASS_INVALID_REQUEST",
			"月次通勤定期検索のリクエスト形式が正しくありません",
			err.Error(),
		))
		return
	}

	// bindしたRequest型とログイン中ユーザーIDをServiceへ渡す
	result := controller.monthlyCommuterPassService.SearchMonthlyCommuterPass(userID, req)

	// Service / Builder / Repository の結果を共通レスポンス形式のJSONでフロントへ返す
	responses.JSON(c, result)
}
