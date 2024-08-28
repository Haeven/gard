// pkg/internal/dash/dash.go
package dash

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const segmentDuration = 5 // Segment duration in seconds

type MPD struct {
	XMLName                   xml.Name `xml:"MPD"`
	XMLNs                     string   `xml:"xmlns,attr"`
	XMLNsXsi                  string   `xml:"xmlns:xsi,attr"`
	XsiSchemaLocation         string   `xml:"xsi:schemaLocation,attr"`
	MediaPresentationDuration string   `xml:"mediaPresentationDuration,attr"`
	MinBufferTime             string   `xml:"minBufferTime,attr"`
	Periods                   []Period `xml:"Period"`
}

type Period struct {
	XMLName        xml.Name        `xml:"Period"`
	Duration       string          `xml:"duration,attr"`
	AdaptationSets []AdaptationSet `xml:"AdaptationSet"`
}

type AdaptationSet struct {
	XMLName         xml.Name         `xml:"AdaptationSet"`
	MimeType        string           `xml:"mimeType,attr"`
	Codecs          string           `xml:"codecs,attr"`
	Width           string           `xml:"width,attr"`
	Height          string           `xml:"height,attr"`
	FrameRate       string           `xml:"frameRate,attr"`
	Representations []Representation `xml:"Representation"`
}

type Representation struct {
	XMLName     xml.Name    `xml:"Representation"`
	ID          string      `xml:"id,attr"`
	Bandwidth   string      `xml:"bandwidth,attr"`
	Codecs      string      `xml:"codecs,attr"`
	Width       string      `xml:"width,attr"`
	Height      string      `xml:"height,attr"`
	FrameRate   string      `xml:"frameRate,attr"`
	BaseURL     string      `xml:"BaseURL"`
	SegmentList SegmentList `xml:"SegmentList"`
}

type SegmentList struct {
	XMLName     xml.Name     `xml:"SegmentList"`
	Duration    string       `xml:"duration,attr"`
	Timescale   string       `xml:"timescale,attr"`
	SegmentURLs []SegmentURL `xml:"SegmentURL"`
}

type SegmentURL struct {
	XMLName xml.Name `xml:"SegmentURL"`
	Media   string   `xml:"media,attr"`
}

func dash(videoFile, outputDir string) error {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating output directory:", err)
		return err
	}

	// Generate segments with ffmpeg
	err = generateSegments(videoFile, outputDir)
	if err != nil {
		fmt.Println("Error generating segments:", err)
		return err
	}

	// Generate MPD file
	err = generateMPD(outputDir)
	if err != nil {
		fmt.Println("Error generating MPD file:", err)
	}

	fmt.Println("MPD file generated successfully")
	return nil
}

func generateSegments(videoFile, outputDir string) error {
	cmd := exec.Command("ffmpeg", "-i", videoFile, "-c", "copy", "-map", "0", "-f", "segment", "-segment_time", strconv.Itoa(segmentDuration), "-segment_format", "mp4", filepath.Join(outputDir, "segment_%03d.mp4"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

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

	// Use a sample representation
	mpd := MPD{
		XMLNs:                     "urn:mpeg:dash:schema:mpd:2011",
		XMLNsXsi:                  "http://www.w3.org/2001/XMLSchema-instance",
		XsiSchemaLocation:         "urn:mpeg:dash:schema:mpd:2011 http://www.mpegdash.org/schemas/2011/MPD.xsd",
		MediaPresentationDuration: "PT" + strconv.Itoa(len(segmentURLs)*segmentDuration) + "S",
		MinBufferTime:             "PT1.5S",
		Periods: []Period{
			{
				Duration: "PT" + strconv.Itoa(len(segmentURLs)*segmentDuration) + "S",
				AdaptationSets: []AdaptationSet{
					{
						MimeType: "video/mp4",
						Codecs:   "avc1.4d401e",
						Representations: []Representation{
							createRepresentation("144p", "1 Mbps", "150000", "144", "256", "25", "https://example.com/video_144p/", segmentURLs),
							// Add more representations for other resolutions as needed
						},
					},
				},
			},
		},
	}

	file, err := os.Create("output.mpd")
	if err != nil {
		return fmt.Errorf("error creating MPD file: %w", err)
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("  ", "    ")
	return encoder.Encode(mpd)
}

func createRepresentation(resolution, bitrate, bandwidth, width, height, frameRate, baseURL string, segmentURLs []SegmentURL) Representation {
	return Representation{
		ID:        resolution,
		Bandwidth: bandwidth,
		Codecs:    "avc1.4d401e",
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
