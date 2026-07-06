package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type PaidLeaveUsageBuilder interface {
	BuildFindActiveUserByIDQuery(targetUserID uint) (*gorm.DB, results.Result)
	BuildFindActiveUsersForPaidLeaveRequiredUseWarningsQuery(targetDate time.Time) (*gorm.DB, results.Result)
	BuildSearchPaidLeaveUsagesQuery(req types.SearchPaidLeaveUsagesRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildSumActivePaidLeaveUsageDaysByUserIDQuery(targetUserID uint) (*gorm.DB, results.Result)
	BuildSumActivePaidLeaveUsageDaysByUserIDAndPeriodQuery(targetUserID uint, periodStart time.Time, periodEnd time.Time) (*gorm.DB, results.Result)
	BuildCreatePaidLeaveUsageModel(req types.CreatePaidLeaveUsageRequest, usageDate time.Time) (models.PaidLeaveUsage, results.Result)
	BuildFindManualPaidLeaveUsageByIDAndUserIDQuery(targetPaidLeaveUsageID uint, targetUserID uint) (*gorm.DB, results.Result)
	BuildUpdatePaidLeaveUsageModel(currentPaidLeaveUsage models.PaidLeaveUsage, req types.UpdatePaidLeaveUsageRequest, usageDate time.Time) (models.PaidLeaveUsage, results.Result)
	BuildDeletePaidLeaveUsageModel(currentPaidLeaveUsage models.PaidLeaveUsage) (models.PaidLeaveUsage, results.Result)
	BuildFindAutomaticPaidLeaveUsageByUserIDAndUsageDateQuery(targetUserID uint, usageDate time.Time) (*gorm.DB, results.Result)
	BuildCreateAutomaticPaidLeaveUsageModel(targetUserID uint, usageDate time.Time) (models.PaidLeaveUsage, results.Result)
	BuildActivateAutomaticPaidLeaveUsageModel(currentPaidLeaveUsage models.PaidLeaveUsage) (models.PaidLeaveUsage, results.Result)
	BuildDeleteAutomaticPaidLeaveUsageModel(currentPaidLeaveUsage models.PaidLeaveUsage) (models.PaidLeaveUsage, results.Result)
}

/*
 * 管理者用有給使用日Builder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Serviceから受け取ったRequestをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Count / Create / Save はRepositoryに任せる
 * ・管理者が追加する過去有給使用日は is_manual = true にする
 */
type paidLeaveUsageBuilder struct {
	db *gorm.DB
}

/*
 * PaidLeaveUsageBuilder生成
 */
func NewPaidLeaveUsageBuilder(db *gorm.DB) PaidLeaveUsageBuilder {
	return &paidLeaveUsageBuilder{db: db}
}

/*
 * 有効ユーザーID検索用クエリ作成
 *
 * 有給使用日の検索・追加・残数取得前に、
 * 対象ユーザーが存在するか確認するために使う。
 *
 * 論理削除済みユーザーは対象外。
 * ADMIN は有給管理の対象外。
 */
func (builder *paidLeaveUsageBuilder) BuildFindActiveUserByIDQuery(targetUserID uint) (*gorm.DB, results.Result) {
	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ACTIVE_USER_BY_ID_QUERY_INVALID_TARGET_USER_ID",
			"対象ユーザー取得条件の作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("id = ?", targetUserID).
		Where("role = ?", "USER").
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_ACTIVE_USER_BY_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給使用日検索用クエリ作成
 *
 * searchQuery：
 * ・一覧取得用
 * ・offset / limit / order を含む
 *
 * countQuery：
 * ・総件数取得用
 * ・offset / limit は含めない
 */

/*
 * 年5日取得義務警告対象候補ユーザー検索用クエリ作成
 *
 * 管理者ホーム画面で、期限が近い年5日取得義務未達ユーザーを抽出する前段として使う。
 *
 * 対象：
 * ・role = USER
 * ・論理削除されていない
 * ・退職日が未設定、または基準日以降
 */
func (builder *paidLeaveUsageBuilder) BuildFindActiveUsersForPaidLeaveRequiredUseWarningsQuery(targetDate time.Time) (*gorm.DB, results.Result) {
	query := builder.db.
		Model(&models.User{}).
		Where("role = ?", "USER").
		Where("is_deleted = ?", false).
		Where("retirement_date IS NULL OR retirement_date >= ?", targetDate).
		Order("hire_date ASC").
		Order("id ASC")

	return query, results.OK(
		nil,
		"BUILD_FIND_ACTIVE_USERS_FOR_PAID_LEAVE_REQUIRED_USE_WARNINGS_QUERY_SUCCESS",
		"",
		nil,
	)
}

func (builder *paidLeaveUsageBuilder) BuildSearchPaidLeaveUsagesQuery(req types.SearchPaidLeaveUsagesRequest) (*gorm.DB, *gorm.DB, results.Result) {
	if req.TargetUserID == 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_PAID_LEAVE_USAGES_QUERY_INVALID_TARGET_USER_ID",
			"有給使用日検索条件の作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

	if req.Offset < 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_PAID_LEAVE_USAGES_QUERY_INVALID_OFFSET",
			"有給使用日検索条件の作成に失敗しました",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	if req.Limit <= 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_PAID_LEAVE_USAGES_QUERY_INVALID_LIMIT",
			"有給使用日検索条件の作成に失敗しました",
			map[string]any{
				"limit": req.Limit,
			},
		)
	}

	searchQuery := builder.db.Model(&models.PaidLeaveUsage{})
	countQuery := builder.db.Model(&models.PaidLeaveUsage{})

	searchQuery = applySearchPaidLeaveUsagesCondition(searchQuery, req)
	countQuery = applySearchPaidLeaveUsagesCondition(countQuery, req)

	searchQuery = searchQuery.
		Order("usage_date DESC").
		Order("id DESC").
		Offset(req.Offset).
		Limit(req.Limit)

	return searchQuery, countQuery, results.OK(
		nil,
		"BUILD_SEARCH_PAID_LEAVE_USAGES_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有効な有給使用日数合計クエリ作成
 *
 * 有給残数計算で使う。
 *
 * 対象：
 * ・指定ユーザーの有給使用日
 * ・論理削除されていないもの
 */
func (builder *paidLeaveUsageBuilder) BuildSumActivePaidLeaveUsageDaysByUserIDQuery(targetUserID uint) (*gorm.DB, results.Result) {
	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_SUM_ACTIVE_PAID_LEAVE_USAGE_DAYS_BY_USER_ID_QUERY_INVALID_TARGET_USER_ID",
			"有給使用日数合計条件の作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	query := builder.db.
		Model(&models.PaidLeaveUsage{}).
		Where("user_id = ?", targetUserID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_SUM_ACTIVE_PAID_LEAVE_USAGE_DAYS_BY_USER_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給使用日作成用Model作成
 *
 * 管理者が過去有給使用日として追加するため、
 * IsManual は必ず true にする。
 */

/*
 * 指定期間内の有効な有給使用日数合計クエリ作成
 *
 * 年5日取得義務の判定で使う。
 *
 * 対象：
 * ・指定ユーザーの有給使用日
 * ・論理削除されていないもの
 * ・usage_date が periodStart 以上
 * ・usage_date が periodEnd 未満
 */
func (builder *paidLeaveUsageBuilder) BuildSumActivePaidLeaveUsageDaysByUserIDAndPeriodQuery(
	targetUserID uint,
	periodStart time.Time,
	periodEnd time.Time,
) (*gorm.DB, results.Result) {
	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_SUM_ACTIVE_PAID_LEAVE_USAGE_DAYS_BY_USER_ID_AND_PERIOD_QUERY_INVALID_TARGET_USER_ID",
			"有給使用日数合計条件の作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	if periodStart.IsZero() {
		return nil, results.BadRequest(
			"BUILD_SUM_ACTIVE_PAID_LEAVE_USAGE_DAYS_BY_USER_ID_AND_PERIOD_QUERY_EMPTY_PERIOD_START",
			"有給使用日数合計条件の作成に失敗しました",
			nil,
		)
	}

	if periodEnd.IsZero() || !periodEnd.After(periodStart) {
		return nil, results.BadRequest(
			"BUILD_SUM_ACTIVE_PAID_LEAVE_USAGE_DAYS_BY_USER_ID_AND_PERIOD_QUERY_INVALID_PERIOD_END",
			"有給使用日数合計条件の作成に失敗しました",
			map[string]any{
				"periodStart": periodStart,
				"periodEnd":   periodEnd,
			},
		)
	}

	query := builder.db.
		Model(&models.PaidLeaveUsage{}).
		Where("user_id = ?", targetUserID).
		Where("is_deleted = ?", false).
		Where("usage_date >= ?", periodStart).
		Where("usage_date < ?", periodEnd)

	return query, results.OK(
		nil,
		"BUILD_SUM_ACTIVE_PAID_LEAVE_USAGE_DAYS_BY_USER_ID_AND_PERIOD_QUERY_SUCCESS",
		"",
		nil,
	)
}

func (builder *paidLeaveUsageBuilder) BuildCreatePaidLeaveUsageModel(
	req types.CreatePaidLeaveUsageRequest,
	usageDate time.Time,
) (models.PaidLeaveUsage, results.Result) {
	if req.TargetUserID == 0 {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_CREATE_PAID_LEAVE_USAGE_MODEL_INVALID_TARGET_USER_ID",
			"有給使用日作成データの作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

	if req.UsageDays <= 0 {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_CREATE_PAID_LEAVE_USAGE_MODEL_INVALID_USAGE_DAYS",
			"有給使用日数が正しくありません",
			map[string]any{
				"usageDays": req.UsageDays,
			},
		)
	}

	paidLeaveUsage := models.PaidLeaveUsage{
		UserID:    req.TargetUserID,
		UsageDate: usageDate,
		UsageDays: req.UsageDays,
		IsManual:  true,
		Memo:      req.Memo,
		IsDeleted: false,
		DeletedAt: nil,
	}

	return paidLeaveUsage, results.OK(
		nil,
		"BUILD_CREATE_PAID_LEAVE_USAGE_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 手動追加の有給使用日ID検索用クエリ作成
 *
 * 更新・削除で使う。
 *
 * 条件：
 * ・対象ID
 * ・対象ユーザーID
 * ・論理削除されていない
 * ・手動追加である
 *
 * 注意：
 * ・勤怠連携や有給申請連携で作られた使用履歴は、
 *   この管理画面から直接編集・削除しない。
 */
func (builder *paidLeaveUsageBuilder) BuildFindManualPaidLeaveUsageByIDAndUserIDQuery(
	targetPaidLeaveUsageID uint,
	targetUserID uint,
) (*gorm.DB, results.Result) {
	if targetPaidLeaveUsageID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MANUAL_PAID_LEAVE_USAGE_BY_ID_AND_USER_ID_QUERY_INVALID_TARGET_PAID_LEAVE_USAGE_ID",
			"有給使用日取得条件の作成に失敗しました",
			map[string]any{
				"targetPaidLeaveUsageId": targetPaidLeaveUsageID,
			},
		)
	}

	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MANUAL_PAID_LEAVE_USAGE_BY_ID_AND_USER_ID_QUERY_INVALID_TARGET_USER_ID",
			"有給使用日取得条件の作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	query := builder.db.
		Model(&models.PaidLeaveUsage{}).
		Where("id = ?", targetPaidLeaveUsageID).
		Where("user_id = ?", targetUserID).
		Where("is_deleted = ?", false).
		Where("is_manual = ?", true)

	return query, results.OK(
		nil,
		"BUILD_FIND_MANUAL_PAID_LEAVE_USAGE_BY_ID_AND_USER_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給使用日更新用Model作成
 */
func (builder *paidLeaveUsageBuilder) BuildUpdatePaidLeaveUsageModel(
	currentPaidLeaveUsage models.PaidLeaveUsage,
	req types.UpdatePaidLeaveUsageRequest,
	usageDate time.Time,
) (models.PaidLeaveUsage, results.Result) {
	if currentPaidLeaveUsage.ID == 0 {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_UPDATE_PAID_LEAVE_USAGE_MODEL_EMPTY_CURRENT_PAID_LEAVE_USAGE",
			"有給使用日更新データの作成に失敗しました",
			nil,
		)
	}

	if currentPaidLeaveUsage.UserID != req.TargetUserID {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_UPDATE_PAID_LEAVE_USAGE_MODEL_USER_ID_MISMATCH",
			"有給使用日更新データの作成に失敗しました",
			map[string]any{
				"currentUserId": currentPaidLeaveUsage.UserID,
				"targetUserId":  req.TargetUserID,
			},
		)
	}

	if !currentPaidLeaveUsage.IsManual {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_UPDATE_PAID_LEAVE_USAGE_MODEL_NOT_MANUAL",
			"手動追加ではない有給使用日は編集できません",
			map[string]any{
				"targetPaidLeaveUsageId": currentPaidLeaveUsage.ID,
			},
		)
	}

	if req.UsageDays <= 0 {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_UPDATE_PAID_LEAVE_USAGE_MODEL_INVALID_USAGE_DAYS",
			"有給使用日数が正しくありません",
			map[string]any{
				"usageDays": req.UsageDays,
			},
		)
	}

	currentPaidLeaveUsage.UsageDate = usageDate
	currentPaidLeaveUsage.UsageDays = req.UsageDays
	currentPaidLeaveUsage.Memo = req.Memo

	/*
	 * 管理者画面から更新できるのは手動追加データのみなので、
	 * 念のため true のまま維持する。
	 */
	currentPaidLeaveUsage.IsManual = true

	return currentPaidLeaveUsage, results.OK(
		nil,
		"BUILD_UPDATE_PAID_LEAVE_USAGE_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給使用日論理削除用Model作成
 */
func (builder *paidLeaveUsageBuilder) BuildDeletePaidLeaveUsageModel(
	currentPaidLeaveUsage models.PaidLeaveUsage,
) (models.PaidLeaveUsage, results.Result) {
	if currentPaidLeaveUsage.ID == 0 {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_DELETE_PAID_LEAVE_USAGE_MODEL_EMPTY_CURRENT_PAID_LEAVE_USAGE",
			"有給使用日削除データの作成に失敗しました",
			nil,
		)
	}

	if !currentPaidLeaveUsage.IsManual {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_DELETE_PAID_LEAVE_USAGE_MODEL_NOT_MANUAL",
			"手動追加ではない有給使用日は削除できません",
			map[string]any{
				"targetPaidLeaveUsageId": currentPaidLeaveUsage.ID,
			},
		)
	}

	now := time.Now()

	currentPaidLeaveUsage.IsDeleted = true
	currentPaidLeaveUsage.DeletedAt = &now

	return currentPaidLeaveUsage, results.OK(
		nil,
		"BUILD_DELETE_PAID_LEAVE_USAGE_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠連携用有給使用日検索クエリ作成
 *
 * 対象：
 * ・対象ユーザー
 * ・対象日
 * ・is_manual = false
 *
 * 注意：
 * ・論理削除済みも含めて取得する
 * ・有給を再選択した場合は、既存データを復活させるために使う
 */
func (builder *paidLeaveUsageBuilder) BuildFindAutomaticPaidLeaveUsageByUserIDAndUsageDateQuery(
	targetUserID uint,
	usageDate time.Time,
) (*gorm.DB, results.Result) {
	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_AUTOMATIC_PAID_LEAVE_USAGE_INVALID_TARGET_USER_ID",
			"勤怠連携用有給使用日取得条件の作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	if usageDate.IsZero() {
		return nil, results.BadRequest(
			"BUILD_FIND_AUTOMATIC_PAID_LEAVE_USAGE_EMPTY_USAGE_DATE",
			"勤怠連携用有給使用日取得条件の作成に失敗しました",
			nil,
		)
	}

	query := builder.db.
		Model(&models.PaidLeaveUsage{}).
		Where("user_id = ?", targetUserID).
		Where("usage_date = ?", usageDate).
		Where("is_manual = ?", false).
		Order("id DESC")

	return query, results.OK(
		nil,
		"BUILD_FIND_AUTOMATIC_PAID_LEAVE_USAGE_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠連携用有給使用日作成Model作成
 */
func (builder *paidLeaveUsageBuilder) BuildCreateAutomaticPaidLeaveUsageModel(
	targetUserID uint,
	usageDate time.Time,
) (models.PaidLeaveUsage, results.Result) {
	if targetUserID == 0 {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_CREATE_AUTOMATIC_PAID_LEAVE_USAGE_INVALID_TARGET_USER_ID",
			"勤怠連携用有給使用日作成データの作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	if usageDate.IsZero() {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_CREATE_AUTOMATIC_PAID_LEAVE_USAGE_EMPTY_USAGE_DATE",
			"勤怠連携用有給使用日作成データの作成に失敗しました",
			nil,
		)
	}

	paidLeaveUsage := models.PaidLeaveUsage{
		UserID:    targetUserID,
		UsageDate: usageDate,
		UsageDays: 1.0,
		IsManual:  false,
		Memo:      "月次勤怠全体保存から登録",
		IsDeleted: false,
		DeletedAt: nil,
	}

	return paidLeaveUsage, results.OK(
		nil,
		"BUILD_CREATE_AUTOMATIC_PAID_LEAVE_USAGE_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠連携用有給使用日復活Model作成
 */
func (builder *paidLeaveUsageBuilder) BuildActivateAutomaticPaidLeaveUsageModel(
	currentPaidLeaveUsage models.PaidLeaveUsage,
) (models.PaidLeaveUsage, results.Result) {
	if currentPaidLeaveUsage.ID == 0 {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_ACTIVATE_AUTOMATIC_PAID_LEAVE_USAGE_EMPTY_CURRENT_DATA",
			"勤怠連携用有給使用日の復活データ作成に失敗しました",
			nil,
		)
	}

	if currentPaidLeaveUsage.IsManual {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_ACTIVATE_AUTOMATIC_PAID_LEAVE_USAGE_MANUAL_DATA",
			"手動追加の有給使用日は勤怠保存から変更できません",
			map[string]any{
				"paidLeaveUsageId": currentPaidLeaveUsage.ID,
			},
		)
	}

	currentPaidLeaveUsage.UsageDays = 1.0
	currentPaidLeaveUsage.IsDeleted = false
	currentPaidLeaveUsage.DeletedAt = nil

	return currentPaidLeaveUsage, results.OK(
		nil,
		"BUILD_ACTIVATE_AUTOMATIC_PAID_LEAVE_USAGE_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠連携用有給使用日論理削除Model作成
 */
func (builder *paidLeaveUsageBuilder) BuildDeleteAutomaticPaidLeaveUsageModel(
	currentPaidLeaveUsage models.PaidLeaveUsage,
) (models.PaidLeaveUsage, results.Result) {
	if currentPaidLeaveUsage.ID == 0 {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_DELETE_AUTOMATIC_PAID_LEAVE_USAGE_EMPTY_CURRENT_DATA",
			"勤怠連携用有給使用日の削除データ作成に失敗しました",
			nil,
		)
	}

	if currentPaidLeaveUsage.IsManual {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_DELETE_AUTOMATIC_PAID_LEAVE_USAGE_MANUAL_DATA",
			"手動追加の有給使用日は勤怠保存から削除できません",
			map[string]any{
				"paidLeaveUsageId": currentPaidLeaveUsage.ID,
			},
		)
	}

	now := time.Now()
	currentPaidLeaveUsage.IsDeleted = true
	currentPaidLeaveUsage.DeletedAt = &now

	return currentPaidLeaveUsage, results.OK(
		nil,
		"BUILD_DELETE_AUTOMATIC_PAID_LEAVE_USAGE_SUCCESS",
		"",
		nil,
	)
}

/*
 * 有給使用日検索条件をGORMクエリへ適用する
 */
func applySearchPaidLeaveUsagesCondition(query *gorm.DB, req types.SearchPaidLeaveUsagesRequest) *gorm.DB {
	query = query.Where("user_id = ?", req.TargetUserID)

	if !req.IncludeDeleted {
		query = query.Where("is_deleted = ?", false)
	}

	return query
}
