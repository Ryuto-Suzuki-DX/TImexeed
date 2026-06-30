package main

import (
	"os"

	"timexeed/backend/internal/database"
	"timexeed/backend/internal/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	appEnv := getEnv("APP_ENV", "development")
	appPort := getEnv("APP_PORT", "8080")
	frontendOrigin := getEnv("FRONTEND_ORIGIN", "")

	if appEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.NewDB()
	if err != nil {
		panic("DB接続に失敗しました: " + err.Error())
	}

	/*
	 * gin.Default() はアクセスログ用の gin.Logger() を自動登録する。
	 * 本番ではAWSやDockerへ通常アクセスログを残さない方針のため、
	 * gin.New() と gin.Recovery() のみ使用する。
	 *
	 * API操作ログの独自ミドルウェアや、
	 * Google Driveへの日次アップロード処理には影響しない。
	 */
	r := gin.New()
	r.Use(gin.Recovery())

	allowOrigins := []string{}

	if appEnv == "production" {
		if frontendOrigin != "" {
			allowOrigins = append(allowOrigins, frontendOrigin)
		}
	} else {
		allowOrigins = append(
			allowOrigins,
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3001",
		)

		if frontendOrigin != "" {
			allowOrigins = append(allowOrigins, frontendOrigin)
		}
	}

	if len(allowOrigins) > 0 {
		r.Use(cors.New(cors.Config{
			AllowOrigins: allowOrigins,
			AllowMethods: []string{
				"GET",
				"POST",
				"PUT",
				"PATCH",
				"DELETE",
				"OPTIONS",
			},
			AllowHeaders: []string{
				"Origin",
				"Content-Type",
				"Authorization",
			},
			AllowCredentials: true,
		}))
	}

	routes.RegisterRoutes(r, db)

	if err := r.Run(":" + appPort); err != nil {
		panic("バックエンドの起動に失敗しました: " + err.Error())
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}
