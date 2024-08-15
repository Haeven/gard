package main

// TODO Add production logging; OPTIONAL: Analytics
//
import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"

	// "io"
	// "net/http"
	"os"
	// "path/filepath"
)

func main() {
	// Create the uploads directory if it doesn't exist
	// os.MkdirAll("uploads", os.ModePerm)

	// Handle file uploads
	http.HandleFunc("/upload", uploadHandler)

	// Start the server

	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	log.Println("Starting server on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))

}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	filePath := "uploads/input.y4m"

	// Create the uploads directory if it doesn't exist
	os.MkdirAll("uploads", os.ModePerm)
	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Unable to create the file", http.StatusInternalServerError)
		log.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Copy the binary data from the request body to the file
	_, err = io.Copy(file, r.Body)
	if err != nil {
		http.Error(w, "Unable to save the file", http.StatusInternalServerError)
		log.Println("Error saving file:", err)
		return
	}
	files, err := ioutil.ReadDir("/uploads")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Println(f.Name())
	}
	output, err := exec.Command("./usr/local/bin/encoder/target/release/rav1e", "uploads/input.y4m", "-o", "output.ivf").Output()
	fmt.Println(output)
	// file, handler, err := r.FormFile("video")
	// if err != nil {
	// 	http.Error(w, "Error retrieving the file", http.StatusBadRequest)
	// 	log.Println("Error retrieving the file:", err)
	// 	return
	// }
	// defer file.Close()
	// fmt.Println(handler)
	fmt.Fprintf(w, "File uploaded successfully:")
}

// func uploadHandler(w http.ResponseWriter, r *http.Request) {
// 	// Limit the size of the request body to 10 MB (10 * 1024 * 1024)
// 	r.ParseMultipartForm(10 << 20)

// 	// Retrieve the file from form data

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
