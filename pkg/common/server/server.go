package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"beanzii/docker-sql-api/pkg/common/config"
	"beanzii/docker-sql-api/pkg/common/database"
	"beanzii/docker-sql-api/pkg/models"
)

var routes = []route{
	newRoute("GET", `\/`, getRoot, false),
	newRoute("GET", `\/api\/devices`, apiGetDevices, true),
	newRoute("PUT", `\/api\/devices`, apiPutDevices, true),
	newRoute("POST", `\/api\/users`, apiPostUsers, true),
}

var serverAddress = "0.0.0.0"

func newRoute(method, pattern string, handler http.HandlerFunc, auth bool) route {
	return route{method, regexp.MustCompile(`^` + pattern + `$`), handler, auth}
}

type route struct {
	method  string
	regex   *regexp.Regexp
	handler http.HandlerFunc
	auth    bool
}

func Start() error {
	// Populate the server struct with configuration from
	// the .env file
	srvCfg, err := config.Parse(config.Server{})
	if err != nil {
		log.Printf("[ERROR] Unable to parse server config to Start(): %s\n", err)
		return err
	}

	// Get port from the srvCfg struct
	port := fmt.Sprint(srvCfg.Port)

	// Define router
	router := http.HandlerFunc(Serve)

	// Wrap router in authentication middleware

	var authHandler http.Handler = router
	authHandler = authRequestHandler(authHandler)

	// Wrap authHandler in logging middleware

	var logHandler http.Handler = authHandler
	logHandler = logRequestHandler(logHandler)

	// Define Server
	s := &http.Server{
		Addr:         serverAddress + ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      logHandler,
	}
	log.Printf("[INFO] Listening on https://%s:%s\n", serverAddress, port)
	if err = s.ListenAndServe(); err != nil {
		log.Printf("[ERROR] s.ListenAndServe() has encountered an error: %s\n", err)
	}
	return err
}

func Serve(w http.ResponseWriter, r *http.Request) {
	var allow []string
	for _, route := range routes {
		matches := route.regex.FindString(r.URL.Path)
		if len(matches) > 0 {
			if r.Method != route.method {
				allow = append(allow, route.method)
				continue
			}
			route.handler(w, r)
			return
		}
	}
	if len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.NotFound(w, r)
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "root\n")
}

func apiGetDevices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	reqId := q["id"]
	if reqId != nil {
		row := database.DB.QueryRow("Select id,name from `devices` where id = ?;", reqId[0])
		if errRowErr := row.Err(); errRowErr != nil {
			log.Fatal(errRowErr)
		}
		var id int
		var name string
		if errRowScan := row.Scan(&id, &name); errRowScan != nil {
			if errRowScan == sql.ErrNoRows {
				fmt.Fprintf(w, "No device found with id: %s\n", reqId[0])
				return
			} else {
				log.Fatal(errRowScan)
			}
		}
		fmt.Fprintf(w, "Device Found:\n id: %d name: %s\n", id, name)
	} else {
		rows, err := database.DB.Query("SELECT id,name FROM `devices`;")
		defer rows.Close()
		if err != nil {
			panic(err.Error())
		}
		fmt.Fprintln(w, "Devices:")
		for rows.Next() {
			var (
				id   int
				name string
			)
			if err := rows.Scan(&id, &name); err != nil {
				log.Fatal(err)
			}
			fmt.Fprintf(w, "id: %d name: %s\n", id, name)
		}
		if err := rows.Err(); err != nil {
			log.Fatalf("[ERROR] Problem getting rows: %s\n", err)
		}
	}
}

func apiPutDevices(w http.ResponseWriter, r *http.Request) {
	// Set content-type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Initialise *models.Device{}
	d := &models.Device{Id: 0, Name: ""}

	// Load values into struct from *http.Request body
	if err := d.Load(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// Update device based on provided id,
	// if no id is supplied, create new device
	id, err := d.Put()
	if err != nil {
		http.NotFound(w, r)
	}

	if id == 0 {
		fmt.Fprintf(w, "Updated Device\n id: %d\n name: %s\n", d.Id, d.Name)
		w.WriteHeader(http.StatusOK)
		return
	}

	fmt.Fprintf(w, "Created Device\n id: %d\n name: %s\n", id, d.Name)
	w.WriteHeader(http.StatusCreated)
}

type httpLog struct {
	StatusCode int
	Method     string
	RemoteAddr string
	URL        string
	Username   string
}
type StatusRecorder struct {
	Status int
	http.ResponseWriter
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		recorder := &StatusRecorder{
			Status:         200,
			ResponseWriter: w,
		}
		h.ServeHTTP(recorder, r)
		username, _, _ := r.BasicAuth()
		ri := httpLog{
			StatusCode: recorder.Status,
			Method:     r.Method,
			URL:        r.URL.String(),
			RemoteAddr: r.RemoteAddr,
			Username:   username,
		}
		logHTTPReq(ri)
	}
	return http.HandlerFunc(fn)
}

