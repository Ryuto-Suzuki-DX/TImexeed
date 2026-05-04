package migrations

import (
	"gorm.io/gorm"
)

/*
 * DBマイグレーション
 *
 * テーブル構造をDBへ反映する
 */
func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate()
}
