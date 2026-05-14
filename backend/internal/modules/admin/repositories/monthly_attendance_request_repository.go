package repositories

import (
	"errors"
	"time"

	"timexeed/backend/internal/models"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者 月次勤怠申請一覧検索用Row
 *
 * 注意：
 * ・users 起点で LEFT JOIN した検索結果を受け取るためのRepository用Row
 * ・未申請は monthly_attendance_requests.id が NULL
 * ・フロント返却用Responseへの変換はServiceで行う
 */
type MonthlyAttendanceRequestSearchRow struct {
	TargetUserID uint `gorm:"column:target_user_id"`

	UserName string `gorm:"column:user_name"`
	Email    string `gorm:"column:email"`

	DepartmentID   *uint   `gorm:"column:department_id"`
	DepartmentName *string `gorm:"column:department_name"`

	MonthlyAttendanceRequestID *uint `gorm:"column:monthly_attendance_request_id"`

	RequestTargetYear  *int `gorm:"column:request_target_year"`
	RequestTargetMonth *int `gorm:"column:request_target_month"`

	Status *string `gorm:"column:status"`

	RequestMemo *string    `gorm:"column:request_memo"`
	RequestedAt *time.Time `gorm:"column:requested_at"`

	ApprovedBy *uint      `gorm:"column:approved_by"`
	ApprovedAt *time.Time `gorm:"column:approved_at"`

	RejectedReason *string    `gorm:"column:rejected_reason"`
	RejectedAt     *time.Time `gorm:"column:rejected_at"`

	CanceledReason *string    `gorm:"column:canceled_reason"`
	CanceledAt     *time.Time `gorm:"column:canceled_at"`

	CreatedAt *time.Time `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

type MonthlyAttendanceRequestRepository interface {
	SearchMonthlyAttendanceRequests(
		query *gorm.DB,
		limit int,
	) ([]MonthlyAttendanceRequestSearchRow, bool, results.Result)

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
 * ・月次勤怠申請一覧検索
 * ・月次勤怠申請1件取得
 * ・月次勤怠申請作成
 * ・月次勤怠申請保存
 *
 * このRepositoryで扱わないもの：
 * ・検索条件の作成
 * ・申請可否の判定
 * ・取り下げ可否の判定
 * ・承認、否認の業務ルール
 * ・Responseへの変換
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
 * 月次勤怠申請一覧検索
 *
 * Builderで作成された users 起点の LEFT JOIN query を実行する。
 *
 * 仕様：
 * ・limit + 1 件取得された場合は hasMore=true として返す
 * ・返却するrowsは limit 件に切り詰める
 *
 * 注意：
 * ・未申請は monthly_attendance_requests.id が NULL のRowとして返る
 * ・NOT_SUBMITTEDへの変換はService側で行う
 */
func (repository *monthlyAttendanceRequestRepository) SearchMonthlyAttendanceRequests(
	query *gorm.DB,
	limit int,
) ([]MonthlyAttendanceRequestSearchRow, bool, results.Result) {
	if query == nil {
		return nil, false, results.InternalServerError(
			"SEARCH_MONTHLY_ATTENDANCE_REQUESTS_QUERY_IS_NIL",
			"月次勤怠申請一覧の取得に失敗しました",
			nil,
		)
	}

	if limit <= 0 {
		return nil, false, results.InternalServerError(
			"SEARCH_MONTHLY_ATTENDANCE_REQUESTS_INVALID_LIMIT",
			"月次勤怠申請一覧の取得に失敗しました",
			map[string]any{
				"limit": limit,
			},
		)
	}

	var rows []MonthlyAttendanceRequestSearchRow

	if err := query.Scan(&rows).Error; err != nil {
		return nil, false, results.InternalServerError(
			"SEARCH_MONTHLY_ATTENDANCE_REQUESTS_FAILED",
			"月次勤怠申請一覧の取得に失敗しました",
			err.Error(),
		)
	}

	hasMore := false

	if len(rows) > limit {
		hasMore = true
		rows = rows[:limit]
	}

	return rows, hasMore, results.OK(
		nil,
		"SEARCH_MONTHLY_ATTENDANCE_REQUESTS_SUCCESS",
		"",
		nil,
	)
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
