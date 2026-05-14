package services

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"
	"timexeed/backend/internal/utils"
)

/*
 * 管理者用勤怠Service interface
 *
 * Controllerや月次勤怠全体保存ServiceがServiceに求める処理だけを定義する。
 *
 * 注意：
 * ・管理者APIでは対象ユーザーIDを targetUserId としてRequestで受け取る
 * ・ControllerではJWTのuserIdを対象ユーザーIDとして使わない
 * ・AttendanceDay は申請状態を持たない
 * ・管理者側では MonthlyAttendanceRequest の状態による編集ロックを行わない
 */
type AttendanceDayService interface {
	SearchAttendanceDays(req types.SearchAttendanceDaysRequest) results.Result
	UpdateAttendanceDay(req types.UpdateAttendanceDayRequest) results.Result
	DeleteAttendanceDay(req types.DeleteAttendanceDayRequest) results.Result
}

/*
 * 管理者用勤怠Service
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
 * ・AttendanceDay 自体は申請状態を持たない
 * ・対象月の申請状態は MonthlyAttendanceRequest から取得する
 * ・MonthlyAttendanceRequest が存在しない場合は未申請扱いにする
 * ・管理者側では月次申請状態を表示用として返す
 * ・管理者側では月次申請状態による編集ロックを行わない
 *
 * 画面表示用メッセージ方針：
 * ・AttendanceDay には SystemMessage を保存しない
 * ・残業、深夜勤務、有給、月次申請状態などの画面表示用メッセージは、
 *   DB保存値ではなく、勤怠データ・休憩データ・月次申請状態などから
 *   表示時に組み立てる
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type attendanceDayService struct {
	attendanceDayBuilder               builders.AttendanceDayBuilder
	attendanceDayRepository            repositories.AttendanceDayRepository
	attendanceTypeRepository           repositories.AttendanceTypeRepository
	monthlyAttendanceRequestBuilder    builders.MonthlyAttendanceRequestBuilder
	monthlyAttendanceRequestRepository repositories.MonthlyAttendanceRequestRepository
}

/*
 * AttendanceDayService生成
 */
func NewAttendanceDayService(
	attendanceDayBuilder builders.AttendanceDayBuilder,
	attendanceDayRepository repositories.AttendanceDayRepository,
	attendanceTypeRepository repositories.AttendanceTypeRepository,
	monthlyAttendanceRequestBuilder builders.MonthlyAttendanceRequestBuilder,
	monthlyAttendanceRequestRepository repositories.MonthlyAttendanceRequestRepository,
) *attendanceDayService {
	return &attendanceDayService{
		attendanceDayBuilder:               attendanceDayBuilder,
		attendanceDayRepository:            attendanceDayRepository,
		attendanceTypeRepository:           attendanceTypeRepository,
		monthlyAttendanceRequestBuilder:    monthlyAttendanceRequestBuilder,
		monthlyAttendanceRequestRepository: monthlyAttendanceRequestRepository,
	}
}

/*
 * models.AttendanceDayをフロント返却用AttendanceDayResponseへ変換する
 *
 * AttendanceDay は申請状態を持たない。
 * 月次申請状態は SearchAttendanceDaysResponse の上位に載せる。
 *
 * 注意：
 * ・AttendanceDay model から SystemMessage は削除済み
 * ・ここでは SystemMessage を詰めない
 */
func toAttendanceDayResponse(attendanceDay models.AttendanceDay) types.AttendanceDayResponse {
	return types.AttendanceDayResponse{
		ID:     attendanceDay.ID,
		UserID: attendanceDay.UserID,

		WorkDate: attendanceDay.WorkDate,

		PlanAttendanceTypeID:   attendanceDay.PlanAttendanceTypeID,
		ActualAttendanceTypeID: attendanceDay.ActualAttendanceTypeID,

		PlanStartAt:   attendanceDay.PlanStartAt,
		PlanEndAt:     attendanceDay.PlanEndAt,
		ActualStartAt: attendanceDay.ActualStartAt,
		ActualEndAt:   attendanceDay.ActualEndAt,

		LateFlag:       attendanceDay.LateFlag,
		EarlyLeaveFlag: attendanceDay.EarlyLeaveFlag,
		AbsenceFlag:    attendanceDay.AbsenceFlag,
		SickLeaveFlag:  attendanceDay.SickLeaveFlag,

		RemoteWorkAllowanceFlag: attendanceDay.RemoteWorkAllowanceFlag,

		TransportFrom:   attendanceDay.TransportFrom,
		TransportTo:     attendanceDay.TransportTo,
		TransportMethod: attendanceDay.TransportMethod,
		TransportAmount: attendanceDay.TransportAmount,

		IsDeleted: attendanceDay.IsDeleted,
		CreatedAt: attendanceDay.CreatedAt,
		UpdatedAt: attendanceDay.UpdatedAt,
		DeletedAt: attendanceDay.DeletedAt,
	}
}

