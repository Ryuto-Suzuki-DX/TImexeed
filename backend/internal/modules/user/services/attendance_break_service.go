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
 * ControllerがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・従業員APIでは userId / targetUserId をRequestで受け取らない
 * ・ControllerでAuthMiddleware由来のuserIdを取得し、Serviceへ渡す
 */
type AttendanceBreakService interface {
	SearchAttendanceBreaks(userID uint, req types.SearchAttendanceBreaksRequest) results.Result
	CreateAttendanceBreak(userID uint, req types.CreateAttendanceBreakRequest) results.Result
	UpdateAttendanceBreak(userID uint, req types.UpdateAttendanceBreakRequest) results.Result
	DeleteAttendanceBreak(userID uint, req types.DeleteAttendanceBreakRequest) results.Result
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
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type attendanceBreakService struct {
	attendanceBreakBuilder    builders.AttendanceBreakBuilder
	attendanceBreakRepository repositories.AttendanceBreakRepository
	attendanceDayBuilder      builders.AttendanceDayBuilder
	attendanceDayRepository   repositories.AttendanceDayRepository
}

/*
 * AttendanceBreakService生成
 */
func NewAttendanceBreakService(
	attendanceBreakBuilder builders.AttendanceBreakBuilder,
	attendanceBreakRepository repositories.AttendanceBreakRepository,
	attendanceDayBuilder builders.AttendanceDayBuilder,
	attendanceDayRepository repositories.AttendanceDayRepository,
) *attendanceBreakService {
	return &attendanceBreakService{
		attendanceBreakBuilder:    attendanceBreakBuilder,
		attendanceBreakRepository: attendanceBreakRepository,
		attendanceDayBuilder:      attendanceDayBuilder,
		attendanceDayRepository:   attendanceDayRepository,
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

	// 対象日を日付型へ変換する
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

	// Builderで対象勤怠取得用クエリを作成する
	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(userID, workDate)
	if buildFindAttendanceDayResult.Error {
		return buildFindAttendanceDayResult
	}

	// Repositoryで対象勤怠を取得する
	attendanceDay, findAttendanceDayResult := service.attendanceDayRepository.FindAttendanceDay(findAttendanceDayQuery)
	if findAttendanceDayResult.Error {
		return findAttendanceDayResult
	}

	// Builderで休憩検索用クエリを作成する
	query, buildResult := service.attendanceBreakBuilder.BuildSearchAttendanceBreaksQuery(attendanceDay.ID)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryで休憩一覧を取得する
	attendanceBreaks, findResult := service.attendanceBreakRepository.FindAttendanceBreaks(query)
	if findResult.Error {
		return findResult
	}

	// DBモデルをフロント返却用Responseへ変換する
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
 * userID + workDate で勤怠日を取得し、その勤怠日に休憩を作成する。
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

	// 対象日を日付型へ変換する
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

	// 休憩開始日時を変換する
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

	// 休憩終了日時を変換する
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

	// Builderで対象勤怠取得用クエリを作成する
	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(userID, workDate)
	if buildFindAttendanceDayResult.Error {
		return buildFindAttendanceDayResult
	}

	// Repositoryで対象勤怠を取得する
	attendanceDay, findAttendanceDayResult := service.attendanceDayRepository.FindAttendanceDay(findAttendanceDayQuery)
	if findAttendanceDayResult.Error {
		return findAttendanceDayResult
	}

	// 既存勤怠がある場合は、変更可能か確認する
	editableResult := validateAttendanceDayEditable(attendanceDay, "CREATE_ATTENDANCE_BREAK")
	if editableResult.Error {
		return editableResult
	}

	// Builderで休憩作成用Modelを作る
	attendanceBreak, buildCreateResult := service.attendanceBreakBuilder.BuildCreateAttendanceBreakModel(
		attendanceDay.ID,
		req,
		breakStartAt,
		breakEndAt,
	)
	if buildCreateResult.Error {
		return buildCreateResult
	}

	// Repositoryで休憩を作成する
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
 * userID + workDate で勤怠日を取得し、指定された休憩を更新する。
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

	// 対象日を日付型へ変換する
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

	// 休憩開始日時を変換する
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

	// 休憩終了日時を変換する
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

	// Builderで対象勤怠取得用クエリを作成する
	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(userID, workDate)
	if buildFindAttendanceDayResult.Error {
		return buildFindAttendanceDayResult
	}

	// Repositoryで対象勤怠を取得する
	attendanceDay, findAttendanceDayResult := service.attendanceDayRepository.FindAttendanceDay(findAttendanceDayQuery)
	if findAttendanceDayResult.Error {
		return findAttendanceDayResult
	}

	// 既存勤怠がある場合は、変更可能か確認する
	editableResult := validateAttendanceDayEditable(attendanceDay, "UPDATE_ATTENDANCE_BREAK")
	if editableResult.Error {
		return editableResult
	}

	// Builderで対象休憩取得用クエリを作成する
	findAttendanceBreakQuery, buildFindAttendanceBreakResult := service.attendanceBreakBuilder.BuildFindAttendanceBreakByIDAndAttendanceDayIDQuery(
		req.AttendanceBreakID,
		attendanceDay.ID,
	)
	if buildFindAttendanceBreakResult.Error {
		return buildFindAttendanceBreakResult
	}

	// Repositoryで対象休憩を取得する
	currentAttendanceBreak, findAttendanceBreakResult := service.attendanceBreakRepository.FindAttendanceBreak(findAttendanceBreakQuery)
	if findAttendanceBreakResult.Error {
		return findAttendanceBreakResult
	}

	// Builderで休憩更新用Modelを作る
	attendanceBreak, buildUpdateResult := service.attendanceBreakBuilder.BuildUpdateAttendanceBreakModel(
		currentAttendanceBreak,
		req,
		breakStartAt,
		breakEndAt,
	)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	// Repositoryで休憩を保存する
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
 * userID + workDate で勤怠日を取得し、指定された休憩を論理削除する。
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

	// 対象日を日付型へ変換する
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

	// Builderで対象勤怠取得用クエリを作成する
	findAttendanceDayQuery, buildFindAttendanceDayResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(userID, workDate)
	if buildFindAttendanceDayResult.Error {
		return buildFindAttendanceDayResult
	}

	// Repositoryで対象勤怠を取得する
	attendanceDay, findAttendanceDayResult := service.attendanceDayRepository.FindAttendanceDay(findAttendanceDayQuery)
	if findAttendanceDayResult.Error {
		return findAttendanceDayResult
	}

	// 既存勤怠がある場合は、変更可能か確認する
	editableResult := validateAttendanceDayEditable(attendanceDay, "DELETE_ATTENDANCE_BREAK")
	if editableResult.Error {
		return editableResult
	}

	// Builderで対象休憩取得用クエリを作成する
	findAttendanceBreakQuery, buildFindAttendanceBreakResult := service.attendanceBreakBuilder.BuildFindAttendanceBreakByIDAndAttendanceDayIDQuery(
		req.AttendanceBreakID,
		attendanceDay.ID,
	)
	if buildFindAttendanceBreakResult.Error {
		return buildFindAttendanceBreakResult
	}

	// Repositoryで対象休憩を取得する
	currentAttendanceBreak, findAttendanceBreakResult := service.attendanceBreakRepository.FindAttendanceBreak(findAttendanceBreakQuery)
	if findAttendanceBreakResult.Error {
		return findAttendanceBreakResult
	}

	// Builderで休憩論理削除用Modelを作る
	deletedAttendanceBreak, buildDeleteResult := service.attendanceBreakBuilder.BuildDeleteAttendanceBreakModel(currentAttendanceBreak)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	// Repositoryで休憩を保存する
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
