package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 従業員用休憩Service interface
 *
 * Controllerや月次勤怠全体保存ServiceがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・従業員APIでは userId / targetUserId をRequestで受け取らない
 * ・ControllerでAuthMiddleware由来のuserIdを取得し、Serviceへ渡す
 * ・AttendanceBreak は申請状態を持たない
 * ・編集可否は MonthlyAttendanceRequest を見て判断する
 */
type AttendanceBreakService interface {
	SearchAttendanceBreaks(userID uint, req types.SearchAttendanceBreaksRequest) results.Result
	CreateAttendanceBreak(userID uint, req types.CreateAttendanceBreakRequest) results.Result
	UpdateAttendanceBreak(userID uint, req types.UpdateAttendanceBreakRequest) results.Result
	DeleteAttendanceBreak(userID uint, req types.DeleteAttendanceBreakRequest) results.Result
	UpdateAttendanceBreaksByWorkDate(userID uint, req types.UpdateAttendanceBreaksByWorkDateRequest) results.Result
}

/*
 * 従業員用休憩Service
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
 * 状態管理：
 * ・AttendanceBreak 自体は申請状態を持たない
 * ・対象月の申請状態は MonthlyAttendanceRequest から取得する
 * ・MonthlyAttendanceRequest が存在しない場合は未申請扱いにする
 *
 * 月次全体保存での休憩保存方針：
 * ・削除 → 全新規作成はしない
 * ・リクエストにIDがある休憩は更新する
 * ・リクエストにIDがない休憩は新規作成する
 * ・DBに存在するがリクエストから消えた休憩は論理削除する
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type attendanceBreakService struct {
	attendanceBreakBuilder             builders.AttendanceBreakBuilder
	attendanceBreakRepository          repositories.AttendanceBreakRepository
	attendanceDayBuilder               builders.AttendanceDayBuilder
	attendanceDayRepository            repositories.AttendanceDayRepository
	monthlyAttendanceRequestBuilder    builders.MonthlyAttendanceRequestBuilder
	monthlyAttendanceRequestRepository repositories.MonthlyAttendanceRequestRepository
}

/*
 * AttendanceBreakService生成
 */
func NewAttendanceBreakService(
	attendanceBreakBuilder builders.AttendanceBreakBuilder,
	attendanceBreakRepository repositories.AttendanceBreakRepository,
	attendanceDayBuilder builders.AttendanceDayBuilder,
	attendanceDayRepository repositories.AttendanceDayRepository,
	monthlyAttendanceRequestBuilder builders.MonthlyAttendanceRequestBuilder,
	monthlyAttendanceRequestRepository repositories.MonthlyAttendanceRequestRepository,
) *attendanceBreakService {
	return &attendanceBreakService{
		attendanceBreakBuilder:             attendanceBreakBuilder,
		attendanceBreakRepository:          attendanceBreakRepository,
		attendanceDayBuilder:               attendanceDayBuilder,
		attendanceDayRepository:            attendanceDayRepository,
		monthlyAttendanceRequestBuilder:    monthlyAttendanceRequestBuilder,
		monthlyAttendanceRequestRepository: monthlyAttendanceRequestRepository,
	}
}

/*
 * models.AttendanceBreakをフロント返却用AttendanceBreakResponseへ変換する
 *
 * 日付はtime.Timeのまま返す。
 * 表示形式の整形はフロント側で行う。
 */
func toAttendanceBreakResponse(attendanceBreak models.AttendanceBreak) types.AttendanceBreakResponse {
	return types.AttendanceBreakResponse{
		ID:              attendanceBreak.ID,
		AttendanceDayID: attendanceBreak.AttendanceDayID,
		BreakStartAt:    attendanceBreak.BreakStartAt,
		BreakEndAt:      attendanceBreak.BreakEndAt,
		BreakMemo:       attendanceBreak.BreakMemo,
		IsDeleted:       attendanceBreak.IsDeleted,
		CreatedAt:       attendanceBreak.CreatedAt,
		UpdatedAt:       attendanceBreak.UpdatedAt,
		DeletedAt:       attendanceBreak.DeletedAt,
	}
}

/*
 * 対象月の休憩が編集可能か確認する
 *
 * PENDING / APPROVED の場合は従業員側から作成・更新・削除できない。
 * NOT_SUBMITTED / REJECTED / CANCELED の場合は編集可能。
 */
