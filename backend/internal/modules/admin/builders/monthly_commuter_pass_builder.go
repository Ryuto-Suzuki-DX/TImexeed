package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type MonthlyCommuterPassBuilder interface {
	BuildFindMonthlyCommuterPassesByUserIDAndTargetYearMonthQuery(targetUserID uint, targetYear int, targetMonth int) (*gorm.DB, results.Result)
	BuildCreateMonthlyCommuterPassModel(targetUserID uint, targetYear int, targetMonth int, req types.UpdateMonthlyCommuterPassItemRequest) (models.MonthlyCommuterPass, results.Result)
	BuildUpdateMonthlyCommuterPassModel(currentMonthlyCommuterPass models.MonthlyCommuterPass, targetUserID uint, targetYear int, targetMonth int, req types.UpdateMonthlyCommuterPassItemRequest) (models.MonthlyCommuterPass, results.Result)
	BuildDeleteMonthlyCommuterPassModel(currentMonthlyCommuterPass models.MonthlyCommuterPass) (models.MonthlyCommuterPass, results.Result)
}

type monthlyCommuterPassBuilder struct {
	db *gorm.DB
}

func NewMonthlyCommuterPassBuilder(db *gorm.DB) MonthlyCommuterPassBuilder {
	return &monthlyCommuterPassBuilder{db: db}
}

func (builder *monthlyCommuterPassBuilder) BuildFindMonthlyCommuterPassesByUserIDAndTargetYearMonthQuery(
	targetUserID uint,
	targetYear int,
	targetMonth int,
) (*gorm.DB, results.Result) {
	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MONTHLY_COMMUTER_PASSES_QUERY_INVALID_TARGET_USER_ID",
			"月次通勤定期取得条件の作成に失敗しました",
			map[string]any{"targetUserId": targetUserID},
		)
	}

	if targetYear <= 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MONTHLY_COMMUTER_PASSES_QUERY_INVALID_TARGET_YEAR",
			"月次通勤定期取得条件の作成に失敗しました",
			map[string]any{"targetYear": targetYear},
		)
	}

	if targetMonth < 1 || targetMonth > 12 {
		return nil, results.BadRequest(
			"BUILD_FIND_MONTHLY_COMMUTER_PASSES_QUERY_INVALID_TARGET_MONTH",
			"月次通勤定期取得条件の作成に失敗しました",
			map[string]any{"targetMonth": targetMonth},
		)
	}

	query := builder.db.
		Model(&models.MonthlyCommuterPass{}).
		Where("user_id = ?", targetUserID).
		Where("target_year = ?", targetYear).
		Where("target_month = ?", targetMonth).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_MONTHLY_COMMUTER_PASSES_QUERY_SUCCESS",
		"",
		nil,
	)
}

func (builder *monthlyCommuterPassBuilder) BuildCreateMonthlyCommuterPassModel(
	targetUserID uint,
	targetYear int,
	targetMonth int,
	req types.UpdateMonthlyCommuterPassItemRequest,
) (models.MonthlyCommuterPass, results.Result) {
	validationResult := validateMonthlyCommuterPassTarget(targetUserID, targetYear, targetMonth, "BUILD_CREATE_MONTHLY_COMMUTER_PASS_MODEL")
	if validationResult.Error {
		return models.MonthlyCommuterPass{}, validationResult
	}

	monthlyCommuterPass := models.MonthlyCommuterPass{
		UserID:         targetUserID,
		TargetYear:     targetYear,
		TargetMonth:    targetMonth,
		CommuterFrom:   req.CommuterFrom,
		CommuterTo:     req.CommuterTo,
		CommuterMethod: req.CommuterMethod,
		CommuterAmount: req.CommuterAmount,
		IsDeleted:      false,
		DeletedAt:      nil,
	}

	return monthlyCommuterPass, results.OK(
		nil,
		"BUILD_CREATE_MONTHLY_COMMUTER_PASS_MODEL_SUCCESS",
		"",
		nil,
	)
}

