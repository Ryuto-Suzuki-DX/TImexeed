package services

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 従業員用勤怠Service interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・従業員APIでは userId / targetUserId をRequestで受け取らない
 * ・ControllerでAuthMiddleware由来のuserIdを取得し、Serviceへ渡す
 */
type AttendanceDayService interface {
	SearchAttendanceDays(userID uint, req types.SearchAttendanceDaysRequest) results.Result
	UpdateAttendanceDay(userID uint, req types.UpdateAttendanceDayRequest) results.Result
	DeleteAttendanceDay(userID uint, req types.DeleteAttendanceDayRequest) results.Result
}

/*
 * 従業員用勤怠Service
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
type attendanceDayService struct {
	attendanceDayBuilder     builders.AttendanceDayBuilder
	attendanceDayRepository  repositories.AttendanceDayRepository
	attendanceTypeRepository repositories.AttendanceTypeRepository
}

/*
 * AttendanceDayService生成
 */
func NewAttendanceDayService(
	attendanceDayBuilder builders.AttendanceDayBuilder,
	attendanceDayRepository repositories.AttendanceDayRepository,
	attendanceTypeRepository repositories.AttendanceTypeRepository,
) *attendanceDayService {
	return &attendanceDayService{
		attendanceDayBuilder:     attendanceDayBuilder,
		attendanceDayRepository:  attendanceDayRepository,
		attendanceTypeRepository: attendanceTypeRepository,
	}
}

/*
 * models.AttendanceDayをフロント返却用AttendanceDayResponseへ変換する
 *
 * 日付はtime.Time / *time.Timeのまま返す。
 * 表示形式の整形はフロント側で行う。
 */
func toAttendanceDayResponse(attendanceDay models.AttendanceDay) types.AttendanceDayResponse {
	return types.AttendanceDayResponse{
		ID:       attendanceDay.ID,
		WorkDate: attendanceDay.WorkDate,

		PlanAttendanceTypeID:   attendanceDay.PlanAttendanceTypeID,
		ActualAttendanceTypeID: attendanceDay.ActualAttendanceTypeID,

		PlanStartAt:   attendanceDay.PlanStartAt,
		PlanEndAt:     attendanceDay.PlanEndAt,
		ActualStartAt: attendanceDay.ActualStartAt,
		ActualEndAt:   attendanceDay.ActualEndAt,

		RequestStatus:  attendanceDay.RequestStatus,
		RequestMemo:    attendanceDay.RequestMemo,
		ApprovedBy:     attendanceDay.ApprovedBy,
		ApprovedAt:     attendanceDay.ApprovedAt,
		RejectedReason: attendanceDay.RejectedReason,

		LateFlag:       attendanceDay.LateFlag,
		EarlyLeaveFlag: attendanceDay.EarlyLeaveFlag,
		AbsenceFlag:    attendanceDay.AbsenceFlag,
		SickLeaveFlag:  attendanceDay.SickLeaveFlag,

		SystemMessage: attendanceDay.SystemMessage,

		TransportFrom:   attendanceDay.TransportFrom,
		TransportTo:     attendanceDay.TransportTo,
		TransportMethod: attendanceDay.TransportMethod,
		TransportAmount: attendanceDay.TransportAmount,

		MonthlyStatus: attendanceDay.MonthlyStatus,

		IsDeleted: attendanceDay.IsDeleted,
		CreatedAt: attendanceDay.CreatedAt,
		UpdatedAt: attendanceDay.UpdatedAt,
		DeletedAt: attendanceDay.DeletedAt,
	}
}

/*
 * 勤怠検索
 *
 * 対象年月のログイン中ユーザー本人の勤怠を取得する。
 */
