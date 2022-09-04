package database

import (
	"beanzii/docker-sql-api/pkg/config"
	"database/sql"
	"fmt"
	"log"

	"github.com/go-sql-driver/mysql"
)

func Open() (*sql.DB, error) {
	var err error
	// Pull database environment information from .env
	dbcfg, err := config.Parse(config.Database{})
	if err != nil {
		log.Printf("[ERROR] Unable to parse database config to Open(): %s\n", err)
		return nil, err
	}
	// Connect to MySQL server
	dsn := mysql.NewConfig()
	dsn.User = dbcfg.User
	dsn.Passwd = dbcfg.Password
	dsn.Addr = fmt.Sprintf("%s:%d", dbcfg.Host, dbcfg.Port)
	dsn.DBName = dbcfg.Name
	dsn.Net = dbcfg.ConnectionType
	dsnString := dsn.FormatDSN()
	db, err := sql.Open(dbcfg.Driver, dsnString)

	if err != nil {
		log.Printf("[ERROR] Unable to Open() MySQL Connection to database: %s\n", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Printf("[ERROR] Unable to Ping() MySQL: %s\n", err)
		return nil, err
	}

	log.Printf("[INFO] Connected to MySQL: %s/%s\n", dbcfg.Host, dbcfg.Name)
	return db, nil
}
