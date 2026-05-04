package routes

import (
	"timexeed/backend/internal/handlers"
	"timexeed/backend/internal/middlewares"

	adminBuilders "timexeed/backend/internal/admin/builders"
	adminControllers "timexeed/backend/internal/admin/controllers"
	adminRepositories "timexeed/backend/internal/admin/repositories"
	adminServices "timexeed/backend/internal/admin/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/*
 * 管理者用ルート登録
 *
 * 依存関係とそのAPIを記載
 *
 * 所属管理					ユーザーの所属
 * ユーザー管理				ユーザーのCRUD
 * 有給管理					サービス使用前の既存有給を入れる
 * 有給申請承認				有給申請を承認する
 * 月次申請承認				各月の承認
 * 勤怠					    全ユーザーの勤怠閲覧
 * 休憩					   	各日の休憩
 * 各日交通費				各日の交通費
 * 各月交通定期				各月の交通定期
 *
 *
 *
 *
 */
func RegisterAdminRoutes(r *gin.Engine, db *gorm.DB) {
	authHandler := handlers.NewAuthHandler(db)

	/*
	 * 管理者用 依存関係
	 */

	// 所属管理
	departmentRepository := adminRepositories.NewDepartmentRepository()
	departmentBuilder := adminBuilders.NewDepartmentBuilder()
	departmentService := adminServices.NewDepartmentService(
		db,
		departmentRepository,
		departmentBuilder,
	)
	departmentController := adminControllers.NewDepartmentController(
		departmentService,
	)

	// ユーザー管理
	userRepository := adminRepositories.NewUserRepository()
	userBuilder := adminBuilders.NewUserBuilder()
	userService := adminServices.NewUserService(
		db,
		userRepository,
		userBuilder,
		departmentRepository,
		departmentBuilder,
	)
	userController := adminControllers.NewUserController(
		userService,
	)

	// 有給管理
	paidLeaveRepository := adminRepositories.NewPaidLeaveRepository()
	paidLeaveBuilder := adminBuilders.NewPaidLeaveBuilder()
	paidLeaveService := adminServices.NewPaidLeaveService(
		db,
		paidLeaveRepository,
		paidLeaveBuilder,
		departmentRepository,
		departmentBuilder,
	)
	paidLeaveController := adminControllers.NewPaidLeaveController(
		paidLeaveService,
	)

	// 有給申請承認
	paidLeaveRequestRepository := adminRepositories.NewPaidLeaveRequestRepository()
	paidLeaveRequestBuilder := adminBuilders.NewPaidLeaveRequestBuilder()
	paidLeaveRequestService := adminServices.NewPaidLeaveRequestService(
		db,
		paidLeaveRequestRepository,
		paidLeaveRequestBuilder,
	)
	paidLeaveRequestController := adminControllers.NewPaidLeaveRequestController(
		paidLeaveRequestService,
	)

	// 月次勤怠申請承認
	monthlyAttendanceRequestRepository := adminRepositories.NewMonthlyAttendanceRequestRepository()
	monthlyAttendanceRequestBuilder := adminBuilders.NewMonthlyAttendanceRequestBuilder()
	monthlyAttendanceRequestService := adminServices.NewMonthlyAttendanceRequestService(
		db,
		monthlyAttendanceRequestRepository,
		monthlyAttendanceRequestBuilder,
	)
	monthlyAttendanceRequestController := adminControllers.NewMonthlyAttendanceRequestController(
		monthlyAttendanceRequestService,
	)

	// 管理者用 勤怠
	attendanceRecordRepository := adminRepositories.NewAttendanceRecordRepository()
	attendanceRecordBuilder := adminBuilders.NewAttendanceRecordBuilder()
	attendanceRecordService := adminServices.NewAttendanceRecordService(
		db,
		attendanceRecordRepository,
		attendanceRecordBuilder,
	)
	attendanceRecordController := adminControllers.NewAttendanceRecordController(
		attendanceRecordService,
	)

	// 勤怠休憩
	attendanceBreakRepository := adminRepositories.NewAttendanceBreakRepository()
	attendanceBreakBuilder := adminBuilders.NewAttendanceBreakBuilder()
	attendanceBreakService := adminServices.NewAttendanceBreakService(
		db,
		attendanceBreakRepository,
		attendanceBreakBuilder,
	)
	attendanceBreakController := adminControllers.NewAttendanceBreakController(
		attendanceBreakService,
	)

	// 勤怠交通費管理
	attendanceTransportationRepository := adminRepositories.NewAttendanceTransportationRepository()
	attendanceTransportationBuilder := adminBuilders.NewAttendanceTransportationBuilder()
	attendanceTransportationService := adminServices.NewAttendanceTransportationService(
		db,
		attendanceTransportationRepository,
		attendanceTransportationBuilder,
	)
	attendanceTransportationController := adminControllers.NewAttendanceTransportationController(
		attendanceTransportationService,
	)

	// 交通定期管理
	commuterPassRepository := adminRepositories.NewCommuterPassRepository()
	commuterPassBuilder := adminBuilders.NewCommuterPassBuilder()
	commuterPassService := adminServices.NewCommuterPassService(
		db,
		commuterPassRepository,
		commuterPassBuilder,
	)
	commuterPassController := adminControllers.NewCommuterPassController(
		commuterPassService,
	)

	/*
	 * 管理者用API
	 * ログイン必須 + 管理者権限必須
	 */
	admin := r.Group("/admin")
	admin.Use(
		middlewares.AuthMiddleware(),
		middlewares.AdminMiddleware(),
	)
	{
		// 管理者マイページ確認用
		admin.GET("/me", authHandler.Me)

		/*
		 * ユーザー管理
		 */
		// 検索
		admin.GET("/users", userController.SearchUsers)
		// 詳細
		admin.GET("/users/:id", userController.GetUser)
		// 新規作成
		admin.POST("/users", userController.CreateUser)
		// 更新
		admin.PUT("/users/:id", userController.UpdateUser)
		// 削除
		admin.DELETE("/users/:id", userController.DeleteUser)

		/*
		 * 所属管理
		 */
		// 検索
		admin.GET("/departments", departmentController.SearchDepartments)
		// 詳細
		admin.GET("/departments/:id", departmentController.GetDepartment)
		// 新規作成
		admin.POST("/departments", departmentController.CreateDepartment)
		// 更新
		admin.PUT("/departments/:id", departmentController.UpdateDepartment)
		// 削除
		admin.DELETE("/departments/:id", departmentController.DeleteDepartment)

		/*
		 * 有給管理
		 */
		//
		admin.GET("/paid-leaves", paidLeaveController.SearchPaidLeaveSummaries)
		// 詳細
		admin.GET("/paid-leaves/:userId", paidLeaveController.GetPaidLeaveDetail)
		// 過去分調整有給　新規作成
		admin.POST("/paid-leave-adjustments", paidLeaveController.CreatePaidLeaveAdjustment)
		// 過去分調整有給　詳細
		admin.GET("/paid-leave-adjustments/:id", paidLeaveController.GetPaidLeaveAdjustment)
		// 過去分調整有給　更新
		admin.PUT("/paid-leave-adjustments/:id", paidLeaveController.UpdatePaidLeaveAdjustment)
		// 過去分調整有給　削除
		admin.DELETE("/paid-leave-adjustments/:id", paidLeaveController.DeletePaidLeaveAdjustment)

		/*
		 * 有給申請承認管理
		 */
		// 検索
		admin.GET("/paid-leave-requests", paidLeaveRequestController.SearchPaidLeaveRequests)
		// 申請
		admin.PUT("/paid-leave-requests/:id/approve", paidLeaveRequestController.ApprovePaidLeaveRequest)
		// 否認
		admin.PUT("/paid-leave-requests/:id/reject", paidLeaveRequestController.RejectPaidLeaveRequest)

		/*
		 * 管理者用 勤怠閲覧・編集
		 */
		// 検索
		admin.GET("/attendance-records", attendanceRecordController.SearchAttendanceRecords)
		// 詳細
		admin.GET("/attendance-records/:id", attendanceRecordController.GetAttendanceRecord)
		// 新規作成
		admin.POST("/attendance-records", attendanceRecordController.CreateAttendanceRecord)
		// 更新
		admin.PUT("/attendance-records/:id", attendanceRecordController.UpdateAttendanceRecord)
		// 削除
		admin.DELETE("/attendance-records/:id", attendanceRecordController.DeleteAttendanceRecord)

		/*
		 * 管理者用 勤怠休憩閲覧・編集
		 */
		// 検索
		admin.GET("/attendance-breaks", attendanceBreakController.SearchAttendanceBreaks)
		// 詳細
		admin.GET("/attendance-breaks/:id", attendanceBreakController.GetAttendanceBreak)
		// 新規作成
		admin.POST("/attendance-breaks", attendanceBreakController.CreateAttendanceBreak)
		// 更新
		admin.PUT("/attendance-breaks/:id", attendanceBreakController.UpdateAttendanceBreak)
		// 削除
		admin.DELETE("/attendance-breaks/:id", attendanceBreakController.DeleteAttendanceBreak)

		/*
		 * 管理者用 勤怠交通費閲覧・編集
		 */
		// 検索
		admin.GET("/attendance-transportations", attendanceTransportationController.SearchAttendanceTransportations)
		// 詳細
		admin.GET("/attendance-transportations/:id", attendanceTransportationController.GetAttendanceTransportation)
		// 新規作成
		admin.POST("/attendance-transportations", attendanceTransportationController.CreateAttendanceTransportation)
		// 更新
		admin.PUT("/attendance-transportations/:id", attendanceTransportationController.UpdateAttendanceTransportation)
		// 削除
		admin.DELETE("/attendance-transportations/:id", attendanceTransportationController.DeleteAttendanceTransportation)

		/*
		 * 管理者用 定期閲覧・編集
		 */
		// 検索
		admin.GET("/commuter-passes", commuterPassController.SearchCommuterPasses)
		// 詳細
		admin.GET("/commuter-passes/:id", commuterPassController.GetCommuterPass)
		// 新規作成
		admin.POST("/commuter-passes", commuterPassController.CreateCommuterPass)
		// 更新
		admin.PUT("/commuter-passes/:id", commuterPassController.UpdateCommuterPass)
		// 削除
		admin.DELETE("/commuter-passes/:id", commuterPassController.DeleteCommuterPass)

		/*
		 * 月次勤怠申請承認管理
		 */
		// 検索
		admin.GET("/monthly-attendance-requests", monthlyAttendanceRequestController.SearchMonthlyAttendanceRequests)
		// 承認
		admin.PUT("/monthly-attendance-requests/:id/approve", monthlyAttendanceRequestController.ApproveMonthlyAttendanceRequest)
		// 否認
		admin.PUT("/monthly-attendance-requests/:id/reject", monthlyAttendanceRequestController.RejectMonthlyAttendanceRequest)
	}
}
