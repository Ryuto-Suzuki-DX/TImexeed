package routes

import (
	"timexeed/backend/internal/handlers"
	"timexeed/backend/internal/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/*
 * 〇 全体ルート登録
 *
 * このファイルは、アプリ全体の入口だけを管理する。
 *
 * ヘルスチェック
 * 	GET /health
 * 		バックエンドが起動しているか確認する
 *
 * 	GET /db-health
 * 		バックエンドからDBへ接続できるか確認する
 *
 * 認証API
 * 	POST /auth/login
 * 		ログインする
 *
 * 	GET /auth/me
 * 		JWTトークンを使ってログイン中のユーザー情報を取得する
 *
 * 管理者API
 * 	/admin/*
 * 		admin_routes.go に分離する
 *
 * 従業員API
 * 	/user/*
 * 		user_routes.go に分離する
 */
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
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", middlewares.AuthMiddleware(), authHandler.Me)
	}

	/*
	 * 管理者用API
	 */
	RegisterAdminRoutes(r, db)

	/*
	 * 従業員用API
	 */
	RegisterUserRoutes(r, db)
}
