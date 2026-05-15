package services

import (
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"

	"gorm.io/gorm"
)

/*
 * 管理者用ユーザー給与詳細Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 */
type UserSalaryDetailService interface {
	SearchUserSalaryDetails(req types.SearchUserSalaryDetailsRequest) results.Result
	GetUserSalaryDetail(req types.GetUserSalaryDetailRequest) results.Result
	CreateUserSalaryDetail(req types.CreateUserSalaryDetailRequest) results.Result
	UpdateUserSalaryDetail(req types.UpdateUserSalaryDetailRequest) results.Result
	DeleteUserSalaryDetail(req types.DeleteUserSalaryDetailRequest) results.Result
}

/*
 * 管理者用ユーザー給与詳細Service
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
 * ・日付文字列はServiceでtime.Time / *time.Timeへ変換する
 */
type userSalaryDetailService struct {
	userSalaryDetailBuilder    builders.UserSalaryDetailBuilder
	userSalaryDetailRepository repositories.UserSalaryDetailRepository
}

/*
 * UserSalaryDetailService生成
 */
func NewUserSalaryDetailService(
	userSalaryDetailBuilder builders.UserSalaryDetailBuilder,
	userSalaryDetailRepository repositories.UserSalaryDetailRepository,
) UserSalaryDetailService {
	return &userSalaryDetailService{
		userSalaryDetailBuilder:    userSalaryDetailBuilder,
		userSalaryDetailRepository: userSalaryDetailRepository,
	}
}

/*
 * models.UserSalaryDetailをフロント返却用UserSalaryDetailResponseへ変換する
 *
 * EffectiveFrom / EffectiveTo は yyyy-MM-dd 形式で返す。
 * CreatedAt / UpdatedAt / DeletedAt は RFC3339 形式で返す。
 */
func toUserSalaryDetailResponse(userSalaryDetail models.UserSalaryDetail) types.UserSalaryDetailResponse {
	return types.UserSalaryDetailResponse{
		ID: userSalaryDetail.ID,

		UserID: userSalaryDetail.UserID,

		SalaryType: userSalaryDetail.SalaryType,

		BaseAmount: userSalaryDetail.BaseAmount,

		ExtraAllowanceAmount: userSalaryDetail.ExtraAllowanceAmount,
		ExtraAllowanceMemo:   userSalaryDetail.ExtraAllowanceMemo,

		FixedDeductionAmount: userSalaryDetail.FixedDeductionAmount,
		FixedDeductionMemo:   userSalaryDetail.FixedDeductionMemo,

		IsPayrollTarget: userSalaryDetail.IsPayrollTarget,

		EffectiveFrom: userSalaryFormatDate(userSalaryDetail.EffectiveFrom),
		EffectiveTo:   userSalaryFormatOptionalDate(userSalaryDetail.EffectiveTo),

		Memo: userSalaryDetail.Memo,

		IsDeleted: userSalaryDetail.IsDeleted,
		CreatedAt: userSalaryFormatDateTime(userSalaryDetail.CreatedAt),
		UpdatedAt: userSalaryFormatDateTime(userSalaryDetail.UpdatedAt),
		DeletedAt: userSalaryFormatDeletedAt(userSalaryDetail.DeletedAt),
	}
}

/*
 * 日付を yyyy-MM-dd 形式に変換する
 */
func userSalaryFormatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.Format("2006-01-02")
}

/*
 * 任意日付を yyyy-MM-dd 形式に変換する
 */
func userSalaryFormatOptionalDate(value *time.Time) *string {
	if value == nil || value.IsZero() {
		return nil
	}

	formattedValue := value.Format("2006-01-02")
	return &formattedValue
}

/*
 * 日時をRFC3339形式に変換する
 */
func userSalaryFormatDateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.Format(time.RFC3339)
}

/*
 * gorm.DeletedAtをRFC3339形式に変換する
 */
func userSalaryFormatDeletedAt(value gorm.DeletedAt) *string {
	if !value.Valid {
		return nil
	}

	formattedValue := value.Time.Format(time.RFC3339)
	return &formattedValue
}

/*
 * yyyy-MM-dd形式の日付文字列をtime.Timeへ変換する
 */
