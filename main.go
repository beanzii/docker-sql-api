package main

import (
	"beanzii/docker-sql-api/pkg/common/config"
	"beanzii/docker-sql-api/pkg/common/database"
	"beanzii/docker-sql-api/pkg/common/server"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// config.Load(envLocation)  using .env file path from os.Args[1]
	if err := loadDotEnv(); err != nil {
		panic(err)
	}

	// Open DB Connection
	if err := database.Open(); err != nil {
		panic(err)
	}

	// Start http server
	if err := server.Start(); err != nil {
		panic(err)
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
