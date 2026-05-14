package services

import (
	"strings"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
)

/*
 * 管理者用月次勤怠申請Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・管理者APIでは対象ユーザーIDを targetUserId としてRequestで受け取る
 * ・管理者は対象ユーザーの月次申請状態取得、申請、取り下げができる
 * ・管理者は月次勤怠申請の承認、否認ができる
 * ・管理者は月次勤怠申請一覧を検索できる
 * ・管理者側では月次申請状態による勤怠編集ロックを行わない
 */
type MonthlyAttendanceRequestService interface {
	SearchMonthlyAttendanceRequests(req types.SearchMonthlyAttendanceRequestsRequest) results.Result
	GetMonthlyAttendanceRequestStatus(req types.GetMonthlyAttendanceRequestStatusRequest) results.Result
	SubmitMonthlyAttendanceRequest(req types.SubmitMonthlyAttendanceRequestRequest) results.Result
	CancelMonthlyAttendanceRequest(req types.CancelMonthlyAttendanceRequestRequest) results.Result
	ApproveMonthlyAttendanceRequest(loginAdminID uint, req types.ApproveMonthlyAttendanceRequestRequest) results.Result
	RejectMonthlyAttendanceRequest(loginAdminID uint, req types.RejectMonthlyAttendanceRequestRequest) results.Result
}

/*
 * 管理者用月次勤怠申請Service
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリや保存用Modelを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB処理を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 *
 * このServiceで扱うもの：
 * ・月次勤怠申請一覧検索
 * ・対象ユーザーの対象月の月次申請状態取得
 * ・対象ユーザーの対象月の月次申請
 * ・対象ユーザーの申請中の月次申請取り下げ
 * ・月次勤怠申請の承認
 * ・月次勤怠申請の否認
 *
 * このServiceで扱わないもの：
 * ・勤怠日別データの更新
 * ・休憩データの更新
 * ・月次通勤定期の更新
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 * ・月次申請状態は勤怠編集ロックには使わない
 */
type monthlyAttendanceRequestService struct {
	monthlyAttendanceRequestBuilder    builders.MonthlyAttendanceRequestBuilder
	monthlyAttendanceRequestRepository repositories.MonthlyAttendanceRequestRepository
}

/*
 * MonthlyAttendanceRequestService生成
 */
func NewMonthlyAttendanceRequestService(
	monthlyAttendanceRequestBuilder builders.MonthlyAttendanceRequestBuilder,
	monthlyAttendanceRequestRepository repositories.MonthlyAttendanceRequestRepository,
) *monthlyAttendanceRequestService {
	return &monthlyAttendanceRequestService{
		monthlyAttendanceRequestBuilder:    monthlyAttendanceRequestBuilder,
		monthlyAttendanceRequestRepository: monthlyAttendanceRequestRepository,
	}
}

/*
 * 月次勤怠申請状態から画面制御用フラグを作る
 *
 * 注意：
 * ・Editable はuser側と型を合わせるために返す
 * ・管理者側では Editable=false でも勤怠編集自体は許可する
 *
 * NOT_SUBMITTED / REJECTED / CANCELED
 * 	→ 申請可能、取り下げ不可、承認不可、否認不可
 *
 * PENDING
 * 	→ 申請不可、取り下げ可能、承認可能、否認可能
 *
 * APPROVED
 * 	→ 申請不可、取り下げ不可、承認不可、否認不可
 */
func buildMonthlyAttendanceRequestFlags(
	status string,
) (editable bool, canSubmit bool, canCancel bool, canApprove bool, canReject bool) {
	switch status {
	case "PENDING":
		return false, false, true, true, true
	case "APPROVED":
		return false, false, false, false, false
	case "REJECTED":
		return true, true, false, false, false
	case "CANCELED":
		return true, true, false, false, false
	case "NOT_SUBMITTED":
		return true, true, false, false, false
	default:
		return false, false, false, false, false
	}
}

/*
 * 未申請状態のResponseを作成する
 *
 * 未申請はDBレコードなしで表現する。
 * そのため、フロント返却時だけ NOT_SUBMITTED として返す。
 */
