package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type AttendanceDayRepository interface {
	FindAttendanceDays(query *gorm.DB) ([]models.AttendanceDay, results.Result)
	FindAttendanceDay(query *gorm.DB) (models.AttendanceDay, results.Result)
	CreateAttendanceDay(attendanceDay models.AttendanceDay) (models.AttendanceDay, results.Result)
	SaveAttendanceDay(attendanceDay models.AttendanceDay) (models.AttendanceDay, results.Result)
}

/*
 * 管理者用勤怠Repository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・DBへのCreate / Saveを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 * ・勤怠の月次申請状態チェックなどはServiceに任せる
 * ・AttendanceDay は申請状態を持たない
 * ・AttendanceDay は画面表示用メッセージを持たない
 * ・管理者側では月次申請状態による編集ロックを行わない
 */
type attendanceDayRepository struct {
	db *gorm.DB
}

/*
 * AttendanceDayRepository生成
 */
func NewAttendanceDayRepository(db *gorm.DB) AttendanceDayRepository {
	return &attendanceDayRepository{db: db}
}

/*
 * 勤怠一覧取得
 */
func (repository *attendanceDayRepository) FindAttendanceDays(query *gorm.DB) ([]models.AttendanceDay, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_ATTENDANCE_DAYS_QUERY_IS_NIL",
			"勤怠一覧の取得に失敗しました",
			nil,
		)
	}

	var attendanceDays []models.AttendanceDay

	if err := query.Find(&attendanceDays).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_ATTENDANCE_DAYS_FAILED",
			"勤怠一覧の取得に失敗しました",
			err.Error(),
		)
	}

	return attendanceDays, results.OK(
		nil,
		"FIND_ATTENDANCE_DAYS_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠1件取得
 */
func (repository *attendanceDayRepository) FindAttendanceDay(query *gorm.DB) (models.AttendanceDay, results.Result) {
	if query == nil {
		return models.AttendanceDay{}, results.InternalServerError(
			"FIND_ATTENDANCE_DAY_QUERY_IS_NIL",
			"勤怠情報の取得に失敗しました",
			nil,
		)
	}

	var attendanceDay models.AttendanceDay

	if err := query.First(&attendanceDay).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AttendanceDay{}, results.NotFound(
				"ATTENDANCE_DAY_NOT_FOUND",
				"対象日の勤怠が見つかりません",
				nil,
			)
		}

		return models.AttendanceDay{}, results.InternalServerError(
			"FIND_ATTENDANCE_DAY_FAILED",
			"勤怠情報の取得に失敗しました",
			err.Error(),
		)
	}

	return attendanceDay, results.OK(
		nil,
		"FIND_ATTENDANCE_DAY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠作成
 */
func (repository *attendanceDayRepository) CreateAttendanceDay(attendanceDay models.AttendanceDay) (models.AttendanceDay, results.Result) {
	if err := repository.db.Create(&attendanceDay).Error; err != nil {
		return models.AttendanceDay{}, results.InternalServerError(
			"CREATE_ATTENDANCE_DAY_FAILED",
			"勤怠の作成に失敗しました",
			err.Error(),
		)
	}

	return attendanceDay, results.OK(
		nil,
		"CREATE_ATTENDANCE_DAY_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤怠保存
 *
 * 更新・論理削除で使う。
 */
func (repository *attendanceDayRepository) SaveAttendanceDay(attendanceDay models.AttendanceDay) (models.AttendanceDay, results.Result) {
	if attendanceDay.ID == 0 {
		return models.AttendanceDay{}, results.InternalServerError(
			"SAVE_ATTENDANCE_DAY_EMPTY_ID",
			"勤怠情報の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&attendanceDay).Error; err != nil {
		return models.AttendanceDay{}, results.InternalServerError(
			"SAVE_ATTENDANCE_DAY_FAILED",
			"勤怠情報の保存に失敗しました",
			err.Error(),
		)
	}

	return attendanceDay, results.OK(
		nil,
		"SAVE_ATTENDANCE_DAY_SUCCESS",
		"",
		nil,
	)
}
