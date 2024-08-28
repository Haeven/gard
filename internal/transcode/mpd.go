// internal/transcode/mpd.go
package mpd

import (
	"os"

	dash "github.com/zencoder/go-dash"
)

// GenerateMPD generates a DASH MPD file for a given video with detailed bitrate settings
func GenerateMPD(videoPath string) (string, error) {
	mpd := dash.NewMPD()
	mpd.SetProfile(dash.ProfileLive)
	mpd.SetMinBufferTime(1000)

	// Define the bitrate settings for each resolution
	resolutions := map[string]map[string]int{
		"4K": {
			"standard": 44 * 1000, // 44 Mbps
			"high":     66 * 1000, // 66 Mbps
		},
		"1440p": {
			"standard": 20 * 1000, // 20 Mbps
			"high":     30 * 1000, // 30 Mbps
		},
		"1080p": {
			"standard": 10 * 1000, // 10 Mbps
			"high":     15 * 1000, // 15 Mbps
		},
		"720p": {
			"standard": 6.5 * 1000, // 6.5 Mbps
			"high":     9.5 * 1000, // 9.5 Mbps
		},
		"480p": {
			"standard": 0, // Not supported for HDR
			"high":     0, // Not supported for HDR
		},
		"360p": {
			"standard": 0, // Not supported for HDR
			"high":     0, // Not supported for HDR
		},
	}

	// Define frame rates
	frameRates := []int{24, 30, 50, 60}

	// Add AdaptationSets for each resolution
	for resolution, bitrates := range resolutions {
		for _, framerate := range frameRates {
			if bitrate, ok := bitrates["standard"]; ok && bitrate > 0 {
				addRepresentation(mpd, resolution, bitrate, framerate, "standard")
			}
			if bitrate, ok := bitrates["high"]; ok && bitrate > 0 {
				addRepresentation(mpd, resolution, bitrate, framerate, "high")
			}
		}
	}

	// Save MPD to file
	mpdFilePath := videoPath + ".mpd"
	file, err := os.Create(mpdFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Write MPD data to file
	if err := mpd.Write(file); err != nil {
		return "", err
	}

	return mpdFilePath, nil
}

// addRepresentation adds a video representation to the MPD
func addRepresentation(mpd *dash.MPD, resolution string, bitrate int, frameRate int, profile string) {
	var width, height int
	var codecs string

	// Set dimensions and codecs based on resolution
	switch resolution {
	case "4K":
		width, height = 3840, 2160
		codecs = "vp09.00.10.08" // VP9
	case "1440p":
		width, height = 2560, 1440
		codecs = "vp09.00.10.08" // VP9
	case "1080p":
		width, height = 1920, 1080
		codecs = "vp09.00.10.08" // VP9
	case "720p":
		width, height = 1280, 720
		codecs = "vp09.00.10.08" // VP9
	case "480p":
		width, height = 854, 480
		codecs = "vp09.00.10.08" // VP9
	case "360p":
		width, height = 640, 360
		codecs = "vp09.00.10.08" // VP9
	default:
		return
	}

	adaptationSet := dash.NewAdaptationSet(dash.NewContentType("video"))
	adaptationSet.SetSegmentAlignment(true)
	adaptationSet.SetMaxWidth(width)
	adaptationSet.SetMaxHeight(height)
	adaptationSet.SetMaxFrameRate(frameRate)

	// Create a representation for the given resolution
	rep := dash.NewRepresentation()
	rep.SetBandwidth(bitrate)
	rep.SetWidth(width)
	rep.SetHeight(height)
	rep.SetFrameRate(frameRate)
	rep.SetCodecs(codecs) // Set codec for video

	adaptationSet.AddRepresentation(rep)
	mpd.AddAdaptationSet(adaptationSet)
}
