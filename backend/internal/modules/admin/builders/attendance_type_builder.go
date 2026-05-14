package builders

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/admin/types"
	"timexeed/backend/internal/results"

	"gorm.io/gorm"
)

/*
 * 管理者用勤務区分マスタBuilder interface
 *
 * ServiceがBuilderに求める処理だけを定義する。
 *
 * 管理者側では勤務区分マスタの作成・更新・削除はしない。
 * 勤怠編集画面で使う選択肢取得用の検索クエリだけを作成する。
 */
type AttendanceTypeBuilder interface {
	BuildSearchAttendanceTypesQuery(req types.SearchAttendanceTypesRequest) (*gorm.DB, results.Result)
}

/*
 * 管理者用勤務区分マスタBuilder
 *
 * 役割：
 * ・Serviceから受け取ったRequestをもとにGORMクエリを作成する
 * ・Builder内で発生したエラーはBuilderでcode/message/detailsを作って返す
 *
 * 注意：
 * ・DB実行はしない
 * ・FindはRepositoryに任せる
 * ・管理者側では有効な勤務区分だけ取得する
 */
type attendanceTypeBuilder struct {
	db *gorm.DB
}

/*
 * AttendanceTypeBuilder生成
 */
func NewAttendanceTypeBuilder(db *gorm.DB) AttendanceTypeBuilder {
	return &attendanceTypeBuilder{db: db}
}

/*
 * 勤務区分マスタ検索用クエリ作成
 *
 * 管理者側の勤怠編集画面で使うため、
 * ・論理削除されていない
 * ・有効状態
 * の勤務区分だけを取得する。
 *
 * ページネーションはしない。
 * キーワード検索もしない。
 * 勤務区分は件数が少なく、フロントではプルダウン表示が基本のため、
 * 全件取得する。
 */
func (builder *attendanceTypeBuilder) BuildSearchAttendanceTypesQuery(req types.SearchAttendanceTypesRequest) (*gorm.DB, results.Result) {
	query := builder.db.
		Model(&models.AttendanceType{}).
		Where("is_deleted = ?", false).
		Where("is_active = ?", true).
		Order("display_order ASC").
		Order("id ASC")

	return query, results.OK(
		nil,
		"BUILD_SEARCH_ATTENDANCE_TYPES_QUERY_SUCCESS",
		"",
		nil,
	)
}
