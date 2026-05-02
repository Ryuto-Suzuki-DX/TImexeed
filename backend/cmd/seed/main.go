package main

import (
	"log"

	"timexeed/backend/internal/database"
	"timexeed/backend/internal/database/seeders"
)

/*
 * seeder手動実行コマンド
 */
func main() {
	db, err := database.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := seeders.RunSeeders(db); err != nil {
		log.Fatal(err)
	}

	log.Println("seeder completed")
}
