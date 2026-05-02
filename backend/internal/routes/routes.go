package routes

import (
	"timexeed/backend/internal/handlers"
	"timexeed/backend/internal/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	adminBuilders "timexeed/backend/internal/admin/builders"
	adminControllers "timexeed/backend/internal/admin/controllers"
	adminRepositories "timexeed/backend/internal/admin/repositories"
	adminServices "timexeed/backend/internal/admin/services"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {

	// ヘルスチェック
	healthHandler := handlers.NewHealthHandler(db)
	authHandler := handlers.NewAuthHandler(db)

	// 所属管理
	departmentRepository := adminRepositories.NewDepartmentRepository()
	departmentBuilder := adminBuilders.NewDepartmentBuilder()
	departmentService := adminServices.NewDepartmentService(db, departmentRepository, departmentBuilder)
	departmentController := adminControllers.NewDepartmentController(departmentService)

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
	userController := adminControllers.NewUserController(userService)

	/*
	 * ヘルスチェック
	 */
	r.GET("/health", healthHandler.Health)
	r.GET("/db-health", healthHandler.DBHealth)

	/*
	 * 認証API
	 */
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)

		// ログイン中ユーザー確認
		auth.GET("/me", middlewares.AuthMiddleware(), authHandler.Me)
	}

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

		// ユーザー管理
		admin.GET("/users", userController.SearchUsers)
		admin.GET("/users/:id", userController.GetUser)
		admin.POST("/users", userController.CreateUser)
		admin.PUT("/users/:id", userController.UpdateUser)
		admin.DELETE("/users/:id", userController.DeleteUser)

		// 所属管理
		admin.GET("/departments", departmentController.SearchDepartments)
		admin.GET("/departments/:id", departmentController.GetDepartment)
		admin.POST("/departments", departmentController.CreateDepartment)
		admin.PUT("/departments/:id", departmentController.UpdateDepartment)
		admin.DELETE("/departments/:id", departmentController.DeleteDepartment)
	}

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

		// 今後ここに一般ユーザー用APIを追加する
		// user.GET("/attendance", userAttendanceHandler.SearchMyAttendance)
		// user.GET("/salary", userSalaryHandler.SearchMySalary)
		// user.GET("/drive-files", userDriveFileHandler.SearchMyDriveFiles)
	}
}