func toNotSubmittedMonthlyAttendanceRequestResponse(
	targetUserID uint,
	targetYear int,
	targetMonth int,
) types.MonthlyAttendanceRequestResponse {
	editable, canSubmit, canCancel, canApprove, canReject := buildMonthlyAttendanceRequestFlags("NOT_SUBMITTED")

	return types.MonthlyAttendanceRequestResponse{
		ID: nil,

		TargetUserID: targetUserID,
		TargetYear:   targetYear,
		TargetMonth:  targetMonth,
		Status:       "NOT_SUBMITTED",
		Exists:       false,

		Editable:   editable,
		CanSubmit:  canSubmit,
		CanCancel:  canCancel,
		CanApprove: canApprove,
		CanReject:  canReject,

		RequestMemo: nil,
		RequestedAt: nil,

		ApprovedBy: nil,
		ApprovedAt: nil,

		RejectedReason: nil,
		RejectedAt:     nil,

		CanceledReason: nil,
		CanceledAt:     nil,

		CreatedAt: nil,
		UpdatedAt: nil,
	}
}

/*
 * models.MonthlyAttendanceRequestをフロント返却用MonthlyAttendanceRequestResponseへ変換する
 *
 * 日付はtime.Time / *time.Timeのまま返す。
 * 表示形式の整形はフロント側で行う。
 */
func toMonthlyAttendanceRequestResponse(
	monthlyAttendanceRequest models.MonthlyAttendanceRequest,
) types.MonthlyAttendanceRequestResponse {
	id := monthlyAttendanceRequest.ID
	createdAt := monthlyAttendanceRequest.CreatedAt
	updatedAt := monthlyAttendanceRequest.UpdatedAt

	editable, canSubmit, canCancel, canApprove, canReject := buildMonthlyAttendanceRequestFlags(monthlyAttendanceRequest.Status)

	return types.MonthlyAttendanceRequestResponse{
		ID: &id,

		TargetUserID: monthlyAttendanceRequest.UserID,
		TargetYear:   monthlyAttendanceRequest.TargetYear,
		TargetMonth:  monthlyAttendanceRequest.TargetMonth,
		Status:       monthlyAttendanceRequest.Status,
		Exists:       true,

		Editable:   editable,
		CanSubmit:  canSubmit,
		CanCancel:  canCancel,
		CanApprove: canApprove,
		CanReject:  canReject,

		RequestMemo: monthlyAttendanceRequest.RequestMemo,
		RequestedAt: monthlyAttendanceRequest.RequestedAt,

		ApprovedBy: monthlyAttendanceRequest.ApprovedBy,
		ApprovedAt: monthlyAttendanceRequest.ApprovedAt,

		RejectedReason: monthlyAttendanceRequest.RejectedReason,
		RejectedAt:     monthlyAttendanceRequest.RejectedAt,

		CanceledReason: monthlyAttendanceRequest.CanceledReason,
		CanceledAt:     monthlyAttendanceRequest.CanceledAt,

		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}
}

/*
 * 検索Rowをフロント返却用MonthlyAttendanceRequestResponseへ変換する
 *
 * 注意：
 * ・monthly_attendance_requests.id が NULL の場合は未申請として返す
 * ・未申請はDBには保存しない
 */
