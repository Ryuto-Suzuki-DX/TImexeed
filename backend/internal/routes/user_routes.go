package routes

import (
	"timexeed/backend/internal/middlewares"

	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/controllers"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/*
 * 〇 従業員用APIルート登録
 *
 * 一般ユーザー、つまり従業員だけが使うAPIをここにまとめる。
 *
 * 重要：
 * 	管理者APIと従業員APIは完全に分離する。
 * 	管理者画面から使うAPIはここには書かない。
 *
 * ルール：
 * 	URLにIDを載せない。
 * 	userId は request body にも載せない。
 *
 * 従業員APIで扱うID：
 * 	ログイン中ユーザー本人のID
 * 		→ JWTから取得する
 *
 * 	attendanceId などの対象データID
 * 		→ request body で受け取る
 *
 * 注意：
 * 	従業員APIでは、必ず
 * 	「JWTから取得したログイン中ユーザーID」
 * 	と
 * 	「対象データの所有者 user_id」
 * 	が一致することを確認する。
 */
func RegisterUserRoutes(r *gin.Engine, db *gorm.DB) {

	// 勤怠区分マスタ
	attendanceTypeBuilder := builders.NewAttendanceTypeBuilder(db)
	attendanceTypeRepository := repositories.NewAttendanceTypeRepository(db)
	attendanceTypeService := services.NewAttendanceTypeService(attendanceTypeBuilder, attendanceTypeRepository)
	attendanceTypeController := controllers.NewAttendanceTypeController(attendanceTypeService)

	// 勤怠
	attendanceDayBuilder := builders.NewAttendanceDayBuilder(db)
	attendanceDayRepository := repositories.NewAttendanceDayRepository(db)
	attendanceDayService := services.NewAttendanceDayService(attendanceDayBuilder, attendanceDayRepository, attendanceTypeRepository)
	attendanceDayController := controllers.NewAttendanceDayController(attendanceDayService)

	// 休憩
	attendanceBreakBuilder := builders.NewAttendanceBreakBuilder(db)
	attendanceBreakRepository := repositories.NewAttendanceBreakRepository(db)
	attendanceBreakService := services.NewAttendanceBreakService(attendanceBreakBuilder, attendanceBreakRepository, attendanceDayBuilder, attendanceDayRepository)
	attendanceBreakController := controllers.NewAttendanceBreakController(attendanceBreakService)

	// 月次通勤定期
	monthlyCommuterPassBuilder := builders.NewMonthlyCommuterPassBuilder(db)
	monthlyCommuterPassRepository := repositories.NewMonthlyCommuterPassRepository(db)
	monthlyCommuterPassService := services.NewMonthlyCommuterPassService(monthlyCommuterPassBuilder, monthlyCommuterPassRepository)
	monthlyCommuterPassController := controllers.NewMonthlyCommuterPassController(monthlyCommuterPassService)
	user := r.Group("/user")

	// 月次勤怠全体保存
	monthlyAttendanceService := services.NewMonthlyAttendanceService(attendanceDayService, attendanceBreakService, monthlyCommuterPassService)
	monthlyAttendanceController := controllers.NewMonthlyAttendanceController(monthlyAttendanceService)

	/*
	 * 従業員APIは、
	 * 1. JWT認証済みであること
	 * 2. role が USER であること
	 * を必須にする。
	 */
	user.Use(
		middlewares.AuthMiddleware(),
		middlewares.UserMiddleware(),
	)

	{
		// 勤怠区分マスタ(検索のみ)
		user.POST("/attendance-types/search", attendanceTypeController.SearchAttendanceTypes)

		// 勤怠
		user.POST("/attendance-days/search", attendanceDayController.SearchAttendanceDays)
		user.POST("/attendance-days/delete", attendanceDayController.DeleteAttendanceDay)

		// 休憩
		user.POST("/attendance-breaks/search", attendanceBreakController.SearchAttendanceBreaks)
		user.POST("/attendance-breaks/create", attendanceBreakController.CreateAttendanceBreak)
		user.POST("/attendance-breaks/delete", attendanceBreakController.DeleteAttendanceBreak)

		// 月次通勤定期
		user.POST("/monthly-commuter-passes/search", monthlyCommuterPassController.SearchMonthlyCommuterPass)
		user.POST("/monthly-commuter-passes/delete", monthlyCommuterPassController.DeleteMonthlyCommuterPass)

		// 月次勤怠全体保存(勤怠・休憩・月次通勤定期)
		user.POST("/monthly-attendances/update", monthlyAttendanceController.UpdateMonthlyAttendance)
	}
}