func parseUserSalaryDate(value string, fieldName string, errorCode string, errorMessage string) (time.Time, results.Result) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return time.Time{}, results.BadRequest(
			errorCode,
			errorMessage,
			map[string]any{
				"field": fieldName,
			},
		)
	}

	parsedValue, err := time.Parse("2006-01-02", trimmedValue)
	if err != nil {
		return time.Time{}, results.BadRequest(
			errorCode,
			errorMessage,
			map[string]any{
				"field": fieldName,
				"value": value,
			},
		)
	}

	return parsedValue, results.OK(nil, "", "", nil)
}

/*
 * yyyy-MM-dd形式の任意日付文字列を*time.Timeへ変換する
 */
func parseOptionalUserSalaryDate(value *string, fieldName string, errorCode string, errorMessage string) (*time.Time, results.Result) {
	if value == nil {
		return nil, results.OK(nil, "", "", nil)
	}

	trimmedValue := strings.TrimSpace(*value)
	if trimmedValue == "" {
		return nil, results.OK(nil, "", "", nil)
	}

	parsedValue, err := time.Parse("2006-01-02", trimmedValue)
	if err != nil {
		return nil, results.BadRequest(
			errorCode,
			errorMessage,
			map[string]any{
				"field": fieldName,
				"value": *value,
			},
		)
	}

	return &parsedValue, results.OK(nil, "", "", nil)
}

/*
 * 給与区分を検証する
 */
func validateUserSalaryType(salaryType string, errorCode string) results.Result {
	switch strings.TrimSpace(salaryType) {
	case types.SalaryTypeMonthly, types.SalaryTypeHourly, types.SalaryTypeDaily:
		return results.OK(nil, "", "", nil)
	default:
		return results.BadRequest(
			errorCode,
			"給与区分が正しくありません",
			map[string]any{
				"salaryType": salaryType,
				"allowed": []string{
					types.SalaryTypeMonthly,
					types.SalaryTypeHourly,
					types.SalaryTypeDaily,
				},
			},
		)
	}
}

/*
 * 金額項目を検証する
 */
func validateUserSalaryAmounts(
	baseAmount int,
	extraAllowanceAmount int,
	fixedDeductionAmount int,
	errorCode string,
) results.Result {
	if baseAmount < 0 {
		return results.BadRequest(
			errorCode,
			"基本金額は0円以上で入力してください",
			map[string]any{
				"field": "baseAmount",
				"value": baseAmount,
			},
		)
	}

	if extraAllowanceAmount < 0 {
		return results.BadRequest(
			errorCode,
			"その他固定手当は0円以上で入力してください",
			map[string]any{
				"field": "extraAllowanceAmount",
				"value": extraAllowanceAmount,
			},
		)
	}

	if fixedDeductionAmount < 0 {
		return results.BadRequest(
			errorCode,
			"その他固定控除は0円以上で入力してください",
			map[string]any{
				"field": "fixedDeductionAmount",
				"value": fixedDeductionAmount,
			},
		)
	}

	return results.OK(nil, "", "", nil)
}

/*
 * 適用期間を検証する
 */
func validateUserSalaryEffectivePeriod(
	effectiveFrom time.Time,
	effectiveTo *time.Time,
	errorCode string,
) results.Result {
	if effectiveTo == nil {
		return results.OK(nil, "", "", nil)
	}

	if effectiveTo.Before(effectiveFrom) {
		return results.BadRequest(
			errorCode,
			"適用終了日は適用開始日以降の日付を入力してください",
			map[string]any{
				"effectiveFrom": userSalaryFormatDate(effectiveFrom),
				"effectiveTo":   userSalaryFormatDate(*effectiveTo),
			},
		)
	}

	return results.OK(nil, "", "", nil)
}

