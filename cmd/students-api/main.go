package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bikipramanik/students-api/internal/config"
)

func main() {
	//load config

	cfg := config.MustLoad()

	//database setup
	//setup routes
	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Studetnts api "))
	})
	//setup server

	server := http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: router,
	}

	fmt.Print("Server started")
	err := server.ListenAndServe()

	if err != nil {
		log.Fatal("Failed to load Server")
	}

}
