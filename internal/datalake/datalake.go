package datalake

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
func (c *SeaweedFSClient) UploadFile(videoFile string) (map[string]string, error) {
	// Step 1: Generate segments and MPD file
	outputDir := "output_segments"
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("error creating output directory: %w", err)
	}

	err = generateSegments(videoFile, outputDir)
	if err != nil {
		return nil, fmt.Errorf("error generating segments: %w", err)
	}

	err = generateMPD(outputDir)
	if err != nil {
		return nil, fmt.Errorf("error generating MPD file: %w", err)
	}

	// Step 2: Upload the original file
	originalFileID, err := c.uploadFile(videoFile)
	if err != nil {
		return nil, fmt.Errorf("error uploading original file: %w", err)
	}

	// Step 3: Upload segments and MPD file
	files := []string{"output.mpd"}
	segments, err := filepath.Glob(filepath.Join(outputDir, "segment_*.mp4"))
	if err != nil {
		return nil, fmt.Errorf("error listing segment files: %w", err)
	}
	files = append(files, segments...)

	fileIDs := make(map[string]string)
	fileIDs["original"] = originalFileID
	for _, file := range files {
		fid, err := c.uploadFile(file)
		if err != nil {
			return nil, fmt.Errorf("error uploading file %s: %w", file, err)
		}
		fileIDs[filepath.Base(file)] = fid
	}

	return fileIDs, nil
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

// generateSegments generates video segments using ffmpeg
func generateSegments(videoFile, outputDir string) error {
	cmd := exec.Command("ffmpeg", "-i", videoFile, "-c", "copy", "-map", "0", "-f", "segment", "-segment_time", "10", "-segment_format", "mp4", filepath.Join(outputDir, "segment_%03d.mp4"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// generateMPD generates an MPD file for the video segments
func generateMPD(outputDir string) error {
	segmentFiles, err := filepath.Glob(filepath.Join(outputDir, "segment_*.mp4"))
	if err != nil {
		return fmt.Errorf("error reading segment files: %w", err)
	}

	var segmentURLs []SegmentURL
	for _, file := range segmentFiles {
		filename := filepath.Base(file)
		segmentURLs = append(segmentURLs, SegmentURL{Media: filename})
	}

	mpd := MPD{
		XMLNs:                     "urn:mpeg:dash:schema:mpd:2011",
		XMLNsXsi:                  "http://www.w3.org/2001/XMLSchema-instance",
		XsiSchemaLocation:         "urn:mpeg:dash:schema:mpd:2011 http://www.mpegdash.org/schemas/2011/MPD.xsd",
		MediaPresentationDuration: "PT" + strconv.Itoa(len(segmentURLs)*10) + "S", // Example duration, adjust as needed
		MinBufferTime:             "PT1.5S",
		Periods: []Period{
			{
				Duration: "PT" + strconv.Itoa(len(segmentURLs)*10) + "S", // Example duration, adjust as needed
				AdaptationSets: []AdaptationSet{
					{
						MimeType: "video/mp4",
						Codecs:   "vp09.00.10.08",
						Representations: []Representation{
							createRepresentation("144p", "1 Mbps", "150000", "144", "256", "25", "https://example.com/video_144p/", segmentURLs),
							// Add more representations for other resolutions as needed
						},
					},
				},
			},
		},
	}

	file, err := os.Create(filepath.Join(outputDir, "output.mpd"))
	if err != nil {
		return fmt.Errorf("error creating MPD file: %w", err)
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("  ", "    ")
	return encoder.Encode(mpd)
}

// createRepresentation creates a Representation for the MPD file
func createRepresentation(resolution, bitrate, bandwidth, width, height, frameRate, baseURL string, segmentURLs []SegmentURL) Representation {
	return Representation{
		ID:        resolution,
		Bandwidth: bandwidth,
		Codecs:    "vp09.00.10.08",
		Width:     width,
		Height:    height,
		FrameRate: frameRate,
		BaseURL:   baseURL,
		SegmentList: SegmentList{
			Duration:    "2000000",
			Timescale:   "1000000",
			SegmentURLs: segmentURLs,
		},
	}
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