func (builder *monthlyCommuterPassBuilder) BuildUpdateMonthlyCommuterPassModel(
	currentMonthlyCommuterPass models.MonthlyCommuterPass,
	targetUserID uint,
	targetYear int,
	targetMonth int,
	req types.UpdateMonthlyCommuterPassItemRequest,
) (models.MonthlyCommuterPass, results.Result) {
	if currentMonthlyCommuterPass.ID == 0 {
		return models.MonthlyCommuterPass{}, results.BadRequest(
			"BUILD_UPDATE_MONTHLY_COMMUTER_PASS_MODEL_EMPTY_CURRENT_MONTHLY_COMMUTER_PASS",
			"月次通勤定期更新データの作成に失敗しました",
			nil,
		)
	}

	validationResult := validateMonthlyCommuterPassTarget(targetUserID, targetYear, targetMonth, "BUILD_UPDATE_MONTHLY_COMMUTER_PASS_MODEL")
	if validationResult.Error {
		return models.MonthlyCommuterPass{}, validationResult
	}

	if currentMonthlyCommuterPass.UserID != targetUserID ||
		currentMonthlyCommuterPass.TargetYear != targetYear ||
		currentMonthlyCommuterPass.TargetMonth != targetMonth {
		return models.MonthlyCommuterPass{}, results.Conflict(
			"BUILD_UPDATE_MONTHLY_COMMUTER_PASS_MODEL_TARGET_MISMATCH",
			"月次通勤定期更新対象が一致しません",
			map[string]any{
				"monthlyCommuterPassId": currentMonthlyCommuterPass.ID,
				"targetUserId":          targetUserID,
				"targetYear":            targetYear,
				"targetMonth":           targetMonth,
			},
		)
	}

	currentMonthlyCommuterPass.CommuterFrom = req.CommuterFrom
	currentMonthlyCommuterPass.CommuterTo = req.CommuterTo
	currentMonthlyCommuterPass.CommuterMethod = req.CommuterMethod
	currentMonthlyCommuterPass.CommuterAmount = req.CommuterAmount
	currentMonthlyCommuterPass.IsDeleted = false
	currentMonthlyCommuterPass.DeletedAt = nil

	return currentMonthlyCommuterPass, results.OK(
		nil,
		"BUILD_UPDATE_MONTHLY_COMMUTER_PASS_MODEL_SUCCESS",
		"",
		nil,
	)
}

func (builder *monthlyCommuterPassBuilder) BuildDeleteMonthlyCommuterPassModel(
	currentMonthlyCommuterPass models.MonthlyCommuterPass,
) (models.MonthlyCommuterPass, results.Result) {
	if currentMonthlyCommuterPass.ID == 0 {
		return models.MonthlyCommuterPass{}, results.BadRequest(
			"BUILD_DELETE_MONTHLY_COMMUTER_PASS_MODEL_EMPTY_CURRENT_MONTHLY_COMMUTER_PASS",
			"月次通勤定期削除データの作成に失敗しました",
			nil,
		)
	}

	now := time.Now()
	currentMonthlyCommuterPass.IsDeleted = true
	currentMonthlyCommuterPass.DeletedAt = &now

	return currentMonthlyCommuterPass, results.OK(
		nil,
		"BUILD_DELETE_MONTHLY_COMMUTER_PASS_MODEL_SUCCESS",
		"",
		nil,
	)
}

func validateMonthlyCommuterPassTarget(
	targetUserID uint,
	targetYear int,
	targetMonth int,
	actionCode string,
) results.Result {
	if targetUserID == 0 {
		return results.BadRequest(
			actionCode+"_INVALID_TARGET_USER_ID",
			"対象ユーザーIDが正しくありません",
			map[string]any{"targetUserId": targetUserID},
		)
	}

	if targetYear <= 0 {
		return results.BadRequest(
			actionCode+"_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{"targetYear": targetYear},
		)
	}

	if targetMonth < 1 || targetMonth > 12 {
		return results.BadRequest(
			actionCode+"_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{"targetMonth": targetMonth},
		)
	}

	return results.OK(nil, actionCode+"_VALID_TARGET", "", nil)
}
