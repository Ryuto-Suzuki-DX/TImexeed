package routes

import (
	"timexeed/backend/internal/handlers"
	"timexeed/backend/internal/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	healthHandler := handlers.NewHealthHandler(db)
	authHandler := handlers.NewAuthHandler(db)

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

		// 今後ここに管理者用APIを追加する
		// admin.GET("/users", adminUserHandler.SearchUsers)
		// admin.POST("/users", adminUserHandler.CreateUser)
		// admin.GET("/attendance", adminAttendanceHandler.SearchAttendance)
		// admin.GET("/salary", adminSalaryHandler.SearchSalary)
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
