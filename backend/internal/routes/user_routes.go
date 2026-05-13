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

	// 有給（残数取得のみ）
	paidLeaveBuilder := builders.NewPaidLeaveBuilder(db)
	paidLeaveRepository := repositories.NewPaidLeaveRepository(db)
	paidLeaveService := services.NewPaidLeaveService(paidLeaveBuilder, paidLeaveRepository)
	paidLeaveController := controllers.NewPaidLeaveController(paidLeaveService)

	// 月次勤怠全体保存
	monthlyAttendanceService := services.NewMonthlyAttendanceService(attendanceDayService, attendanceBreakService, monthlyCommuterPassService, attendanceTypeService, paidLeaveService)
	monthlyAttendanceController := controllers.NewMonthlyAttendanceController(monthlyAttendanceService)

	// お知らせ機能
	notificationBuilder := builders.NewNotificationBuilder(db)
	notificationRepository := repositories.NewNotificationRepository(db)
	notificationService := services.NewNotificationService(notificationBuilder, notificationRepository)
	notificationController := controllers.NewNotificationController(notificationService)

	user := r.Group("/user")

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
		// 勤怠区分マスタ（検索のみ）
		user.POST("/attendance-types/search", attendanceTypeController.SearchAttendanceTypes)

		// 月次勤怠申請
		user.POST("/monthly-attendance-requests/status", monthlyAttendanceRequestController.GetMonthlyAttendanceRequestStatus)
		user.POST("/monthly-attendance-requests/submit", monthlyAttendanceRequestController.SubmitMonthlyAttendanceRequest)
		user.POST("/monthly-attendance-requests/cancel", monthlyAttendanceRequestController.CancelMonthlyAttendanceRequest)

		// 勤怠
		user.POST("/attendance-days/search", attendanceDayController.SearchAttendanceDays)

		// 休憩
		user.POST("/attendance-breaks/search", attendanceBreakController.SearchAttendanceBreaks)

		// 月次通勤定期
		user.POST("/monthly-commuter-passes/search", monthlyCommuterPassController.SearchMonthlyCommuterPass)

		// 有給（残数取得のみ）
		user.GET("/paid-leave/balance", paidLeaveController.GetPaidLeaveBalance)

		// 月次勤怠全体保存（勤怠・休憩・月次通勤定期）
		user.POST("/monthly-attendances/update", monthlyAttendanceController.UpdateMonthlyAttendance)

		// お知らせ機能
		user.POST("/notifications/search", notificationController.SearchNotifications)
		user.POST("/notifications/read", notificationController.ReadNotification)
		user.POST("/notifications/unread-count", notificationController.CountUnreadNotifications)
	}
}
