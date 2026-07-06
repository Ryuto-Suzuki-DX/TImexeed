package services

import (
	"time"

	"timexeed/backend/internal/constants"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 従業員用有給Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type PaidLeaveService interface {
	GetPaidLeaveBalance(userID uint) results.Result
	ValidateMonthlyAttendancePaidLeaveBalance(
		userID uint,
		req types.UpdateMonthlyAttendanceRequest,
		paidLeaveAttendanceTypeIDs map[uint]bool,
	) results.Result
	SyncAutomaticPaidLeaveUsage(
		userID uint,
		workDate string,
		shouldUsePaidLeave bool,
	) results.Result
}

/*
 * 従業員用有給Service
 *
 * 役割：
 * ・Controllerから受け取ったログインユーザーIDをもとに処理を進める
 * ・現在の有給残数を返す
 * ・月次勤怠全体保存前に、保存後の有給残数を確認する
 * ・勤怠画面由来の有給使用履歴を日付単位で同期する
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builder/Repositoryのエラーはそのまま返す
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・勤怠画面由来の有給使用履歴は IsManual=false とする
 * ・管理者が手動登録した IsManual=true の履歴は変更しない
 */
type paidLeaveService struct {
	paidLeaveBuilder    builders.PaidLeaveBuilder
	paidLeaveRepository repositories.PaidLeaveRepository
}

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
 */
func (service *paidLeaveService) GetPaidLeaveBalance(userID uint) results.Result {
	if userID == 0 {
		return results.BadRequest(
			"GET_PAID_LEAVE_BALANCE_INVALID_USER_ID",
			"有給残数取得の対象ユーザーが正しくありません",
			map[string]any{"userId": userID},
		)
	}

	now := time.Now()

	userQuery, buildUserResult := service.paidLeaveBuilder.BuildFindActiveUserByIDQuery(userID)
	if buildUserResult.Error {
		return buildUserResult
	}

	user, findUserResult := service.paidLeaveRepository.FindUser(userQuery)
	if findUserResult.Error {
		return findUserResult
	}

	usedDaysQuery, buildUsedDaysResult := service.paidLeaveBuilder.BuildSumActivePaidLeaveUsageDaysByUserIDQuery(userID)
	if buildUsedDaysResult.Error {
		return buildUsedDaysResult
	}

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
			UserID:                   userID,
			TotalGrantedDays:         totalGrantedDays,
			UsedDays:                 usedDays,
			RemainingDays:            remainingDays,
			NextGrantDate:            nextGrantDate,
			NextGrantDays:            nextGrantDays,
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
 * 月次勤怠全体保存前の有給残数チェック
 *
 * 現在の有給使用履歴と、今回の月次保存後に必要になる
 * 勤怠画面由来の有給使用履歴との差分を計算する。
 *
 * 注意：
 * ・勤怠画面の有給は1日単位で扱う
 * ・同じ日を再保存するだけなら新規消費として数えない
 * ・有給を外す日については現在の自動登録分を差し引く
 * ・同じ日付がRequest内に重複しても一度だけ判定する
 */
func (service *paidLeaveService) ValidateMonthlyAttendancePaidLeaveBalance(
	userID uint,
	req types.UpdateMonthlyAttendanceRequest,
	paidLeaveAttendanceTypeIDs map[uint]bool,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"VALIDATE_MONTHLY_ATTENDANCE_PAID_LEAVE_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	userQuery, buildUserResult := service.paidLeaveBuilder.BuildFindActiveUserByIDQuery(userID)
	if buildUserResult.Error {
		return buildUserResult
	}

	user, findUserResult := service.paidLeaveRepository.FindUser(userQuery)
	if findUserResult.Error {
		return findUserResult
	}

	usedDaysQuery, buildUsedDaysResult := service.paidLeaveBuilder.BuildSumActivePaidLeaveUsageDaysByUserIDQuery(userID)
	if buildUsedDaysResult.Error {
		return buildUsedDaysResult
	}

	usedDays, sumResult := service.paidLeaveRepository.SumPaidLeaveUsageDays(usedDaysQuery)
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

		findQuery, buildFindResult := service.paidLeaveBuilder.BuildFindAutomaticPaidLeaveUsageByUserIDAndUsageDateQuery(
			userID,
			usageDate,
		)
		if buildFindResult.Error {
			return buildFindResult
		}

		currentPaidLeaveUsage, findResult := service.paidLeaveRepository.FindPaidLeaveUsage(findQuery)
		hasActiveAutomaticUsage := false

		if findResult.Error {
			if findResult.Code != "USER_PAID_LEAVE_USAGE_NOT_FOUND" {
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
 * shouldUsePaidLeave=true：
 * ・未登録なら IsManual=false で作成
 * ・論理削除済みなら復活
 * ・有効なデータが既にあれば何もしない
 *
 * shouldUsePaidLeave=false：
 * ・有効な IsManual=false データがあれば論理削除
 * ・未登録または削除済みなら何もしない
 */
func (service *paidLeaveService) SyncAutomaticPaidLeaveUsage(
	userID uint,
	workDate string,
	shouldUsePaidLeave bool,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"SYNC_AUTOMATIC_PAID_LEAVE_USAGE_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
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

	findQuery, buildFindResult := service.paidLeaveBuilder.BuildFindAutomaticPaidLeaveUsageByUserIDAndUsageDateQuery(
		userID,
		usageDate,
	)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentPaidLeaveUsage, findResult := service.paidLeaveRepository.FindPaidLeaveUsage(findQuery)

	if findResult.Error {
		if findResult.Code != "USER_PAID_LEAVE_USAGE_NOT_FOUND" {
			return findResult
		}

		if !shouldUsePaidLeave {
			return results.OK(nil, "SYNC_AUTOMATIC_PAID_LEAVE_USAGE_NOTHING_TO_DELETE", "", nil)
		}

		paidLeaveUsage, buildCreateResult := service.paidLeaveBuilder.BuildCreateAutomaticPaidLeaveUsageModel(
			userID,
			usageDate,
		)
		if buildCreateResult.Error {
			return buildCreateResult
		}

		_, createResult := service.paidLeaveRepository.CreatePaidLeaveUsage(paidLeaveUsage)
		if createResult.Error {
			return createResult
		}

		return results.OK(nil, "SYNC_AUTOMATIC_PAID_LEAVE_USAGE_CREATED", "", nil)
	}

	if shouldUsePaidLeave {
		if !currentPaidLeaveUsage.IsDeleted {
			return results.OK(nil, "SYNC_AUTOMATIC_PAID_LEAVE_USAGE_ALREADY_ACTIVE", "", nil)
		}

		activatedPaidLeaveUsage, buildActivateResult := service.paidLeaveBuilder.BuildActivateAutomaticPaidLeaveUsageModel(currentPaidLeaveUsage)
		if buildActivateResult.Error {
			return buildActivateResult
		}

		_, saveResult := service.paidLeaveRepository.SavePaidLeaveUsage(activatedPaidLeaveUsage)
		if saveResult.Error {
			return saveResult
		}

		return results.OK(nil, "SYNC_AUTOMATIC_PAID_LEAVE_USAGE_ACTIVATED", "", nil)
	}

	if currentPaidLeaveUsage.IsDeleted {
		return results.OK(nil, "SYNC_AUTOMATIC_PAID_LEAVE_USAGE_ALREADY_DELETED", "", nil)
	}

	deletedPaidLeaveUsage, buildDeleteResult := service.paidLeaveBuilder.BuildDeleteAutomaticPaidLeaveUsageModel(currentPaidLeaveUsage)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	_, saveResult := service.paidLeaveRepository.SavePaidLeaveUsage(deletedPaidLeaveUsage)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(nil, "SYNC_AUTOMATIC_PAID_LEAVE_USAGE_DELETED", "", nil)
}

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

func calculateNextGrant(hireDate time.Time, targetDate time.Time) (*time.Time, float64) {
	for _, rule := range constants.PaidLeaveGrantRules {
		grantDate := hireDate.AddDate(0, rule.AfterMonths, 0)
		if grantDate.After(targetDate) {
			return &grantDate, rule.GrantDays
		}
	}
	return nil, 0
}

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

	if latestGrantDate == nil || latestGrantDays < 10 {
		return nil, 0
	}

	deadline := latestGrantDate.AddDate(1, 0, 0)
	remainingRequiredDays := constants.PaidLeaveRequiredUseDays - usedDays
	if remainingRequiredDays < 0 {
		remainingRequiredDays = 0
	}

	return &deadline, remainingRequiredDays
}