/*
 * 検索
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
func (service *userSalaryDetailService) SearchUserSalaryDetails(req types.SearchUserSalaryDetailsRequest) results.Result {
	if req.TargetUserID == 0 {
		return results.BadRequest(
			"SEARCH_USER_SALARY_DETAILS_EMPTY_TARGET_USER_ID",
			"対象ユーザーが指定されていません",
			nil,
		)
	}

	// ページング検索条件を共通関数で正規化する
	normalizedCondition, normalizeResult := utils.NormalizePageSearchCondition(
		utils.PageSearchCondition{
			Keyword: "",
			Offset:  req.Offset,
			Limit:   req.Limit,
		},
		"SEARCH_USER_SALARY_DETAILS_INVALID_OFFSET",
		"検索開始位置が正しくありません",
	)
	if normalizeResult.Error {
		return normalizeResult
	}

	req.Offset = normalizedCondition.Offset
	req.Limit = normalizedCondition.Limit

	// Builderで一覧検索用クエリと件数取得用クエリを作成する
	searchQuery, countQuery, buildResult := service.userSalaryDetailBuilder.BuildSearchUserSalaryDetailsQuery(req)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryでユーザー給与詳細一覧を取得する
	userSalaryDetails, findResult := service.userSalaryDetailRepository.FindUserSalaryDetails(searchQuery)
	if findResult.Error {
		return findResult
	}

	// Repositoryで検索条件に一致する総件数を取得する
	total, countResult := service.userSalaryDetailRepository.CountUserSalaryDetails(countQuery)
	if countResult.Error {
		return countResult
	}

	// DBモデルをフロント返却用Responseへ変換する
	userSalaryDetailResponses := make([]types.UserSalaryDetailResponse, 0, len(userSalaryDetails))
	for _, userSalaryDetail := range userSalaryDetails {
		userSalaryDetailResponses = append(userSalaryDetailResponses, toUserSalaryDetailResponse(userSalaryDetail))
	}

	hasMore := utils.HasMore(total, req.Offset, len(userSalaryDetails))

	return results.OK(
		types.SearchUserSalaryDetailsResponse{
			UserSalaryDetails: userSalaryDetailResponses,
			HasMore:           hasMore,
		},
		"SEARCH_USER_SALARY_DETAILS_SUCCESS",
		"ユーザー給与詳細一覧を取得しました",
		nil,
	)
}

/*
 * 単体情報取得
 */
func (service *userSalaryDetailService) GetUserSalaryDetail(req types.GetUserSalaryDetailRequest) results.Result {
	if req.UserSalaryDetailID == 0 {
		return results.BadRequest(
			"GET_USER_SALARY_DETAIL_EMPTY_ID",
			"取得対象のユーザー給与詳細が指定されていません",
			nil,
		)
	}

	// Builderで単体取得用クエリを作成する
	query, buildResult := service.userSalaryDetailBuilder.BuildFindUserSalaryDetailByIDQuery(req.UserSalaryDetailID)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryでユーザー給与詳細を取得する
	userSalaryDetail, findResult := service.userSalaryDetailRepository.FindUserSalaryDetail(query)
	if findResult.Error {
		return findResult
	}

	return results.OK(
		types.GetUserSalaryDetailResponse{
			UserSalaryDetail: toUserSalaryDetailResponse(userSalaryDetail),
		},
		"GET_USER_SALARY_DETAIL_SUCCESS",
		"ユーザー給与詳細を取得しました",
		nil,
	)
}

/*
 * 新規作成
 */
