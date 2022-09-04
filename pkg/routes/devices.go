package routes

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"beanzii/docker-sql-api/pkg/database"
	"beanzii/docker-sql-api/pkg/models"
)

func Devices(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Routing /api/devices")
	switch r.Method {
	case "GET", "":
		getApiDevice(w, r)
	case "PUT":
		putApiDevice(w, r)
	case "POST":
		postApiDevice(w, r)
	case "DELETE":
		deleteApiDevice(w, r)
	default:
		w.Header().Add("Allowed", "GET,PUT,POST,DELETE")
		w.WriteHeader(405)
	}
}

func getApiDevice(w http.ResponseWriter, r *http.Request) {
	// Set content-type to application/json
	w.Header().Set("Content-Type", "application/json")

	db, err := database.Open()
	if err != nil {
		log.Printf("[ERROR] Unable to open database: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	q := r.URL.Query()
	if reqId := q["id"]; reqId != nil {
		row := db.QueryRow("Select id,name,model from `devices` where id = ?;", reqId[0])
		if err := row.Err(); err != nil {
			log.Printf("[ERROR] Unable to query rows for getApiDevice(): %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		d := models.Device{}
		if err := row.Scan(&d.Id, &d.Name, &d.Model); err != nil {
			if err == sql.ErrNoRows {
				_ = json.NewEncoder(w).Encode("")
				return
			} else {
				log.Printf("[ERROR] Unable to scan rows for getApiDevice(): %s\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		_ = json.NewEncoder(w).Encode(d)
	}
	rows, err := db.Query("SELECT id,name,model FROM `devices`;")
	if err != nil {
		log.Printf("[ERROR] Unable to query rows for getApiDevice(): %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	d := models.Device{}
	var ds []models.Device
	for rows.Next() {
		if err := rows.Scan(&d.Id, &d.Name, &d.Model); err != nil {
			log.Printf("[ERROR] Unable to scan rows for getApiDevice(): %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ds = append(ds, d)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Unable to scan all rows for getApiDevice(): %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(ds)
}

func putApiDevice(w http.ResponseWriter, r *http.Request) {

	// Initialise *models.Device{}
	db, err := database.Open()
	if err != nil {
		log.Printf("[ERROR] Unable to open database: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	d := models.Device{}
	defer db.Close()
	// Load values into struct from *http.Request body
	if err := d.Load(r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update device based on provided id,
	// if no id is supplied, create new device
	tx, err := db.Begin()
	if err != nil {
		log.Printf("[ERROR] Unable to begin apiPutDevice() database transaction: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("INSERT INTO `devices` (`id`, `name`, `model`) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE name = values(name), model = values(model);")
	if err != nil {
		log.Printf("[ERROR] Unable to prepare apiPutDevice() database statement: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()
	if _, err := stmt.Exec(d.Id, d.Name, d.Model); err != nil {
		log.Printf("[ERROR] Unable to execute apiPutDevice() database statement: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		log.Printf("[ERROR] Unable to commit apiPutDevice() database transaction: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if d.Id != 0 {
		_ = json.NewEncoder(w).Encode(d)
		return
	}
	var id int
	row := db.QueryRow("select id from devices where name = ?;", d.Name)
	if err := row.Scan(&id); err != nil {
		log.Printf("[ERROR] Unable to query databse for new id")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	d.Id = id
	_ = json.NewEncoder(w).Encode(d)
}

func postApiDevice(w http.ResponseWriter, r *http.Request) {

}

func deleteApiDevice(w http.ResponseWriter, r *http.Request) {

}
