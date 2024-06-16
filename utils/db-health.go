package utils

import (
	"go-test/db-utils/models"
	"gorm.io/gorm"
	"log"
	"time"
)

func DataBaseHealthPollingLoop(db *gorm.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sqlDB, err := db.DB()
			if err != nil {
				// try to recreate DB connection here
				log.Fatalf("Error getting DB instance: %v", err)
			}
			err = sqlDB.Ping()
			if err != nil {
				// try to recreate DB connection here
				log.Fatalf("Database connection lost: %v", err)
			}
			// check if model does need migration
			if !db.Migrator().HasTable(&models.Animal{}) {
				// try to migrate gorm model here
				log.Fatalf("Table corrupted: %v", err)
			}
			log.Println("Database connection is healthy.")
		}
	}
}
