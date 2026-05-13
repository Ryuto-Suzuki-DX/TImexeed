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
	BuildSearchPaidLeaveUsagesQuery(req types.SearchPaidLeaveUsagesRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildSumActivePaidLeaveUsageDaysByUserIDQuery(targetUserID uint) (*gorm.DB, results.Result)
	BuildCreatePaidLeaveUsageModel(req types.CreatePaidLeaveUsageRequest, usageDate time.Time) (models.PaidLeaveUsage, results.Result)
	BuildFindManualPaidLeaveUsageByIDAndUserIDQuery(targetPaidLeaveUsageID uint, targetUserID uint) (*gorm.DB, results.Result)
	BuildUpdatePaidLeaveUsageModel(currentPaidLeaveUsage models.PaidLeaveUsage, req types.UpdatePaidLeaveUsageRequest, usageDate time.Time) (models.PaidLeaveUsage, results.Result)
	BuildDeletePaidLeaveUsageModel(currentPaidLeaveUsage models.PaidLeaveUsage) (models.PaidLeaveUsage, results.Result)
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
 * 有給使用日検索条件をGORMクエリへ適用する
 */
func applySearchPaidLeaveUsagesCondition(query *gorm.DB, req types.SearchPaidLeaveUsagesRequest) *gorm.DB {
	query = query.Where("user_id = ?", req.TargetUserID)

	if !req.IncludeDeleted {
		query = query.Where("is_deleted = ?", false)
	}

	return query
}
