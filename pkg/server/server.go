// pkg/internal/server/server.go

package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	datalake "gard/pkg/internal/datalake"
)

var client *datalake.SeaweedFSClient

func init() {
	client = datalake.NewSeaweedFSClient()
}

// SetupRoutes sets up the HTTP routes for the server
func SetupRoutes() {
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/download", DownloadHandler)
	http.HandleFunc("/delete", DeleteHandler)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 32 MB is the default max memory for ParseMultipartForm
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Use the original filename
	tempFile, err := os.CreateTemp("", "upload-*"+filepath.Ext(header.Filename))
	if err != nil {
		http.Error(w, "Failed to create temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	// Save file to temporary location
	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Rewind the file for reading
	if _, err := tempFile.Seek(0, 0); err != nil {
		http.Error(w, "Failed to reset file for reading: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fileID, mpdFileID, err := client.UploadFile(tempFile.Name())
	if err != nil {
		http.Error(w, "Failed to upload file to SeaweedFS: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully. File ID: %s, MPD File ID: %s", fileID, mpdFileID)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "Missing file ID", http.StatusBadRequest)
		return
	}

	// Get the MPD content
	mpd, err := client.GetMPDContent(fileID) //TODO change GetMPDContent fileid argument to take videofile then generate the mpd file name ([videofile].mpd)
	if err != nil {
		log.Fatalf("Failed to get MPD content: %v", err)
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
	resolution := "720p"
	videoData, err := client.GetRepresentation(fileID, resolution)
	if err != nil {
		log.Fatalf("Failed to get representation: %v", err)
	}
	if err != nil {
		http.Error(w, "Failed to download file from SeaweedFS", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/webm")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.webm", fileID))
	w.WriteHeader(http.StatusOK)
	w.Write(videoData)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "Missing file ID", http.StatusBadRequest)
		return
	}

	err := client.DeleteFile(fileID)
	if err != nil {
		http.Error(w, "Failed to delete file from SeaweedFS", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File deleted successfully.")
}
