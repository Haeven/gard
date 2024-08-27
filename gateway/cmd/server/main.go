package server

import (
	"fmt"
	"log"
	"net/http"

	"gard-gateway/internal/server"
)

func main() {
	// Set up HTTP handlers
	server.SetupRoutes()

	// Start the server
	fmt.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