func (service *attendanceBreakService) validateMonthlyAttendanceEditable(
	userID uint,
	targetYear int,
	targetMonth int,
	actionCode string,
) results.Result {
	query, buildResult := service.monthlyAttendanceRequestBuilder.BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
		userID,
		targetYear,
		targetMonth,
	)
	if buildResult.Error {
		return buildResult
	}

	monthlyAttendanceRequest, findResult := service.monthlyAttendanceRequestRepository.FindMonthlyAttendanceRequest(query)

	if findResult.Error && findResult.Code == "MONTHLY_ATTENDANCE_REQUEST_NOT_FOUND" {
		return results.OK(
			nil,
			actionCode+"_MONTHLY_ATTENDANCE_EDITABLE",
			"",
			nil,
		)
	}

	if findResult.Error {
		return findResult
	}

	if monthlyAttendanceRequest.Status == "PENDING" {
		return results.Conflict(
			actionCode+"_MONTHLY_ATTENDANCE_REQUEST_PENDING",
			"月次申請中のため、休憩を変更できません",
			map[string]any{
				"targetYear":  targetYear,
				"targetMonth": targetMonth,
				"status":      monthlyAttendanceRequest.Status,
			},
		)
	}

	if monthlyAttendanceRequest.Status == "APPROVED" {
		return results.Conflict(
			actionCode+"_MONTHLY_ATTENDANCE_REQUEST_APPROVED",
			"月次承認済みのため、休憩を変更できません",
			map[string]any{
				"targetYear":  targetYear,
				"targetMonth": targetMonth,
				"status":      monthlyAttendanceRequest.Status,
			},
		)
	}

	return results.OK(
		nil,
		actionCode+"_MONTHLY_ATTENDANCE_EDITABLE",
		"",
		nil,
	)
}

/*
 * 休憩検索
 *
 * userID + workDate で勤怠日を取得し、その勤怠日に紐づく休憩一覧を取得する。
 */
func (service *attendanceBreakService) SearchAttendanceBreaks(
	userID uint,
	req types.SearchAttendanceBreaksRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"SEARCH_ATTENDANCE_BREAKS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	workDate, err := utils.ParseDate(req.WorkDate)
	if err != nil {
		return results.BadRequest(
			"SEARCH_ATTENDANCE_BREAKS_INVALID_WORK_DATE",
			"対象日の形式が正しくありません",
			map[string]any{
				"workDate": req.WorkDate,
				"format":   "yyyy-MM-dd",
			},
		)
	}

	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(userID, workDate)
	if buildFindAttendanceDayResult.Error {
		return buildFindAttendanceDayResult
	}

	attendanceDay, findAttendanceDayResult := service.attendanceDayRepository.FindAttendanceDay(findAttendanceDayQuery)
	if findAttendanceDayResult.Error {
		return findAttendanceDayResult
	}

	query, buildResult := service.attendanceBreakBuilder.BuildSearchAttendanceBreaksQuery(attendanceDay.ID)
	if buildResult.Error {
		return buildResult
	}

	attendanceBreaks, findResult := service.attendanceBreakRepository.FindAttendanceBreaks(query)
	if findResult.Error {
		return findResult
	}

	attendanceBreakResponses := make([]types.AttendanceBreakResponse, 0, len(attendanceBreaks))
	for _, attendanceBreak := range attendanceBreaks {
		attendanceBreakResponses = append(attendanceBreakResponses, toAttendanceBreakResponse(attendanceBreak))
	}

	return results.OK(
		types.SearchAttendanceBreaksResponse{
			WorkDate:         req.WorkDate,
			AttendanceBreaks: attendanceBreakResponses,
		},
		"SEARCH_ATTENDANCE_BREAKS_SUCCESS",
		"休憩一覧を取得しました",
		nil,
	)
}

/*
 * 休憩作成
 *
 * APIとして直接公開しない。
 * monthly_attendances/update の月次全体保存から内部的に使う。
 */