/*
 * 対象月の月次勤怠申請状態を取得する
 *
 * MonthlyAttendanceRequest が存在しない場合は未申請として返す。
 *
 * 注意：
 * ・管理者側ではこの状態を編集ロックには使わない
 * ・画面表示用として返す
 */
func (service *attendanceDayService) getMonthlyAttendanceRequestResponse(
	targetUserID uint,
	targetYear int,
	targetMonth int,
) (types.MonthlyAttendanceRequestResponse, results.Result) {
	query, buildResult := service.monthlyAttendanceRequestBuilder.BuildFindMonthlyAttendanceRequestByUserIDAndTargetYearMonthQuery(
		targetUserID,
		targetYear,
		targetMonth,
	)
	if buildResult.Error {
		return types.MonthlyAttendanceRequestResponse{}, buildResult
	}

	monthlyAttendanceRequest, findResult := service.monthlyAttendanceRequestRepository.FindMonthlyAttendanceRequest(query)

	if findResult.Error && findResult.Code == "MONTHLY_ATTENDANCE_REQUEST_NOT_FOUND" {
		return toNotSubmittedMonthlyAttendanceRequestResponse(targetUserID, targetYear, targetMonth), results.OK(
			nil,
			"GET_MONTHLY_ATTENDANCE_REQUEST_FOR_ATTENDANCE_DAY_NOT_SUBMITTED",
			"",
			nil,
		)
	}

	if findResult.Error {
		return types.MonthlyAttendanceRequestResponse{}, findResult
	}

	return toMonthlyAttendanceRequestResponse(monthlyAttendanceRequest), results.OK(
		nil,
		"GET_MONTHLY_ATTENDANCE_REQUEST_FOR_ATTENDANCE_DAY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠検索
 *
 * 対象年月の対象ユーザーの勤怠を取得する。
 * 対象月の月次申請状態も一緒に返す。
 */
func (service *attendanceDayService) SearchAttendanceDays(
	req types.SearchAttendanceDaysRequest,
) results.Result {
	if req.TargetUserID == 0 {
		return results.BadRequest(
			"SEARCH_ATTENDANCE_DAYS_INVALID_TARGET_USER_ID",
			"対象ユーザーIDが正しくありません",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
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

	monthlyAttendanceRequestResponse, monthlyAttendanceRequestResult := service.getMonthlyAttendanceRequestResponse(
		req.TargetUserID,
		req.TargetYear,
		req.TargetMonth,
	)
	if monthlyAttendanceRequestResult.Error {
		return monthlyAttendanceRequestResult
	}

	query, buildResult := service.attendanceDayBuilder.BuildSearchAttendanceDaysQuery(req)
	if buildResult.Error {
		return buildResult
	}

	attendanceDays, findResult := service.attendanceDayRepository.FindAttendanceDays(query)
	if findResult.Error {
		return findResult
	}

	attendanceDayResponses := make([]types.AttendanceDayResponse, 0, len(attendanceDays))
	for _, attendanceDay := range attendanceDays {
		attendanceDayResponses = append(attendanceDayResponses, toAttendanceDayResponse(attendanceDay))
	}

	return results.OK(
		types.SearchAttendanceDaysResponse{
			TargetUserID:             req.TargetUserID,
			TargetYear:               req.TargetYear,
			TargetMonth:              req.TargetMonth,
			MonthlyAttendanceRequest: monthlyAttendanceRequestResponse,
			AttendanceDays:           attendanceDayResponses,
		},
		"SEARCH_ATTENDANCE_DAYS_SUCCESS",
		"勤怠一覧を取得しました",
		nil,
	)
}

/*
 * 勤怠更新
 *
 * APIとして直接公開しない。
 * monthly_attendances/update の月次全体保存から内部的に使う。
 *
 * 仕様：
 * ・targetUserId + workDate で既存勤怠を検索する
 * ・存在しなければ新規作成する
 * ・存在すれば更新する
 * ・休日は予定・実績ともに時間を保存しない
 * ・syncPlanActual = true の勤務区分は、commonStartAt / commonEndAt を plan / actual の両方へ反映する
 * ・通常勤務は actualAttendanceTypeId をフロントに要求せず、予定区分IDと同じ値を実績区分IDとして保存する
 * ・欠勤、病欠、遅刻、早退は actualAttendanceTypeId ではなく各Flagで表現する
 * ・夜勤は勤務区分ではなく、actualStartAt / actualEndAt から集計時に深夜時間として計算する
 * ・管理者側では MonthlyAttendanceRequest による編集ロックを行わない
 */
func (service *attendanceDayService) UpdateAttendanceDay(
	req types.UpdateAttendanceDayRequest,
) results.Result {
	if req.TargetUserID == 0 {
		return results.BadRequest(
			"UPDATE_ATTENDANCE_DAY_INVALID_TARGET_USER_ID",
			"対象ユーザーIDが正しくありません",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

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
	 * 実績区分IDは予定区分IDと同じ値を保存する。
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

		req.RemoteWorkAllowanceFlag = false

		req.TransportFrom = nil
		req.TransportTo = nil
		req.TransportMethod = nil
		req.TransportAmount = nil
	} else if attendanceType.SyncPlanActual {
		/*
		 * 予定・実績同期対象の場合
		 *
		 * 有給、特別休暇、休職、介護休業、育児休業などは、
		 * commonStartAt / commonEndAt を plan / actual の両方へ反映する。
		 *
		 * 欠勤、病欠、遅刻、早退は勤務区分ではなく実績状態なので、
		 * ここには含めない。
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
		 * 通常勤務の場合
		 *
		 * 実績区分は予定区分と同じ勤務区分IDを保存する。
		 * 欠勤、病欠、遅刻、早退は actualAttendanceTypeId ではなく、
		 * LateFlag / EarlyLeaveFlag / AbsenceFlag / SickLeaveFlag で表現する。
		 *
		 * 夜勤は勤務区分ではない。
		 * 深夜時間は actualStartAt / actualEndAt から集計時に計算する。
		 */
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

		actualAttendanceTypeID = req.PlanAttendanceTypeID
		planStartAt = parsedPlanStartAt
		planEndAt = parsedPlanEndAt
		actualStartAt = parsedActualStartAt
		actualEndAt = parsedActualEndAt
	}

	findQuery, buildFindResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(req.TargetUserID, workDate)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentAttendanceDay, findResult := service.attendanceDayRepository.FindAttendanceDay(findQuery)

	if findResult.Error && findResult.Code == "ATTENDANCE_DAY_NOT_FOUND" {
		attendanceDay, buildCreateResult := service.attendanceDayBuilder.BuildCreateAttendanceDayModel(
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
 * 現時点ではAPIとして直接公開しない。
 * 必要になった場合の内部用として残す。
 *
 * 注意：
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
func (service *attendanceDayService) DeleteAttendanceDay(
	req types.DeleteAttendanceDayRequest,
) results.Result {
	if req.TargetUserID == 0 {
		return results.BadRequest(
			"DELETE_ATTENDANCE_DAY_INVALID_TARGET_USER_ID",
			"対象ユーザーIDが正しくありません",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

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

	findQuery, buildFindResult := service.attendanceDayBuilder.BuildFindAttendanceDayByUserIDAndWorkDateQuery(req.TargetUserID, workDate)
	if buildFindResult.Error {
		return buildFindResult
	}

	currentAttendanceDay, findResult := service.attendanceDayRepository.FindAttendanceDay(findQuery)
	if findResult.Error {
		return findResult
	}

	deletedAttendanceDay, buildDeleteResult := service.attendanceDayBuilder.BuildDeleteAttendanceDayModel(currentAttendanceDay)
	if buildDeleteResult.Error {
		return buildDeleteResult
	}

	_, saveResult := service.attendanceDayRepository.SaveAttendanceDay(deletedAttendanceDay)
	if saveResult.Error {
		return saveResult
	}

	return results.OK(
		types.DeleteAttendanceDayResponse{
			TargetUserID: req.TargetUserID,
			WorkDate:     req.WorkDate,
		},
		"DELETE_ATTENDANCE_DAY_SUCCESS",
		"勤怠を削除しました",
		nil,
	)
}