func (service *attendanceDayService) SearchAttendanceDays(
	userID uint,
	req types.SearchAttendanceDaysRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"SEARCH_ATTENDANCE_DAYS_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	if req.TargetYear <= 0 {
		return results.BadRequest(
			"SEARCH_ATTENDANCE_DAYS_INVALID_TARGET_YEAR",
			"対象年が正しくありません",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return results.BadRequest(
			"SEARCH_ATTENDANCE_DAYS_INVALID_TARGET_MONTH",
			"対象月が正しくありません",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	// Builderで勤怠検索用クエリを作成する
	query, buildResult := service.attendanceDayBuilder.BuildSearchAttendanceDaysQuery(userID, req)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryで勤怠一覧を取得する
	attendanceDays, findResult := service.attendanceDayRepository.FindAttendanceDays(query)
	if findResult.Error {
		return findResult
	}

	// DBモデルをフロント返却用Responseへ変換する
	attendanceDayResponses := make([]types.AttendanceDayResponse, 0, len(attendanceDays))
	for _, attendanceDay := range attendanceDays {
		attendanceDayResponses = append(attendanceDayResponses, toAttendanceDayResponse(attendanceDay))
	}

	return results.OK(
		types.SearchAttendanceDaysResponse{
			TargetYear:     req.TargetYear,
			TargetMonth:    req.TargetMonth,
			AttendanceDays: attendanceDayResponses,
		},
		"SEARCH_ATTENDANCE_DAYS_SUCCESS",
		"勤怠一覧を取得しました",
		nil,
	)
}

/*
 * 勤怠更新
 *
 * 画面上は月次一覧の1行を更新する操作。
 *
 * 仕様：
 * ・userID + workDate で既存勤怠を検索する
 * ・存在しなければ新規作成する
 * ・存在すれば更新する
 * ・休日は予定・実績ともに時間を保存しない
 * ・syncPlanActual = true の勤務区分は、commonStartAt / commonEndAt を plan / actual の両方へ反映する
 */
func (service *attendanceDayService) UpdateAttendanceDay(
	userID uint,
	req types.UpdateAttendanceDayRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"UPDATE_ATTENDANCE_DAY_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	// 対象日を日付型へ変換する
	workDate, err := utils.ParseDate(req.WorkDate)
	if err != nil {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_DAY_INVALID_WORK_DATE",
			"対象日の形式が正しくありません",
			map[string]any{
				"workDate": req.WorkDate,
				"format":   "yyyy-MM-dd",
			},
		)
	}

	// 選択された予定勤務区分を取得する
	attendanceType, findAttendanceTypeResult := service.attendanceTypeRepository.FindAttendanceTypeByID(req.PlanAttendanceTypeID)
	if findAttendanceTypeResult.Error {
		return findAttendanceTypeResult
	}

	var actualAttendanceTypeID uint
	var planStartAt *time.Time
	var planEndAt *time.Time
	var actualStartAt *time.Time
	var actualEndAt *time.Time

	/*
	 * 休日の場合
	 *
	 * 休日だけは予定にも実績にも時間を保存しない。
	 * syncPlanActual=true でも commonStartAt / commonEndAt は要求しない。
	 */
	if attendanceType.Code == "HOLIDAY" {
		actualAttendanceTypeID = req.PlanAttendanceTypeID

		planStartAt = nil
		planEndAt = nil
		actualStartAt = nil
		actualEndAt = nil

		req.ActualAttendanceTypeID = nil

		req.CommonStartAt = nil
		req.CommonEndAt = nil

		req.PlanStartAt = nil
		req.PlanEndAt = nil

		req.ActualStartAt = nil
		req.ActualEndAt = nil

		req.LateFlag = false
		req.EarlyLeaveFlag = false
		req.AbsenceFlag = false
		req.SickLeaveFlag = false

		req.TransportFrom = nil
		req.TransportTo = nil
		req.TransportMethod = nil
		req.TransportAmount = nil
	} else if attendanceType.SyncPlanActual {
		/*
		 * 予定・実績同期対象の場合
		 *
		 * 有給、欠勤、病欠、休職、介護休業などは、
		 * commonStartAt / commonEndAt を plan / actual の両方へ反映する。
		 */
		commonStartAt, err := utils.ParseOptionalDateTime(req.CommonStartAt)
		if err != nil {
			return results.BadRequest(
				"UPDATE_ATTENDANCE_DAY_INVALID_COMMON_START_AT",
				"共通開始日時の形式が正しくありません",
				map[string]any{
					"commonStartAt": req.CommonStartAt,
					"format":        "RFC3339",
				},
			)
		}

		commonEndAt, err := utils.ParseOptionalDateTime(req.CommonEndAt)
		if err != nil {
			return results.BadRequest(
				"UPDATE_ATTENDANCE_DAY_INVALID_COMMON_END_AT",
				"共通終了日時の形式が正しくありません",
				map[string]any{
					"commonEndAt": req.CommonEndAt,
					"format":      "RFC3339",
				},
			)
		}

		if commonStartAt == nil || commonEndAt == nil {
			return results.BadRequest(
				"UPDATE_ATTENDANCE_DAY_EMPTY_COMMON_TIME",
				"共通時間を入力してください",
				nil,
			)
		}

		actualAttendanceTypeID = req.PlanAttendanceTypeID
		planStartAt = commonStartAt
		planEndAt = commonEndAt
		actualStartAt = commonStartAt
		actualEndAt = commonEndAt

		req.LateFlag = false
		req.EarlyLeaveFlag = false
		req.AbsenceFlag = false
		req.SickLeaveFlag = false
	} else {
		/*
		 * 通常勤務・夜勤など、予定と実績を分ける区分。
		 */
		if req.ActualAttendanceTypeID == nil || *req.ActualAttendanceTypeID == 0 {
			return results.BadRequest(
				"UPDATE_ATTENDANCE_DAY_EMPTY_ACTUAL_ATTENDANCE_TYPE_ID",
				"実績区分を選択してください",
				nil,
			)
		}

		// 選択された実績勤務区分を取得する
		_, findActualAttendanceTypeResult := service.attendanceTypeRepository.FindAttendanceTypeByID(*req.ActualAttendanceTypeID)
		if findActualAttendanceTypeResult.Error {
			return findActualAttendanceTypeResult
		}

		parsedPlanStartAt, err := utils.ParseOptionalDateTime(req.PlanStartAt)
		if err != nil {
			return results.BadRequest(
				"UPDATE_ATTENDANCE_DAY_INVALID_PLAN_START_AT",
				"予定開始日時の形式が正しくありません",
				map[string]any{
					"planStartAt": req.PlanStartAt,
					"format":      "RFC3339",
				},
			)
		}

		parsedPlanEndAt, err := utils.ParseOptionalDateTime(req.PlanEndAt)
		if err != nil {
			return results.BadRequest(
				"UPDATE_ATTENDANCE_DAY_INVALID_PLAN_END_AT",
				"予定終了日時の形式が正しくありません",
				map[string]any{
					"planEndAt": req.PlanEndAt,
					"format":    "RFC3339",
				},
			)
		}

		parsedActualStartAt, err := utils.ParseOptionalDateTime(req.ActualStartAt)
		if err != nil {
			return results.BadRequest(
				"UPDATE_ATTENDANCE_DAY_INVALID_ACTUAL_START_AT",
				"実績開始日時の形式が正しくありません",
				map[string]any{
					"actualStartAt": req.ActualStartAt,
					"format":        "RFC3339",
				},
			)
		}

		parsedActualEndAt, err := utils.ParseOptionalDateTime(req.ActualEndAt)
		if err != nil {
			return results.BadRequest(
				"UPDATE_ATTENDANCE_DAY_INVALID_ACTUAL_END_AT",
				"実績終了日時の形式が正しくありません",
				map[string]any{
					"actualEndAt": req.ActualEndAt,
					"format":      "RFC3339",
				},
			)
		}

		if parsedPlanStartAt == nil || parsedPlanEndAt == nil {
			return results.BadRequest(
				"UPDATE_ATTENDANCE_DAY_EMPTY_PLAN_TIME",
				"予定時間を入力してください",
				nil,
			)
		}

		if parsedActualStartAt == nil || parsedActualEndAt == nil {
			return results.BadRequest(
				"UPDATE_ATTENDANCE_DAY_EMPTY_ACTUAL_TIME",
				"実績時間を入力してください",
				nil,
			)
		}

		actualAttendanceTypeID = *req.ActualAttendanceTypeID
		planStartAt = parsedPlanStartAt
		planEndAt = parsedPlanEndAt
		actualStartAt = parsedActualStartAt
		actualEndAt = parsedActualEndAt
	}

	// Builderで対象勤怠取得用クエリを作成する
	findQuery, buildFindResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(userID, workDate)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象勤怠を取得する
	currentAttendanceDay, findResult := service.attendanceDayRepository.FindAttendanceDay(findQuery)

	// 対象日の勤怠が存在しない場合は新規作成する
	if findResult.Error && findResult.Code == "ATTENDANCE_DAY_NOT_FOUND" {
		attendanceDay, buildCreateResult := service.attendanceDayBuilder.BuildCreateAttendanceDayModel(
			userID,
			req,
			workDate,
			planStartAt,
			planEndAt,
			actualStartAt,
			actualEndAt,
			actualAttendanceTypeID,
		)
		if buildCreateResult.Error {
			return buildCreateResult
		}

		createdAttendanceDay, createResult := service.attendanceDayRepository.CreateAttendanceDay(attendanceDay)
		if createResult.Error {
			return createResult
		}

		return results.Created(
			types.UpdateAttendanceDayResponse{
				AttendanceDay: toAttendanceDayResponse(createdAttendanceDay),
			},
			"CREATE_ATTENDANCE_DAY_SUCCESS",
			"勤怠を作成しました",
			nil,
		)
	}

	if findResult.Error {
		return findResult
	}

	// 既存勤怠がある場合は、更新可能か確認する
	editableResult := validateAttendanceDayEditable(currentAttendanceDay, "UPDATE_ATTENDANCE_DAY")
	if editableResult.Error {
		return editableResult
	}

	// 対象日の勤怠が存在する場合は更新する
	attendanceDay, buildUpdateResult := service.attendanceDayBuilder.BuildUpdateAttendanceDayModel(
		currentAttendanceDay,
		req,
		workDate,
		planStartAt,
		planEndAt,
		actualStartAt,
		actualEndAt,
		actualAttendanceTypeID,
	)
	if buildUpdateResult.Error {
		return buildUpdateResult
	}

	savedAttendanceDay, saveResult := service.attendanceDayRepository.SaveAttendanceDay(attendanceDay)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.UpdateAttendanceDayResponse{
			AttendanceDay: toAttendanceDayResponse(savedAttendanceDay),
		},
		"UPDATE_ATTENDANCE_DAY_SUCCESS",
		"勤怠を更新しました",
		nil,
	)
}

