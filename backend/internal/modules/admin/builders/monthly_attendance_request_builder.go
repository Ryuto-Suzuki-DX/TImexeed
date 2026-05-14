package builders

import (
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type MonthlyAttendanceRequestBuilder interface {
	BuildSearchMonthlyAttendanceRequestsQuery(
		req types.SearchMonthlyAttendanceRequestsRequest,
		limit int,
	) (*gorm.DB, results.Result)

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
 * ・月次勤怠申請一覧検索クエリ作成
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
 * ・Responseへの変換
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Create / Save / Search はRepositoryに任せる
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
 * 月次勤怠申請一覧検索用クエリ作成
 *
 * 仕様：
 * ・users 起点で検索する
 * ・departments は所属名検索と一覧表示のため LEFT JOIN する
 * ・monthly_attendance_requests は対象年月で LEFT JOIN する
 * ・申請レコードが存在しないユーザーは未申請として扱うため、LEFT JOIN が必須
 * ・未申請は monthly_attendance_requests.id IS NULL で判定する
 *
 * 状態検索：
 * ・NOT_SUBMITTED が含まれる場合は monthly_attendance_requests.id IS NULL を検索対象にする
 * ・PENDING / APPROVED / REJECTED / CANCELED は monthly_attendance_requests.status IN (?) で検索する
 */
func (builder *monthlyAttendanceRequestBuilder) BuildSearchMonthlyAttendanceRequestsQuery(
	req types.SearchMonthlyAttendanceRequestsRequest,
	limit int,
) (*gorm.DB, results.Result) {
	if req.TargetYear <= 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_MONTHLY_ATTENDANCE_REQUESTS_QUERY_INVALID_TARGET_YEAR",
			"月次勤怠申請一覧検索条件の作成に失敗しました",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_MONTHLY_ATTENDANCE_REQUESTS_QUERY_INVALID_TARGET_MONTH",
			"月次勤怠申請一覧検索条件の作成に失敗しました",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	if len(req.Statuses) == 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_MONTHLY_ATTENDANCE_REQUESTS_QUERY_EMPTY_STATUSES",
			"月次勤怠申請一覧検索条件の作成に失敗しました",
			nil,
		)
	}

	if req.Offset < 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_MONTHLY_ATTENDANCE_REQUESTS_QUERY_INVALID_OFFSET",
			"月次勤怠申請一覧検索条件の作成に失敗しました",
			map[string]any{
				"offset": req.Offset,
			},
		)
	}

	if limit <= 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_MONTHLY_ATTENDANCE_REQUESTS_QUERY_INVALID_LIMIT",
			"月次勤怠申請一覧検索条件の作成に失敗しました",
			map[string]any{
				"limit": limit,
			},
		)
	}

	query := builder.db.
		Table("users").
		Select(`
			users.id AS target_user_id,
			users.name AS user_name,
			users.email AS email,

			departments.id AS department_id,
			departments.name AS department_name,

			monthly_attendance_requests.id AS monthly_attendance_request_id,
			monthly_attendance_requests.target_year AS request_target_year,
			monthly_attendance_requests.target_month AS request_target_month,
			monthly_attendance_requests.status AS status,
			monthly_attendance_requests.request_memo AS request_memo,
			monthly_attendance_requests.requested_at AS requested_at,
			monthly_attendance_requests.approved_by AS approved_by,
			monthly_attendance_requests.approved_at AS approved_at,
			monthly_attendance_requests.rejected_reason AS rejected_reason,
			monthly_attendance_requests.rejected_at AS rejected_at,
			monthly_attendance_requests.canceled_reason AS canceled_reason,
			monthly_attendance_requests.canceled_at AS canceled_at,
			monthly_attendance_requests.created_at AS created_at,
			monthly_attendance_requests.updated_at AS updated_at
		`).
		Joins(`
			LEFT JOIN departments
			  ON departments.id = users.department_id
			 AND departments.is_deleted = ?
		`, false).
		Joins(`
			LEFT JOIN monthly_attendance_requests
			  ON monthly_attendance_requests.user_id = users.id
			 AND monthly_attendance_requests.target_year = ?
			 AND monthly_attendance_requests.target_month = ?
			 AND monthly_attendance_requests.is_deleted = ?
		`, req.TargetYear, req.TargetMonth, false)

	if !req.IncludeDeletedUsers {
		query = query.Where("users.is_deleted = ?", false)
	}

	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"

		query = query.Where(
			`(
				users.name LIKE ?
				OR users.email LIKE ?
				OR departments.name LIKE ?
			)`,
			likeKeyword,
			likeKeyword,
			likeKeyword,
		)
	}

	hasNotSubmitted := false
	dbStatuses := make([]string, 0)

	for _, status := range req.Statuses {
		switch status {
		case "NOT_SUBMITTED":
			hasNotSubmitted = true
		default:
			dbStatuses = append(dbStatuses, status)
		}
	}

	if hasNotSubmitted && len(dbStatuses) > 0 {
		query = query.Where(
			`(
				monthly_attendance_requests.id IS NULL
				OR monthly_attendance_requests.status IN ?
			)`,
			dbStatuses,
		)
	} else if hasNotSubmitted {
		query = query.Where("monthly_attendance_requests.id IS NULL")
	} else {
		query = query.Where("monthly_attendance_requests.status IN ?", dbStatuses)
	}

	query = query.
		Order("users.id ASC").
		Offset(req.Offset).
		Limit(limit + 1)

	return query, results.OK(
		nil,
		"BUILD_SEARCH_MONTHLY_ATTENDANCE_REQUESTS_QUERY_SUCCESS",
		"",
		nil,
	)
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
