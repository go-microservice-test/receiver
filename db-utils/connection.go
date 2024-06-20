package dbutils

import (
	"fmt"
	"go-test/db-utils/migrations"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"time"
)

func Connect(DBUser, DBPassword, DBHost, DBName, DBSSLMode, DBPort string) *gorm.DB {
	// construct a connection string
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s", DBUser, DBPassword, DBHost, DBPort, DBName, DBSSLMode)
	// connect to database
	var db *gorm.DB
	var err error

	// retry mechanism
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
		if err == nil {
			fmt.Println("Connected to the database")
			break
		}

		log.Printf("Failed to connect to database, attempt %d/%d: %v", i+1, maxRetries, err)
		time.Sleep(1 * time.Second) // wait for 1 second before retrying
	}

	if err != nil {
		log.Fatal("Could not connect to the database after several attempts: ", err)
	}

	// creating table if not exists
	err = migrations.MigrateAllTables(db)
	if err != nil {
		log.Fatal(err)
	}
	// return database object reference
	return db
}
