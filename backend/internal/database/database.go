package database

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/*
 * DB接続
 * Docker環境では DB_HOST=db、DB_PORT=5432 を使用する
 * ローカル起動ではデフォルトで localhost:15432 を使用する
 */
func NewDB() (*gorm.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "15432")
	user := getEnv("DB_USER", "timexeed")
	password := getEnv("DB_PASSWORD", "timexeedpass")
	dbname := getEnv("DB_NAME", "timexeed_db")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Tokyo",
		host,
		port,
		user,
		password,
		dbname,
		sslmode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

/*
 * 環境変数を取得する
 * 未設定の場合はデフォルト値を使う
 */
func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}
