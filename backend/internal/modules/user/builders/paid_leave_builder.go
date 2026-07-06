package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type PaidLeaveBuilder interface {
	BuildFindActiveUserByIDQuery(userID uint) (*gorm.DB, results.Result)
	BuildSumActivePaidLeaveUsageDaysByUserIDQuery(userID uint) (*gorm.DB, results.Result)
	BuildFindAutomaticPaidLeaveUsageByUserIDAndUsageDateQuery(userID uint, usageDate time.Time) (*gorm.DB, results.Result)
	BuildCreateAutomaticPaidLeaveUsageModel(userID uint, usageDate time.Time) (models.PaidLeaveUsage, results.Result)
	BuildActivateAutomaticPaidLeaveUsageModel(currentPaidLeaveUsage models.PaidLeaveUsage) (models.PaidLeaveUsage, results.Result)
	BuildDeleteAutomaticPaidLeaveUsageModel(currentPaidLeaveUsage models.PaidLeaveUsage) (models.PaidLeaveUsage, results.Result)
}

/*
 * 従業員用有給Builder
 *
 * 役割：
 * ・Serviceから受け取ったログインユーザーIDをもとにGORMクエリを作成する
 * ・勤怠画面から保存される有給使用履歴のModelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Sum / Create / Save はRepositoryに任せる
 * ・従業員APIなので targetUserId は使わない
 * ・勤怠画面由来の有給使用履歴は IsManual=false とする
 */
type paidLeaveBuilder struct {
	db *gorm.DB
}

func NewPaidLeaveBuilder(db *gorm.DB) PaidLeaveBuilder {
	return &paidLeaveBuilder{db: db}
}

func (builder *paidLeaveBuilder) BuildFindActiveUserByIDQuery(userID uint) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ACTIVE_USER_BY_ID_QUERY_INVALID_USER_ID",
			"ユーザー取得条件の作成に失敗しました",
			map[string]any{"userId": userID},
		)
	}

	query := builder.db.
		Model(&models.User{}).
		Where("id = ?", userID).
		Where("is_deleted = ?", false)

	return query, results.OK(nil, "BUILD_FIND_ACTIVE_USER_BY_ID_QUERY_SUCCESS", "", nil)
}

func (builder *paidLeaveBuilder) BuildSumActivePaidLeaveUsageDaysByUserIDQuery(userID uint) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_SUM_ACTIVE_PAID_LEAVE_USAGE_DAYS_BY_USER_ID_QUERY_INVALID_USER_ID",
			"有給使用日数合計条件の作成に失敗しました",
			map[string]any{"userId": userID},
		)
	}

	query := builder.db.
		Model(&models.PaidLeaveUsage{}).
		Where("user_id = ?", userID).
		Where("is_deleted = ?", false)

	return query, results.OK(nil, "BUILD_SUM_ACTIVE_PAID_LEAVE_USAGE_DAYS_BY_USER_ID_QUERY_SUCCESS", "", nil)
}

/*
 * 勤怠画面由来の有給使用履歴検索用クエリ
 *
 * 論理削除済みも取得対象にする。
 * 再び有給へ変更された場合に、既存データを復活させるため。
 */
func (builder *paidLeaveBuilder) BuildFindAutomaticPaidLeaveUsageByUserIDAndUsageDateQuery(
	userID uint,
	usageDate time.Time,
) (*gorm.DB, results.Result) {
	if userID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_AUTOMATIC_PAID_LEAVE_USAGE_INVALID_USER_ID",
			"勤怠連携用有給使用日取得条件の作成に失敗しました",
			map[string]any{"userId": userID},
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
		Where("user_id = ?", userID).
		Where("usage_date = ?", usageDate).
		Where("is_manual = ?", false).
		Order("id DESC")

	return query, results.OK(nil, "BUILD_FIND_AUTOMATIC_PAID_LEAVE_USAGE_SUCCESS", "", nil)
}

func (builder *paidLeaveBuilder) BuildCreateAutomaticPaidLeaveUsageModel(
	userID uint,
	usageDate time.Time,
) (models.PaidLeaveUsage, results.Result) {
	if userID == 0 {
		return models.PaidLeaveUsage{}, results.BadRequest(
			"BUILD_CREATE_AUTOMATIC_PAID_LEAVE_USAGE_INVALID_USER_ID",
			"勤怠連携用有給使用日作成データの作成に失敗しました",
			map[string]any{"userId": userID},
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
		UserID:    userID,
		UsageDate: usageDate,
		UsageDays: 1.0,
		IsManual:  false,
		Memo:      "月次勤怠全体保存から登録",
		IsDeleted: false,
		DeletedAt: nil,
	}

	return paidLeaveUsage, results.OK(nil, "BUILD_CREATE_AUTOMATIC_PAID_LEAVE_USAGE_SUCCESS", "", nil)
}

func (builder *paidLeaveBuilder) BuildActivateAutomaticPaidLeaveUsageModel(
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
			map[string]any{"paidLeaveUsageId": currentPaidLeaveUsage.ID},
		)
	}

	currentPaidLeaveUsage.UsageDays = 1.0
	currentPaidLeaveUsage.IsDeleted = false
	currentPaidLeaveUsage.DeletedAt = nil

	return currentPaidLeaveUsage, results.OK(nil, "BUILD_ACTIVATE_AUTOMATIC_PAID_LEAVE_USAGE_SUCCESS", "", nil)
}

func (builder *paidLeaveBuilder) BuildDeleteAutomaticPaidLeaveUsageModel(
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
			map[string]any{"paidLeaveUsageId": currentPaidLeaveUsage.ID},
		)
	}

	now := time.Now()
	currentPaidLeaveUsage.IsDeleted = true
	currentPaidLeaveUsage.DeletedAt = &now

	return currentPaidLeaveUsage, results.OK(nil, "BUILD_DELETE_AUTOMATIC_PAID_LEAVE_USAGE_SUCCESS", "", nil)
}
