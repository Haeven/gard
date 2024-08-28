// pkg/internal/datalake/datalake.go
package datalake

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	codec "gard/pkg/internal/codec"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// SeaweedFSClient represents a client for interacting with SeaweedFS
type SeaweedFSClient struct {
	MasterURL string
	VolumeURL string
}

// NewSeaweedFSClient creates a new SeaweedFSClient
func NewSeaweedFSClient() *SeaweedFSClient {
	return &SeaweedFSClient{
		MasterURL: os.Getenv("SEAWEEDFS_MASTER"),
		VolumeURL: os.Getenv("SEAWEEDFS_VOLUME"),
	}
}

// UploadFile uploads a file to SeaweedFS
func (c *SeaweedFSClient) UploadFile(videoFile string) (map[string]string, string, error) {
	// Step 1: Generate segments and MPD file
	outputDir := "output_segments"
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return nil, "", fmt.Errorf("error creating output directory: %w", err)
	}

	err = codec.GenerateSegments(videoFile, outputDir)
	if err != nil {
		return nil, "", fmt.Errorf("error generating segments: %w", err)
	}

	err = codec.GenerateMPD(outputDir)
	if err != nil {
		return nil, "", fmt.Errorf("error generating MPD file: %w", err)
	}

	// Step 2: Upload the original file
	originalFileID, err := c.uploadFile(videoFile)
	if err != nil {
		return nil, "", fmt.Errorf("error uploading original file: %w", err)
	}

	// Step 3: Upload segments and MPD file
	files := []string{"output.mpd"}
	segments, err := filepath.Glob(filepath.Join(outputDir, fmt.Sprintf("%s_segment_*.mp4", videoFile)))
	if err != nil {
		return nil, "", fmt.Errorf("error listing segment files: %w", err)
	}
	files = append(files, segments...)

	fileIDs := make(map[string]string)
	fileIDs["original"] = originalFileID
	var mpdFileID string
	for _, file := range files {
		fid, err := c.uploadFile(file)
		if err != nil {
			return nil, "", fmt.Errorf("error uploading file %s: %w", file, err)
		}
		if filepath.Ext(file) == ".mpd" {
			mpdFileID = fid
		} else {
			fileIDs[filepath.Base(file)] = fid
		}
	}

	return fileIDs, mpdFileID, nil
}
func (c *SeaweedFSClient) DownloadFile(fileID string) ([]byte, error) {
	// First, check if the requested file is the MPD
	if filepath.Ext(fileID) == ".mpd" {
		return c.downloadSingleFile(fileID)
	}

	// If it's not the MPD, assume it's a segment file
	return c.downloadSingleFile(fileID)
}

func (c *SeaweedFSClient) downloadSingleFile(fileID string) ([]byte, error) {
	downloadURL := fmt.Sprintf("http://%s/%s", c.VolumeURL, fileID)
	fmt.Printf("Attempting to download file from: %s\n", downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to download file, status: %s, body: %s", resp.Status, string(body))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	fmt.Printf("Successfully downloaded %d bytes\n", len(data))
	return data, nil
}

// DeleteFile deletes a file from SeaweedFS using a file ID
func (c *SeaweedFSClient) DeleteFile(fileID string) error {
	deleteURL := fmt.Sprintf("http://%s/%s", c.VolumeURL, fileID)
	request, err := http.NewRequest("DELETE", deleteURL, nil)
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

// uploadFile uploads a single file to SeaweedFS and returns the file ID
func (c *SeaweedFSClient) uploadFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	uploadURL, err := c.getUploadURL()
	if err != nil {
		return "", err
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}
	writer.Close()

	request, err := http.NewRequest("POST", uploadURL, &requestBody)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload file, status: %s, body: %s", resp.Status, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	fid, ok := result["name"]
	if !ok {
		return "", fmt.Errorf("fid not found in response: %+v", result)
	}

	fidStr, ok := fid.(string)
	if !ok {
		return "", fmt.Errorf("fid is not a string: %v", fid)
	}

	return fidStr, nil
}

// getUploadURL gets an upload URL from the SeaweedFS master
func (c *SeaweedFSClient) getUploadURL() (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/dir/assign", c.MasterURL))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get upload URL, status: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s/%s", result["publicUrl"], result["fid"]), nil
}

func (c *SeaweedFSClient) GetMPDContent(mpdFileID string) (*MPD, error) {
	mpdData, err := c.downloadSingleFile(mpdFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to download MPD file: %v", err)
	}

	var mpd MPD
	err = xml.Unmarshal(mpdData, &mpd)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MPD file: %v", err)
	}

	return &mpd, nil
}

func (c *SeaweedFSClient) GetRepresentation(mpdFileID, resolution string) ([]byte, error) {
	mpd, err := c.GetMPDContent(mpdFileID)
	if err != nil {
		return nil, err
	}

	// Find the requested representation
	var targetRep *Representation
	for _, period := range mpd.Periods {
		for _, adaptationSet := range period.AdaptationSets {
			for _, rep := range adaptationSet.Representations {
				if rep.ID == resolution {
					targetRep = &rep
					break
				}
			}
			if targetRep != nil {
				break
			}
		}
		if targetRep != nil {
			break
		}
	}

	if targetRep == nil {
		return nil, fmt.Errorf("representation not found for resolution: %s", resolution)
	}

	// Download and concatenate all segments for the representation
	var fullVideo []byte

	for _, segmentURL := range targetRep.SegmentList.SegmentURLs {
		segmentData, err := c.downloadSingleFile(segmentURL.Media)
		if err != nil {
			return nil, fmt.Errorf("failed to download segment %s: %v", segmentURL.Media, err)
		}
		fullVideo = append(fullVideo, segmentData...)
	}

	return fullVideo, nil
}

// SegmentURL represents the URL of a segment in the MPD file
type SegmentURL struct {
	Media string `xml:"media,attr"`
}

// MPD represents the root element of the MPD file
type MPD struct {
	XMLName                   xml.Name `xml:"MPD"`
	XMLNs                     string   `xml:"xmlns,attr"`
	XMLNsXsi                  string   `xml:"xmlns:xsi,attr"`
	XsiSchemaLocation         string   `xml:"xsi:schemaLocation,attr"`
	MediaPresentationDuration string   `xml:"mediaPresentationDuration,attr"`
	MinBufferTime             string   `xml:"minBufferTime,attr"`
	Periods                   []Period `xml:"Period"`
}

// Period represents a period in the MPD file
type Period struct {
	Duration       string          `xml:"duration,attr"`
	AdaptationSets []AdaptationSet `xml:"AdaptationSet"`
}

// AdaptationSet represents an adaptation set in the MPD file
type AdaptationSet struct {
	MimeType        string           `xml:"mimeType,attr"`
	Codecs          string           `xml:"codecs,attr"`
	Representations []Representation `xml:"Representation"`
}

// Representation represents a representation in the MPD file
type Representation struct {
	ID          string      `xml:"id,attr"`
	Bandwidth   string      `xml:"bandwidth,attr"`
	Codecs      string      `xml:"codecs,attr"`
	Width       string      `xml:"width,attr"`
	Height      string      `xml:"height,attr"`
	FrameRate   string      `xml:"frameRate,attr"`
	BaseURL     string      `xml:"BaseURL"`
	SegmentList SegmentList `xml:"SegmentList"`
}

// SegmentList represents the segment list in the MPD file
type SegmentList struct {
	Duration    string       `xml:"duration,attr"`
	Timescale   string       `xml:"timescale,attr"`
	SegmentURLs []SegmentURL `xml:"SegmentURL"`
}
