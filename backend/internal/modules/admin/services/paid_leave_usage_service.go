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
	SearchPaidLeaveRequiredUseWarnings(req types.SearchPaidLeaveRequiredUseWarningsRequest) results.Result
	CreatePaidLeaveUsage(req types.CreatePaidLeaveUsageRequest) results.Result
	UpdatePaidLeaveUsage(req types.UpdatePaidLeaveUsageRequest) results.Result
	DeletePaidLeaveUsage(req types.DeletePaidLeaveUsageRequest) results.Result
	ValidateMonthlyAttendancePaidLeaveBalance(req types.UpdateMonthlyAttendanceRequest, paidLeaveAttendanceTypeIDs map[uint]bool) results.Result
	SyncAutomaticPaidLeaveUsage(targetUserID uint, workDate string, shouldUsePaidLeave bool) results.Result
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
	requiredUseStartDate, requiredUseDeadline, hasRequiredUsePeriod := calculateRequiredUsePeriod(user.HireDate, now)

	usedDaysInRequiredPeriod := 0.0
	if hasRequiredUsePeriod && requiredUseDeadline != nil {
		usedDaysInRequiredPeriodQuery, buildUsedDaysInRequiredPeriodResult := service.paidLeaveUsageBuilder.BuildSumActivePaidLeaveUsageDaysByUserIDAndPeriodQuery(
			req.TargetUserID,
			requiredUseStartDate,
			*requiredUseDeadline,
		)
		if buildUsedDaysInRequiredPeriodResult.Error {
			return buildUsedDaysInRequiredPeriodResult
		}

		usedDaysInRequiredPeriodResult, sumUsedDaysInRequiredPeriodResult := service.paidLeaveUsageRepository.SumPaidLeaveUsageDays(usedDaysInRequiredPeriodQuery)
		if sumUsedDaysInRequiredPeriodResult.Error {
			return sumUsedDaysInRequiredPeriodResult
		}

		usedDaysInRequiredPeriod = usedDaysInRequiredPeriodResult
	}

	_, requiredUseRemainingDays := calculateRequiredUseInfo(user.HireDate, now, usedDaysInRequiredPeriod)

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
 * 年5日取得義務警告一覧取得
 *
 * 管理者ホーム画面で、期限が近く、年5日取得義務を満たしていないユーザーを表示するために使う。
 *
 * 判定方針：
 * ・対象は有効なUSERのみ
 * ・直近の10日以上付与日を年5日取得義務の開始日とする
 * ・期限は開始日から1年後
 * ・使用日数は開始日以上、期限未満の有給使用日だけを集計する
 * ・期限まで deadlineWithinDays 日以内、かつ残り必要日数があるユーザーだけ返す
 */
