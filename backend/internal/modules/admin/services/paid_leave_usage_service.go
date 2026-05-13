package services

import (
	"time"

	"timexeed/backend/internal/constants"
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 管理者用有給使用日Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type PaidLeaveUsageService interface {
	SearchPaidLeaveUsages(req types.SearchPaidLeaveUsagesRequest) results.Result
	GetPaidLeaveBalance(req types.GetPaidLeaveBalanceRequest) results.Result
	CreatePaidLeaveUsage(req types.CreatePaidLeaveUsageRequest) results.Result
	UpdatePaidLeaveUsage(req types.UpdatePaidLeaveUsageRequest) results.Result
	DeletePaidLeaveUsage(req types.DeletePaidLeaveUsageRequest) results.Result
}

/*
 * 管理者用有給使用日Service
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや更新用Modelを作成する
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
type paidLeaveUsageService struct {
	paidLeaveUsageBuilder    builders.PaidLeaveUsageBuilder
	paidLeaveUsageRepository repositories.PaidLeaveUsageRepository
}

/*
 * PaidLeaveUsageService生成
 */
func NewPaidLeaveUsageService(
	paidLeaveUsageBuilder builders.PaidLeaveUsageBuilder,
	paidLeaveUsageRepository repositories.PaidLeaveUsageRepository,
) *paidLeaveUsageService {
	return &paidLeaveUsageService{
		paidLeaveUsageBuilder:    paidLeaveUsageBuilder,
		paidLeaveUsageRepository: paidLeaveUsageRepository,
	}
}

/*
 * models.PaidLeaveUsageをフロント返却用PaidLeaveUsageResponseへ変換する
 *
 * 日付はtime.Time / *time.Timeのまま返す。
 * 表示形式の整形はフロント側で行う。
 */
func toPaidLeaveUsageResponse(paidLeaveUsage models.PaidLeaveUsage) types.PaidLeaveUsageResponse {
	return types.PaidLeaveUsageResponse{
		ID:        paidLeaveUsage.ID,
		UserID:    paidLeaveUsage.UserID,
		UsageDate: paidLeaveUsage.UsageDate,
		UsageDays: paidLeaveUsage.UsageDays,
		IsManual:  paidLeaveUsage.IsManual,
		Memo:      paidLeaveUsage.Memo,
		IsDeleted: paidLeaveUsage.IsDeleted,
		CreatedAt: paidLeaveUsage.CreatedAt,
		UpdatedAt: paidLeaveUsage.UpdatedAt,
		DeletedAt: paidLeaveUsage.DeletedAt,
	}
}

/*
 * 有給使用日検索
 *
 * ページング方針：
 * ・初回は offset=0, limit=50
 * ・さらに表示するときは、フロントで現在表示済みの件数を offset として送る
 * ・limit が未指定、0以下の場合は 50件にする
 * ・limit が 50件を超える場合も 50件に丸める
 *
 * hasMore：
 * ・総件数 total が offset + 今回取得件数 より多ければ true
 * ・それ以下なら false
 */
func (service *paidLeaveUsageService) SearchPaidLeaveUsages(req types.SearchPaidLeaveUsagesRequest) results.Result {
	// ページング検索条件を共通関数で正規化する
	normalizedCondition, normalizeResult := utils.NormalizePageSearchCondition(
		utils.PageSearchCondition{
			Keyword: "",
			Offset:  req.Offset,
			Limit:   req.Limit,
		},
		"SEARCH_PAID_LEAVE_USAGES_INVALID_OFFSET",
		"検索開始位置が正しくありません",
	)
	if normalizeResult.Error {
		return normalizeResult
	}

	req.Offset = normalizedCondition.Offset
	req.Limit = normalizedCondition.Limit

	// Builderで対象ユーザー取得用クエリを作成する
	userQuery, buildUserResult := service.paidLeaveUsageBuilder.BuildFindActiveUserByIDQuery(req.TargetUserID)
	if buildUserResult.Error {
		return buildUserResult
	}

	// Repositoryで対象ユーザーの存在確認をする
	_, findUserResult := service.paidLeaveUsageRepository.FindUser(userQuery)
	if findUserResult.Error {
		return findUserResult
	}

	// Builderで一覧検索用クエリと件数取得用クエリを作成する
	searchQuery, countQuery, buildSearchResult := service.paidLeaveUsageBuilder.BuildSearchPaidLeaveUsagesQuery(req)
	if buildSearchResult.Error {
		return buildSearchResult
	}

	// Repositoryで有給使用日一覧を取得する
	paidLeaveUsages, findResult := service.paidLeaveUsageRepository.FindPaidLeaveUsages(searchQuery)
	if findResult.Error {
		return findResult
	}

	// Repositoryで検索条件に一致する総件数を取得する
	total, countResult := service.paidLeaveUsageRepository.CountPaidLeaveUsages(countQuery)
	if countResult.Error {
		return countResult
	}

	// DBモデルをフロント返却用Responseへ変換する
	paidLeaveUsageResponses := make([]types.PaidLeaveUsageResponse, 0, len(paidLeaveUsages))
	for _, paidLeaveUsage := range paidLeaveUsages {
		paidLeaveUsageResponses = append(paidLeaveUsageResponses, toPaidLeaveUsageResponse(paidLeaveUsage))
	}

	hasMore := utils.HasMore(total, req.Offset, len(paidLeaveUsages))

	return results.OK(
		types.SearchPaidLeaveUsagesResponse{
			PaidLeaveUsages: paidLeaveUsageResponses,
			Total:           total,
			Offset:          req.Offset,
			Limit:           req.Limit,
			HasMore:         hasMore,
		},
		"SEARCH_PAID_LEAVE_USAGES_SUCCESS",
		"有給使用日一覧を取得しました",
		nil,
	)
}

/*
 * 現時点の有給残数取得
 *
 * 計算方針：
 * ・対象ユーザーの入社日を取得する
 * ・固定値ファイルの付与ルールをもとに、現時点までの付与日数を合計する
 * ・有給使用日テーブルの使用日数を合計する
 * ・残数 = 付与合計 - 使用合計
 *
 * 注意：
 * ・現時点では、出勤率8割判定は未実装
 * ・現時点では、有給申請・勤怠承認との連携分はPaidLeaveUsageに入っている前提
 */
func (service *paidLeaveUsageService) GetPaidLeaveBalance(req types.GetPaidLeaveBalanceRequest) results.Result {
	now := time.Now()

	// Builderで対象ユーザー取得用クエリを作成する
	userQuery, buildUserResult := service.paidLeaveUsageBuilder.BuildFindActiveUserByIDQuery(req.TargetUserID)
	if buildUserResult.Error {
		return buildUserResult
	}

	// Repositoryで対象ユーザーを取得する
	user, findUserResult := service.paidLeaveUsageRepository.FindUser(userQuery)
	if findUserResult.Error {
		return findUserResult
	}

	// Builderで対象ユーザーの有給使用日数合計用クエリを作成する
	usedDaysQuery, buildUsedDaysResult := service.paidLeaveUsageBuilder.BuildSumActivePaidLeaveUsageDaysByUserIDQuery(req.TargetUserID)
	if buildUsedDaysResult.Error {
		return buildUsedDaysResult
	}

	// Repositoryで使用日数合計を取得する
	usedDays, sumResult := service.paidLeaveUsageRepository.SumPaidLeaveUsageDays(usedDaysQuery)
	if sumResult.Error {
		return sumResult
	}

	totalGrantedDays := calculateTotalGrantedDays(user.HireDate, now)
	nextGrantDate, nextGrantDays := calculateNextGrant(user.HireDate, now)
	requiredUseDeadline, requiredUseRemainingDays := calculateRequiredUseInfo(user.HireDate, now, usedDays)

	remainingDays := totalGrantedDays - usedDays

	return results.OK(
		types.PaidLeaveBalanceResponse{
			TargetUserID: req.TargetUserID,

			TotalGrantedDays: totalGrantedDays,
			UsedDays:         usedDays,
			RemainingDays:    remainingDays,

			NextGrantDate: nextGrantDate,
			NextGrantDays: nextGrantDays,

			RequiredUseDays:          constants.PaidLeaveRequiredUseDays,
			RequiredUseDeadline:      requiredUseDeadline,
			RequiredUseRemainingDays: requiredUseRemainingDays,
		},
		"GET_PAID_LEAVE_BALANCE_SUCCESS",
		"有給残数を取得しました",
		nil,
	)
}

/*
 * 過去有給使用日追加
 *
 * 注意：
 * ・管理者が過去分として手動追加するAPI
 * ・isManual はフロントから受け取らない
 * ・Service側で isManual = true を強制する
 */
func (service *paidLeaveUsageService) CreatePaidLeaveUsage(req types.CreatePaidLeaveUsageRequest) results.Result {
	// 有給使用日を日付型へ変換する
	usageDate, err := utils.ParseDate(req.UsageDate)
	if err != nil {
		return results.BadRequest(
			"CREATE_PAID_LEAVE_USAGE_INVALID_USAGE_DATE",
			"有給使用日の形式が正しくありません",
			map[string]any{
				"usageDate": req.UsageDate,
				"format":    "yyyy-MM-dd",
			},
		)
	}

	// 使用日数を検証する
	if !isValidPaidLeaveUsageDays(req.UsageDays) {
		return results.BadRequest(
			"CREATE_PAID_LEAVE_USAGE_INVALID_USAGE_DAYS",
			"有給使用日数が正しくありません",
			map[string]any{
				"usageDays": req.UsageDays,
				"allowed":   []float64{0.5, 1.0},
			},
		)
	}

	// Builderで対象ユーザー取得用クエリを作成する
	userQuery, buildUserResult := service.paidLeaveUsageBuilder.BuildFindActiveUserByIDQuery(req.TargetUserID)
	if buildUserResult.Error {
		return buildUserResult
	}

	// Repositoryで対象ユーザーの存在確認をする
	_, findUserResult := service.paidLeaveUsageRepository.FindUser(userQuery)
	if findUserResult.Error {
		return findUserResult
	}

	// Builderで作成用Modelを作る
	paidLeaveUsage, buildCreateResult := service.paidLeaveUsageBuilder.BuildCreatePaidLeaveUsageModel(
		req,
		usageDate,
	)
	if buildCreateResult.Error {
		return buildCreateResult
	}

	// Repositoryで有給使用日を作成する
	createdPaidLeaveUsage, createResult := service.paidLeaveUsageRepository.CreatePaidLeaveUsage(paidLeaveUsage)
	if createResult.Error {
		return createResult
	}

	return results.Created(
		types.CreatePaidLeaveUsageResponse{
			PaidLeaveUsage: toPaidLeaveUsageResponse(createdPaidLeaveUsage),
		},
		"CREATE_PAID_LEAVE_USAGE_SUCCESS",
		"有給使用日を作成しました",
		nil,
	)
}

/*
 * 過去有給使用日編集
 *
 * 注意：
 * ・対象データが targetUserId のものか確認する
 * ・手動追加データだけ編集可能にする
 */
func (service *paidLeaveUsageService) UpdatePaidLeaveUsage(req types.UpdatePaidLeaveUsageRequest) results.Result {
	// 有給使用日を日付型へ変換する
	usageDate, err := utils.ParseDate(req.UsageDate)
	if err != nil {
		return results.BadRequest(
			"UPDATE_PAID_LEAVE_USAGE_INVALID_USAGE_DATE",
			"有給使用日の形式が正しくありません",
			map[string]any{
				"usageDate": req.UsageDate,
				"format":    "yyyy-MM-dd",
			},
		)
	}

	// 使用日数を検証する
	if !isValidPaidLeaveUsageDays(req.UsageDays) {
		return results.BadRequest(
			"UPDATE_PAID_LEAVE_USAGE_INVALID_USAGE_DAYS",
			"有給使用日数が正しくありません",
			map[string]any{
				"usageDays": req.UsageDays,
				"allowed":   []float64{0.5, 1.0},
			},
		)
	}

	// Builderで対象有給使用日取得用クエリを作成する
	findQuery, buildFindResult := service.paidLeaveUsageBuilder.BuildFindManualPaidLeaveUsageByIDAndUserIDQuery(
		req.TargetPaidLeaveUsageID,
		req.TargetUserID,
	)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象有給使用日を取得する
	currentPaidLeaveUsage, findResult := service.paidLeaveUsageRepository.FindPaidLeaveUsage(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderで更新用Modelを作る
	updatedPaidLeaveUsage, buildUpdateResult := service.paidLeaveUsageBuilder.BuildUpdatePaidLeaveUsageModel(
		currentPaidLeaveUsage,
		req,
		usageDate,
	)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	// Repositoryで有給使用日を更新する
	savedPaidLeaveUsage, saveResult := service.paidLeaveUsageRepository.SavePaidLeaveUsage(updatedPaidLeaveUsage)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.UpdatePaidLeaveUsageResponse{
			PaidLeaveUsage: toPaidLeaveUsageResponse(savedPaidLeaveUsage),
		},
		"UPDATE_PAID_LEAVE_USAGE_SUCCESS",
		"有給使用日を更新しました",
		nil,
	)
}

/*
 * 過去有給使用日削除
 *
 * 注意：
 * ・物理削除ではなく論理削除する
 * ・対象データが targetUserId のものか確認する
 * ・手動追加データだけ削除可能にする
 */
func (service *paidLeaveUsageService) DeletePaidLeaveUsage(req types.DeletePaidLeaveUsageRequest) results.Result {
	// Builderで対象有給使用日取得用クエリを作成する
	findQuery, buildFindResult := service.paidLeaveUsageBuilder.BuildFindManualPaidLeaveUsageByIDAndUserIDQuery(
		req.TargetPaidLeaveUsageID,
		req.TargetUserID,
	)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象有給使用日を取得する
	currentPaidLeaveUsage, findResult := service.paidLeaveUsageRepository.FindPaidLeaveUsage(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderで論理削除用Modelを作る
	deletedPaidLeaveUsage, buildDeleteResult := service.paidLeaveUsageBuilder.BuildDeletePaidLeaveUsageModel(currentPaidLeaveUsage)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	// Repositoryで有給使用日を保存する
	_, saveResult := service.paidLeaveUsageRepository.SavePaidLeaveUsage(deletedPaidLeaveUsage)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeletePaidLeaveUsageResponse{
			TargetUserID:           req.TargetUserID,
			TargetPaidLeaveUsageID: req.TargetPaidLeaveUsageID,
		},
		"DELETE_PAID_LEAVE_USAGE_SUCCESS",
		"有給使用日を削除しました",
		nil,
	)
}

/*
 * 有給使用日数の妥当性チェック
 *
 * 現時点では、
 * ・1日
 * ・半日
 * のみ許可する。
 */
func isValidPaidLeaveUsageDays(usageDays float64) bool {
	return usageDays == 1.0 || usageDays == 0.5
}

/*
 * 現時点までの有給付与合計日数を計算する
 *
 * hireDate:
 * ・入社日
 *
 * targetDate:
 * ・計算基準日
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
