package builders

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type PaidLeaveBuilder interface {
	BuildFindActiveUserByIDQuery(userID uint) (*gorm.DB, results.Result)
	BuildSumActivePaidLeaveUsageDaysByUserIDQuery(userID uint) (*gorm.DB, results.Result)
}

/*
 * 従業員用有給Builder
 *
 * 役割：
 * ・Serviceから受け取ったログインユーザーIDをもとにGORMクエリを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Sum はRepositoryに任せる
 * ・従業員APIなので targetUserId は使わない
 */
type paidLeaveBuilder struct {
	db *gorm.DB
}

/*
 * PaidLeaveBuilder生成
 */
func NewPaidLeaveBuilder(db *gorm.DB) PaidLeaveBuilder {
	return &paidLeaveBuilder{db: db}
}

/*
 * 有効ユーザーID検索用クエリ作成
 *
 * ログイン中ユーザーの有給残数計算で、
 * 入社日を取得するために使う。
 *
 * 論理削除済みユーザーは対象外。
 */
func (builder *paidLeaveBuilder) BuildFindActiveUserByIDQuery(userID uint) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ACTIVE_USER_BY_ID_QUERY_INVALID_USER_ID",
			"ユーザー取得条件の作成に失敗しました",
			map[string]any{
				"userId": userID,
			},
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("id = ?", userID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_ACTIVE_USER_BY_ID_QUERY_SUCCESS",
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
 * ・ログイン中ユーザーの有給使用日
 * ・論理削除されていないもの
 */
func (builder *paidLeaveBuilder) BuildSumActivePaidLeaveUsageDaysByUserIDQuery(userID uint) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_SUM_ACTIVE_PAID_LEAVE_USAGE_DAYS_BY_USER_ID_QUERY_INVALID_USER_ID",
			"有給使用日数合計条件の作成に失敗しました",
			map[string]any{
				"userId": userID,
			},
		)
	}

	query := builder.db.
		Model(&models.PaidLeaveUsage{}).
		Where("user_id = ?", userID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_SUM_ACTIVE_PAID_LEAVE_USAGE_DAYS_BY_USER_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}
