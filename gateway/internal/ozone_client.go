// ozone_client.go

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

// Uploads a file to Apache Ozone
func uploadFile(filename string, ozoneURL string) error {
	// Open the file for reading
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Prepare a form that you will submit to Ozone
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return err
	}

	// Write the data to the form
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	writer.Close()

	// Create a new HTTP request
	request, err := http.NewRequest("PUT", ozoneURL+"/webhdfs/v1/ozone-path?op=CREATE", &requestBody)
	if err != nil {
		return err
	}

	// Set the Content-Type header
	request.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to upload file, status: %s", resp.Status)
	}

	return nil
}

// Downloads a file from Apache Ozone
func downloadFile(objectKey string, ozoneURL string) ([]byte, error) {
	resp, err := http.Get(ozoneURL + "/webhdfs/v1/ozone-path/" + objectKey + "?op=OPEN")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file, status: %s", resp.Status)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
