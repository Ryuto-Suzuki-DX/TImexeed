package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type MonthlyAttendanceRequestBuilder interface {
	BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
		targetUserID uint,
		targetYear int,
		targetMonth int,
	) (*gorm.DB, results.Result)

	BuildFindMonthlyAttendanceRequestByIDQuery(
		targetRequestID uint,
	) (*gorm.DB, results.Result)

	BuildCreateMonthlyAttendanceRequestModel(
		req types.SubmitMonthlyAttendanceRequestRequest,
		now time.Time,
	) (models.MonthlyAttendanceRequest, results.Result)

	BuildResubmitMonthlyAttendanceRequestModel(
		currentMonthlyAttendanceRequest models.MonthlyAttendanceRequest,
		req types.SubmitMonthlyAttendanceRequestRequest,
		now time.Time,
	) (models.MonthlyAttendanceRequest, results.Result)

	BuildCancelMonthlyAttendanceRequestModel(
		currentMonthlyAttendanceRequest models.MonthlyAttendanceRequest,
		req types.CancelMonthlyAttendanceRequestRequest,
		now time.Time,
	) (models.MonthlyAttendanceRequest, results.Result)

	BuildApproveMonthlyAttendanceRequestModel(
		currentMonthlyAttendanceRequest models.MonthlyAttendanceRequest,
		loginAdminID uint,
		now time.Time,
	) (models.MonthlyAttendanceRequest, results.Result)

	BuildRejectMonthlyAttendanceRequestModel(
		currentMonthlyAttendanceRequest models.MonthlyAttendanceRequest,
		req types.RejectMonthlyAttendanceRequestRequest,
		now time.Time,
	) (models.MonthlyAttendanceRequest, results.Result)
}

/*
 * 管理者用月次勤怠申請Builder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Serviceから受け取ったRequestをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * このBuilderで扱うもの：
 * ・対象ユーザー、対象年月の月次勤怠申請取得クエリ作成
 * ・月次勤怠申請IDによる取得クエリ作成
 * ・月次勤怠申請の新規作成Model作成
 * ・月次勤怠申請の再申請Model作成
 * ・月次勤怠申請の取り下げModel作成
 * ・月次勤怠申請の承認Model作成
 * ・月次勤怠申請の否認Model作成
 *
 * このBuilderで扱わないもの：
 * ・DB実行
 * ・申請可否の判定
 * ・取り下げ可否の判定
 * ・承認可否の判定
 * ・否認可否の判定
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Create / Save はRepositoryに任せる
 * ・対象年月の基本バリデーションはServiceでも行う
 * ・BuilderではModel作成に必要な最低限の検証を行う
 * ・管理者側では対象ユーザーIDを request body の targetUserId で受け取る
 */
type monthlyAttendanceRequestBuilder struct {
	db *gorm.DB
}

/*
 * MonthlyAttendanceRequestBuilder生成
 */
func NewMonthlyAttendanceRequestBuilder(db *gorm.DB) MonthlyAttendanceRequestBuilder {
	return &monthlyAttendanceRequestBuilder{db: db}
}

/*
 * ユーザーID + 対象年月で月次勤怠申請1件取得用クエリ作成
 *
 * 状態取得・申請・取り下げ時に使う。
 *
 * 注意：
 * ・targetUserID は管理者が選択した対象ユーザーID
 * ・論理削除済みの月次勤怠申請は対象外
 */
