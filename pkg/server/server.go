package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"beanzii/docker-sql-api/pkg/config"
	"beanzii/docker-sql-api/pkg/handlers"
	"beanzii/docker-sql-api/pkg/middlewares"
	"beanzii/docker-sql-api/pkg/routes"
)

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
	// Define mux
	// New Multiplexer
	mux := new(handlers.MiddlewareMux)
	// Identify Requests
	mux.AppendMiddleware(middlewares.ReqIdRequest)
	// Auth Requests
	mux.AppendMiddleware(middlewares.AuthenicateRequest)
	// Log Response
	mux.AppendMiddleware(middlewares.LogRequest)
	// Serve pages
	mux.HandleFunc("/api/devices", routes.Devices)
	mux.HandleFunc("/api/users", routes.Users)

	// Define Server
	s := &http.Server{
		Addr:         srvCfg.Addr + ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}
	log.Printf("[INFO] Listening on http://%s:%s\n", srvCfg.Addr, port)
	if err = s.ListenAndServe(); err != nil {
		log.Printf("[ERROR] s.ListenAndServe() has encountered an error: %s\n", err)
	}
	return err
}