func (service *paidLeaveUsageService) SearchPaidLeaveRequiredUseWarnings(
	req types.SearchPaidLeaveRequiredUseWarningsRequest,
) results.Result {
	now := time.Now()
	today := truncateDate(now)
	deadlineWithinDays := normalizePaidLeaveRequiredUseWarningDeadlineWithinDays(req.DeadlineWithinDays)

	usersQuery, buildUsersResult := service.paidLeaveUsageBuilder.BuildFindActiveUsersForPaidLeaveRequiredUseWarningsQuery(today)
	if buildUsersResult.Error {
		return buildUsersResult
	}

	users, findUsersResult := service.paidLeaveUsageRepository.FindUsers(usersQuery)
	if findUsersResult.Error {
		return findUsersResult
	}

	warnings := make([]types.PaidLeaveRequiredUseWarningResponse, 0)

	for _, user := range users {
		requiredUseStartDate, requiredUseDeadline, hasRequiredUsePeriod := calculateRequiredUsePeriod(user.HireDate, now)
		if !hasRequiredUsePeriod || requiredUseDeadline == nil {
			continue
		}

		deadlineRemainingDays := calculateDateDiffDays(today, *requiredUseDeadline)
		if deadlineRemainingDays > deadlineWithinDays {
			continue
		}

		usedDaysQuery, buildUsedDaysResult := service.paidLeaveUsageBuilder.BuildSumActivePaidLeaveUsageDaysByUserIDAndPeriodQuery(
			user.ID,
			requiredUseStartDate,
			*requiredUseDeadline,
		)
		if buildUsedDaysResult.Error {
			return buildUsedDaysResult
		}

		usedDaysInRequiredPeriod, sumResult := service.paidLeaveUsageRepository.SumPaidLeaveUsageDays(usedDaysQuery)
		if sumResult.Error {
			return sumResult
		}

		requiredUseRemainingDays := constants.PaidLeaveRequiredUseDays - usedDaysInRequiredPeriod
		if requiredUseRemainingDays <= 0 {
			continue
		}

		warnings = append(warnings, types.PaidLeaveRequiredUseWarningResponse{
			UserID: user.ID,

			UserName:  user.Name,
			UserEmail: user.Email,

			HireDate: user.HireDate,

			RequiredUseStartDate:     requiredUseStartDate,
			RequiredUseDeadline:      requiredUseDeadline,
			DeadlineRemainingDays:    deadlineRemainingDays,
			RequiredUseDays:          constants.PaidLeaveRequiredUseDays,
			UsedDaysInRequiredPeriod: usedDaysInRequiredPeriod,
			RequiredUseRemainingDays: requiredUseRemainingDays,
		})
	}

	return results.OK(
		types.SearchPaidLeaveRequiredUseWarningsResponse{
			Warnings: warnings,
			Total:    len(warnings),
		},
		"SEARCH_PAID_LEAVE_REQUIRED_USE_WARNINGS_SUCCESS",
		"年5日取得義務の警告対象ユーザー一覧を取得しました",
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
 * 月次勤怠全体保存前の有給残数チェック
 *
 * 現在の有給使用履歴と、今回の月次保存後に必要となる
 * 勤怠連携用有給使用履歴との差分を計算して確認する。
 *
 * 注意：
 * ・勤怠画面由来の有給は1日単位で扱う
 * ・手動追加データは現在の使用済み日数としてそのまま含める
 * ・同じ有給を再保存するだけの場合は新規消費として数えない
 * ・有給を外す日については、現在の自動登録分を差し引く
 */
func (service *paidLeaveUsageService) ValidateMonthlyAttendancePaidLeaveBalance(
	req types.UpdateMonthlyAttendanceRequest,
	paidLeaveAttendanceTypeIDs map[uint]bool,
) results.Result {
	if req.TargetUserID == 0 {
		return results.BadRequest(
			"VALIDATE_MONTHLY_ATTENDANCE_PAID_LEAVE_INVALID_TARGET_USER_ID",
			"有給残数確認の対象ユーザーが正しくありません",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

	userQuery, buildUserResult := service.paidLeaveUsageBuilder.BuildFindActiveUserByIDQuery(req.TargetUserID)
	if buildUserResult.Error {
		return buildUserResult
	}

	user, findUserResult := service.paidLeaveUsageRepository.FindUser(userQuery)
	if findUserResult.Error {
		return findUserResult
	}

	usedDaysQuery, buildUsedDaysResult := service.paidLeaveUsageBuilder.BuildSumActivePaidLeaveUsageDaysByUserIDQuery(req.TargetUserID)
	if buildUsedDaysResult.Error {
		return buildUsedDaysResult
	}

	usedDays, sumResult := service.paidLeaveUsageRepository.SumPaidLeaveUsageDays(usedDaysQuery)
	if sumResult.Error {
		return sumResult
	}

	usageDaysDelta := 0.0
	processedWorkDates := make(map[string]bool)

	for _, attendanceDayReq := range req.AttendanceDays {
		if processedWorkDates[attendanceDayReq.WorkDate] {
			continue
		}
		processedWorkDates[attendanceDayReq.WorkDate] = true
		usageDate, err := utils.ParseDate(attendanceDayReq.WorkDate)
		if err != nil {
			return results.BadRequest(
				"VALIDATE_MONTHLY_ATTENDANCE_PAID_LEAVE_INVALID_WORK_DATE",
				"有給残数確認対象日の形式が正しくありません",
				map[string]any{
					"workDate": attendanceDayReq.WorkDate,
					"format":   "yyyy-MM-dd",
				},
			)
		}

		findQuery, buildFindResult := service.paidLeaveUsageBuilder.BuildFindAutomaticPaidLeaveUsageByUserIDAndUsageDateQuery(
			req.TargetUserID,
			usageDate,
		)
		if buildFindResult.Error {
			return buildFindResult
		}

		currentPaidLeaveUsage, findResult := service.paidLeaveUsageRepository.FindPaidLeaveUsage(findQuery)
		hasActiveAutomaticUsage := false

		if findResult.Error {
			if findResult.Code != "PAID_LEAVE_USAGE_NOT_FOUND" {
				return findResult
			}
		} else {
			hasActiveAutomaticUsage = !currentPaidLeaveUsage.IsDeleted
		}

		shouldUsePaidLeave := paidLeaveAttendanceTypeIDs[attendanceDayReq.PlanAttendanceTypeID]

		if shouldUsePaidLeave && !hasActiveAutomaticUsage {
			usageDaysDelta += 1.0
		}

		if !shouldUsePaidLeave && hasActiveAutomaticUsage {
			usageDaysDelta -= 1.0
		}
	}

	totalGrantedDays := calculateTotalGrantedDays(user.HireDate, time.Now())
	projectedUsedDays := usedDays + usageDaysDelta
	projectedRemainingDays := totalGrantedDays - projectedUsedDays

	if projectedRemainingDays < 0 {
		return results.BadRequest(
			"UPDATE_MONTHLY_ATTENDANCE_PAID_LEAVE_BALANCE_NOT_ENOUGH",
			"有給残数が不足しているため、有給を登録できません",
			map[string]any{
				"targetUserId":           req.TargetUserID,
				"totalGrantedDays":       totalGrantedDays,
				"currentUsedDays":        usedDays,
				"usageDaysDelta":         usageDaysDelta,
				"projectedUsedDays":      projectedUsedDays,
				"projectedRemainingDays": projectedRemainingDays,
			},
		)
	}

	return results.OK(
		nil,
		"VALIDATE_MONTHLY_ATTENDANCE_PAID_LEAVE_BALANCE_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠画面由来の有給使用履歴を日付単位で同期する
 *
 * shouldUsePaidLeave = true：
 * ・未登録なら IsManual=false で作成
 * ・論理削除済みなら復活
 * ・有効なデータが既にあれば何もしない
 *
 * shouldUsePaidLeave = false：
 * ・有効な IsManual=false データがあれば論理削除
 * ・未登録または削除済みなら何もしない
 *
 * 管理者が手動登録した IsManual=true のデータは対象外。
 */
func (service *paidLeaveUsageService) SyncAutomaticPaidLeaveUsage(
	targetUserID uint,
	workDate string,
	shouldUsePaidLeave bool,
) results.Result {
	if targetUserID == 0 {
		return results.BadRequest(
			"SYNC_AUTOMATIC_PAID_LEAVE_USAGE_INVALID_TARGET_USER_ID",
			"勤怠連携用有給使用日の同期対象ユーザーが正しくありません",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	usageDate, err := utils.ParseDate(workDate)
	if err != nil {
		return results.BadRequest(
			"SYNC_AUTOMATIC_PAID_LEAVE_USAGE_INVALID_WORK_DATE",
			"勤怠連携用有給使用日の対象日形式が正しくありません",
			map[string]any{
				"workDate": workDate,
				"format":   "yyyy-MM-dd",
			},
		)
	}

	findQuery, buildFindResult := service.paidLeaveUsageBuilder.BuildFindAutomaticPaidLeaveUsageByUserIDAndUsageDateQuery(
		targetUserID,
		usageDate,
	)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentPaidLeaveUsage, findResult := service.paidLeaveUsageRepository.FindPaidLeaveUsage(findQuery)

	if findResult.Error {
		if findResult.Code != "PAID_LEAVE_USAGE_NOT_FOUND" {
			return findResult
		}

		if !shouldUsePaidLeave {
			return results.OK(
				nil,
				"SYNC_AUTOMATIC_PAID_LEAVE_USAGE_NOTHING_TO_DELETE",
				"",
				nil,
			)
		}

		paidLeaveUsage, buildCreateResult := service.paidLeaveUsageBuilder.BuildCreateAutomaticPaidLeaveUsageModel(
			targetUserID,
			usageDate,
		)
		if buildCreateResult.Error {
			return buildCreateResult
		}

		_, createResult := service.paidLeaveUsageRepository.CreatePaidLeaveUsage(paidLeaveUsage)
		if createResult.Error {
			return createResult
		}

		return results.OK(
			nil,
			"SYNC_AUTOMATIC_PAID_LEAVE_USAGE_CREATED",
			"",
			nil,
		)
	}

	if shouldUsePaidLeave {
		if !currentPaidLeaveUsage.IsDeleted {
			return results.OK(
				nil,
				"SYNC_AUTOMATIC_PAID_LEAVE_USAGE_ALREADY_ACTIVE",
				"",
				nil,
			)
		}

		activatedPaidLeaveUsage, buildActivateResult := service.paidLeaveUsageBuilder.BuildActivateAutomaticPaidLeaveUsageModel(currentPaidLeaveUsage)
		if buildActivateResult.Error {
			return buildActivateResult
		}

		_, saveResult := service.paidLeaveUsageRepository.SavePaidLeaveUsage(activatedPaidLeaveUsage)
		if saveResult.Error {
			return saveResult
		}

		return results.OK(
			nil,
			"SYNC_AUTOMATIC_PAID_LEAVE_USAGE_ACTIVATED",
			"",
			nil,
		)
	}

	if currentPaidLeaveUsage.IsDeleted {
		return results.OK(
			nil,
			"SYNC_AUTOMATIC_PAID_LEAVE_USAGE_ALREADY_DELETED",
			"",
			nil,
		)
	}

	deletedPaidLeaveUsage, buildDeleteResult := service.paidLeaveUsageBuilder.BuildDeleteAutomaticPaidLeaveUsageModel(currentPaidLeaveUsage)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	_, saveResult := service.paidLeaveUsageRepository.SavePaidLeaveUsage(deletedPaidLeaveUsage)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		nil,
		"SYNC_AUTOMATIC_PAID_LEAVE_USAGE_DELETED",
		"",
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
 * 年5日取得義務の対象期間を計算する
 *
 * 対象：
 * ・直近の付与日数が10日以上の付与
 *
 * 戻り値：
 * ・開始日：直近の10日以上付与日
 * ・期限：開始日から1年後
 * ・対象期間があるか
 */
func calculateRequiredUsePeriod(hireDate time.Time, targetDate time.Time) (time.Time, *time.Time, bool) {
	var latestGrantDate time.Time
	hasRequiredUsePeriod := false

	for _, rule := range constants.PaidLeaveGrantRules {
		grantDate := hireDate.AddDate(0, rule.AfterMonths, 0)

		if grantDate.After(targetDate) {
			continue
		}

		if rule.GrantDays < 10 {
			continue
		}

		latestGrantDate = grantDate
		hasRequiredUsePeriod = true
	}

	if !hasRequiredUsePeriod {
		return time.Time{}, nil, false
	}

	deadline := latestGrantDate.AddDate(1, 0, 0)

	return latestGrantDate, &deadline, true
}

/*
 * 年5日取得義務の期限と残り必要取得日数を計算する
 *
 * usedDaysInRequiredPeriod：
 * ・直近の10日以上付与日から、その1年後の期限までに取得した有給使用日数
 */
func calculateRequiredUseInfo(hireDate time.Time, targetDate time.Time, usedDaysInRequiredPeriod float64) (*time.Time, float64) {
	_, deadline, hasRequiredUsePeriod := calculateRequiredUsePeriod(hireDate, targetDate)
	if !hasRequiredUsePeriod {
		return nil, 0
	}

	remainingRequiredDays := constants.PaidLeaveRequiredUseDays - usedDaysInRequiredPeriod
	if remainingRequiredDays < 0 {
		remainingRequiredDays = 0
	}

	return deadline, remainingRequiredDays
}

/*
 * 年5日取得義務警告の期限日数条件を補正する
 */
func normalizePaidLeaveRequiredUseWarningDeadlineWithinDays(deadlineWithinDays int) int {
	if deadlineWithinDays <= 0 {
		return 90
	}

	return deadlineWithinDays
}

/*
 * 日付の時刻部分を切り捨てる
 */
func truncatePaidLeaveUsageDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

/*
 * 日付差分を日数で計算する
 */
func calculateDateDiffDays(from time.Time, to time.Time) int {
	fromDate := truncateDate(from)
	toDate := truncateDate(to)

	return int(toDate.Sub(fromDate).Hours() / 24)
}
