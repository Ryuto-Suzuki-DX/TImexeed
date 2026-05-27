package routes

import (
	"context"

	"timexeed/backend/internal/mail"
	"timexeed/backend/internal/middlewares"
	"timexeed/backend/internal/storage"

	"timexeed/backend/internal/modules/user/builders"
	"timexeed/backend/internal/modules/user/controllers"
	"timexeed/backend/internal/modules/user/repositories"
	"timexeed/backend/internal/modules/user/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/*
 * 〇 従業員用APIルート登録
 */
func RegisterUserRoutes(r *gin.Engine, db *gorm.DB) {

	// 勤怠区分マスタ
	attendanceTypeBuilder := builders.NewAttendanceTypeBuilder(db)
	attendanceTypeRepository := repositories.NewAttendanceTypeRepository(db)
	attendanceTypeService := services.NewAttendanceTypeService(attendanceTypeBuilder, attendanceTypeRepository)
	attendanceTypeController := controllers.NewAttendanceTypeController(attendanceTypeService)

	// メール送信
	//
	// 注意：
	// ・環境変数が未設定でもアプリ起動自体は止めない
	// ・未設定の場合、メール送信はスキップされる
	// ・お知らせ作成後のメール送信で使用する
	mailService, _ := mail.NewGmailMailServiceFromEnv(context.Background())

	// お知らせ機能
	//
	// 注意：
	// ・月次勤怠申請Serviceでも通知作成に使うため、月次勤怠申請Serviceより先に生成する
	// ・従業員側Controllerでは検索/既読/未読件数のみを公開する
	// ・通知作成はフロントから直接呼ばず、バックエンド内部処理からService経由で行う
	// ・通知作成後のメール送信でmailServiceを使用する
	notificationBuilder := builders.NewNotificationBuilder(db)
	notificationRepository := repositories.NewNotificationRepository(db)
	notificationService := services.NewNotificationService(notificationBuilder, notificationRepository, mailService)
	notificationController := controllers.NewNotificationController(notificationService)

	// 月次勤怠申請
	//
	// 注意：
	// ・申請/再申請/取り下げ成功後に、本人宛と管理者宛のお知らせを作成する
	// ・通知作成は副処理のため、月次勤怠申請ServiceへnotificationServiceを注入する
	monthlyAttendanceRequestBuilder := builders.NewMonthlyAttendanceRequestBuilder(db)
	monthlyAttendanceRequestRepository := repositories.NewMonthlyAttendanceRequestRepository(db)
	monthlyAttendanceRequestService := services.NewMonthlyAttendanceRequestService(
		monthlyAttendanceRequestBuilder,
		monthlyAttendanceRequestRepository,
		notificationService,
	)
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

	// 月次勤怠全体保存
	monthlyAttendanceSaveService := services.NewMonthlyAttendanceSaveService(attendanceDayService, attendanceBreakService, monthlyCommuterPassService, attendanceTypeService, paidLeaveService)
	monthlyAttendanceSaveController := controllers.NewMonthlyAttendanceSaveController(monthlyAttendanceSaveService)

	// Google Drive
	//
	// 注意：
	// ・環境変数が未設定でもアプリ起動自体は止めない
	// ・未設定の場合、領収書アップロード/表示時にService側でエラーを返す
	googleDriveService, _ := storage.NewGoogleDriveServiceFromEnv(context.Background())

	// 個人情報Driveフォルダ
	personalInformationDriveFolderBuilder := builders.NewPersonalInformationDriveFolderBuilder(db)
	personalInformationDriveFolderRepository := repositories.NewPersonalInformationDriveFolderRepository(db)
	personalInformationDriveFolderService := services.NewPersonalInformationDriveFolderService(personalInformationDriveFolderBuilder, personalInformationDriveFolderRepository)
	personalInformationDriveFolderController := controllers.NewPersonalInformationDriveFolderController(personalInformationDriveFolderService)

	// 共有資料Driveフォルダ
	//
	// 従業員側では閲覧のみ。
	// Driveフォルダ作成・Drive権限同期・external_storage_links参照は管理者側APIで行う。
	sharedDocumentDriveFolderBuilder := builders.NewSharedDocumentDriveFolderBuilder(db)
	sharedDocumentDriveFolderRepository := repositories.NewSharedDocumentDriveFolderRepository(db)
	sharedDocumentDriveFolderService := services.NewSharedDocumentDriveFolderService(sharedDocumentDriveFolderBuilder, sharedDocumentDriveFolderRepository)
	sharedDocumentDriveFolderController := controllers.NewSharedDocumentDriveFolderController(sharedDocumentDriveFolderService)

	// 経費
	expenseBuilder := builders.NewExpenseBuilder(db)
	expenseRepository := repositories.NewExpenseRepository(db)
	expenseService := services.NewExpenseService(expenseBuilder, expenseRepository, googleDriveService)
	expenseController := controllers.NewExpenseController(expenseService)

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

		// 祝日
		user.POST("/holiday-dates/search", holidayDateController.SearchHolidayDates)

		// 有給（残数取得のみ）
		user.GET("/paid-leave/balance", paidLeaveController.GetPaidLeaveBalance)

		// 月次勤怠全体保存（勤怠・休憩・月次通勤定期）
		user.POST("/monthly-attendances/update", monthlyAttendanceSaveController.UpdateMonthlyAttendance)

		// お知らせ機能
		user.POST("/notifications/search", notificationController.SearchNotifications)
		user.POST("/notifications/read", notificationController.ReadNotification)
		user.POST("/notifications/unread-count", notificationController.CountUnreadNotifications)

		// 個人情報Driveフォルダ
		user.POST("/personal-information-drive-folders/get", personalInformationDriveFolderController.GetMyPersonalInformationDriveFolder)

		// 共有資料Driveフォルダ
		user.POST("/shared-document-drive-folders/search", sharedDocumentDriveFolderController.SearchSharedDocumentDriveFolders)
		user.POST("/shared-document-drive-folders/detail", sharedDocumentDriveFolderController.DetailSharedDocumentDriveFolder)

		// 経費
		user.POST("/expenses/search", expenseController.SearchExpenses)
		user.POST("/expenses/detail", expenseController.GetExpenseDetail)
		user.POST("/expenses/create", expenseController.CreateExpense)
		user.POST("/expenses/update", expenseController.UpdateExpense)
		user.POST("/expenses/delete", expenseController.DeleteExpense)
		user.POST("/expenses/receipt/view", expenseController.ViewExpenseReceipt)
	}
}
