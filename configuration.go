package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	RequestsPerMinute int    `json:"REQUESTS_PER_MINUTE""`
	DBHeathInterval   int64  `json:"DATABASE_HEALTH_LOOP_INTERVAL""`
	DBUser            string `json:"DB_USER"`
	DBPassword        string `json:"DB_PASSWORD"`
	DBName            string `json:"DB_NAME"`
	DBHost            string `json:"DB_HOST"`
	DBPort            string `json:"DB_PORT"`
	DBSSLMode         string `json:"DB_SSLMODE"`
}

func LoadConfiguration(file string) Config {
	var config Config
	// open file from a string
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}
