package routes

import (
	"timexeed/backend/internal/modules/admin/builders"
	"timexeed/backend/internal/modules/admin/controllers"
	"timexeed/backend/internal/modules/admin/repositories"
	"timexeed/backend/internal/modules/admin/services"

	"timexeed/backend/internal/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/*
 * 〇 管理者用APIルート登録
 *
 * 管理者だけが使うAPIをここにまとめる。
 *
 * 重要：
 * 	管理者APIと従業員APIは完全に分離する。
 * 	従業員画面から使うAPIはここには書かない。
 *
 * ルール：
 * 	URLにIDを載せない。
 * 	targetUserId や attendanceId などは request body で受け取る。
 *
 * 管理者APIで扱うID：
 * 	操作している管理者本人のID
 * 		→ JWTから取得する
 *
 * 	操作対象のユーザーID
 * 		→ request body の targetUserId / targetUserIds で受け取る
 *
 * 管理者勤怠API方針：
 * 	従業員APIと同じ粒度でAPIを分離する。
 * 	ただし、対象ユーザーはJWTからではなく request body の targetUserId で受け取る。
 * 	月次申請状態に関係なく、管理者は編集・保存できる。
 */
func RegisterAdminRoutes(r *gin.Engine, db *gorm.DB) {

	// 所属
	departmentBuilder := builders.NewDepartmentBuilder(db)
	departmentRepository := repositories.NewDepartmentRepository(db)
	departmentService := services.NewDepartmentService(departmentBuilder, departmentRepository)
	departmentController := controllers.NewDepartmentController(departmentService)

	// ユーザー
	userBuilder := builders.NewUserBuilder(db)
	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userBuilder, userRepository)
	userController := controllers.NewUserController(userService)

	// 勤怠区分マスタ
	attendanceTypeBuilder := builders.NewAttendanceTypeBuilder(db)
	attendanceTypeRepository := repositories.NewAttendanceTypeRepository(db)
	attendanceTypeService := services.NewAttendanceTypeService(attendanceTypeBuilder, attendanceTypeRepository)
	attendanceTypeController := controllers.NewAttendanceTypeController(attendanceTypeService)

	// 月次勤怠申請
	monthlyAttendanceRequestBuilder := builders.NewMonthlyAttendanceRequestBuilder(db)
	monthlyAttendanceRequestRepository := repositories.NewMonthlyAttendanceRequestRepository(db)
	monthlyAttendanceRequestService := services.NewMonthlyAttendanceRequestService(monthlyAttendanceRequestBuilder, monthlyAttendanceRequestRepository)
	monthlyAttendanceRequestController := controllers.NewMonthlyAttendanceRequestController(monthlyAttendanceRequestService)

	// 勤怠
	attendanceDayBuilder := builders.NewAttendanceDayBuilder(db)
	attendanceDayRepository := repositories.NewAttendanceDayRepository(db)
	attendanceDayService := services.NewAttendanceDayService(attendanceDayBuilder, attendanceDayRepository, attendanceTypeRepository, monthlyAttendanceRequestBuilder, monthlyAttendanceRequestRepository)
	attendanceDayController := controllers.NewAttendanceDayController(attendanceDayService)

	// 休憩
	attendanceBreakBuilder := builders.NewAttendanceBreakBuilder(db)
	attendanceBreakRepository := repositories.NewAttendanceBreakRepository(db)
	attendanceBreakService := services.NewAttendanceBreakService(attendanceBreakBuilder, attendanceBreakRepository, attendanceDayBuilder, attendanceDayRepository, monthlyAttendanceRequestBuilder, monthlyAttendanceRequestRepository)
	attendanceBreakController := controllers.NewAttendanceBreakController(attendanceBreakService)

	// 月次通勤定期
	monthlyCommuterPassBuilder := builders.NewMonthlyCommuterPassBuilder(db)
	monthlyCommuterPassRepository := repositories.NewMonthlyCommuterPassRepository(db)
	monthlyCommuterPassService := services.NewMonthlyCommuterPassService(monthlyCommuterPassBuilder, monthlyCommuterPassRepository, monthlyAttendanceRequestBuilder, monthlyAttendanceRequestRepository)
	monthlyCommuterPassController := controllers.NewMonthlyCommuterPassController(monthlyCommuterPassService)

	// 祝日
	holidayDateBuilder := builders.NewHolidayDateBuilder(db)
	holidayDateRepository := repositories.NewHolidayDateRepository(db)
	holidayDateService := services.NewHolidayDateService(holidayDateBuilder, holidayDateRepository)
	holidayDateController := controllers.NewHolidayDateController(holidayDateService)

	// 有給（残数取得のみ）
	paidLeaveBuilder := builders.NewPaidLeaveBuilder(db)
	paidLeaveRepository := repositories.NewPaidLeaveRepository(db)
	paidLeaveService := services.NewPaidLeaveService(paidLeaveBuilder, paidLeaveRepository)
	paidLeaveController := controllers.NewPaidLeaveController(paidLeaveService)

	// 有給使用日
	paidLeaveUsageBuilder := builders.NewPaidLeaveUsageBuilder(db)
	paidLeaveUsageRepository := repositories.NewPaidLeaveUsageRepository(db)
	paidLeaveUsageService := services.NewPaidLeaveUsageService(paidLeaveUsageBuilder, paidLeaveUsageRepository)
	paidLeaveUsageController := controllers.NewPaidLeaveUsageController(paidLeaveUsageService)

	// 月次勤怠全体保存
	monthlyAttendanceService := services.NewMonthlyAttendanceService(attendanceDayService, attendanceBreakService, monthlyCommuterPassService, attendanceTypeService, paidLeaveService)
	monthlyAttendanceController := controllers.NewMonthlyAttendanceController(monthlyAttendanceService)

	admin := r.Group("/admin")

	/*
	 * 管理者APIは、
	 * 1. JWT認証済みであること
	 * 2. role が ADMIN であること
	 * を必須にする。
	 */
	admin.Use(
		middlewares.AuthMiddleware(),
		middlewares.AdminMiddleware(),
	)

	{
		// 所属
		admin.POST("/departments/search", departmentController.SearchDepartments)
		admin.POST("/departments/detail", departmentController.GetDepartmentDetail)
		admin.POST("/departments/create", departmentController.CreateDepartment)
		admin.POST("/departments/update", departmentController.UpdateDepartment)
		admin.POST("/departments/delete", departmentController.DeleteDepartment)

		// ユーザー
		admin.POST("/users/search", userController.SearchUsers)
		admin.POST("/users/detail", userController.GetUserDetail)
		admin.POST("/users/create", userController.CreateUser)
		admin.POST("/users/update", userController.UpdateUser)
		admin.POST("/users/delete", userController.DeleteUser)

		// 勤怠区分マスタ（検索のみ）
		admin.POST("/attendance-types/search", attendanceTypeController.SearchAttendanceTypes)

		// 月次勤怠申請
		admin.POST("/monthly-attendance-requests/status", monthlyAttendanceRequestController.GetMonthlyAttendanceRequestStatus)

		// 勤怠
		admin.POST("/attendance-days/search", attendanceDayController.SearchAttendanceDays)

		// 休憩
		admin.POST("/attendance-breaks/search", attendanceBreakController.SearchAttendanceBreaks)

		// 月次通勤定期
		admin.POST("/monthly-commuter-passes/search", monthlyCommuterPassController.SearchMonthlyCommuterPass)

		// 祝日
		admin.POST("/holiday-dates/import", holidayDateController.ImportHolidayDates)
		admin.POST("/holiday-dates/search", holidayDateController.SearchHolidayDates)

		// 有給（残数取得のみ）
		admin.POST("/paid-leave/balance", paidLeaveController.GetPaidLeaveBalance)

		// 有給使用日
		admin.POST("/paid-leave-usages/search", paidLeaveUsageController.SearchPaidLeaveUsages)
		admin.POST("/paid-leave-usages/balance", paidLeaveUsageController.GetPaidLeaveBalance)
		admin.POST("/paid-leave-usages/create", paidLeaveUsageController.CreatePaidLeaveUsage)
		admin.POST("/paid-leave-usages/update", paidLeaveUsageController.UpdatePaidLeaveUsage)
		admin.POST("/paid-leave-usages/delete", paidLeaveUsageController.DeletePaidLeaveUsage)

		// 月次勤怠全体保存（勤怠・休憩・月次通勤定期）
		admin.POST("/monthly-attendances/update", monthlyAttendanceController.UpdateMonthlyAttendance)
	}
}