func logHTTPReq(ri httpLog) {
	var level string
	switch {
	case ri.StatusCode < 200:
		level = "INFO"
	case ri.StatusCode < 300:
		level = "INFO"
	case ri.StatusCode < 400:
		level = "INFO"
	default:
		level = "ERROR"
	}
	ip := strings.Split(ri.RemoteAddr, ":")
	log.Printf("[%s] Method: %s Status: %d Destination: %s Source: %s Username: %s\n", level, ri.Method, ri.StatusCode, ri.URL, ip[0], ri.Username)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func authRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		for _, route := range routes {
			matches := route.regex.FindString(r.URL.Path)
			if len(matches) > 0 {
				if route.auth == false {
					h.ServeHTTP(w, r)
					return
				}
			}
		}
		log.Printf("[INFO] Authentication required for URL: %s\n", r.URL.Path)
		username, password, ok := r.BasicAuth()
		if ok {
			log.Printf("[INFO] Confirming existence user: %s\n", username)
			userRow := database.DB.QueryRow("Select username from `users` where username = ?;", username)
			if err := userRow.Err(); err != nil {
				log.Fatalln(err)
			}
			var sqlUsername string
			if err := userRow.Scan(&sqlUsername); err != nil {
				if err == sql.ErrNoRows {
					log.Printf("[INFO] User: %s was not found\n", username)
					w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				} else {
					log.Fatalln(err)
				}
			}
			log.Printf("[INFO] Getting hashed password for user: %s\n", username)
			passRow := database.DB.QueryRow("Select password from `users` where username = ?;", username)
			if err := passRow.Err(); err != nil {
				log.Fatalln(err)
			}
			var sqlPassword string
			if err := passRow.Scan(&sqlPassword); err != nil {
				if err == sql.ErrNoRows {
					log.Printf("[INFO] Password for user: %s was not found\n", username)
					w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				} else {
					log.Fatalln(err)
				}
			}
			log.Printf("[INFO] Checking password for user: %s\n", username)
			if CheckPasswordHash(password, sqlPassword) == true {
				log.Printf("[INFO] Successful login for user: %s\n", username)
				h.ServeHTTP(w, r)
				return
			}
			log.Printf("[INFO] Unsuccesful login for user: %s\n", username)
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
	return http.HandlerFunc(fn)
}

func apiPostUsers(w http.ResponseWriter, r *http.Request) {
	tx, errTx := database.DB.Begin()
	if errTx != nil {
		log.Fatal(errTx)
	}
	defer tx.Rollback()
	stmt, errStmt := tx.Prepare("INSERT INTO `users` (`id`, `username`, `password`) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE username = ?, password = ?;")
	if errStmt != nil {
		log.Fatal(errStmt)
	}
	defer stmt.Close()
	var u models.User
	errJson := json.NewDecoder(r.Body).Decode(&u)
	if errJson != nil {
		http.Error(w, errJson.Error(), http.StatusBadRequest)
		return
	}
	pwHash, errHashPassword := hashPassword(u.Password)
	if errHashPassword != nil {
		log.Fatal(errHashPassword)
	}
	if _, errStmtExec := stmt.Exec(u.Id, u.Username, pwHash, u.Username, pwHash); errStmtExec != nil {
		log.Fatal(errStmtExec)
	}
	if errTxCommit := tx.Commit(); errTxCommit != nil {
		log.Fatal(errTxCommit)
	}
	row := database.DB.QueryRow("Select LAST_INSERT_ID();")
	if errRowErr := row.Err(); errRowErr != nil {
		log.Fatal(errRowErr)
	}
	var id int
	if errRowScan := row.Scan(&id); errRowScan != nil {
		log.Fatal(errRowScan)
	}
	if id == 0 {
		id = u.Id
		fmt.Fprintf(w, "Updated User\n id: %d\n username: %s\n", id, u.Username)
		return
	}
	fmt.Fprintf(w, "Created User\n id: %d\n username: %s\n", id, u.Username)
}
