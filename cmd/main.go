// gard/cmd/main.go

package main

import (
	"fmt"
	"gard/pkg/server"
	"log"
	"net/http"
	// "gard/pkg/internal/codec"
	// "gard/pkg/internal/dash"
	// "gard/pkg/internal/storage"
)

func main() {
	// Set up HTTP handlers
	server.SetupRoutes()

	// Start the server
	fmt.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
