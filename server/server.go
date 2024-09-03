package server

import (
	"context"
	"log"

	datalake "gard/pkg/internal/datalake"

	"github.com/segmentio/kafka-go"
)

var client *datalake.SeaweedFSClient

func init() {
	client = datalake.NewSeaweedFSClient()
}

func main() {
	// Create a new Kafka reader for each event type
	uploadReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "upload",
		GroupID: "group-id",
	})
	downloadReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "download",
		GroupID: "group-id",
	})
	deleteReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "delete",
		GroupID: "group-id",
	})

	// Start consumers
	go consumeUploadEvents(uploadReader)
	go consumeDownloadEvents(downloadReader)
	go consumeDeleteEvents(deleteReader)

	// Block forever
	select {}
}

func consumeUploadEvents(r *kafka.Reader) {
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("Failed to read message: %v", err)
		}

		// Process upload event
		handleUploadEvent(m.Value)
	}
}

func handleUploadEvent(data []byte) {
	// Assume data contains the file content and filename
	// Parse data to get file content and filename
	// ...

	fileID, err := client.UploadFile(filename, data)
	if err != nil {
		log.Printf("Failed to upload file to SeaweedFS: %v", err)
		return
	}

	log.Printf("File uploaded successfully. File ID: %s", fileID)
}

func consumeDownloadEvents(r *kafka.Reader) {
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("Failed to read message: %v", err)
		}

		// Process download event
		handleDownloadEvent(m.Value)
	}
}

func handleDownloadEvent(data []byte) {
	// Assume data contains the file ID
	// Parse data to get file ID
	// ...

	fileData, err := client.DownloadFile(fileID)
	if err != nil {
		log.Printf("Failed to download file: %v", err)
		return
	}

	// Process or save fileData as needed
	// ...
}

func consumeDeleteEvents(r *kafka.Reader) {
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("Failed to read message: %v", err)
		}

		// Process delete event
		handleDeleteEvent(m.Value)
	}
}

func handleDeleteEvent(data []byte) {
	// Assume data contains the file ID
	// Parse data to get file ID
	// ...

	err := client.DeleteFile(fileID)
	if err != nil {
		log.Printf("Failed to delete file from SeaweedFS: %v", err)
		return
	}

	log.Printf("File deleted successfully.")
}
