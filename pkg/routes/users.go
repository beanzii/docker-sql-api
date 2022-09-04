package routes

import (
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"beanzii/docker-sql-api/pkg/database"
	"beanzii/docker-sql-api/pkg/models"
)

func Users(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] Routing /api/users")
	switch r.Method {
	case "PUT":
		putApiUser(w, r)
	case "DELETE":
		deleteApiUser(w, r)
	default:
		w.Header().Add("Allowed", "PUT,DELETE")
		w.WriteHeader(405)
	}
}

func putApiUser(w http.ResponseWriter, r *http.Request) {
	// Set content-type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Initialise *models.User{}
	db, err := database.Open()
	if err != nil {
		log.Printf("[ERROR] Unable to open database for apiPutUsers(): %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u := models.User{}
	defer db.Close()
	// Load values into struct from *http.Request body
	if err := u.Load(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Printf("[ERROR] Unable to begin database transaction for apiPutUsers(): %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("INSERT INTO `users` (`id`, `username`, `password`) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE username = ?, password = ?;")
	if err != nil {
		log.Printf("[ERROR] Unable to prepare transaction for apiPutUsers(): %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer stmt.Close()
	pwHash, err := hashPassword(u.Password)
	if err != nil {
		log.Printf("[ERROR] Unable to hash password for for apiPutUsers(): %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := stmt.Exec(u.Id, u.Username, pwHash, u.Username, pwHash); err != nil {
		log.Printf("[ERROR] Unable to execute statement for apiPutUsers(): %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := tx.Commit(); err != nil {
		log.Printf("[ERROR] Unable to commit transaction for apiPutUsers(): %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	row := db.QueryRow("Select LAST_INSERT_ID();")
	if err := row.Err(); err != nil {
		log.Printf("[ERROR] Unable to query last insert id for apiPutUsers(): %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var id int
	if err := row.Scan(&id); err != nil {
		log.Printf("[ERROR] Unable to scan id for apiPutUsers(): %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if u.Id != 0 {
		return
	}

}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func deleteApiUser(w http.ResponseWriter, r *http.Request) {

}
