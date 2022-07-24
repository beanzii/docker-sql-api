package database

import (
	"beanzii/docker-sql-api/pkg/common/config"
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Open() error {
	var err error
	// Pull database environment information from .env
	dbcfg, err := config.Parse(config.Database{})
	if err != nil {
		log.Printf("[ERROR] Unable to parse database config to Open(): %s\n", err)
		return err
	}
	// Connect to MySQL server
	DB, err = sql.Open(fmt.Sprintf("%s", dbcfg.Driver), fmt.Sprintf("%s:%s@tcp(%s:%s)/", dbcfg.User, dbcfg.Password, dbcfg.Host, strconv.Itoa(int(dbcfg.Port))))

	if err != nil {
		log.Printf("[ERROR] Unable to Open() MySQL Connection: %s\n", err)
		return err
	}

	if err = DB.Ping(); err != nil {
		log.Printf("[ERROR] Unable to Ping() MySQL: %s\n", err)
		return err
	}

	log.Printf("[INFO] Connected to MySQL: %s\n", dbcfg.Host)

	// Connect to database
	DB, err = sql.Open(fmt.Sprintf("%s", dbcfg.Driver), fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbcfg.User, dbcfg.Password, dbcfg.Host, strconv.Itoa(int(dbcfg.Port)), dbcfg.Name))
	if err != nil {
		log.Printf("[ERROR] Unable to open connection to database: %s\n", err)
		return err
	}
	log.Printf("[INFO] Connected to database: %s\n", dbcfg.Name)
	return nil
}
