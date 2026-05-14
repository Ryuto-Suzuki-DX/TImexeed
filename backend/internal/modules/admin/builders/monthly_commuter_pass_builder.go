package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type MonthlyCommuterPassBuilder interface {
	BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(targetUserID uint, targetYear int, targetMonth int) (*gorm.DB, results.Result)
	BuildCreateMonthlyCommuterPassModel(req types.UpdateMonthlyCommuterPassRequest) (models.MonthlyCommuterPass, results.Result)
	BuildUpdateMonthlyCommuterPassModel(currentMonthlyCommuterPass models.MonthlyCommuterPass, req types.UpdateMonthlyCommuterPassRequest) (models.MonthlyCommuterPass, results.Result)
	BuildDeleteMonthlyCommuterPassModel(currentMonthlyCommuterPass models.MonthlyCommuterPass) (models.MonthlyCommuterPass, results.Result)
}

/*
 * 管理者用月次通勤定期Builder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Serviceから受け取ったRequestをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Create / Save はRepositoryに任せる
 * ・MonthlyCommuterPass は申請状態を持たない
 * ・月次申請状態は MonthlyAttendanceRequest 側で管理する
 * ・管理者側では月次申請状態による編集ロックを行わない
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
 *
 * 注意：
 * ・targetUserID は管理者が選択した対象ユーザーID
 * ・論理削除済みの月次通勤定期は対象外
 */
func (builder *monthlyCommuterPassBuilder) BuildFindMonthlyCommuterPassByUserIDAndTargetYearMonthQuery(
	targetUserID uint,
	targetYear int,
	targetMonth int,
) (*gorm.DB, results.Result) {
	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MONTHLY_COMMUTER_PASS_QUERY_INVALID_TARGET_USER_ID",
			"月次通勤定期取得条件の作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
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
		Where("user_id = ?", targetUserID).
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
 *
 * 注意：
 * ・MonthlyCommuterPass には月次申請状態を保存しない
 */
func (builder *monthlyCommuterPassBuilder) BuildCreateMonthlyCommuterPassModel(
	req types.UpdateMonthlyCommuterPassRequest,
) (models.MonthlyCommuterPass, results.Result) {
	if req.TargetUserID == 0 {
		return models.MonthlyCommuterPass{}, results.BadRequest(
			"BUILD_CREATE_MONTHLY_COMMUTER_PASS_MODEL_INVALID_TARGET_USER_ID",
			"月次通勤定期作成データの作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
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
		UserID:         req.TargetUserID,
		TargetYear:     req.TargetYear,
		TargetMonth:    req.TargetMonth,
		CommuterFrom:   req.CommuterFrom,
		CommuterTo:     req.CommuterTo,
		CommuterMethod: req.CommuterMethod,
		CommuterAmount: req.CommuterAmount,
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
 *
 * 注意：
 * ・MonthlyCommuterPass には月次申請状態を保存しない
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

	if req.TargetUserID == 0 {
		return models.MonthlyCommuterPass{}, results.BadRequest(
			"BUILD_UPDATE_MONTHLY_COMMUTER_PASS_MODEL_INVALID_TARGET_USER_ID",
			"月次通勤定期更新データの作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

	if currentMonthlyCommuterPass.UserID != req.TargetUserID {
		return models.MonthlyCommuterPass{}, results.Conflict(
			"BUILD_UPDATE_MONTHLY_COMMUTER_PASS_MODEL_USER_ID_MISMATCH",
			"月次通勤定期更新対象のユーザーが一致しません",
			map[string]any{
				"currentUserId": currentMonthlyCommuterPass.UserID,
				"targetUserId":  req.TargetUserID,
			},
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

	if currentMonthlyCommuterPass.TargetYear != req.TargetYear ||
		currentMonthlyCommuterPass.TargetMonth != req.TargetMonth {
		return models.MonthlyCommuterPass{}, results.BadRequest(
			"BUILD_UPDATE_MONTHLY_COMMUTER_PASS_MODEL_TARGET_MONTH_MISMATCH",
			"月次通勤定期更新対象の年月が一致しません",
			map[string]any{
				"currentTargetYear":  currentMonthlyCommuterPass.TargetYear,
				"currentTargetMonth": currentMonthlyCommuterPass.TargetMonth,
				"requestTargetYear":  req.TargetYear,
				"requestTargetMonth": req.TargetMonth,
			},
		)
	}

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
