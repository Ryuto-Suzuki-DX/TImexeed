package services

import (
	"timexeed/backend/internal/models"
	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/types"
	"timexeed/backend/internal/results"
)

/*
 * 従業員用勤務区分マスタService interface
 *
 * ControllerがServiceに求める処理だけを定義する。
 *
 * ユーザー側では勤務区分マスタの作成・更新・削除はしない。
 * 勤怠入力画面で使う選択肢を取得するだけ。
 */
type AttendanceTypeService interface {
	SearchAttendanceTypes(req types.SearchAttendanceTypesRequest) results.Result
}

/*
 * 従業員用勤務区分マスタService
 *
 * 役割：
 * ・Controllerから受け取ったRequestをもとに処理を進める
 * ・Serviceで発生したエラーはServiceでcode/message/detailsを作る
 * ・Builderで検索クエリを作成する
 * ・Builderで発生したエラーはBuilderから返されたResultをそのまま返す
 * ・RepositoryでDB検索を実行する
 * ・Repositoryで発生したエラーはRepositoryから返されたResultをそのまま返す
 * ・成功時はResponse型に変換してControllerへ返す
 *
 * 注意：
 * ・Controllerにはgin.Contextを渡さない
 * ・Serviceではc.JSONしない
 * ・DBへの直接アクセスはRepositoryに任せる
 * ・Builder/Repositoryのエラー文言をServiceで作り直さない
 */
type attendanceTypeService struct {
	attendanceTypeBuilder    builders.AttendanceTypeBuilder
	attendanceTypeRepository repositories.AttendanceTypeRepository
}

/*
 * AttendanceTypeService生成
 */
func NewAttendanceTypeService(
	attendanceTypeBuilder builders.AttendanceTypeBuilder,
	attendanceTypeRepository repositories.AttendanceTypeRepository,
) *attendanceTypeService {
	return &attendanceTypeService{
		attendanceTypeBuilder:    attendanceTypeBuilder,
		attendanceTypeRepository: attendanceTypeRepository,
	}
}

/*
 * models.AttendanceTypeをフロント返却用AttendanceTypeResponseへ変換する
 *
 * フロントはこのレスポンスを見て、
 * ・予定と実績を分けるか
 * ・共通時間入力にするか
 * ・休憩入力を出すか
 * ・交通費入力を出すか
 * ・遅刻、早退、欠勤、病欠の入力を出すか
 * を判断する。
 */
func toAttendanceTypeResponse(attendanceType models.AttendanceType) types.AttendanceTypeResponse {
	return types.AttendanceTypeResponse{
		ID:       attendanceType.ID,
		Code:     attendanceType.Code,
		Name:     attendanceType.Name,
		Category: attendanceType.Category,

		SyncPlanActual: attendanceType.SyncPlanActual,

		AllowActualTimeInput: attendanceType.AllowActualTimeInput,
		AllowBreakInput:      attendanceType.AllowBreakInput,
		AllowTransportInput:  attendanceType.AllowTransportInput,

		AllowLateFlag:       attendanceType.AllowLateFlag,
		AllowEarlyLeaveFlag: attendanceType.AllowEarlyLeaveFlag,
		AllowAbsenceFlag:    attendanceType.AllowAbsenceFlag,
		AllowSickLeaveFlag:  attendanceType.AllowSickLeaveFlag,

		RequiresRequest: attendanceType.RequiresRequest,

		DisplayOrder: attendanceType.DisplayOrder,
	}
}

/*
 * 勤務区分マスタ検索
 *
 * ユーザー側では、
 * ・is_deleted = false
 * ・is_active = true
 * の勤務区分だけ取得する。
 *
 * 用途：
 * ・勤怠入力画面の区分選択肢
 * ・フロント側の入力欄出し分け
 */
func (service *attendanceTypeService) SearchAttendanceTypes(req types.SearchAttendanceTypesRequest) results.Result {
	// Builderで勤務区分マスタ検索用クエリを作成する
	query, buildResult := service.attendanceTypeBuilder.BuildSearchAttendanceTypesQuery(req)
	if buildResult.Error {
		return buildResult
	}

	// Repositoryで勤務区分マスタ一覧を取得する
	attendanceTypes, findResult := service.attendanceTypeRepository.FindAttendanceTypes(query)
	if findResult.Error {
		return findResult
	}

	// DBモデルをフロント返却用Responseへ変換する
	attendanceTypeResponses := make([]types.AttendanceTypeResponse, 0, len(attendanceTypes))
	for _, attendanceType := range attendanceTypes {
		attendanceTypeResponses = append(attendanceTypeResponses, toAttendanceTypeResponse(attendanceType))
	}

	return results.OK(
		types.SearchAttendanceTypesResponse{
			AttendanceTypes: attendanceTypeResponses,
		},
		"SEARCH_ATTENDANCE_TYPES_SUCCESS",
		"勤務区分マスタを取得しました",
		nil,
	)
}
