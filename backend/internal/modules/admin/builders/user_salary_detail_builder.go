package builders

import (
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用ユーザー給与詳細Builder interface
 */
type UserSalaryDetailBuilder interface {
	BuildSearchUserSalaryDetailsQuery(req types.SearchUserSalaryDetailsRequest) (*gorm.DB, *gorm.DB, results.Result)
	BuildFindUserSalaryDetailByIDQuery(userSalaryDetailID uint) (*gorm.DB, results.Result)
	BuildCountOverlapUserSalaryDetailsQuery(userID uint, effectiveFrom time.Time, effectiveTo *time.Time) (*gorm.DB, results.Result)
	BuildCountOverlapUserSalaryDetailsExceptIDQuery(userID uint, userSalaryDetailID uint, effectiveFrom time.Time, effectiveTo *time.Time) (*gorm.DB, results.Result)
	BuildCreateUserSalaryDetailModel(req types.CreateUserSalaryDetailRequest, effectiveFrom time.Time, effectiveTo *time.Time) (models.UserSalaryDetail, results.Result)
	BuildUpdateUserSalaryDetailModel(currentUserSalaryDetail models.UserSalaryDetail, req types.UpdateUserSalaryDetailRequest, effectiveFrom time.Time, effectiveTo *time.Time) (models.UserSalaryDetail, results.Result)
	BuildDeleteUserSalaryDetailModel(currentUserSalaryDetail models.UserSalaryDetail) (models.UserSalaryDetail, results.Result)
}

/*
 * 管理者用ユーザー給与詳細Builder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Serviceから受け取ったRequestをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Count / Create / Save はRepositoryに任せる
 * ・給与区分、金額、適用期間の業務検証はServiceに任せる
 */
type userSalaryDetailBuilder struct {
	db *gorm.DB
}

/*
 * UserSalaryDetailBuilder生成
 */
func NewUserSalaryDetailBuilder(db *gorm.DB) UserSalaryDetailBuilder {
	return &userSalaryDetailBuilder{
		db: db,
	}
}

/*
 * ユーザー給与詳細検索用クエリ作成
 *
 * searchQuery：
 * ・一覧取得用
 * ・offset / limit / order を含む
 *
 * countQuery：
 * ・総件数取得用
 * ・offset / limit は含めない
 */
func (builder *userSalaryDetailBuilder) BuildSearchUserSalaryDetailsQuery(req types.SearchUserSalaryDetailsRequest) (*gorm.DB, *gorm.DB, results.Result) {
	if req.TargetUserID == 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_USER_SALARY_DETAILS_QUERY_EMPTY_TARGET_USER_ID",
			"ユーザー給与詳細検索条件の作成に失敗しました",
			nil,
		)
	}

	if req.Offset < 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_USER_SALARY_DETAILS_QUERY_INVALID_OFFSET",
			"ユーザー給与詳細検索条件の作成に失敗しました",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	if req.Limit <= 0 {
		return nil, nil, results.BadRequest(
			"BUILD_SEARCH_USER_SALARY_DETAILS_QUERY_INVALID_LIMIT",
			"ユーザー給与詳細検索条件の作成に失敗しました",
			map[string]any{
				"limit": req.Limit,
			},
		)
	}

	searchQuery := builder.db.Model(&models.UserSalaryDetail{})
	countQuery := builder.db.Model(&models.UserSalaryDetail{})

	searchQuery = applySearchUserSalaryDetailsCondition(searchQuery, req)
	countQuery = applySearchUserSalaryDetailsCondition(countQuery, req)

	searchQuery = searchQuery.
		Order("effective_from DESC").
		Order("id DESC").
		Offset(req.Offset).
		Limit(req.Limit)

	return searchQuery, countQuery, results.OK(
		nil,
		"BUILD_SEARCH_USER_SALARY_DETAILS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー給与詳細ID検索用クエリ作成
 *
 * 論理削除済みユーザー給与詳細は対象外にする。
 */
func (builder *userSalaryDetailBuilder) BuildFindUserSalaryDetailByIDQuery(userSalaryDetailID uint) (*gorm.DB, results.Result) {
	if userSalaryDetailID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_USER_SALARY_DETAIL_BY_ID_QUERY_INVALID_ID",
			"ユーザー給与詳細取得条件の作成に失敗しました",
			map[string]any{
				"userSalaryDetailId": userSalaryDetailID,
			},
		)
	}

	query := builder.db.
		Model(&models.UserSalaryDetail{}).
		Where("id = ?", userSalaryDetailID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_USER_SALARY_DETAIL_BY_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 同一ユーザーの適用期間重複確認クエリ作成
 *
 * 新規作成時に使う。
 * effective_to が NULL のレコードは、終了日なしとして扱う。
 */
func (builder *userSalaryDetailBuilder) BuildCountOverlapUserSalaryDetailsQuery(
	userID uint,
	effectiveFrom time.Time,
	effectiveTo *time.Time,
) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_COUNT_OVERLAP_USER_SALARY_DETAILS_QUERY_EMPTY_USER_ID",
			"ユーザー給与詳細の適用期間重複確認条件の作成に失敗しました",
			nil,
		)
	}

	if effectiveFrom.IsZero() {
		return nil, results.BadRequest(
			"BUILD_COUNT_OVERLAP_USER_SALARY_DETAILS_QUERY_EMPTY_EFFECTIVE_FROM",
			"ユーザー給与詳細の適用期間重複確認条件の作成に失敗しました",
			nil,
		)
	}

	query := builder.db.
		Model(&models.UserSalaryDetail{}).
		Where("user_id = ?", userID).
		Where("is_deleted = ?", false)

	query = applyUserSalaryEffectivePeriodOverlapCondition(query, effectiveFrom, effectiveTo)

	return query, results.OK(
		nil,
		"BUILD_COUNT_OVERLAP_USER_SALARY_DETAILS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 指定ユーザー給与詳細以外の同一ユーザー適用期間重複確認クエリ作成
 *
 * 更新時に使う。
 * effective_to が NULL のレコードは、終了日なしとして扱う。
 */
func (builder *userSalaryDetailBuilder) BuildCountOverlapUserSalaryDetailsExceptIDQuery(
	userID uint,
	userSalaryDetailID uint,
	effectiveFrom time.Time,
	effectiveTo *time.Time,
) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_COUNT_OVERLAP_USER_SALARY_DETAILS_EXCEPT_ID_QUERY_EMPTY_USER_ID",
			"ユーザー給与詳細の適用期間重複確認条件の作成に失敗しました",
			nil,
		)
	}

	if userSalaryDetailID == 0 {
		return nil, results.BadRequest(
			"BUILD_COUNT_OVERLAP_USER_SALARY_DETAILS_EXCEPT_ID_QUERY_INVALID_ID",
			"ユーザー給与詳細の適用期間重複確認条件の作成に失敗しました",
			map[string]any{
				"userSalaryDetailId": userSalaryDetailID,
			},
		)
	}

	if effectiveFrom.IsZero() {
		return nil, results.BadRequest(
			"BUILD_COUNT_OVERLAP_USER_SALARY_DETAILS_EXCEPT_ID_QUERY_EMPTY_EFFECTIVE_FROM",
			"ユーザー給与詳細の適用期間重複確認条件の作成に失敗しました",
			nil,
		)
	}

	query := builder.db.
		Model(&models.UserSalaryDetail{}).
		Where("user_id = ?", userID).
		Where("id <> ?", userSalaryDetailID).
		Where("is_deleted = ?", false)

	query = applyUserSalaryEffectivePeriodOverlapCondition(query, effectiveFrom, effectiveTo)

	return query, results.OK(
		nil,
		"BUILD_COUNT_OVERLAP_USER_SALARY_DETAILS_EXCEPT_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー給与詳細作成用Model作成
 */
func (builder *userSalaryDetailBuilder) BuildCreateUserSalaryDetailModel(
	req types.CreateUserSalaryDetailRequest,
	effectiveFrom time.Time,
	effectiveTo *time.Time,
) (models.UserSalaryDetail, results.Result) {
	if req.TargetUserID == 0 {
		return models.UserSalaryDetail{}, results.BadRequest(
			"BUILD_CREATE_USER_SALARY_DETAIL_MODEL_EMPTY_TARGET_USER_ID",
			"ユーザー給与詳細作成データの作成に失敗しました",
			nil,
		)
	}

	if strings.TrimSpace(req.SalaryType) == "" {
		return models.UserSalaryDetail{}, results.BadRequest(
			"BUILD_CREATE_USER_SALARY_DETAIL_MODEL_EMPTY_SALARY_TYPE",
			"ユーザー給与詳細作成データの作成に失敗しました",
			nil,
		)
	}

	if effectiveFrom.IsZero() {
		return models.UserSalaryDetail{}, results.BadRequest(
			"BUILD_CREATE_USER_SALARY_DETAIL_MODEL_EMPTY_EFFECTIVE_FROM",
			"ユーザー給与詳細作成データの作成に失敗しました",
			nil,
		)
	}

	userSalaryDetail := models.UserSalaryDetail{
		UserID: req.TargetUserID,

		SalaryType: strings.TrimSpace(req.SalaryType),

		BaseAmount: req.BaseAmount,

		ExtraAllowanceAmount: req.ExtraAllowanceAmount,
		ExtraAllowanceMemo:   req.ExtraAllowanceMemo,

		FixedDeductionAmount: req.FixedDeductionAmount,
		FixedDeductionMemo:   req.FixedDeductionMemo,

		IsPayrollTarget: req.IsPayrollTarget,

		EffectiveFrom: effectiveFrom,
		EffectiveTo:   effectiveTo,

		Memo: req.Memo,

		IsDeleted: false,
	}

	return userSalaryDetail, results.OK(
		nil,
		"BUILD_CREATE_USER_SALARY_DETAIL_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー給与詳細更新用Model作成
 */
func (builder *userSalaryDetailBuilder) BuildUpdateUserSalaryDetailModel(
	currentUserSalaryDetail models.UserSalaryDetail,
	req types.UpdateUserSalaryDetailRequest,
	effectiveFrom time.Time,
	effectiveTo *time.Time,
) (models.UserSalaryDetail, results.Result) {
	if currentUserSalaryDetail.ID == 0 {
		return models.UserSalaryDetail{}, results.BadRequest(
			"BUILD_UPDATE_USER_SALARY_DETAIL_MODEL_EMPTY_CURRENT_USER_SALARY_DETAIL",
			"ユーザー給与詳細更新データの作成に失敗しました",
			nil,
		)
	}

	if strings.TrimSpace(req.SalaryType) == "" {
		return models.UserSalaryDetail{}, results.BadRequest(
			"BUILD_UPDATE_USER_SALARY_DETAIL_MODEL_EMPTY_SALARY_TYPE",
			"ユーザー給与詳細更新データの作成に失敗しました",
			nil,
		)
	}

	if effectiveFrom.IsZero() {
		return models.UserSalaryDetail{}, results.BadRequest(
			"BUILD_UPDATE_USER_SALARY_DETAIL_MODEL_EMPTY_EFFECTIVE_FROM",
			"ユーザー給与詳細更新データの作成に失敗しました",
			nil,
		)
	}

	currentUserSalaryDetail.SalaryType = strings.TrimSpace(req.SalaryType)

	currentUserSalaryDetail.BaseAmount = req.BaseAmount

	currentUserSalaryDetail.ExtraAllowanceAmount = req.ExtraAllowanceAmount
	currentUserSalaryDetail.ExtraAllowanceMemo = req.ExtraAllowanceMemo

	currentUserSalaryDetail.FixedDeductionAmount = req.FixedDeductionAmount
	currentUserSalaryDetail.FixedDeductionMemo = req.FixedDeductionMemo

	currentUserSalaryDetail.IsPayrollTarget = req.IsPayrollTarget

	currentUserSalaryDetail.EffectiveFrom = effectiveFrom
	currentUserSalaryDetail.EffectiveTo = effectiveTo

	currentUserSalaryDetail.Memo = req.Memo

	return currentUserSalaryDetail, results.OK(
		nil,
		"BUILD_UPDATE_USER_SALARY_DETAIL_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー給与詳細論理削除用Model作成
 */
func (builder *userSalaryDetailBuilder) BuildDeleteUserSalaryDetailModel(currentUserSalaryDetail models.UserSalaryDetail) (models.UserSalaryDetail, results.Result) {
	if currentUserSalaryDetail.ID == 0 {
		return models.UserSalaryDetail{}, results.BadRequest(
			"BUILD_DELETE_USER_SALARY_DETAIL_MODEL_EMPTY_CURRENT_USER_SALARY_DETAIL",
			"ユーザー給与詳細削除データの作成に失敗しました",
			nil,
		)
	}

	now := time.Now()

	currentUserSalaryDetail.IsDeleted = true
	currentUserSalaryDetail.DeletedAt = gorm.DeletedAt{
		Time:  now,
		Valid: true,
	}

	return currentUserSalaryDetail, results.OK(
		nil,
		"BUILD_DELETE_USER_SALARY_DETAIL_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザー給与詳細検索条件をGORMクエリへ適用する
 */
func applySearchUserSalaryDetailsCondition(query *gorm.DB, req types.SearchUserSalaryDetailsRequest) *gorm.DB {
	query = query.Where("user_id = ?", req.TargetUserID)

	if !req.IncludeDeleted {
		query = query.Where("is_deleted = ?", false)
	}

	return query
}

/*
 * 適用期間重複条件をGORMクエリへ適用する
 *
 * 重複判定：
 * ・既存開始日 <= 新規終了日
 * ・既存終了日がNULL、または既存終了日 >= 新規開始日
 *
 * 新規終了日がNULLの場合：
 * ・既存終了日がNULL、または既存終了日 >= 新規開始日
 */
func applyUserSalaryEffectivePeriodOverlapCondition(query *gorm.DB, effectiveFrom time.Time, effectiveTo *time.Time) *gorm.DB {
	if effectiveTo == nil {
		return query.Where("effective_to IS NULL OR effective_to >= ?", effectiveFrom)
	}

	return query.
		Where("effective_from <= ?", *effectiveTo).
		Where("effective_to IS NULL OR effective_to >= ?", effectiveFrom)
}
