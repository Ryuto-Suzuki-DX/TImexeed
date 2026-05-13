package services

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
)

/*
 * 従業員用月次勤怠申請Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・従業員APIでは userId / targetUserId をRequestで受け取らない
 * ・ControllerでAuthMiddleware由来のuserIdを取得し、Serviceへ渡す
 * ・管理者による承認、否認はadmin側のServiceで行う
 */
type MonthlyAttendanceRequestService interface {
	GetMonthlyAttendanceRequestStatus(userID uint, req types.GetMonthlyAttendanceRequestStatusRequest) results.Result
	SubmitMonthlyAttendanceRequest(userID uint, req types.SubmitMonthlyAttendanceRequestRequest) results.Result
	CancelMonthlyAttendanceRequest(userID uint, req types.CancelMonthlyAttendanceRequestRequest) results.Result
}

/*
 * 従業員用月次勤怠申請Service
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
 * ・対象月の月次申請状態取得
 * ・対象月の月次申請
 * ・申請中の月次申請取り下げ
 *
 * このServiceで扱わないもの：
 * ・管理者による承認
 * ・管理者による否認
 * ・勤怠日別データの更新
 * ・休憩データの更新
 * ・月次通勤定期の更新
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
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
 * NOT_SUBMITTED / REJECTED / CANCELED
 * 	→ 編集可能、申請可能、取り下げ不可
 *
 * PENDING
 * 	→ 編集不可、申請不可、取り下げ可能
 *
 * APPROVED
 * 	→ 編集不可、申請不可、取り下げ不可
 */
func buildMonthlyAttendanceRequestFlags(status string) (editable bool, canSubmit bool, canCancel bool) {
	switch status {
	case "PENDING":
		return false, false, true
	case "APPROVED":
		return false, false, false
	case "REJECTED":
		return true, true, false
	case "CANCELED":
		return true, true, false
	case "NOT_SUBMITTED":
		return true, true, false
	default:
		return false, false, false
	}
}

/*
 * 未申請状態のResponseを作成する
 *
 * 未申請はDBレコードなしで表現する。
 * そのため、フロント返却時だけ NOT_SUBMITTED として返す。
 */
func toNotSubmittedMonthlyAttendanceRequestResponse(
	targetYear int,
	targetMonth int,
) types.MonthlyAttendanceRequestResponse {
	editable, canSubmit, canCancel := buildMonthlyAttendanceRequestFlags("NOT_SUBMITTED")

	return types.MonthlyAttendanceRequestResponse{
		ID:          nil,
		TargetYear:  targetYear,
		TargetMonth: targetMonth,
		Status:      "NOT_SUBMITTED",
		Exists:      false,

		Editable:  editable,
		CanSubmit: canSubmit,
		CanCancel: canCancel,

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

	editable, canSubmit, canCancel := buildMonthlyAttendanceRequestFlags(monthlyAttendanceRequest.Status)

	return types.MonthlyAttendanceRequestResponse{
		ID:          &id,
		TargetYear:  monthlyAttendanceRequest.TargetYear,
		TargetMonth: monthlyAttendanceRequest.TargetMonth,
		Status:      monthlyAttendanceRequest.Status,
		Exists:      true,

		Editable:  editable,
		CanSubmit: canSubmit,
		CanCancel: canCancel,

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
 * 月次勤怠申請状態取得
 *
 * 対象年月のログイン中ユーザー本人の月次勤怠申請状態を取得する。
 *
 * 仕様：
 * ・MonthlyAttendanceRequest が存在する場合は、その状態を返す
 * ・存在しない場合は、未申請として NOT_SUBMITTED を返す
 */
func (service *monthlyAttendanceRequestService) GetMonthlyAttendanceRequestStatus(
	userID uint,
	req types.GetMonthlyAttendanceRequestStatusRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"GET_MONTHLY_ATTENDANCE_REQUEST_STATUS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	validateResult := validateMonthlyAttendanceRequestTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"GET_MONTHLY_ATTENDANCE_REQUEST_STATUS",
	)
	if validateResult.Error {
		return validateResult
	}

	// Builderで月次勤怠申請検索用クエリを作成する
	query, buildResult := service.monthlyAttendanceRequestBuilder.BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
		userID,
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
 * 対象年月のログイン中ユーザー本人の月次勤怠を申請する。
 *
 * 仕様：
 * ・未申請の場合は新規作成する
 * ・REJECTED / CANCELED の場合は再申請として PENDING に戻す
 * ・PENDING の場合は二重申請として拒否する
 * ・APPROVED の場合は承認済みのため拒否する
 */
func (service *monthlyAttendanceRequestService) SubmitMonthlyAttendanceRequest(
	userID uint,
	req types.SubmitMonthlyAttendanceRequestRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"SUBMIT_MONTHLY_ATTENDANCE_REQUEST_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	validateResult := validateMonthlyAttendanceRequestTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"SUBMIT_MONTHLY_ATTENDANCE_REQUEST",
	)
	if validateResult.Error {
		return validateResult
	}

	// Builderで月次勤怠申請検索用クエリを作成する
	query, buildFindResult := service.monthlyAttendanceRequestBuilder.BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
		userID,
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
			userID,
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
 * 対象年月のログイン中ユーザー本人の月次勤怠申請を取り下げる。
 *
 * 仕様：
 * ・取り下げできるのは PENDING のみ
 * ・未申請、承認済み、否認済み、取り下げ済みは拒否する
 */
func (service *monthlyAttendanceRequestService) CancelMonthlyAttendanceRequest(
	userID uint,
	req types.CancelMonthlyAttendanceRequestRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"CANCEL_MONTHLY_ATTENDANCE_REQUEST_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	validateResult := validateMonthlyAttendanceRequestTargetMonth(
		req.TargetYear,
		req.TargetMonth,
		"CANCEL_MONTHLY_ATTENDANCE_REQUEST",
	)
	if validateResult.Error {
		return validateResult
	}

	// Builderで月次勤怠申請検索用クエリを作成する
	query, buildFindResult := service.monthlyAttendanceRequestBuilder.BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
		userID,
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
					"targetYear":  req.TargetYear,
					"targetMonth": req.TargetMonth,
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
