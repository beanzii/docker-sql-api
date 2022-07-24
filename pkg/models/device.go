package models

import (
	"beanzii/docker-sql-api/pkg/common/database"
	"encoding/json"
	"log"
	"net/http"
)

type Device struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Model string `json:"model"`
}

func (d *Device) Load(r *http.Request) error {
	return json.NewDecoder(r.Body).Decode(&d)
}

func (d *Device) Put() (int, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		log.Printf("[ERROR] Unable to begin d.Post() database transaction: %s\n", err)
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("INSERT INTO `devices` (`id`, `name`) VALUES (?, ?) ON DUPLICATE KEY UPDATE name = ?;")
	if err != nil {
		log.Printf("[ERROR] Unable to prepare d.Post() database statement: %s\n", err)
	}
	defer stmt.Close()
	if _, err := stmt.Exec(d.Id, d.Name, d.Name); err != nil {
		log.Printf("[ERROR] Unable to execute d.Post() database statement: %s\n", err)
	}
	if err := tx.Commit(); err != nil {
		log.Printf("[ERROR] Unable to commit d.Post() database transaction: %s\n", err)
	}
	row := database.DB.QueryRow("Select LAST_INSERT_ID();")
	if err := row.Err(); err != nil {
		log.Printf("[ERROR] Unable to query d.Post() LAST_INSERT_ID(): %s\n", err)
	}
	var id int
	if err := row.Scan(&id); err != nil {
		log.Printf("[ERROR] Unable to scan rows for d.Post() id: %s\n", err)
	}
	return id, nil
}

func (d *Device) Delete() {

}
