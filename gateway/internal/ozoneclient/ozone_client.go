// internal/ozoneclient/ozone_client"
package ozone_client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

// UploadFile uploads a file to Apache Ozone
func UploadFile(filename string, ozoneURL string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	writer.Close()

	request, err := http.NewRequest("PUT", ozoneURL+"/webhdfs/v1/ozone-path?op=CREATE", &requestBody)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

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

// DownloadFile downloads a file from Apache Ozone
func DownloadFile(objectKey string, ozoneURL string) ([]byte, error) {
	resp, err := http.Get(ozoneURL + "/webhdfs/v1/ozone-path/" + objectKey + "?op=OPEN")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file, status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// DeleteFile deletes a file from Apache Ozone
func DeleteFile(objectKey string, ozoneURL string) error {
	request, err := http.NewRequest("DELETE", ozoneURL+"/webhdfs/v1/ozone-path/"+objectKey+"?op=DELETE", nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete file, status: %s", resp.Status)
	}

	return nil
}