func (service *userSalaryDetailService) CreateUserSalaryDetail(req types.CreateUserSalaryDetailRequest) results.Result {
	if req.TargetUserID == 0 {
		return results.BadRequest(
			"CREATE_USER_SALARY_DETAIL_EMPTY_TARGET_USER_ID",
			"対象ユーザーが指定されていません",
			nil,
		)
	}

	if validateResult := validateUserSalaryType(req.SalaryType, "CREATE_USER_SALARY_DETAIL_INVALID_SALARY_TYPE"); validateResult.Error {
		return validateResult
	}

	if validateResult := validateUserSalaryAmounts(
		req.BaseAmount,
		req.ExtraAllowanceAmount,
		req.FixedDeductionAmount,
		"CREATE_USER_SALARY_DETAIL_INVALID_AMOUNT",
	); validateResult.Error {
		return validateResult
	}

	effectiveFrom, parseEffectiveFromResult := parseUserSalaryDate(
		req.EffectiveFrom,
		"effectiveFrom",
		"CREATE_USER_SALARY_DETAIL_INVALID_EFFECTIVE_FROM",
		"適用開始日が正しくありません",
	)
	if parseEffectiveFromResult.Error {
		return parseEffectiveFromResult
	}

	effectiveTo, parseEffectiveToResult := parseOptionalUserSalaryDate(
		req.EffectiveTo,
		"effectiveTo",
		"CREATE_USER_SALARY_DETAIL_INVALID_EFFECTIVE_TO",
		"適用終了日が正しくありません",
	)
	if parseEffectiveToResult.Error {
		return parseEffectiveToResult
	}

	if validateResult := validateUserSalaryEffectivePeriod(
		effectiveFrom,
		effectiveTo,
		"CREATE_USER_SALARY_DETAIL_INVALID_EFFECTIVE_PERIOD",
	); validateResult.Error {
		return validateResult
	}

	// Builderで同一ユーザーの適用期間重複確認用クエリを作成する
	overlapCountQuery, buildOverlapCountResult := service.userSalaryDetailBuilder.BuildCountOverlapUserSalaryDetailsQuery(
		req.TargetUserID,
		effectiveFrom,
		effectiveTo,
	)
	if buildOverlapCountResult.Error {
		return buildOverlapCountResult
	}

	// Repositoryで適用期間重複確認を実行する
	overlapCount, overlapCountResult := service.userSalaryDetailRepository.CountUserSalaryDetails(overlapCountQuery)
	if overlapCountResult.Error {
		return overlapCountResult
	}

	if overlapCount > 0 {
		return results.Conflict(
			"CREATE_USER_SALARY_DETAIL_EFFECTIVE_PERIOD_OVERLAP",
			"同じユーザーで適用期間が重なる給与詳細が既に登録されています",
			map[string]any{
				"targetUserId":  req.TargetUserID,
				"effectiveFrom": userSalaryFormatDate(effectiveFrom),
				"effectiveTo":   userSalaryFormatOptionalDate(effectiveTo),
				"overlapCount":  overlapCount,
			},
		)
	}

	// Builderで作成用Modelを作る
	userSalaryDetail, buildUserSalaryDetailResult := service.userSalaryDetailBuilder.BuildCreateUserSalaryDetailModel(
		req,
		effectiveFrom,
		effectiveTo,
	)
	if buildUserSalaryDetailResult.Error {
		return buildUserSalaryDetailResult
	}

	// Repositoryでユーザー給与詳細を作成する
	createdUserSalaryDetail, createResult := service.userSalaryDetailRepository.CreateUserSalaryDetail(userSalaryDetail)
	if createResult.Error {
		return createResult
	}

	return results.Created(
		types.CreateUserSalaryDetailResponse{
			UserSalaryDetail: toUserSalaryDetailResponse(createdUserSalaryDetail),
		},
		"CREATE_USER_SALARY_DETAIL_SUCCESS",
		"ユーザー給与詳細を作成しました",
		nil,
	)
}

/*
 * 更新
 */