func (service *attendanceBreakService) CreateAttendanceBreak(
	userID uint,
	req types.CreateAttendanceBreakRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"CREATE_ATTENDANCE_BREAK_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	workDate, err := utils.ParseDate(req.WorkDate)
	if err != nil {
		return results.BadRequest(
			"CREATE_ATTENDANCE_BREAK_INVALID_WORK_DATE",
			"対象日の形式が正しくありません",
			map[string]any{
				"workDate": req.WorkDate,
				"format":   "yyyy-MM-dd",
			},
		)
	}

	editableResult := service.validateMonthlyAttendanceEditable(
		userID,
		workDate.Year(),
		int(workDate.Month()),
		"CREATE_ATTENDANCE_BREAK",
	)
	if editableResult.Error {
		return editableResult
	}

	breakStartAt, err := utils.ParseDateTime(req.BreakStartAt)
	if err != nil {
		return results.BadRequest(
			"CREATE_ATTENDANCE_BREAK_INVALID_BREAK_START_AT",
			"休憩開始日時の形式が正しくありません",
			map[string]any{
				"breakStartAt": req.BreakStartAt,
				"format":       "RFC3339",
			},
		)
	}

	breakEndAt, err := utils.ParseDateTime(req.BreakEndAt)
	if err != nil {
		return results.BadRequest(
			"CREATE_ATTENDANCE_BREAK_INVALID_BREAK_END_AT",
			"休憩終了日時の形式が正しくありません",
			map[string]any{
				"breakEndAt": req.BreakEndAt,
				"format":     "RFC3339",
			},
		)
	}

	if !breakEndAt.After(breakStartAt) {
		return results.BadRequest(
			"CREATE_ATTENDANCE_BREAK_INVALID_TIME_RANGE",
			"休憩終了日時は休憩開始日時より後にしてください",
			map[string]any{
				"breakStartAt": req.BreakStartAt,
				"breakEndAt":   req.BreakEndAt,
			},
		)
	}

	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(userID, workDate)
	if buildFindAttendanceDayResult.Error {
		return buildFindAttendanceDayResult
	}

	attendanceDay, findAttendanceDayResult := service.attendanceDayRepository.FindAttendanceDay(findAttendanceDayQuery)
	if findAttendanceDayResult.Error {
		return findAttendanceDayResult
	}

	attendanceBreak, buildCreateResult := service.attendanceBreakBuilder.BuildCreateAttendanceBreakModel(
		attendanceDay.ID,
		req,
		breakStartAt,
		breakEndAt,
	)
	if buildCreateResult.Error {
		return buildCreateResult
	}

	createdAttendanceBreak, createResult := service.attendanceBreakRepository.CreateAttendanceBreak(attendanceBreak)
	if createResult.Error {
		return createResult
	}

	return results.Created(
		types.CreateAttendanceBreakResponse{
			AttendanceBreak: toAttendanceBreakResponse(createdAttendanceBreak),
		},
		"CREATE_ATTENDANCE_BREAK_SUCCESS",
		"休憩を作成しました",
		nil,
	)
}

/*
 * 休憩更新
 *
 * APIとして直接公開しない。
 * monthly_attendances/update の月次全体保存から内部的に使う。
 */
func (service *attendanceBreakService) UpdateAttendanceBreak(
	userID uint,
	req types.UpdateAttendanceBreakRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"UPDATE_ATTENDANCE_BREAK_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	workDate, err := utils.ParseDate(req.WorkDate)
	if err != nil {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_BREAK_INVALID_WORK_DATE",
			"対象日の形式が正しくありません",
			map[string]any{
				"workDate": req.WorkDate,
				"format":   "yyyy-MM-dd",
			},
		)
	}

	editableResult := service.validateMonthlyAttendanceEditable(
		userID,
		workDate.Year(),
		int(workDate.Month()),
		"UPDATE_ATTENDANCE_BREAK",
	)
	if editableResult.Error {
		return editableResult
	}

	breakStartAt, err := utils.ParseDateTime(req.BreakStartAt)
	if err != nil {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_BREAK_INVALID_BREAK_START_AT",
			"休憩開始日時の形式が正しくありません",
			map[string]any{
				"breakStartAt": req.BreakStartAt,
				"format":       "RFC3339",
			},
		)
	}

	breakEndAt, err := utils.ParseDateTime(req.BreakEndAt)
	if err != nil {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_BREAK_INVALID_BREAK_END_AT",
			"休憩終了日時の形式が正しくありません",
			map[string]any{
				"breakEndAt": req.BreakEndAt,
				"format":     "RFC3339",
			},
		)
	}

	if !breakEndAt.After(breakStartAt) {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_BREAK_INVALID_TIME_RANGE",
			"休憩終了日時は休憩開始日時より後にしてください",
			map[string]any{
				"breakStartAt": req.BreakStartAt,
				"breakEndAt":   req.BreakEndAt,
			},
		)
	}

	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(userID, workDate)
	if buildFindAttendanceDayResult.Error {
		return buildFindAttendanceDayResult
	}

	attendanceDay, findAttendanceDayResult := service.attendanceDayRepository.FindAttendanceDay(findAttendanceDayQuery)
	if findAttendanceDayResult.Error {
		return findAttendanceDayResult
	}

	findAttendanceBreakQuery, buildFindAttendanceBreakResult := service.attendanceBreakBuilder.BuildFindAttendanceBreakByIDAndAttendanceDayIDQuery(
		req.AttendanceBreakID,
		attendanceDay.ID,
	)
	if buildFindAttendanceBreakResult.Error {
		return buildFindAttendanceBreakResult
	}

	currentAttendanceBreak, findAttendanceBreakResult := service.attendanceBreakRepository.FindAttendanceBreak(findAttendanceBreakQuery)
	if findAttendanceBreakResult.Error {
		return findAttendanceBreakResult
	}

	attendanceBreak, buildUpdateResult := service.attendanceBreakBuilder.BuildUpdateAttendanceBreakModel(
		currentAttendanceBreak,
		req,
		breakStartAt,
		breakEndAt,
	)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	savedAttendanceBreak, saveResult := service.attendanceBreakRepository.SaveAttendanceBreak(attendanceBreak)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.UpdateAttendanceBreakResponse{
			AttendanceBreak: toAttendanceBreakResponse(savedAttendanceBreak),
		},
		"UPDATE_ATTENDANCE_BREAK_SUCCESS",
		"休憩を更新しました",
		nil,
	)
}

