package services

import (
	"time"

	"timexeed/backend/internal/constants"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
)

/*
 * 従業員用有給Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type PaidLeaveService interface {
	GetPaidLeaveBalance(userID uint) results.Result
}

/*
 * 従業員用有給Service
 *
 * 役割：
 * ・Controllerから受け取ったログインユーザーIDをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB処理を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type paidLeaveService struct {
	paidLeaveBuilder    builders.PaidLeaveBuilder
	paidLeaveRepository repositories.PaidLeaveRepository
}

/*
 * PaidLeaveService生成
 */
func NewPaidLeaveService(
	paidLeaveBuilder builders.PaidLeaveBuilder,
	paidLeaveRepository repositories.PaidLeaveRepository,
) *paidLeaveService {
	return &paidLeaveService{
		paidLeaveBuilder:    paidLeaveBuilder,
		paidLeaveRepository: paidLeaveRepository,
	}
}

/*
 * 現時点の有給残数取得
 *
 * 計算方針：
 * ・ログイン中ユーザーの入社日を取得する
 * ・固定値ファイルの付与ルールをもとに、現時点までの付与日数を合計する
 * ・有給使用日テーブルの使用日数を合計する
 * ・残数 = 付与合計 - 使用合計
 *
 * 注意：
 * ・現時点では、出勤率8割判定は未実装
 * ・現時点では、有給申請・勤怠承認との連携分はPaidLeaveUsageに入っている前提
 */
func (service *paidLeaveService) GetPaidLeaveBalance(userID uint) results.Result {
	if userID == 0 {
		return results.BadRequest(
			"GET_PAID_LEAVE_BALANCE_INVALID_USER_ID",
			"有給残数取得の対象ユーザーが正しくありません",
			map[string]any{
				"userId": userID,
			},
		)
	}

	now := time.Now()

	// Builderでログインユーザー取得用クエリを作成する
	userQuery, buildUserResult := service.paidLeaveBuilder.BuildFindActiveUserByIDQuery(userID)
	if buildUserResult.Error {
		return buildUserResult
	}

	// Repositoryでログインユーザーを取得する
	user, findUserResult := service.paidLeaveRepository.FindUser(userQuery)
	if findUserResult.Error {
		return findUserResult
	}

	// Builderで有給使用日数合計用クエリを作成する
	usedDaysQuery, buildUsedDaysResult := service.paidLeaveBuilder.BuildSumActivePaidLeaveUsageDaysByUserIDQuery(userID)
	if buildUsedDaysResult.Error {
		return buildUsedDaysResult
	}

	// Repositoryで使用日数合計を取得する
	usedDays, sumResult := service.paidLeaveRepository.SumPaidLeaveUsageDays(usedDaysQuery)
	if sumResult.Error {
		return sumResult
	}

	totalGrantedDays := calculateTotalGrantedDays(user.HireDate, now)
	nextGrantDate, nextGrantDays := calculateNextGrant(user.HireDate, now)
	requiredUseDeadline, requiredUseRemainingDays := calculateRequiredUseInfo(user.HireDate, now, usedDays)

	remainingDays := totalGrantedDays - usedDays

	return results.OK(
		types.PaidLeaveBalanceResponse{
			UserID: userID,

			TotalGrantedDays: totalGrantedDays,
			UsedDays:         usedDays,
			RemainingDays:    remainingDays,

			NextGrantDate: nextGrantDate,
			NextGrantDays: nextGrantDays,

			RequiredUseDays:          constants.PaidLeaveRequiredUseDays,
			RequiredUseDeadline:      requiredUseDeadline,
			RequiredUseRemainingDays: requiredUseRemainingDays,
		},
		"GET_USER_PAID_LEAVE_BALANCE_SUCCESS",
		"有給残数を取得しました",
		nil,
	)
}

/*
 * 現時点までの有給付与合計日数を計算する
 */
func calculateTotalGrantedDays(hireDate time.Time, targetDate time.Time) float64 {
	totalGrantedDays := 0.0

	for _, rule := range constants.PaidLeaveGrantRules {
		grantDate := hireDate.AddDate(0, rule.AfterMonths, 0)

		if grantDate.After(targetDate) {
			continue
		}

		totalGrantedDays += rule.GrantDays
	}

	return totalGrantedDays
}

/*
 * 次回付与予定日と次回付与日数を計算する
 */
func calculateNextGrant(hireDate time.Time, targetDate time.Time) (*time.Time, float64) {
	for _, rule := range constants.PaidLeaveGrantRules {
		grantDate := hireDate.AddDate(0, rule.AfterMonths, 0)

		if grantDate.After(targetDate) {
			return &grantDate, rule.GrantDays
		}
	}

	return nil, 0
}

/*
 * 年5日取得義務の期限と残り必要取得日数を計算する
 *
 * 現時点では簡易版：
 * ・直近の付与日数が10日以上の場合のみ対象
 * ・期限は直近付与日から1年後
 * ・使用日数は全期間合計を使う
 *
 * 注意：
 * ・本来は「付与日から1年以内に何日取得したか」で判定する
 * ・後で厳密化する場合は、付与日以降の使用日だけを集計する必要がある
 */
func calculateRequiredUseInfo(hireDate time.Time, targetDate time.Time, usedDays float64) (*time.Time, float64) {
	var latestGrantDate *time.Time
	var latestGrantDays float64

	for _, rule := range constants.PaidLeaveGrantRules {
		grantDate := hireDate.AddDate(0, rule.AfterMonths, 0)

		if grantDate.After(targetDate) {
			continue
		}

		latestGrantDate = &grantDate
		latestGrantDays = rule.GrantDays
	}

	if latestGrantDate == nil {
		return nil, 0
	}

	if latestGrantDays < 10 {
		return nil, 0
	}

	deadline := latestGrantDate.AddDate(1, 0, 0)

	remainingRequiredDays := constants.PaidLeaveRequiredUseDays - usedDays
	if remainingRequiredDays < 0 {
		remainingRequiredDays = 0
	}

	return &deadline, remainingRequiredDays
}
