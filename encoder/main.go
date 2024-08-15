package main

// TODO Add production logging; OPTIONAL: Analytics
//
import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	// "io"
	// "net/http"
	// "os"
	// "path/filepath"
)

func main() {
	// Create the uploads directory if it doesn't exist
	// os.MkdirAll("uploads", os.ModePerm)

	// Handle file uploads
	http.HandleFunc("/upload", uploadHandler)

	// Start the server
	output, err := exec.Command("./engine/target/release/rav1e", "input.y4m", "-o", "output.ivf").Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(output)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "File uploaded successfully:")
}

// func uploadHandler(w http.ResponseWriter, r *http.Request) {
// 	// Limit the size of the request body to 10 MB (10 * 1024 * 1024)
// 	r.ParseMultipartForm(10 << 20)

// 	// Retrieve the file from form data
// 	file, handler, err := r.FormFile("video")
// 	if err != nil {
// 		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
// 		log.Println("Error retrieving the file:", err)
// 		return
// 	}
// 	defer file.Close()

// 	// Create the file on the server
// 	dst, err := os.Create(filepath.Join("uploads", handler.Filename))
// 	if err != nil {
// 		http.Error(w, "Error saving the file", http.StatusInternalServerError)
// 		log.Println("Error saving the file:", err)
// 		return
// 	}
// 	defer dst.Close()

// 	// Copy the uploaded file to the destination file
// 	if _, err := io.Copy(dst, file); err != nil {
// 		http.Error(w, "Error saving the file", http.StatusInternalServerError)
// 		log.Println("Error copying the file:", err)
// 		return
// 	}

// 	// Respond to the client

// }

// func main() {

// }