/*
 * 休憩削除
 *
 * APIとして直接公開しない。
 * monthly_attendances/update の月次全体保存から内部的に使う。
 */
func (service *attendanceBreakService) DeleteAttendanceBreak(
	userID uint,
	req types.DeleteAttendanceBreakRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"DELETE_ATTENDANCE_BREAK_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	workDate, err := utils.ParseDate(req.WorkDate)
	if err != nil {
		return results.BadRequest(
			"DELETE_ATTENDANCE_BREAK_INVALID_WORK_DATE",
			"対象日の形式が正しくありません",
			map[string]any{
				"workDate": req.WorkDate,
				"format":   "yyyy-MM-dd",
			},
		)
	}

	editableResult := service.validateMonthlyAttendanceEditable(
		userID,
		workDate.Year(),
		int(workDate.Month()),
		"DELETE_ATTENDANCE_BREAK",
	)
	if editableResult.Error {
		return editableResult
	}

	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(userID, workDate)
	if buildFindAttendanceDayResult.Error {
		return buildFindAttendanceDayResult
	}

	attendanceDay, findAttendanceDayResult := service.attendanceDayRepository.FindAttendanceDay(findAttendanceDayQuery)
	if findAttendanceDayResult.Error {
		return findAttendanceDayResult
	}

	findAttendanceBreakQuery, buildFindAttendanceBreakResult := service.attendanceBreakBuilder.BuildFindAttendanceBreakByIDAndAttendanceDayIDQuery(
		req.AttendanceBreakID,
		attendanceDay.ID,
	)
	if buildFindAttendanceBreakResult.Error {
		return buildFindAttendanceBreakResult
	}

	currentAttendanceBreak, findAttendanceBreakResult := service.attendanceBreakRepository.FindAttendanceBreak(findAttendanceBreakQuery)
	if findAttendanceBreakResult.Error {
		return findAttendanceBreakResult
	}

	deletedAttendanceBreak, buildDeleteResult := service.attendanceBreakBuilder.BuildDeleteAttendanceBreakModel(currentAttendanceBreak)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	_, saveResult := service.attendanceBreakRepository.SaveAttendanceBreak(deletedAttendanceBreak)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteAttendanceBreakResponse{
			WorkDate:          req.WorkDate,
			AttendanceBreakID: req.AttendanceBreakID,
		},
		"DELETE_ATTENDANCE_BREAK_SUCCESS",
		"休憩を削除しました",
		nil,
	)
}

/*
 * 対象日の休憩を差分保存する
 *
 * monthly_attendances/update の月次全体保存から内部的に使う。
 *
 * 保存方針：
 * ・リクエストにIDがある休憩は更新する
 * ・リクエストにIDがない休憩は新規作成する
 * ・DBに存在するがリクエストから消えた休憩は論理削除する
 *
 * 注意：
 * ・このメソッド自体はAPIとして直接公開しない
 * ・月次申請中、月次承認済みの場合は保存できない
 * ・休憩の所属チェックは userID + workDate で取得した AttendanceDay 配下に限定する
 */
