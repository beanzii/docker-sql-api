package middlewares

import (
	"database/sql"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"beanzii/docker-sql-api/pkg/database"
	"beanzii/docker-sql-api/pkg/models"
)

func AuthenicateRequest(w http.ResponseWriter, r *http.Request, next func(http.ResponseWriter, *http.Request)) {
	log.Printf("[INFO] Authentication required for URL: %s\n", r.URL.Path)
	username, password, ok := r.BasicAuth()
	if ok {
		log.Printf("[INFO] Confirming existence of user: %s\n", username)
		db, err := database.Open()
		if err != nil {
			log.Printf("[ERROR] Unable to open database: %s\n", err)
			authUnauthorised(w)
		}
		defer db.Close()
		userRow := db.QueryRow("Select username,password from `users` where username = ?;", username)
		if err := userRow.Err(); err != nil {
			log.Printf("[ERROR] Unable to query userRow: %s\n", err)
			authUnauthorised(w)
		}
		u := models.User{}
		if err := userRow.Scan(&u.Username, &u.Password); err != nil {
			if err == sql.ErrNoRows {
				log.Printf("[INFO] Unable to find user: %s\n", username)
				authUnauthorised(w)
			} else {
				log.Printf("[ERROR] Unable to scan userRow: %s\n", err)
				authUnauthorised(w)
			}
		}
		log.Printf("[INFO] Checking password for user: %s\n", username)
		if checkPasswordHash(password, u.Password) {
			log.Printf("[INFO] Successful login for user: %s\n", username)
			next(w, r)
			return
		}
		log.Printf("[INFO] Unsuccesful login for user: %s\n", username)
	}
	authUnauthorised(w)
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func authUnauthorised(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
}