func toMonthlyAttendanceRequestResponseFromSearchRow(
	row repositories.MonthlyAttendanceRequestSearchRow,
	targetYear int,
	targetMonth int,
) types.MonthlyAttendanceRequestResponse {
	if row.MonthlyAttendanceRequestID == nil {
		return toNotSubmittedMonthlyAttendanceRequestResponse(
			row.TargetUserID,
			targetYear,
			targetMonth,
		)
	}

	status := ""
	if row.Status != nil {
		status = *row.Status
	}

	editable, canSubmit, canCancel, canApprove, canReject := buildMonthlyAttendanceRequestFlags(status)

	return types.MonthlyAttendanceRequestResponse{
		ID: row.MonthlyAttendanceRequestID,

		TargetUserID: row.TargetUserID,
		TargetYear:   targetYear,
		TargetMonth:  targetMonth,
		Status:       status,
		Exists:       true,

		Editable:   editable,
		CanSubmit:  canSubmit,
		CanCancel:  canCancel,
		CanApprove: canApprove,
		CanReject:  canReject,

		RequestMemo: row.RequestMemo,
		RequestedAt: row.RequestedAt,

		ApprovedBy: row.ApprovedBy,
		ApprovedAt: row.ApprovedAt,

		RejectedReason: row.RejectedReason,
		RejectedAt:     row.RejectedAt,

		CanceledReason: row.CanceledReason,
		CanceledAt:     row.CanceledAt,

		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

/*
 * 検索Rowを月次勤怠申請一覧Rowへ変換する
 */
func toMonthlyAttendanceRequestListRow(
	row repositories.MonthlyAttendanceRequestSearchRow,
	targetYear int,
	targetMonth int,
) types.MonthlyAttendanceRequestListRow {
	return types.MonthlyAttendanceRequestListRow{
		TargetUserID: row.TargetUserID,

		UserName: row.UserName,
		Email:    row.Email,

		DepartmentID:   row.DepartmentID,
		DepartmentName: row.DepartmentName,

		TargetYear:  targetYear,
		TargetMonth: targetMonth,

		MonthlyAttendanceRequest: toMonthlyAttendanceRequestResponseFromSearchRow(
			row,
			targetYear,
			targetMonth,
		),
	}
}

/*
 * 対象ユーザーIDのバリデーション
 */
func validateMonthlyAttendanceRequestTargetUserID(
	targetUserID uint,
	actionCode string,
) results.Result {
	if targetUserID == 0 {
		return results.BadRequest(
			actionCode+"_INVALID_TARGET_USER_ID",
			"対象ユーザーIDが正しくありません",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	return results.OK(
		nil,
		actionCode+"_VALID_TARGET_USER_ID",
		"",
		nil,
	)
}

/*
 * 対象年月のバリデーション
 */
func validateMonthlyAttendanceRequestTargetMonth(
	targetYear int,
	targetMonth int,
	actionCode string,
) results.Result {
	if targetYear <= 0 {
		return results.BadRequest(
			actionCode+"_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{
				"targetYear": targetYear,
			},
		)
	}

	if targetMonth < 1 || targetMonth > 12 {
		return results.BadRequest(
			actionCode+"_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{
				"targetMonth": targetMonth,
			},
		)
	}

	return results.OK(
		nil,
		actionCode+"_VALID_TARGET_MONTH",
		"",
		nil,
	)
}

/*
 * 管理者IDのバリデーション
 */
func validateMonthlyAttendanceRequestAdminID(
	loginAdminID uint,
	actionCode string,
) results.Result {
	if loginAdminID == 0 {
		return results.Unauthorized(
			actionCode+"_INVALID_ADMIN_ID",
			"認証情報の管理者IDが正しくありません",
			nil,
		)
	}

	return results.OK(
		nil,
		actionCode+"_VALID_ADMIN_ID",
		"",
		nil,
	)
}

/*
 * 月次勤怠申請一覧検索ステータスのバリデーション
 */
func validateSearchMonthlyAttendanceRequestStatuses(
	statuses []string,
) results.Result {
	if len(statuses) == 0 {
		return results.BadRequest(
			"SEARCH_MONTHLY_ATTENDANCE_REQUESTS_EMPTY_STATUSES",
			"申請状態を1つ以上選択してください",
			nil,
		)
	}

	allowedStatuses := map[string]bool{
		"NOT_SUBMITTED": true,
		"PENDING":       true,
		"APPROVED":      true,
		"REJECTED":      true,
		"CANCELED":      true,
	}

	for _, status := range statuses {
		trimmedStatus := strings.TrimSpace(status)

		if !allowedStatuses[trimmedStatus] {
			return results.BadRequest(
				"SEARCH_MONTHLY_ATTENDANCE_REQUESTS_INVALID_STATUS",
				"申請状態の指定が正しくありません",
				map[string]any{
					"status": status,
				},
			)
		}
	}

	return results.OK(
		nil,
		"SEARCH_MONTHLY_ATTENDANCE_REQUESTS_VALID_STATUSES",
		"",
		nil,
	)
}

/*
 * 月次勤怠申請一覧検索用limitを補正する
 */
func normalizeMonthlyAttendanceRequestSearchLimit(limit int) int {
	if limit <= 0 {
		return 20
	}

	if limit > 100 {
		return 100
	}

	return limit
}

/*
 * 月次勤怠申請一覧検索用offsetを補正する
 */
func normalizeMonthlyAttendanceRequestSearchOffset(offset int) int {
	if offset < 0 {
		return 0
	}

	return offset
}

/*
 * 月次勤怠申請一覧検索
 *
 * 仕様：
 * ・users 起点で検索する
 * ・対象年月の monthly_attendance_requests を LEFT JOIN する
 * ・申請レコードが存在しないユーザーは NOT_SUBMITTED として返す
 * ・ユーザー名、メール、所属名でキーワード検索できる
 * ・statuses 複数選択で対象状態を絞り込む
 * ・offset / limit / hasMore でページングする
 */
func (service *monthlyAttendanceRequestService) SearchMonthlyAttendanceRequests(
	req types.SearchMonthlyAttendanceRequestsRequest,
) results.Result {
	validateMonthResult := validateMonthlyAttendanceRequestTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"SEARCH_MONTHLY_ATTENDANCE_REQUESTS",
	)
	if validateMonthResult.Error {
		return validateMonthResult
	}

	validateStatusesResult := validateSearchMonthlyAttendanceRequestStatuses(req.Statuses)
	if validateStatusesResult.Error {
		return validateStatusesResult
	}

	req.Offset = normalizeMonthlyAttendanceRequestSearchOffset(req.Offset)
	req.Limit = normalizeMonthlyAttendanceRequestSearchLimit(req.Limit)

	query, buildResult := service.monthlyAttendanceRequestBuilder.BuildSearchMonthlyAttendanceRequestsQuery(
		req,
		req.Limit,
	)
	if buildResult.Error {
		return buildResult
	}

	rows, hasMore, searchResult := service.monthlyAttendanceRequestRepository.SearchMonthlyAttendanceRequests(
		query,
		req.Limit,
	)
	if searchResult.Error {
		return searchResult
	}

	monthlyAttendanceRequestRows := make([]types.MonthlyAttendanceRequestListRow, 0, len(rows))

	for _, row := range rows {
		monthlyAttendanceRequestRows = append(
			monthlyAttendanceRequestRows,
			toMonthlyAttendanceRequestListRow(
				row,
				req.TargetYear,
				req.TargetMonth,
			),
		)
	}

	return results.OK(
		types.SearchMonthlyAttendanceRequestsResponse{
			MonthlyAttendanceRequests: monthlyAttendanceRequestRows,
			HasMore:                   hasMore,
		},
		"SEARCH_MONTHLY_ATTENDANCE_REQUESTS_SUCCESS",
		"月次勤怠申請一覧を取得しました",
		nil,
	)
}

/*
 * 月次勤怠申請状態取得
 *
 * 対象ユーザーの対象年月の月次勤怠申請状態を取得する。
 *
 * 仕様：
 * ・MonthlyAttendanceRequest が存在する場合は、その状態を返す
 * ・存在しない場合は、未申請として NOT_SUBMITTED を返す
 */
func (service *monthlyAttendanceRequestService) GetMonthlyAttendanceRequestStatus(
	req types.GetMonthlyAttendanceRequestStatusRequest,
) results.Result {
	validateUserResult := validateMonthlyAttendanceRequestTargetUserID(
		req.TargetUserID,
		"GET_MONTHLY_ATTENDANCE_REQUEST_STATUS",
	)
	if validateUserResult.Error {
		return validateUserResult
	}

	validateMonthResult := validateMonthlyAttendanceRequestTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"GET_MONTHLY_ATTENDANCE_REQUEST_STATUS",
	)
	if validateMonthResult.Error {
		return validateMonthResult
	}

	// Builderで月次勤怠申請検索用クエリを作成する
	query, buildResult := service.monthlyAttendanceRequestBuilder.BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
		req.TargetUserID,
		req.TargetYear,
		req.TargetMonth,
	)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryで月次勤怠申請を取得する
	monthlyAttendanceRequest, findResult := service.monthlyAttendanceRequestRepository.FindMonthlyAttendanceRequest(query)

	// 対象月の月次勤怠申請が存在しない場合は未申請として返す
	if findResult.Error && findResult.Code == "MONTHLY_ATTENDANCE_REQUEST_NOT_FOUND" {
		return results.OK(
			types.GetMonthlyAttendanceRequestStatusResponse{
				MonthlyAttendanceRequest: toNotSubmittedMonthlyAttendanceRequestResponse(
					req.TargetUserID,
					req.TargetYear,
					req.TargetMonth,
				),
			},
			"GET_MONTHLY_ATTENDANCE_REQUEST_STATUS_SUCCESS",
			"月次勤怠申請状態を取得しました",
			nil,
		)
	}

	if findResult.Error {
		return findResult
	}

	return results.OK(
		types.GetMonthlyAttendanceRequestStatusResponse{
			MonthlyAttendanceRequest: toMonthlyAttendanceRequestResponse(monthlyAttendanceRequest),
		},
		"GET_MONTHLY_ATTENDANCE_REQUEST_STATUS_SUCCESS",
		"月次勤怠申請状態を取得しました",
		nil,
	)
}

