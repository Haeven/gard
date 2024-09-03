package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/segmentio/kafka-go"
	datalake "gard/pkg/internal/datalake"
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

	// Use the original filename
	tempFile, err := os.CreateTemp("", "upload-*"+filepath.Ext(filename))
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		return
	}
	defer os.Remove(tempFile.Name())

	// Save file to temporary location
	_, err = io.Copy(tempFile, bytes.NewReader(data))
	if err != nil {
		log.Printf("Failed to save file: %v", err)
		return
	}

	// Rewind the file for reading
	if _, err := tempFile.Seek(0, 0); err != nil {
		log.Printf("Failed to reset file for reading: %v", err)
		return
	}

	fileID, mpdFileID, err := client.UploadFile(tempFile.Name())
	if err != nil {
		log.Printf("Failed to upload file to SeaweedFS: %v", err)
		return
	}

	log.Printf("File uploaded successfully. File ID: %s, MPD File ID: %s", fileID, mpdFileID)
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
	// Assume data contains the file ID and resolution
	// Parse data to get file ID and resolution
	// ...

	// Get the MPD content
	mpd, err := client.GetMPDContent(fileID)
	if err != nil {
		log.Printf("Failed to get MPD content: %v", err)
		return
	}

	// Print available resolutions
	for _, period := range mpd.Periods {
		for _, adaptationSet := range period.AdaptationSets {
			for _, rep := range adaptationSet.Representations {
				fmt.Printf("Available resolution: %s\n", rep.ID)
			}
		}
	}

	// Download a specific representation
	videoData, err := client.GetRepresentation(fileID, resolution)
	if err != nil {
		log.Printf("Failed to get representation: %v", err)
		return
	}

	// Save or process videoData as needed
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