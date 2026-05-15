package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Entry struct {
	Email        string
	FirstName    string
	LastName     string
	Area         string
	Group        string
	Function     string
	Gender       string
	Local        string
	District     string
	Status       string
	PreferredDay string
}

func main() {
	loadConfig()
	initDB()

	http.HandleFunc("/meet-greet/register", RegistrationHandler)
	http.HandleFunc("/internal/reports", ReportsHandler)
	http.HandleFunc("/status", StatusHandler)
	http.HandleFunc("/ping", PingHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	// FIX M-5: configure server timeouts to prevent Slowloris-style attacks
	// and resource exhaustion from slow or idle clients.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(srv.ListenAndServe())
}