/*
 * 月次勤怠申請
 *
 * 対象ユーザーの対象年月の月次勤怠を申請する。
 *
 * 仕様：
 * ・未申請の場合は新規作成する
 * ・REJECTED / CANCELED の場合は再申請として PENDING に戻す
 * ・PENDING の場合は二重申請として拒否する
 * ・APPROVED の場合は承認済みのため拒否する
 *
 * 注意：
 * ・管理者による代理申請でも、月次申請状態の遷移ルールはuser側と同じ
 */
func (service *monthlyAttendanceRequestService) SubmitMonthlyAttendanceRequest(
	req types.SubmitMonthlyAttendanceRequestRequest,
) results.Result {
	validateUserResult := validateMonthlyAttendanceRequestTargetUserID(
		req.TargetUserID,
		"SUBMIT_MONTHLY_ATTENDANCE_REQUEST",
	)
	if validateUserResult.Error {
		return validateUserResult
	}

	validateMonthResult := validateMonthlyAttendanceRequestTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"SUBMIT_MONTHLY_ATTENDANCE_REQUEST",
	)
	if validateMonthResult.Error {
		return validateMonthResult
	}

	// Builderで月次勤怠申請検索用クエリを作成する
	query, buildFindResult := service.monthlyAttendanceRequestBuilder.BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
		req.TargetUserID,
		req.TargetYear,
		req.TargetMonth,
	)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで既存の月次勤怠申請を取得する
	currentMonthlyAttendanceRequest, findResult := service.monthlyAttendanceRequestRepository.FindMonthlyAttendanceRequest(query)

	// 未申請の場合は新規作成する
	if findResult.Error && findResult.Code == "MONTHLY_ATTENDANCE_REQUEST_NOT_FOUND" {
		monthlyAttendanceRequest, buildCreateResult := service.monthlyAttendanceRequestBuilder.BuildCreateMonthlyAttendanceRequestModel(
			req,
			time.Now(),
		)
		if buildCreateResult.Error {
			return buildCreateResult
		}

		createdMonthlyAttendanceRequest, createResult := service.monthlyAttendanceRequestRepository.CreateMonthlyAttendanceRequest(monthlyAttendanceRequest)
		if createResult.Error {
			return createResult
		}

		return results.Created(
			types.SubmitMonthlyAttendanceRequestResponse{
				MonthlyAttendanceRequest: toMonthlyAttendanceRequestResponse(createdMonthlyAttendanceRequest),
			},
			"SUBMIT_MONTHLY_ATTENDANCE_REQUEST_SUCCESS",
			"月次勤怠を申請しました",
			nil,
		)
	}

	if findResult.Error {
		return findResult
	}

	// 申請中の場合は二重申請を拒否する
	if currentMonthlyAttendanceRequest.Status == "PENDING" {
		return results.Conflict(
			"SUBMIT_MONTHLY_ATTENDANCE_REQUEST_ALREADY_PENDING",
			"すでに月次勤怠を申請中です",
			map[string]any{
				"monthlyAttendanceRequestId": currentMonthlyAttendanceRequest.ID,
				"status":                     currentMonthlyAttendanceRequest.Status,
			},
		)
	}

	// 承認済みの場合は再申請できない
	if currentMonthlyAttendanceRequest.Status == "APPROVED" {
		return results.Conflict(
			"SUBMIT_MONTHLY_ATTENDANCE_REQUEST_ALREADY_APPROVED",
			"月次勤怠は承認済みのため、再申請できません",
			map[string]any{
				"monthlyAttendanceRequestId": currentMonthlyAttendanceRequest.ID,
				"status":                     currentMonthlyAttendanceRequest.Status,
			},
		)
	}

	// 否認済み、取り下げ済みの場合は再申請する
	if currentMonthlyAttendanceRequest.Status == "REJECTED" ||
		currentMonthlyAttendanceRequest.Status == "CANCELED" {
		monthlyAttendanceRequest, buildUpdateResult := service.monthlyAttendanceRequestBuilder.BuildResubmitMonthlyAttendanceRequestModel(
			currentMonthlyAttendanceRequest,
			req,
			time.Now(),
		)
		if buildUpdateResult.Error {
			return buildUpdateResult
		}

		savedMonthlyAttendanceRequest, saveResult := service.monthlyAttendanceRequestRepository.SaveMonthlyAttendanceRequest(monthlyAttendanceRequest)
		if saveResult.Error {
			return saveResult
		}

		return results.OK(
			types.SubmitMonthlyAttendanceRequestResponse{
				MonthlyAttendanceRequest: toMonthlyAttendanceRequestResponse(savedMonthlyAttendanceRequest),
			},
			"RESUBMIT_MONTHLY_ATTENDANCE_REQUEST_SUCCESS",
			"月次勤怠を再申請しました",
			nil,
		)
	}

	return results.Conflict(
		"SUBMIT_MONTHLY_ATTENDANCE_REQUEST_INVALID_STATUS",
		"現在の月次勤怠申請状態では申請できません",
		map[string]any{
			"monthlyAttendanceRequestId": currentMonthlyAttendanceRequest.ID,
			"status":                     currentMonthlyAttendanceRequest.Status,
		},
	)
}

