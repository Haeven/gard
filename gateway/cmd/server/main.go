// main.go

package main

import (
	"fmt"
	"log"
)

func main() {
	// Define the Ozone URL (replace with the actual URL for your setup)
	ozoneURL := "http://ozone-container:9878"

	// Upload a .webm video file
	err := uploadFile("path/to/your/video.webm", ozoneURL)
	if err != nil {
		log.Fatalf("Error uploading file: %v", err)
	}
	fmt.Println("File uploaded successfully.")

	// Download a .webm video file
	data, err := downloadFile("your-video-key.webm", ozoneURL)
	if err != nil {
		log.Fatalf("Error downloading file: %v", err)
	}
	fmt.Println("File downloaded successfully, size:", len(data))
}
