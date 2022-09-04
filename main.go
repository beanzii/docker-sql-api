package main

import (
	"beanzii/docker-sql-api/pkg/config"
	"beanzii/docker-sql-api/pkg/database"
	"beanzii/docker-sql-api/pkg/server"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// config.Load(envLocation)  using .env file path from os.Args[1]
	if err := loadDotEnv(); err != nil {
		exit(err)
	}

	// Open DB Connection
	db, err := database.Open()
	if err != nil {
		exit(err)
	}
	defer db.Close()

	// Start http server
	if err := server.Start(); err != nil {
		exit(err)
	}
}

func loadDotEnv() error {
	var envPath string
	// Confirm arguments exist
	if len(os.Args) < 2 {
		envPath = `./.env`
	} else {
		envPath = os.Args[1]
	}

	return config.Load(envPath)
}

func exit(err error) {
	log.Printf("[EXIT] Progarm exited due to: %s", err)
	os.Exit(1)
}