/*
 * 月次勤怠申請取り下げ
 *
 * 対象ユーザーの対象年月の月次勤怠申請を取り下げる。
 *
 * 仕様：
 * ・取り下げできるのは PENDING のみ
 * ・未申請、承認済み、否認済み、取り下げ済みは拒否する
 */
func (service *monthlyAttendanceRequestService) CancelMonthlyAttendanceRequest(
	req types.CancelMonthlyAttendanceRequestRequest,
) results.Result {
	validateUserResult := validateMonthlyAttendanceRequestTargetUserID(
		req.TargetUserID,
		"CANCEL_MONTHLY_ATTENDANCE_REQUEST",
	)
	if validateUserResult.Error {
		return validateUserResult
	}

	validateMonthResult := validateMonthlyAttendanceRequestTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"CANCEL_MONTHLY_ATTENDANCE_REQUEST",
	)
	if validateMonthResult.Error {
		return validateMonthResult
	}

	// Builderで月次勤怠申請検索用クエリを作成する
	query, buildFindResult := service.monthlyAttendanceRequestBuilder.BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
		req.TargetUserID,
		req.TargetYear,
		req.TargetMonth,
	)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで既存の月次勤怠申請を取得する
	currentMonthlyAttendanceRequest, findResult := service.monthlyAttendanceRequestRepository.FindMonthlyAttendanceRequest(query)
	if findResult.Error {
		if findResult.Code == "MONTHLY_ATTENDANCE_REQUEST_NOT_FOUND" {
			return results.Conflict(
				"CANCEL_MONTHLY_ATTENDANCE_REQUEST_NOT_SUBMITTED",
				"未申請のため、月次勤怠申請を取り下げできません",
				map[string]any{
					"targetUserId": req.TargetUserID,
					"targetYear":   req.TargetYear,
					"targetMonth":  req.TargetMonth,
				},
			)
		}

		return findResult
	}

	// 申請中以外は取り下げ不可
	if currentMonthlyAttendanceRequest.Status != "PENDING" {
		return results.Conflict(
			"CANCEL_MONTHLY_ATTENDANCE_REQUEST_INVALID_STATUS",
			"申請中ではないため、月次勤怠申請を取り下げできません",
			map[string]any{
				"monthlyAttendanceRequestId": currentMonthlyAttendanceRequest.ID,
				"status":                     currentMonthlyAttendanceRequest.Status,
			},
		)
	}

	// Builderで取り下げ用Modelを作る
	monthlyAttendanceRequest, buildCancelResult := service.monthlyAttendanceRequestBuilder.BuildCancelMonthlyAttendanceRequestModel(
		currentMonthlyAttendanceRequest,
		req,
		time.Now(),
	)
	if buildCancelResult.Error {
		return buildCancelResult
	}

	// Repositoryで月次勤怠申請を保存する
	savedMonthlyAttendanceRequest, saveResult := service.monthlyAttendanceRequestRepository.SaveMonthlyAttendanceRequest(monthlyAttendanceRequest)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.CancelMonthlyAttendanceRequestResponse{
			MonthlyAttendanceRequest: toMonthlyAttendanceRequestResponse(savedMonthlyAttendanceRequest),
		},
		"CANCEL_MONTHLY_ATTENDANCE_REQUEST_SUCCESS",
		"月次勤怠申請を取り下げました",
		nil,
	)
}

