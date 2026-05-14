package repositories

import (
	"errors"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

type MonthlyAttendanceRequestRepository interface {
	FindMonthlyAttendanceRequest(query *gorm.DB) (models.MonthlyAttendanceRequest, results.Result)
	CreateMonthlyAttendanceRequest(monthlyAttendanceRequest models.MonthlyAttendanceRequest) (models.MonthlyAttendanceRequest, results.Result)
	SaveMonthlyAttendanceRequest(monthlyAttendanceRequest models.MonthlyAttendanceRequest) (models.MonthlyAttendanceRequest, results.Result)
}

/*
 * 管理者用月次勤怠申請Repository
 *
 * 役割：
 * ・Builderで作成されたGORMクエリを実行する
 * ・DBへのCreate / Saveを実行する
 * ・Repository内で発生したエラーはRepositoryでcode/message/detailsを作って返す
 *
 * このRepositoryで扱うもの：
 * ・月次勤怠申請1件取得
 * ・月次勤怠申請作成
 * ・月次勤怠申請保存
 *
 * このRepositoryで扱わないもの：
 * ・検索条件の作成
 * ・申請可否の判定
 * ・取り下げ可否の判定
 * ・承認、否認の業務ルール
 *
 * 注意：
 * ・検索条件や業務ルールは作らない
 * ・クエリ作成はBuilderに任せる
 * ・月次勤怠申請の状態判定はServiceに任せる
 */
type monthlyAttendanceRequestRepository struct {
	db *gorm.DB
}

/*
 * MonthlyAttendanceRequestRepository生成
 */
func NewMonthlyAttendanceRequestRepository(db *gorm.DB) MonthlyAttendanceRequestRepository {
	return &monthlyAttendanceRequestRepository{db: db}
}

/*
 * 月次勤怠申請1件取得
 *
 * Builderで作成された query を実行する。
 *
 * 主な用途：
 * ・対象年月の申請状態取得
 * ・申請時の既存レコード確認
 * ・取り下げ時の既存レコード確認
 * ・承認時の対象レコード確認
 * ・否認時の対象レコード確認
 *
 * 注意：
 * ・レコードが存在しない場合は MONTHLY_ATTENDANCE_REQUEST_NOT_FOUND を返す
 * ・未申請扱いにするかどうかはService側で判断する
 */
func (repository *monthlyAttendanceRequestRepository) FindMonthlyAttendanceRequest(
	query *gorm.DB,
) (models.MonthlyAttendanceRequest, results.Result) {
	if query == nil {
		return models.MonthlyAttendanceRequest{}, results.InternalServerError(
			"FIND_MONTHLY_ATTENDANCE_REQUEST_QUERY_IS_NIL",
			"月次勤怠申請情報の取得に失敗しました",
			nil,
		)
	}

	var monthlyAttendanceRequest models.MonthlyAttendanceRequest

	if err := query.First(&monthlyAttendanceRequest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.MonthlyAttendanceRequest{}, results.NotFound(
				"MONTHLY_ATTENDANCE_REQUEST_NOT_FOUND",
				"対象月の月次勤怠申請が見つかりません",
				nil,
			)
		}

		return models.MonthlyAttendanceRequest{}, results.InternalServerError(
			"FIND_MONTHLY_ATTENDANCE_REQUEST_FAILED",
			"月次勤怠申請情報の取得に失敗しました",
			err.Error(),
		)
	}

	return monthlyAttendanceRequest, results.OK(
		nil,
		"FIND_MONTHLY_ATTENDANCE_REQUEST_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次勤怠申請作成
 *
 * 未申請の対象月を新規申請するときに使う。
 */
func (repository *monthlyAttendanceRequestRepository) CreateMonthlyAttendanceRequest(
	monthlyAttendanceRequest models.MonthlyAttendanceRequest,
) (models.MonthlyAttendanceRequest, results.Result) {
	if err := repository.db.Create(&monthlyAttendanceRequest).Error; err != nil {
		return models.MonthlyAttendanceRequest{}, results.InternalServerError(
			"CREATE_MONTHLY_ATTENDANCE_REQUEST_FAILED",
			"月次勤怠申請の作成に失敗しました",
			err.Error(),
		)
	}

	return monthlyAttendanceRequest, results.OK(
		nil,
		"CREATE_MONTHLY_ATTENDANCE_REQUEST_SUCCESS",
		"",
		nil,
	)
}

/*
 * 月次勤怠申請保存
 *
 * 再申請・取り下げ・承認・否認で使う。
 */
func (repository *monthlyAttendanceRequestRepository) SaveMonthlyAttendanceRequest(
	monthlyAttendanceRequest models.MonthlyAttendanceRequest,
) (models.MonthlyAttendanceRequest, results.Result) {
	if monthlyAttendanceRequest.ID == 0 {
		return models.MonthlyAttendanceRequest{}, results.InternalServerError(
			"SAVE_MONTHLY_ATTENDANCE_REQUEST_EMPTY_ID",
			"月次勤怠申請情報の保存に失敗しました",
			nil,
		)
	}

	if err := repository.db.Save(&monthlyAttendanceRequest).Error; err != nil {
		return models.MonthlyAttendanceRequest{}, results.InternalServerError(
			"SAVE_MONTHLY_ATTENDANCE_REQUEST_FAILED",
			"月次勤怠申請情報の保存に失敗しました",
			err.Error(),
		)
	}

	return monthlyAttendanceRequest, results.OK(
		nil,
		"SAVE_MONTHLY_ATTENDANCE_REQUEST_SUCCESS",
		"",
		nil,
	)
}
