package main

import (
	"fmt"
	"log"
	"net/http"

	"gateway/internal/server"
	// "gateway/internal/codec"
	// "gateway/internal/dash"
	// "gateway/internal/storage"
)

func main() {
	// Set up HTTP handlers
	server.SetupRoutes()

	// Start the server
	fmt.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
