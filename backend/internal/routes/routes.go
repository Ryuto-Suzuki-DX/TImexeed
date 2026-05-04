package routes

import (
	"timexeed/backend/internal/handlers"
	"timexeed/backend/internal/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/*
 * 〇全体ルート登録
 *
 * ヘルスチェック
 *	：health 	バックエンド接続確認
 *	：db-health バックエンドからDBの接続確認
 *
 * 認証API
 *	：login		ログイン用
 *	：me		JWTトークンを使ってログイン中のユーザー情報を取得する
 *
 * 管理者
 * 	：admin_routes.go
 * ユーザー
 *	：user_routes.go
 *
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
	RegisterAdminRoutes(r, db, authHandler)

	/*
	 * 一般ユーザー用API
	 */
	RegisterUserRoutes(r, db, authHandler)
}