func (builder *monthlyAttendanceRequestBuilder) BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
	targetUserID uint,
	targetYear int,
	targetMonth int,
) (*gorm.DB, results.Result) {
	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MONTHLY_ATTENDANCE_REQUEST_QUERY_INVALID_TARGET_USER_ID",
			"月次勤怠申請取得条件の作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	if targetYear <= 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MONTHLY_ATTENDANCE_REQUEST_QUERY_INVALID_TARGET_YEAR",
			"月次勤怠申請取得条件の作成に失敗しました",
			map[string]any{
				"targetYear": targetYear,
			},
		)
	}

	if targetMonth < 1 || targetMonth > 12 {
		return nil, results.BadRequest(
			"BUILD_FIND_MONTHLY_ATTENDANCE_REQUEST_QUERY_INVALID_TARGET_MONTH",
			"月次勤怠申請取得条件の作成に失敗しました",
			map[string]any{
				"targetMonth": targetMonth,
			},
		)
	}

	query := builder.db.
		Model(&models.MonthlyAttendanceRequest{}).
		Where("user_id = ?", targetUserID).
		Where("target_year = ?", targetYear).
		Where("target_month = ?", targetMonth).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_MONTHLY_ATTENDANCE_REQUEST_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次勤怠申請IDで1件取得用クエリ作成
 *
 * 承認・否認時に使う。
 */
func (builder *monthlyAttendanceRequestBuilder) BuildFindMonthlyAttendanceRequestByIDQuery(
	targetRequestID uint,
) (*gorm.DB, results.Result) {
	if targetRequestID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_MONTHLY_ATTENDANCE_REQUEST_BY_ID_QUERY_INVALID_TARGET_REQUEST_ID",
			"月次勤怠申請取得条件の作成に失敗しました",
			map[string]any{
				"targetRequestId": targetRequestID,
			},
		)
	}

	query := builder.db.
		Model(&models.MonthlyAttendanceRequest{}).
		Where("id = ?", targetRequestID).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_MONTHLY_ATTENDANCE_REQUEST_BY_ID_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次勤怠申請 新規作成用Model作成
 *
 * 未申請、つまりMonthlyAttendanceRequestが存在しない場合に使う。
 *
 * 作成時は必ず PENDING として保存する。
 */
