package routes

import (
	"timexeed/backend/internal/handlers"
	"timexeed/backend/internal/middlewares"

	userBuilders "timexeed/backend/internal/user/builders"
	userControllers "timexeed/backend/internal/user/controllers"
	userRepositories "timexeed/backend/internal/user/repositories"
	userServices "timexeed/backend/internal/user/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/*
 * 一般ユーザー用ルート登録
 *
 * 勤怠					    全ユーザーの勤怠閲覧
 * 休憩					   	各日の休憩
 * 各日交通費				各日の交通費
 * 各月交通定期				各月の交通定期
 * 有給管理					サービス使用前の既存有給を入れる
 * 有給申請承認				有給申請を承認する
 * 月次申請承認				各月の承認
 * 有給付与					？？？？？？？？？？？？？？？？？？？？
 * 有給残数					有給残数表示用
 * 有給申請					申請がない、、、作成したあとに申請じゃなくて？　削除もない
 * 月次申請					検索いるかこれ？GETだろ　月ごとの勤怠申請　これは新規作成で申請
 *
 *
 *
 *
 *
 *
 *
 *
 */
func RegisterUserRoutes(r *gin.Engine, db *gorm.DB) {
	authHandler := handlers.NewAuthHandler(db)

	/*
	 * 一般ユーザー用 依存関係
	 */

	// 勤怠管理
	attendanceRecordRepository := userRepositories.NewAttendanceRecordRepository()
	attendanceRecordBuilder := userBuilders.NewAttendanceRecordBuilder()
	attendanceRecordService := userServices.NewAttendanceRecordService(
		db,
		attendanceRecordRepository,
		attendanceRecordBuilder,
	)
	attendanceRecordController := userControllers.NewAttendanceRecordController(
		attendanceRecordService,
	)

	// 勤怠休憩
	attendanceBreakRepository := userRepositories.NewAttendanceBreakRepository()
	attendanceBreakBuilder := userBuilders.NewAttendanceBreakBuilder()
	attendanceBreakService := userServices.NewAttendanceBreakService(
		db,
		attendanceBreakRepository,
		attendanceBreakBuilder,
	)
	attendanceBreakController := userControllers.NewAttendanceBreakController(
		attendanceBreakService,
	)

	// 交通費
	attendanceTransportationRepository := userRepositories.NewAttendanceTransportationRepository()
	attendanceTransportationBuilder := userBuilders.NewAttendanceTransportationBuilder()
	attendanceTransportationService := userServices.NewAttendanceTransportationService(
		db,
		attendanceTransportationRepository,
		attendanceTransportationBuilder,
	)
	attendanceTransportationController := userControllers.NewAttendanceTransportationController(
		attendanceTransportationService,
	)

	// 交通定期
	commuterPassRepository := userRepositories.NewCommuterPassRepository()
	commuterPassBuilder := userBuilders.NewCommuterPassBuilder()
	commuterPassService := userServices.NewCommuterPassService(
		db,
		commuterPassRepository,
		commuterPassBuilder,
	)
	commuterPassController := userControllers.NewCommuterPassController(
		commuterPassService,
	)

	// 有給付与管理
	paidLeaveGrantRepository := userRepositories.NewPaidLeaveGrantRepository()
	paidLeaveGrantBuilder := userBuilders.NewPaidLeaveGrantBuilder()
	paidLeaveGrantService := userServices.NewPaidLeaveGrantService(
		db,
		paidLeaveGrantRepository,
		paidLeaveGrantBuilder,
	)
	paidLeaveGrantController := userControllers.NewPaidLeaveGrantController(
		paidLeaveGrantService,
	)

	// 有給残数管理
	paidLeaveBalanceRepository := userRepositories.NewPaidLeaveBalanceRepository()
	paidLeaveBalanceBuilder := userBuilders.NewPaidLeaveBalanceBuilder()
	paidLeaveBalanceService := userServices.NewPaidLeaveBalanceService(
		db,
		paidLeaveBalanceRepository,
		paidLeaveBalanceBuilder,
	)
	paidLeaveBalanceController := userControllers.NewPaidLeaveBalanceController(
		paidLeaveBalanceService,
	)

	// 有給申請管理
	paidLeaveRequestRepository := userRepositories.NewPaidLeaveRequestRepository()
	paidLeaveRequestBuilder := userBuilders.NewPaidLeaveRequestBuilder()
	paidLeaveRequestService := userServices.NewPaidLeaveRequestService(
		db,
		paidLeaveRequestRepository,
		paidLeaveRequestBuilder,
	)
	paidLeaveRequestController := userControllers.NewPaidLeaveRequestController(
		paidLeaveRequestService,
	)

	// 月次勤怠申請管理
	monthlyAttendanceRequestRepository := userRepositories.NewMonthlyAttendanceRequestRepository()
	monthlyAttendanceRequestBuilder := userBuilders.NewMonthlyAttendanceRequestBuilder()
	monthlyAttendanceRequestService := userServices.NewMonthlyAttendanceRequestService(
		db,
		monthlyAttendanceRequestRepository,
		monthlyAttendanceRequestBuilder,
	)
	monthlyAttendanceRequestController := userControllers.NewMonthlyAttendanceRequestController(
		monthlyAttendanceRequestService,
	)

	/*
	 * 一般ユーザー用API
	 * ログイン必須
	 */
	user := r.Group("/user")
	user.Use(
		middlewares.AuthMiddleware(),
	)
	{
		// ユーザーマイページ確認用
		user.GET("/me", authHandler.Me)

		/*
		 * 勤怠管理
		 */
		// 検索
		user.GET("/attendance-records", attendanceRecordController.SearchAttendanceRecords)
		// 詳細
		user.GET("/attendance-records/:id", attendanceRecordController.GetAttendanceRecord)
		// 新規作成
		user.POST("/attendance-records", attendanceRecordController.CreateAttendanceRecord)
		// 更新
		user.PUT("/attendance-records/:id", attendanceRecordController.UpdateAttendanceRecord)
		// 削除
		user.DELETE("/attendance-records/:id", attendanceRecordController.DeleteAttendanceRecord)

		/*
		 * 勤怠休憩管理
		 */
		// 検索
		user.GET("/attendance-breaks", attendanceBreakController.SearchAttendanceBreaks)
		// 詳細
		user.GET("/attendance-breaks/:id", attendanceBreakController.GetAttendanceBreak)
		// 新規作成
		user.POST("/attendance-breaks", attendanceBreakController.CreateAttendanceBreak)
		// 更新
		user.PUT("/attendance-breaks/:id", attendanceBreakController.UpdateAttendanceBreak)
		// 削除
		user.DELETE("/attendance-breaks/:id", attendanceBreakController.DeleteAttendanceBreak)

		/*
		 * 勤怠交通費管理
		 */
		// 検索
		user.GET("/attendance-transportations", attendanceTransportationController.SearchAttendanceTransportations)
		// 詳細
		user.GET("/attendance-transportations/:id", attendanceTransportationController.GetAttendanceTransportation)
		// 新規作成
		user.POST("/attendance-transportations", attendanceTransportationController.CreateAttendanceTransportation)
		// 更新
		user.PUT("/attendance-transportations/:id", attendanceTransportationController.UpdateAttendanceTransportation)
		// 削除
		user.DELETE("/attendance-transportations/:id", attendanceTransportationController.DeleteAttendanceTransportation)

		/*
		 * 定期管理
		 */
		// 検索
		user.GET("/commuter-passes", commuterPassController.SearchCommuterPasses)
		// 詳細
		user.GET("/commuter-passes/:id", commuterPassController.GetCommuterPass)
		// 新規作成
		user.POST("/commuter-passes", commuterPassController.CreateCommuterPass)
		// 更新
		user.PUT("/commuter-passes/:id", commuterPassController.UpdateCommuterPass)
		// 削除
		user.DELETE("/commuter-passes/:id", commuterPassController.DeleteCommuterPass)

		/*
		 * 有給付与
		 */
		// 検索
		user.GET("/paid-leave-grants", paidLeaveGrantController.SearchPaidLeaveGrants)
		// 詳細
		user.GET("/paid-leave-grants/:id", paidLeaveGrantController.GetPaidLeaveGrant)
		// 新規作成
		user.POST("/paid-leave-grants", paidLeaveGrantController.CreatePaidLeaveGrant)
		// 更新
		user.PUT("/paid-leave-grants/:id", paidLeaveGrantController.UpdatePaidLeaveGrant)
		// 削除
		user.DELETE("/paid-leave-grants/:id", paidLeaveGrantController.DeletePaidLeaveGrant)

		/*
		 * 有給残数
		 */
		// 有給の残数取得
		user.GET("/paid-leave-balance", paidLeaveBalanceController.GetPaidLeaveBalance)

		/*
		 * 有給申請
		 */
		// 検索
		user.GET("/paid-leave-requests", paidLeaveRequestController.SearchPaidLeaveRequests)
		// 新規作成
		user.POST("/paid-leave-requests", paidLeaveRequestController.CreatePaidLeaveRequest)
		// 取り下げ
		user.PATCH("/paid-leave-requests/:id/withdraw", paidLeaveRequestController.WithdrawPaidLeaveRequest)

		/*
		 * 月次勤怠申請管理
		 */
		// 検索　→　これいる？　あってもGETじぇね？
		user.GET("/monthly-attendance-requests", monthlyAttendanceRequestController.SearchMonthlyAttendanceRequests)
		// 新規作成
		user.POST("/monthly-attendance-requests", monthlyAttendanceRequestController.CreateMonthlyAttendanceRequest)
	}
}
