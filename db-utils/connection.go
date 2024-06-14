package dbutils

import (
	"fmt"
	"go-test/db-utils/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func Connect(DBUser, DBPassword, DBHost, DBName, DBSSLMode, DBPort string) *gorm.DB {
	// construct a connection string
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s", DBUser, DBPassword, DBHost, DBPort, DBName, DBSSLMode)
	// connect to database
	var err error
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to the database")

	// creating table if not exists
	err = db.AutoMigrate(&models.Animal{})
	if err != nil {
		log.Fatal(err)
	}
	// return database object reference
	return db
}