func (builder *monthlyAttendanceRequestBuilder) BuildCreateMonthlyAttendanceRequestModel(
	req types.SubmitMonthlyAttendanceRequestRequest,
	now time.Time,
) (models.MonthlyAttendanceRequest, results.Result) {
	if req.TargetUserID == 0 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_CREATE_MONTHLY_ATTENDANCE_REQUEST_MODEL_INVALID_TARGET_USER_ID",
			"月次勤怠申請データの作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

	if req.TargetYear <= 0 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_CREATE_MONTHLY_ATTENDANCE_REQUEST_MODEL_INVALID_TARGET_YEAR",
			"月次勤怠申請データの作成に失敗しました",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_CREATE_MONTHLY_ATTENDANCE_REQUEST_MODEL_INVALID_TARGET_MONTH",
			"月次勤怠申請データの作成に失敗しました",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	if now.IsZero() {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_CREATE_MONTHLY_ATTENDANCE_REQUEST_MODEL_EMPTY_NOW",
			"月次勤怠申請データの作成に失敗しました",
			nil,
		)
	}

	monthlyAttendanceRequest := models.MonthlyAttendanceRequest{
		UserID:      req.TargetUserID,
		TargetYear:  req.TargetYear,
		TargetMonth: req.TargetMonth,

		Status:      "PENDING",
		RequestMemo: req.RequestMemo,
		RequestedAt: &now,

		ApprovedBy: nil,
		ApprovedAt: nil,

		RejectedReason: nil,
		RejectedAt:     nil,

		CanceledReason: nil,
		CanceledAt:     nil,

		IsDeleted: false,
	}

	return monthlyAttendanceRequest, results.OK(
		nil,
		"BUILD_CREATE_MONTHLY_ATTENDANCE_REQUEST_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次勤怠申請 再申請用Model作成
 *
 * REJECTED / CANCELED の月次勤怠申請を PENDING に戻す。
 *
 * 注意：
 * ・承認者、承認日時はクリアする
 * ・否認理由、否認日時はクリアする
 * ・取り下げ理由、取り下げ日時はクリアする
 * ・申請メモ、申請日時は新しい内容に更新する
 */
func (builder *monthlyAttendanceRequestBuilder) BuildResubmitMonthlyAttendanceRequestModel(
	currentMonthlyAttendanceRequest models.MonthlyAttendanceRequest,
	req types.SubmitMonthlyAttendanceRequestRequest,
	now time.Time,
) (models.MonthlyAttendanceRequest, results.Result) {
	if currentMonthlyAttendanceRequest.ID == 0 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_RESUBMIT_MONTHLY_ATTENDANCE_REQUEST_MODEL_EMPTY_CURRENT_REQUEST",
			"月次勤怠再申請データの作成に失敗しました",
			nil,
		)
	}

	if req.TargetUserID == 0 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_RESUBMIT_MONTHLY_ATTENDANCE_REQUEST_MODEL_INVALID_TARGET_USER_ID",
			"月次勤怠再申請データの作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

	if currentMonthlyAttendanceRequest.UserID != req.TargetUserID {
		return models.MonthlyAttendanceRequest{}, results.Conflict(
			"BUILD_RESUBMIT_MONTHLY_ATTENDANCE_REQUEST_MODEL_USER_ID_MISMATCH",
			"月次勤怠再申請対象のユーザーが一致しません",
			map[string]any{
				"currentUserId": currentMonthlyAttendanceRequest.UserID,
				"targetUserId":  req.TargetUserID,
			},
		)
	}

	if req.TargetYear <= 0 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_RESUBMIT_MONTHLY_ATTENDANCE_REQUEST_MODEL_INVALID_TARGET_YEAR",
			"月次勤怠再申請データの作成に失敗しました",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_RESUBMIT_MONTHLY_ATTENDANCE_REQUEST_MODEL_INVALID_TARGET_MONTH",
			"月次勤怠再申請データの作成に失敗しました",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	if currentMonthlyAttendanceRequest.TargetYear != req.TargetYear ||
		currentMonthlyAttendanceRequest.TargetMonth != req.TargetMonth {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_RESUBMIT_MONTHLY_ATTENDANCE_REQUEST_MODEL_TARGET_MONTH_MISMATCH",
			"月次勤怠再申請データの作成に失敗しました",
			map[string]any{
				"currentTargetYear":  currentMonthlyAttendanceRequest.TargetYear,
				"currentTargetMonth": currentMonthlyAttendanceRequest.TargetMonth,
				"requestTargetYear":  req.TargetYear,
				"requestTargetMonth": req.TargetMonth,
			},
		)
	}

	if now.IsZero() {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_RESUBMIT_MONTHLY_ATTENDANCE_REQUEST_MODEL_EMPTY_NOW",
			"月次勤怠再申請データの作成に失敗しました",
			nil,
		)
	}

	currentMonthlyAttendanceRequest.Status = "PENDING"
	currentMonthlyAttendanceRequest.RequestMemo = req.RequestMemo
	currentMonthlyAttendanceRequest.RequestedAt = &now

	currentMonthlyAttendanceRequest.ApprovedBy = nil
	currentMonthlyAttendanceRequest.ApprovedAt = nil

	currentMonthlyAttendanceRequest.RejectedReason = nil
	currentMonthlyAttendanceRequest.RejectedAt = nil

	currentMonthlyAttendanceRequest.CanceledReason = nil
	currentMonthlyAttendanceRequest.CanceledAt = nil

	return currentMonthlyAttendanceRequest, results.OK(
		nil,
		"BUILD_RESUBMIT_MONTHLY_ATTENDANCE_REQUEST_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次勤怠申請 取り下げ用Model作成
 *
 * PENDING の月次勤怠申請を CANCELED にする。
 *
 * 注意：
 * ・取り下げ可否の判定はServiceで行う
 * ・Builderでは現在ステータスの判定はしない
 */
func (builder *monthlyAttendanceRequestBuilder) BuildCancelMonthlyAttendanceRequestModel(
	currentMonthlyAttendanceRequest models.MonthlyAttendanceRequest,
	req types.CancelMonthlyAttendanceRequestRequest,
	now time.Time,
) (models.MonthlyAttendanceRequest, results.Result) {
	if currentMonthlyAttendanceRequest.ID == 0 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_CANCEL_MONTHLY_ATTENDANCE_REQUEST_MODEL_EMPTY_CURRENT_REQUEST",
			"月次勤怠申請取り下げデータの作成に失敗しました",
			nil,
		)
	}

	if req.TargetUserID == 0 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_CANCEL_MONTHLY_ATTENDANCE_REQUEST_MODEL_INVALID_TARGET_USER_ID",
			"月次勤怠申請取り下げデータの作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

	if currentMonthlyAttendanceRequest.UserID != req.TargetUserID {
		return models.MonthlyAttendanceRequest{}, results.Conflict(
			"BUILD_CANCEL_MONTHLY_ATTENDANCE_REQUEST_MODEL_USER_ID_MISMATCH",
			"月次勤怠申請取り下げ対象のユーザーが一致しません",
			map[string]any{
				"currentUserId": currentMonthlyAttendanceRequest.UserID,
				"targetUserId":  req.TargetUserID,
			},
		)
	}

	if req.TargetYear <= 0 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_CANCEL_MONTHLY_ATTENDANCE_REQUEST_MODEL_INVALID_TARGET_YEAR",
			"月次勤怠申請取り下げデータの作成に失敗しました",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_CANCEL_MONTHLY_ATTENDANCE_REQUEST_MODEL_INVALID_TARGET_MONTH",
			"月次勤怠申請取り下げデータの作成に失敗しました",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	if currentMonthlyAttendanceRequest.TargetYear != req.TargetYear ||
		currentMonthlyAttendanceRequest.TargetMonth != req.TargetMonth {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_CANCEL_MONTHLY_ATTENDANCE_REQUEST_MODEL_TARGET_MONTH_MISMATCH",
			"月次勤怠申請取り下げデータの作成に失敗しました",
			map[string]any{
				"currentTargetYear":  currentMonthlyAttendanceRequest.TargetYear,
				"currentTargetMonth": currentMonthlyAttendanceRequest.TargetMonth,
				"requestTargetYear":  req.TargetYear,
				"requestTargetMonth": req.TargetMonth,
			},
		)
	}

	if now.IsZero() {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_CANCEL_MONTHLY_ATTENDANCE_REQUEST_MODEL_EMPTY_NOW",
			"月次勤怠申請取り下げデータの作成に失敗しました",
			nil,
		)
	}

	currentMonthlyAttendanceRequest.Status = "CANCELED"
	currentMonthlyAttendanceRequest.CanceledReason = req.CanceledReason
	currentMonthlyAttendanceRequest.CanceledAt = &now

	return currentMonthlyAttendanceRequest, results.OK(
		nil,
		"BUILD_CANCEL_MONTHLY_ATTENDANCE_REQUEST_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次勤怠申請 承認用Model作成
 *
 * PENDING の月次勤怠申請を APPROVED にする。
 *
 * 注意：
 * ・承認可否の判定はServiceで行う
 * ・Builderでは現在ステータスの判定はしない
 */
func (builder *monthlyAttendanceRequestBuilder) BuildApproveMonthlyAttendanceRequestModel(
	currentMonthlyAttendanceRequest models.MonthlyAttendanceRequest,
	loginAdminID uint,
	now time.Time,
) (models.MonthlyAttendanceRequest, results.Result) {
	if currentMonthlyAttendanceRequest.ID == 0 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_APPROVE_MONTHLY_ATTENDANCE_REQUEST_MODEL_EMPTY_CURRENT_REQUEST",
			"月次勤怠申請承認データの作成に失敗しました",
			nil,
		)
	}

	if loginAdminID == 0 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_APPROVE_MONTHLY_ATTENDANCE_REQUEST_MODEL_INVALID_ADMIN_ID",
			"月次勤怠申請承認データの作成に失敗しました",
			map[string]any{
				"adminId": loginAdminID,
			},
		)
	}

	if now.IsZero() {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_APPROVE_MONTHLY_ATTENDANCE_REQUEST_MODEL_EMPTY_NOW",
			"月次勤怠申請承認データの作成に失敗しました",
			nil,
		)
	}

	currentMonthlyAttendanceRequest.Status = "APPROVED"
	currentMonthlyAttendanceRequest.ApprovedBy = &loginAdminID
	currentMonthlyAttendanceRequest.ApprovedAt = &now

	currentMonthlyAttendanceRequest.RejectedReason = nil
	currentMonthlyAttendanceRequest.RejectedAt = nil

	currentMonthlyAttendanceRequest.CanceledReason = nil
	currentMonthlyAttendanceRequest.CanceledAt = nil

	return currentMonthlyAttendanceRequest, results.OK(
		nil,
		"BUILD_APPROVE_MONTHLY_ATTENDANCE_REQUEST_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次勤怠申請 否認用Model作成
 *
 * PENDING の月次勤怠申請を REJECTED にする。
 *
 * 注意：
 * ・否認可否の判定はServiceで行う
 * ・Builderでは現在ステータスの判定はしない
 * ・MonthlyAttendanceRequest model には rejectedBy がないため、
 *   否認した管理者IDは保存しない
 */
func (builder *monthlyAttendanceRequestBuilder) BuildRejectMonthlyAttendanceRequestModel(
	currentMonthlyAttendanceRequest models.MonthlyAttendanceRequest,
	req types.RejectMonthlyAttendanceRequestRequest,
	now time.Time,
) (models.MonthlyAttendanceRequest, results.Result) {
	if currentMonthlyAttendanceRequest.ID == 0 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_REJECT_MONTHLY_ATTENDANCE_REQUEST_MODEL_EMPTY_CURRENT_REQUEST",
			"月次勤怠申請否認データの作成に失敗しました",
			nil,
		)
	}

	if req.TargetRequestID == 0 {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_REJECT_MONTHLY_ATTENDANCE_REQUEST_MODEL_INVALID_TARGET_REQUEST_ID",
			"月次勤怠申請否認データの作成に失敗しました",
			map[string]any{
				"targetRequestId": req.TargetRequestID,
			},
		)
	}

	if req.RejectedReason == "" {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_REJECT_MONTHLY_ATTENDANCE_REQUEST_MODEL_EMPTY_REJECTED_REASON",
			"月次勤怠申請否認データの作成に失敗しました",
			nil,
		)
	}

	if now.IsZero() {
		return models.MonthlyAttendanceRequest{}, results.BadRequest(
			"BUILD_REJECT_MONTHLY_ATTENDANCE_REQUEST_MODEL_EMPTY_NOW",
			"月次勤怠申請否認データの作成に失敗しました",
			nil,
		)
	}

	rejectedReason := req.RejectedReason

	currentMonthlyAttendanceRequest.Status = "REJECTED"
	currentMonthlyAttendanceRequest.RejectedReason = &rejectedReason
	currentMonthlyAttendanceRequest.RejectedAt = &now

	currentMonthlyAttendanceRequest.ApprovedBy = nil
	currentMonthlyAttendanceRequest.ApprovedAt = nil

	currentMonthlyAttendanceRequest.CanceledReason = nil
	currentMonthlyAttendanceRequest.CanceledAt = nil

	return currentMonthlyAttendanceRequest, results.OK(
		nil,
		"BUILD_REJECT_MONTHLY_ATTENDANCE_REQUEST_MODEL_SUCCESS",
		"",
		nil,
	)
}