/*
 * 勤怠削除
 *
 * userID + workDate で対象勤怠を取得し、論理削除する。
 */
func (service *attendanceDayService) DeleteAttendanceDay(
	userID uint,
	req types.DeleteAttendanceDayRequest,
) results.Result {
	if userID == 0 {
		return results.Unauthorized(
			"DELETE_ATTENDANCE_DAY_INVALID_USER_ID",
			"認証情報のユーザーIDが正しくありません",
			nil,
		)
	}

	// 対象日を日付型へ変換する
	workDate, err := utils.ParseDate(req.WorkDate)
	if err != nil {
		return results.BadRequest(
			"DELETE_ATTENDANCE_DAY_INVALID_WORK_DATE",
			"対象日の形式が正しくありません",
			map[string]any{
				"workDate": req.WorkDate,
				"format":   "yyyy-MM-dd",
			},
		)
	}

	// Builderで対象勤怠取得用クエリを作成する
	findQuery, buildFindResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(userID, workDate)
	if buildFindResult.Error {
		return buildFindResult
	}

	// Repositoryで対象勤怠を取得する
	currentAttendanceDay, findResult := service.attendanceDayRepository.FindAttendanceDay(findQuery)
	if findResult.Error {
		return findResult
	}

	// 既存勤怠がある場合は、削除可能か確認する
	editableResult := validateAttendanceDayEditable(currentAttendanceDay, "DELETE_ATTENDANCE_DAY")
	if editableResult.Error {
		return editableResult
	}

	// Builderで論理削除用Modelを作る
	deletedAttendanceDay, buildDeleteResult := service.attendanceDayBuilder.BuildDeleteAttendanceDayModel(currentAttendanceDay)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	// Repositoryで勤怠を保存する
	_, saveResult := service.attendanceDayRepository.SaveAttendanceDay(deletedAttendanceDay)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteAttendanceDayResponse{
			WorkDate: req.WorkDate,
		},
		"DELETE_ATTENDANCE_DAY_SUCCESS",
		"勤怠を削除しました",
		nil,
	)
}