func (service *attendanceBreakService) UpdateAttendanceBreaksByWorkDate(
	userID uint,
	req types.UpdateAttendanceBreaksByWorkDateRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"UPDATE_ATTENDANCE_BREAKS_BY_WORK_DATE_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	workDate, err := utils.ParseDate(req.WorkDate)
	if err != nil {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_BREAKS_BY_WORK_DATE_INVALID_WORK_DATE",
			"対象日の形式が正しくありません",
			map[string]any{
				"workDate": req.WorkDate,
				"format":   "yyyy-MM-dd",
			},
		)
	}

	editableResult := service.validateMonthlyAttendanceEditable(
		userID,
		workDate.Year(),
		int(workDate.Month()),
		"UPDATE_ATTENDANCE_BREAKS_BY_WORK_DATE",
	)
	if editableResult.Error {
		return editableResult
	}

	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(userID, workDate)
	if buildFindAttendanceDayResult.Error {
		return buildFindAttendanceDayResult
	}

	attendanceDay, findAttendanceDayResult := service.attendanceDayRepository.FindAttendanceDay(findAttendanceDayQuery)
	if findAttendanceDayResult.Error {
		return findAttendanceDayResult
	}

	searchAttendanceBreaksQuery, buildSearchAttendanceBreaksResult := service.attendanceBreakBuilder.BuildSearchAttendanceBreaksQuery(attendanceDay.ID)
	if buildSearchAttendanceBreaksResult.Error {
		return buildSearchAttendanceBreaksResult
	}

	currentAttendanceBreaks, findAttendanceBreaksResult := service.attendanceBreakRepository.FindAttendanceBreaks(searchAttendanceBreaksQuery)
	if findAttendanceBreaksResult.Error {
		return findAttendanceBreaksResult
	}

	currentAttendanceBreakMap := make(map[uint]models.AttendanceBreak, len(currentAttendanceBreaks))
	for _, currentAttendanceBreak := range currentAttendanceBreaks {
		currentAttendanceBreakMap[currentAttendanceBreak.ID] = currentAttendanceBreak
	}

	requestedAttendanceBreakIDMap := make(map[uint]bool)

	createdCount := 0
	updatedCount := 0
	deletedCount := 0
	savedCount := 0

	for _, attendanceBreakReq := range req.Breaks {
		if attendanceBreakReq.AttendanceBreakID != nil && *attendanceBreakReq.AttendanceBreakID > 0 {
			requestedAttendanceBreakIDMap[*attendanceBreakReq.AttendanceBreakID] = true

			if _, exists := currentAttendanceBreakMap[*attendanceBreakReq.AttendanceBreakID]; !exists {
				return results.BadRequest(
					"UPDATE_ATTENDANCE_BREAKS_BY_WORK_DATE_BREAK_NOT_IN_TARGET_DAY",
					"対象日の休憩ではないため更新できません",
					map[string]any{
						"workDate":          req.WorkDate,
						"attendanceBreakId": *attendanceBreakReq.AttendanceBreakID,
					},
				)
			}

			updateAttendanceBreakResult := service.UpdateAttendanceBreak(
				userID,
				types.UpdateAttendanceBreakRequest{
					WorkDate:          req.WorkDate,
					AttendanceBreakID: *attendanceBreakReq.AttendanceBreakID,
					BreakStartAt:      attendanceBreakReq.BreakStartAt,
					BreakEndAt:        attendanceBreakReq.BreakEndAt,
					BreakMemo:         attendanceBreakReq.BreakMemo,
				},
			)

			if updateAttendanceBreakResult.Error {
				return updateAttendanceBreakResult
			}

			updatedCount++
			savedCount++

			continue
		}

		createAttendanceBreakResult := service.CreateAttendanceBreak(
			userID,
			types.CreateAttendanceBreakRequest{
				WorkDate:     req.WorkDate,
				BreakStartAt: attendanceBreakReq.BreakStartAt,
				BreakEndAt:   attendanceBreakReq.BreakEndAt,
				BreakMemo:    attendanceBreakReq.BreakMemo,
			},
		)

		if createAttendanceBreakResult.Error {
			return createAttendanceBreakResult
		}

		createdCount++
		savedCount++
	}

	for _, currentAttendanceBreak := range currentAttendanceBreaks {
		if requestedAttendanceBreakIDMap[currentAttendanceBreak.ID] {
			continue
		}

		deleteAttendanceBreakResult := service.DeleteAttendanceBreak(
			userID,
			types.DeleteAttendanceBreakRequest{
				WorkDate:          req.WorkDate,
				AttendanceBreakID: currentAttendanceBreak.ID,
			},
		)

		if deleteAttendanceBreakResult.Error {
			return deleteAttendanceBreakResult
		}

		deletedCount++
	}

	return results.OK(
		types.UpdateAttendanceBreaksByWorkDateResponse{
			WorkDate:                    req.WorkDate,
			SavedAttendanceBreakCount:   savedCount,
			CreatedAttendanceBreakCount: createdCount,
			UpdatedAttendanceBreakCount: updatedCount,
			DeletedAttendanceBreakCount: deletedCount,
		},
		"UPDATE_ATTENDANCE_BREAKS_BY_WORK_DATE_SUCCESS",
		"休憩を保存しました",
		nil,
	)
}
