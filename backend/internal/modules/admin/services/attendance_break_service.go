package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 管理者用休憩Service interface
 *
 * Controllerや月次勤怠全体保存ServiceがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・管理者APIでは対象ユーザーIDを targetUserId としてRequestで受け取る
 * ・ControllerではJWTのuserIdを対象ユーザーIDとして使わない
 * ・AttendanceBreak は申請状態を持たない
 * ・管理者側では MonthlyAttendanceRequest の状態による編集ロックを行わない
 */
type AttendanceBreakService interface {
	SearchAttendanceBreaks(req types.SearchAttendanceBreaksRequest) results.Result
	CreateAttendanceBreak(req types.CreateAttendanceBreakRequest) results.Result
	UpdateAttendanceBreak(req types.UpdateAttendanceBreakRequest) results.Result
	DeleteAttendanceBreak(req types.DeleteAttendanceBreakRequest) results.Result
	UpdateAttendanceBreaksByWorkDate(req types.UpdateAttendanceBreaksByWorkDateRequest) results.Result
}

/*
 * 管理者用休憩Service
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
 * ・対象月の申請状態は MonthlyAttendanceRequest 側で管理する
 * ・管理者側では月次申請状態による編集ロックを行わない
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
 * 対象ユーザーIDのバリデーション
 */
func validateAttendanceBreakTargetUserID(
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
 * 休憩検索
 *
 * targetUserId + workDate で勤怠日を取得し、その勤怠日に紐づく休憩一覧を取得する。
 */
func (service *attendanceBreakService) SearchAttendanceBreaks(
	req types.SearchAttendanceBreaksRequest,
) results.Result {
	validateUserResult := validateAttendanceBreakTargetUserID(
		req.TargetUserID,
		"SEARCH_ATTENDANCE_BREAKS",
	)
	if validateUserResult.Error {
		return validateUserResult
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

	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(req.TargetUserID, workDate)
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
			TargetUserID:     req.TargetUserID,
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
 *
 * 注意：
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
func (service *attendanceBreakService) CreateAttendanceBreak(
	req types.CreateAttendanceBreakRequest,
) results.Result {
	validateUserResult := validateAttendanceBreakTargetUserID(
		req.TargetUserID,
		"CREATE_ATTENDANCE_BREAK",
	)
	if validateUserResult.Error {
		return validateUserResult
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

	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(req.TargetUserID, workDate)
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
 *
 * 注意：
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
func (service *attendanceBreakService) UpdateAttendanceBreak(
	req types.UpdateAttendanceBreakRequest,
) results.Result {
	validateUserResult := validateAttendanceBreakTargetUserID(
		req.TargetUserID,
		"UPDATE_ATTENDANCE_BREAK",
	)
	if validateUserResult.Error {
		return validateUserResult
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

	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(req.TargetUserID, workDate)
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
 *
 * 注意：
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
func (service *attendanceBreakService) DeleteAttendanceBreak(
	req types.DeleteAttendanceBreakRequest,
) results.Result {
	validateUserResult := validateAttendanceBreakTargetUserID(
		req.TargetUserID,
		"DELETE_ATTENDANCE_BREAK",
	)
	if validateUserResult.Error {
		return validateUserResult
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

	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(req.TargetUserID, workDate)
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
			TargetUserID:      req.TargetUserID,
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
 * ・管理者側では月次申請状態による編集ロックを行わない
 * ・休憩の所属チェックは targetUserId + workDate で取得した AttendanceDay 配下に限定する
 */
func (service *attendanceBreakService) UpdateAttendanceBreaksByWorkDate(
	req types.UpdateAttendanceBreaksByWorkDateRequest,
) results.Result {
	validateUserResult := validateAttendanceBreakTargetUserID(
		req.TargetUserID,
		"UPDATE_ATTENDANCE_BREAKS_BY_WORK_DATE",
	)
	if validateUserResult.Error {
		return validateUserResult
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

	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(req.TargetUserID, workDate)
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
						"targetUserId":      req.TargetUserID,
						"workDate":          req.WorkDate,
						"attendanceBreakId": *attendanceBreakReq.AttendanceBreakID,
					},
				)
			}

			updateAttendanceBreakResult := service.UpdateAttendanceBreak(
				types.UpdateAttendanceBreakRequest{
					TargetUserID:      req.TargetUserID,
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
			types.CreateAttendanceBreakRequest{
				TargetUserID: req.TargetUserID,
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
			types.DeleteAttendanceBreakRequest{
				TargetUserID:      req.TargetUserID,
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
			TargetUserID:                req.TargetUserID,
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
