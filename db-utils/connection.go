package db_utils

import (
	"database/sql"
	"fmt"
	"log"
)

func Connect(DBUser, DBPassword, DBHost, DBName, DBSSLMode, DBPort string) *sql.DB {
	// construct a connection string
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s", DBUser, DBPassword, DBHost, DBPort, DBName, DBSSLMode)
	// connect to database
	var err error
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	// Verify the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to the database")

	// creating table if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS animals (
    	id SERIAL PRIMARY KEY,
    	name VARCHAR(100),
    	type INT,
    	description VARCHAR(500),
    	is_active BOOLEAN
    )`)
	if err != nil {
		log.Fatal(err)
	}
	// return database object reference
	return db
}