/*
 * 勤怠更新・削除が可能か確認する
 *
 * 月次申請中・月次承認済み・休暇申請中・休暇承認済みの場合は、
 * 従業員側から更新・削除できない。
 */
func validateAttendanceDayEditable(attendanceDay models.AttendanceDay, actionCode string) results.Result {
	if attendanceDay.MonthlyStatus == "PENDING" {
		return results.Conflict(
			actionCode+"_MONTHLY_STATUS_PENDING",
			"月次申請中のため、勤怠を変更できません",
			map[string]any{
				"attendanceDayId": attendanceDay.ID,
				"monthlyStatus":   attendanceDay.MonthlyStatus,
			},
		)
	}

	if attendanceDay.MonthlyStatus == "APPROVED" {
		return results.Conflict(
			actionCode+"_MONTHLY_STATUS_APPROVED",
			"月次承認済みのため、勤怠を変更できません",
			map[string]any{
				"attendanceDayId": attendanceDay.ID,
				"monthlyStatus":   attendanceDay.MonthlyStatus,
			},
		)
	}

	if attendanceDay.RequestStatus == "PENDING" {
		return results.Conflict(
			actionCode+"_REQUEST_STATUS_PENDING",
			"申請中の勤怠のため、変更できません",
			map[string]any{
				"attendanceDayId": attendanceDay.ID,
				"requestStatus":   attendanceDay.RequestStatus,
			},
		)
	}

	if attendanceDay.RequestStatus == "APPROVED" {
		return results.Conflict(
			actionCode+"_REQUEST_STATUS_APPROVED",
			"承認済みの申請があるため、変更できません",
			map[string]any{
				"attendanceDayId": attendanceDay.ID,
				"requestStatus":   attendanceDay.RequestStatus,
			},
		)
	}

	return results.OK(
		nil,
		actionCode+"_EDITABLE",
		"",
		nil,
	)
}
