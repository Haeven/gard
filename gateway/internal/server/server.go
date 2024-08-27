package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	ozone_client "gard-gateway/internal/ozoneclient"
)

// SetupRoutes sets up the HTTP routes for the server
func SetupRoutes() {
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/download", DownloadHandler)
	http.HandleFunc("/delete", DeleteHandler)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	ozoneURL := "http://ozone:9878" // Updated to use Docker network alias

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", "upload-*.webm")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	err = ozone_client.UploadFile(tempFile.Name(), ozoneURL)
	if err != nil {
		http.Error(w, "Failed to upload file to Ozone", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File uploaded successfully.")
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	ozoneURL := "http://ozone:9878" // Updated to use Docker network alias

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing file key", http.StatusBadRequest)
		return
	}

	data, err := ozone_client.DownloadFile(key, ozoneURL)
	if err != nil {
		http.Error(w, "Failed to download file from Ozone", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/webm")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(key)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ozoneURL := "http://ozone:9878" // Updated to use Docker network alias

	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing file key", http.StatusBadRequest)
		return
	}

	err := ozone_client.DeleteFile(key, ozoneURL)
	if err != nil {
		http.Error(w, "Failed to delete file from Ozone", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File deleted successfully.")
}
