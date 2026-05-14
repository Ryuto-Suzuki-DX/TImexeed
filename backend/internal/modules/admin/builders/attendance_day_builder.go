package builders

import (
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type AttendanceDayBuilder interface {
	BuildSearchAttendanceDaysQuery(req types.SearchAttendanceDaysRequest) (*gorm.DB, results.Result)
	BuildFindAttendanceDayByUserIDAndWorkDateQuery(targetUserID uint, workDate time.Time) (*gorm.DB, results.Result)
	BuildCreateAttendanceDayModel(
		req types.UpdateAttendanceDayRequest,
		workDate time.Time,
		planStartAt *time.Time,
		planEndAt *time.Time,
		actualStartAt *time.Time,
		actualEndAt *time.Time,
		actualAttendanceTypeID uint,
	) (models.AttendanceDay, results.Result)
	BuildUpdateAttendanceDayModel(
		currentAttendanceDay models.AttendanceDay,
		req types.UpdateAttendanceDayRequest,
		workDate time.Time,
		planStartAt *time.Time,
		planEndAt *time.Time,
		actualStartAt *time.Time,
		actualEndAt *time.Time,
		actualAttendanceTypeID uint,
	) (models.AttendanceDay, results.Result)
	BuildDeleteAttendanceDayModel(currentAttendanceDay models.AttendanceDay) (models.AttendanceDay, results.Result)
}

/*
 * 管理者用勤怠Builder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Serviceから受け取ったRequestをもとにDB保存用Modelを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・Find / Create / Save はRepositoryに任せる
 * ・日付文字列、日時文字列の変換はServiceで行う
 * ・Builderでは変換済みの time.Time / *time.Time を受け取る
 * ・AttendanceDay は申請状態を持たない
 * ・AttendanceDay は画面表示用SystemMessageを持たない
 * ・月次申請状態は MonthlyAttendanceRequest 側で管理する
 * ・管理者側では対象ユーザーIDを request body の targetUserId で受け取る
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
type attendanceDayBuilder struct {
	db *gorm.DB
}

/*
 * AttendanceDayBuilder生成
 */
func NewAttendanceDayBuilder(db *gorm.DB) AttendanceDayBuilder {
	return &attendanceDayBuilder{db: db}
}

/*
 * 勤怠検索用クエリ作成
 *
 * 対象年月の対象ユーザーの勤怠を取得する。
 *
 * 注意：
 * ・targetUserId は管理者が選択した対象ユーザーID
 * ・論理削除済みの勤怠は対象外
 */
func (builder *attendanceDayBuilder) BuildSearchAttendanceDaysQuery(
	req types.SearchAttendanceDaysRequest,
) (*gorm.DB, results.Result) {
	if req.TargetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_ATTENDANCE_DAYS_QUERY_INVALID_TARGET_USER_ID",
			"勤怠検索条件の作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

	if req.TargetYear <= 0 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_ATTENDANCE_DAYS_QUERY_INVALID_TARGET_YEAR",
			"勤怠検索条件の作成に失敗しました",
			map[string]any{
				"targetYear": req.TargetYear,
			},
		)
	}

	if req.TargetMonth < 1 || req.TargetMonth > 12 {
		return nil, results.BadRequest(
			"BUILD_SEARCH_ATTENDANCE_DAYS_QUERY_INVALID_TARGET_MONTH",
			"勤怠検索条件の作成に失敗しました",
			map[string]any{
				"targetMonth": req.TargetMonth,
			},
		)
	}

	startDate := time.Date(req.TargetYear, time.Month(req.TargetMonth), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	query := builder.db.
		Model(&models.AttendanceDay{}).
		Preload("PlanAttendanceType").
		Preload("ActualAttendanceType").
		Where("user_id = ?", req.TargetUserID).
		Where("work_date >= ?", startDate).
		Where("work_date < ?", endDate).
		Where("is_deleted = ?", false).
		Order("work_date ASC").
		Order("id ASC")

	return query, results.OK(
		nil,
		"BUILD_SEARCH_ATTENDANCE_DAYS_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * ユーザーID + 対象日で勤怠1件取得用クエリ作成
 *
 * 更新・削除時に使う。
 *
 * 注意：
 * ・targetUserID は管理者が選択した対象ユーザーID
 * ・workDate はServiceで utils.ParseDate 済みの time.Time
 * ・論理削除済みの勤怠は対象外
 */
func (builder *attendanceDayBuilder) BuildFindAttendanceDayByUserIDAndWorkDateQuery(
	targetUserID uint,
	workDate time.Time,
) (*gorm.DB, results.Result) {
	if targetUserID == 0 {
		return nil, results.BadRequest(
			"BUILD_FIND_ATTENDANCE_DAY_QUERY_INVALID_TARGET_USER_ID",
			"勤怠取得条件の作成に失敗しました",
			map[string]any{
				"targetUserId": targetUserID,
			},
		)
	}

	if workDate.IsZero() {
		return nil, results.BadRequest(
			"BUILD_FIND_ATTENDANCE_DAY_QUERY_EMPTY_WORK_DATE",
			"勤怠取得条件の作成に失敗しました",
			nil,
		)
	}

	query := builder.db.
		Model(&models.AttendanceDay{}).
		Preload("PlanAttendanceType").
		Preload("ActualAttendanceType").
		Where("user_id = ?", targetUserID).
		Where("work_date = ?", workDate).
		Where("is_deleted = ?", false)

	return query, results.OK(
		nil,
		"BUILD_FIND_ATTENDANCE_DAY_QUERY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠作成用Model作成
 *
 * 画面上は「更新」操作だが、対象日の勤怠が未登録の場合は新規作成する。
 *
 * 注意：
 * ・commonStartAt / commonEndAt はModelに持たせない
 * ・Service側で syncPlanActual を見て plan / actual へ変換済みの値を受け取る
 * ・AttendanceDay には申請状態を保存しない
 * ・AttendanceDay には画面表示用SystemMessageを保存しない
 */
func (builder *attendanceDayBuilder) BuildCreateAttendanceDayModel(
	req types.UpdateAttendanceDayRequest,
	workDate time.Time,
	planStartAt *time.Time,
	planEndAt *time.Time,
	actualStartAt *time.Time,
	actualEndAt *time.Time,
	actualAttendanceTypeID uint,
) (models.AttendanceDay, results.Result) {
	if req.TargetUserID == 0 {
		return models.AttendanceDay{}, results.BadRequest(
			"BUILD_CREATE_ATTENDANCE_DAY_MODEL_INVALID_TARGET_USER_ID",
			"勤怠作成データの作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

	if workDate.IsZero() {
		return models.AttendanceDay{}, results.BadRequest(
			"BUILD_CREATE_ATTENDANCE_DAY_MODEL_EMPTY_WORK_DATE",
			"勤怠作成データの作成に失敗しました",
			nil,
		)
	}

	if req.PlanAttendanceTypeID == 0 {
		return models.AttendanceDay{}, results.BadRequest(
			"BUILD_CREATE_ATTENDANCE_DAY_MODEL_EMPTY_PLAN_ATTENDANCE_TYPE_ID",
			"勤怠作成データの作成に失敗しました",
			nil,
		)
	}

	if actualAttendanceTypeID == 0 {
		return models.AttendanceDay{}, results.BadRequest(
			"BUILD_CREATE_ATTENDANCE_DAY_MODEL_EMPTY_ACTUAL_ATTENDANCE_TYPE_ID",
			"勤怠作成データの作成に失敗しました",
			nil,
		)
	}

	attendanceDay := models.AttendanceDay{
		UserID:                  req.TargetUserID,
		WorkDate:                workDate,
		PlanAttendanceTypeID:    req.PlanAttendanceTypeID,
		ActualAttendanceTypeID:  actualAttendanceTypeID,
		PlanStartAt:             planStartAt,
		PlanEndAt:               planEndAt,
		ActualStartAt:           actualStartAt,
		ActualEndAt:             actualEndAt,
		LateFlag:                req.LateFlag,
		EarlyLeaveFlag:          req.EarlyLeaveFlag,
		AbsenceFlag:             req.AbsenceFlag,
		SickLeaveFlag:           req.SickLeaveFlag,
		RemoteWorkAllowanceFlag: req.RemoteWorkAllowanceFlag,
		TransportFrom:           req.TransportFrom,
		TransportTo:             req.TransportTo,
		TransportMethod:         req.TransportMethod,
		TransportAmount:         req.TransportAmount,
		IsDeleted:               false,
	}

	return attendanceDay, results.OK(
		nil,
		"BUILD_CREATE_ATTENDANCE_DAY_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠更新用Model作成
 *
 * 対象日の勤怠が登録済みの場合に更新する。
 *
 * 注意：
 * ・commonStartAt / commonEndAt はModelに持たせない
 * ・Service側で syncPlanActual を見て plan / actual へ変換済みの値を受け取る
 * ・AttendanceDay には申請状態を保存しない
 * ・AttendanceDay には画面表示用SystemMessageを保存しない
 */
func (builder *attendanceDayBuilder) BuildUpdateAttendanceDayModel(
	currentAttendanceDay models.AttendanceDay,
	req types.UpdateAttendanceDayRequest,
	workDate time.Time,
	planStartAt *time.Time,
	planEndAt *time.Time,
	actualStartAt *time.Time,
	actualEndAt *time.Time,
	actualAttendanceTypeID uint,
) (models.AttendanceDay, results.Result) {
	if currentAttendanceDay.ID == 0 {
		return models.AttendanceDay{}, results.BadRequest(
			"BUILD_UPDATE_ATTENDANCE_DAY_MODEL_EMPTY_CURRENT_ATTENDANCE_DAY",
			"勤怠更新データの作成に失敗しました",
			nil,
		)
	}

	if req.TargetUserID == 0 {
		return models.AttendanceDay{}, results.BadRequest(
			"BUILD_UPDATE_ATTENDANCE_DAY_MODEL_INVALID_TARGET_USER_ID",
			"勤怠更新データの作成に失敗しました",
			map[string]any{
				"targetUserId": req.TargetUserID,
			},
		)
	}

	if currentAttendanceDay.UserID != req.TargetUserID {
		return models.AttendanceDay{}, results.Conflict(
			"BUILD_UPDATE_ATTENDANCE_DAY_MODEL_USER_ID_MISMATCH",
			"勤怠更新対象のユーザーが一致しません",
			map[string]any{
				"currentUserId": currentAttendanceDay.UserID,
				"targetUserId":  req.TargetUserID,
			},
		)
	}

	if workDate.IsZero() {
		return models.AttendanceDay{}, results.BadRequest(
			"BUILD_UPDATE_ATTENDANCE_DAY_MODEL_EMPTY_WORK_DATE",
			"勤怠更新データの作成に失敗しました",
			nil,
		)
	}

	if req.PlanAttendanceTypeID == 0 {
		return models.AttendanceDay{}, results.BadRequest(
			"BUILD_UPDATE_ATTENDANCE_DAY_MODEL_EMPTY_PLAN_ATTENDANCE_TYPE_ID",
			"勤怠更新データの作成に失敗しました",
			nil,
		)
	}

	if actualAttendanceTypeID == 0 {
		return models.AttendanceDay{}, results.BadRequest(
			"BUILD_UPDATE_ATTENDANCE_DAY_MODEL_EMPTY_ACTUAL_ATTENDANCE_TYPE_ID",
			"勤怠更新データの作成に失敗しました",
			nil,
		)
	}

	currentAttendanceDay.WorkDate = workDate
	currentAttendanceDay.PlanAttendanceTypeID = req.PlanAttendanceTypeID
	currentAttendanceDay.ActualAttendanceTypeID = actualAttendanceTypeID
	currentAttendanceDay.PlanStartAt = planStartAt
	currentAttendanceDay.PlanEndAt = planEndAt
	currentAttendanceDay.ActualStartAt = actualStartAt
	currentAttendanceDay.ActualEndAt = actualEndAt
	currentAttendanceDay.LateFlag = req.LateFlag
	currentAttendanceDay.EarlyLeaveFlag = req.EarlyLeaveFlag
	currentAttendanceDay.AbsenceFlag = req.AbsenceFlag
	currentAttendanceDay.SickLeaveFlag = req.SickLeaveFlag
	currentAttendanceDay.RemoteWorkAllowanceFlag = req.RemoteWorkAllowanceFlag
	currentAttendanceDay.TransportFrom = req.TransportFrom
	currentAttendanceDay.TransportTo = req.TransportTo
	currentAttendanceDay.TransportMethod = req.TransportMethod
	currentAttendanceDay.TransportAmount = req.TransportAmount

	return currentAttendanceDay, results.OK(
		nil,
		"BUILD_UPDATE_ATTENDANCE_DAY_MODEL_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠論理削除用Model作成
 *
 * 現時点ではAPIとして直接公開しない。
 * 必要になった場合の内部用として残す。
 */
func (builder *attendanceDayBuilder) BuildDeleteAttendanceDayModel(
	currentAttendanceDay models.AttendanceDay,
) (models.AttendanceDay, results.Result) {
	if currentAttendanceDay.ID == 0 {
		return models.AttendanceDay{}, results.BadRequest(
			"BUILD_DELETE_ATTENDANCE_DAY_MODEL_EMPTY_CURRENT_ATTENDANCE_DAY",
			"勤怠削除データの作成に失敗しました",
			nil,
		)
	}

	now := time.Now()

	currentAttendanceDay.IsDeleted = true
	currentAttendanceDay.DeletedAt = &now

	return currentAttendanceDay, results.OK(
		nil,
		"BUILD_DELETE_ATTENDANCE_DAY_MODEL_SUCCESS",
		"",
		nil,
	)
}
