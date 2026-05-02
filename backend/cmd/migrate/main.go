package main

import (
	"log"

	"timexeed/backend/internal/database"
	"timexeed/backend/internal/database/migrations"
)

/*
 * migration手動実行コマンド
 */
func main() {
	db, err := database.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := migrations.RunMigrations(db); err != nil {
		log.Fatal(err)
	}

	log.Println("migration completed")
}
