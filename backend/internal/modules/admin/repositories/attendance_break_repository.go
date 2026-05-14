package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type AttendanceBreakRepository interface {
	FindAttendanceBreaks(query *gorm.DB) ([]models.AttendanceBreak, results.Result)
	FindAttendanceBreak(query *gorm.DB) (models.AttendanceBreak, results.Result)
	CreateAttendanceBreak(attendanceBreak models.AttendanceBreak) (models.AttendanceBreak, results.Result)
	SaveAttendanceBreak(attendanceBreak models.AttendanceBreak) (models.AttendanceBreak, results.Result)
}

/*
 * 管理者用休憩Repository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・DBへのCreate / Saveを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 * ・休憩の月次申請状態チェックなどはServiceに任せる
 * ・AttendanceBreak は申請状態を持たない
 * ・管理者側では月次申請状態による編集ロックを行わない
 *
 * 月次全体保存での休憩保存方針：
 * ・IDなしの休憩は CreateAttendanceBreak で作成する
 * ・IDありの休憩は SaveAttendanceBreak で更新する
 * ・リクエストから消えた休憩は SaveAttendanceBreak で論理削除する
 */
type attendanceBreakRepository struct {
	db *gorm.DB
}

/*
 * AttendanceBreakRepository生成
 */
func NewAttendanceBreakRepository(db *gorm.DB) AttendanceBreakRepository {
	return &attendanceBreakRepository{db: db}
}

/*
 * 休憩一覧取得
 */
func (repository *attendanceBreakRepository) FindAttendanceBreaks(query *gorm.DB) ([]models.AttendanceBreak, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_ATTENDANCE_BREAKS_QUERY_IS_NIL",
			"休憩一覧の取得に失敗しました",
			nil,
		)
	}

	var attendanceBreaks []models.AttendanceBreak

	if err := query.Find(&attendanceBreaks).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_ATTENDANCE_BREAKS_FAILED",
			"休憩一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return attendanceBreaks, results.OK(
		nil,
		"FIND_ATTENDANCE_BREAKS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 休憩1件取得
 */
func (repository *attendanceBreakRepository) FindAttendanceBreak(query *gorm.DB) (models.AttendanceBreak, results.Result) {
	if query == nil {
		return models.AttendanceBreak{}, results.InternalServerError(
			"FIND_ATTENDANCE_BREAK_QUERY_IS_NIL",
			"休憩情報の取得に失敗しました",
			nil,
		)
	}

	var attendanceBreak models.AttendanceBreak

	if err := query.First(&attendanceBreak).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AttendanceBreak{}, results.NotFound(
				"ATTENDANCE_BREAK_NOT_FOUND",
				"対象の休憩が見つかりません",
				nil,
			)
		}

		return models.AttendanceBreak{}, results.InternalServerError(
			"FIND_ATTENDANCE_BREAK_FAILED",
			"休憩情報の取得に失敗しました",
			err.Error(),
		)
	}

	return attendanceBreak, results.OK(
		nil,
		"FIND_ATTENDANCE_BREAK_SUCCESS",
		"",
		nil,
	)
}

/*
 * 休憩作成
 */
func (repository *attendanceBreakRepository) CreateAttendanceBreak(attendanceBreak models.AttendanceBreak) (models.AttendanceBreak, results.Result) {
	if err := repository.db.Create(&attendanceBreak).Error; err != nil {
		return models.AttendanceBreak{}, results.InternalServerError(
			"CREATE_ATTENDANCE_BREAK_FAILED",
			"休憩の作成に失敗しました",
			err.Error(),
		)
	}

	return attendanceBreak, results.OK(
		nil,
		"CREATE_ATTENDANCE_BREAK_SUCCESS",
		"",
		nil,
	)
}

/*
 * 休憩保存
 *
 * 更新・論理削除で使う。
 */
func (repository *attendanceBreakRepository) SaveAttendanceBreak(attendanceBreak models.AttendanceBreak) (models.AttendanceBreak, results.Result) {
	if attendanceBreak.ID == 0 {
		return models.AttendanceBreak{}, results.InternalServerError(
			"SAVE_ATTENDANCE_BREAK_EMPTY_ID",
			"休憩情報の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&attendanceBreak).Error; err != nil {
		return models.AttendanceBreak{}, results.InternalServerError(
			"SAVE_ATTENDANCE_BREAK_FAILED",
			"休憩情報の保存に失敗しました",
			err.Error(),
		)
	}

	return attendanceBreak, results.OK(
		nil,
		"SAVE_ATTENDANCE_BREAK_SUCCESS",
		"",
		nil,
	)
}