/*
 * 月次勤怠申請承認
 *
 * 指定された月次勤怠申請を承認する。
 *
 * 仕様：
 * ・承認できるのは PENDING のみ
 * ・承認者IDはJWTから取得した管理者IDを使う
 */
func (service *monthlyAttendanceRequestService) ApproveMonthlyAttendanceRequest(
	loginAdminID uint,
	req types.ApproveMonthlyAttendanceRequestRequest,
) results.Result {
	validateAdminResult := validateMonthlyAttendanceRequestAdminID(
		loginAdminID,
		"APPROVE_MONTHLY_ATTENDANCE_REQUEST",
	)
	if validateAdminResult.Error {
		return validateAdminResult
	}

	if req.TargetRequestID == 0 {
		return results.BadRequest(
			"APPROVE_MONTHLY_ATTENDANCE_REQUEST_INVALID_TARGET_REQUEST_ID",
			"月次勤怠申請IDが正しくありません",
			map[string]any{
				"targetRequestId": req.TargetRequestID,
			},
		)
	}

	query, buildFindResult := service.monthlyAttendanceRequestBuilder.BuildFindMonthlyAttendanceRequestByIDQuery(req.TargetRequestID)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentMonthlyAttendanceRequest, findResult := service.monthlyAttendanceRequestRepository.FindMonthlyAttendanceRequest(query)
	if findResult.Error {
		return findResult
	}

	if currentMonthlyAttendanceRequest.Status != "PENDING" {
		return results.Conflict(
			"APPROVE_MONTHLY_ATTENDANCE_REQUEST_INVALID_STATUS",
			"申請中ではないため、月次勤怠申請を承認できません",
			map[string]any{
				"monthlyAttendanceRequestId": currentMonthlyAttendanceRequest.ID,
				"status":                     currentMonthlyAttendanceRequest.Status,
			},
		)
	}

	monthlyAttendanceRequest, buildApproveResult := service.monthlyAttendanceRequestBuilder.BuildApproveMonthlyAttendanceRequestModel(
		currentMonthlyAttendanceRequest,
		loginAdminID,
		time.Now(),
	)
	if buildApproveResult.Error {
		return buildApproveResult
	}

	savedMonthlyAttendanceRequest, saveResult := service.monthlyAttendanceRequestRepository.SaveMonthlyAttendanceRequest(monthlyAttendanceRequest)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.ApproveMonthlyAttendanceRequestResponse{
			MonthlyAttendanceRequest: toMonthlyAttendanceRequestResponse(savedMonthlyAttendanceRequest),
		},
		"APPROVE_MONTHLY_ATTENDANCE_REQUEST_SUCCESS",
		"月次勤怠申請を承認しました",
		nil,
	)
}

