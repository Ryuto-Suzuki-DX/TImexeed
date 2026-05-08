package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type MonthlyCommuterPassBuilder interface {
	BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(userID uint, targetYear int, targetMonth int) (*gorm.DB, results.Result)
	BuildCreateMonthlyCommuterPassModel(userID uint, req types.UpdateMonthlyCommuterPassRequest) (models.MonthlyCommuterPass, results.Result)
	BuildUpdateMonthlyCommuterPassModel(currentMonthlyCommuterPass models.MonthlyCommuterPass, req types.UpdateMonthlyCommuterPassRequest) (models.MonthlyCommuterPass, results.Result)
	BuildDeleteMonthlyCommuterPassModel(currentMonthlyCommuterPass models.MonthlyCommuterPass) (models.MonthlyCommuterPass, results.Result)
}

/*
 * 従業員用月次通勤定期Builder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Serviceから受け取ったRequestをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Create / Save はRepositoryに任せる
 */
type monthlyCommuterPassBuilder struct {
	db *gorm.DB
}

/*
 * MonthlyCommuterPassBuilder生成
 */
func NewMonthlyCommuterPassBuilder(db *gorm.DB) MonthlyCommuterPassBuilder {
	return &monthlyCommuterPassBuilder{db: db}
}

/*
 * ユーザーID + 対象年月で月次通勤定期1件取得用クエリ作成
 *
 * 検索・更新・削除時に使う。
 */
func (builder *monthlyCommuterPassBuilder) BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(
	userID uint,
	targetYear int,
	targetMonth int,
) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MONTHLY_COMMUTER_PASS_QUERY_INVALID_USER_ID",
			"月次通勤定期取得条件の作成に失敗しました",
			map[string]any{
				"userId": userID,
			},
		)
	}

	if targetYear <= 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MONTHLY_COMMUTER_PASS_QUERY_INVALID_TARGET_YEAR",
			"月次通勤定期取得条件の作成に失敗しました",
			map[string]any{
				"targetYear": targetYear,
			},
		)
	}

	if targetMonth < 1 || targetMonth > 12 {
		return nil, results.BadRequest(
			"BUILD_FIND_MONTHLY_COMMUTER_PASS_QUERY_INVALID_TARGET_MONTH",
			"月次通勤定期取得条件の作成に失敗しました",
			map[string]any{
				"targetMonth": targetMonth,
			},
		)
	}

	query := builder.db.
		Model(&models.MonthlyCommuterPass{}).
		Where("user_id = ?", userID).
		Where("target_year = ?", targetYear).
		Where("target_month = ?", targetMonth).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_MONTHLY_COMMUTER_PASS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次通勤定期作成用Model作成
 *
 * 画面上は「更新」操作だが、対象年月の通勤定期が未登録の場合は新規作成する。
 */
func (builder *monthlyCommuterPassBuilder) BuildCreateMonthlyCommuterPassModel(
	userID uint,
	req types.UpdateMonthlyCommuterPassRequest,
) (models.MonthlyCommuterPass, results.Result) {
	if userID == 0 {
		return models.MonthlyCommuterPass{}, results.BadRequest(
			"BUILD_CREATE_MONTHLY_COMMUTER_PASS_MODEL_INVALID_USER_ID",
			"月次通勤定期作成データの作成に失敗しました",
			map[string]any{
				"userId": userID,
			},
		)
	}

	if req.TargetYear <= 0 {
		return models.MonthlyCommuterPass{}, results.BadRequest(
			"BUILD_CREATE_MONTHLY_COMMUTER_PASS_MODEL_INVALID_TARGET_YEAR",
			"月次通勤定期作成データの作成に失敗しました",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return models.MonthlyCommuterPass{}, results.BadRequest(
			"BUILD_CREATE_MONTHLY_COMMUTER_PASS_MODEL_INVALID_TARGET_MONTH",
			"月次通勤定期作成データの作成に失敗しました",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	monthlyCommuterPass := models.MonthlyCommuterPass{
		UserID:         userID,
		TargetYear:     req.TargetYear,
		TargetMonth:    req.TargetMonth,
		CommuterFrom:   req.CommuterFrom,
		CommuterTo:     req.CommuterTo,
		CommuterMethod: req.CommuterMethod,
		CommuterAmount: req.CommuterAmount,
		MonthlyStatus:  "DRAFT",
		IsDeleted:      false,
	}

	return monthlyCommuterPass, results.OK(
		nil,
		"BUILD_CREATE_MONTHLY_COMMUTER_PASS_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次通勤定期更新用Model作成
 *
 * 対象年月の通勤定期が登録済みの場合に更新する。
 */
func (builder *monthlyCommuterPassBuilder) BuildUpdateMonthlyCommuterPassModel(
	currentMonthlyCommuterPass models.MonthlyCommuterPass,
	req types.UpdateMonthlyCommuterPassRequest,
) (models.MonthlyCommuterPass, results.Result) {
	if currentMonthlyCommuterPass.ID == 0 {
		return models.MonthlyCommuterPass{}, results.BadRequest(
			"BUILD_UPDATE_MONTHLY_COMMUTER_PASS_MODEL_EMPTY_CURRENT_MONTHLY_COMMUTER_PASS",
			"月次通勤定期更新データの作成に失敗しました",
			nil,
		)
	}

	if req.TargetYear <= 0 {
		return models.MonthlyCommuterPass{}, results.BadRequest(
			"BUILD_UPDATE_MONTHLY_COMMUTER_PASS_MODEL_INVALID_TARGET_YEAR",
			"月次通勤定期更新データの作成に失敗しました",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return models.MonthlyCommuterPass{}, results.BadRequest(
			"BUILD_UPDATE_MONTHLY_COMMUTER_PASS_MODEL_INVALID_TARGET_MONTH",
			"月次通勤定期更新データの作成に失敗しました",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	currentMonthlyCommuterPass.TargetYear = req.TargetYear
	currentMonthlyCommuterPass.TargetMonth = req.TargetMonth
	currentMonthlyCommuterPass.CommuterFrom = req.CommuterFrom
	currentMonthlyCommuterPass.CommuterTo = req.CommuterTo
	currentMonthlyCommuterPass.CommuterMethod = req.CommuterMethod
	currentMonthlyCommuterPass.CommuterAmount = req.CommuterAmount

	return currentMonthlyCommuterPass, results.OK(
		nil,
		"BUILD_UPDATE_MONTHLY_COMMUTER_PASS_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次通勤定期論理削除用Model作成
 */
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
