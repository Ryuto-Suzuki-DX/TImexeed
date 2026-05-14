package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type AttendanceTypeRepository interface {
	FindAttendanceTypes(query *gorm.DB) ([]models.AttendanceType, results.Result)
	FindAttendanceTypeByID(attendanceTypeID uint) (models.AttendanceType, results.Result)
}

/*
 * 管理者用勤務区分マスタRepository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・勤務区分マスタの1件取得を行う
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 * ・管理者側では作成 / 更新 / 削除は行わない
 */
type attendanceTypeRepository struct {
	db *gorm.DB
}

/*
 * AttendanceTypeRepository生成
 */
func NewAttendanceTypeRepository(db *gorm.DB) AttendanceTypeRepository {
	return &attendanceTypeRepository{db: db}
}

/*
 * 勤務区分マスタ一覧取得
 *
 * 勤怠編集画面のプルダウン表示に使う。
 */
func (repository *attendanceTypeRepository) FindAttendanceTypes(query *gorm.DB) ([]models.AttendanceType, results.Result) {
	if query == nil {
		return nil, results.InternalServerError(
			"FIND_ATTENDANCE_TYPES_QUERY_IS_NIL",
			"勤務区分マスタの取得に失敗しました",
			nil,
		)
	}

	var attendanceTypes []models.AttendanceType

	if err := query.Find(&attendanceTypes).Error; err != nil {
		return nil, results.InternalServerError(
			"FIND_ATTENDANCE_TYPES_FAILED",
			"勤務区分マスタの取得に失敗しました",
			err.Error(),
		)
	}

	return attendanceTypes, results.OK(
		nil,
		"FIND_ATTENDANCE_TYPES_SUCCESS",
		"",
		nil,
	)
}

/*
 * 勤務区分マスタ1件取得
 *
 * 勤怠更新時に、選択された勤務区分の設定を確認するために使う。
 *
 * 例：
 * ・syncPlanActual = true なら commonStartAt / commonEndAt を plan / actual に反映する
 * ・syncPlanActual = false なら plan / actual を別々に扱う
 */
func (repository *attendanceTypeRepository) FindAttendanceTypeByID(attendanceTypeID uint) (models.AttendanceType, results.Result) {
	if attendanceTypeID == 0 {
		return models.AttendanceType{}, results.BadRequest(
			"FIND_ATTENDANCE_TYPE_INVALID_ATTENDANCE_TYPE_ID",
			"勤務区分マスタの取得に失敗しました",
			map[string]any{
				"attendanceTypeId": attendanceTypeID,
			},
		)
	}

	var attendanceType models.AttendanceType

	if err := repository.db.
		Where("id = ?", attendanceTypeID).
		Where("is_deleted = ?", false).
		Where("is_active = ?", true).
		First(&attendanceType).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AttendanceType{}, results.NotFound(
				"ATTENDANCE_TYPE_NOT_FOUND",
				"勤務区分マスタが見つかりません",
				map[string]any{
					"attendanceTypeId": attendanceTypeID,
				},
			)
		}

		return models.AttendanceType{}, results.InternalServerError(
			"FIND_ATTENDANCE_TYPE_FAILED",
			"勤務区分マスタの取得に失敗しました",
			err.Error(),
		)
	}

	return attendanceType, results.OK(
		nil,
		"FIND_ATTENDANCE_TYPE_SUCCESS",
		"",
		nil,
	)
}