/*
 * 月次勤怠申請否認
 *
 * 指定された月次勤怠申請を否認する。
 *
 * 仕様：
 * ・否認できるのは PENDING のみ
 * ・否認理由は必須
 *
 * 注意：
 * ・MonthlyAttendanceRequest model には rejectedBy がないため、
 *   否認した管理者IDは保存しない
 */
func (service *monthlyAttendanceRequestService) RejectMonthlyAttendanceRequest(
	loginAdminID uint,
	req types.RejectMonthlyAttendanceRequestRequest,
) results.Result {
	validateAdminResult := validateMonthlyAttendanceRequestAdminID(
		loginAdminID,
		"REJECT_MONTHLY_ATTENDANCE_REQUEST",
	)
	if validateAdminResult.Error {
		return validateAdminResult
	}

	if req.TargetRequestID == 0 {
		return results.BadRequest(
			"REJECT_MONTHLY_ATTENDANCE_REQUEST_INVALID_TARGET_REQUEST_ID",
			"月次勤怠申請IDが正しくありません",
			map[string]any{
				"targetRequestId": req.TargetRequestID,
			},
		)
	}

	if req.RejectedReason == "" {
		return results.BadRequest(
			"REJECT_MONTHLY_ATTENDANCE_REQUEST_EMPTY_REJECTED_REASON",
			"否認理由を入力してください",
			nil,
		)
	}

	query, buildFindResult := service.monthlyAttendanceRequestBuilder.BuildFindMonthlyAttendanceRequestByIDQuery(req.TargetRequestID)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentMonthlyAttendanceRequest, findResult := service.monthlyAttendanceRequestRepository.FindMonthlyAttendanceRequest(query)
	if findResult.Error {
		return findResult
	}

	if currentMonthlyAttendanceRequest.Status != "PENDING" {
		return results.Conflict(
			"REJECT_MONTHLY_ATTENDANCE_REQUEST_INVALID_STATUS",
			"申請中ではないため、月次勤怠申請を否認できません",
			map[string]any{
				"monthlyAttendanceRequestId": currentMonthlyAttendanceRequest.ID,
				"status":                     currentMonthlyAttendanceRequest.Status,
			},
		)
	}

	monthlyAttendanceRequest, buildRejectResult := service.monthlyAttendanceRequestBuilder.BuildRejectMonthlyAttendanceRequestModel(
		currentMonthlyAttendanceRequest,
		req,
		time.Now(),
	)
	if buildRejectResult.Error {
		return buildRejectResult
	}

	savedMonthlyAttendanceRequest, saveResult := service.monthlyAttendanceRequestRepository.SaveMonthlyAttendanceRequest(monthlyAttendanceRequest)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.RejectMonthlyAttendanceRequestResponse{
			MonthlyAttendanceRequest: toMonthlyAttendanceRequestResponse(savedMonthlyAttendanceRequest),
		},
		"REJECT_MONTHLY_ATTENDANCE_REQUEST_SUCCESS",
		"月次勤怠申請を否認しました",
		nil,
	)
}