func (service *userSalaryDetailService) UpdateUserSalaryDetail(req types.UpdateUserSalaryDetailRequest) results.Result {
	if req.UserSalaryDetailID == 0 {
		return results.BadRequest(
			"UPDATE_USER_SALARY_DETAIL_EMPTY_ID",
			"更新対象のユーザー給与詳細が指定されていません",
			nil,
		)
	}

	if validateResult := validateUserSalaryType(req.SalaryType, "UPDATE_USER_SALARY_DETAIL_INVALID_SALARY_TYPE"); validateResult.Error {
		return validateResult
	}

	if validateResult := validateUserSalaryAmounts(
		req.BaseAmount,
		req.ExtraAllowanceAmount,
		req.FixedDeductionAmount,
		"UPDATE_USER_SALARY_DETAIL_INVALID_AMOUNT",
	); validateResult.Error {
		return validateResult
	}

	effectiveFrom, parseEffectiveFromResult := parseUserSalaryDate(
		req.EffectiveFrom,
		"effectiveFrom",
		"UPDATE_USER_SALARY_DETAIL_INVALID_EFFECTIVE_FROM",
		"適用開始日が正しくありません",
	)
	if parseEffectiveFromResult.Error {
		return parseEffectiveFromResult
	}

	effectiveTo, parseEffectiveToResult := parseOptionalUserSalaryDate(
		req.EffectiveTo,
		"effectiveTo",
		"UPDATE_USER_SALARY_DETAIL_INVALID_EFFECTIVE_TO",
		"適用終了日が正しくありません",
	)
	if parseEffectiveToResult.Error {
		return parseEffectiveToResult
	}

	if validateResult := validateUserSalaryEffectivePeriod(
		effectiveFrom,
		effectiveTo,
		"UPDATE_USER_SALARY_DETAIL_INVALID_EFFECTIVE_PERIOD",
	); validateResult.Error {
		return validateResult
	}

	// Builderで対象ユーザー給与詳細取得用クエリを作成する
	findQuery, buildFindResult := service.userSalaryDetailBuilder.BuildFindUserSalaryDetailByIDQuery(req.UserSalaryDetailID)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象ユーザー給与詳細を取得する
	currentUserSalaryDetail, findResult := service.userSalaryDetailRepository.FindUserSalaryDetail(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderで同一ユーザーの適用期間重複確認用クエリを作成する
	overlapCountQuery, buildOverlapCountResult := service.userSalaryDetailBuilder.BuildCountOverlapUserSalaryDetailsExceptIDQuery(
		currentUserSalaryDetail.UserID,
		req.UserSalaryDetailID,
		effectiveFrom,
		effectiveTo,
	)
	if buildOverlapCountResult.Error {
		return buildOverlapCountResult
	}

	// Repositoryで適用期間重複確認を実行する
	overlapCount, overlapCountResult := service.userSalaryDetailRepository.CountUserSalaryDetails(overlapCountQuery)
	if overlapCountResult.Error {
		return overlapCountResult
	}

	if overlapCount > 0 {
		return results.Conflict(
			"UPDATE_USER_SALARY_DETAIL_EFFECTIVE_PERIOD_OVERLAP",
			"同じユーザーで適用期間が重なる給与詳細が既に登録されています",
			map[string]any{
				"userSalaryDetailId": req.UserSalaryDetailID,
				"userId":             currentUserSalaryDetail.UserID,
				"effectiveFrom":      userSalaryFormatDate(effectiveFrom),
				"effectiveTo":        userSalaryFormatOptionalDate(effectiveTo),
				"overlapCount":       overlapCount,
			},
		)
	}

	// Builderで更新用Modelを作る
	updatedUserSalaryDetail, buildUpdateResult := service.userSalaryDetailBuilder.BuildUpdateUserSalaryDetailModel(
		currentUserSalaryDetail,
		req,
		effectiveFrom,
		effectiveTo,
	)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	// Repositoryでユーザー給与詳細を更新する
	savedUserSalaryDetail, saveResult := service.userSalaryDetailRepository.SaveUserSalaryDetail(updatedUserSalaryDetail)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.UpdateUserSalaryDetailResponse{
			UserSalaryDetail: toUserSalaryDetailResponse(savedUserSalaryDetail),
		},
		"UPDATE_USER_SALARY_DETAIL_SUCCESS",
		"ユーザー給与詳細を更新しました",
		nil,
	)
}

/*
 * 論理削除
 */
func (service *userSalaryDetailService) DeleteUserSalaryDetail(req types.DeleteUserSalaryDetailRequest) results.Result {
	if req.UserSalaryDetailID == 0 {
		return results.BadRequest(
			"DELETE_USER_SALARY_DETAIL_EMPTY_ID",
			"削除対象のユーザー給与詳細が指定されていません",
			nil,
		)
	}

	// Builderで対象ユーザー給与詳細取得用クエリを作成する
	findQuery, buildFindResult := service.userSalaryDetailBuilder.BuildFindUserSalaryDetailByIDQuery(req.UserSalaryDetailID)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象ユーザー給与詳細を取得する
	currentUserSalaryDetail, findResult := service.userSalaryDetailRepository.FindUserSalaryDetail(findQuery)
	if findResult.Error {
		return findResult
	}

	// Builderで論理削除用Modelを作る
	deletedUserSalaryDetail, buildDeleteResult := service.userSalaryDetailBuilder.BuildDeleteUserSalaryDetailModel(currentUserSalaryDetail)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	// Repositoryでユーザー給与詳細を保存する
	_, saveResult := service.userSalaryDetailRepository.SaveUserSalaryDetail(deletedUserSalaryDetail)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteUserSalaryDetailResponse{
			UserSalaryDetailID: req.UserSalaryDetailID,
		},
		"DELETE_USER_SALARY_DETAIL_SUCCESS",
		"ユーザー給与詳細を削除しました",
		nil,
	)
}
